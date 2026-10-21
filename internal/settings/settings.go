// Package settings keeps the application's own settings (what closing the
// window does, which terminal "Apri terminale" launches) in the key/value
// settings table of the application database.
package settings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// CloseBehavior is what closing the window does.
type CloseBehavior string

const (
	// CloseHide hides the window in the tray (the default).
	CloseHide CloseBehavior = "hide"
	// CloseQuit quits the application.
	CloseQuit CloseBehavior = "quit"
	// CloseAsk asks every time.
	CloseAsk CloseBehavior = "ask"
)

// DefaultTerminalCommand is the "Apri terminale" command until the user
// changes it; {dir} is replaced by the Project folder.
const DefaultTerminalCommand = "konsole --workdir {dir}"

// ErrInvalidCloseBehavior is returned by Save for a value that is not one of
// the CloseBehavior constants.
var ErrInvalidCloseBehavior = errors.New("settings: invalid close behavior")

const (
	keyCloseBehavior   = "close_behavior"
	keyTerminalCommand = "terminal_command"
)

// Settings are the application settings.
type Settings struct {
	CloseBehavior   CloseBehavior `json:"closeBehavior"`
	TerminalCommand string        `json:"terminalCommand"`
}

// Defaults returns the settings of a fresh installation.
func Defaults() Settings {
	return Settings{CloseBehavior: CloseHide, TerminalCommand: DefaultTerminalCommand}
}

// Valid reports whether b is one of the CloseBehavior constants.
func (b CloseBehavior) Valid() bool {
	return b == CloseHide || b == CloseQuit || b == CloseAsk
}

// Store reads and writes the settings on the application database.
type Store struct {
	db *sql.DB
}

// New returns a Store over a database opened with db.Open.
func New(conn *sql.DB) *Store {
	return &Store{db: conn}
}

// Get returns the saved settings; a setting never saved (or saved with a value
// that is no longer valid) has its default.
func (s *Store) Get(ctx context.Context) (Settings, error) {
	out := Defaults()
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM settings WHERE key IN (?, ?)`, keyCloseBehavior, keyTerminalCommand)
	if err != nil {
		return Settings{}, fmt.Errorf("settings: read: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return Settings{}, fmt.Errorf("settings: read: %w", err)
		}
		switch key {
		case keyCloseBehavior:
			if b := CloseBehavior(value); b.Valid() {
				out.CloseBehavior = b
			}
		case keyTerminalCommand:
			if strings.TrimSpace(value) != "" {
				out.TerminalCommand = value
			}
		}
	}
	if err := rows.Err(); err != nil {
		return Settings{}, fmt.Errorf("settings: read: %w", err)
	}
	return out, nil
}

// Save stores the settings. An unknown close behavior is refused with
// ErrInvalidCloseBehavior; a blank terminal command restores the default and
// one that cannot be split into arguments is refused. Nothing is saved when
// it fails.
func (s *Store) Save(ctx context.Context, v Settings) error {
	if !v.CloseBehavior.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidCloseBehavior, v.CloseBehavior)
	}
	v.TerminalCommand = strings.TrimSpace(v.TerminalCommand)
	if v.TerminalCommand == "" {
		v.TerminalCommand = DefaultTerminalCommand
	}
	if _, err := TerminalArgv(v.TerminalCommand, ""); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("settings: save: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	for key, value := range map[string]string{keyCloseBehavior: string(v.CloseBehavior), keyTerminalCommand: v.TerminalCommand} {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value); err != nil {
			return fmt.Errorf("settings: save %s: %w", key, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("settings: save: %w", err)
	}
	return nil
}
