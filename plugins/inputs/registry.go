package inputs

import "Dana"

type Creator func() Dana.Input

var Inputs = make(map[string]Creator)

func Add(name string, creator Creator) {
	Inputs[name] = creator
}
