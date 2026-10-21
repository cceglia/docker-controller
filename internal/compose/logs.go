package compose

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// logTail is how many past lines a log stream starts with before following.
const logTail = "500"

// LogRequest asks for the log stream of a Project, or of one of its Services.
type LogRequest struct {
	// Project is the compose name of the Project.
	Project string `json:"project"`
	// Service, when not empty, restricts the stream to that Service; compose
	// is asked for it, the others are not read at all.
	Service string `json:"service"`
}

// LogLine is one line of a log stream, already attributed to its Service.
type LogLine struct {
	// Service is the Service whose Container wrote the line. It is empty for
	// a line that did not come from a Container (compose's own messages) and
	// holds the raw compose prefix when it matches no known Service.
	Service string `json:"service"`
	// Timestamp is the time Docker recorded for the line (RFC 3339, as
	// compose prints it); empty when the line carries none. It is always
	// requested: whether to show it is up to the caller.
	Timestamp string `json:"timestamp"`
	// Text is the line without prefix, timestamp or terminal colour codes.
	Text string `json:"text"`
	// Stderr marks a message of compose itself (for example a Service that
	// does not exist); Container output always arrives as not Stderr.
	Stderr bool `json:"stderr"`
}

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
		onLine(parseLogLine(line, owners))
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
// The usual prefix, "<service>-<replica>", is resolved by parseLogLine.
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

// ansi matches the terminal escape sequences Containers write for colours.
var ansi = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

// parseLogLine splits "<prefix> | <timestamp> <text>".
func parseLogLine(line string, owners map[string]string) LogLine {
	prefix, rest, found := strings.Cut(line, " | ")
	prefix = strings.TrimRight(prefix, " ")
	if !found || prefix == "" || strings.ContainsAny(prefix, " \t") {
		return LogLine{Text: ansi.ReplaceAllString(line, "")}
	}
	out := LogLine{Service: serviceOf(prefix, owners), Text: rest}
	ts, text, hasText := strings.Cut(rest, " ")
	if _, err := time.Parse(time.RFC3339Nano, ts); err == nil {
		out.Timestamp = ts
		out.Text = ""
		if hasText {
			out.Text = text
		}
	}
	out.Text = ansi.ReplaceAllString(out.Text, "")
	return out
}

// serviceOf resolves a compose line prefix to a Service name.
func serviceOf(prefix string, owners map[string]string) string {
	if s, ok := owners[prefix]; ok {
		return s
	}
	if i := strings.LastIndexByte(prefix, '-'); i > 0 && isNumber(prefix[i+1:]) {
		if s, ok := owners[prefix[:i]]; ok {
			return s
		}
		// A Service of a Project Docker no longer shows: the replica number
		// is not part of its name.
		return prefix[:i]
	}
	return prefix
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
