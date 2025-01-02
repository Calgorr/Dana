//go:build !custom || inputs || inputs.postgresql_extensible

package all

import _ "Dana/plugins/inputs/postgresql_extensible" // register plugin
