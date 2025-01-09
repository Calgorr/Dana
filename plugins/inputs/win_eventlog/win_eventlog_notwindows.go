//go:generate ../../../tools/readme_config_includer/generator
//go:build !windows

package win_eventlog

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type WinEventLog struct {
	Log Dana.Logger `toml:"-"`
}

func (w *WinEventLog) Init() error {
	w.Log.Warn("current platform is not supported")
	return nil
}
func (*WinEventLog) SampleConfig() string            { return sampleConfig }
func (*WinEventLog) Gather(_ Dana.Accumulator) error { return nil }
func (*WinEventLog) Start(_ Dana.Accumulator) error  { return nil }
func (*WinEventLog) Stop()                           {}

func init() {
	inputs.Add("win_eventlog", func() Dana.Input {
		return &WinEventLog{}
	})
}
