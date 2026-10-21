package compose_test

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"dockercontroller/internal/compose"
)

// The logs fixtures are real output of
//
//	docker compose -p <name> -f <file> --profile '*' logs --no-color --timestamps --tail N
//
// captured read-only. The container output of a Project arrives on stdout;
// stderr only carries compose's own diagnostics.

// logArgs is what the Manager must run to follow the logs of a Project: the
// Project's selection, every Profile (so Containers of Profiles that are not
// active are still seen), the last 500 lines then the stream, always with
// timestamps (the UI decides whether to show them) and without colours.
func logArgs(t *testing.T, m *compose.Manager, project string, service ...string) []string {
	t.Helper()
	p := find(t, m.Snapshot(), project)
	args := []string{"compose", "-p", project}
	for _, f := range p.ComposeFiles {
		args = append(args, "-f", f)
	}
	args = append(args, "--project-directory", p.Folder, "--profile", "*",
		"logs", "--follow", "--tail", "500", "--timestamps", "--no-color")
	return append(args, service...)
}

// logCollector gathers the lines a Logs call delivers.
type logCollector struct {
	mu    sync.Mutex
	lines []compose.LogLine
}

func (c *logCollector) on(l compose.LogLine) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lines = append(c.lines, l)
}

func (c *logCollector) snapshot() []compose.LogLine {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]compose.LogLine(nil), c.lines...)
}

// waitLines waits until n lines have been delivered.
func (c *logCollector) waitLines(t *testing.T, n int) []compose.LogLine {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if got := c.snapshot(); len(got) >= n {
			return got
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("got %d lines, want %d: %+v", len(c.snapshot()), n, c.snapshot())
	return nil
}

// startLogs runs Logs in the background and returns its result channel.
func startLogs(ctx context.Context, m *compose.Manager, req compose.LogRequest, c *logCollector) <-chan error {
	done := make(chan error, 1)
	go func() { done <- m.Logs(ctx, req, c.on) }()
	return done
}

func waitDone(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("Logs did not return")
		return nil
	}
}

func TestLogs_RunsFollowWithLast500LinesAndTheProjectSelection(t *testing.T) {
	f := realDocker(t)
	m, _ := activityManager(t, f)
	args := logArgs(t, m, "sunlinkdocker")
	f.RespondStream(nil, nil, nil, args...)

	c := &logCollector{}
	if err := waitDone(t, startLogs(context.Background(), m, compose.LogRequest{Project: "sunlinkdocker"}, c)); err != nil {
		t.Fatal(err)
	}

	var got []string
	for _, call := range f.Calls() {
		if strings.Contains(strings.Join(call, " "), " logs ") {
			got = call
		}
	}
	if strings.Join(got, " ") != strings.Join(args, " ") {
		t.Errorf("ran %v, want %v", got, args)
	}
}

func TestLogs_PassesTheServiceFilterToCompose(t *testing.T) {
	f := realDocker(t)
	m, _ := activityManager(t, f)
	args := logArgs(t, m, "sunlinkdocker", "db")
	f.RespondStream(nil, nil, nil, args...)

	c := &logCollector{}
	if err := waitDone(t, startLogs(context.Background(), m, compose.LogRequest{Project: "sunlinkdocker", Service: "db"}, c)); err != nil {
		t.Fatal(err)
	}

	last := f.Calls()[len(f.Calls())-1]
	if strings.Join(last, " ") != strings.Join(args, " ") {
		t.Errorf("ran %v, want %v", last, args)
	}
}

func TestLogs_DeliversLinesAttributedToTheirServiceWithTimestamp(t *testing.T) {
	f := realDocker(t)
	m, _ := activityManager(t, f)
	f.RespondStream(fixture(t, "logs-sunlinkdocker.txt"), nil, nil, logArgs(t, m, "sunlinkdocker")...)

	c := &logCollector{}
	if err := waitDone(t, startLogs(context.Background(), m, compose.LogRequest{Project: "sunlinkdocker"}, c)); err != nil {
		t.Fatal(err)
	}

	got := c.snapshot()
	if len(got) != 24 {
		t.Fatalf("got %d lines, want 24", len(got))
	}
	// "buffer-1" is the replica of Service "buffer"; the Container is named
	// "sunlinkdocker-buffer-1".
	want := compose.LogLine{
		Service:   "buffer",
		Timestamp: "2026-10-21T12:17:37.376631333Z",
		Text:      "Server listening on 0.0.0.0 port 22.",
	}
	if got[4] != want {
		t.Errorf("line 4 = %+v, want %+v", got[4], want)
	}
	// Hyphenated Service names, with a prefix padded to the longest name.
	if got[17].Service != "concentrator" || got[19].Service != "nodejs-api" {
		t.Errorf("services = %q, %q", got[17].Service, got[19].Service)
	}
	for i, l := range got {
		if l.Service == "" || l.Timestamp == "" || l.Stderr {
			t.Errorf("line %d not attributed: %+v", i, l)
		}
	}
}

func TestLogs_KeepsEmptyLinesAndStripsTerminalColours(t *testing.T) {
	f := realDocker(t)
	m, _ := activityManager(t, f)
	f.RespondStream(fixture(t, "logs-sunlinkdocker.txt"), nil, nil, logArgs(t, m, "sunlinkdocker")...)

	c := &logCollector{}
	waitDone(t, startLogs(context.Background(), m, compose.LogRequest{Project: "sunlinkdocker"}, c))

	got := c.snapshot()
	if got[9].Service != "scheduler" || got[9].Text != "" {
		t.Errorf("empty line = %+v", got[9])
	}
	if want := "2026-10-21T12:17:40Z INF Starting provider *traefik.Provider"; got[10].Service != "traefik" || got[10].Text != want {
		t.Errorf("coloured line = %+v, want text %q", got[10], want)
	}
}

func TestLogs_AttributesAServiceWhoseContainerKeepsTheServiceName(t *testing.T) {
	f := realDocker(t)
	m, _ := activityManager(t, f)
	f.RespondStream(fixture(t, "logs-personalvpn.txt"), nil, nil, logArgs(t, m, "personalvpn")...)

	c := &logCollector{}
	waitDone(t, startLogs(context.Background(), m, compose.LogRequest{Project: "personalvpn"}, c))

	got := c.snapshot()
	if len(got) != 3 || got[0].Service != "personal-vpn" {
		t.Errorf("lines = %+v", got)
	}
}

func TestLogs_ComposeDiagnosticsHaveNoServiceAndAServiceNoLongerShownKeepsItsName(t *testing.T) {
	f := realDocker(t)
	m, _ := activityManager(t, f)
	out := "ghost-1  | 2026-10-21T12:00:00.000000001Z still here\n" +
		"no prefix at all\n"
	f.RespondStream([]byte(out), []byte("no such service: nosuch"), nil, logArgs(t, m, "sunlinkdocker")...)

	c := &logCollector{}
	waitDone(t, startLogs(context.Background(), m, compose.LogRequest{Project: "sunlinkdocker"}, c))

	want := []compose.LogLine{
		{Service: "ghost", Timestamp: "2026-10-21T12:00:00.000000001Z", Text: "still here"},
		{Text: "no prefix at all"},
		{Text: "no such service: nosuch", Stderr: true},
	}
	got := c.snapshot()
	if len(got) != len(want) {
		t.Fatalf("lines = %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestLogs_EndsOnCancellationWithoutError_AndLeaksNoGoroutine(t *testing.T) {
	f := realDocker(t)
	m, _ := activityManager(t, f)
	args := logArgs(t, m, "sunlinkdocker")
	f.RespondStream(fixture(t, "logs-sunlinkdocker.txt"), nil, nil, args...)
	f.Follow(args...)
	before := runtime.NumGoroutine() // after the database and the Manager are up

	ctx, cancel := context.WithCancel(context.Background())
	c := &logCollector{}
	done := startLogs(ctx, m, compose.LogRequest{Project: "sunlinkdocker"}, c)
	c.waitLines(t, 24)
	select {
	case err := <-done:
		t.Fatalf("Logs ended before the cancellation: %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	cancel()
	if err := waitDone(t, done); err != nil {
		t.Errorf("a cancelled stream is not an error, got %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if after := runtime.NumGoroutine(); after > before {
		t.Errorf("goroutines: %d before, %d after the stream ended", before, after)
	}
}

func TestLogs_ReturnsComposeFailureWhenTheStreamEndsOnItsOwn(t *testing.T) {
	f := realDocker(t)
	m, _ := activityManager(t, f)
	boom := errors.New("exit status 1")
	f.RespondStream(nil, []byte("no such service: nosuch"), boom, logArgs(t, m, "sunlinkdocker", "nosuch")...)

	err := m.Logs(context.Background(), compose.LogRequest{Project: "sunlinkdocker", Service: "nosuch"}, func(compose.LogLine) {})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want %v", err, boom)
	}
}

func TestLogs_RefusesAnUnknownProjectOrAServiceThatLooksLikeAnOption(t *testing.T) {
	f := realDocker(t)
	m, _ := activityManager(t, f)
	noop := func(compose.LogLine) {}

	err := m.Logs(context.Background(), compose.LogRequest{Project: "nope"}, noop)
	if !errors.Is(err, compose.ErrUnknownProject) {
		t.Errorf("unknown project: %v", err)
	}
	err = m.Logs(context.Background(), compose.LogRequest{Project: "sunlinkdocker", Service: "--since=1h"}, noop)
	if !errors.Is(err, compose.ErrInvalidService) {
		t.Errorf("option-like service: %v", err)
	}
	for _, call := range f.Calls() {
		if strings.Contains(strings.Join(call, " "), " logs ") {
			t.Errorf("a refused request still ran %v", call)
		}
	}
}
