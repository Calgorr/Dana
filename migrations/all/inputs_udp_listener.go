//go:build !custom || (migrations && (inputs || inputs.udp_listener))

package all

import _ "Dana/migrations/inputs_udp_listener" // register migration
