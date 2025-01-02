//go:build !custom || inputs || inputs.dns_query

package all

import _ "Dana/plugins/inputs/dns_query" // register plugin
