//go:build !linux

package infiniband

import (
	"Dana"
	"Dana/plugins/inputs"
)

func (i *Infiniband) Init() error {
	i.Log.Warn("Current platform is not supported")
	return nil
}

func (*Infiniband) Gather(_ Dana.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("infiniband", func() Dana.Input {
		return &Infiniband{}
	})
}
