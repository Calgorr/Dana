//go:build !linux

package wireless

import (
	"Dana"
	"Dana/plugins/inputs"
)

func (w *Wireless) Init() error {
	w.Log.Warn("Current platform is not supported")
	return nil
}

func (*Wireless) Gather(_ Dana.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("wireless", func() Dana.Input {
		return &Wireless{}
	})
}
