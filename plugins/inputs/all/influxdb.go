//go:build !custom || inputs || inputs.influxdb

package all

import _ "Dana/plugins/inputs/influxdb" // register plugin
