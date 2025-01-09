//go:build windows

package socketstat

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Socketstat struct {
	Log Dana.Logger `toml:"-"`
}

func (s *Socketstat) Init() error {
	s.Log.Warn("current platform is not supported")
	return nil
}
func (*Socketstat) SampleConfig() string            { return sampleConfig }
func (*Socketstat) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("socketstat", func() Dana.Input {
		return &Socketstat{}
	})
}
