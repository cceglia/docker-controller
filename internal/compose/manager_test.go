package compose_test

import (
	"context"
	"errors"
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

// The fixtures in testdata are real output captured read-only from the
// projects on the developer machine:
//
//	docker compose ls -a --format json
//	docker compose -p <name> ps -a --format json
//	docker compose -p <name> -f <file> --project-directory <dir> config --profiles
//
// None of those projects defines a Profile, so config-profiles-none.txt is
// the real (empty) output. config-profiles.txt is derived: the real output of
// `config --profiles` for a minimal compose file whose services use the
// profiles `debug` and `tools` (see profiles_test.go).
func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

var (
	lsArgs       = []string{"compose", "ls", "-a", "--format", "json"}
	profilesArgs = func(project, file string) []string {
		return []string{"compose", "-p", project, "-f", file, "--project-directory", filepath.Dir(file), "config", "--profiles"}
	}
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
	for name, file := range map[string]string{
		"sunlinkdocker":      "/home/cristiano/Sunlink/Dockers/SunlinkDocker/docker-compose.yml",
		"stellantisvpn":      "/home/cristiano/Sunlink/Dockers/StellantisVpn/docker-compose.yml",
		"personalvpn":        "/home/cristiano/Personal/Dockers/PersonalVpn/docker-compose.yml",
		"sunlinkvpn":         "/home/cristiano/Personal/Dockers/SunlinkVpn/docker-compose.yml",
		"travelgrouplaravel": "/home/cristiano/Personal/Lavoro/TravelGroupLaravel/docker-compose.yml",
	} {
		f.Respond(fixture(t, "config-profiles-none.txt"), profilesArgs(name, file)...)
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

func service(t *testing.T, p compose.Project, name string) compose.Service {
	t.Helper()
	for _, s := range p.Services {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("service %q not in project %q: %+v", name, p.Name, p.Services)
	return compose.Service{}
}

func TestRefresh_ExposesContainersGroupedByServiceFromRealOutput(t *testing.T) {
	m := newManager(t, realDocker(t), tempDB(t))
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	s := m.Snapshot()

	sl := find(t, s, "sunlinkdocker")
	if len(sl.Services) != 14 {
		t.Fatalf("sunlinkdocker has %d services, want 14: %+v", len(sl.Services), sl.Services)
	}

	// Image, state, health and a published port, as Docker reported them.
	redis := service(t, sl, "redis")
	if len(redis.Containers) != 1 {
		t.Fatalf("redis containers = %+v", redis.Containers)
	}
	c := redis.Containers[0]
	want := compose.Container{
		Name: "sunlinkdocker-redis-1", Image: "redis:7-alpine", State: "running", Health: "healthy",
		// 0.0.0.0 and :: publish the same port: it is listed once.
		Ports: []compose.Port{{Published: 6379, Target: 6379, Protocol: "tcp"}},
	}
	if !reflect.DeepEqual(c, want) {
		t.Errorf("redis container = %+v, want %+v", c, want)
	}

	// Only published ports are listed: traefik also exposes an unpublished 80/tcp.
	traefik := service(t, sl, "traefik").Containers[0]
	if !reflect.DeepEqual(traefik.Ports, []compose.Port{{Published: 8080, Target: 8080, Protocol: "tcp"}}) {
		t.Errorf("traefik ports = %+v", traefik.Ports)
	}
	// A published port may map to a different container port.
	ts := service(t, sl, "timescale").Containers[0]
	if !reflect.DeepEqual(ts.Ports, []compose.Port{{Published: 5433, Target: 5432, Protocol: "tcp"}}) {
		t.Errorf("timescale ports = %+v", ts.Ports)
	}
	// A service without published ports has an empty (not nil) list.
	if cron := service(t, sl, "cron").Containers[0]; cron.Ports == nil || len(cron.Ports) != 0 {
		t.Errorf("cron ports = %#v, want empty non-nil", cron.Ports)
	}

	// A stopped container keeps its service.
	win := service(t, find(t, s, "personalvpn"), "personal-vpn").Containers[0]
	if win.State != "running" {
		t.Errorf("personal-vpn state = %q", win.State)
	}
	st := find(t, s, "stellantisvpn")
	if len(st.Services) != 1 || st.Services[0].Containers[0].State != "exited" {
		t.Errorf("stellantisvpn services = %+v", st.Services)
	}

	// A Project with no Container has no Services, as an empty list.
	rm := find(t, s, "sunlinkvpn")
	if rm.Services == nil || len(rm.Services) != 0 {
		t.Errorf("removed project services = %#v, want empty non-nil", rm.Services)
	}
}

func TestRefresh_ReplicatedServiceShowsAllItsContainers(t *testing.T) {
	// No captured project has replicas, so the second replica is derived from
	// the real `redis` line of ps-sunlinkdocker.json: only the container name and
	// the host port change (replicas of a published service use a port range).
	lines := strings.Split(strings.TrimSpace(string(fixture(t, "ps-sunlinkdocker.json"))), "\n")
	var replica string
	for _, l := range lines {
		if strings.Contains(l, `"Service":"redis"`) {
			replica = strings.NewReplacer("sunlinkdocker-redis-1", "sunlinkdocker-redis-2", "6379", "6380").Replace(l)
		}
	}
	if replica == "" || !strings.Contains(replica, "sunlinkdocker-redis-2") {
		t.Fatal("fixture has no redis line to derive a replica from")
	}
	f := realDocker(t)
	f.Respond([]byte(strings.Join(append(lines, replica), "\n")), psArgs("sunlinkdocker")...)

	m := newManager(t, f, tempDB(t))
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	p := find(t, m.Snapshot(), "sunlinkdocker")
	if len(p.Services) != 14 || p.Total != 15 {
		t.Errorf("%d services, %d containers; want 14 services, 15 containers", len(p.Services), p.Total)
	}
	var names []string
	for _, c := range service(t, p, "redis").Containers {
		names = append(names, c.Name)
	}
	if !reflect.DeepEqual(names, []string{"sunlinkdocker-redis-1", "sunlinkdocker-redis-2"}) {
		t.Errorf("redis containers = %v", names)
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

// lsOf builds a `compose ls -a --format json` answer for one Project.
func lsOf(name string, files ...string) []byte {
	return []byte(`[{"Name":"` + name + `","Status":"exited(1)","ConfigFiles":"` + strings.Join(files, ",") + `"}]`)
}

func TestRefresh_FlagsAProjectWhoseComposeFilesAreGone(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "docker-compose.yml")
	other := filepath.Join(dir, "docker-compose.override.yml")
	for _, f := range []string{file, other} {
		if err := os.WriteFile(f, []byte("services: {}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	f := composefake.New()
	f.Respond(lsOf("tmpproject", file, other), lsArgs...)
	f.Respond(nil, psArgs("tmpproject")...)
	m := newManager(t, f, tempDB(t))

	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if find(t, m.Snapshot(), "tmpproject").ComposeMissing {
		t.Fatal("files exist: not missing")
	}

	if err := os.Remove(other); err != nil { // only one of the two files
		t.Fatal(err)
	}
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !find(t, m.Snapshot(), "tmpproject").ComposeMissing {
		t.Error("a deleted compose file must be flagged")
	}

	// Still flagged once Docker forgets the Project (after a Down).
	f.Respond([]byte("[]"), lsArgs...)
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	p := find(t, m.Snapshot(), "tmpproject")
	if !p.ComposeMissing || p.Status != compose.StatusRemoved {
		t.Errorf("got missing=%v status=%s, want missing Rimosso", p.ComposeMissing, p.Status)
	}
}

func TestWatch_FirstSnapshotIsLoadingUntilTheFirstRefresh(t *testing.T) {
	for _, tc := range []struct {
		name      string
		docker    func() *composefake.Fake
		reachable bool
	}{
		{"docker reachable", func() *composefake.Fake { return realDocker(t) }, true},
		{"docker unreachable", func() *composefake.Fake {
			f := composefake.New()
			f.RespondErr(nil, errors.New("Cannot connect to the Docker daemon"), lsArgs...)
			return f
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newManager(t, tc.docker(), tempDB(t))
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ch := m.Watch(ctx)

			if first := <-ch; !first.Loading {
				t.Errorf("first snapshot must be Loading (nothing was checked yet): %+v", first)
			}
			if !m.Snapshot().Loading {
				t.Error("Snapshot() before the first Refresh must be Loading")
			}
			if err := m.Refresh(ctx); err != nil {
				t.Fatal(err)
			}
			s := <-ch
			if s.Loading || s.DockerReachable != tc.reachable {
				t.Errorf("after the first Refresh: Loading=%v DockerReachable=%v, want false %v", s.Loading, s.DockerReachable, tc.reachable)
			}
		})
	}
}
