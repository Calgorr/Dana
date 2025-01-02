//go:build !custom || parsers || parsers.collectd

package all

import _ "Dana/plugins/parsers/collectd" // register plugin
