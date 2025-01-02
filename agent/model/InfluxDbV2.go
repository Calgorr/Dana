package model

type InfluxDbV2 struct {
	ID                    int      `json:"id"`
	ServiceAddress        string   `json:"service_address"`
	MaxUndeliveredMetrics int      `json:"max_undelivered_metrics"`
	ReadTimeout           string   `json:"read_timeout"`
	WriteTimeout          string   `json:"write_timeout"`
	MaxBodySize           string   `json:"max_body_size"`
	BucketTag             string   `json:"bucket_tag"`
	TLSAllowedCacerts     []string `json:"tls_allowed_cacerts"`
	TLSCert               string   `json:"tls_cert"`
	TLSKey                string   `json:"tls_key"`
	Token                 string   `json:"token"`
	ParserType            string   `json:"parser_type"`
}
