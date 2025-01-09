//go:build !linux && !freebsd

package zfs

import (
	"Dana"
	"Dana/plugins/inputs"
)

func (*Zfs) Gather(_ Dana.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("zfs", func() Dana.Input {
		return &Zfs{}
	})
}
