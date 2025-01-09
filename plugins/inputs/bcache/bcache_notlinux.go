//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package bcache

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Bcache struct {
	Log Dana.Logger `toml:"-"`
}

func (*Bcache) SampleConfig() string { return sampleConfig }

func (b *Bcache) Init() error {
	b.Log.Warn("Current platform is not supported")
	return nil
}

func (*Bcache) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("bcache", func() Dana.Input {
		return &Bcache{}
	})
}
