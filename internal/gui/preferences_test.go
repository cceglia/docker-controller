package gui

import (
	"fmt"
	"strings"
	"testing"

	"dockercontroller/internal/settings"
)

func TestCloseBehaviorChoices_ListTheThreeBehaviors(t *testing.T) {
	want := []settings.CloseBehavior{settings.CloseHide, settings.CloseQuit, settings.CloseAsk}

	if len(closeBehaviorChoices) != len(want) {
		t.Fatalf("%d choices, want %d", len(closeBehaviorChoices), len(want))
	}
	for i, behavior := range want {
		choice := closeBehaviorChoices[i]
		if choice.behavior != behavior {
			t.Errorf("choice %d = %q, want %q", i, choice.behavior, behavior)
		}
		if strings.TrimSpace(choice.label) == "" {
			t.Errorf("choice %d has no label", i)
		}
	}
}

func TestCloseBehaviorIndexAndAt_RoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name     string
		behavior settings.CloseBehavior
		index    int
	}{
		{"hide", settings.CloseHide, 0},
		{"quit", settings.CloseQuit, 1},
		{"ask", settings.CloseAsk, 2},
		{"unknown shows the default", settings.CloseBehavior("maybe"), 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := closeBehaviorIndex(tc.behavior); got != tc.index {
				t.Errorf("closeBehaviorIndex(%q) = %d, want %d", tc.behavior, got, tc.index)
			}
			want := tc.behavior
			if !want.Valid() {
				want = settings.Defaults().CloseBehavior
			}
			if got := closeBehaviorAt(tc.index); got != want {
				t.Errorf("closeBehaviorAt(%d) = %q, want %q", tc.index, got, want)
			}
		})
	}
}

func TestPreferencesSaveOutcome(t *testing.T) {
	for _, tc := range []struct {
		name        string
		err         error
		wantMessage string
		wantSaved   bool
	}{
		{"saved", nil, settingsSavedText, true},
		{
			"invalid terminal command",
			fmt.Errorf("%w: unterminated quote", settings.ErrInvalidTerminalCommand),
			"Comando del terminale non valido",
			false,
		},
		{
			"invalid close behavior",
			fmt.Errorf("%w: %q", settings.ErrInvalidCloseBehavior, "maybe"),
			"Comportamento alla chiusura non valido.",
			false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message, saved := preferencesSaveOutcome(tc.err)
			if saved != tc.wantSaved {
				t.Errorf("saved = %v, want %v", saved, tc.wantSaved)
			}
			if !strings.Contains(message, tc.wantMessage) {
				t.Errorf("message = %q, want it to contain %q", message, tc.wantMessage)
			}
			if strings.Contains(message, "settings:") {
				t.Errorf("the English error leaked: %q", message)
			}
		})
	}
}
