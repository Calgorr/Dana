//go:build !custom || inputs || inputs.fail2ban

package all

import _ "Dana/plugins/inputs/fail2ban" // register plugin
