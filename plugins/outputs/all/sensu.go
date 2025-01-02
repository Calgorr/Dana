//go:build !custom || outputs || outputs.sensu

package all

import _ "Dana/plugins/outputs/sensu" // register plugin
