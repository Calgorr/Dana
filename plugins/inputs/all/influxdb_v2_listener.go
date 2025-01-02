//go:build !custom || inputs || inputs.influxdb_v2_listener

package all

import _ "Dana/plugins/inputs/influxdb_v2_listener" // register plugin
