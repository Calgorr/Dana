//go:build !custom || (migrations && (inputs || inputs.snmp_legacy))

package all

import _ "Dana/migrations/inputs_snmp_legacy" // register migration
