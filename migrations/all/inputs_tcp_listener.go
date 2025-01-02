//go:build !custom || (migrations && (inputs || inputs.tcp_listener))

package all

import _ "Dana/migrations/inputs_tcp_listener" // register migration
