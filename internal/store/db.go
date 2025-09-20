package store

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

// NewSQLite opens new SQLite database.
func NewSQLite(dbFilePath string, requiredVer uint) (*sql.DB, error) {
	// 1. Open the database
	db, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening DB: %w", err)
	}

	// 2. Limit number of connections due to SQLite doesn't support multiple connections
	db.SetMaxOpenConns(1)

	// 3. Perform database migration
	if err = migrateDB(db, requiredVer); err != nil {
		return nil, fmt.Errorf("migration error: %w", err)
	}

	return db, nil
}

// Migration scripts
//
//go:embed migration/*.sql
var migrationFS embed.FS

// migrateDB performs database migration.
func migrateDB(db *sql.DB, requiredVer uint) error {
	data, err := iofs.New(migrationFS, "migration")
	if err != nil {
		return fmt.Errorf("failed to load migration scripts: %w", err)
	}

	instance, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to initialize database driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", data, "sqlite", instance)
	if err != nil {
		return fmt.Errorf("failed to initialize migration: %w", err)
	}

	if err = m.Migrate(requiredVer); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to perform migration: %w", err)
	}

	ver, dirty, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get database version: %w", err)
	}

	switch {
	case dirty:
		return errors.New("database is dirty")
	case ver != requiredVer:
		return fmt.Errorf("unexpected database version: '%d', expected '%d'", ver, requiredVer)
	default:
		return nil
	}
}
