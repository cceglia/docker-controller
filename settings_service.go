package main

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"dockercontroller/internal/settings"
)

// closeRequestedEvent asks the frontend to show the "Chiedi ogni volta" dialog.
const closeRequestedEvent = "close-requested"

// windowCloser decides what closing the window does. It is a thin Wails
// adapter: the choice itself lives in internal/settings.
type windowCloser struct {
	app      *application.App
	window   application.Window
	settings *settings.Store
	// quitting is set once the application is shutting down, so the closing of
	// the windows done by Wails during the shutdown is let through.
	quitting atomic.Bool
}

// install hooks the closing of the window. A hook (not a listener) is used so
// the default close can be cancelled.
func (c *windowCloser) install() {
	c.app.OnShutdown(func() { c.quitting.Store(true) })
	c.window.RegisterHook(events.Common.WindowClosing, c.onClosing)
}

func (c *windowCloser) onClosing(e *application.WindowEvent) {
	if c.quitting.Load() {
		return
	}
	e.Cancel()
	cfg, err := c.settings.Get(context.Background())
	if err != nil {
		// Never trap the user: fall back to the default behaviour.
		c.app.Logger.Error("close behavior", "error", err)
		cfg = settings.Defaults()
	}
	c.perform(cfg.CloseBehavior)
}

// perform carries out a choice made by the user or read from the settings.
func (c *windowCloser) perform(b settings.CloseBehavior) {
	switch b {
	case settings.CloseQuit:
		go c.app.Quit() // not from inside the closing hook
	case settings.CloseAsk:
		c.app.Event.Emit(closeRequestedEvent)
	default:
		c.window.Hide()
	}
}

// SettingsService is the thin Wails binding over the settings module.
type SettingsService struct {
	store  *settings.Store
	ctx    context.Context
	closer *windowCloser
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
