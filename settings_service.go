package main

import (
	"context"
	"fmt"

	"dockercontroller/internal/settings"
)

// closeRequestedEvent asks the frontend to show the "Chiedi ogni volta" dialog.
const closeRequestedEvent = "close-requested"

// SettingsService is the thin Wails binding over the settings module.
type SettingsService struct {
	store  *settings.Store
	ctx    context.Context
	closer *mainWindow
}

// Get returns the application settings.
func (s *SettingsService) Get() (settings.Settings, error) {
	cfg, err := s.store.Get(s.ctx)
	return cfg, userError(err)
}

// Save stores the application settings.
func (s *SettingsService) Save(v settings.Settings) error {
	return userError(s.store.Save(s.ctx, v))
}

// ResolveClose carries out the answer to the "Chiedi ogni volta" dialog: choice
// is "hide" or "quit"; with remember, it also becomes the saved close behavior.
func (s *SettingsService) ResolveClose(choice settings.CloseBehavior, remember bool) error {
	if choice != settings.CloseHide && choice != settings.CloseQuit {
		return userError(fmt.Errorf("%w: %q", settings.ErrInvalidCloseBehavior, choice))
	}
	if remember {
		cfg, err := s.store.Get(s.ctx)
		if err != nil {
			return userError(err)
		}
		cfg.CloseBehavior = choice
		if err := s.store.Save(s.ctx, cfg); err != nil {
			return userError(err)
		}
	}
	s.closer.perform(choice)
	return nil
}
