package main

import (
	"context"
	"fmt"
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

// logStreams maps the id of each open log stream to its cancellation.
type logStreams struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

// add registers id; it fails when id is already in use.
func (l *logStreams) add(id string, cancel context.CancelFunc) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, taken := l.cancels[id]; taken {
		return fmt.Errorf("flusso di log già aperto: %s", id)
	}
	if l.cancels == nil {
		l.cancels = map[string]context.CancelFunc{}
	}
	l.cancels[id] = cancel
	return nil
}

// stop cancels the stream and forgets it; an unknown id is ignored.
func (l *logStreams) stop(id string) {
	l.mu.Lock()
	cancel := l.cancels[id]
	delete(l.cancels, id)
	l.mu.Unlock()
	if cancel != nil {
		cancel()
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
		return err
	}
	go func() {
		defer s.logs.stop(id)
		err := s.manager.Logs(ctx, compose.LogRequest{Project: project, Service: service}, func(l compose.LogLine) {
			s.emit(logEvent, LogEvent{ID: id, Type: LogLineEvent, Line: l})
		})
		if ctx.Err() == nil { // not stopped: tell the frontend why it ended
			e := LogEvent{ID: id, Type: LogEndedEvent}
			if err != nil {
				e.Error = err.Error()
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
