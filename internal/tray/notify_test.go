package tray

import (
	"strings"
	"testing"

	"dockercontroller/internal/compose"
)

func ev(id string, typ compose.ActivityEventType, project string, kind compose.ActionKind, errText string) compose.ActivityEvent {
	return compose.ActivityEvent{ID: compose.ActivityID(id), Type: typ, Project: project, Kind: kind, Error: errText, Summary: compose.Summarize(errText)}
}

func TestNotifier_TrayLaunchedActivityNotifiesOnSuccess(t *testing.T) {
	var n Notifier
	act := compose.Action{Project: "personalvpn", Kind: compose.ActionRestart}
	n.Launching(act)

	n.Observe(ev("act-1", compose.ActivityStarted, "personalvpn", compose.ActionRestart, ""), "VPN")
	if _, ok := n.Observe(ev("act-1", compose.ActivityLine, "personalvpn", compose.ActionRestart, ""), "VPN"); ok {
		t.Fatal("output lines must not notify")
	}
	got, ok := n.Observe(ev("act-1", compose.ActivityFinished, "personalvpn", compose.ActionRestart, ""), "VPN")
	if !ok {
		t.Fatal("no notification at the end of a tray Activity")
	}
	if got.Title != "Docker Controller" || got.Body != "Restart di VPN completato" {
		t.Errorf("got %+v", got)
	}
}

func TestNotifier_ServiceActivityNamesTheService(t *testing.T) {
	var n Notifier
	n.Launching(compose.Action{Project: "p", Service: "web", Kind: compose.ActionStop})
	e := ev("act-1", compose.ActivityStarted, "p", compose.ActionStop, "")
	e.Service = "web"
	n.Observe(e, "p")
	e.Type = compose.ActivityFinished
	got, ok := n.Observe(e, "p")
	if !ok || got.Body != "Stop di web (p) completato" {
		t.Errorf("got %+v ok=%v", got, ok)
	}
}

func TestNotifier_FailureNotifiesWhateverTheOrigin(t *testing.T) {
	var n Notifier // never launched from the tray
	n.Observe(ev("act-7", compose.ActivityStarted, "p", compose.ActionPull, ""), "p")
	got, ok := n.Observe(ev("act-7", compose.ActivityFinished, "p", compose.ActionPull, "progress\nError response from daemon: pull access denied\nsecond line"), "p")
	if !ok {
		t.Fatal("failed Activity must notify")
	}
	if got.Body != "Pull di p fallito: Error response from daemon: pull access denied" {
		t.Errorf("body = %q", got.Body)
	}
}

func TestNotifier_FailureBodyIsTruncated(t *testing.T) {
	var n Notifier
	got, ok := n.Observe(ev("act-1", compose.ActivityFinished, "p", compose.ActionUp, strings.Repeat("x", 500)), "p")
	if !ok || len([]rune(got.Body)) > 250 {
		t.Errorf("ok=%v len=%d", ok, len([]rune(got.Body)))
	}
}

func TestNotifier_SuccessOutsideTheTrayIsSilent(t *testing.T) {
	var n Notifier
	n.Observe(ev("act-2", compose.ActivityStarted, "p", compose.ActionUp, ""), "p")
	if _, ok := n.Observe(ev("act-2", compose.ActivityFinished, "p", compose.ActionUp, ""), "p"); ok {
		t.Fatal("a window-launched success must not notify")
	}
}

func TestNotifier_AbortedLaunchIsForgotten(t *testing.T) {
	var n Notifier
	act := compose.Action{Project: "p", Kind: compose.ActionUp}
	n.Launching(act)
	n.Aborted(act) // Do refused it
	n.Observe(ev("act-3", compose.ActivityStarted, "p", compose.ActionUp, ""), "p")
	if _, ok := n.Observe(ev("act-3", compose.ActivityFinished, "p", compose.ActionUp, ""), "p"); ok {
		t.Fatal("an aborted tray launch must not claim a later Activity")
	}
}

func TestNotifier_FinishedActivityIsNotNotifiedTwice(t *testing.T) {
	var n Notifier
	n.Launching(compose.Action{Project: "p", Kind: compose.ActionUp})
	n.Observe(ev("act-1", compose.ActivityStarted, "p", compose.ActionUp, ""), "p")
	n.Observe(ev("act-1", compose.ActivityFinished, "p", compose.ActionUp, ""), "p")
	if _, ok := n.Observe(ev("act-1", compose.ActivityFinished, "p", compose.ActionUp, ""), "p"); ok {
		t.Fatal("second finished event notified again")
	}
}
