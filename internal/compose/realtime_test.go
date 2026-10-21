package compose_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/composefake"
	"dockercontroller/internal/db"
)

// eventsArgs is what the Manager must ask Docker to follow container events.
var eventsArgs = []string{"events", "--filter", "type=container", "--format", "json"}

const (
	longInterval = time.Hour // an interval that never fires during a test
	fastRetry    = 5 * time.Millisecond
	waitFor      = 3 * time.Second
)

// startRun creates a Manager with the given intervals and runs it until the
// test ends. Run is observed through Watch only.
func startRun(t *testing.T, docker compose.Docker, refreshEvery, retryEvery time.Duration) (*compose.Manager, <-chan compose.Snapshot) {
	t.Helper()
	conn, err := db.Open(tempDB(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	m, err := compose.New(compose.Config{Docker: docker, DB: conn, RefreshEvery: refreshEvery, RetryEvery: retryEvery})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	ch := m.Watch(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.Run(ctx, nil)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(waitFor):
			t.Error("Run did not stop after the context was cancelled")
		}
	})
	return m, ch
}

// until returns the first snapshot satisfying ok.
func until(t *testing.T, ch <-chan compose.Snapshot, what string, ok func(compose.Snapshot) bool) compose.Snapshot {
	t.Helper()
	timeout := time.After(waitFor)
	for {
		select {
		case s := <-ch:
			if ok(s) {
				return s
			}
		case <-timeout:
			t.Fatalf("timed out waiting for: %s", what)
		}
	}
}

func count(f *composefake.Fake, args []string) int {
	n := 0
	for _, c := range f.Calls() {
		if strings.Join(c, " ") == strings.Join(args, " ") {
			n++
		}
	}
	return n
}

// containerEvent derives a lifecycle event of a project from a real captured
// event (testdata/events-noise.jsonl, a sunlinkdocker exec_die): only the
// action and the project identity are changed.
func containerEvent(t *testing.T, action, project string) string {
	t.Helper()
	line := strings.Split(strings.TrimSpace(string(fixture(t, "events-noise.jsonl"))), "\n")[0]
	line = strings.Replace(line, `"Action":"exec_die"`, `"Action":"`+action+`"`, 1)
	line = strings.Replace(line, `"com.docker.compose.project":"sunlinkdocker"`, `"com.docker.compose.project":"`+project+`"`, 1)
	if !strings.Contains(line, `"`+project+`"`) || !strings.Contains(line, `"Action":"`+action+`"`) {
		t.Fatal("fixture event could not be adapted")
	}
	return line
}

func TestRun_ContainerEventRefreshesOnlyTheAffectedProject(t *testing.T) {
	f := realDocker(t)
	f.Follow(eventsArgs...)
	release := f.Hold(eventsArgs...)
	f.RespondStream([]byte(containerEvent(t, "start", "stellantisvpn")), nil, nil, eventsArgs...)
	m, ch := startRun(t, f, longInterval, fastRetry)

	until(t, ch, "initial refresh", func(s compose.Snapshot) bool { return len(s.Projects) == 5 })
	if got := find(t, m.Snapshot(), "stellantisvpn").Status; got != compose.StatusStopped {
		t.Fatalf("stellantisvpn starts %s, want Fermo", got)
	}
	// The container comes up from a terminal; then its event arrives.
	f.Respond(fixture(t, "ps-personalvpn.json"), psArgs("stellantisvpn")...)
	lsBefore, psBefore := count(f, lsArgs), count(f, psArgs("sunlinkdocker"))
	release()

	s := until(t, ch, "stellantisvpn running", func(s compose.Snapshot) bool {
		return find(t, s, "stellantisvpn").Status == compose.StatusActive
	})
	if p := find(t, s, "stellantisvpn"); p.Running != 1 || p.Total != 1 {
		t.Errorf("got %d/%d, want 1/1", p.Running, p.Total)
	}
	if n := count(f, lsArgs); n != lsBefore {
		t.Errorf("a full refresh ran on a container event (%d ls calls, was %d)", n, lsBefore)
	}
	if n := count(f, psArgs("sunlinkdocker")); n != psBefore {
		t.Errorf("an unrelated project was refreshed (%d ps calls, was %d)", n, psBefore)
	}
}

func TestRun_EventsThatDoNotChangeAContainerAreIgnored(t *testing.T) {
	f := realDocker(t)
	f.Follow(eventsArgs...)
	release := f.Hold(eventsArgs...)
	f.RespondStream(fixture(t, "events-noise.jsonl"), nil, nil, eventsArgs...)
	_, ch := startRun(t, f, longInterval, fastRetry)

	until(t, ch, "initial refresh", func(s compose.Snapshot) bool { return len(s.Projects) == 5 })
	before := count(f, psArgs("sunlinkdocker"))
	release()
	time.Sleep(100 * time.Millisecond) // let the stream be consumed
	if n := count(f, psArgs("sunlinkdocker")); n != before {
		t.Errorf("exec events triggered %d refreshes", n-before)
	}
}

func TestRun_EventWithoutChangeDoesNotNotify(t *testing.T) {
	f := realDocker(t)
	f.Follow(eventsArgs...)
	release := f.Hold(eventsArgs...)
	f.RespondStream([]byte(containerEvent(t, "die", "personalvpn")), nil, nil, eventsArgs...)
	_, ch := startRun(t, f, longInterval, fastRetry)

	until(t, ch, "initial refresh", func(s compose.Snapshot) bool { return len(s.Projects) == 5 })
	before := count(f, psArgs("personalvpn"))
	release()
	deadline := time.After(waitFor)
	for count(f, psArgs("personalvpn")) == before { // the event was handled
		select {
		case <-deadline:
			t.Fatal("the event was never handled")
		case <-time.After(5 * time.Millisecond):
		}
	}
	select {
	case s := <-ch:
		t.Fatalf("unexpected notification for an unchanged snapshot: %+v", s)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestRun_FullRefreshIsTheSafetyNet(t *testing.T) {
	f := realDocker(t)
	f.Follow(eventsArgs...) // docker events stays silent
	_, ch := startRun(t, f, 20*time.Millisecond, fastRetry)

	until(t, ch, "initial refresh", func(s compose.Snapshot) bool { return len(s.Projects) == 5 })
	f.Respond(fixture(t, "ps-personalvpn.json"), psArgs("stellantisvpn")...) // no event announces it
	until(t, ch, "periodic refresh", func(s compose.Snapshot) bool {
		return find(t, s, "stellantisvpn").Status == compose.StatusActive
	})
}

func TestRun_RecoversByItselfWhenDockerComesBack(t *testing.T) {
	f := composefake.New()
	f.RespondErr(nil, errors.New("Cannot connect to the Docker daemon"), lsArgs...)
	f.Follow(eventsArgs...)
	release := f.Hold(eventsArgs...)
	f.RespondStream([]byte(containerEvent(t, "start", "stellantisvpn")), nil, nil, eventsArgs...)
	_, ch := startRun(t, f, longInterval, fastRetry)

	until(t, ch, "docker unreachable", func(s compose.Snapshot) bool { return !s.DockerReachable })
	deadline := time.After(waitFor)
	for count(f, lsArgs) < 3 { // it keeps retrying at the retry interval
		select {
		case <-deadline:
			t.Fatalf("only %d attempts while docker was down", count(f, lsArgs))
		case <-time.After(5 * time.Millisecond):
		}
	}

	// Docker is back, with the real projects; nobody tells the Manager.
	f.Respond(fixture(t, "ls.json"), lsArgs...)
	for _, p := range []string{"sunlinkdocker", "stellantisvpn", "personalvpn"} {
		f.Respond(fixture(t, "ps-"+p+".json"), psArgs(p)...)
	}
	for _, p := range []string{"sunlinkvpn", "travelgrouplaravel"} {
		f.Respond(nil, psArgs(p)...)
	}
	until(t, ch, "docker reachable again", func(s compose.Snapshot) bool {
		return s.DockerReachable && len(s.Projects) == 5
	})

	// The events stream is back too: an event is applied.
	f.Respond(fixture(t, "ps-personalvpn.json"), psArgs("stellantisvpn")...)
	release()
	until(t, ch, "event after recovery", func(s compose.Snapshot) bool {
		return find(t, s, "stellantisvpn").Status == compose.StatusActive
	})
}

func TestRun_ResumesWhenDockerGoesDownAndUpAgain(t *testing.T) {
	f := realDocker(t)
	f.Follow(eventsArgs...)
	_, ch := startRun(t, f, 20*time.Millisecond, fastRetry)
	until(t, ch, "initial refresh", func(s compose.Snapshot) bool { return s.DockerReachable && len(s.Projects) == 5 })

	f.RespondErr(nil, errors.New("Cannot connect to the Docker daemon"), lsArgs...)
	s := until(t, ch, "docker unreachable", func(s compose.Snapshot) bool { return !s.DockerReachable })
	if len(s.Projects) != 5 {
		t.Errorf("last known projects dropped: %d", len(s.Projects))
	}

	f.Respond(fixture(t, "ls.json"), lsArgs...)
	until(t, ch, "docker reachable again", func(s compose.Snapshot) bool { return s.DockerReachable })
}
