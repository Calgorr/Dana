//go:build windows

package varnish

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Varnish struct {
	Log Dana.Logger `toml:"-"`
}

func (v *Varnish) Init() error {
	v.Log.Warn("current platform is not supported")
	return nil
}
func (*Varnish) SampleConfig() string            { return sampleConfig }
func (*Varnish) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("varnish", func() Dana.Input {
		return &Varnish{}
	})
}
