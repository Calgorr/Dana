//go:build windows

package processes

import (
	"Dana"
	"Dana/plugins/inputs"
)

type Processes struct {
	Log Dana.Logger
}

func (e *Processes) Init() error {
	e.Log.Warn("Current platform is not supported")
	return nil
}

func (e *Processes) Gather(_ Dana.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("processes", func() Dana.Input {
		return &Processes{}
	})
}
