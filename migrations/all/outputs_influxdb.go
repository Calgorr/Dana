//go:build !custom || (migrations && (outputs || outputs.influxdb))

package all

import _ "Dana/migrations/outputs_influxdb" // register migration
