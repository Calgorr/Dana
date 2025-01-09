package aggregators

import "Dana"

type Creator func() Dana.Aggregator

var Aggregators = make(map[string]Creator)

func Add(name string, creator Creator) {
	Aggregators[name] = creator
}
