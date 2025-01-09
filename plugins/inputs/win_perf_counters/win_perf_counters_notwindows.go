//go:generate ../../../tools/readme_config_includer/generator
//go:build !windows

package win_perf_counters

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type WinPerfCounters struct {
	Log Dana.Logger `toml:"-"`
}

func (*WinPerfCounters) SampleConfig() string { return sampleConfig }

func (w *WinPerfCounters) Init() error {
	w.Log.Warn("current platform is not supported")
	return nil
}

func (*WinPerfCounters) Gather(Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("win_perf_counters", func() Dana.Input {
		return &WinPerfCounters{}
	})
}
