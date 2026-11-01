package gui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"dockercontroller/internal/compose"
)

func TestLogStreams_StopCancelsARegisteredStream(t *testing.T) {
	var l LogStreams
	ctx, cancel := context.WithCancel(context.Background())
	if err := l.Add("a", cancel); err != nil {
		t.Fatal(err)
	}
	l.Stop("a")
	if ctx.Err() == nil {
		t.Error("the stream was not cancelled")
	}
}

func TestLogStreams_StopBeforeAddCancelsTheStreamOnceItRegisters(t *testing.T) {
	var l LogStreams
	l.Stop("a") // StopLogs overtook StartLogs
	ctx, cancel := context.WithCancel(context.Background())
	err := l.Add("a", cancel)
	if !errors.Is(err, ErrLogStopped) {
		t.Fatalf("err = %v, want ErrLogStopped", err)
	}
	if ctx.Err() == nil {
		t.Error("a stream stopped before registering must be cancelled")
	}
	// The id is free again afterwards.
	if err := l.Add("a", func() {}); err != nil {
		t.Errorf("second add: %v", err)
	}
}

func TestLogStreams_StopAllCancelsEveryOpenStream(t *testing.T) {
	var l LogStreams
	ctxA, cancelA := context.WithCancel(context.Background())
	ctxB, cancelB := context.WithCancel(context.Background())
	if err := l.Add("a", cancelA); err != nil {
		t.Fatal(err)
	}
	if err := l.Add("b", cancelB); err != nil {
		t.Fatal(err)
	}
	l.StopAll()
	if ctxA.Err() == nil || ctxB.Err() == nil {
		t.Error("StopAll left a stream running")
	}
	// A StartLogs still in flight after the window is destroyed must not start.
	ctxC, cancelC := context.WithCancel(context.Background())
	if err := l.Add("c", cancelC); !errors.Is(err, ErrLogsClosed) {
		t.Errorf("err = %v, want ErrLogsClosed", err)
	}
	if ctxC.Err() == nil {
		t.Error("a stream added after StopAll must be cancelled")
	}
	// The window is created again: streams are accepted.
	l.Resume()
	if err := l.Add("c", func() {}); err != nil {
		t.Errorf("add after resume: %v", err)
	}
}

func TestLogStreams_AddRefusesADuplicateID(t *testing.T) {
	var l LogStreams
	if err := l.Add("a", func() {}); err != nil {
		t.Fatal(err)
	}
	if err := l.Add("a", func() {}); err == nil {
		t.Error("duplicate id accepted")
	}
}

func TestLogStreams_StoppedIDsAreBounded(t *testing.T) {
	var l LogStreams
	for i := 0; i < 10*maxStoppedLogIDs; i++ {
		l.Stop(string(rune('a'+i%26)) + string(rune(i)))
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
	emit := func(e LogEvent) { events = append(events, e) }

	StreamLogs(context.Background(), "s1", compose.LogRequest{Project: "p"}, source, emit)

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
	StreamLogs(ctx, "s1", compose.LogRequest{}, source, func(LogEvent) { n++ })
	if n != 0 {
		t.Errorf("%d events after a stop, want none", n)
	}
}
