package gui

import "dockercontroller/internal/settings"

const (
	// preferencesTitle is the title of the settings window.
	preferencesTitle = "Impostazioni"
	// saveSettingsText is the action that stores the settings.
	saveSettingsText = "Salva"
	// settingsSavedText confirms a save of the settings.
	settingsSavedText = "Impostazioni salvate."
	// closeBehaviorGroupTitle and closeBehaviorGroupHint describe the group of
	// the close behavior.
	closeBehaviorGroupTitle = "Alla chiusura della finestra"
	closeBehaviorGroupHint  = "Cosa fa l'app quando chiudi la finestra. Restando nel tray la riapri " +
		"con un click sull'icona; la cronologia delle Attività va persa."
	// terminalGroupTitle and terminalGroupHint describe the group of the
	// terminal command.
	terminalGroupTitle = "Comando del terminale"
	terminalGroupHint  = "{dir} viene sostituito con la cartella del Progetto. " +
		"Vuoto per il comando predefinito."
	// terminalRowTitle names the field of the command.
	terminalRowTitle = "Comando"
)

// closeBehaviorChoice is one option of the close behavior: the label it shows
// and the value it saves.
type closeBehaviorChoice struct {
	label    string
	behavior settings.CloseBehavior
}

// closeBehaviorChoices are the three behaviors, in the order the window shows
// them.
var closeBehaviorChoices = []closeBehaviorChoice{
	{label: "Chiudi la finestra e resta nel tray", behavior: settings.CloseHide},
	{label: "Esci dall'app", behavior: settings.CloseQuit},
	{label: "Chiedi ogni volta", behavior: settings.CloseAsk},
}

// closeBehaviorIndex is the place of a behavior in the choices. A value the
// settings do not know shows the default.
func closeBehaviorIndex(b settings.CloseBehavior) int {
	for i, choice := range closeBehaviorChoices {
		if choice.behavior == b {
			return i
		}
	}
	return 0
}

// closeBehaviorAt is the behavior of a place in the choices. A place out of the
// list is the default.
func closeBehaviorAt(i int) settings.CloseBehavior {
	if i < 0 || i >= len(closeBehaviorChoices) {
		return settings.Defaults().CloseBehavior
	}
	return closeBehaviorChoices[i].behavior
}

// preferencesSaveOutcome turns the result of a save into the message of the
// window and says whether the settings are saved. The refusal of an invalid
// terminal command is the Italian message of UserError.
func preferencesSaveOutcome(err error) (message string, saved bool) {
	if err != nil {
		return UserError(err).Error(), false
	}
	return settingsSavedText, true
}
