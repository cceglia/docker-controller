// Package db opens the application's single SQLite file and applies the
// versioned migrations embedded in the binary.
package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migsqlite "github.com/golang-migrate/migrate/v4/database/sqlite" // modernc.org/sqlite based, no cgo
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite" // database/sql driver "sqlite"
)

// migrationTimeout bounds the schema checks done while opening.
const migrationTimeout = 30 * time.Second

//go:embed migrations/*.sql
var migrationFiles embed.FS

// DefaultPath is ~/.local/share/docker-controller/docker-controller.db.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("db: home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share", "docker-controller", "docker-controller.db"), nil
}

// Open opens (creating it and its directory when needed) the database at path
// and brings the schema up to date. The path is injectable so tests can use a
// temporary file.
func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("db: create directory: %w", err)
	}
	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("db: open %s: %w", path, err)
	}
	// One connection: the app is a single local process and this avoids
	// SQLITE_BUSY between concurrent goroutines.
	conn.SetMaxOpenConns(1)
	if err := migrateUp(conn); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

// legacyVersion is the schema version of a database created by the former
// hand-rolled runner (PRAGMA user_version) that has no schema_migrations table
// yet; 0 when the database is new or already managed by golang-migrate.
func legacyVersion(ctx context.Context, conn *sql.DB) (int, error) {
	var managed int
	err := conn.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?", migsqlite.DefaultMigrationsTable).Scan(&managed)
	if err != nil {
		return 0, fmt.Errorf("db: look for %s: %w", migsqlite.DefaultMigrationsTable, err)
	}
	if managed > 0 {
		return 0, nil
	}
	var current int
	if err := conn.QueryRowContext(ctx, "PRAGMA user_version").Scan(&current); err != nil {
		return 0, fmt.Errorf("db: read legacy schema version: %w", err)
	}
	return current, nil
}

// migrateUp applies the embedded migrations with golang-migrate. A database
// left by the former runner (user_version >= 1, no schema_migrations) is
// adopted at its version first, so its tables and rows are kept untouched.
func migrateUp(conn *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), migrationTimeout)
	defer cancel()

	legacy, err := legacyVersion(ctx, conn)
	if err != nil {
		return err
	}
	src, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("db: migration source: %w", err)
	}
	driver, err := migsqlite.WithInstance(conn, &migsqlite.Config{DatabaseName: "docker-controller"})
	if err != nil {
		return fmt.Errorf("db: migration driver: %w", err)
	}
	// m.Close is never called: it would close conn, which the caller owns.
	m, err := migrate.NewWithInstance("iofs", src, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("db: migrator: %w", err)
	}
	if legacy >= 1 {
		if err := m.Force(legacy); err != nil {
			return fmt.Errorf("db: adopt legacy schema version %d: %w", legacy, err)
		}
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("db: apply migrations: %w", err)
	}
	return nil
}
