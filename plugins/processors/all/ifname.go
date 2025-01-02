//go:build !custom || processors || processors.ifname

package all

import _ "Dana/plugins/processors/ifname" // register plugin
