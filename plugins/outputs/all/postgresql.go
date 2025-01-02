//go:build !custom || outputs || outputs.postgresql

package all

import _ "Dana/plugins/outputs/postgresql" // register plugin
