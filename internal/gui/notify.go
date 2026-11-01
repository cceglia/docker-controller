package gui

import (
	"sync"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/notify"
	"dockercontroller/internal/tray"
)

// Notifier turns Activity events into desktop notifications. It feeds
// tray.Notifier, which decides what deserves one, names the Project as the list
// does and hands the result to the Sender. It has no Qt, so a test drives it
// with synthetic events and a Sender that is not D-Bus. It is safe for
// concurrent use: the Manager delivers the events and the Qt thread writes the
// names of the last snapshot.
type Notifier struct {
	// policy decides which events deserve a notification.
	policy tray.Notifier
	sender notify.Sender
	// mu guards labels, which the Qt thread writes and the Manager reads.
	mu     sync.Mutex
	labels map[string]string // Display name by Project
}

// NewNotifier returns a Notifier that shows through sender.
func NewNotifier(sender notify.Sender) *Notifier {
	return &Notifier{sender: sender, labels: map[string]string{}}
}

// OnClick registers the function that runs when the user clicks one of the
// notifications of this Notifier. A Sender that cannot report a click ignores
// the call: the notification shows and the click does nothing.
func (n *Notifier) OnClick(fn func()) {
	if clickable, ok := n.sender.(notify.Clickable); ok {
		clickable.SetOnClick(fn)
	}
}

// SetProjects records the Display names of the last snapshot, so a notification
// names a Project as the list does. A Project without a Display name keeps its
// compose name.
func (n *Notifier) SetProjects(projects []compose.Project) {
	labels := make(map[string]string, len(projects))
	for _, project := range projects {
		if project.DisplayName != "" {
			labels[project.Name] = project.DisplayName
		}
	}
	n.mu.Lock()
	n.labels = labels
	n.mu.Unlock()
}

// launch runs an Action chosen in the tray through start and announces it to the
// Notifier, so its end is notified. A Manager that refuses the Action withdraws
// the announcement and gets a notification of its own, because a refused Action
// never becomes an Activity and the Activity rules would never report it. The
// outcome of the Action is unchanged: it comes back to the caller.
func (n *Notifier) launch(action compose.Action, start func(compose.Action) (compose.ActivityID, error)) error {
	n.policy.Launching(action)
	if _, err := start(action); err != nil {
		n.policy.Aborted(action)
		n.sender.Send(windowTitle, UserError(err).Error())
		return err
	}
	return nil
}

// Observe feeds one Activity event and shows the notification it deserves, if
// any. Observe returns at once: the Sender sends in the background.
func (n *Notifier) Observe(e compose.ActivityEvent) {
	notification, ok := n.policy.Observe(e, n.label(e.Project))
	if ok {
		n.sender.Send(notification.Title, notification.Body)
	}
}

// label is the Display name of a Project, or its compose name.
func (n *Notifier) label(project string) string {
	n.mu.Lock()
	defer n.mu.Unlock()
	if name := n.labels[project]; name != "" {
		return name
	}
	return project
}
