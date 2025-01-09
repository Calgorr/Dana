//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package mdstat

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Mdstat struct {
	Log Dana.Logger `toml:"-"`
}

func (*Mdstat) SampleConfig() string { return sampleConfig }

func (m *Mdstat) Init() error {
	m.Log.Warn("Current platform is not supported")
	return nil
}

func (*Mdstat) Gather(Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("mdstat", func() Dana.Input {
		return &Mdstat{}
	})
}
