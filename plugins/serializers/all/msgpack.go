//go:build !custom || serializers || serializers.msgpack

package all

import (
	_ "Dana/plugins/serializers/msgpack" // register plugin
)
