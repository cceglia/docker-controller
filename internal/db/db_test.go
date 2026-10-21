package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

// legacyDB creates a database the way the old hand-rolled runner did: the
// schema, PRAGMA user_version = 1, no schema_migrations table, and one row.
func legacyDB(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "legacy.db")
	conn, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	body, err := migrationFiles.ReadFile("migrations/0001_init.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, stmt := range []string{
		string(body),
		"PRAGMA user_version = 1",
		"INSERT INTO projects (name, folder) VALUES ('legacy', '/srv/legacy')",
	} {
		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func migrationVersion(t *testing.T, conn *sql.DB) (version int, dirty bool) {
	t.Helper()
	err := conn.QueryRowContext(context.Background(), "SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty)
	if err != nil {
		t.Fatalf("reading schema_migrations: %v", err)
	}
	return version, dirty
}

func countProjects(t *testing.T, conn *sql.DB) int {
	t.Helper()
	var n int
	if err := conn.QueryRowContext(context.Background(), "SELECT count(*) FROM projects").Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestOpen_FreshDatabaseIsMigrated(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "fresh.db")
	conn, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if v, dirty := migrationVersion(t, conn); v != 1 || dirty {
		t.Fatalf("version = %d dirty = %v, want 1 false", v, dirty)
	}
	if n := countProjects(t, conn); n != 0 {
		t.Fatalf("projects = %d, want 0", n)
	}
}

func TestOpen_AdoptsLegacyDatabaseWithoutDataLoss(t *testing.T) {
	path := legacyDB(t)
	conn, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if v, dirty := migrationVersion(t, conn); v != 1 || dirty {
		t.Fatalf("version = %d dirty = %v, want 1 false", v, dirty)
	}
	if n := countProjects(t, conn); n != 1 {
		t.Fatalf("legacy row lost: projects = %d, want 1", n)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}

	// A second open is a no-op.
	conn, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if v, dirty := migrationVersion(t, conn); v != 1 || dirty {
		t.Fatalf("after reopen: version = %d dirty = %v, want 1 false", v, dirty)
	}
	if n := countProjects(t, conn); n != 1 {
		t.Fatalf("after reopen: projects = %d, want 1", n)
	}
}

func TestOpen_CompletesAHalfFinishedLegacyAdoption(t *testing.T) {
	path := legacyDB(t)
	// A crash after the migration driver created schema_migrations but before
	// the version was recorded leaves an empty table behind.
	raw, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.ExecContext(context.Background(), "CREATE TABLE schema_migrations (version uint64, dirty bool)"); err != nil {
		t.Fatal(err)
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}

	conn, err := Open(path)
	if err != nil {
		t.Fatalf("open after a half-finished adoption: %v", err)
	}
	defer conn.Close()
	if v, dirty := migrationVersion(t, conn); v != 1 || dirty {
		t.Fatalf("version = %d dirty = %v, want 1 false", v, dirty)
	}
	if n := countProjects(t, conn); n != 1 {
		t.Fatalf("legacy row lost: projects = %d, want 1", n)
	}
}
