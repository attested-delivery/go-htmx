package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate applies every pending goose migration under migrations/,
// embedded into the binary so a fresh deploy needs no separate migration
// step or file (AD-9). Migrations run against conn directly (not through
// DB.BeginWrite) since goose manages its own transactions per migration
// file; callers should pass a *sql.DB opened the same way DB.Write is —
// a single connection is enough and avoids goose racing itself.
func Migrate(conn *sql.DB) error {
	goose.SetBaseFS(migrationsFS)
	defer goose.SetBaseFS(nil)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("db: set goose dialect: %w", err)
	}
	if err := goose.Up(conn, "migrations"); err != nil {
		return fmt.Errorf("db: run migrations: %w", err)
	}
	return nil
}
