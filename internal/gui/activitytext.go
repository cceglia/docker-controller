package gui

const (
	// emptyActivityText is the line of the page before the first Activity.
	emptyActivityText = "Nessuna Attività in questa sessione."
	// runningActivityText is the text of a row while its Activity runs.
	runningActivityText = "Attività in corso."
)

// activityTooltip is what the row says on hover: the outcome of a finished
// Activity, nothing more than its state while it runs.
func activityTooltip(a Activity) string {
	switch a.State {
	case ActivityOK:
		return ActivitySuccessText(a)
	case ActivityFailed:
		return ActivityToastText(a)
	}
	return runningActivityText
}
