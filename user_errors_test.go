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
		{"no compose file", fmt.Errorf("%w: %w", compose.ErrInvalidProject, compose.ErrNoComposeFile), "Serve almeno un file compose."},
		{"unusable compose file", fmt.Errorf("%w: %w", compose.ErrInvalidProject, compose.ErrComposeFileUnusable), "file compose"},
		{"other invalid project", fmt.Errorf("%w: no usable name", compose.ErrInvalidProject), "Progetto non valido"},
		{"activity running", fmt.Errorf("%w: x", compose.ErrActivityRunning), "attività in corso"},
		{"unknown project", fmt.Errorf("%w: %q", compose.ErrUnknownProject, "x"), "Progetto sconosciuto."},
		{"unknown action", compose.ErrUnknownAction, "Azione sconosciuta."},
		{"invalid service", compose.ErrInvalidService, "Nome del servizio non valido."},
		{"project only", compose.ErrProjectOnly, "intero Progetto"},
		{"not removed", compose.ErrProjectNotRemoved, "Rimosso"},
		{"docker unreachable", compose.ErrDockerUnreachable, "Docker non è raggiungibile."},
		{"close behavior", settings.ErrInvalidCloseBehavior, "Comportamento alla chiusura non valido."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := userError(tc.err)
			if got == nil || !strings.Contains(got.Error(), tc.want) {
				t.Errorf("userError(%v) = %v, want it to contain %q", tc.err, got, tc.want)
			}
			if strings.Contains(got.Error(), "compose:") {
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
	if got == nil || got.Error() != "Operazione non riuscita: disk on fire" {
		t.Errorf("got %v", got)
	}
}

func TestUserError_NilStaysNil(t *testing.T) {
	if userError(nil) != nil {
		t.Error("nil must stay nil")
	}
}
