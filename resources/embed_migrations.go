//go:build !no_migrations

package resources

import "embed"

//go:embed migrations/*.sql
var DatabaseMigrations embed.FS
