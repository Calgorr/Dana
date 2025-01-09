//go:generate ../../../tools/readme_config_includer/generator
package printer

import (
	_ "embed"
	"fmt"

	"Dana"
	"Dana/plugins/processors"
	"Dana/plugins/serializers/influx"
)

//go:embed sample.conf
var sampleConfig string

type Printer struct {
	influx.Serializer
}

func (*Printer) SampleConfig() string {
	return sampleConfig
}

func (p *Printer) Apply(in ...Dana.Metric) []Dana.Metric {
	for _, metric := range in {
		octets, err := p.Serialize(metric)
		if err != nil {
			continue
		}
		fmt.Print(string(octets))
	}
	return in
}

func init() {
	processors.Add("printer", func() Dana.Processor {
		return &Printer{}
	})
}
