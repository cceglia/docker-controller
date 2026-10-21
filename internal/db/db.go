// Package db opens the application's single SQLite file and applies the
// versioned migrations embedded in the binary.
package db

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "modernc.org/sqlite" // database/sql driver "sqlite"
)

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
	if err := migrate(conn); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

type migration struct {
	version int
	sql     string
}

func loadMigrations() ([]migration, error) {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("db: read migrations: %w", err)
	}
	var out []migration
	for _, e := range entries {
		prefix, _, ok := strings.Cut(e.Name(), "_")
		if !ok {
			return nil, fmt.Errorf("db: migration %q has no version prefix", e.Name())
		}
		v, err := strconv.Atoi(prefix)
		if err != nil {
			return nil, fmt.Errorf("db: migration %q: bad version: %w", e.Name(), err)
		}
		body, err := migrationFiles.ReadFile("migrations/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("db: read migration %q: %w", e.Name(), err)
		}
		out = append(out, migration{version: v, sql: string(body)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })
	return out, nil
}

// migrate applies, in order and each in its own transaction, the migrations
// newer than PRAGMA user_version.
func migrate(conn *sql.DB) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}
	var current int
	if err := conn.QueryRow("PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("db: read schema version: %w", err)
	}
	for _, m := range migrations {
		if m.version <= current {
			continue
		}
		tx, err := conn.Begin()
		if err != nil {
			return fmt.Errorf("db: begin migration %d: %w", m.version, err)
		}
		if _, err := tx.Exec(m.sql); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("db: apply migration %d: %w", m.version, err)
		}
		// PRAGMA does not accept bound parameters.
		if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", m.version)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("db: record migration %d: %w", m.version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("db: commit migration %d: %w", m.version, err)
		}
	}
	return nil
}
