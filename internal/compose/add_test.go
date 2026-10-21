package compose_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/composefake"
	"dockercontroller/internal/db"
)

// writeCompose writes a compose file named file inside dir (created when
// needed) and returns its path.
func writeCompose(t *testing.T, dir, file, content string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, file)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestProposeName(t *testing.T) {
	tests := []struct {
		name    string
		folder  string
		content string
		want    string
	}{
		{"name in the file wins over the folder", "folder", "name: from-file\nservices:\n  web:\n    image: nginx\n", "from-file"},
		{"quoted name with a comment", "folder", "name: \"quoted-name\" # the project\nservices: {}\n", "quoted-name"},
		{"name after the services", "folder", "services:\n  web:\n    image: nginx\nname: late\n", "late"},
		{"a nested name is not the project name", "myapp", "services:\n  web:\n    name: nested\n    image: nginx\n", "myapp"},
		{"folder name when there is no name", "myapp", "services: {}\n", "myapp"},
		{"folder name is normalised like compose does", "My_App.v2 Test", "services: {}\n", "my_appv2test"},
		{"leading dashes and underscores are dropped", "_-app", "services: {}\n", "app"},
		{"name in the file is normalised too", "folder", "name: My.App\n", "myapp"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := writeCompose(t, filepath.Join(t.TempDir(), tt.folder), "compose.yaml", tt.content)
			got, err := compose.ProposeName(file)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("ProposeName = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProposeName_FailsForAFileThatCannotBeRead(t *testing.T) {
	if _, err := compose.ProposeName(filepath.Join(t.TempDir(), "missing.yaml")); !errors.Is(err, compose.ErrInvalidProject) {
		t.Errorf("err = %v, want ErrInvalidProject", err)
	}
}

// addManager is a Manager on dbPath over the captured Docker output, with its
// Activities recorded.
func addManager(t *testing.T, dbPath string) (*compose.Manager, *recorder, *composefake.Fake) {
	t.Helper()
	conn, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	rec := &recorder{}
	f := realDocker(t)
	m, err := compose.New(compose.Config{Docker: f, DB: conn, OnActivity: rec.on})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	return m, rec, f
}

func TestAddProject_AppearsRemovedWithItsFilesInOrder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "newapp")
	main := writeCompose(t, dir, "compose.yaml", "services: {}\n")
	override := writeCompose(t, dir, "compose.override.yaml", "services: {}\n")
	m, _, _ := addManager(t, tempDB(t))

	if err := m.AddProject(context.Background(), "newapp", []string{main, override}); err != nil {
		t.Fatal(err)
	}

	p := find(t, m.Snapshot(), "newapp")
	if p.Status != compose.StatusRemoved || p.Total != 0 {
		t.Errorf("status = %s with %d Containers, want Rimosso", p.Status, p.Total)
	}
	if p.Folder != dir {
		t.Errorf("folder = %q, want %q", p.Folder, dir)
	}
	if want := []string{main, override}; !reflect.DeepEqual(p.ComposeFiles, want) {
		t.Errorf("files = %v, want %v", p.ComposeFiles, want)
	}
	if p.ComposeMissing {
		t.Error("the files exist: ComposeMissing must be false")
	}
}

func TestAddProject_ActionsPassEveryFileInSavedOrderAndTheProjectDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "newapp")
	main := writeCompose(t, dir, "compose.yaml", "services: {}\n")
	override := writeCompose(t, dir, "compose.override.yaml", "services: {}\n")
	m, rec, f := addManager(t, tempDB(t))
	if err := m.AddProject(context.Background(), "newapp", []string{main, override}); err != nil {
		t.Fatal(err)
	}

	want := []string{"compose", "-p", "newapp", "-f", main, "-f", override, "--project-directory", dir, "up", "-d"}
	f.RespondStream([]byte("ok"), nil, nil, want...)
	id, err := m.Do(context.Background(), compose.Action{Project: "newapp", Kind: compose.ActionUp})
	if err != nil {
		t.Fatal(err)
	}
	if evs := rec.finished(t, id); evs[len(evs)-1].Error != "" {
		t.Errorf("Up failed: %s (Docker was not asked for %v)", evs[len(evs)-1].Error, want)
	}
}

func TestAddProject_RejectsADuplicateName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "newapp")
	main := writeCompose(t, dir, "compose.yaml", "services: {}\n")
	m, _, _ := addManager(t, tempDB(t))

	// A name Docker already reported.
	err := m.AddProject(context.Background(), "stellantisvpn", []string{main})
	if !errors.Is(err, compose.ErrDuplicateProject) {
		t.Fatalf("err = %v, want ErrDuplicateProject", err)
	}
	if err.Error() == compose.ErrDuplicateProject.Error() {
		t.Errorf("the message should say which name is taken: %v", err)
	}
	// A name added by hand before.
	if err := m.AddProject(context.Background(), "newapp", []string{main}); err != nil {
		t.Fatal(err)
	}
	if err := m.AddProject(context.Background(), "newapp", []string{main}); !errors.Is(err, compose.ErrDuplicateProject) {
		t.Errorf("second add: err = %v, want ErrDuplicateProject", err)
	}
	// The saved Project is untouched.
	if got := find(t, m.Snapshot(), "stellantisvpn").ComposeFiles; len(got) == 1 && got[0] == main {
		t.Error("the duplicate overwrote the saved Project")
	}
}

func TestAddProject_RefusesWhatComposeCouldNotUse(t *testing.T) {
	dir := t.TempDir()
	main := writeCompose(t, dir, "compose.yaml", "services: {}\n")
	tests := []struct {
		name  string
		pname string
		files []string
	}{
		{"no files", "app", nil},
		{"empty name", "", []string{main}},
		{"invalid name", "My App", []string{main}},
		{"relative path", "app", []string{"compose.yaml"}},
		{"missing file", "app", []string{filepath.Join(dir, "missing.yaml")}},
		{"a folder instead of a file", "app", []string{dir}},
		{"missing extra file", "app", []string{main, filepath.Join(dir, "missing.yaml")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _, _ := addManager(t, tempDB(t))
			n := len(m.Snapshot().Projects)
			if err := m.AddProject(context.Background(), tt.pname, tt.files); !errors.Is(err, compose.ErrInvalidProject) {
				t.Errorf("err = %v, want ErrInvalidProject", err)
			}
			if got := len(m.Snapshot().Projects); got != n {
				t.Errorf("%d Projects, want %d: nothing must be saved", got, n)
			}
		})
	}
}

func TestAddProject_SurvivesReopeningTheDatabase(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "newapp")
	main := writeCompose(t, dir, "compose.yaml", "services: {}\n")
	override := writeCompose(t, dir, "compose.override.yaml", "services: {}\n")
	path := tempDB(t)
	first, _, _ := addManager(t, path)
	if err := first.AddProject(context.Background(), "newapp", []string{main, override}); err != nil {
		t.Fatal(err)
	}

	second, _, _ := addManager(t, path)
	p := find(t, second.Snapshot(), "newapp")
	if want := []string{main, override}; !reflect.DeepEqual(p.ComposeFiles, want) {
		t.Errorf("files = %v, want %v", p.ComposeFiles, want)
	}
	if p.Folder != dir || p.Status != compose.StatusRemoved {
		t.Errorf("folder %q status %s, want %q Rimosso", p.Folder, p.Status, dir)
	}
}
