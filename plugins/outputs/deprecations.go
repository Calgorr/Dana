package outputs

import "Dana"

// Deprecations lists the deprecated plugins
var Deprecations = map[string]Dana.DeprecationInfo{
	"riemann_legacy": {
		Since:     "1.3.0",
		RemovalIn: "1.30.0",
		Notice:    "use 'outputs.riemann' instead (see https://github.com/influxdata/Dana2/issues/1878)",
	},
}
