package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"dockercontroller/internal/compose"
)

func TestLogStreams_StopCancelsARegisteredStream(t *testing.T) {
	var l logStreams
	ctx, cancel := context.WithCancel(context.Background())
	if err := l.add("a", cancel); err != nil {
		t.Fatal(err)
	}
	l.stop("a")
	if ctx.Err() == nil {
		t.Error("the stream was not cancelled")
	}
}

func TestLogStreams_StopBeforeAddCancelsTheStreamOnceItRegisters(t *testing.T) {
	var l logStreams
	l.stop("a") // StopLogs overtook StartLogs
	ctx, cancel := context.WithCancel(context.Background())
	err := l.add("a", cancel)
	if !errors.Is(err, errLogStopped) {
		t.Fatalf("err = %v, want errLogStopped", err)
	}
	if ctx.Err() == nil {
		t.Error("a stream stopped before registering must be cancelled")
	}
	// The id is free again afterwards.
	if err := l.add("a", func() {}); err != nil {
		t.Errorf("second add: %v", err)
	}
}

func TestLogStreams_AddRefusesADuplicateID(t *testing.T) {
	var l logStreams
	if err := l.add("a", func() {}); err != nil {
		t.Fatal(err)
	}
	if err := l.add("a", func() {}); err == nil {
		t.Error("duplicate id accepted")
	}
}

func TestLogStreams_StoppedIDsAreBounded(t *testing.T) {
	var l logStreams
	for i := 0; i < 10*maxStoppedLogIDs; i++ {
		l.stop(string(rune('a'+i%26)) + string(rune(i)))
	}
	if n := len(l.stopped); n > maxStoppedLogIDs {
		t.Errorf("%d remembered ids, want at most %d", n, maxStoppedLogIDs)
	}
}

func TestAppendLastLines_KeepsOnlyTheLastN(t *testing.T) {
	var lines []string
	for i := 0; i < 100; i++ {
		lines = appendLastLines(lines, fmt.Sprint("l", i), maxEndedStderrLines)
	}
	if len(lines) != maxEndedStderrLines {
		t.Fatalf("kept %d lines, want %d", len(lines), maxEndedStderrLines)
	}
	if lines[0] != "l80" || lines[len(lines)-1] != "l99" {
		t.Errorf("kept %v, want l80..l99", lines)
	}
}

func TestStreamLogs_EndedEventCarriesTheStderrDetailOfCompose(t *testing.T) {
	source := func(_ context.Context, _ compose.LogRequest, onLine func(compose.LogLine)) error {
		onLine(compose.LogLine{Text: "web  | hello"})
		onLine(compose.LogLine{Text: "no configuration file provided: not found", Stderr: true})
		return errors.New("exit status 1")
	}
	var events []LogEvent
	emit := func(_ string, data any) { events = append(events, data.(LogEvent)) }

	streamLogs(context.Background(), "s1", compose.LogRequest{Project: "p"}, source, emit)

	if len(events) != 3 {
		t.Fatalf("%d events, want 3: %v", len(events), events)
	}
	last := events[2]
	if last.Type != LogEndedEvent || !strings.Contains(last.Error, "no configuration file provided") {
		t.Errorf("ended event = %+v, want the stderr detail", last)
	}
}

func TestStreamLogs_StoppedStreamSendsNoEndedEvent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	source := func(context.Context, compose.LogRequest, func(compose.LogLine)) error {
		cancel()
		return nil
	}
	n := 0
	streamLogs(ctx, "s1", compose.LogRequest{}, source, func(string, any) { n++ })
	if n != 0 {
		t.Errorf("%d events after a stop, want none", n)
	}
}

func TestStartLogs_DuplicateIDReturnsAnItalianMessage(t *testing.T) {
	s := &ProjectsService{ctx: context.Background()}
	if err := s.logs.add("dup", func() {}); err != nil {
		t.Fatal(err)
	}
	err := s.StartLogs("dup", "p", "")
	if err == nil || err.Error() != "Flusso dei log già aperto." {
		t.Errorf("err = %v, want the Italian message", err)
	}
}
