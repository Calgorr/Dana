//go:build !custom || (migrations && (outputs || outputs.riemann_legacy))

package all

import _ "Dana/migrations/outputs_riemann_legacy" // register migration
