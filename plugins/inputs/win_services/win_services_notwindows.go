//go:generate ../../../tools/readme_config_includer/generator
//go:build !windows

package win_services

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type WinServices struct {
	Log Dana.Logger `toml:"-"`
}

func (*WinServices) SampleConfig() string { return sampleConfig }

func (w *WinServices) Init() error {
	w.Log.Warn("Current platform is not supported")
	return nil
}
func (*WinServices) Gather(Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("win_services", func() Dana.Input {
		return &WinServices{}
	})
}
