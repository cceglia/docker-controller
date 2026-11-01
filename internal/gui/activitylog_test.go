package gui

import (
	"fmt"
	"slices"
	"testing"

	"dockercontroller/internal/compose"
)

// started, output, finished and failed build the events the Manager sends.
func started(id compose.ActivityID, project, service string, kind compose.ActionKind) compose.ActivityEvent {
	return compose.ActivityEvent{ID: id, Type: compose.ActivityStarted, Project: project, Service: service, Kind: kind}
}

func output(id compose.ActivityID, project, service string, kind compose.ActionKind, text string, stderr bool) compose.ActivityEvent {
	return compose.ActivityEvent{ID: id, Type: compose.ActivityLine, Project: project, Service: service, Kind: kind, Line: text, Stderr: stderr}
}

func finished(id compose.ActivityID, project, service string, kind compose.ActionKind, err, summary string) compose.ActivityEvent {
	return compose.ActivityEvent{ID: id, Type: compose.ActivityFinished, Project: project, Service: service, Kind: kind, Error: err, Summary: summary}
}

func TestActivityHeader(t *testing.T) {
	for _, tc := range []struct {
		name     string
		activity Activity
		want     string
	}{
		{"running on the whole Project", Activity{State: ActivityRunning, Project: "shop", Kind: compose.ActionUp}, "⏳ Up · shop"},
		{"running on a Service", Activity{State: ActivityRunning, Project: "shop", Service: "db", Kind: compose.ActionStart}, "⏳ Start · shop/db"},
		{"finished ok", Activity{State: ActivityOK, Project: "shop", Kind: compose.ActionStop}, "✓ Stop · shop"},
		{"failed on a Service", Activity{State: ActivityFailed, Project: "shop", Service: "db", Kind: compose.ActionRestart}, "✗ Restart · shop/db"},
		{"down with volumes", Activity{State: ActivityRunning, Project: "shop", Kind: compose.ActionDownVolumes}, "⏳ Down con volumi · shop"},
		{"unknown kind keeps its name", Activity{State: ActivityRunning, Project: "shop", Kind: compose.ActionKind("inspect")}, "⏳ inspect · shop"},
		{"unknown state keeps the row", Activity{State: ActivityState("queued"), Project: "shop", Kind: compose.ActionUp}, "Up · shop"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ActivityHeader(tc.activity); got != tc.want {
				t.Errorf("ActivityHeader() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestActivityTarget(t *testing.T) {
	for _, tc := range []struct {
		name    string
		project string
		service string
		want    string
	}{
		{"the whole Project", "shop", "", "shop"},
		{"one Service of the Project", "shop", "db", "shop/db"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ActivityTarget(tc.project, tc.service); got != tc.want {
				t.Errorf("ActivityTarget(%q, %q) = %q, want %q", tc.project, tc.service, got, tc.want)
			}
		})
	}
}

func TestActivitySuccessText(t *testing.T) {
	for _, tc := range []struct {
		name     string
		activity Activity
		want     string
	}{
		{"on the whole Project", Activity{State: ActivityOK, Project: "shop", Kind: compose.ActionUp}, "Up su shop riuscito."},
		{"on a Service", Activity{State: ActivityOK, Project: "shop", Service: "db", Kind: compose.ActionStart}, "Start su shop/db riuscito."},
		{"unknown kind keeps its name", Activity{State: ActivityOK, Project: "shop", Kind: compose.ActionKind("inspect")}, "inspect su shop riuscito."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ActivitySuccessText(tc.activity); got != tc.want {
				t.Errorf("ActivitySuccessText() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestActivityToastText(t *testing.T) {
	for _, tc := range []struct {
		name     string
		activity Activity
		want     string
	}{
		{
			"with the summary of the backend",
			Activity{State: ActivityFailed, Project: "shop", Kind: compose.ActionUp, Summary: "no such image"},
			"Up su shop non riuscito: no such image",
		},
		{
			"without a summary",
			Activity{State: ActivityFailed, Project: "shop", Kind: compose.ActionUp},
			"Up su shop non riuscito",
		},
		{
			"on a Service",
			Activity{State: ActivityFailed, Project: "shop", Service: "db", Kind: compose.ActionDownService, Summary: "exit 1"},
			"Down su shop/db non riuscito: exit 1",
		},
		{
			"an unknown kind keeps its name",
			Activity{State: ActivityFailed, Project: "shop", Kind: compose.ActionKind("inspect"), Summary: "boom"},
			"inspect su shop non riuscito: boom",
		},
		{
			// The notice carries one line: the summary the backend chose, not
			// the whole stderr that the output of the Activity keeps.
			"the notice shows the summary and not the error",
			Activity{
				State:   ActivityFailed,
				Project: "shop",
				Kind:    compose.ActionBuild,
				Error:   "progress 1\nprogress 2\nerror: no such image\n",
				Summary: "error: no such image",
			},
			"Build su shop non riuscito: error: no such image",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ActivityToastText(tc.activity); got != tc.want {
				t.Errorf("ActivityToastText() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestActivityOutcome(t *testing.T) {
	for _, tc := range []struct {
		name     string
		activity Activity
		want     string
	}{
		{"a running Activity has none", Activity{State: ActivityRunning, Project: "shop", Kind: compose.ActionUp}, ""},
		{"a success names the Action", Activity{State: ActivityOK, Project: "shop", Kind: compose.ActionUp}, "Up su shop riuscito."},
		{"a failure shows the error", Activity{State: ActivityFailed, Project: "shop", Kind: compose.ActionUp, Error: "boom"}, "boom"},
		{"the error loses its last newline", Activity{State: ActivityFailed, Project: "shop", Kind: compose.ActionUp, Error: "a\nb\n"}, "a\nb"},
		{"a state the view does not know has none", Activity{State: ActivityState("queued"), Project: "shop", Kind: compose.ActionUp}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ActivityOutcome(tc.activity); got != tc.want {
				t.Errorf("ActivityOutcome() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestActivityOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		activity Activity
		want     string
	}{
		{
			"a running Activity shows its lines",
			Activity{State: ActivityRunning, Lines: []ActivityLine{{Text: "one"}, {Text: "two"}}},
			"one\ntwo\n",
		},
		{
			"a success ends with the outcome",
			Activity{State: ActivityOK, Project: "shop", Kind: compose.ActionUp, Lines: []ActivityLine{{Text: "one"}}},
			"one\nUp su shop riuscito.\n",
		},
		{
			"a failure ends with the error",
			Activity{State: ActivityFailed, Project: "shop", Kind: compose.ActionUp, Lines: []ActivityLine{{Text: "one", Stderr: true}}, Error: "boom\n"},
			"one\nboom\n",
		},
		{
			"a failure whose error already reached the output says it once",
			Activity{State: ActivityFailed, Project: "shop", Kind: compose.ActionUp, Lines: []ActivityLine{{Text: "Pulling web...", Stderr: true}, {Text: "error: no such image", Stderr: true}}, Error: "error: no such image\n"},
			"Pulling web...\nerror: no such image\n",
		},
		{
			"an Activity without output",
			Activity{State: ActivityRunning},
			"",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ActivityOutput(tc.activity); got != tc.want {
				t.Errorf("ActivityOutput() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestActivityGlyph(t *testing.T) {
	for _, tc := range []struct {
		name  string
		state ActivityState
		want  string
	}{
		{"running", ActivityRunning, "⏳"},
		{"ok", ActivityOK, "✓"},
		{"failed", ActivityFailed, "✗"},
		{"unknown", ActivityState("queued"), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ActivityGlyph(tc.state); got != tc.want {
				t.Errorf("ActivityGlyph(%q) = %q, want %q", tc.state, got, tc.want)
			}
		})
	}
}

func TestActivityLog_KeepsTheHistoryNewestFirst(t *testing.T) {
	log := NewActivityLog()
	log.Push(started("act-1", "shop", "", compose.ActionUp))
	log.Push(started("act-2", "blog", "db", compose.ActionStart))
	log.Push(output("act-2", "blog", "db", compose.ActionStart, "starting", false))
	log.Push(finished("act-2", "blog", "db", compose.ActionStart, "", ""))

	got := log.Activities()
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != "act-2" || got[1].ID != "act-1" {
		t.Errorf("order = [%q %q], want [act-2 act-1]", got[0].ID, got[1].ID)
	}
	if len(got[0].Lines) != 1 || got[0].Lines[0].Text != "starting" {
		t.Errorf("lines of act-2 = %v, want the one line of the output", got[0].Lines)
	}
	if got[0].State != ActivityOK {
		t.Errorf("state of act-2 = %q, want %q", got[0].State, ActivityOK)
	}
}

func TestActivityLog_State(t *testing.T) {
	for _, tc := range []struct {
		name   string
		events []compose.ActivityEvent
		want   ActivityState
	}{
		{
			"a new Activity runs",
			[]compose.ActivityEvent{started("act-1", "shop", "", compose.ActionUp)},
			ActivityRunning,
		},
		{
			"a finished Activity with no error worked",
			[]compose.ActivityEvent{
				started("act-1", "shop", "", compose.ActionUp),
				finished("act-1", "shop", "", compose.ActionUp, "", ""),
			},
			ActivityOK,
		},
		{
			"a finished Activity with an error failed",
			[]compose.ActivityEvent{
				started("act-1", "shop", "", compose.ActionUp),
				finished("act-1", "shop", "", compose.ActionUp, "boom", "boom"),
			},
			ActivityFailed,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			log := NewActivityLog()
			for _, e := range tc.events {
				log.Push(e)
			}
			got, ok := log.Activity("act-1")
			if !ok {
				t.Fatal("Activity(act-1) not found")
			}
			if got.State != tc.want {
				t.Errorf("state = %q, want %q", got.State, tc.want)
			}
		})
	}
}

func TestActivityLog_KeepsTheSummaryOfAFailure(t *testing.T) {
	log := NewActivityLog()
	log.Push(started("act-1", "shop", "", compose.ActionUp))
	log.Push(finished("act-1", "shop", "", compose.ActionUp, "error: no such image", "no such image"))

	got, ok := log.Activity("act-1")
	if !ok {
		t.Fatal("Activity(act-1) not found")
	}
	if got.Error != "error: no such image" {
		t.Errorf("Error = %q, want the error of the event", got.Error)
	}
	if got.Summary != "no such image" {
		t.Errorf("Summary = %q, want the summary of the event", got.Summary)
	}
}

func TestActivityLog_IgnoresAnUnknownActivity(t *testing.T) {
	log := NewActivityLog()
	log.Push(output("act-1", "shop", "", compose.ActionUp, "late", false))
	log.Push(finished("act-1", "shop", "", compose.ActionUp, "boom", "boom"))

	if got := log.Activities(); len(got) != 0 {
		t.Errorf("Activities() = %v, want none", got)
	}
	if _, ok := log.Activity("act-1"); ok {
		t.Error("Activity(act-1) found, want none")
	}
}

func TestActivityLog_KeepsTheLastLines(t *testing.T) {
	log := NewActivityLog()
	log.Push(started("act-1", "shop", "", compose.ActionUp))
	for i := 0; i <= ActivityLineLimit; i++ {
		log.Push(output("act-1", "shop", "", compose.ActionUp, fmt.Sprintf("line %d", i), false))
	}

	got, ok := log.Activity("act-1")
	if !ok {
		t.Fatal("Activity(act-1) not found")
	}
	if len(got.Lines) != ActivityLineLimit {
		t.Fatalf("len(Lines) = %d, want %d", len(got.Lines), ActivityLineLimit)
	}
	if got.Lines[0].Text != "line 1" {
		t.Errorf("first line = %q, want %q", got.Lines[0].Text, "line 1")
	}
	if last := got.Lines[len(got.Lines)-1].Text; last != fmt.Sprintf("line %d", ActivityLineLimit) {
		t.Errorf("last line = %q, want %q", last, fmt.Sprintf("line %d", ActivityLineLimit))
	}
	if got.Dropped != 1 {
		t.Errorf("Dropped = %d, want 1", got.Dropped)
	}
}

func TestActivityLog_NotifiesAfterAChange(t *testing.T) {
	log := NewActivityLog()
	calls := 0
	log.OnChange(func() { calls++ })

	log.Push(started("act-1", "shop", "", compose.ActionUp))
	if calls != 1 {
		t.Fatalf("calls after started = %d, want 1", calls)
	}
	log.Push(output("act-1", "shop", "", compose.ActionUp, "one", false))
	if calls != 2 {
		t.Fatalf("calls after a line = %d, want 2", calls)
	}
	log.Push(output("act-unknown", "shop", "", compose.ActionUp, "lost", false))
	if calls != 2 {
		t.Fatalf("calls after an ignored line = %d, want 2", calls)
	}
	log.Push(finished("act-1", "shop", "", compose.ActionUp, "", ""))
	if calls != 3 {
		t.Fatalf("calls after finished = %d, want 3", calls)
	}
}

func TestActivityLog_Clear(t *testing.T) {
	// The history of the window before: two Activities, one of them failed.
	before := []compose.ActivityEvent{
		started("act-1", "shop", "", compose.ActionUp),
		output("act-1", "shop", "", compose.ActionUp, "one", false),
		finished("act-1", "shop", "", compose.ActionUp, "boom", "boom"),
		started("act-2", "blog", "db", compose.ActionStart),
	}

	for _, tc := range []struct {
		name   string
		before []compose.ActivityEvent
		after  []compose.ActivityEvent
		want   []compose.ActivityID
	}{
		{
			name: "an empty log stays empty",
		},
		{
			name:   "the history of the window before is gone",
			before: before,
		},
		{
			name:   "an Activity after the clear appears alone",
			before: before,
			after:  []compose.ActivityEvent{started("act-3", "shop", "", compose.ActionStop)},
			want:   []compose.ActivityID{"act-3"},
		},
		{
			name:   "the events of an Activity before the clear stay out",
			before: before,
			after: []compose.ActivityEvent{
				output("act-2", "blog", "db", compose.ActionStart, "late", false),
				finished("act-2", "blog", "db", compose.ActionStart, "", ""),
				started("act-3", "shop", "", compose.ActionStop),
			},
			want: []compose.ActivityID{"act-3"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			log := NewActivityLog()
			for _, e := range tc.before {
				log.Push(e)
			}
			log.Clear()
			for _, e := range tc.after {
				log.Push(e)
			}

			var ids []compose.ActivityID
			for _, a := range log.Activities() {
				ids = append(ids, a.ID)
			}
			if !slices.Equal(ids, tc.want) {
				t.Errorf("Activities() = %v, want %v", ids, tc.want)
			}

			// An Activity of the window before the clear is gone, even when
			// the view asks for it by its ID.
			for _, e := range tc.before {
				if e.Type != compose.ActivityStarted || slices.Contains(tc.want, e.ID) {
					continue
				}
				if _, ok := log.Activity(e.ID); ok {
					t.Errorf("Activity(%q) found after the clear, want none", e.ID)
				}
			}
		})
	}
}

func TestActivityLog_ClearNotifiesTheSubscriber(t *testing.T) {
	log := NewActivityLog()
	calls := 0
	log.OnChange(func() { calls++ })

	log.Push(started("act-1", "shop", "", compose.ActionUp))
	log.Clear()

	if calls != 2 {
		t.Fatalf("calls = %d, want 2: the clear must notify the subscriber", calls)
	}
	if got := log.Activities(); len(got) != 0 {
		t.Errorf("Activities() = %v, want none", got)
	}
}

func TestActivityLog_ReadIsACopy(t *testing.T) {
	log := NewActivityLog()
	log.Push(started("act-1", "shop", "", compose.ActionUp))
	log.Push(output("act-1", "shop", "", compose.ActionUp, "one", false))

	got := log.Activities()
	got[0].State = ActivityFailed
	got[0].Lines[0].Text = "tampered"

	again, ok := log.Activity("act-1")
	if !ok {
		t.Fatal("Activity(act-1) not found")
	}
	if again.State != ActivityRunning {
		t.Errorf("state = %q, want %q: the view must not change the log", again.State, ActivityRunning)
	}
	if again.Lines[0].Text != "one" {
		t.Errorf("line = %q, want %q: the view must not change the log", again.Lines[0].Text, "one")
	}
}
