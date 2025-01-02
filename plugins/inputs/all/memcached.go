//go:build !custom || inputs || inputs.memcached

package all

import _ "Dana/plugins/inputs/memcached" // register plugin
