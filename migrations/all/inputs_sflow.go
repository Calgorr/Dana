//go:build !custom || (migrations && (inputs || inputs.sflow))

package all

import _ "Dana/migrations/inputs_sflow" // register migration
