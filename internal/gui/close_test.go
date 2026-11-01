package gui

import (
	"strings"
	"testing"

	"dockercontroller/internal/settings"
)

func TestCloseActionFor(t *testing.T) {
	for _, tc := range []struct {
		name     string
		behavior settings.CloseBehavior
		want     closeAction
	}{
		{"hide destroys the window", settings.CloseHide, closeHide},
		{"quit ends the application", settings.CloseQuit, closeQuit},
		{"ask opens the dialog", settings.CloseAsk, closeAsk},
		{"unknown keeps the default", settings.CloseBehavior("maybe"), closeHide},
		{"empty keeps the default", settings.CloseBehavior(""), closeHide},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := closeActionFor(tc.behavior); got != tc.want {
				t.Errorf("closeActionFor(%q) = %v, want %v", tc.behavior, got, tc.want)
			}
		})
	}
}

func TestAskCloseText_SaysThatTheActivityHistoryIsLost(t *testing.T) {
	if askCloseHeading != "Chiudere la finestra?" {
		t.Errorf("heading = %q", askCloseHeading)
	}
	for _, want := range []string{"tray", "cronologia delle Attività va persa"} {
		if !strings.Contains(askCloseBody, want) {
			t.Errorf("body = %q, want it to say %q", askCloseBody, want)
		}
	}
}

func TestAskCloseAnswers_AreHideQuitAndCancel(t *testing.T) {
	want := []struct {
		label    string
		behavior settings.CloseBehavior
	}{
		{"Nascondi", settings.CloseHide},
		{"Esci", settings.CloseQuit},
		{"Annulla", ""},
	}
	if len(askCloseAnswers) != len(want) {
		t.Fatalf("%d answers, want %d", len(askCloseAnswers), len(want))
	}
	for i, w := range want {
		got := askCloseAnswers[i]
		if got.label != w.label || got.behavior != w.behavior {
			t.Errorf("answer %d = %+v, want label %q and behavior %q", i, got, w.label, w.behavior)
		}
	}
}

func TestCloseAnswerFor(t *testing.T) {
	for _, tc := range []struct {
		name     string
		response string
		behavior settings.CloseBehavior
		ok       bool
	}{
		{"hide", "hide", settings.CloseHide, true},
		{"quit", "quit", settings.CloseQuit, true},
		{"cancel", "cancel", "", false},
		{"closed without an answer", "close", "", false},
		{"unknown response", "other", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			answer, ok := closeAnswerFor(tc.response)
			if ok != tc.ok {
				t.Fatalf("closeAnswerFor(%q) ok = %v, want %v", tc.response, ok, tc.ok)
			}
			if answer.behavior != tc.behavior {
				t.Errorf("behavior = %q, want %q", answer.behavior, tc.behavior)
			}
		})
	}
}

func TestCloseToSave(t *testing.T) {
	for _, tc := range []struct {
		name     string
		choice   settings.CloseBehavior
		remember bool
		want     settings.CloseBehavior
		save     bool
	}{
		{"remembered hide", settings.CloseHide, true, settings.CloseHide, true},
		{"remembered quit", settings.CloseQuit, true, settings.CloseQuit, true},
		{"ask is never saved as an answer", settings.CloseAsk, true, "", false},
		{"empty answer", "", true, "", false},
		{"not remembered hide", settings.CloseHide, false, "", false},
		{"not remembered quit", settings.CloseQuit, false, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, save := closeToSave(tc.choice, tc.remember)
			if got != tc.want || save != tc.save {
				t.Errorf("closeToSave(%q, %v) = (%q, %v), want (%q, %v)",
					tc.choice, tc.remember, got, save, tc.want, tc.save)
			}
		})
	}
}
