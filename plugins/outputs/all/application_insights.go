//go:build !custom || outputs || outputs.application_insights

package all

import _ "Dana/plugins/outputs/application_insights" // register plugin
