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

// latestVersion is the schema version this binary writes. A database on a
// greater version is newer than the binary and cannot be opened safely.
const latestVersion = 3

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
// hand-rolled runner (PRAGMA user_version) whose adoption by golang-migrate is
// not recorded yet: schema_migrations is missing or has no row (an adoption
// that crashed after the driver created the table, before Force). It is 0
// when the database is new or already managed by golang-migrate.
func legacyVersion(ctx context.Context, conn *sql.DB) (int, error) {
	var managed int
	err := conn.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?", migsqlite.DefaultMigrationsTable).Scan(&managed)
	if err != nil {
		return 0, fmt.Errorf("db: look for %s: %w", migsqlite.DefaultMigrationsTable, err)
	}
	if managed > 0 {
		var recorded int
		// The table name is a constant of the driver, not user input.
		query := "SELECT count(*) FROM " + migsqlite.DefaultMigrationsTable
		if err := conn.QueryRowContext(ctx, query).Scan(&recorded); err != nil {
			return 0, fmt.Errorf("db: read %s: %w", migsqlite.DefaultMigrationsTable, err)
		}
		if recorded > 0 {
			return 0, nil
		}
	}
	var current int
	if err := conn.QueryRowContext(ctx, "PRAGMA user_version").Scan(&current); err != nil {
		return 0, fmt.Errorf("db: read legacy schema version: %w", err)
	}
	return current, nil
}

// migrateUp applies the embedded migrations with golang-migrate. A database
// left by the former runner (user_version >= 1, no recorded schema_migrations row) is
// adopted at its version first, so its tables and rows are kept untouched.
//
// The schema changes in three versions: 0002 creates the Groups table, a Go
// step fills it from the names on the Projects, and 0003 drops the old column.
// The three steps run in one opening, and an opening on a newer database is an
// error.
func migrateUp(conn *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), migrationTimeout)
	defer cancel()

	legacy, err := legacyVersion(ctx, conn)
	if err != nil {
		return err
	}
	// m.Close is never called: it would close conn, which the caller owns.
	m, err := newMigrator(conn)
	if err != nil {
		return err
	}
	if legacy >= 1 {
		if _, err := chooseMigration(legacy); err != nil {
			return err
		}
		if err := m.Force(legacy); err != nil {
			return fmt.Errorf("db: adopt legacy schema version %d: %w", legacy, err)
		}
	}
	for {
		version, err := schemaVersion(m)
		if err != nil {
			return err
		}
		choice, err := chooseMigration(version)
		if err != nil {
			return err
		}
		switch choice {
		case migrateToGroupsSchema:
			if err := m.Migrate(2); err != nil {
				return fmt.Errorf("db: apply migrations to version 2: %w", err)
			}
		case migrateGroupData:
			if err := MigrateGroupNames(ctx, conn); err != nil {
				return fmt.Errorf("db: migrate the Group names: %w", err)
			}
			if err := applyLastMigration(ctx, m, conn); err != nil {
				return err
			}
		case migrationDone:
			return nil
		}
	}
}

// migrationChoice is what the opening does next for the schema version it read.
type migrationChoice int

const (
	// migrateToGroupsSchema brings versions 0 and 1 up to the 0002.
	migrateToGroupsSchema migrationChoice = iota
	// migrateGroupData runs the Go step of the Group names and then the 0003.
	migrateGroupData
	// migrationDone is version 3: nothing to do.
	migrationDone
)

// chooseMigration is the pure decision of the opening. The versions 0 and 1
// need the 0002, the 2 needs the data step and the 0003, and the 3 needs
// nothing. A greater version is an error.
func chooseMigration(version int) (migrationChoice, error) {
	switch {
	case version == 0 || version == 1:
		return migrateToGroupsSchema, nil
	case version == 2:
		return migrateGroupData, nil
	case version == latestVersion:
		return migrationDone, nil
	default:
		return 0, fmt.Errorf("db: schema version %d is not supported: the newest is %d", version, latestVersion)
	}
}

// schemaVersion reads the version recorded in schema_migrations. A database
// without a record is version 0. A dirty database is an error: a failed
// migration left the schema in an unknown state.
func schemaVersion(m *migrate.Migrate) (int, error) {
	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("db: read schema version: %w", err)
	}
	if dirty {
		return 0, fmt.Errorf("db: schema left dirty at version %d by a failed migration", version)
	}
	return int(version), nil
}

// applyLastMigration applies the 0003. A database that does not have the old
// column any more has nothing to drop, so it takes the version without running
// the SQL.
func applyLastMigration(ctx context.Context, m *migrate.Migrate, conn *sql.DB) error {
	has, err := columnExists(ctx, conn, "projects", "group_name")
	if err != nil {
		return err
	}
	if !has {
		if err := m.Force(latestVersion); err != nil {
			return fmt.Errorf("db: record schema version %d: %w", latestVersion, err)
		}
		return nil
	}
	if err := m.Migrate(latestVersion); err != nil {
		return fmt.Errorf("db: apply migration %d: %w", latestVersion, err)
	}
	return nil
}

// newMigrator builds the migrator over the embedded files and the connection.
func newMigrator(conn *sql.DB) (*migrate.Migrate, error) {
	src, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("db: migration source: %w", err)
	}
	driver, err := migsqlite.WithInstance(conn, &migsqlite.Config{DatabaseName: "docker-controller"})
	if err != nil {
		return nil, fmt.Errorf("db: migration driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "sqlite", driver)
	if err != nil {
		return nil, fmt.Errorf("db: migrator: %w", err)
	}
	return m, nil
}

// queryer is what a transaction and the database have in common, so a schema
// check works inside either.
type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// columnExists reports whether a table has a column. The table name is a
// constant of this package, never input, and the column is read from the
// result.
func columnExists(ctx context.Context, q queryer, table, column string) (bool, error) {
	rows, err := q.QueryContext(ctx, "PRAGMA table_info("+table+")")
	if err != nil {
		return false, fmt.Errorf("db: read the columns of %s: %w", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			cid     int
			name    string
			kind    string
			notNull int
			dflt    sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &kind, &notNull, &dflt, &pk); err != nil {
			return false, fmt.Errorf("db: read the columns of %s: %w", table, err)
		}
		if name == column {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("db: read the columns of %s: %w", table, err)
	}
	return false, nil
}
