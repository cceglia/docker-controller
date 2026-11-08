package view

import (
	"errors"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/settings"
)

// italianMessages is the single place where the errors of the backend become
// the Italian text the user reads. The Go errors stay idiomatic English; the
// views pass what they return through UserError. Order matters: a more specific
// error comes before the general one it also wraps.
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
	{compose.ErrUnknownGroup, "Gruppo sconosciuto."},
	{compose.ErrDuplicateGroup, "Esiste già un Gruppo con questo nome."},
	{compose.ErrInvalidGroupName, "Nome del Gruppo non valido: non può essere vuoto, né «Nessun Gruppo», né «Senza Gruppo»."},
	{compose.ErrDockerUnreachable, "Docker non è raggiungibile."},
	{settings.ErrInvalidCloseBehavior, "Comportamento alla chiusura non valido."},
	{settings.ErrInvalidTerminalCommand, "Comando del terminale non valido: controlla virgolette e backslash, e che non sia vuoto."},
	{ErrInvalidPort, "Porta non valida."},
	{ErrNoFolder, NoFolderCommandText},
	{ErrFolderMissing, "La cartella del Progetto non esiste."},
	{ErrCannotOpen, "Impossibile aprire: xdg-open non è riuscito a partire."},
	{ErrTerminalStart, "Impossibile avviare il terminale: controlla il comando nelle Impostazioni."},
}

// Errors of the views that open things on the desktop.
var (
	// ErrInvalidPort reports a port outside 1-65535.
	ErrInvalidPort = errors.New("invalid port")
	// ErrNoFolder reports a Project without a folder.
	ErrNoFolder = errors.New("project has no folder")
	// ErrFolderMissing reports a Project folder that no longer exists.
	ErrFolderMissing = errors.New("project folder does not exist")
	// ErrCannotOpen reports that xdg-open did not start.
	ErrCannotOpen = errors.New("cannot open with xdg-open")
	// ErrTerminalStart reports that the terminal command did not start.
	ErrTerminalStart = errors.New("cannot start the terminal")
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

// UserError returns err with an Italian message: the one registered for the
// error it wraps, or an Italian prefix followed by its own text.
func UserError(err error) error {
	if err == nil {
		return nil
	}
	if msg, ok := KnownMessage(err); ok {
		return WithMessage(err, msg)
	}
	return WithMessage(err, unknownErrorPrefix+err.Error())
}

// KnownMessage returns the Italian message registered for err. A package that
// owns an error view does not know asks for its text here, so one fact keeps
// one sentence.
func KnownMessage(err error) (string, bool) {
	for _, m := range italianMessages {
		if errors.Is(err, m.err) {
			return m.msg, true
		}
	}
	return "", false
}

// WithMessage returns err with an Italian message and keeps the original error
// in the chain, so errors.Is keeps working. A package that owns an error view
// does not know builds its user text with it.
func WithMessage(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &localizedError{msg: msg, cause: err}
}
