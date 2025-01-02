//go:build !custom || inputs || inputs.snmp_trap

package all

import _ "Dana/plugins/inputs/snmp_trap" // register plugin
