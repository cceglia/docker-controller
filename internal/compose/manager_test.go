package compose_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/composefake"
	"dockercontroller/internal/db"
)

// The fixtures in testdata are real output captured read-only from the
// projects on the developer machine:
//
//	docker compose ls -a --format json
//	docker compose -p <name> ps -a --format json
func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

var (
	lsArgs = []string{"compose", "ls", "-a", "--format", "json"}
	psArgs = func(project string) []string {
		return []string{"compose", "-p", project, "ps", "-a", "--format", "json"}
	}
)

// realDocker answers like the machine did when the fixtures were captured.
func realDocker(t *testing.T) *composefake.Fake {
	t.Helper()
	f := composefake.New()
	f.Respond(fixture(t, "ls.json"), lsArgs...)
	for _, p := range []string{"sunlinkdocker", "stellantisvpn", "personalvpn"} {
		f.Respond(fixture(t, "ps-"+p+".json"), psArgs(p)...)
	}
	// The remaining projects of the capture have no fixture of their own.
	for _, p := range []string{"sunlinkvpn", "travelgrouplaravel"} {
		f.Respond(nil, psArgs(p)...)
	}
	return f
}

func newManager(t *testing.T, docker compose.Docker, dbPath string) *compose.Manager {
	t.Helper()
	conn, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	m, err := compose.New(compose.Config{Docker: docker, DB: conn})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func tempDB(t *testing.T) string { return filepath.Join(t.TempDir(), "test.db") }

func find(t *testing.T, s compose.Snapshot, name string) compose.Project {
	t.Helper()
	for _, p := range s.Projects {
		if p.Name == name {
			return p
		}
	}
	t.Fatalf("project %q not in snapshot: %+v", name, s.Projects)
	return compose.Project{}
}

func TestRefresh_ComputesStatusAndCountsFromRealOutput(t *testing.T) {
	m := newManager(t, realDocker(t), tempDB(t))
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	s := m.Snapshot()

	if !s.DockerReachable {
		t.Fatal("docker should be reachable")
	}
	cases := []struct {
		name            string
		status          compose.Status
		running, total  int
		folder          string
		firstComposeEnd string
	}{
		{"sunlinkdocker", compose.StatusActive, 14, 14, "/home/cristiano/Sunlink/Dockers/SunlinkDocker", "SunlinkDocker/docker-compose.yml"},
		{"stellantisvpn", compose.StatusStopped, 0, 1, "/home/cristiano/Sunlink/Dockers/StellantisVpn", "StellantisVpn/docker-compose.yml"},
		{"personalvpn", compose.StatusActive, 1, 1, "/home/cristiano/Personal/Dockers/PersonalVpn", "PersonalVpn/docker-compose.yml"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := find(t, s, c.name)
			if p.Status != c.status || p.Running != c.running || p.Total != c.total {
				t.Errorf("got %s %d/%d, want %s %d/%d", p.Status, p.Running, p.Total, c.status, c.running, c.total)
			}
			if p.Folder != c.folder {
				t.Errorf("folder = %q, want %q", p.Folder, c.folder)
			}
			if len(p.ComposeFiles) != 1 || !strings.HasSuffix(p.ComposeFiles[0], c.firstComposeEnd) {
				t.Errorf("compose files = %v", p.ComposeFiles)
			}
		})
	}
}

func TestRefresh_PartialStatus(t *testing.T) {
	// Same real sunlinkdocker output, with one container turned off.
	lines := strings.Split(strings.TrimSpace(string(fixture(t, "ps-sunlinkdocker.json"))), "\n")
	lines[0] = strings.Replace(lines[0], `"State":"running"`, `"State":"exited"`, 1)
	if lines[0] == strings.Split(strings.TrimSpace(string(fixture(t, "ps-sunlinkdocker.json"))), "\n")[0] {
		t.Fatal("fixture has no running container to switch off")
	}
	f := realDocker(t)
	f.Respond([]byte(strings.Join(lines, "\n")), psArgs("sunlinkdocker")...)

	m := newManager(t, f, tempDB(t))
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	p := find(t, m.Snapshot(), "sunlinkdocker")
	if p.Status != compose.StatusPartial || p.Running != 13 || p.Total != 14 {
		t.Errorf("got %s %d/%d, want Parziale 13/14", p.Status, p.Running, p.Total)
	}
}

func TestRefresh_RemovedProjectStaysListedAcrossRestarts(t *testing.T) {
	path := tempDB(t)
	first := newManager(t, realDocker(t), path)
	if err := first.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}

	// Docker no longer knows any project (e.g. after a Down).
	empty := composefake.New()
	empty.Respond([]byte("[]"), lsArgs...)

	same := newManager(t, empty, path) // a new process on the same DB
	if err := same.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	s := same.Snapshot()
	if len(s.Projects) != 5 {
		t.Fatalf("want the 5 saved projects, got %d", len(s.Projects))
	}
	p := find(t, s, "sunlinkdocker")
	if p.Status != compose.StatusRemoved || p.Running != 0 || p.Total != 0 {
		t.Errorf("got %s %d/%d, want Rimosso 0/0", p.Status, p.Running, p.Total)
	}
	if p.Folder != "/home/cristiano/Sunlink/Dockers/SunlinkDocker" || len(p.ComposeFiles) != 1 {
		t.Errorf("saved folder/files lost: %+v", p)
	}
}

func TestRefresh_NoDuplicatesWhenDiscoveredAgain(t *testing.T) {
	path := tempDB(t)
	for i := 0; i < 2; i++ {
		m := newManager(t, realDocker(t), path)
		if err := m.Refresh(context.Background()); err != nil {
			t.Fatal(err)
		}
		if n := len(m.Snapshot().Projects); n != 5 {
			t.Fatalf("round %d: %d projects, want 5", i, n)
		}
	}
}

func TestRefresh_DockerUnreachable(t *testing.T) {
	f := composefake.New()
	f.RespondErr(nil, errors.New("Cannot connect to the Docker daemon"), lsArgs...)
	m := newManager(t, f, tempDB(t))

	if err := m.Refresh(context.Background()); err != nil {
		t.Fatalf("an unreachable docker is a state, not an error: %v", err)
	}
	if m.Snapshot().DockerReachable {
		t.Error("DockerReachable should be false")
	}
}

func TestWatch_NotifiesOnlyWhenSomethingChanges(t *testing.T) {
	f := realDocker(t)
	m := newManager(t, f, tempDB(t))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := m.Watch(ctx)

	recv := func(want bool) compose.Snapshot {
		t.Helper()
		select {
		case s := <-ch:
			if !want {
				t.Fatalf("unexpected notification: %+v", s)
			}
			return s
		case <-time.After(150 * time.Millisecond):
			if want {
				t.Fatal("expected a notification")
			}
		}
		return compose.Snapshot{}
	}

	recv(true) // the current (still empty) snapshot on subscription

	if err := m.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if s := recv(true); len(s.Projects) != 5 {
		t.Fatalf("got %d projects", len(s.Projects))
	}

	if err := m.Refresh(ctx); err != nil { // nothing changed
		t.Fatal(err)
	}
	recv(false)

	// A container stops: only then a new snapshot arrives.
	f.Respond(fixture(t, "ps-stellantisvpn.json"), psArgs("personalvpn")...)
	if err := m.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	recv(true)
}
