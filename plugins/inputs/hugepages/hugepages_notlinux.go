//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package hugepages

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Hugepages struct {
	Log Dana.Logger `toml:"-"`
}

func (*Hugepages) SampleConfig() string {
	return sampleConfig
}

func (h *Hugepages) Init() error {
	h.Log.Warn("Current platform is not supported")
	return nil
}

func (*Hugepages) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("hugepages", func() Dana.Input {
		return &Hugepages{}
	})
}
