package logs

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"dockercontroller/internal/compose"
)

// LogSource opens the log stream of a Project, or of one of its Services, and
// delivers each line to onLine, which is never called concurrently. It blocks
// until ctx is cancelled or until compose ends the stream.
// compose.Manager.Logs satisfies it.
type LogSource func(ctx context.Context, req compose.LogRequest, onLine func(compose.LogLine)) error

// LogEnd tells how a stream ended by itself.
type LogEnd struct {
	// Stopped is true when the view stopped the stream: no line follows.
	Stopped bool
	// Error is the reason compose ended the stream, in Italian, and is empty
	// when it ended without a failure.
	Error string
}

// ErrLogStopped reports that a stream was stopped before it registered: it
// must not start.
var ErrLogStopped = errors.New("log stream already stopped")

// ErrLogOpen reports a stream id that is already in use.
var ErrLogOpen = errors.New("log stream already open")

// ErrLogsClosed reports that StopAll ran: the window is gone and no stream may
// start until Resume.
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
	// so a Stop that overtakes its Add is not lost. Bounded: the view also
	// stops streams that already ended on their own.
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
// new stream starts, so an opening still in flight cannot outlive the window.
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

// maxEndedStderrLines bounds the stderr lines kept to explain how a stream
// ended.
const maxEndedStderrLines = 20

// appendLastLines appends line to lines, keeping only the last n.
func appendLastLines(lines []string, line string, n int) []string {
	lines = append(lines, line)
	if len(lines) > n {
		lines = slices.Delete(lines, 0, len(lines)-n)
	}
	return lines
}

// StreamLogs runs the stream and returns how it ended. When the view stopped
// it with ctx, the result says so and carries no error: a stop is not a
// failure. When compose ended it on its own, the result carries the Italian
// reason, with the last stderr lines of compose as detail.
func StreamLogs(
	ctx context.Context,
	req compose.LogRequest,
	source LogSource,
	onLine func(compose.LogLine),
) LogEnd {
	var stderr []string // only touched by the callback, which is never concurrent
	err := source(ctx, req, func(l compose.LogLine) {
		if l.Stderr {
			stderr = appendLastLines(stderr, l.Text, maxEndedStderrLines)
		}
		onLine(l)
	})
	if ctx.Err() != nil {
		return LogEnd{Stopped: true}
	}
	end := LogEnd{}
	if err != nil {
		end.Error = logEndedMessage(err, stderr)
	}
	return end
}
