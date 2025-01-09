//go:build !linux

package linux_cpu

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type LinuxCPU struct {
	Log Dana.Logger `toml:"-"`
}

func (*LinuxCPU) SampleConfig() string { return sampleConfig }

func (l *LinuxCPU) Init() error {
	l.Log.Warn("Current platform is not supported")
	return nil
}

func (*LinuxCPU) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("linux_cpu", func() Dana.Input {
		return &LinuxCPU{}
	})
}
