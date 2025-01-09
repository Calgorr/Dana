package processors

import (
	"Dana"
	"Dana/models"
)

// NewStreamingProcessorFromProcessor is a converter that turns a standard
// processor into a streaming processor
func NewStreamingProcessorFromProcessor(p Dana.Processor) Dana.StreamingProcessor {
	sp := &streamingProcessor{
		processor: p,
	}
	return sp
}

type streamingProcessor struct {
	processor Dana.Processor
	acc       Dana.Accumulator
	Log       Dana.Logger
}

func (sp *streamingProcessor) SampleConfig() string {
	return sp.processor.SampleConfig()
}

func (sp *streamingProcessor) Start(acc Dana.Accumulator) error {
	sp.acc = acc
	return nil
}

func (sp *streamingProcessor) Add(m Dana.Metric, acc Dana.Accumulator) error {
	for _, m := range sp.processor.Apply(m) {
		acc.AddMetric(m)
	}
	return nil
}

func (sp *streamingProcessor) Stop() {
}

// Make the streamingProcessor of type Initializer to be able
// to call the Init method of the wrapped processor if
// needed
func (sp *streamingProcessor) Init() error {
	models.SetLoggerOnPlugin(sp.processor, sp.Log)
	if p, ok := sp.processor.(Dana.Initializer); ok {
		err := p.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

// Unwrap lets you retrieve the original Dana.Processor from the
// StreamingProcessor. This is necessary because the toml Unmarshaller won't
// look inside composed types.
func (sp *streamingProcessor) Unwrap() Dana.Processor {
	return sp.processor
}
