package processors

import "Dana"

type Creator func() Dana.Processor
type StreamingCreator func() Dana.StreamingProcessor

// HasUnwrap indicates the presence of an Unwrap() function to retrieve the
// underlying Dana.Processor.
type HasUnwrap interface {
	Unwrap() Dana.Processor
}

// all processors are streaming processors.
// Dana.Processor processors are upgraded to Dana.StreamingProcessor
var Processors = make(map[string]StreamingCreator)

// Add adds a Dana.Processor processor
func Add(name string, creator Creator) {
	Processors[name] = upgradeToStreamingProcessor(creator)
}

// AddStreaming adds a Dana.StreamingProcessor streaming processor
func AddStreaming(name string, creator StreamingCreator) {
	Processors[name] = creator
}

func upgradeToStreamingProcessor(oldCreator Creator) StreamingCreator {
	return func() Dana.StreamingProcessor {
		return NewStreamingProcessorFromProcessor(oldCreator())
	}
}
