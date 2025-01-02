//go:build !custom || inputs || inputs.conntrack

package all

import _ "Dana/plugins/inputs/conntrack" // register plugin
