package inputs_kafka_consumer_legacy

import (
	"github.com/influxdata/toml/ast"

	"Dana/migrations"
)

const msg = `
    This plugin cannot be migrated automatically and requires manual intervention!
`

// Migration function
func migrate(_ *ast.Table) ([]byte, string, error) {
	return nil, msg, nil
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginMigration("inputs.kafka_consumer_legacy", migrate)
}
