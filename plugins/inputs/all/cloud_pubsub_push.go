//go:build !custom || inputs || inputs.cloud_pubsub_push

package all

import _ "Dana/plugins/inputs/cloud_pubsub_push" // register plugin
