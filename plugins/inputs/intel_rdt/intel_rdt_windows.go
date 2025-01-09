//go:build windows

package intel_rdt

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type IntelRDT struct {
	Log Dana.Logger `toml:"-"`
}

func (*IntelRDT) SampleConfig() string { return sampleConfig }

func (i *IntelRDT) Init() error {
	i.Log.Warn("Current platform is not supported")
	return nil
}

func (*IntelRDT) Start(_ Dana.Accumulator) error { return nil }

func (*IntelRDT) Gather(_ Dana.Accumulator) error { return nil }

func (*IntelRDT) Stop() {}

func init() {
	inputs.Add("intel_rdt", func() Dana.Input {
		return &IntelRDT{}
	})
}
