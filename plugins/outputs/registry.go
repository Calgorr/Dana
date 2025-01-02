package outputs

import (
	"Dana"
)

type Creator func() telegraf.Output

var Outputs = make(map[string]Creator)

func Add(name string, creator Creator) {
	Outputs[name] = creator
}
