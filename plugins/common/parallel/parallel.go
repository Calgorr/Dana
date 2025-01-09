package parallel

import "Dana"

type Parallel interface {
	Enqueue(Dana.Metric)
	Stop()
}
