package gui

import "dockercontroller/internal/compose"

const (
	// panelCloseText is the button that hides the panel.
	panelCloseText = "Chiudi"
	// activityPageTitle is the title of the page of the Activities.
	activityPageTitle = "Attività"
	// logPageTitle is the title of the page of the logs.
	logPageTitle = "Log"
	// emptyLogText is the line of the Log tab before a Project is open.
	emptyLogText = "Nessun log aperto. Scegli un Progetto o apri i log di un Servizio."
	// panelAccessibleName is what a screen reader calls the bottom panel.
	panelAccessibleName = "Attività e Log"
)

// The texts of the bar and of the page of the logs.
const (
	// logSearchLabelText is the visible label of the search entry.
	logSearchLabelText = "Cerca"
	// logTimestampsText is the label of the switch of the clocks.
	logTimestampsText = "Timestamp"
	// logPauseText and logResumeText are the two faces of the pause button.
	logPauseText  = "Pausa"
	logResumeText = "Riprendi"
	// logWaitingText is the line before the first line arrives, and
	// logNoMatchText the line when the search hides every line.
	logWaitingText = "In attesa di righe…"
	logNoMatchText = "Nessuna riga corrisponde alla ricerca."
	// logEndedText is the line under the text when compose closed the stream.
	logEndedText = "Il flusso dei log è terminato."
	// showLogsText is the button of the detail that opens the logs of a whole
	// Project, and serviceLogText the one of a single Service.
	showLogsText   = "Mostra log"
	serviceLogText = "Log"
)

// panelActivity returns the Activity the bottom panel must open on after a
// change: the newest Activity of the history it has not shown yet. It reports
// false when every Activity of the history is already shown, so an output line
// never reopens the panel. It reads no widget, so a test drives it.
func panelActivity(shown map[compose.ActivityID]bool, activities []Activity) (compose.ActivityID, bool) {
	for _, a := range activities {
		if !shown[a.ID] {
			return a.ID, true
		}
	}
	return "", false
}
