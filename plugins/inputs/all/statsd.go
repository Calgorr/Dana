//go:build !custom || inputs || inputs.statsd

package all

import _ "Dana/plugins/inputs/statsd" // register plugin
