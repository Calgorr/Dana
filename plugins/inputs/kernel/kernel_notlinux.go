//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package kernel

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Kernel struct {
	Log Dana.Logger `toml:"-"`
}

func (*Kernel) SampleConfig() string { return sampleConfig }

func (k *Kernel) Init() error {
	k.Log.Warn("Current platform is not supported")
	return nil
}

func (*Kernel) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("kernel", func() Dana.Input {
		return &Kernel{}
	})
}
