//go:generate ../../../tools/readme_config_includer/generator
package discard

import (
	_ "embed"

	"Dana"
	"Dana/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type Discard struct{}

func (*Discard) SampleConfig() string {
	return sampleConfig
}

func (d *Discard) Connect() error { return nil }
func (d *Discard) Close() error   { return nil }
func (d *Discard) Write(_ []Dana.Metric) error {
	return nil
}

func init() {
	outputs.Add("discard", func() Dana.Output { return &Discard{} })
}
