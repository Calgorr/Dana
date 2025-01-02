//go:build !custom || inputs || inputs.syslog

package all

import _ "Dana/plugins/inputs/syslog" // register plugin
