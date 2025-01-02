//go:build !custom || (migrations && (inputs || inputs.gnmi))

package all

import _ "Dana/migrations/inputs_gnmi" // register migration
