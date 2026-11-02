package logs

import (
	"errors"
	"strings"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/view"
)

// logOpenText is the Italian message of ErrLogOpen.
const logOpenText = "Flusso dei log già aperto."

// UserError returns err with an Italian message: the one view registers for the
// error it wraps, or the one of the Log stream. ErrLogOpen lives in this
// package, so this wrapper adds its message to the ones of view.
func UserError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrLogOpen) {
		return view.WithMessage(err, logOpenText)
	}
	return view.UserError(err)
}

// logEndedMessage is the Italian reason a log stream ended on its own: the
// message of a known error, otherwise a generic one with the most informative
// stderr line of compose when there is one. The raw command line of a failed
// process is never shown.
func logEndedMessage(err error, stderr []string) string {
	if msg, ok := view.KnownMessage(err); ok {
		return msg
	}
	const base = "Flusso dei log interrotto"
	if detail := compose.Summarize(strings.Join(stderr, "\n")); detail != "" {
		return base + ": " + detail
	}
	return base + "."
}
