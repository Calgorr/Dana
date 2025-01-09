//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package slab

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Slab struct {
	Log Dana.Logger `toml:"-"`
}

func (s *Slab) Init() error {
	s.Log.Warn("current platform is not supported")
	return nil
}
func (*Slab) SampleConfig() string            { return sampleConfig }
func (*Slab) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("slab", func() Dana.Input {
		return &Slab{}
	})
}
