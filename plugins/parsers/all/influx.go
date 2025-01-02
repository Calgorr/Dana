//go:build !custom || parsers || parsers.influx

package all

import (
	_ "Dana/plugins/parsers/influx"                 // register plugin
	_ "Dana/plugins/parsers/influx/influx_upstream" // register plugin
)
