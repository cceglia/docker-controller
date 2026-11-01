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
		tray     bool
		want     closeAction
	}{
		{"hide with a tray hides the window", settings.CloseHide, true, closeHide},
		{"quit ends the application", settings.CloseQuit, true, closeQuit},
		{"ask opens the dialog", settings.CloseAsk, true, closeAsk},
		{"unknown keeps the default", settings.CloseBehavior("maybe"), true, closeHide},
		{"empty keeps the default", settings.CloseBehavior(""), true, closeHide},
		{"hide without a tray quits", settings.CloseHide, false, closeQuit},
		{"ask without a tray quits", settings.CloseAsk, false, closeQuit},
		{"quit without a tray quits", settings.CloseQuit, false, closeQuit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := closeActionFor(tc.behavior, tc.tray); got != tc.want {
				t.Errorf("closeActionFor(%q, %v) = %v, want %v", tc.behavior, tc.tray, got, tc.want)
			}
		})
	}
}

func TestAskCloseText_DoesNotPromiseTheLossOfTheHistory(t *testing.T) {
	if askCloseHeading != "Chiudere la finestra?" {
		t.Errorf("heading = %q", askCloseHeading)
	}
	for _, want := range []string{"tray", "cronologia delle Attività"} {
		if !strings.Contains(askCloseBody, want) {
			t.Errorf("body = %q, want it to say %q", askCloseBody, want)
		}
	}
	for _, unwanted := range []string{"persa", "perdi", "va perso"} {
		if strings.Contains(askCloseBody, unwanted) {
			t.Errorf("body = %q, want it to leave the history alone, not to say %q", askCloseBody, unwanted)
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
	got := availableCloseAnswers(true)
	if len(got) != len(want) {
		t.Fatalf("%d answers, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i].label != w.label || got[i].behavior != w.behavior {
			t.Errorf("answer %d = %+v, want label %q and behavior %q", i, got[i], w.label, w.behavior)
		}
	}
}

func TestAvailableCloseAnswers_WithoutATrayLeaveNoHide(t *testing.T) {
	got := availableCloseAnswers(false)

	if len(got) != 2 {
		t.Fatalf("%d answers without a tray, want Esci and Annulla", len(got))
	}
	if got[0].behavior != settings.CloseQuit || got[1].behavior != "" {
		t.Errorf("answers = %+v, want Esci and Annulla", got)
	}
}

func TestCloseQuestionBody_WithoutATrayDoesNotOfferHiding(t *testing.T) {
	with := closeQuestionBody(true)
	if !strings.Contains(with, "nascondere") {
		t.Errorf("body = %q, want it to offer hiding with a tray", with)
	}
	without := closeQuestionBody(false)
	for _, unwanted := range []string{"nascondere", "restare nel tray"} {
		if strings.Contains(without, unwanted) {
			t.Errorf("body = %q, want it to offer no hiding without a tray, not %q", without, unwanted)
		}
	}
	if !strings.Contains(without, "esce dall'app") {
		t.Errorf("body = %q, want it to say that closing quits", without)
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
