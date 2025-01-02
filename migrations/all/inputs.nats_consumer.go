//go:build !custom || (migrations && (inputs || inputs.nats_consumer))

package all

import _ "Dana/migrations/inputs_nats_consumer" // register migration
