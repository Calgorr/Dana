//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package dpdk

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Dpdk struct {
	Log Dana.Logger `toml:"-"`
}

func (*Dpdk) SampleConfig() string { return sampleConfig }

func (d *Dpdk) Init() error {
	d.Log.Warn("Current platform is not supported")
	return nil
}

func (*Dpdk) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("dpdk", func() Dana.Input {
		return &Dpdk{}
	})
}
