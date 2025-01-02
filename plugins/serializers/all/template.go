//go:build !custom || serializers || serializers.template

package all

import (
	_ "Dana/plugins/serializers/template" // register plugin
)
