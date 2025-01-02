//go:build !custom || inputs || inputs.iptables

package all

import _ "Dana/plugins/inputs/iptables" // register plugin
