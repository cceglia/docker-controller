package view

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/settings"
)

func TestUserError_TranslatesKnownErrorsToItalian(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"duplicate project", fmt.Errorf("%w: %q", compose.ErrDuplicateProject, "x"), "Esiste già un Progetto con questo nome."},
		{"invalid name", fmt.Errorf("%w: %w: %q", compose.ErrInvalidProject, compose.ErrInvalidName, "My App"), "Nome non valido"},
		{"no compose file", fmt.Errorf("%w: %w", compose.ErrInvalidProject, compose.ErrNoComposeFile), "Serve almeno un File compose."},
		{"unusable compose file", fmt.Errorf("%w: %w", compose.ErrInvalidProject, compose.ErrComposeFileUnusable), "File compose"},
		{"other invalid project", fmt.Errorf("%w: no usable name", compose.ErrInvalidProject), "Progetto non valido"},
		{"activity running", fmt.Errorf("%w: x", compose.ErrActivityRunning), "Attività in corso"},
		{"unknown project", fmt.Errorf("%w: %q", compose.ErrUnknownProject, "x"), "Progetto sconosciuto."},
		{"unknown action", compose.ErrUnknownAction, "Azione sconosciuta."},
		{"invalid service", compose.ErrInvalidService, "Nome del Servizio non valido."},
		{"project only", compose.ErrProjectOnly, "intero Progetto"},
		{"not removed", compose.ErrProjectNotRemoved, "Rimosso"},
		{"docker unreachable", compose.ErrDockerUnreachable, "Docker non è raggiungibile."},
		{"close behavior", settings.ErrInvalidCloseBehavior, "Comportamento alla chiusura non valido."},
		{"terminal command", fmt.Errorf("%w: unterminated quote", settings.ErrInvalidTerminalCommand), "Comando del terminale non valido"},
		{"invalid port", fmt.Errorf("%w: %d", ErrInvalidPort, 0), "Porta non valida."},
		{"no folder", fmt.Errorf("%w: %s", ErrNoFolder, "p"), "non ha una cartella"},
		{"folder missing", fmt.Errorf("%w: %s", ErrFolderMissing, "/x"), "cartella del Progetto non esiste"},
		{"cannot open", fmt.Errorf("%w: %w", ErrCannotOpen, errors.New("exec: not found")), "Impossibile aprire"},
		{"terminal start", fmt.Errorf("%w: %w", ErrTerminalStart, errors.New("exec: not found")), "Impossibile avviare il terminale"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := UserError(tc.err)
			if got == nil || !strings.Contains(got.Error(), tc.want) {
				t.Errorf("UserError(%v) = %v, want it to contain %q", tc.err, got, tc.want)
			}
			if strings.Contains(got.Error(), "compose:") || strings.Contains(got.Error(), "Operazione non riuscita") {
				t.Errorf("English text leaked: %v", got)
			}
			if !errors.Is(got, tc.err) {
				t.Errorf("the original error must stay in the chain")
			}
		})
	}
}

func TestUserError_UnknownErrorGetsAnItalianPrefixAndTheDetail(t *testing.T) {
	got := UserError(errors.New("disk on fire"))
	if got == nil || got.Error() != "Errore: disk on fire" {
		t.Errorf("got %v", got)
	}
}

func TestUserError_NilStaysNil(t *testing.T) {
	if UserError(nil) != nil {
		t.Error("nil must stay nil")
	}
}

func TestUserError_ItalianTextUsesTheGlossary(t *testing.T) {
	for _, m := range italianMessages {
		for _, bad := range []string{"operazione", "elenco", "attività", "servizio", "container"} {
			if strings.Contains(m.msg, bad) {
				t.Errorf("%q uses %q instead of the glossary term", m.msg, bad)
			}
		}
	}
}

func TestKnownMessage_ReportsOnlyTheKnownErrors(t *testing.T) {
	if msg, ok := KnownMessage(fmt.Errorf("%w: x", compose.ErrUnknownProject)); !ok || msg != "Progetto sconosciuto." {
		t.Errorf("KnownMessage = %q, %t, want the Italian message", msg, ok)
	}
	if msg, ok := KnownMessage(errors.New("disk on fire")); ok || msg != "" {
		t.Errorf("KnownMessage = %q, %t, want no message", msg, ok)
	}
}

func TestKnownMessage_TranslatesTheGroupErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"unknown", fmt.Errorf("%w: %q", compose.ErrUnknownGroup, "Casa"), "Gruppo sconosciuto."},
		{"duplicate", fmt.Errorf("%w: %q", compose.ErrDuplicateGroup, "Casa"), "Esiste già un Gruppo con questo nome."},
		{"invalid", fmt.Errorf("%w: %q", compose.ErrInvalidGroupName, ""), "Nome del Gruppo non valido: non può essere vuoto, né «Nessun Gruppo», né «Senza Gruppo»."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := KnownMessage(tc.err)
			if !ok || got != tc.want {
				t.Errorf("KnownMessage(%v) = %q, %t, want %q", tc.err, got, ok, tc.want)
			}
		})
	}

	// The message of a bad Group name must not read as the one of a bad
	// Project name: they are two different fields of two different forms.
	groupMsg, _ := KnownMessage(compose.ErrInvalidGroupName)
	projectMsg, _ := KnownMessage(compose.ErrInvalidName)
	if groupMsg == projectMsg {
		t.Errorf("the Group and the Project invalid names share the message %q", groupMsg)
	}
}

func TestWithMessage_KeepsTheChain(t *testing.T) {
	cause := errors.New("boom")
	got := WithMessage(cause, "Messaggio.")
	if got == nil || got.Error() != "Messaggio." {
		t.Fatalf("WithMessage = %v, want the Italian message", got)
	}
	if !errors.Is(got, cause) {
		t.Error("the original error must stay in the chain")
	}
	if WithMessage(nil, "Messaggio.") != nil {
		t.Error("nil must stay nil")
	}
}
