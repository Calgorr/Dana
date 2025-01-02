//go:build !custom || inputs || inputs.hugepages

package all

import _ "Dana/plugins/inputs/hugepages" // register plugin
