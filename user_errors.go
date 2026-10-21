package main

import (
	"errors"
	"strings"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/settings"
)

// italianMessages is the single place where the errors of the backend become
// the Italian text the user reads. The Go errors stay idiomatic English; the
// bindings pass what they return through userError. Order matters: a more
// specific error comes before the general one it also wraps.
var italianMessages = []struct {
	err error
	msg string
}{
	{compose.ErrDuplicateProject, "Esiste già un Progetto con questo nome."},
	{compose.ErrInvalidName, "Nome non valido: usa lettere minuscole, cifre, - e _."},
	{compose.ErrNoComposeFile, "Serve almeno un File compose."},
	{compose.ErrComposeFileUnusable, "Il File compose non esiste, non è leggibile o non è un percorso assoluto."},
	{compose.ErrInvalidProject, "Progetto non valido: i File compose indicati non sono utilizzabili."},
	{compose.ErrActivityRunning, "C'è già un'Attività in corso su questo Progetto."},
	{compose.ErrUnknownProject, "Progetto sconosciuto."},
	{compose.ErrUnknownAction, "Azione sconosciuta."},
	{compose.ErrInvalidService, "Nome del Servizio non valido."},
	{compose.ErrProjectOnly, "Questa Azione riguarda l'intero Progetto, non un singolo Servizio."},
	{compose.ErrProjectNotRemoved, "Si può togliere dalla lista solo un Progetto Rimosso, senza Container."},
	{compose.ErrDockerUnreachable, "Docker non è raggiungibile."},
	{settings.ErrInvalidCloseBehavior, "Comportamento alla chiusura non valido."},
	{settings.ErrInvalidTerminalCommand, "Comando del terminale non valido: controlla virgolette e backslash, e che non sia vuoto."},
	{errInvalidPort, "Porta non valida."},
	{errNoFolder, "Il Progetto non ha una cartella."},
	{errFolderMissing, "La cartella del Progetto non esiste."},
	{errLogOpen, "Flusso dei log già aperto."},
	{errCannotOpen, "Impossibile aprire: xdg-open non è riuscito a partire."},
	{errTerminalStart, "Impossibile avviare il terminale: controlla il comando nelle Impostazioni."},
}

// Errors of the bindings that open things on the desktop.
var (
	errInvalidPort   = errors.New("invalid port")
	errNoFolder      = errors.New("project has no folder")
	errFolderMissing = errors.New("project folder does not exist")
	errCannotOpen    = errors.New("cannot open with xdg-open")
	errTerminalStart = errors.New("cannot start the terminal")
)

// unknownErrorPrefix introduces the detail of an error without a message above.
const unknownErrorPrefix = "Errore: "

// localizedError is an error whose text is Italian and whose chain still holds
// the original error, so errors.Is keeps working.
type localizedError struct {
	msg   string
	cause error
}

func (e *localizedError) Error() string { return e.msg }
func (e *localizedError) Unwrap() error { return e.cause }

// userError returns err with an Italian message: the one registered for the
// error it wraps, or an Italian prefix followed by its own text.
func userError(err error) error {
	if err == nil {
		return nil
	}
	for _, m := range italianMessages {
		if errors.Is(err, m.err) {
			return &localizedError{msg: m.msg, cause: err}
		}
	}
	return &localizedError{msg: unknownErrorPrefix + err.Error(), cause: err}
}

// logEndedMessage is the Italian reason a log stream ended on its own: the
// message of a known error, otherwise a generic one with the most informative
// stderr line of compose when there is one. The raw command line of a failed
// process is never shown.
func logEndedMessage(err error, stderr []string) string {
	for _, m := range italianMessages {
		if errors.Is(err, m.err) {
			return m.msg
		}
	}
	const base = "Flusso dei log interrotto"
	if detail := compose.Summarize(strings.Join(stderr, "\n")); detail != "" {
		return base + ": " + detail
	}
	return base + "."
}
