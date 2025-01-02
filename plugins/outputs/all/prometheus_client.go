//go:build !custom || outputs || outputs.prometheus_client

package all

import _ "Dana/plugins/outputs/prometheus_client" // register plugin
