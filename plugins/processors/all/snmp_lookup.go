//go:build !custom || processors || processors.snmp_lookup

package all

import _ "Dana/plugins/processors/snmp_lookup" // register plugin
