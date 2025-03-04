package parsers

import (
	"Dana"
)

// Creator is the function to create a new parser
type Creator func(defaultMetricName string) Dana.Parser

// Parsers contains the registry of all known parsers (following the new style)
var Parsers = make(map[string]Creator)

// Add adds a parser to the registry. Usually this function is called in the plugin's init function
func Add(name string, creator Creator) {
	Parsers[name] = creator
}
