package compose_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/composefake"
	"dockercontroller/internal/db"
)

// recorder collects the Activity events a Manager delivers.
type recorder struct {
	mu     sync.Mutex
	events []compose.ActivityEvent
}

func (r *recorder) on(e compose.ActivityEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
}

func (r *recorder) forActivity(id compose.ActivityID) []compose.ActivityEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []compose.ActivityEvent
	for _, e := range r.events {
		if e.ID == id {
			out = append(out, e)
		}
	}
	return out
}

// finished waits for the final event of the Activity and returns all its events.
func (r *recorder) finished(t *testing.T, id compose.ActivityID) []compose.ActivityEvent {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		evs := r.forActivity(id)
		if n := len(evs); n > 0 && evs[n-1].Type == compose.ActivityFinished {
			return evs
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("activity %s did not finish; events: %+v", id, r.forActivity(id))
	return nil
}

func activityManager(t *testing.T, docker compose.Docker) (*compose.Manager, *recorder) {
	t.Helper()
	conn, err := db.Open(tempDB(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	rec := &recorder{}
	m, err := compose.New(compose.Config{Docker: docker, DB: conn, OnActivity: rec.on})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	return m, rec
}

// composeArgs is what the Manager must run for a Project: the project name,
// its compose files in saved order and its folder as project directory, then
// the verb.
func composeArgs(t *testing.T, m *compose.Manager, project string, verb ...string) []string {
	t.Helper()
	p := find(t, m.Snapshot(), project)
	args := []string{"compose", "-p", project}
	for _, f := range p.ComposeFiles {
		args = append(args, "-f", f)
	}
	args = append(args, "--project-directory", p.Folder)
	return append(args, verb...)
}

func TestDo_RunsEachActionWithItsComposeArguments(t *testing.T) {
	cases := []struct {
		kind compose.ActionKind
		verb []string
	}{
		{compose.ActionUp, []string{"up", "-d"}},
		{compose.ActionBuild, []string{"up", "-d", "--build"}},
		{compose.ActionStop, []string{"stop"}},
		{compose.ActionRestart, []string{"restart"}},
		{compose.ActionPull, []string{"pull"}},
	}
	for _, c := range cases {
		t.Run(string(c.kind), func(t *testing.T) {
			f := realDocker(t)
			m, rec := activityManager(t, f)
			args := composeArgs(t, m, "stellantisvpn", c.verb...)
			f.RespondStream([]byte("ok"), nil, nil, args...)

			id, err := m.Do(context.Background(), compose.Action{Project: "stellantisvpn", Kind: c.kind})
			if err != nil {
				t.Fatal(err)
			}
			rec.finished(t, id)

			calls := f.Calls()
			var got []string
			for _, call := range calls {
				if strings.Contains(strings.Join(call, " "), c.verb[0]) && call[len(call)-1] != "json" {
					got = call
				}
			}
			if strings.Join(got, " ") != strings.Join(args, " ") {
				t.Errorf("ran %v, want %v", got, args)
			}
		})
	}
}

func TestDo_DeliversOutputLineByLineThenOutcomeAndRefreshesSnapshot(t *testing.T) {
	f := realDocker(t)
	m, rec := activityManager(t, f)
	if st := find(t, m.Snapshot(), "stellantisvpn").Status; st != compose.StatusStopped {
		t.Fatalf("precondition: %s", st)
	}
	args := composeArgs(t, m, "stellantisvpn", "up", "-d")
	f.RespondStream([]byte("creating\nstarted"), []byte("progress on stderr"), nil, args...)
	// After the Up, Docker reports the container as running.
	running := strings.Replace(string(fixture(t, "ps-stellantisvpn.json")), `"State":"exited"`, `"State":"running"`, 1)
	f.Respond([]byte(running), psArgs("stellantisvpn")...)

	id, err := m.Do(context.Background(), compose.Action{Project: "stellantisvpn", Kind: compose.ActionUp})
	if err != nil {
		t.Fatal(err)
	}
	evs := rec.finished(t, id)

	var lines []string
	for _, e := range evs {
		if e.Type == compose.ActivityLine {
			lines = append(lines, e.Line)
		}
	}
	if strings.Join(lines, "|") != "creating|started|progress on stderr" {
		t.Errorf("lines = %v", lines)
	}
	if evs[0].Type != compose.ActivityStarted || evs[0].Project != "stellantisvpn" || evs[0].Kind != compose.ActionUp {
		t.Errorf("first event = %+v", evs[0])
	}
	if last := evs[len(evs)-1]; last.Error != "" {
		t.Errorf("success outcome carries error %q", last.Error)
	}

	p := find(t, m.Snapshot(), "stellantisvpn")
	if p.Status != compose.StatusActive || p.Activity != nil {
		t.Errorf("after the Activity: status %s activity %+v", p.Status, p.Activity)
	}
}

func TestDo_ErrorOutcomeCarriesComposeStderr(t *testing.T) {
	f := realDocker(t)
	m, rec := activityManager(t, f)
	args := composeArgs(t, m, "stellantisvpn", "up", "-d")
	f.RespondStream([]byte("creating"), []byte("Error response from daemon: port is already allocated"),
		errors.New("exit status 1"), args...)

	id, err := m.Do(context.Background(), compose.Action{Project: "stellantisvpn", Kind: compose.ActionUp})
	if err != nil {
		t.Fatal(err)
	}
	evs := rec.finished(t, id)
	if last := evs[len(evs)-1]; !strings.Contains(last.Error, "port is already allocated") {
		t.Errorf("outcome error = %q, want the compose stderr", last.Error)
	}
}

func TestDo_FailureSummaryPicksTheInformativeLine(t *testing.T) {
	cases := []struct {
		name   string
		stderr string
		want   string
	}{
		{
			name:   "real pull failure: progress lines then the daemon error",
			stderr: " Image registry.invalid/nonexistent:1 Pulling \n Image registry.invalid/nonexistent:1 Error \nError response from daemon: pull access denied for registry.invalid/nonexistent",
			want:   "Error response from daemon: pull access denied for registry.invalid/nonexistent",
		},
		{
			name:   "no error word: the last non-empty line",
			stderr: "first\nlast line\n\n",
			want:   "last line",
		},
		{
			name:   "long line is truncated",
			stderr: "Error " + strings.Repeat("x", 500),
			want:   "…",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := realDocker(t)
			m, rec := activityManager(t, f)
			args := composeArgs(t, m, "stellantisvpn", "pull")
			f.RespondStream(nil, []byte(c.stderr), errors.New("exit status 1"), args...)

			id, err := m.Do(context.Background(), compose.Action{Project: "stellantisvpn", Kind: compose.ActionPull})
			if err != nil {
				t.Fatal(err)
			}
			evs := rec.finished(t, id)
			last := evs[len(evs)-1]
			if c.want == "…" {
				if !strings.HasSuffix(last.Summary, "…") || len([]rune(last.Summary)) > 201 {
					t.Errorf("summary = %q, want truncated", last.Summary)
				}
			} else if last.Summary != c.want {
				t.Errorf("summary = %q, want %q", last.Summary, c.want)
			}
			if !strings.Contains(last.Error, strings.Split(strings.TrimSpace(c.stderr), "\n")[0]) {
				t.Errorf("full Error must keep every line, got %q", last.Error)
			}
		})
	}
}

func TestDo_SuccessHasNoSummary(t *testing.T) {
	f := realDocker(t)
	m, rec := activityManager(t, f)
	args := composeArgs(t, m, "stellantisvpn", "stop")
	f.RespondStream([]byte("ok"), nil, nil, args...)
	id, err := m.Do(context.Background(), compose.Action{Project: "stellantisvpn", Kind: compose.ActionStop})
	if err != nil {
		t.Fatal(err)
	}
	evs := rec.finished(t, id)
	if last := evs[len(evs)-1]; last.Summary != "" {
		t.Errorf("summary = %q", last.Summary)
	}
}

func TestDo_OneActivityPerProjectButProjectsInParallel(t *testing.T) {
	f := realDocker(t)
	m, rec := activityManager(t, f)
	a1 := composeArgs(t, m, "stellantisvpn", "stop")
	a2 := composeArgs(t, m, "personalvpn", "stop")
	release1 := f.Hold(a1...)
	release2 := f.Hold(a2...)
	f.RespondStream(nil, nil, nil, a1...)
	f.RespondStream(nil, nil, nil, a2...)

	id1, err := m.Do(context.Background(), compose.Action{Project: "stellantisvpn", Kind: compose.ActionStop})
	if err != nil {
		t.Fatal(err)
	}
	if act := find(t, m.Snapshot(), "stellantisvpn").Activity; act == nil || act.ID != id1 || act.Kind != compose.ActionStop {
		t.Errorf("running Activity not exposed in the snapshot: %+v", act)
	}

	if _, err := m.Do(context.Background(), compose.Action{Project: "stellantisvpn", Kind: compose.ActionUp}); !errors.Is(err, compose.ErrActivityRunning) {
		t.Errorf("second Activity on the same Project: err = %v, want ErrActivityRunning", err)
	}
	id2, err := m.Do(context.Background(), compose.Action{Project: "personalvpn", Kind: compose.ActionStop})
	if err != nil {
		t.Fatalf("a different Project must be allowed: %v", err)
	}

	release1()
	release2()
	rec.finished(t, id1)
	rec.finished(t, id2)

	// Once ended, the Project accepts a new Activity.
	f.RespondStream(nil, nil, nil, composeArgs(t, m, "stellantisvpn", "up", "-d")...)
	id3, err := m.Do(context.Background(), compose.Action{Project: "stellantisvpn", Kind: compose.ActionUp})
	if err != nil {
		t.Fatal(err)
	}
	rec.finished(t, id3)
}

func TestDo_RejectsUnknownProjectAndAction(t *testing.T) {
	m, _ := activityManager(t, realDocker(t))
	if _, err := m.Do(context.Background(), compose.Action{Project: "nope", Kind: compose.ActionUp}); !errors.Is(err, compose.ErrUnknownProject) {
		t.Errorf("err = %v, want ErrUnknownProject", err)
	}
	if _, err := m.Do(context.Background(), compose.Action{Project: "stellantisvpn", Kind: "explode"}); !errors.Is(err, compose.ErrUnknownAction) {
		t.Errorf("err = %v, want ErrUnknownAction", err)
	}
}

func TestDo_ServiceActionsPassTheServiceNameToCompose(t *testing.T) {
	cases := []struct {
		kind compose.ActionKind
		verb string
	}{
		{compose.ActionStart, "start"},
		{compose.ActionStop, "stop"},
		{compose.ActionRestart, "restart"},
	}
	for _, c := range cases {
		t.Run(string(c.kind), func(t *testing.T) {
			f := realDocker(t)
			m, rec := activityManager(t, f)
			args := composeArgs(t, m, "sunlinkdocker", c.verb, "redis")
			f.RespondStream([]byte("ok"), nil, nil, args...)

			id, err := m.Do(context.Background(), compose.Action{Project: "sunlinkdocker", Service: "redis", Kind: c.kind})
			if err != nil {
				t.Fatal(err)
			}
			rec.finished(t, id)

			var found bool
			for _, call := range f.Calls() {
				found = found || strings.Join(call, " ") == strings.Join(args, " ")
			}
			if !found {
				t.Errorf("never ran %v; calls: %v", args, f.Calls())
			}
		})
	}
}

func TestDo_ServiceActionIsAnActivityOfTheProject(t *testing.T) {
	f := realDocker(t)
	m, rec := activityManager(t, f)
	a1 := composeArgs(t, m, "sunlinkdocker", "restart", "redis")
	release := f.Hold(a1...)
	f.RespondStream(nil, nil, nil, a1...)

	id, err := m.Do(context.Background(), compose.Action{Project: "sunlinkdocker", Service: "redis", Kind: compose.ActionRestart})
	if err != nil {
		t.Fatal(err)
	}
	if act := find(t, m.Snapshot(), "sunlinkdocker").Activity; act == nil || act.ID != id || act.Service != "redis" {
		t.Errorf("running Activity = %+v, want it exposed with its Service", act)
	}

	// Another Service of the same Project, or the whole Project, must wait.
	for _, a := range []compose.Action{
		{Project: "sunlinkdocker", Service: "nats", Kind: compose.ActionStop},
		{Project: "sunlinkdocker", Kind: compose.ActionStop},
	} {
		if _, err := m.Do(context.Background(), a); !errors.Is(err, compose.ErrActivityRunning) {
			t.Errorf("%+v: err = %v, want ErrActivityRunning", a, err)
		}
	}
	release()
	rec.finished(t, id)
}

func TestDo_RejectsAServiceNameThatLooksLikeAnOption(t *testing.T) {
	m, _ := activityManager(t, realDocker(t))
	_, err := m.Do(context.Background(), compose.Action{Project: "sunlinkdocker", Service: "--all", Kind: compose.ActionStop})
	if !errors.Is(err, compose.ErrInvalidService) {
		t.Errorf("err = %v, want ErrInvalidService", err)
	}
}

func TestWatch_NotifiesRightAfterAnActivityEnds(t *testing.T) {
	f := realDocker(t)
	m, rec := activityManager(t, f)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := m.Watch(ctx)
	<-ch // current snapshot

	args := composeArgs(t, m, "stellantisvpn", "up", "-d")
	release := f.Hold(args...)
	f.RespondStream(nil, nil, nil, args...)
	running := strings.Replace(string(fixture(t, "ps-stellantisvpn.json")), `"State":"exited"`, `"State":"running"`, 1)
	f.Respond([]byte(running), psArgs("stellantisvpn")...)

	id, err := m.Do(context.Background(), compose.Action{Project: "stellantisvpn", Kind: compose.ActionUp})
	if err != nil {
		t.Fatal(err)
	}
	s := <-ch // the Activity started
	if find(t, s, "stellantisvpn").Activity == nil {
		t.Fatalf("the start must be notified with the running Activity: %+v", s)
	}
	release()
	select {
	case s = <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("no notification after the Activity ended")
	}
	p := find(t, s, "stellantisvpn")
	if p.Activity != nil || p.Status != compose.StatusActive {
		t.Errorf("notified snapshot: status %s activity %+v", p.Status, p.Activity)
	}
	rec.finished(t, id)
}

var _ = composefake.New

// actionCalls returns the calls that are not the discovery/refresh traffic of
// the Manager, i.e. the ones that ran an Action.
func actionCalls(f *composefake.Fake) [][]string {
	var out [][]string
	for _, call := range f.Calls() {
		if len(call) > 1 && call[1] == "ls" {
			continue
		}
		if call[len(call)-1] == "json" { // ps -a --format json
			continue
		}
		if call[len(call)-1] == "--profiles" { // config --profiles
			continue
		}
		out = append(out, call)
	}
	return out
}

func TestDo_DownRunsDownAndDownVolumesRunsDownDashV(t *testing.T) {
	cases := []struct {
		name string
		kind compose.ActionKind
		verb []string
	}{
		{"down", compose.ActionDown, []string{"down"}},
		{"down with volumes", compose.ActionDownVolumes, []string{"down", "-v"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := realDocker(t)
			m, rec := activityManager(t, f)
			args := composeArgs(t, m, "stellantisvpn", append([]string{"--profile", "*"}, c.verb...)...)
			f.RespondStream(nil, nil, nil, args...)

			id, err := m.Do(context.Background(), compose.Action{Project: "stellantisvpn", Kind: c.kind})
			if err != nil {
				t.Fatal(err)
			}
			rec.finished(t, id)

			calls := actionCalls(f)
			if len(calls) != 1 || strings.Join(calls[0], " ") != strings.Join(args, " ") {
				t.Errorf("ran %v, want exactly %v", calls, args)
			}
		})
	}
}

func TestDo_VolumesAreRemovedOnlyByDownWithVolumes(t *testing.T) {
	kinds := []compose.ActionKind{
		compose.ActionUp, compose.ActionStart, compose.ActionStop, compose.ActionRestart,
		compose.ActionPull, compose.ActionBuild, compose.ActionDown,
	}
	for _, kind := range kinds {
		for _, service := range []string{"", "redis"} {
			if service != "" && kind == compose.ActionDown {
				continue // Down is rejected for a Service, see below
			}
			t.Run(string(kind)+"/"+service, func(t *testing.T) {
				f := realDocker(t)
				m, rec := activityManager(t, f)
				// Accept whatever compose verb this kind maps to.
				p := find(t, m.Snapshot(), "sunlinkdocker")
				base := []string{"compose", "-p", "sunlinkdocker"}
				for _, file := range p.ComposeFiles {
					base = append(base, "-f", file)
				}
				for _, verb := range [][]string{{"up", "-d"}, {"start"}, {"stop"}, {"restart"}, {"pull"}, {"up", "-d", "--build"}, {"down"}} {
					args := append(append([]string(nil), base...), verb...)
					if service != "" {
						args = append(args, service)
					}
					f.RespondStream(nil, nil, nil, args...)
				}
				id, err := m.Do(context.Background(), compose.Action{Project: "sunlinkdocker", Service: service, Kind: kind})
				if err != nil {
					t.Fatal(err)
				}
				rec.finished(t, id)
				for _, call := range actionCalls(f) {
					for _, a := range call {
						if a == "-v" || a == "--volumes" {
							t.Errorf("%s must never remove volumes, ran %v", kind, call)
						}
					}
				}
			})
		}
	}
}

func TestDo_DownActionsAreOnlyForTheWholeProject(t *testing.T) {
	for _, kind := range []compose.ActionKind{compose.ActionDown, compose.ActionDownVolumes} {
		t.Run(string(kind), func(t *testing.T) {
			f := realDocker(t)
			m, _ := activityManager(t, f)
			_, err := m.Do(context.Background(), compose.Action{Project: "sunlinkdocker", Service: "redis", Kind: kind})
			if !errors.Is(err, compose.ErrProjectOnly) {
				t.Errorf("err = %v, want ErrProjectOnly", err)
			}
			if calls := actionCalls(f); len(calls) != 0 {
				t.Errorf("nothing must reach Docker, got %v", calls)
			}
			if find(t, m.Snapshot(), "sunlinkdocker").Activity != nil {
				t.Error("a rejected Action must not leave an Activity behind")
			}
		})
	}
}

func TestDo_ProjectStaysListedAsRemovedAfterADown(t *testing.T) {
	for _, kind := range []compose.ActionKind{compose.ActionDown, compose.ActionDownVolumes} {
		t.Run(string(kind), func(t *testing.T) {
			f := realDocker(t)
			m, rec := activityManager(t, f)
			verb := []string{"down"}
			if kind == compose.ActionDownVolumes {
				verb = append(verb, "-v")
			}
			f.RespondStream(nil, nil, nil, composeArgs(t, m, "personalvpn", append([]string{"--profile", "*"}, verb...)...)...)
			// From now on Docker no longer knows the Project, as after a real Down.
			f.Respond([]byte("[]"), lsArgs...)

			id, err := m.Do(context.Background(), compose.Action{Project: "personalvpn", Kind: kind})
			if err != nil {
				t.Fatal(err)
			}
			rec.finished(t, id)

			p := find(t, m.Snapshot(), "personalvpn")
			if p.Status != compose.StatusRemoved || p.Total != 0 || len(p.ComposeFiles) != 1 {
				t.Errorf("got %s %d/%d files=%v, want it kept as Rimosso", p.Status, p.Running, p.Total, p.ComposeFiles)
			}
		})
	}
}
