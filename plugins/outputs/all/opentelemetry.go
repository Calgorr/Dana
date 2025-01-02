//go:build !custom || outputs || outputs.opentelemetry

package all

import _ "Dana/plugins/outputs/opentelemetry" // register plugin
