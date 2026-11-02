package forms

import (
	"fmt"
	"strings"
	"testing"

	"dockercontroller/internal/settings"
)

func TestCloseBehaviorChoices_ListTheThreeBehaviors(t *testing.T) {
	want := []settings.CloseBehavior{settings.CloseHide, settings.CloseQuit, settings.CloseAsk}

	if len(CloseBehaviorChoices) != len(want) {
		t.Fatalf("%d choices, want %d", len(CloseBehaviorChoices), len(want))
	}
	for i, behavior := range want {
		choice := CloseBehaviorChoices[i]
		if choice.Behavior != behavior {
			t.Errorf("choice %d = %q, want %q", i, choice.Behavior, behavior)
		}
		if strings.TrimSpace(choice.Label) == "" {
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
			if got := CloseBehaviorIndex(tc.behavior); got != tc.index {
				t.Errorf("CloseBehaviorIndex(%q) = %d, want %d", tc.behavior, got, tc.index)
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

func TestCloseBehaviorAvailable(t *testing.T) {
	for _, tc := range []struct {
		name     string
		behavior settings.CloseBehavior
		tray     bool
		want     bool
	}{
		{"hide with a tray", settings.CloseHide, true, true},
		{"hide without a tray", settings.CloseHide, false, false},
		{"quit without a tray", settings.CloseQuit, false, true},
		{"ask without a tray", settings.CloseAsk, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := CloseBehaviorAvailable(tc.behavior, tc.tray); got != tc.want {
				t.Errorf("CloseBehaviorAvailable(%q, %v) = %v, want %v", tc.behavior, tc.tray, got, tc.want)
			}
		})
	}
}

func TestCloseBehaviorShown(t *testing.T) {
	for _, tc := range []struct {
		name     string
		behavior settings.CloseBehavior
		tray     bool
		want     settings.CloseBehavior
	}{
		{"hide with a tray", settings.CloseHide, true, settings.CloseHide},
		{"quit stays quit", settings.CloseQuit, true, settings.CloseQuit},
		{"ask stays ask", settings.CloseAsk, true, settings.CloseAsk},
		// Hiding needs the tray: without one the window would never come back,
		// and the close quits. The dialog shows what the close really does.
		{"hide without a tray reads as quit", settings.CloseHide, false, settings.CloseQuit},
		{"ask without a tray stays ask", settings.CloseAsk, false, settings.CloseAsk},
		{"unknown shows the default", settings.CloseBehavior("maybe"), true, settings.CloseHide},
		{"unknown without a tray reads as quit", settings.CloseBehavior("maybe"), false, settings.CloseQuit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := CloseBehaviorShown(tc.behavior, tc.tray); got != tc.want {
				t.Errorf("CloseBehaviorShown(%q, %v) = %q, want %q", tc.behavior, tc.tray, got, tc.want)
			}
		})
	}
}

func TestValidateTerminalCommand(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
		want    string
	}{
		{"empty means the default", "", ""},
		{"blanks mean the default", "   \t", ""},
		{"plain command", "konsole --workdir {dir}", ""},
		{"quoted argument", `konsole -e sh -c "cd {dir} && ls"`, ""},
		{"single quotes", "konsole --workdir '{dir}'", ""},
		{"unterminated quote", `konsole --workdir "dir`, "Comando del terminale non valido"},
		{"trailing backslash", `konsole {dir}\`, "Comando del terminale non valido"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidateTerminalCommand(tc.command)
			if !strings.Contains(got, tc.want) {
				t.Errorf("ValidateTerminalCommand(%q) = %q, want it to contain %q", tc.command, got, tc.want)
			}
			if strings.Contains(got, "settings:") {
				t.Errorf("the English error leaked: %q", got)
			}
		})
	}
}

func TestStoredSettings(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value settings.Settings
		want  settings.Settings
	}{
		{
			name:  "a command stays as it is",
			value: settings.Settings{CloseBehavior: settings.CloseAsk, TerminalCommand: "konsole --workdir {dir}"},
			want:  settings.Settings{CloseBehavior: settings.CloseAsk, TerminalCommand: "konsole --workdir {dir}"},
		},
		{
			name:  "blank spaces around a command go",
			value: settings.Settings{CloseBehavior: settings.CloseQuit, TerminalCommand: "  konsole {dir} "},
			want:  settings.Settings{CloseBehavior: settings.CloseQuit, TerminalCommand: "konsole {dir}"},
		},
		{
			name:  "a blank command is the default one",
			value: settings.Settings{CloseBehavior: settings.CloseHide, TerminalCommand: "   "},
			want:  settings.Settings{CloseBehavior: settings.CloseHide, TerminalCommand: settings.DefaultTerminalCommand},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := StoredSettings(tc.value); got != tc.want {
				t.Errorf("StoredSettings(%+v) = %+v, want %+v", tc.value, got, tc.want)
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
		{"saved", nil, SettingsSavedText, true},
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
