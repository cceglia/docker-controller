package forms

import (
	"context"

	"dockercontroller/internal/settings"
)

// CloseAction is what closing the window does, once the saved behavior is read.
type CloseAction int

const (
	// CloseHide takes the window off the screen and leaves the application in
	// the tray. The window, the open Project and the history stay, so the tray
	// brings them back.
	CloseHide CloseAction = iota
	// CloseQuit stops the streams and quits the application.
	CloseQuit
	// CloseAsk asks the user every time.
	CloseAsk
)

// CloseActionFor reads a saved close behavior. Without a tray the window could
// not come back, so every behavior quits. A value the settings do not know keeps
// the default of a fresh installation, so closing never traps the user in the
// window.
func CloseActionFor(b settings.CloseBehavior, tray bool) CloseAction {
	if !tray {
		return CloseQuit
	}
	switch b {
	case settings.CloseQuit:
		return CloseQuit
	case settings.CloseAsk:
		return CloseAsk
	default:
		return CloseHide
	}
}

const (
	// AskCloseHeading is the title of the "Chiedi ogni volta" dialog.
	AskCloseHeading = "Chiudere la finestra?"
	// AskCloseBody is the body of the dialog: it says what the two answers do,
	// and it promises nothing that does not happen. The window is hidden and
	// not destroyed, so the open Project and the history of the Activities wait
	// for the next click on the tray icon.
	AskCloseBody = "Puoi nascondere la finestra e restare nel tray, con i Progetti sotto controllo, " +
		"oppure uscire dall'app. Riaprendo la finestra ritrovi il Progetto aperto e la cronologia delle Attività."
	// RememberCloseText is the control that saves the answer as the setting.
	RememberCloseText = "Ricorda la scelta"
	// AskCloseBodyNoTray is the body on a desktop without a tray: the window
	// could not come back, so the body offers no hiding.
	AskCloseBodyNoTray = "Questo desktop non ha un'icona nella tray, quindi la finestra non può " +
		"restare nascosta. Chiuderla esce dall'app."
	// closeHideResponse and closeQuitResponse are the responses of the dialog.
	closeHideResponse = "hide"
	closeQuitResponse = "quit"
	// closeCancelResponse is the response of "Annulla" and of a dismissed dialog.
	closeCancelResponse = "cancel"
)

// CloseQuestionBody is the body of the question a desktop gets. Without a tray
// the window cannot stay hidden, so the body never offers it.
func CloseQuestionBody(tray bool) string {
	if tray {
		return AskCloseBody
	}
	return AskCloseBodyNoTray
}

// CloseAnswer is one button of the ask dialog.
type CloseAnswer struct {
	Response string
	Label    string
	// Behavior is what the button does; it is empty for "Annulla", which does
	// nothing and leaves the window open.
	Behavior settings.CloseBehavior
}

// AskCloseAnswers are the answers of the dialog, in the order the window shows
// them.
var AskCloseAnswers = []CloseAnswer{
	{Response: closeHideResponse, Label: "Nascondi", Behavior: settings.CloseHide},
	{Response: closeQuitResponse, Label: "Esci", Behavior: settings.CloseQuit},
	{Response: closeCancelResponse, Label: "Annulla"},
}

// AvailableCloseAnswers are the answers a desktop offers. Without a tray there
// is no way back to a hidden window, so "Nascondi" is not there at all.
func AvailableCloseAnswers(tray bool) []CloseAnswer {
	if tray {
		return AskCloseAnswers
	}
	answers := make([]CloseAnswer, 0, len(AskCloseAnswers))
	for _, answer := range AskCloseAnswers {
		if answer.Behavior == settings.CloseHide {
			continue
		}
		answers = append(answers, answer)
	}
	return answers
}

// closeAnswerFor finds the answer of a response. It reports false for
// "Annulla" and for a dialog closed without an answer: both do nothing.
func closeAnswerFor(response string) (CloseAnswer, bool) {
	for _, answer := range AskCloseAnswers {
		if answer.Response == response && answer.Behavior != "" {
			return answer, true
		}
	}
	return CloseAnswer{}, false
}

// CloseToSave returns the behavior a remembered answer saves, and whether to
// save at all. "Chiedi ogni volta" is not an answer, and "Annulla" saves
// nothing, so only a real answer saves the setting.
func CloseToSave(choice settings.CloseBehavior, remember bool) (settings.CloseBehavior, bool) {
	if !remember || (choice != settings.CloseHide && choice != settings.CloseQuit) {
		return "", false
	}
	return choice, true
}

// SaveCloseBehavior saves value as the close behavior and keeps the other
// settings. Without a store there is nothing to save.
func SaveCloseBehavior(ctx context.Context, store *settings.Store, value settings.CloseBehavior) error {
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
