package store_test

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"dockercontroller/internal/compose/model"
	"dockercontroller/internal/compose/store"
	"dockercontroller/internal/db"
)

// open returns a Store on a fresh, migrated database in a temporary directory.
func open(t *testing.T) *store.Store {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return store.New(conn)
}

// list returns the saved Projects, failing the test on error.
func list(t *testing.T, s *store.Store) []model.Project {
	t.Helper()
	projects, err := s.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	return projects
}

func TestUpsertDiscovered_SavesFolderAndFiles(t *testing.T) {
	s := open(t)

	err := s.UpsertDiscovered(t.Context(), model.Discovered{
		Name:         "app",
		ComposeFiles: []string{"/srv/app/compose.yaml", "/srv/app/override.yaml"},
	})
	if err != nil {
		t.Fatalf("UpsertDiscovered: %v", err)
	}

	projects := list(t, s)
	if len(projects) != 1 {
		t.Fatalf("List returned %d projects, want 1", len(projects))
	}
	p := projects[0]
	if p.Name != "app" {
		t.Errorf("Name = %q, want %q", p.Name, "app")
	}
	if p.Folder != "/srv/app" {
		t.Errorf("Folder = %q, want %q", p.Folder, "/srv/app")
	}
	if want := []string{"/srv/app/compose.yaml", "/srv/app/override.yaml"}; !slices.Equal(p.ComposeFiles, want) {
		t.Errorf("ComposeFiles = %q, want %q", p.ComposeFiles, want)
	}
	if p.DisplayName != "" || p.Group != "" {
		t.Errorf("DisplayName, Group = %q, %q, want empty", p.DisplayName, p.Group)
	}
	if p.ActiveProfiles == nil || len(p.ActiveProfiles) != 0 {
		t.Errorf("ActiveProfiles = %v, want empty and not nil", p.ActiveProfiles)
	}
	if p.Profiles == nil || len(p.Profiles) != 0 {
		t.Errorf("Profiles = %v, want empty and not nil", p.Profiles)
	}
}

func TestUpsertDiscovered_RefreshesFilesAndKeepsUserSettings(t *testing.T) {
	s := open(t)
	if err := s.Add(t.Context(), model.Discovered{Name: "app", ComposeFiles: []string{"/old/compose.yaml"}}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	name, group, profiles := "Frontend", "Work", []string{"dev"}
	if _, err := s.UpdateSettings(t.Context(), "app", model.Settings{
		DisplayName: &name, Group: &group, Profiles: profiles,
	}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	if err := s.UpsertDiscovered(t.Context(), model.Discovered{
		Name:         "app",
		ComposeFiles: []string{"/new/compose.yaml", "/new/extra.yaml"},
	}); err != nil {
		t.Fatalf("UpsertDiscovered: %v", err)
	}

	p := list(t, s)[0]
	if p.Folder != "/new" {
		t.Errorf("Folder = %q, want %q", p.Folder, "/new")
	}
	if want := []string{"/new/compose.yaml", "/new/extra.yaml"}; !slices.Equal(p.ComposeFiles, want) {
		t.Errorf("ComposeFiles = %q, want %q", p.ComposeFiles, want)
	}
	if p.DisplayName != "Frontend" || p.Group != "Work" {
		t.Errorf("DisplayName, Group = %q, %q, want %q, %q", p.DisplayName, p.Group, "Frontend", "Work")
	}
	if want := []string{"dev"}; !slices.Equal(p.ActiveProfiles, want) {
		t.Errorf("ActiveProfiles = %q, want %q", p.ActiveProfiles, want)
	}
}

func TestAdd_SavesProject(t *testing.T) {
	s := open(t)

	err := s.Add(t.Context(), model.Discovered{
		Name:         "app",
		ComposeFiles: []string{"/srv/app/compose.yaml", "/srv/app/prod.yaml"},
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	projects := list(t, s)
	if len(projects) != 1 || projects[0].Name != "app" {
		t.Fatalf("List = %+v, want the saved project app", projects)
	}
	if want := []string{"/srv/app/compose.yaml", "/srv/app/prod.yaml"}; !slices.Equal(projects[0].ComposeFiles, want) {
		t.Errorf("ComposeFiles = %q, want %q", projects[0].ComposeFiles, want)
	}
}

func TestAdd_DuplicateNameIsRefused(t *testing.T) {
	s := open(t)
	if err := s.Add(t.Context(), model.Discovered{Name: "app", ComposeFiles: []string{"/first/compose.yaml"}}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	err := s.Add(t.Context(), model.Discovered{Name: "app", ComposeFiles: []string{"/second/compose.yaml"}})
	if !errors.Is(err, model.ErrDuplicateProject) {
		t.Fatalf("Add duplicate error = %v, want ErrDuplicateProject", err)
	}

	projects := list(t, s)
	if want := []string{"/first/compose.yaml"}; len(projects) != 1 || !slices.Equal(projects[0].ComposeFiles, want) {
		t.Errorf("List = %+v, want the first project unchanged", projects)
	}
}

func TestUpdateSettings_AppliesOnlyTheFieldsThatAreSet(t *testing.T) {
	s := open(t)
	if err := s.Add(t.Context(), model.Discovered{Name: "app", ComposeFiles: []string{"/srv/app/compose.yaml"}}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	name, group, blank := "Frontend", "Work", ""
	steps := []struct {
		name     string
		settings model.Settings
		display  string
		group    string
		profiles []string
	}{
		{
			name:     "display name only",
			settings: model.Settings{DisplayName: &name},
			display:  "Frontend",
			profiles: []string{},
		},
		{
			name:     "group only keeps the display name",
			settings: model.Settings{Group: &group},
			display:  "Frontend",
			group:    "Work",
			profiles: []string{},
		},
		{
			name:     "profiles only keep the rest",
			settings: model.Settings{Profiles: []string{"dev", "ops"}},
			display:  "Frontend",
			group:    "Work",
			profiles: []string{"dev", "ops"},
		},
		{
			name:     "an empty profile list deactivates them all",
			settings: model.Settings{Profiles: []string{}},
			display:  "Frontend",
			group:    "Work",
			profiles: []string{},
		},
		{
			name:     "an empty display name clears it",
			settings: model.Settings{DisplayName: &blank},
			display:  "",
			group:    "Work",
			profiles: []string{},
		},
	}
	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			found, err := s.UpdateSettings(t.Context(), "app", step.settings)
			if err != nil {
				t.Fatalf("UpdateSettings: %v", err)
			}
			if !found {
				t.Fatalf("UpdateSettings found = false, want true")
			}
			p := list(t, s)[0]
			if p.DisplayName != step.display {
				t.Errorf("DisplayName = %q, want %q", p.DisplayName, step.display)
			}
			if p.Group != step.group {
				t.Errorf("Group = %q, want %q", p.Group, step.group)
			}
			if !slices.Equal(p.ActiveProfiles, step.profiles) {
				t.Errorf("ActiveProfiles = %q, want %q", p.ActiveProfiles, step.profiles)
			}
		})
	}
}

func TestUpdateSettings_UnknownProjectIsReportedAsNotFound(t *testing.T) {
	s := open(t)
	name := "Ghost"

	found, err := s.UpdateSettings(t.Context(), "ghost", model.Settings{DisplayName: &name})
	if err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if found {
		t.Errorf("UpdateSettings found = true, want false")
	}
	if projects := list(t, s); len(projects) != 0 {
		t.Errorf("List = %+v, want no project", projects)
	}
}

func TestList_SortsNamesIgnoringCase(t *testing.T) {
	s := open(t)
	for _, name := range []string{"beta", "Alpha", "gamma"} {
		if err := s.Add(t.Context(), model.Discovered{Name: name, ComposeFiles: []string{"/srv/" + name + "/compose.yaml"}}); err != nil {
			t.Fatalf("Add %s: %v", name, err)
		}
	}

	var got []string
	for _, p := range list(t, s) {
		got = append(got, p.Name)
	}
	if want := []string{"Alpha", "beta", "gamma"}; !slices.Equal(got, want) {
		t.Errorf("names = %q, want %q", got, want)
	}
}

func TestRemove_DeletesProjectFilesAndProfiles(t *testing.T) {
	s := open(t)
	if err := s.Add(t.Context(), model.Discovered{Name: "app", ComposeFiles: []string{"/srv/app/compose.yaml"}}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := s.UpdateSettings(t.Context(), "app", model.Settings{Profiles: []string{"dev"}}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	if err := s.Remove(t.Context(), "app"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if projects := list(t, s); len(projects) != 0 {
		t.Fatalf("List = %+v, want no project", projects)
	}

	// Adding the name again must not resurrect the old files or profiles.
	if err := s.Add(t.Context(), model.Discovered{Name: "app", ComposeFiles: []string{"/other/compose.yaml"}}); err != nil {
		t.Fatalf("Add again: %v", err)
	}
	p := list(t, s)[0]
	if want := []string{"/other/compose.yaml"}; !slices.Equal(p.ComposeFiles, want) {
		t.Errorf("ComposeFiles = %q, want %q", p.ComposeFiles, want)
	}
	if len(p.ActiveProfiles) != 0 {
		t.Errorf("ActiveProfiles = %q, want empty", p.ActiveProfiles)
	}
}

func TestRemove_UnknownProjectChangesNothing(t *testing.T) {
	s := open(t)
	if err := s.Add(t.Context(), model.Discovered{Name: "app", ComposeFiles: []string{"/srv/app/compose.yaml"}}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	if err := s.Remove(t.Context(), "ghost"); err != nil {
		t.Fatalf("Remove unknown: %v", err)
	}
	if projects := list(t, s); len(projects) != 1 {
		t.Errorf("List = %+v, want the saved project", projects)
	}
}
