//go:build !custom || inputs || inputs.kernel_vmstat

package all

import _ "Dana/plugins/inputs/kernel_vmstat" // register plugin
