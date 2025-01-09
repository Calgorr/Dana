//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux || !amd64

package intel_baseband

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Baseband struct {
	Log Dana.Logger `toml:"-"`
}

func (*Baseband) SampleConfig() string { return sampleConfig }

func (b *Baseband) Init() error {
	b.Log.Warn("Current platform is not supported")
	return nil
}
func (*Baseband) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("intel_baseband", func() Dana.Input {
		return &Baseband{}
	})
}
