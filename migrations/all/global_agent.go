//go:build !custom || migrations

package all

import _ "Dana/migrations/global_agent" // register migration
