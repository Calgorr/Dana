//go:generate ../../../tools/readme_config_includer/generator
package dmcache

import (
	_ "embed"

	"Dana"
	"Dana/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type DMCache struct {
	PerDevice bool `toml:"per_device"`

	getCurrentStatus func() ([]string, error)
}

func (*DMCache) SampleConfig() string {
	return sampleConfig
}

func init() {
	inputs.Add("dmcache", func() Dana.Input {
		return &DMCache{
			PerDevice:        true,
			getCurrentStatus: dmSetupStatus,
		}
	})
}
