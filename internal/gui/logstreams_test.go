package gui

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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
		l.Stop(strconv.Itoa(i))
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

func TestStreamLogs_EndCarriesTheStderrDetailOfCompose(t *testing.T) {
	source := func(_ context.Context, _ compose.LogRequest, onLine func(compose.LogLine)) error {
		onLine(compose.LogLine{Service: "web", Text: "hello"})
		onLine(compose.LogLine{Text: "no configuration file provided: not found", Stderr: true})
		return errors.New("exit status 1")
	}
	var lines []compose.LogLine
	onLine := func(l compose.LogLine) { lines = append(lines, l) }

	end := StreamLogs(context.Background(), compose.LogRequest{Project: "p"}, source, onLine)

	if end.Stopped {
		t.Error("the stream ended on its own, not on a stop")
	}
	if !strings.Contains(end.Error, "no configuration file provided") {
		t.Errorf("end = %+v, want the stderr detail", end)
	}
	if len(lines) != 2 {
		t.Fatalf("%d lines, want 2: %v", len(lines), lines)
	}
	if !lines[1].Stderr {
		t.Error("the second line lost its stderr mark")
	}
}

func TestStreamLogs_CleanEndHasNoError(t *testing.T) {
	source := func(_ context.Context, _ compose.LogRequest, onLine func(compose.LogLine)) error {
		onLine(compose.LogLine{Service: "web", Text: "bye"})
		return nil
	}
	var lines int
	end := StreamLogs(context.Background(), compose.LogRequest{}, source, func(compose.LogLine) { lines++ })

	if end.Stopped {
		t.Error("a stream that compose closed is not stopped by the view")
	}
	if end.Error != "" {
		t.Errorf("end error = %q, want none", end.Error)
	}
	if lines != 1 {
		t.Errorf("%d lines, want 1", lines)
	}
}

func TestStreamLogs_StoppedStreamReportsNoEnd(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	source := func(context.Context, compose.LogRequest, func(compose.LogLine)) error {
		cancel()
		return nil
	}
	n := 0
	end := StreamLogs(ctx, compose.LogRequest{}, source, func(compose.LogLine) { n++ })
	if n != 0 {
		t.Errorf("%d lines after a stop, want none", n)
	}
	if !end.Stopped || end.Error != "" {
		t.Errorf("end = %+v, want a stop without an error", end)
	}
}
