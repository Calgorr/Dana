package httpconfig

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/peterbourgon/unixtransport"

	"Dana"
	"Dana/config"
	"Dana/plugins/common/cookie"
	"Dana/plugins/common/oauth"
	"Dana/plugins/common/proxy"
	"Dana/plugins/common/tls"
)

// Common HTTP client struct.
type HTTPClientConfig struct {
	Timeout               config.Duration `toml:"timeout"`
	IdleConnTimeout       config.Duration `toml:"idle_conn_timeout"`
	MaxIdleConns          int             `toml:"max_idle_conn"`
	MaxIdleConnsPerHost   int             `toml:"max_idle_conn_per_host"`
	ResponseHeaderTimeout config.Duration `toml:"response_timeout"`

	proxy.HTTPProxy
	tls.ClientConfig
	oauth.OAuth2Config
	cookie.CookieAuthConfig
}

func (h *HTTPClientConfig) CreateClient(ctx context.Context, log Dana.Logger) (*http.Client, error) {
	tlsCfg, err := h.ClientConfig.TLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to set TLS config: %w", err)
	}

	prox, err := h.HTTPProxy.Proxy()
	if err != nil {
		return nil, fmt.Errorf("failed to set proxy: %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig:       tlsCfg,
		Proxy:                 prox,
		IdleConnTimeout:       time.Duration(h.IdleConnTimeout),
		MaxIdleConns:          h.MaxIdleConns,
		MaxIdleConnsPerHost:   h.MaxIdleConnsPerHost,
		ResponseHeaderTimeout: time.Duration(h.ResponseHeaderTimeout),
	}

	// Register "http+unix" and "https+unix" protocol handler.
	unixtransport.Register(transport)

	client := &http.Client{
		Transport: transport,
	}

	// While CreateOauth2Client returns a http.Client keeping the Transport configuration,
	// it does not keep other http.Client parameters (e.g. Timeout).
	client = h.OAuth2Config.CreateOauth2Client(ctx, client)

	if h.CookieAuthConfig.URL != "" {
		if err := h.CookieAuthConfig.Start(client, log, clock.New()); err != nil {
			return nil, err
		}
	}

	timeout := h.Timeout
	if timeout == 0 {
		timeout = config.Duration(time.Second * 5)
	}
	client.Timeout = time.Duration(timeout)

	return client, nil
}
