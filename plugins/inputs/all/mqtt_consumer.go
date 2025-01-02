//go:build !custom || inputs || inputs.mqtt_consumer

package all

import _ "Dana/plugins/inputs/mqtt_consumer" // register plugin
