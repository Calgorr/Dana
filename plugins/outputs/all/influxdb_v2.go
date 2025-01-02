//go:build !custom || outputs || outputs.influxdb_v2

package all

import _ "Dana/plugins/outputs/influxdb_v2" // register plugin
