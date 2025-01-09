//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package intel_dlb

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type IntelDLB struct {
	Log Dana.Logger `toml:"-"`
}

func (*IntelDLB) SampleConfig() string { return sampleConfig }

func (i *IntelDLB) Init() error {
	i.Log.Warn("Current platform is not supported")
	return nil
}
func (*IntelDLB) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("intel_dlb", func() Dana.Input {
		return &IntelDLB{}
	})
}
