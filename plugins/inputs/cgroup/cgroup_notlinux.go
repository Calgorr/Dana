//go:build !linux

package cgroup

import (
	"Dana"
)

func (*CGroup) Gather(_ Dana.Accumulator) error {
	return nil
}
