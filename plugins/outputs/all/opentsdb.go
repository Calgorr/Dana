//go:build !custom || outputs || outputs.opentsdb

package all

import _ "Dana/plugins/outputs/opentsdb" // register plugin
