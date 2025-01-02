//go:build !custom || inputs || inputs.influxdb_listener

package all

import _ "Dana/plugins/inputs/influxdb_listener" // register plugin
