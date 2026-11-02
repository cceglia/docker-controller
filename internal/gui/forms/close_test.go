package forms

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
		want     CloseAction
	}{
		{"hide with a tray hides the window", settings.CloseHide, true, CloseHide},
		{"quit ends the application", settings.CloseQuit, true, CloseQuit},
		{"ask opens the dialog", settings.CloseAsk, true, CloseAsk},
		{"unknown keeps the default", settings.CloseBehavior("maybe"), true, CloseHide},
		{"empty keeps the default", settings.CloseBehavior(""), true, CloseHide},
		{"hide without a tray quits", settings.CloseHide, false, CloseQuit},
		{"ask without a tray quits", settings.CloseAsk, false, CloseQuit},
		{"quit without a tray quits", settings.CloseQuit, false, CloseQuit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := CloseActionFor(tc.behavior, tc.tray); got != tc.want {
				t.Errorf("CloseActionFor(%q, %v) = %v, want %v", tc.behavior, tc.tray, got, tc.want)
			}
		})
	}
}

func TestAskCloseText_DoesNotPromiseTheLossOfTheHistory(t *testing.T) {
	if AskCloseHeading != "Chiudere la finestra?" {
		t.Errorf("heading = %q", AskCloseHeading)
	}
	for _, want := range []string{"tray", "cronologia delle Attività"} {
		if !strings.Contains(AskCloseBody, want) {
			t.Errorf("body = %q, want it to say %q", AskCloseBody, want)
		}
	}
	for _, unwanted := range []string{"persa", "perdi", "va perso"} {
		if strings.Contains(AskCloseBody, unwanted) {
			t.Errorf("body = %q, want it to leave the history alone, not to say %q", AskCloseBody, unwanted)
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
	got := AvailableCloseAnswers(true)
	if len(got) != len(want) {
		t.Fatalf("%d answers, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i].Label != w.label || got[i].Behavior != w.behavior {
			t.Errorf("answer %d = %+v, want label %q and behavior %q", i, got[i], w.label, w.behavior)
		}
	}
}

func TestAvailableCloseAnswers_WithoutATrayLeaveNoHide(t *testing.T) {
	got := AvailableCloseAnswers(false)

	if len(got) != 2 {
		t.Fatalf("%d answers without a tray, want Esci and Annulla", len(got))
	}
	if got[0].Behavior != settings.CloseQuit || got[1].Behavior != "" {
		t.Errorf("answers = %+v, want Esci and Annulla", got)
	}
}

func TestCloseQuestionBody_WithoutATrayDoesNotOfferHiding(t *testing.T) {
	with := CloseQuestionBody(true)
	if !strings.Contains(with, "nascondere") {
		t.Errorf("body = %q, want it to offer hiding with a tray", with)
	}
	without := CloseQuestionBody(false)
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
			if answer.Behavior != tc.behavior {
				t.Errorf("behavior = %q, want %q", answer.Behavior, tc.behavior)
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
			got, save := CloseToSave(tc.choice, tc.remember)
			if got != tc.want || save != tc.save {
				t.Errorf("CloseToSave(%q, %v) = (%q, %v), want (%q, %v)",
					tc.choice, tc.remember, got, save, tc.want, tc.save)
			}
		})
	}
}
