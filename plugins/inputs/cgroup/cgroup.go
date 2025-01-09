//go:generate ../../../tools/readme_config_includer/generator
package cgroup

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type CGroup struct {
	Paths []string `toml:"paths"`
	Files []string `toml:"files"`

	logged map[string]bool
}

func (*CGroup) SampleConfig() string {
	return sampleConfig
}

func (cg *CGroup) Init() error {
	cg.logged = make(map[string]bool)

	return nil
}

func init() {
	inputs.Add("cgroup", func() Dana.Input { return &CGroup{} })
}
