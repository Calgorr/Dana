//go:build !linux

package cgroup

import (
	"Dana"
)

func (*CGroup) Gather(_ telegraf.Accumulator) error {
	return nil
}
