package parallel

import (
	"sync"

	"Dana"
)

type Unordered struct {
	wg      sync.WaitGroup
	acc     Dana.Accumulator
	fn      func(Dana.Metric) []Dana.Metric
	inQueue chan Dana.Metric
}

func NewUnordered(
	acc Dana.Accumulator,
	fn func(Dana.Metric) []Dana.Metric,
	workerCount int,
) *Unordered {
	p := &Unordered{
		acc:     acc,
		inQueue: make(chan Dana.Metric, workerCount),
		fn:      fn,
	}

	// start workers
	p.wg.Add(1)
	go func() {
		p.startWorkers(workerCount)
		p.wg.Done()
	}()

	return p
}

func (p *Unordered) startWorkers(count int) {
	wg := sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			for metric := range p.inQueue {
				for _, m := range p.fn(metric) {
					p.acc.AddMetric(m)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func (p *Unordered) Stop() {
	close(p.inQueue)
	p.wg.Wait()
}

func (p *Unordered) Enqueue(m Dana.Metric) {
	p.inQueue <- m
}
