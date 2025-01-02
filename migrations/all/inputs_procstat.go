//go:build !custom || (migrations && (inputs || inputs.procstat))

package all

import _ "Dana/migrations/inputs_procstat" // register migration
