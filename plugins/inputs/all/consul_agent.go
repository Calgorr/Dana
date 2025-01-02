//go:build !custom || inputs || inputs.consul_agent

package all

import _ "Dana/plugins/inputs/consul_agent" // register plugin
