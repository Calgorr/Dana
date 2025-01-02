package parallel

import "Dana"

type Parallel interface {
	Enqueue(telegraf.Metric)
	Stop()
}
