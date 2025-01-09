//go:build !linux

package ethtool

import (
	"Dana"
	"Dana/plugins/inputs"
)

type Ethtool struct {
	// This is the list of interface names to include
	InterfaceInclude []string `toml:"interface_include"`

	// This is the list of interface names to ignore
	InterfaceExclude []string `toml:"interface_exclude"`

	// Behavior regarding metrics for downed interfaces
	DownInterfaces string `toml:" down_interfaces"`

	// This is the list of namespace names to include
	NamespaceInclude []string `toml:"namespace_include"`

	// This is the list of namespace names to ignore
	NamespaceExclude []string `toml:"namespace_exclude"`

	// Normalization on the key names
	NormalizeKeys []string `toml:"normalize_keys"`

	Log Dana.Logger `toml:"-"`
}

func (e *Ethtool) Init() error {
	e.Log.Warn("Current platform is not supported")
	return nil
}

func (*Ethtool) Gather(_ Dana.Accumulator) error {
	return nil
}

func init() {
	inputs.Add(pluginName, func() Dana.Input {
		return &Ethtool{}
	})
}
