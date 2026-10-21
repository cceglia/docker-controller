package main

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
		{"invalid port", fmt.Errorf("%w: %d", errInvalidPort, 0), "Porta non valida."},
		{"no folder", fmt.Errorf("%w: %s", errNoFolder, "p"), "non ha una cartella"},
		{"folder missing", fmt.Errorf("%w: %s", errFolderMissing, "/x"), "cartella del Progetto non esiste"},
		{"log open", fmt.Errorf("%w: id", errLogOpen), "già aperto"},
		{"cannot open", fmt.Errorf("%w: %w", errCannotOpen, errors.New("exec: not found")), "Impossibile aprire"},
		{"terminal start", fmt.Errorf("%w: %w", errTerminalStart, errors.New("exec: not found")), "Impossibile avviare il terminale"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := userError(tc.err)
			if got == nil || !strings.Contains(got.Error(), tc.want) {
				t.Errorf("userError(%v) = %v, want it to contain %q", tc.err, got, tc.want)
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
	got := userError(errors.New("disk on fire"))
	if got == nil || got.Error() != "Errore: disk on fire" {
		t.Errorf("got %v", got)
	}
}

func TestUserError_NilStaysNil(t *testing.T) {
	if userError(nil) != nil {
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

func TestLogEndedMessage(t *testing.T) {
	exit := errors.New("docker compose -p x logs exit status 1")
	for _, tc := range []struct {
		name   string
		err    error
		stderr []string
		want   string
	}{
		{"known error is translated", fmt.Errorf("%w: %q", compose.ErrUnknownProject, "x"), nil, "Progetto sconosciuto."},
		{"stderr detail is kept", exit, []string{"warn", "Error response from daemon: boom"}, "Flusso dei log interrotto: Error response from daemon: boom"},
		{"no detail, no command line", exit, nil, "Flusso dei log interrotto."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := logEndedMessage(tc.err, tc.stderr); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
