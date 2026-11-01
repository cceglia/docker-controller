package gui

import (
	"context"

	"dockercontroller/internal/settings"
)

// closeAction is what closing the window does, once the saved behavior is read.
type closeAction int

const (
	// closeHide takes the window off the screen and leaves the application in
	// the tray. The window, the open Project and the history stay, so the tray
	// brings them back.
	closeHide closeAction = iota
	// closeQuit stops the streams and quits the application.
	closeQuit
	// closeAsk asks the user every time.
	closeAsk
)

// closeActionFor reads a saved close behavior. Without a tray the window could
// not come back, so every behavior quits. A value the settings do not know keeps
// the default of a fresh installation, so closing never traps the user in the
// window.
func closeActionFor(b settings.CloseBehavior, tray bool) closeAction {
	if !tray {
		return closeQuit
	}
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
	// askCloseBody is the body of the dialog: it says what the two answers do,
	// and it promises nothing that does not happen. The window is hidden and
	// not destroyed, so the open Project and the history of the Activities wait
	// for the next click on the tray icon.
	askCloseBody = "Puoi nascondere la finestra e restare nel tray, con i Progetti sotto controllo, " +
		"oppure uscire dall'app. Riaprendo la finestra ritrovi il Progetto aperto e la cronologia delle Attività."
	// rememberCloseText is the control that saves the answer as the setting.
	rememberCloseText = "Ricorda la scelta"
	// askCloseBodyNoTray is the body on a desktop without a tray: the window
	// could not come back, so the body offers no hiding.
	askCloseBodyNoTray = "Questo desktop non ha un'icona nella tray, quindi la finestra non può " +
		"restare nascosta. Chiuderla esce dall'app."
	// closeHideResponse and closeQuitResponse are the responses of the dialog.
	closeHideResponse = "hide"
	closeQuitResponse = "quit"
	// closeCancelResponse is the response of "Annulla" and of a dismissed dialog.
	closeCancelResponse = "cancel"
)

// closeQuestionBody is the body of the question a desktop gets. Without a tray
// the window cannot stay hidden, so the body never offers it.
func closeQuestionBody(tray bool) string {
	if tray {
		return askCloseBody
	}
	return askCloseBodyNoTray
}

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

// availableCloseAnswers are the answers a desktop offers. Without a tray there
// is no way back to a hidden window, so "Nascondi" is not there at all.
func availableCloseAnswers(tray bool) []closeAnswer {
	if tray {
		return askCloseAnswers
	}
	answers := make([]closeAnswer, 0, len(askCloseAnswers))
	for _, answer := range askCloseAnswers {
		if answer.behavior == settings.CloseHide {
			continue
		}
		answers = append(answers, answer)
	}
	return answers
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
