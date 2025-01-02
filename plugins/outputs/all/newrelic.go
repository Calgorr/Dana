//go:build !custom || outputs || outputs.newrelic

package all

import _ "Dana/plugins/outputs/newrelic" // register plugin
