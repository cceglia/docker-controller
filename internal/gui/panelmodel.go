package gui

import "dockercontroller/internal/compose"

const (
	// panelCloseText is the button that hides the panel.
	panelCloseText = "Chiudi"
	// activityPageTitle is the title of the page of the Activities.
	activityPageTitle = "Attività"
	// logPageTitle is the title of the page of the logs.
	logPageTitle = "Log"
	// emptyLogText is the line of the Log tab until the ticket that streams it.
	emptyLogText = "Nessun log aperto."
	// panelAccessibleName is what a screen reader calls the bottom panel.
	panelAccessibleName = "Attività e Log"
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
