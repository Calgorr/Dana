//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package kernel_vmstat

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type KernelVmstat struct {
	Log Dana.Logger `toml:"-"`
}

func (*KernelVmstat) SampleConfig() string { return sampleConfig }

func (k *KernelVmstat) Init() error {
	k.Log.Warn("Current platform is not supported")
	return nil
}

func (*KernelVmstat) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("kernel_vmstat", func() Dana.Input {
		return &KernelVmstat{}
	})
}
