//go:build !custom || inputs || inputs.redis_sentinel

package all

import _ "Dana/plugins/inputs/redis_sentinel" // register plugin
