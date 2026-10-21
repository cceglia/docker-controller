package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"dockercontroller/internal/compose"
)

// logEvent carries a LogEvent to the frontend.
const logEvent = "log"

// LogEventType tells what a LogEvent reports.
type LogEventType string

const (
	// LogLineEvent delivers one line of a stream.
	LogLineEvent LogEventType = "line"
	// LogEndedEvent is the last event of a stream that ended on its own
	// (compose failed or exited); Error says why. A stream the frontend
	// stopped with StopLogs sends nothing more.
	LogEndedEvent LogEventType = "ended"
)

// LogEvent is one event of the log stream the frontend named ID.
type LogEvent struct {
	ID   string          `json:"id"`
	Type LogEventType    `json:"type"`
	Line compose.LogLine `json:"line"`
	// Error is set for a LogEndedEvent that is a failure.
	Error string `json:"error"`
}

// errLogStopped reports that StopLogs for the id came before StartLogs
// registered it: the stream must not start.
var errLogStopped = errors.New("flusso di log già fermato")

// maxStoppedLogIDs bounds the ids remembered by logStreams.stop.
const maxStoppedLogIDs = 64

// logStreams maps the id of each open log stream to its cancellation.
type logStreams struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
	// stopped holds the latest ids stopped while not registered (oldest first),
	// so a StopLogs that overtakes its StartLogs is not lost. Bounded: the
	// frontend also stops streams that already ended on their own.
	stopped []string
}

// add registers id; it fails when id is already in use. When id was stopped
// before it registered, cancel is called at once and errLogStopped returned.
func (l *logStreams) add(id string, cancel context.CancelFunc) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if i := slices.Index(l.stopped, id); i >= 0 {
		l.stopped = slices.Delete(l.stopped, i, i+1)
		cancel()
		return errLogStopped
	}
	if _, taken := l.cancels[id]; taken {
		return fmt.Errorf("flusso di log già aperto: %s", id)
	}
	if l.cancels == nil {
		l.cancels = map[string]context.CancelFunc{}
	}
	l.cancels[id] = cancel
	return nil
}

// stop cancels the stream and forgets it. An id not registered (yet) is
// remembered, so the stream is cancelled if it registers afterwards.
func (l *logStreams) stop(id string) {
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

// release forgets a stream that ended on its own, without remembering the id.
func (l *logStreams) release(id string) {
	l.mu.Lock()
	cancel := l.cancels[id]
	delete(l.cancels, id)
	l.mu.Unlock()
	if cancel != nil {
		cancel() // frees the context of the stream
	}
}

// StartLogs opens the log stream id (chosen by the frontend, so it can listen
// before the first line arrives) of a Project, or of one of its Services when
// service is not empty. Lines and the end arrive through the "log" event
// until StopLogs, or the end of the application, cancels the stream, which
// terminates the `compose logs` process.
func (s *ProjectsService) StartLogs(id, project, service string) error {
	ctx, cancel := context.WithCancel(s.ctx)
	if err := s.logs.add(id, cancel); err != nil {
		cancel()
		if errors.Is(err, errLogStopped) {
			return nil // stopped before it started: nothing to run
		}
		return err
	}
	go func() {
		defer s.logs.release(id)
		err := s.manager.Logs(ctx, compose.LogRequest{Project: project, Service: service}, func(l compose.LogLine) {
			s.emit(logEvent, LogEvent{ID: id, Type: LogLineEvent, Line: l})
		})
		if ctx.Err() == nil { // not stopped: tell the frontend why it ended
			e := LogEvent{ID: id, Type: LogEndedEvent}
			if err != nil {
				e.Error = userError(err).Error()
			}
			s.emit(logEvent, e)
		}
	}()
	return nil
}

// StopLogs stops the log stream id and terminates its `compose logs` process.
// An unknown id (already ended) is ignored.
func (s *ProjectsService) StopLogs(id string) {
	s.logs.stop(id)
}
