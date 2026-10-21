package compose

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ActionKind is what an Action does. Down and Down with volumes (ticket 05)
// join the table below; nothing else in the Activity machinery has to change.
type ActionKind string

const (
	ActionUp ActionKind = "up"
	// ActionStart starts existing Containers; it is meant for a Service.
	ActionStart   ActionKind = "start"
	ActionStop    ActionKind = "stop"
	ActionRestart ActionKind = "restart"
	ActionPull    ActionKind = "pull"
	ActionBuild   ActionKind = "build"
)

// verbs maps every ActionKind to the compose arguments that follow the
// project selection.
var verbs = map[ActionKind][]string{
	ActionUp:      {"up", "-d"},
	ActionStart:   {"start"},
	ActionStop:    {"stop"},
	ActionRestart: {"restart"},
	ActionPull:    {"pull"},
	ActionBuild:   {"up", "-d", "--build"},
}

// Action is a request to run an ActionKind on a Project, or on one of its
// Services. Ticket 08 adds the active Profiles; it only extends composeArgs.
type Action struct {
	// Project is the compose name of the Project.
	Project string `json:"project"`
	// Service, when not empty, restricts the Action to that Service (all its
	// Containers). It is still an Activity of the whole Project.
	Service string     `json:"service"`
	Kind    ActionKind `json:"kind"`
}

// ActivityID identifies one Activity (one execution of an Action).
type ActivityID string

// ActivityRef is the running Activity of a Project, as shown in the snapshot.
type ActivityRef struct {
	ID ActivityID `json:"id"`
	// Service is the Service the Activity targets, "" for the whole Project.
	Service string     `json:"service"`
	Kind    ActionKind `json:"kind"`
}

// ActivityEventType tells what an ActivityEvent reports.
type ActivityEventType string

const (
	// ActivityStarted is the first event of an Activity.
	ActivityStarted ActivityEventType = "started"
	// ActivityLine delivers one line of output.
	ActivityLine ActivityEventType = "line"
	// ActivityFinished is the last event: the outcome. Error is empty on
	// success; otherwise it holds the stderr of compose.
	ActivityFinished ActivityEventType = "finished"
)

// ActivityEvent is delivered to Config.OnActivity, in order for each Activity
// and never concurrently.
type ActivityEvent struct {
	ID      ActivityID        `json:"id"`
	Type    ActivityEventType `json:"type"`
	Project string            `json:"project"`
	// Service is the Service the Action targets, "" for the whole Project.
	Service string     `json:"service"`
	Kind    ActionKind `json:"kind"`
	// Line and Stderr are set for ActivityLine events.
	Line   string `json:"line"`
	Stderr bool   `json:"stderr"`
	// Error is set for a failed ActivityFinished event.
	Error string `json:"error"`
}

var (
	// ErrActivityRunning is returned when the Project already has an Activity.
	ErrActivityRunning = errors.New("compose: the Project already has an Activity in progress")
	// ErrUnknownProject is returned for a Project that is not saved.
	ErrUnknownProject = errors.New("compose: unknown Project")
	// ErrUnknownAction is returned for an ActionKind that does not exist.
	ErrUnknownAction = errors.New("compose: unknown Action")
	// ErrInvalidService is returned for a Service name compose could take for
	// an option.
	ErrInvalidService = errors.New("compose: invalid Service name")
)

// composeArgs builds the Docker arguments of an Action on a Project.
func composeArgs(p Project, a Action) ([]string, error) {
	verb, ok := verbs[a.Kind]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownAction, a.Kind)
	}
	if strings.HasPrefix(a.Service, "-") {
		return nil, fmt.Errorf("%w: %q", ErrInvalidService, a.Service)
	}
	args := []string{"compose", "-p", p.Name}
	for _, f := range p.ComposeFiles {
		args = append(args, "-f", f)
	}
	args = append(args, verb...)
	if a.Service != "" {
		args = append(args, a.Service)
	}
	return args, nil
}

// Do starts an Action on a Project and returns the id of its Activity without
// waiting for it. Output and outcome are delivered to Config.OnActivity; the
// snapshot shows the Activity while it runs and is refreshed when it ends.
// Cancelling ctx stops the Activity. A Project accepts one Activity at a time
// (ErrActivityRunning); different Projects run in parallel.
func (m *Manager) Do(ctx context.Context, a Action) (ActivityID, error) {
	m.mu.Lock()
	var project *Project
	for i := range m.base.Projects {
		if m.base.Projects[i].Name == a.Project {
			project = &m.base.Projects[i]
			break
		}
	}
	if project == nil {
		m.mu.Unlock()
		return "", fmt.Errorf("%w: %q", ErrUnknownProject, a.Project)
	}
	args, err := composeArgs(*project, a)
	if err != nil {
		m.mu.Unlock()
		return "", err
	}
	if _, busy := m.running[a.Project]; busy {
		m.mu.Unlock()
		return "", fmt.Errorf("%w: %s", ErrActivityRunning, a.Project)
	}
	m.nextID++
	id := ActivityID(fmt.Sprintf("act-%d", m.nextID))
	m.running[a.Project] = ActivityRef{ID: id, Service: a.Service, Kind: a.Kind}
	m.recompose()
	m.mu.Unlock()

	m.emit(ActivityEvent{ID: id, Type: ActivityStarted, Project: a.Project, Service: a.Service, Kind: a.Kind})
	go m.runActivity(ctx, id, a, args)
	return id, nil
}

func (m *Manager) runActivity(ctx context.Context, id ActivityID, a Action, args []string) {
	var stderr []string
	err := m.docker.Stream(ctx, args, func(line string, isErr bool) {
		if isErr {
			stderr = append(stderr, line)
		}
		m.emit(ActivityEvent{ID: id, Type: ActivityLine, Project: a.Project, Service: a.Service, Kind: a.Kind, Line: line, Stderr: isErr})
	})

	outcome := ActivityEvent{ID: id, Type: ActivityFinished, Project: a.Project, Service: a.Service, Kind: a.Kind}
	if err != nil {
		outcome.Error = strings.TrimSpace(strings.Join(stderr, "\n"))
		if outcome.Error == "" || ctx.Err() != nil {
			outcome.Error = err.Error()
		}
	}

	// The Activity is over: free the Project and show what Docker now says in
	// one notification. A failed refresh still has to clear the Activity.
	m.mu.Lock()
	delete(m.running, a.Project)
	m.mu.Unlock()
	if rerr := m.Refresh(context.WithoutCancel(ctx)); rerr != nil {
		m.mu.Lock()
		m.recompose()
		m.mu.Unlock()
	}
	m.emit(outcome)
}

// emit delivers an event; deliveries are serialised.
func (m *Manager) emit(e ActivityEvent) {
	if m.onActivity == nil {
		return
	}
	m.emitMu.Lock()
	defer m.emitMu.Unlock()
	m.onActivity(e)
}
