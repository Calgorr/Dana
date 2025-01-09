//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package sysstat

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Sysstat struct {
	Log Dana.Logger `toml:"-"`
}

func (s *Sysstat) Init() error {
	s.Log.Warn("current platform is not supported")
	return nil
}
func (*Sysstat) SampleConfig() string            { return sampleConfig }
func (*Sysstat) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("sysstat", func() Dana.Input {
		return &Sysstat{}
	})
}
