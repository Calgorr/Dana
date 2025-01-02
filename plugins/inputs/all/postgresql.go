//go:build !custom || inputs || inputs.postgresql

package all

import _ "Dana/plugins/inputs/postgresql" // register plugin
