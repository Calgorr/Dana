//go:build !custom || (migrations && (inputs || inputs.cassandra))

package all

import _ "Dana/migrations/inputs_cassandra" // register migration
