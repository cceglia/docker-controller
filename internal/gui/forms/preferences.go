package forms

import (
	"strings"

	"dockercontroller/internal/gui/view"
	"dockercontroller/internal/settings"
)

const (
	// PreferencesTitle is the title of the settings window.
	PreferencesTitle = "Impostazioni"
	// SaveSettingsText is the action that stores the settings.
	SaveSettingsText = "Salva"
	// SettingsSavedText confirms a save of the settings.
	SettingsSavedText = "Impostazioni salvate."
	// CloseBehaviorGroupTitle and CloseBehaviorGroupHint describe the group of
	// the close behavior. The hint says what hiding the window keeps, because
	// the window is hidden and not destroyed.
	CloseBehaviorGroupTitle = "Alla chiusura della finestra"
	CloseBehaviorGroupHint  = "Cosa fa l'app quando chiudi la finestra. Con Nascondi la finestra va nel tray " +
		"e la riapri con un click sull'icona: il Progetto aperto e la cronologia delle Attività restano."
	// NoTrayHintText explains why "Nascondi" cannot be chosen: the desktop has
	// no icon to bring the window back.
	NoTrayHintText = "Questo desktop non ha un'icona nella tray: senza di essa la finestra non si può " +
		"riaprire, quindi la chiusura esce dall'app."
	// TerminalGroupTitle and TerminalGroupHint describe the group of the
	// terminal command.
	TerminalGroupTitle = "Comando del terminale"
	TerminalGroupHint  = "{dir} viene sostituito con la cartella del Progetto. " +
		"Vuoto per il comando predefinito."
	// TerminalRowTitle names the field of the command.
	TerminalRowTitle = "Comando"
)

// CloseBehaviorChoice is one option of the close behavior: the label it shows
// and the value it saves.
type CloseBehaviorChoice struct {
	Label    string
	Behavior settings.CloseBehavior
}

// CloseBehaviorChoices are the three behaviors, in the order the window shows
// them.
var CloseBehaviorChoices = []CloseBehaviorChoice{
	{Label: "Nascondi la finestra nella tray", Behavior: settings.CloseHide},
	{Label: "Chiudi l'app", Behavior: settings.CloseQuit},
	{Label: "Chiedi ogni volta", Behavior: settings.CloseAsk},
}

// CloseBehaviorIndex is the place of a behavior in the choices. A value the
// settings do not know shows the default.
func CloseBehaviorIndex(b settings.CloseBehavior) int {
	for i, choice := range CloseBehaviorChoices {
		if choice.Behavior == b {
			return i
		}
	}
	return 0
}

// closeBehaviorAt is the behavior of a place in the choices. A place out of the
// list is the default.
func closeBehaviorAt(i int) settings.CloseBehavior {
	if i < 0 || i >= len(CloseBehaviorChoices) {
		return settings.Defaults().CloseBehavior
	}
	return CloseBehaviorChoices[i].Behavior
}

// CloseBehaviorAvailable reports whether a behavior can be chosen: hiding the
// window needs a tray, because without one the user could not bring it back.
func CloseBehaviorAvailable(b settings.CloseBehavior, tray bool) bool {
	return tray || b != settings.CloseHide
}

// CloseBehaviorShown is the choice the dialog shows. A value the settings do not
// know reads as the default, and hiding without a tray reads as the quit the
// close really does, so the dialog never offers a window that cannot come back.
func CloseBehaviorShown(b settings.CloseBehavior, tray bool) settings.CloseBehavior {
	if !b.Valid() {
		b = settings.Defaults().CloseBehavior
	}
	if !CloseBehaviorAvailable(b, tray) {
		return settings.CloseQuit
	}
	return b
}

// ValidateTerminalCommand returns the Italian message for a command that cannot
// run, and "" when it can. A blank command is valid: the Store keeps the default
// one. The check runs before the save, so the message sits under the field and
// the dialog stays open.
func ValidateTerminalCommand(command string) string {
	if strings.TrimSpace(command) == "" {
		return ""
	}
	if _, err := settings.TerminalArgv(command, ""); err != nil {
		return view.UserError(err).Error()
	}
	return ""
}

// StoredSettings is the value the Store keeps from the form: a blank command
// becomes the default one, and the spaces around a command go. The window keeps
// this copy, so what it reads later is what the database holds.
func StoredSettings(v settings.Settings) settings.Settings {
	v.TerminalCommand = strings.TrimSpace(v.TerminalCommand)
	if v.TerminalCommand == "" {
		v.TerminalCommand = settings.DefaultTerminalCommand
	}
	return v
}

// preferencesSaveOutcome turns the result of a save into the message of the
// window and says whether the settings are saved. The refusal of an invalid
// terminal command is the Italian message of UserError.
func preferencesSaveOutcome(err error) (message string, saved bool) {
	if err != nil {
		return view.UserError(err).Error(), false
	}
	return SettingsSavedText, true
}
