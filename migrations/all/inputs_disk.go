//go:build !custom || (migrations && (inputs || inputs.disk))

package all

import _ "Dana/migrations/inputs_disk" // register migration
