//go:build !custom || serializers || serializers.cloudevents

package all

import (
	_ "Dana/plugins/serializers/cloudevents" // register plugin
)
