package compose

import (
	"context"
	"fmt"
	"strings"

	"dockercontroller/internal/compose/model"
	"dockercontroller/internal/compose/parse"
)

// logTail is how many past lines a log stream starts with before following.
const logTail = "500"

// LogRequest and LogLine live in model; the root re-exports them for gui,
// tray and main.
type (
	LogRequest = model.LogRequest
	LogLine    = model.LogLine
)

// logsArgs builds the Docker arguments that follow the logs of a Project.
// Every Profile is passed (like Down does): a Container started under a
// Profile that is not active now still has logs worth reading, and naming its
// Service would otherwise fail. Timestamps are always requested and colours
// never, so the output is parsed one way only.
func logsArgs(p Project, service string) []string {
	args := append(projectArgs(p), "--profile", "*",
		"logs", "--follow", "--tail", logTail, "--timestamps", "--no-color")
	if service != "" {
		args = append(args, service)
	}
	return args
}

// Logs streams the logs of a Project, or of one of its Services: the last 500
// lines, then the new ones as they come, each delivered to onLine already
// attributed to its Service. onLine is never called concurrently. Logs blocks
// until ctx is cancelled, which is the normal way to end a stream and returns
// nil, or until compose itself ends, which returns its failure.
func (m *Manager) Logs(ctx context.Context, req LogRequest, onLine func(LogLine)) error {
	project, ok := m.project(req.Project)
	if !ok {
		return fmt.Errorf("%w: %q", ErrUnknownProject, req.Project)
	}
	if strings.HasPrefix(req.Service, "-") {
		return fmt.Errorf("%w: %q", ErrInvalidService, req.Service)
	}
	owners := logOwners(project)

	err := m.docker.Stream(ctx, logsArgs(project, req.Service), func(line string, stderr bool) {
		if stderr {
			onLine(LogLine{Text: line, Stderr: true})
			return
		}
		onLine(parse.LogLine(line, owners))
	})
	if ctx.Err() != nil {
		return nil
	}
	return err
}

// project returns a copy of the saved Project called name.
func (m *Manager) project(name string) (Project, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.base.Projects {
		if p.Name == name {
			return p, true
		}
	}
	return Project{}, false
}

// logOwners maps what compose can print as a line prefix to a Service: the
// name of a Container (when it was given its own name) or the Service itself.
// The usual prefix, "<service>-<replica>", is resolved by parse.LogLine.
func logOwners(p Project) map[string]string {
	owners := map[string]string{}
	for _, s := range p.Services {
		owners[s.Name] = s.Name
		for _, c := range s.Containers {
			owners[c.Name] = s.Name
		}
	}
	return owners
}
