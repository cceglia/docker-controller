package gui

import (
	"strings"
	"sync"

	"dockercontroller/internal/compose"
)

// ActivityLineLimit is the most output lines kept for one Activity. The old
// view had no cap on the output of an Activity; 5000 is the cap this one uses,
// the same number the log panel keeps.
const ActivityLineLimit = 5000

// ActivityState is where an Activity stands.
type ActivityState string

const (
	// ActivityRunning is an Activity whose command is still running.
	ActivityRunning ActivityState = "running"
	// ActivityOK is an Activity that ended with success.
	ActivityOK ActivityState = "ok"
	// ActivityFailed is an Activity that ended with an error.
	ActivityFailed ActivityState = "error"
)

// ActivityLine is one line of the output of an Activity.
type ActivityLine struct {
	Text   string
	Stderr bool
}

// Activity is one Activity of the session, as the bottom panel shows it.
type Activity struct {
	ID      compose.ActivityID
	Project string
	// Service is the Service the Action targets, "" for the whole Project.
	Service string
	Kind    compose.ActionKind
	State   ActivityState
	// Error is the stderr of a failed Activity, "" otherwise.
	Error string
	// Summary is the one line of Error the backend chose for the toast.
	Summary string
	Lines   []ActivityLine
	// Dropped is how many lines the cap removed from the front of Lines, so a
	// view that appends only the new lines can tell they moved.
	Dropped int
}

// ActivityTarget names what an Action runs on: the Project, or one of its
// Services as "project/service".
func ActivityTarget(project, service string) string {
	if service == "" {
		return project
	}
	return project + "/" + service
}

// ActivityLabel is the Action and its target: "Up su shop/db".
func ActivityLabel(a Activity) string {
	return ActionLabel(a.Kind) + " su " + ActivityTarget(a.Project, a.Service)
}

// ActivityGlyph is the marker of a state, the one the old view showed.
func ActivityGlyph(state ActivityState) string {
	switch state {
	case ActivityRunning:
		return "⏳"
	case ActivityOK:
		return "✓"
	case ActivityFailed:
		return "✗"
	}
	return ""
}

// ActivityHeader is the row of the history: the state, the Action and its
// target, as "⏳ Up · shop/db".
func ActivityHeader(a Activity) string {
	row := ActionLabel(a.Kind) + " · " + ActivityTarget(a.Project, a.Service)
	if glyph := ActivityGlyph(a.State); glyph != "" {
		return glyph + " " + row
	}
	return row
}

// ActivitySuccessText is the outcome of a finished Activity that worked.
func ActivitySuccessText(a Activity) string {
	return ActivityLabel(a) + " riuscito."
}

// ActivityFailureText is the outcome of a failed Activity: the error compose
// printed, without its last newline.
func ActivityFailureText(a Activity) string {
	return strings.TrimRight(a.Error, "\n")
}

// ActivityToastText is the message of the toast of a failed Activity, with the
// one-line Summary the backend chose. It is the message of the old view.
func ActivityToastText(a Activity) string {
	message := ActivityLabel(a) + " non riuscito"
	if a.Summary != "" {
		message += ": " + a.Summary
	}
	return message
}

// ActivityOutcome is the line under the output of a finished Activity: the
// positive result or the error. It is "" while the Activity runs.
func ActivityOutcome(a Activity) string {
	switch a.State {
	case ActivityOK:
		return ActivitySuccessText(a)
	case ActivityFailed:
		return ActivityFailureText(a)
	}
	return ""
}

// linesText is the text of some output lines, one line each.
func linesText(lines []ActivityLine) string {
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(line.Text)
		b.WriteString("\n")
	}
	return b.String()
}

// ActivityOutput is the whole text of the output of an Activity: its lines and,
// when it is over, the outcome on the last line. The outcome of a failure is the
// stderr of compose, which the Manager already streamed as lines: the message
// must not appear twice.
func ActivityOutput(a Activity) string {
	text := linesText(a.Lines)
	outcome := ActivityOutcome(a)
	if outcome == "" {
		return text
	}
	if strings.HasSuffix(strings.TrimRight(text, "\n"), strings.TrimSpace(outcome)) {
		return text
	}
	return text + outcome + "\n"
}

// ActivityLog keeps the Activities of the session, newest first, with their
// output and outcome. It is the single source of the panel: the Manager pushes
// its events into it, the view reads it and subscribes. It has no GTK, so a
// test drives it with synthetic events.
//
// The log never stops an Activity and never waits for the view: Push holds the
// lock only to change the history, and calls the subscriber after unlocking.
// The application clears it when it builds a new window, so the history of a
// destroyed window never comes back.
type ActivityLog struct {
	mu         sync.Mutex
	activities []Activity // newest first
	index      map[compose.ActivityID]int
	onChange   func()
}

// NewActivityLog creates an empty log.
func NewActivityLog() *ActivityLog {
	return &ActivityLog{index: map[compose.ActivityID]int{}}
}

// Push appends the event to the history and returns at once. A Started event
// puts a new Activity at the front; a Line appends to its output, keeping the
// last ActivityLineLimit lines; a Finished event sets its state and its error.
// An event for an Activity the log does not know changes nothing. The
// subscriber, when set, runs after a change, on the goroutine that pushed.
func (l *ActivityLog) Push(e compose.ActivityEvent) {
	changed := false

	l.mu.Lock()
	switch e.Type {
	case compose.ActivityStarted:
		if _, known := l.index[e.ID]; !known {
			l.activities = append([]Activity{{
				ID:      e.ID,
				Project: e.Project,
				Service: e.Service,
				Kind:    e.Kind,
				State:   ActivityRunning,
			}}, l.activities...)
			l.reindex()
			changed = true
		}
	case compose.ActivityLine:
		if i, known := l.index[e.ID]; known {
			a := &l.activities[i]
			a.Lines = append(a.Lines, ActivityLine{Text: e.Line, Stderr: e.Stderr})
			if over := len(a.Lines) - ActivityLineLimit; over > 0 {
				// Drop the oldest lines in place, so the history keeps the
				// memory it already has.
				n := copy(a.Lines, a.Lines[over:])
				a.Lines = a.Lines[:n]
				a.Dropped += over
			}
			changed = true
		}
	case compose.ActivityFinished:
		if i, known := l.index[e.ID]; known {
			a := &l.activities[i]
			a.State = ActivityOK
			a.Error = ""
			a.Summary = ""
			if e.Error != "" {
				a.State = ActivityFailed
				a.Error = e.Error
				a.Summary = e.Summary
			}
			changed = true
		}
	}
	notify := l.onChange
	l.mu.Unlock()

	if changed && notify != nil {
		notify()
	}
}

// Clear empties the history: a window opens with no Activity, as the dialog of
// the close says. The subscriber, when set, runs after the clear, so a view
// still on screen shows nothing. An Activity that ran before the clear stays
// out: its later events name an ID the log no longer knows.
func (l *ActivityLog) Clear() {
	l.mu.Lock()
	l.activities = nil
	l.index = map[compose.ActivityID]int{}
	notify := l.onChange
	l.mu.Unlock()

	if notify != nil {
		notify()
	}
}

// Activities returns the Activities of the session in display order, newest
// first. The view owns the result: it is a copy.
func (l *ActivityLog) Activities() []Activity {
	l.mu.Lock()
	defer l.mu.Unlock()

	out := make([]Activity, len(l.activities))
	for i, a := range l.activities {
		out[i] = copyActivity(a)
	}
	return out
}

// Activity returns one Activity of the session. The view owns the result.
func (l *ActivityLog) Activity(id compose.ActivityID) (Activity, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	i, ok := l.index[id]
	if !ok {
		return Activity{}, false
	}
	return copyActivity(l.activities[i]), true
}

// OnChange registers the subscriber the log calls after every change. The
// subscriber must return at once: the Manager waits for Push, not for the
// view. A nil function removes the subscription.
func (l *ActivityLog) OnChange(fn func()) {
	l.mu.Lock()
	l.onChange = fn
	l.mu.Unlock()
}

// reindex rebuilds the position of each Activity. The caller holds the lock.
func (l *ActivityLog) reindex() {
	l.index = make(map[compose.ActivityID]int, len(l.activities))
	for i, a := range l.activities {
		l.index[a.ID] = i
	}
}

// copyActivity copies an Activity together with its output.
func copyActivity(a Activity) Activity {
	a.Lines = append([]ActivityLine(nil), a.Lines...)
	return a
}
