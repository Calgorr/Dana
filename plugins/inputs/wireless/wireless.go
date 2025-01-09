//go:generate ../../../tools/readme_config_includer/generator
package wireless

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Wireless is used to store configuration values.
type Wireless struct {
	HostProc string      `toml:"host_proc"`
	Log      Dana.Logger `toml:"-"`
}

func (*Wireless) SampleConfig() string {
	return sampleConfig
}

func init() {
	inputs.Add("wireless", func() Dana.Input {
		return &Wireless{}
	})
}
