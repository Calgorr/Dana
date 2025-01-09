//go:build !linux

package synproxy

import (
	"Dana"
	"Dana/plugins/inputs"
)

func (k *Synproxy) Init() error {
	k.Log.Warn("Current platform is not supported")
	return nil
}

func (*Synproxy) Gather(_ Dana.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("synproxy", func() Dana.Input {
		return &Synproxy{}
	})
}
