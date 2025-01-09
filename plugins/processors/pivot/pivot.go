//go:generate ../../../tools/readme_config_includer/generator
package pivot

import (
	_ "embed"

	"Dana"
	"Dana/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Pivot struct {
	TagKey   string `toml:"tag_key"`
	ValueKey string `toml:"value_key"`
}

func (*Pivot) SampleConfig() string {
	return sampleConfig
}

func (p *Pivot) Apply(metrics ...Dana.Metric) []Dana.Metric {
	for _, m := range metrics {
		key, ok := m.GetTag(p.TagKey)
		if !ok {
			continue
		}

		value, ok := m.GetField(p.ValueKey)
		if !ok {
			continue
		}

		m.RemoveTag(p.TagKey)
		m.RemoveField(p.ValueKey)
		m.AddField(key, value)
	}
	return metrics
}

func init() {
	processors.Add("pivot", func() Dana.Processor {
		return &Pivot{}
	})
}
