//go:build !custom || (migrations && (inputs || inputs.mqtt_consumer))

package all

import _ "Dana/migrations/inputs_mqtt_consumer" // register migration
