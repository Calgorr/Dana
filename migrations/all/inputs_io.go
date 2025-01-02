//go:build !custom || (migrations && (inputs || inputs.io || inputs.diskio))

package all

import _ "Dana/migrations/inputs_io" // register migration
