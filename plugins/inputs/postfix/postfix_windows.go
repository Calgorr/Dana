//go:build windows

package postfix

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Postfix struct {
	Log Dana.Logger `toml:"-"`
}

func (*Postfix) SampleConfig() string { return sampleConfig }

func (p *Postfix) Init() error {
	p.Log.Warn("Current platform is not supported")
	return nil
}

func (*Postfix) Gather(_ Dana.Accumulator) error { return nil }

func init() {
	inputs.Add("postfix", func() Dana.Input {
		return &Postfix{}
	})
}
