//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux || !amd64

package intel_pmt

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type IntelPMT struct {
	Log Dana.Logger `toml:"-"`
}

func (*IntelPMT) SampleConfig() string { return sampleConfig }

func (p *IntelPMT) Init() error {
	p.Log.Warn("Current platform is not supported")
	return nil
}

func (*IntelPMT) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("intel_pmt", func() Dana.Input {
		return &IntelPMT{}
	})
}
