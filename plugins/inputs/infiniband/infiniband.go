//go:generate ../../../tools/readme_config_includer/generator
package infiniband

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Infiniband struct {
	Log telegraf.Logger `toml:"-"`
}

func (*Infiniband) SampleConfig() string {
	return sampleConfig
}

// Initialise plugin
func init() {
	inputs.Add("infiniband", func() telegraf.Input { return &Infiniband{} })
}
