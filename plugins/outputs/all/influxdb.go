//go:build !custom || outputs || outputs.influxdb

package all

import _ "Dana/plugins/outputs/influxdb" // register plugin
