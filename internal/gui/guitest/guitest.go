// Package guitest holds the helpers the gui tests share: the builders of the
// Activity events, the comparison of two string slices and the wait for a
// signal. It is test support, not production code: only test files import it,
// and it imports compose and never the root package gui.
package guitest

import (
	"testing"
	"time"

	"dockercontroller/internal/compose"
)

// Started builds the event the Manager sends when an Activity starts.
func Started(id compose.ActivityID, project, service string, kind compose.ActionKind) compose.ActivityEvent {
	return compose.ActivityEvent{ID: id, Type: compose.ActivityStarted, Project: project, Service: service, Kind: kind}
}

// Output builds the event the Manager sends for one line of the output of an
// Activity.
func Output(id compose.ActivityID, project, service string, kind compose.ActionKind, text string, stderr bool) compose.ActivityEvent {
	return compose.ActivityEvent{ID: id, Type: compose.ActivityLine, Project: project, Service: service, Kind: kind, Line: text, Stderr: stderr}
}

// Finished builds the event the Manager sends when an Activity ends. An empty
// err marks a success, and summary is the one line of a failure.
func Finished(id compose.ActivityID, project, service string, kind compose.ActionKind, err, summary string) compose.ActivityEvent {
	return compose.ActivityEvent{ID: id, Type: compose.ActivityFinished, Project: project, Service: service, Kind: kind, Error: err, Summary: summary}
}

// AssertStrings compares two runs of strings, treating nil and empty as the
// same list of nothing.
func AssertStrings(t *testing.T, what string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %q, want %q", what, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s = %q, want %q", what, got, want)
		}
	}
}

// WaitSignal fails the test when signal does not arrive in a short time, so a
// deadlock of the streamer reads as a failure and not as a hang.
func WaitSignal(t *testing.T, signal <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}
