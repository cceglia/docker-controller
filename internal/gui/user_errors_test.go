package gui

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"dockercontroller/internal/compose"
)

// TestUserError_LogOpenKeepsTheItalianMessage checks the message of the Log
// stream, which stays in this package until the Log errors move to logs.
func TestUserError_LogOpenKeepsTheItalianMessage(t *testing.T) {
	got := UserError(fmt.Errorf("%w: id", ErrLogOpen))
	if got == nil || !strings.Contains(got.Error(), "già aperto") {
		t.Errorf("UserError(ErrLogOpen) = %v, want the Italian message", got)
	}
	if !errors.Is(got, ErrLogOpen) {
		t.Errorf("the original error must stay in the chain")
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
