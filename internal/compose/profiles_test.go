package compose_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/composefake"
	"dockercontroller/internal/db"
)

// profilesProject saves a Project added by hand whose compose file defines the
// Profiles debug and tools, and returns its manager, the fake and the path of
// its compose file. No project on the developer machine uses Profiles, so
// testdata/config-profiles.txt is the real output of `config --profiles` for
// this minimal file:
//
//	services:
//	  web:  {image: nginx}
//	  dbg:  {image: busybox, profiles: [debug]}
//	  tool: {image: busybox, profiles: [tools, debug]}
func profilesProject(t *testing.T, dbPath string) (*compose.Manager, *composefake.Fake, string, *recorder) {
	t.Helper()
	dir := t.TempDir()
	file := filepath.Join(dir, "docker-compose.yml")
	if err := os.WriteFile(file, []byte("services:\n  web: {image: nginx}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	f := realDocker(t)
	f.Respond(nil, psArgs("withprofiles")...)
	f.Respond(fixture(t, "config-profiles.txt"), profilesArgs("withprofiles", file)...)
	conn, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	rec := &recorder{}
	m, err := compose.New(compose.Config{Docker: f, DB: conn, OnActivity: rec.on})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := m.AddProject(context.Background(), "withprofiles", []string{file}); err != nil {
		t.Fatal(err)
	}
	return m, f, file, rec
}

func configCalls(f *composefake.Fake) int {
	n := 0
	for _, c := range f.Calls() {
		if c[len(c)-1] == "--profiles" {
			n++
		}
	}
	return n
}

func TestProfiles_AvailableAreReadFromComposeConfigAndExposedInTheSnapshot(t *testing.T) {
	m, _, _, _ := profilesProject(t, tempDB(t))
	s := m.Snapshot()

	p := find(t, s, "withprofiles")
	if want := []string{"debug", "tools"}; !reflect.DeepEqual(p.Profiles, want) {
		t.Errorf("Profiles = %v, want %v", p.Profiles, want)
	}
	if p.ActiveProfiles == nil || len(p.ActiveProfiles) != 0 {
		t.Errorf("ActiveProfiles = %#v, want empty and not nil", p.ActiveProfiles)
	}
	// A real Project without Profiles: empty, not nil (the UI iterates it).
	if q := find(t, s, "sunlinkdocker"); q.Profiles == nil || len(q.Profiles) != 0 {
		t.Errorf("Profiles of sunlinkdocker = %#v, want empty and not nil", q.Profiles)
	}
}

func TestProfiles_AreReadAgainOnlyWhenTheComposeFilesChange(t *testing.T) {
	m, f, file, _ := profilesProject(t, tempDB(t))
	before := configCalls(f)

	// Full refreshes and Docker events must not run `config` again.
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := configCalls(f); got != before {
		t.Errorf("a refresh with unchanged files ran config %d more times", got-before)
	}

	// Editing a compose file may add or remove Profiles.
	later := time.Now().Add(time.Hour)
	if err := os.Chtimes(file, later, later); err != nil {
		t.Fatal(err)
	}
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := configCalls(f); got != before+1 {
		t.Errorf("config ran %d times after the file changed, want %d", got, before+1)
	}
}

func TestUpdateSettings_ActiveProfilesAreSavedAndSurviveReopeningTheDB(t *testing.T) {
	path := tempDB(t)
	m, _, file, _ := profilesProject(t, path)

	if err := m.UpdateSettings(context.Background(), "withprofiles", compose.Settings{Profiles: []string{" tools", "debug", "tools", ""}}); err != nil {
		t.Fatal(err)
	}
	if got, want := find(t, m.Snapshot(), "withprofiles").ActiveProfiles, []string{"debug", "tools"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ActiveProfiles = %v, want %v (trimmed, once each, sorted)", got, want)
	}

	// A field left out leaves the Profiles untouched.
	if err := m.UpdateSettings(context.Background(), "withprofiles", compose.Settings{Group: ptr("Lavoro")}); err != nil {
		t.Fatal(err)
	}
	if got := find(t, m.Snapshot(), "withprofiles").ActiveProfiles; len(got) != 2 {
		t.Errorf("ActiveProfiles = %v after a Group-only update", got)
	}

	// A new process on the same DB still has them (also after re-discovery).
	f2 := realDocker(t)
	f2.Respond(nil, psArgs("withprofiles")...)
	f2.Respond(fixture(t, "config-profiles.txt"), profilesArgs("withprofiles", file)...)
	again := newManager(t, f2, path)
	if err := again.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got, want := find(t, again.Snapshot(), "withprofiles").ActiveProfiles, []string{"debug", "tools"}; !reflect.DeepEqual(got, want) {
		t.Errorf("after reopening: %v, want %v", got, want)
	}

	// An empty list deactivates every Profile.
	if err := again.UpdateSettings(context.Background(), "withprofiles", compose.Settings{Profiles: []string{}}); err != nil {
		t.Fatal(err)
	}
	if got := find(t, again.Snapshot(), "withprofiles").ActiveProfiles; len(got) != 0 {
		t.Errorf("ActiveProfiles = %v, want none", got)
	}
}

func TestUpdateSettings_ChangingProfilesRunsNothingInDocker(t *testing.T) {
	m, f, _, _ := profilesProject(t, tempDB(t))
	calls := len(f.Calls())

	if err := m.UpdateSettings(context.Background(), "withprofiles", compose.Settings{Profiles: []string{"debug"}}); err != nil {
		t.Fatal(err)
	}
	if got := f.Calls(); len(got) != calls {
		t.Errorf("changing a Profile reached Docker: %v", got[calls:])
	}
	if p := find(t, m.Snapshot(), "withprofiles"); p.Activity != nil {
		t.Errorf("changing a Profile started an Activity: %+v", p.Activity)
	}
}

// profileArgs is what the Manager must run for withprofiles.
func profileArgs(t *testing.T, m *compose.Manager, profiles []string, verb ...string) []string {
	t.Helper()
	p := find(t, m.Snapshot(), "withprofiles")
	args := []string{"compose", "-p", "withprofiles"}
	for _, f := range p.ComposeFiles {
		args = append(args, "-f", f)
	}
	args = append(args, "--project-directory", p.Folder)
	for _, pr := range profiles {
		args = append(args, "--profile", pr)
	}
	return append(args, verb...)
}

func TestDo_ActiveProfilesArePassedToEveryActionButDown(t *testing.T) {
	cases := []struct {
		kind    compose.ActionKind
		service string
		verb    []string
	}{
		{compose.ActionUp, "", []string{"up", "-d"}},
		{compose.ActionStop, "", []string{"stop"}},
		{compose.ActionRestart, "", []string{"restart"}},
		{compose.ActionPull, "", []string{"pull"}},
		{compose.ActionBuild, "", []string{"up", "-d", "--build"}},
		{compose.ActionUp, "dbg", []string{"up", "-d", "dbg"}},
		{compose.ActionStart, "dbg", []string{"start", "dbg"}},
		{compose.ActionStop, "dbg", []string{"stop", "dbg"}},
		{compose.ActionRestart, "dbg", []string{"restart", "dbg"}},
		{compose.ActionPull, "dbg", []string{"pull", "dbg"}},
		{compose.ActionBuild, "dbg", []string{"up", "-d", "--build", "dbg"}},
	}
	for _, c := range cases {
		t.Run(string(c.kind)+"/"+c.service, func(t *testing.T) {
			m, f, _, rec := profilesProject(t, tempDB(t))
			if err := m.UpdateSettings(context.Background(), "withprofiles", compose.Settings{Profiles: []string{"debug", "tools"}}); err != nil {
				t.Fatal(err)
			}
			want := profileArgs(t, m, []string{"debug", "tools"}, c.verb...)
			f.RespondStream(nil, nil, nil, want...)

			id, err := m.Do(context.Background(), compose.Action{Project: "withprofiles", Service: c.service, Kind: c.kind})
			if err != nil {
				t.Fatal(err)
			}
			rec.finished(t, id)

			calls := actionCalls(f)
			if len(calls) != 1 || strings.Join(calls[0], " ") != strings.Join(want, " ") {
				t.Errorf("ran %v, want exactly %v", calls, want)
			}
		})
	}
}

func TestDo_WithoutActiveProfilesNoProfileFlagIsPassed(t *testing.T) {
	m, f, _, rec := profilesProject(t, tempDB(t))
	want := profileArgs(t, m, nil, "up", "-d")
	f.RespondStream(nil, nil, nil, want...)
	id, err := m.Do(context.Background(), compose.Action{Project: "withprofiles", Kind: compose.ActionUp})
	if err != nil {
		t.Fatal(err)
	}
	rec.finished(t, id)
	if calls := actionCalls(f); len(calls) != 1 || strings.Join(calls[0], " ") != strings.Join(want, " ") {
		t.Errorf("ran %v, want exactly %v", calls, want)
	}
}

func TestDo_DownAndDownWithVolumesActOnAllProfilesWhateverIsActive(t *testing.T) {
	cases := []struct {
		kind    compose.ActionKind
		verb    []string
		volumes bool
	}{
		{compose.ActionDown, []string{"down"}, false},
		{compose.ActionDownVolumes, []string{"down", "-v"}, true},
	}
	for _, c := range cases {
		t.Run(string(c.kind), func(t *testing.T) {
			m, f, _, rec := profilesProject(t, tempDB(t))
			if err := m.UpdateSettings(context.Background(), "withprofiles", compose.Settings{Profiles: []string{"debug"}}); err != nil {
				t.Fatal(err)
			}
			// compose understands --profile '*' as "every Profile": it also
			// covers Profiles added to the file since they were last read.
			want := profileArgs(t, m, []string{"*"}, c.verb...)
			f.RespondStream(nil, nil, nil, want...)

			id, err := m.Do(context.Background(), compose.Action{Project: "withprofiles", Kind: c.kind})
			if err != nil {
				t.Fatal(err)
			}
			rec.finished(t, id)

			calls := actionCalls(f)
			if len(calls) != 1 || strings.Join(calls[0], " ") != strings.Join(want, " ") {
				t.Fatalf("ran %v, want exactly %v", calls, want)
			}
			hasV := false
			for _, a := range calls[0] {
				hasV = hasV || a == "-v"
			}
			if hasV != c.volumes {
				t.Errorf("-v present = %v, want %v", hasV, c.volumes)
			}
		})
	}
}
