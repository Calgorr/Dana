//go:build !custom || secretstores || secretstores.systemd

package all

import _ "Dana/plugins/secretstores/systemd" // register plugin
