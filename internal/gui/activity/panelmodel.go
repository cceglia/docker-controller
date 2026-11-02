package activity

import "dockercontroller/internal/compose"

const (
	// PanelCloseText is the button that hides the panel.
	PanelCloseText = "Chiudi"
	// ActivityPageTitle is the title of the page of the Activities.
	ActivityPageTitle = "Attività"
	// LogPageTitle is the title of the page of the logs.
	LogPageTitle = "Log"
	// EmptyLogText is the line of the Log tab before a Project is open.
	EmptyLogText = "Nessun log aperto. Scegli un Progetto o apri i log di un Servizio."
	// PanelAccessibleName is what a screen reader calls the bottom panel.
	PanelAccessibleName = "Attività e Log"
)

// The texts of the bar and of the page of the logs.
const (
	// LogSearchLabelText is the visible label of the search entry.
	LogSearchLabelText = "Cerca"
	// LogTimestampsText is the label of the switch of the clocks.
	LogTimestampsText = "Timestamp"
	// LogPauseText and LogResumeText are the two faces of the pause button.
	LogPauseText  = "Pausa"
	LogResumeText = "Riprendi"
	// LogWaitingText is the line before the first line arrives, and
	// LogNoMatchText the line when the search hides every line.
	LogWaitingText = "In attesa di righe…"
	LogNoMatchText = "Nessuna riga corrisponde alla ricerca."
	// LogEndedText is the line under the text when compose closed the stream.
	LogEndedText = "Il flusso dei log è terminato."
	// ShowLogsText is the button of the detail that opens the logs of a whole
	// Project, and ServiceLogText the one of a single Service.
	ShowLogsText   = "Mostra log"
	ServiceLogText = "Log"
)

// PanelActivity returns the Activity the bottom panel must open on after a
// change: the newest Activity of the history it has not shown yet. It reports
// false when every Activity of the history is already shown, so an output line
// never reopens the panel. It reads no widget, so a test drives it.
func PanelActivity(shown map[compose.ActivityID]bool, activities []Activity) (compose.ActivityID, bool) {
	for _, a := range activities {
		if !shown[a.ID] {
			return a.ID, true
		}
	}
	return "", false
}
