//go:build !custom || inputs || inputs.dpdk

package all

import _ "Dana/plugins/inputs/dpdk" // register plugin
