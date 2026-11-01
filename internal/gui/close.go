package gui

import (
	"context"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"dockercontroller/internal/settings"
)

// closeAction is what closing the window does, once the saved behavior is read.
type closeAction int

const (
	// closeHide destroys the window and leaves the application in the tray.
	closeHide closeAction = iota
	// closeQuit stops the streams and quits the application.
	closeQuit
	// closeAsk asks the user every time.
	closeAsk
)

// closeActionFor reads a saved close behavior. A value the settings do not know
// keeps the default of a fresh installation, so closing never traps the user in
// the window.
func closeActionFor(b settings.CloseBehavior) closeAction {
	switch b {
	case settings.CloseQuit:
		return closeQuit
	case settings.CloseAsk:
		return closeAsk
	default:
		return closeHide
	}
}

const (
	// askCloseHeading is the title of the "Chiedi ogni volta" dialog.
	askCloseHeading = "Chiudere la finestra?"
	// askCloseBody is the body of the dialog: it says what the user loses.
	askCloseBody = "Puoi chiudere la finestra e restare nel tray, con i Progetti sotto controllo, " +
		"oppure uscire dall'app. La cronologia delle Attività va persa."
	// rememberCloseText is the control that saves the answer as the setting.
	rememberCloseText = "Ricorda la scelta"
	// closeHideResponse and closeQuitResponse are the responses of the dialog.
	closeHideResponse = "hide"
	closeQuitResponse = "quit"
	// closeCancelResponse is the response of "Annulla" and of a dismissed dialog.
	closeCancelResponse = "cancel"
)

// closeAnswer is one button of the ask dialog.
type closeAnswer struct {
	response string
	label    string
	// behavior is what the button does; it is empty for "Annulla", which does
	// nothing and leaves the window open.
	behavior settings.CloseBehavior
}

// askCloseAnswers are the answers of the dialog, in the order GTK shows them.
var askCloseAnswers = []closeAnswer{
	{response: closeHideResponse, label: "Nascondi", behavior: settings.CloseHide},
	{response: closeQuitResponse, label: "Esci", behavior: settings.CloseQuit},
	{response: closeCancelResponse, label: "Annulla"},
}

// closeAnswerFor finds the answer of a response. It reports false for
// "Annulla" and for a dialog closed without an answer: both do nothing.
func closeAnswerFor(response string) (closeAnswer, bool) {
	for _, answer := range askCloseAnswers {
		if answer.response == response && answer.behavior != "" {
			return answer, true
		}
	}
	return closeAnswer{}, false
}

// closeToSave returns the behavior a remembered answer saves, and whether to
// save at all. "Chiedi ogni volta" is not an answer, and "Annulla" saves
// nothing, so only a real answer saves the setting.
func closeToSave(choice settings.CloseBehavior, remember bool) (settings.CloseBehavior, bool) {
	if !remember || (choice != settings.CloseHide && choice != settings.CloseQuit) {
		return "", false
	}
	return choice, true
}

// onCloseRequest decides what the close of the window does. The saved behavior
// comes from the database, so the read runs off the GTK loop. The window is
// kept open until the answer is known, and the answer runs on the loop.
func (s *shell) onCloseRequest() bool {
	if s.app == nil || s.app.quitting {
		return false
	}
	if s.closing {
		return true
	}
	store := s.app.store
	ctx := s.app.ctx
	go func() {
		// Never trap the user: a settings that cannot be read applies the
		// default behavior.
		cfg := settings.Defaults()
		if store != nil {
			if saved, err := store.Get(ctx); err == nil {
				cfg = saved
			}
		}
		action := closeActionFor(cfg.CloseBehavior)
		glib.IdleAdd(func() { s.performClose(action) })
	}()
	return true
}

// performClose carries out a close behavior chosen by the user or saved in the
// settings.
func (s *shell) performClose(action closeAction) {
	switch action {
	case closeQuit:
		s.app.quit()
	case closeAsk:
		s.askClose()
	default:
		s.hideWindow()
	}
}

// hideWindow stops the streams and then destroys the window, so the application
// stays in the tray with no window. The destroy runs when the loop is idle, not
// inside the handler that asked for the close.
func (s *shell) hideWindow() {
	if s.closing {
		return
	}
	s.closing = true
	s.stopStreams()
	glib.IdleAdd(func() {
		// The application forgets the window, and closes the settings window
		// that belongs to it, before the window goes away.
		s.app.dropWindow(s)
		s.window.Destroy()
	})
}

// stopStreams ends Utilizzo risorse and every log stream, so no `docker stats`
// or `docker compose logs` outlives the window.
func (s *shell) stopStreams() {
	if s.detail == nil || s.detail.panel == nil {
		return
	}
	s.detail.panel.logs.StopAll()
	s.stopUsage()
}

// askClose opens the "Chiedi ogni volta" dialog with the option to remember the
// answer. A dialog already on screen is brought forward instead of duplicated.
func (s *shell) askClose() {
	if s.ask != nil {
		s.ask.Present()
		return
	}
	dialog := adw.NewMessageDialog(&s.window.ApplicationWindow.Window, askCloseHeading, askCloseBody)
	remember := gtk.NewCheckButtonWithLabel(rememberCloseText)
	dialog.SetExtraChild(remember)
	for _, answer := range askCloseAnswers {
		dialog.AddResponse(answer.response, answer.label)
	}
	dialog.SetCloseResponse(closeCancelResponse)
	dialog.SetDefaultResponse(closeCancelResponse)
	dialog.SetResponseAppearance(closeQuitResponse, adw.ResponseDestructive)
	dialog.ConnectResponse(func(response string) {
		s.ask = nil
		answer, ok := closeAnswerFor(response)
		if !ok {
			return
		}
		s.answerClose(answer.behavior, remember.Active())
	})
	s.ask = dialog
	dialog.Present()
}

// answerClose carries out the answer of the dialog. A remembered answer is
// saved first, so the choice still holds after a restart.
func (s *shell) answerClose(choice settings.CloseBehavior, remember bool) {
	value, save := closeToSave(choice, remember)
	if !save {
		s.performClose(closeActionFor(choice))
		return
	}
	store := s.app.store
	ctx := s.app.ctx
	go func() {
		err := saveCloseBehavior(ctx, store, value)
		glib.IdleAdd(func() {
			if err != nil {
				s.toast(UserError(err).Error())
				return
			}
			s.performClose(closeActionFor(value))
		})
	}()
}

// saveCloseBehavior saves value as the close behavior and keeps the other
// settings. Without a store there is nothing to save.
func saveCloseBehavior(ctx context.Context, store *settings.Store, value settings.CloseBehavior) error {
	if store == nil {
		return nil
	}
	cfg, err := store.Get(ctx)
	if err != nil {
		return err
	}
	cfg.CloseBehavior = value
	return store.Save(ctx, cfg)
}
