//go:build !custom || outputs || outputs.event_hubs

package all

import _ "Dana/plugins/outputs/event_hubs" // register plugin
