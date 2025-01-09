//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux || !amd64

package intel_pmu

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type IntelPMU struct {
	Log Dana.Logger `toml:"-"`
}

func (*IntelPMU) SampleConfig() string { return sampleConfig }

func (i *IntelPMU) Init() error {
	i.Log.Warn("Current platform is not supported")
	return nil
}

func (*IntelPMU) Start(_ Dana.Accumulator) error { return nil }

func (*IntelPMU) Gather(_ Dana.Accumulator) error { return nil }

func (*IntelPMU) Stop() {}

func init() {
	inputs.Add("intel_pmu", func() Dana.Input {
		return &IntelPMU{}
	})
}
