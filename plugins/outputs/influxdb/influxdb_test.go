package influxdb_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"Dana"
	"Dana/config"
	"Dana/metric"
	"Dana/plugins/common/tls"
	"Dana/plugins/outputs/influxdb"
	"Dana/testutil"
)

type MockClient struct {
	URLF            func() string
	WriteF          func() error
	CreateDatabaseF func() error
	DatabaseF       func() string
	CloseF          func()

	log Dana.Logger
}

func (c *MockClient) URL() string {
	return c.URLF()
}

func (c *MockClient) Write(context.Context, []Dana.Metric) error {
	return c.WriteF()
}

func (c *MockClient) CreateDatabase(context.Context, string) error {
	return c.CreateDatabaseF()
}

func (c *MockClient) Database() string {
	return c.DatabaseF()
}

func (c *MockClient) Close() {
	c.CloseF()
}

func (c *MockClient) SetLogger(log Dana.Logger) {
	c.log = log
}

func TestDeprecatedURLSupport(t *testing.T) {
	var actual *influxdb.UDPConfig
	output := influxdb.InfluxDB{
		URLs: []string{"udp://localhost:8089"},
		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			actual = config
			return &MockClient{}, nil
		},
	}

	output.Log = testutil.Logger{}

	err := output.Connect()
	require.NoError(t, err)
	require.Equal(t, "udp://localhost:8089", actual.URL.String())
}

func TestDefaultURL(t *testing.T) {
	var actual *influxdb.HTTPConfig
	output := influxdb.InfluxDB{
		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			actual = config
			return &MockClient{
				DatabaseF: func() string {
					return "Dana2"
				},
				CreateDatabaseF: func() error {
					return nil
				},
			}, nil
		},
	}

	output.Log = testutil.Logger{}

	err := output.Connect()
	require.NoError(t, err)
	require.Equal(t, "http://localhost:8086", actual.URL.String())
}

func TestConnectUDPConfig(t *testing.T) {
	var actual *influxdb.UDPConfig

	output := influxdb.InfluxDB{
		URLs:       []string{"udp://localhost:8089"},
		UDPPayload: config.Size(42),

		CreateUDPClientF: func(config *influxdb.UDPConfig) (influxdb.Client, error) {
			actual = config
			return &MockClient{}, nil
		},
	}
	output.Log = testutil.Logger{}

	err := output.Connect()
	require.NoError(t, err)

	require.Equal(t, "udp://localhost:8089", actual.URL.String())
	require.Equal(t, 42, actual.MaxPayloadSize)
	require.NotNil(t, actual.Serializer)
}

func TestConnectHTTPConfig(t *testing.T) {
	var actual *influxdb.HTTPConfig

	output := influxdb.InfluxDB{
		URLs:             []string{"http://localhost:8086"},
		Database:         "Dana2",
		RetentionPolicy:  "default",
		WriteConsistency: "any",
		Timeout:          config.Duration(5 * time.Second),
		Username:         config.NewSecret([]byte("guy")),
		Password:         config.NewSecret([]byte("smiley")),
		UserAgent:        "Dana2",
		HTTPProxy:        "http://localhost:8086",
		HTTPHeaders: map[string]string{
			"x": "y",
		},
		ContentEncoding: "gzip",
		ClientConfig: tls.ClientConfig{
			InsecureSkipVerify: true,
		},

		CreateHTTPClientF: func(config *influxdb.HTTPConfig) (influxdb.Client, error) {
			actual = config
			return &MockClient{
				DatabaseF: func() string {
					return "Dana2"
				},
				CreateDatabaseF: func() error {
					return nil
				},
			}, nil
		},
	}

	output.Log = testutil.Logger{}

	err := output.Connect()
	require.NoError(t, err)

	require.Equal(t, output.URLs[0], actual.URL.String())
	require.Equal(t, output.UserAgent, actual.UserAgent)
	require.Equal(t, time.Duration(output.Timeout), actual.Timeout)
	require.Equal(t, output.Username, actual.Username)
	require.Equal(t, output.Password, actual.Password)
	require.Equal(t, output.HTTPProxy, actual.Proxy.String())
	require.Equal(t, output.HTTPHeaders, actual.Headers)
	require.Equal(t, output.ContentEncoding, actual.ContentEncoding)
	require.Equal(t, output.Database, actual.Database)
	require.Equal(t, output.RetentionPolicy, actual.RetentionPolicy)
	require.Equal(t, output.WriteConsistency, actual.Consistency)
	require.NotNil(t, actual.TLSConfig)
	require.NotNil(t, actual.Serializer)

	require.Equal(t, output.Database, actual.Database)
}

func TestWriteRecreateDatabaseIfDatabaseNotFound(t *testing.T) {
	output := influxdb.InfluxDB{
		URLs: []string{"http://localhost:8086"},
		CreateHTTPClientF: func(*influxdb.HTTPConfig) (influxdb.Client, error) {
			return &MockClient{
				DatabaseF: func() string {
					return "Dana2"
				},
				CreateDatabaseF: func() error {
					return nil
				},
				WriteF: func() error {
					return &influxdb.DatabaseNotFoundError{
						APIError: influxdb.APIError{
							StatusCode:  http.StatusNotFound,
							Title:       "404 Not Found",
							Description: `database not found "Dana2"`,
						},
					}
				},
				URLF: func() string {
					return "http://localhost:8086"
				},
			}, nil
		},
	}

	output.Log = testutil.Logger{}

	err := output.Connect()
	require.NoError(t, err)

	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	metrics := []Dana.Metric{m}

	err = output.Write(metrics)
	// We only have one URL, so we expect an error
	require.Error(t, err)
}

func TestInfluxDBLocalAddress(t *testing.T) {
	output := influxdb.InfluxDB{
		URLs:      []string{"http://localhost:8086"},
		LocalAddr: "localhost",

		CreateHTTPClientF: func(_ *influxdb.HTTPConfig) (influxdb.Client, error) {
			return &MockClient{
				DatabaseF: func() string {
					return "Dana2"
				},
				CreateDatabaseF: func() error {
					return nil
				},
			}, nil
		},
	}

	require.NoError(t, output.Connect())
}
