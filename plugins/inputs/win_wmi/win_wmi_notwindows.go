//go:generate ../../../tools/readme_config_includer/generator
//go:build !windows

package win_wmi

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Wmi struct {
	Log Dana.Logger `toml:"-"`
}

func (w *Wmi) Init() error {
	w.Log.Warn("current platform is not supported")
	return nil
}
func (*Wmi) SampleConfig() string            { return sampleConfig }
func (*Wmi) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("win_wmi", func() Dana.Input { return &Wmi{} })
}
