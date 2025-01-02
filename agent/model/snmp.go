package model

type Snmp struct {
	ID             int      `json:"id"`
	ServiceAddress string   `json:"service_address"`
	Path           []string `json:"path"`
	Timeout        string   `json:"timeout"`
	Version        string   `json:"version"`
	SecName        string   `json:"sec_name"`
	AuthProtocol   string   `json:"auth_protocol"`
	AuthPassword   string   `json:"auth_password"`
	SecLevel       string   `json:"sec_level"`
	PrivProtocol   string   `json:"priv_protocol"`
	PrivPassword   string   `json:"priv_password"`
}
