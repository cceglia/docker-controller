package activity

const (
	// EmptyText is the line of the page before the first Activity.
	EmptyText = "Nessuna Attività in questa sessione."
	// RunningText is the text of a row while its Activity runs.
	RunningText = "Attività in corso."
)

// Tooltip is what the row says on hover: the outcome of a finished
// Activity, nothing more than its state while it runs.
func Tooltip(a Activity) string {
	switch a.State {
	case ActivityOK:
		return ActivitySuccessText(a)
	case ActivityFailed:
		return ActivityToastText(a)
	}
	return RunningText
}
