package main

import (
	"context"
	"errors"
	"testing"
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
