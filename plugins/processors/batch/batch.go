package batch

import (
	"Dana"
	"Dana/plugins/processors"
	_ "embed"
	"strconv"
	"sync/atomic"
)

//go:embed sample.conf
var sampleConfig string

type Batch struct {
	BatchTag     string `toml:"batch_tag"`
	NumBatches   uint64 `toml:"batches"`
	SkipExisting bool   `toml:"skip_existing"`

	// the number of metrics that have been processed so far
	count atomic.Uint64
}

func (*Batch) SampleConfig() string {
	return sampleConfig
}

func (b *Batch) Apply(in ...Dana.Metric) []Dana.Metric {
	out := make([]Dana.Metric, 0, len(in))
	for _, m := range in {
		if b.SkipExisting && m.HasTag(b.BatchTag) {
			out = append(out, m)
			continue
		}

		oldCount := b.count.Add(1) - 1
		batchID := oldCount % b.NumBatches
		m.AddTag(b.BatchTag, strconv.FormatUint(batchID, 10))
		out = append(out, m)
	}

	return out
}

func init() {
	processors.Add("batch", func() Dana.Processor {
		return &Batch{}
	})
}
