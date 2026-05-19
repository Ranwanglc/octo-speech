package migration

import (
	"database/sql"
	"embed"

	migrate "github.com/rubenv/sql-migrate"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func Run(db *sql.DB) (int, error) {
	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationFS,
		Root:       "migrations",
	}
	return migrate.Exec(db, "mysql", migrations, migrate.Up)
}
