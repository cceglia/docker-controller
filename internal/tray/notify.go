package tray

import (
	"fmt"
	"sync"

	"dockercontroller/internal/compose"
)

// Notification is a desktop notification to show.
type Notification struct {
	Title string
	Body  string
}

// Notifier decides which Activity events deserve a notification: the end of
// every Activity launched from the tray, and every failed Activity whatever
// its origin. It is safe for concurrent use.
//
// The id of an Activity is only known once compose.Manager.Do returns, but its
// "started" event is delivered before that and the "finished" one may be too.
// So the tray announces a launch with Launching and the Notifier claims the
// next matching "started" event.
type Notifier struct {
	mu      sync.Mutex
	pending map[string]compose.Action // by Project: launched, id not seen yet
	tray    map[compose.ActivityID]struct{}
}

// Launching must be called just before the tray calls Manager.Do.
func (n *Notifier) Launching(a compose.Action) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.pending == nil {
		n.pending = map[string]compose.Action{}
	}
	n.pending[a.Project] = a
}

// Aborted must be called when Manager.Do refused the Action.
func (n *Notifier) Aborted(a compose.Action) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.pending[a.Project] == a {
		delete(n.pending, a.Project)
	}
}

// Observe feeds an Activity event; name is the Project's label for messages.
// It reports the notification to show, if any.
func (n *Notifier) Observe(e compose.ActivityEvent, name string) (Notification, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	switch e.Type {
	case compose.ActivityStarted:
		a, ok := n.pending[e.Project]
		if ok && a.Service == e.Service && a.Kind == e.Kind {
			delete(n.pending, e.Project)
			if n.tray == nil {
				n.tray = map[compose.ActivityID]struct{}{}
			}
			n.tray[e.ID] = struct{}{}
		}
	case compose.ActivityFinished:
		_, fromTray := n.tray[e.ID]
		delete(n.tray, e.ID)
		if !fromTray && e.Error == "" {
			return Notification{}, false
		}
		return Notification{Title: "Docker Controller", Body: body(e, name)}, true
	}
	return Notification{}, false
}

func body(e compose.ActivityEvent, name string) string {
	target := name
	if e.Service != "" {
		target = fmt.Sprintf("%s (%s)", e.Service, name)
	}
	what := fmt.Sprintf("%s di %s", kindLabel(e.Kind), target)
	if e.Error == "" {
		return what + " completato"
	}
	if e.Summary == "" {
		return what + " fallito"
	}
	return what + " fallito: " + e.Summary
}

func kindLabel(k compose.ActionKind) string {
	switch k {
	case compose.ActionUp:
		return "Up"
	case compose.ActionStart:
		return "Start"
	case compose.ActionStop:
		return "Stop"
	case compose.ActionRestart:
		return "Restart"
	case compose.ActionPull:
		return "Pull"
	case compose.ActionBuild:
		return "Build"
	case compose.ActionDown:
		return "Down"
	case compose.ActionDownVolumes:
		return "Down con volumi"
	}
	return string(k)
}
