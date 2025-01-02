//go:build !custom || inputs || inputs.docker_log

package all

import _ "Dana/plugins/inputs/docker_log" // register plugin
