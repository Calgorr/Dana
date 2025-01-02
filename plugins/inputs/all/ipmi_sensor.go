//go:build !custom || inputs || inputs.ipmi_sensor

package all

import _ "Dana/plugins/inputs/ipmi_sensor" // register plugin
