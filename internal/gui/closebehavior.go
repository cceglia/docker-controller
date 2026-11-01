package gui

import (
	"context"

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

// askCloseAnswers are the answers of the dialog, in the order the window shows
// them.
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
