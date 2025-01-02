//go:build !custom || inputs || inputs.systemd_units

package all

import _ "Dana/plugins/inputs/systemd_units" // register plugin
