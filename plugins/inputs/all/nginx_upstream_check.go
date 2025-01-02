//go:build !custom || inputs || inputs.nginx_upstream_check

package all

import _ "Dana/plugins/inputs/nginx_upstream_check" // register plugin
