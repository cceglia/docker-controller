package main

import (
	"errors"

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
	{compose.ErrNoComposeFile, "Serve almeno un file compose."},
	{compose.ErrComposeFileUnusable, "Il file compose non esiste, non è leggibile o non è un percorso assoluto."},
	{compose.ErrInvalidProject, "Progetto non valido: i file compose indicati non sono utilizzabili."},
	{compose.ErrActivityRunning, "C'è già un'attività in corso su questo Progetto."},
	{compose.ErrUnknownProject, "Progetto sconosciuto."},
	{compose.ErrUnknownAction, "Azione sconosciuta."},
	{compose.ErrInvalidService, "Nome del servizio non valido."},
	{compose.ErrProjectOnly, "Questa azione riguarda l'intero Progetto, non un singolo servizio."},
	{compose.ErrProjectNotRemoved, "Si può togliere dall'elenco solo un Progetto Rimosso, senza container."},
	{compose.ErrDockerUnreachable, "Docker non è raggiungibile."},
	{settings.ErrInvalidCloseBehavior, "Comportamento alla chiusura non valido."},
}

// unknownErrorPrefix introduces the detail of an error without a message above.
const unknownErrorPrefix = "Operazione non riuscita: "

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
