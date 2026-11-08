package db

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
)

// legacyProject is one row of the database the former hand-rolled runner left.
type legacyProject struct {
	name        string
	folder      string
	displayName string
	group       string
	files       []string
	profiles    []string
}

// legacyDB creates a database the way the old hand-rolled runner did: the
// schema of the version 1, PRAGMA user_version = 1, no schema_migrations table,
// and the rows in legacy.
func legacyDB(t *testing.T, projects ...legacyProject) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "legacy.db")
	conn := rawDB(t, path)
	defer conn.Close()
	body, err := migrationFiles.ReadFile("migrations/0001_init.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, stmt := range []string{string(body), "PRAGMA user_version = 1"} {
		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range projects {
		if _, err := conn.ExecContext(ctx,
			`INSERT INTO projects (name, folder, display_name, group_name) VALUES (?, ?, NULLIF(?, ''), NULLIF(?, ''))`,
			p.name, p.folder, p.displayName, p.group); err != nil {
			t.Fatal(err)
		}
		for i, file := range p.files {
			if _, err := conn.ExecContext(ctx,
				`INSERT INTO project_files (project_name, position, path) VALUES (?, ?, ?)`,
				p.name, i, file); err != nil {
				t.Fatal(err)
			}
		}
		for _, profile := range p.profiles {
			if _, err := conn.ExecContext(ctx,
				`INSERT INTO active_profiles (project_name, profile) VALUES (?, ?)`,
				p.name, profile); err != nil {
				t.Fatal(err)
			}
		}
	}
	return path
}

// rawDB opens the database file with the pragmas Open uses, without migrating.
func rawDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	conn, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatal(err)
	}
	conn.SetMaxOpenConns(1)
	return conn
}

// migratorFor builds the same migrator Open uses over an open connection.
func migratorFor(t *testing.T, conn *sql.DB) *migrate.Migrate {
	t.Helper()
	m, err := newMigrator(conn)
	if err != nil {
		t.Fatalf("newMigrator: %v", err)
	}
	return m
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

// groupNames reads the saved Groups in their order.
func groupNames(t *testing.T, conn *sql.DB) []string {
	t.Helper()
	rows, err := conn.QueryContext(context.Background(), "SELECT name FROM groups ORDER BY position")
	if err != nil {
		t.Fatalf("reading groups: %v", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// projectGroup reads the Group name a Project carries after the migration.
func projectGroup(t *testing.T, conn *sql.DB, name string) string {
	t.Helper()
	var group string
	if err := conn.QueryRowContext(context.Background(),
		`SELECT COALESCE(g.name, '') FROM projects p LEFT JOIN groups g ON g.id = p.group_id WHERE p.name = ?`,
		name).Scan(&group); err != nil {
		t.Fatalf("reading the Group of %s: %v", name, err)
	}
	return group
}

func TestOpen_FreshDatabaseIsMigrated(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "fresh.db")
	conn, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if v, dirty := migrationVersion(t, conn); v != latestVersion || dirty {
		t.Fatalf("version = %d dirty = %v, want %d false", v, dirty, latestVersion)
	}
	if n := countProjects(t, conn); n != 0 {
		t.Fatalf("projects = %d, want 0", n)
	}
	if has, err := columnExists(context.Background(), conn, "projects", "group_name"); err != nil || has {
		t.Fatalf("projects.group_name exists = %t (%v), want the old column gone", has, err)
	}
	if got := groupNames(t, conn); len(got) != 0 {
		t.Errorf("groups = %v, want none", got)
	}
}

func TestOpen_AdoptsLegacyDatabaseWithoutDataLoss(t *testing.T) {
	path := legacyDB(t,
		legacyProject{name: "alpha", folder: "/srv/alpha", displayName: "Alfa", group: "Lavoro",
			files: []string{"/srv/alpha/compose.yaml", "/srv/alpha/override.yaml"}, profiles: []string{"debug"}},
		legacyProject{name: "bravo", folder: "/srv/bravo", group: " lavoro "},
		legacyProject{name: "charlie", folder: "/srv/charlie", group: "Città"},
		legacyProject{name: "delta", folder: "/srv/delta", group: "CITTÀ"},
		legacyProject{name: "echo", folder: "/srv/echo", group: "\u00a0Straße\u00a0"},
		legacyProject{name: "foxtrot", folder: "/srv/foxtrot", group: "STRASSE"},
		legacyProject{name: "golf", folder: "/srv/golf", group: "Senza Gruppo"},
		legacyProject{name: "hotel", folder: "/srv/hotel", group: "   "},
		legacyProject{name: "india", folder: "/srv/india", group: "Area 51"},
	)
	conn, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if v, dirty := migrationVersion(t, conn); v != latestVersion || dirty {
		t.Fatalf("version = %d dirty = %v, want %d false", v, dirty, latestVersion)
	}
	if n := countProjects(t, conn); n != 9 {
		t.Fatalf("projects = %d, want 9", n)
	}

	// Every Group kept the Projects and the spelling of the first Project in
	// compose name order; the spaces around the name went, the inner space
	// stayed.
	for project, want := range map[string]string{
		"alpha":   "Lavoro",
		"bravo":   "Lavoro",
		"charlie": "Città",
		"delta":   "Città",
		"echo":    "Straße",
		"foxtrot": "STRASSE",
		"golf":    "",
		"hotel":   "",
		"india":   "Area 51",
	} {
		if got := projectGroup(t, conn, project); got != want {
			t.Errorf("%s Group = %q, want %q", project, got, want)
		}
	}

	// The initial order is alphabetical without regard to case; the simple
	// folding keeps Straße and STRASSE apart.
	want := []string{"Area 51", "Città", "Lavoro", "STRASSE", "Straße"}
	got := groupNames(t, conn)
	if len(got) != len(want) {
		t.Fatalf("Groups = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Groups = %v, want %v", got, want)
		}
	}

	// The Display name and the Active Profiles stay on their Project, and the
	// compose files stay in their order.
	var display string
	if err := conn.QueryRowContext(context.Background(),
		`SELECT COALESCE(display_name, '') FROM projects WHERE name = 'alpha'`).Scan(&display); err != nil {
		t.Fatal(err)
	}
	if display != "Alfa" {
		t.Errorf("display_name of alpha = %q, want Alfa", display)
	}
	rows, err := conn.QueryContext(context.Background(),
		`SELECT path FROM project_files WHERE project_name = 'alpha' ORDER BY position`)
	if err != nil {
		t.Fatal(err)
	}
	var files []string
	for rows.Next() {
		var file string
		if err := rows.Scan(&file); err != nil {
			t.Fatal(err)
		}
		files = append(files, file)
	}
	rows.Close()
	if len(files) != 2 || files[0] != "/srv/alpha/compose.yaml" || files[1] != "/srv/alpha/override.yaml" {
		t.Errorf("compose files = %v, want the two saved ones", files)
	}
	var profile string
	if err := conn.QueryRowContext(context.Background(),
		`SELECT profile FROM active_profiles WHERE project_name = 'alpha'`).Scan(&profile); err != nil {
		t.Fatal(err)
	}
	if profile != "debug" {
		t.Errorf("active profile = %q, want debug", profile)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}

	// A second open is a no-op: the Groups, their order and the links stay.
	conn, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if v, dirty := migrationVersion(t, conn); v != latestVersion || dirty {
		t.Fatalf("after reopen: version = %d dirty = %v, want %d false", v, dirty, latestVersion)
	}
	if n := countProjects(t, conn); n != 9 {
		t.Fatalf("after reopen: projects = %d, want 9", n)
	}
	again := groupNames(t, conn)
	if len(again) != len(want) {
		t.Fatalf("after reopen: Groups = %v, want %v", again, want)
	}
	for i := range want {
		if again[i] != want[i] {
			t.Fatalf("after reopen: Groups = %v, want %v", again, want)
		}
	}
	if got := projectGroup(t, conn, "bravo"); got != "Lavoro" {
		t.Errorf("after reopen: bravo Group = %q, want Lavoro", got)
	}
}

func TestOpen_UpgradesAManagedVersion1Database(t *testing.T) {
	path := filepath.Join(t.TempDir(), "managed.db")
	conn := rawDB(t, path)
	if err := migratorFor(t, conn).Migrate(1); err != nil {
		t.Fatalf("Migrate(1): %v", err)
	}
	if _, err := conn.ExecContext(context.Background(),
		`INSERT INTO projects (name, folder, group_name) VALUES ('app', '/srv/app', 'Lavoro')`); err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}

	conn, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if v, dirty := migrationVersion(t, conn); v != latestVersion || dirty {
		t.Fatalf("version = %d dirty = %v, want %d false", v, dirty, latestVersion)
	}
	if got := projectGroup(t, conn, "app"); got != "Lavoro" {
		t.Errorf("app Group = %q, want Lavoro", got)
	}
}

func TestOpen_CompletesAHalfFinishedLegacyAdoption(t *testing.T) {
	path := legacyDB(t, legacyProject{name: "legacy", folder: "/srv/legacy"})
	// A crash after the migration driver created schema_migrations but before
	// the version was recorded leaves an empty table behind.
	raw := rawDB(t, path)
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
	if v, dirty := migrationVersion(t, conn); v != latestVersion || dirty {
		t.Fatalf("version = %d dirty = %v, want %d false", v, dirty, latestVersion)
	}
	if n := countProjects(t, conn); n != 1 {
		t.Fatalf("legacy row lost: projects = %d, want 1", n)
	}
}

func TestChooseMigration_DecidesFromTheSchemaVersion(t *testing.T) {
	for _, tc := range []struct {
		version int
		want    migrationChoice
		wantErr bool
	}{
		{0, migrateToGroupsSchema, false},
		{1, migrateToGroupsSchema, false},
		{2, migrateGroupData, false},
		{3, migrationDone, false},
		{4, 0, true},
	} {
		got, err := chooseMigration(tc.version)
		if tc.wantErr {
			if err == nil {
				t.Errorf("chooseMigration(%d) = %d, want an error", tc.version, got)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Errorf("chooseMigration(%d) = %d, %v, want %d", tc.version, got, err, tc.want)
		}
	}
}

func TestMigrateGroupNames_IsIdempotentAtVersion2(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	m := migratorFor(t, conn)
	if err := m.Migrate(2); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("Migrate(2): %v", err)
	}
	ctx := context.Background()
	for _, p := range []struct{ name, group string }{
		{"bravo", " lavoro "},
		{"alpha", "Lavoro"},
		{"charlie", "Città"},
		{"delta", "CITTÀ"},
	} {
		if _, err := conn.ExecContext(ctx,
			`INSERT INTO projects (name, folder, group_name) VALUES (?, ?, ?)`,
			p.name, "/srv/"+p.name, p.group); err != nil {
			t.Fatal(err)
		}
	}

	if err := MigrateGroupNames(ctx, conn); err != nil {
		t.Fatalf("first run: %v", err)
	}
	first := groupNames(t, conn)
	groupsAfterFirst := map[string]string{}
	for _, project := range []string{"alpha", "bravo", "charlie", "delta"} {
		groupsAfterFirst[project] = projectGroup(t, conn, project)
	}

	if err := MigrateGroupNames(ctx, conn); err != nil {
		t.Fatalf("second run: %v", err)
	}
	second := groupNames(t, conn)
	if len(first) != len(second) {
		t.Fatalf("second run: Groups = %v, want %v", second, first)
	}
	for i := range first {
		if second[i] != first[i] {
			t.Fatalf("second run: Groups = %v, want %v", second, first)
		}
	}
	for project, want := range groupsAfterFirst {
		if got := projectGroup(t, conn, project); got != want {
			t.Errorf("second run: %s Group = %q, want %q", project, got, want)
		}
	}
	if want := []string{"Città", "Lavoro"}; len(first) != len(want) || first[0] != want[0] || first[1] != want[1] {
		t.Errorf("Groups = %v, want %v", first, want)
	}
}

func TestMigrateGroupNames_DoesNothingWithoutTheColumn(t *testing.T) {
	conn, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if err := MigrateGroupNames(context.Background(), conn); err != nil {
		t.Fatalf("MigrateGroupNames without group_name: %v", err)
	}
	if got := groupNames(t, conn); len(got) != 0 {
		t.Errorf("Groups = %v, want none", got)
	}
}

func TestOpen_Version2WithoutGroupNameReachesVersion3(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	// A database the provisional 0002 left at version 2 without the old column.
	if err := migratorFor(t, conn).Force(2); err != nil {
		t.Fatalf("Force(2): %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}

	conn, err = Open(path)
	if err != nil {
		t.Fatalf("open a version 2 without group_name: %v", err)
	}
	defer conn.Close()
	if v, dirty := migrationVersion(t, conn); v != latestVersion || dirty {
		t.Fatalf("version = %d dirty = %v, want %d false", v, dirty, latestVersion)
	}
}

func TestDown_MigratesFrom3To1KeepingTheData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ctx := context.Background()
	for _, stmt := range []string{
		`INSERT INTO groups (id, name, position) VALUES (1, 'Lavoro', 0), (2, 'Casa', 1), (3, 'Vuoto', 2)`,
		`INSERT INTO projects (name, folder, display_name, group_id) VALUES ('app', '/srv/app', 'App', 1)`,
		`INSERT INTO projects (name, folder, group_id) VALUES ('other', '/srv/other', 2)`,
		`INSERT INTO projects (name, folder) VALUES ('solo', '/srv/solo')`,
		`INSERT INTO project_files (project_name, position, path) VALUES ('app', 0, '/srv/app/compose.yaml')`,
		`INSERT INTO active_profiles (project_name, profile) VALUES ('app', 'debug')`,
	} {
		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			t.Fatal(err)
		}
	}

	if err := migratorFor(t, conn).Migrate(1); err != nil {
		t.Fatalf("Migrate(1): %v", err)
	}
	if v, dirty := migrationVersion(t, conn); v != 1 || dirty {
		t.Fatalf("version = %d dirty = %v, want 1 false", v, dirty)
	}
	if n := countProjects(t, conn); n != 3 {
		t.Fatalf("projects = %d, want 3", n)
	}
	if has, err := columnExists(ctx, conn, "projects", "group_name"); err != nil || !has {
		t.Fatalf("projects.group_name = %t (%v), want the column back", has, err)
	}
	var groupTables int
	if err := conn.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'groups'").Scan(&groupTables); err != nil {
		t.Fatal(err)
	}
	if groupTables != 0 {
		t.Errorf("the groups table still exists, want it gone")
	}
	for project, want := range map[string]string{"app": "Lavoro", "other": "Casa", "solo": ""} {
		var group sql.NullString
		if err := conn.QueryRowContext(ctx,
			`SELECT group_name FROM projects WHERE name = ?`, project).Scan(&group); err != nil {
			t.Fatal(err)
		}
		if group.String != want {
			t.Errorf("%s group_name = %q, want %q", project, group.String, want)
		}
	}
	var display string
	if err := conn.QueryRowContext(ctx,
		`SELECT COALESCE(display_name, '') FROM projects WHERE name = 'app'`).Scan(&display); err != nil {
		t.Fatal(err)
	}
	if display != "App" {
		t.Errorf("display_name = %q, want App", display)
	}
	var file, profile string
	if err := conn.QueryRowContext(ctx,
		`SELECT path FROM project_files WHERE project_name = 'app'`).Scan(&file); err != nil {
		t.Fatal(err)
	}
	if file != "/srv/app/compose.yaml" {
		t.Errorf("compose file = %q, want the saved one", file)
	}
	if err := conn.QueryRowContext(ctx,
		`SELECT profile FROM active_profiles WHERE project_name = 'app'`).Scan(&profile); err != nil {
		t.Fatal(err)
	}
	if profile != "debug" {
		t.Errorf("active profile = %q, want debug", profile)
	}
	// The empty Group had no Project to carry its name down.
	var leftover int
	if err := conn.QueryRowContext(ctx,
		`SELECT count(*) FROM projects WHERE group_name = 'Vuoto'`).Scan(&leftover); err != nil {
		t.Fatal(err)
	}
	if leftover != 0 {
		t.Errorf("%d Projects carry the empty Group, want none", leftover)
	}
}

func TestDown_MigratesFrom2To1KeepingTheGroupName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := migratorFor(t, conn).Migrate(2); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("Migrate(2): %v", err)
	}
	ctx := context.Background()
	if _, err := conn.ExecContext(ctx,
		`INSERT INTO projects (name, folder, group_name) VALUES ('app', '/srv/app', 'Lavoro')`); err != nil {
		t.Fatal(err)
	}

	if err := migratorFor(t, conn).Migrate(1); err != nil {
		t.Fatalf("Migrate(1): %v", err)
	}
	var group sql.NullString
	if err := conn.QueryRowContext(ctx,
		`SELECT group_name FROM projects WHERE name = 'app'`).Scan(&group); err != nil {
		t.Fatal(err)
	}
	if group.String != "Lavoro" {
		t.Errorf("group_name = %q, want Lavoro", group.String)
	}
}

func TestDown_AFailedMigrationLeavesDirtyAndForceClearsIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	conn, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	// A column that is already there makes the down of the 0003 fail.
	if _, err := conn.ExecContext(context.Background(),
		"ALTER TABLE projects ADD COLUMN group_name TEXT"); err != nil {
		t.Fatal(err)
	}

	m := migratorFor(t, conn)
	if err := m.Steps(-1); err == nil {
		t.Fatal("the down of the 0003 should fail")
	}
	if version, dirty, err := m.Version(); err != nil || !dirty || version != 2 {
		t.Fatalf("after the failure: version = %d dirty = %v err = %v, want 2 true", version, dirty, err)
	}

	// Force restores the recorded version and clears the flag, as the test of
	// the failure does before it goes on.
	if err := m.Force(latestVersion); err != nil {
		t.Fatalf("Force: %v", err)
	}
	if version, dirty, err := m.Version(); err != nil || dirty || version != latestVersion {
		t.Fatalf("after Force: version = %d dirty = %v err = %v, want %d false", version, dirty, err, latestVersion)
	}
}
