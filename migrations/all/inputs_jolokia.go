//go:build !custom || (migrations && (inputs || inputs.jolokia))

package all

import _ "Dana/migrations/inputs_jolokia" // register migration
