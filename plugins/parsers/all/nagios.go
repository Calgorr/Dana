//go:build !custom || parsers || parsers.nagios

package all

import _ "Dana/plugins/parsers/nagios" // register plugin
