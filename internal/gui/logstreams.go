package gui

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"dockercontroller/internal/compose"
)

// LogEventType tells what a LogEvent reports.
type LogEventType string

const (
	// LogLineEvent delivers one line of a stream.
	LogLineEvent LogEventType = "line"
	// LogEndedEvent is the last event of a stream that ended on its own
	// (compose failed or exited); Error says why. A stream the view stopped
	// with StopLogs sends nothing more.
	LogEndedEvent LogEventType = "ended"
)

// LogEvent is one event of the log stream the view named ID.
type LogEvent struct {
	ID   string          `json:"id"`
	Type LogEventType    `json:"type"`
	Line compose.LogLine `json:"line"`
	// Error is set for a LogEndedEvent that is a failure.
	Error string `json:"error"`
}

// ErrLogStopped reports that StopLogs for the id came before StartLogs
// registered it: the stream must not start.
var ErrLogStopped = errors.New("log stream already stopped")

// ErrLogOpen reports a stream id that is already in use.
var ErrLogOpen = errors.New("log stream already open")

// ErrLogsClosed reports that StopAllLogs ran: the window is gone and no
// stream may start until resume.
var ErrLogsClosed = errors.New("log streams stopped with the window")

// maxStoppedLogIDs bounds the ids remembered by LogStreams.Stop.
const maxStoppedLogIDs = 64

// LogStreams maps the id of each open log stream to its cancellation.
type LogStreams struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
	// closed is set by StopAll: no stream starts until the window is created
	// again and Resume runs.
	closed bool
	// stopped holds the latest ids stopped while not registered (oldest first),
	// so a StopLogs that overtakes its StartLogs is not lost. Bounded: the
	// view also stops streams that already ended on their own.
	stopped []string
}

// Add registers id; it fails when id is already in use. When id was stopped
// before it registered, cancel is called at once and ErrLogStopped returned.
// When every stream was stopped with the window, cancel is called at once and
// ErrLogsClosed returned.
func (l *LogStreams) Add(id string, cancel context.CancelFunc) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		cancel()
		return ErrLogsClosed
	}
	if i := slices.Index(l.stopped, id); i >= 0 {
		l.stopped = slices.Delete(l.stopped, i, i+1)
		cancel()
		return ErrLogStopped
	}
	if _, taken := l.cancels[id]; taken {
		return fmt.Errorf("%w: %s", ErrLogOpen, id)
	}
	if l.cancels == nil {
		l.cancels = map[string]context.CancelFunc{}
	}
	l.cancels[id] = cancel
	return nil
}

// Stop cancels the stream and forgets it. An id not registered (yet) is
// remembered, so the stream is cancelled if it registers afterwards.
func (l *LogStreams) Stop(id string) {
	l.mu.Lock()
	cancel, open := l.cancels[id]
	delete(l.cancels, id)
	if !open && !slices.Contains(l.stopped, id) {
		l.stopped = append(l.stopped, id)
		if len(l.stopped) > maxStoppedLogIDs {
			l.stopped = l.stopped[len(l.stopped)-maxStoppedLogIDs:]
		}
	}
	l.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// Release forgets a stream that ended on its own, without remembering the id.
func (l *LogStreams) Release(id string) {
	l.mu.Lock()
	cancel := l.cancels[id]
	delete(l.cancels, id)
	l.mu.Unlock()
	if cancel != nil {
		cancel() // frees the context of the stream
	}
}

// StopAll cancels every open stream and forgets it. The ids are not
// remembered: they belong to the window that is going away. Until Resume, no
// new stream starts, so a StartLogs still in flight cannot outlive the window.
func (l *LogStreams) StopAll() {
	l.mu.Lock()
	l.closed = true
	cancels := l.cancels
	l.cancels = nil
	l.mu.Unlock()
	for _, cancel := range cancels {
		cancel()
	}
}

// Resume allows new streams after the window is created again.
func (l *LogStreams) Resume() {
	l.mu.Lock()
	l.closed = false
	l.mu.Unlock()
}

// maxEndedStderrLines bounds the stderr lines kept to explain how a stream ended.
const maxEndedStderrLines = 20

// appendLastLines appends line to lines, keeping only the last n.
func appendLastLines(lines []string, line string, n int) []string {
	lines = append(lines, line)
	if len(lines) > n {
		lines = slices.Delete(lines, 0, len(lines)-n)
	}
	return lines
}

// StreamLogs runs the stream, emitting each line, and when it ends on its own
// (ctx not cancelled) a LogEndedEvent that explains why, with the last stderr
// lines of compose as detail.
func StreamLogs(
	ctx context.Context,
	id string,
	req compose.LogRequest,
	logs func(context.Context, compose.LogRequest, func(compose.LogLine)) error,
	emit func(LogEvent),
) {
	var stderr []string // only touched by the callback, which is never concurrent
	err := logs(ctx, req, func(l compose.LogLine) {
		if l.Stderr {
			stderr = appendLastLines(stderr, l.Text, maxEndedStderrLines)
		}
		emit(LogEvent{ID: id, Type: LogLineEvent, Line: l})
	})
	if ctx.Err() != nil {
		return // stopped: nothing more is sent
	}
	e := LogEvent{ID: id, Type: LogEndedEvent}
	if err != nil {
		e.Error = logEndedMessage(err, stderr)
	}
	emit(e)
}
