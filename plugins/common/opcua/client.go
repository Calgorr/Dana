package opcua

import (
	"context"
	"errors"
	"fmt"
	"log" //nolint:depguard // just for debug
	"net/url"
	"strconv"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/ua"

	"Dana"
	"Dana/config"
	"Dana/internal/choice"
)

type OpcUAWorkarounds struct {
	AdditionalValidStatusCodes []string `toml:"additional_valid_status_codes"`
}

type ConnectionState opcua.ConnState

const (
	Closed       ConnectionState = ConnectionState(opcua.Closed)
	Connected    ConnectionState = ConnectionState(opcua.Connected)
	Connecting   ConnectionState = ConnectionState(opcua.Connecting)
	Disconnected ConnectionState = ConnectionState(opcua.Disconnected)
	Reconnecting ConnectionState = ConnectionState(opcua.Reconnecting)
)

func (c ConnectionState) String() string {
	return opcua.ConnState(c).String()
}

type OpcUAClientConfig struct {
	Endpoint       string          `toml:"endpoint"`
	SecurityPolicy string          `toml:"security_policy"`
	SecurityMode   string          `toml:"security_mode"`
	Certificate    string          `toml:"certificate"`
	PrivateKey     string          `toml:"private_key"`
	Username       config.Secret   `toml:"username"`
	Password       config.Secret   `toml:"password"`
	AuthMethod     string          `toml:"auth_method"`
	ConnectTimeout config.Duration `toml:"connect_timeout"`
	RequestTimeout config.Duration `toml:"request_timeout"`
	ClientTrace    bool            `toml:"client_trace"`

	OptionalFields []string         `toml:"optional_fields"`
	Workarounds    OpcUAWorkarounds `toml:"workarounds"`
	SessionTimeout config.Duration  `toml:"session_timeout"`
}

func (o *OpcUAClientConfig) Validate() error {
	if err := o.validateOptionalFields(); err != nil {
		return fmt.Errorf("invalid 'optional_fields': %w", err)
	}

	return o.validateEndpoint()
}

func (o *OpcUAClientConfig) validateOptionalFields() error {
	validFields := []string{"DataType"}
	return choice.CheckSlice(o.OptionalFields, validFields)
}

func (o *OpcUAClientConfig) validateEndpoint() error {
	if o.Endpoint == "" {
		return errors.New("endpoint url is empty")
	}

	_, err := url.Parse(o.Endpoint)
	if err != nil {
		return errors.New("endpoint url is invalid")
	}

	switch o.SecurityPolicy {
	case "None", "Basic128Rsa15", "Basic256", "Basic256Sha256", "auto":
	default:
		return fmt.Errorf("invalid security type %q in %q", o.SecurityPolicy, o.Endpoint)
	}

	switch o.SecurityMode {
	case "None", "Sign", "SignAndEncrypt", "auto":
	default:
		return fmt.Errorf("invalid security type %q in %q", o.SecurityMode, o.Endpoint)
	}

	return nil
}

func (o *OpcUAClientConfig) CreateClient(Dana2Logger Dana.Logger) (*OpcUAClient, error) {
	err := o.Validate()
	if err != nil {
		return nil, err
	}

	if o.ClientTrace {
		debug.Enable = true
		debug.Logger = log.New(&DebugLogger{Log: Dana2Logger}, "", 0)
	}

	c := &OpcUAClient{
		Config: o,
		Log:    Dana2Logger,
	}
	c.Log.Debug("Initialising OpcUAClient")

	err = c.setupWorkarounds()
	return c, err
}

type OpcUAClient struct {
	Config *OpcUAClientConfig
	Log    Dana.Logger

	Client *opcua.Client

	opts  []opcua.Option
	codes []ua.StatusCode
}

// / setupOptions read the endpoints from the specified server and setup all authentication
func (o *OpcUAClient) SetupOptions() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.Config.ConnectTimeout))
	defer cancel()
	// Get a list of the endpoints for our target server
	endpoints, err := opcua.GetEndpoints(ctx, o.Config.Endpoint)
	if err != nil {
		return err
	}

	if o.Config.Certificate == "" && o.Config.PrivateKey == "" {
		if o.Config.SecurityPolicy != "None" || o.Config.SecurityMode != "None" {
			o.Log.Debug("Generating self-signed certificate")
			cert, privateKey, err := generateCert("urn:Dana2:gopcua:client", 2048,
				o.Config.Certificate, o.Config.PrivateKey, 365*24*time.Hour)
			if err != nil {
				return err
			}

			o.Config.Certificate = cert
			o.Config.PrivateKey = privateKey
		}
	}

	o.Log.Debug("Configuring OPC UA connection options")
	o.opts, err = o.generateClientOpts(endpoints)

	return err
}

func (o *OpcUAClient) setupWorkarounds() error {
	o.codes = []ua.StatusCode{ua.StatusOK}
	for _, c := range o.Config.Workarounds.AdditionalValidStatusCodes {
		val, err := strconv.ParseUint(c, 0, 32) // setting 32 bits to allow for safe conversion
		if err != nil {
			return err
		}
		o.codes = append(o.codes, ua.StatusCode(val))
	}

	return nil
}

func (o *OpcUAClient) StatusCodeOK(code ua.StatusCode) bool {
	for _, val := range o.codes {
		if val == code {
			return true
		}
	}
	return false
}

// Connect to an OPC UA device
func (o *OpcUAClient) Connect(ctx context.Context) error {
	o.Log.Debug("Connecting OPC UA Client to server")
	u, err := url.Parse(o.Config.Endpoint)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "opc.tcp":
		if err := o.SetupOptions(); err != nil {
			return err
		}

		if o.Client != nil {
			o.Log.Warnf("Closing connection to %q as already connected", u)
			if err := o.Client.Close(ctx); err != nil {
				// Only log the error but to not bail-out here as this prevents
				// reconnections for multiple parties (see e.g. #9523).
				o.Log.Errorf("Closing connection failed: %v", err)
			}
		}

		o.Client, err = opcua.NewClient(o.Config.Endpoint, o.opts...)
		if err != nil {
			return fmt.Errorf("error in new client: %w", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.Config.ConnectTimeout))
		defer cancel()
		if err := o.Client.Connect(ctx); err != nil {
			return fmt.Errorf("error in Client Connection: %w", err)
		}
		o.Log.Debug("Connected to OPC UA Server")

	default:
		return fmt.Errorf("unsupported scheme %q in endpoint. Expected opc.tcp", u.Scheme)
	}
	return nil
}

func (o *OpcUAClient) Disconnect(ctx context.Context) error {
	o.Log.Debug("Disconnecting from OPC UA Server")
	u, err := url.Parse(o.Config.Endpoint)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "opc.tcp":
		// We can't do anything about failing to close a connection
		err := o.Client.Close(ctx)
		o.Client = nil
		return err
	default:
		return errors.New("invalid controller")
	}
}

func (o *OpcUAClient) State() ConnectionState {
	if o.Client == nil {
		return Disconnected
	}
	return ConnectionState(o.Client.State())
}
