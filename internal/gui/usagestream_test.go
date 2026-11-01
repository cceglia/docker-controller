package gui

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"dockercontroller/internal/compose"
)

// waitSignal fails the test when signal does not arrive in a short time, so a
// deadlock of the streamer reads as a failure and not as a hang.
func waitSignal(t *testing.T, signal <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}

// testStreamer builds a streamer that ends with the test.
func testStreamer(t *testing.T, source UsageSource) *UsageStreamer {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return NewUsageStreamer(ctx, source)
}

func TestUsageStreamer_StartIsIdempotent(t *testing.T) {
	var runs atomic.Int32
	started := make(chan struct{}, 2)
	source := func(ctx context.Context, _ func(compose.UsageSample)) error {
		runs.Add(1)
		started <- struct{}{}
		<-ctx.Done()
		return nil
	}
	u := testStreamer(t, source)
	stopped := make(chan struct{})
	u.OnStop = func() { close(stopped) }

	u.Start()
	u.Start()
	waitSignal(t, started, "the stream")
	u.Stop()
	waitSignal(t, stopped, "the stopped notice")

	if got := runs.Load(); got != 1 {
		t.Errorf("%d streams, want 1", got)
	}
}

func TestUsageStreamer_StopDoesNotBlockTheCaller(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{})
	source := func(ctx context.Context, _ func(compose.UsageSample)) error {
		close(started)
		<-ctx.Done()
		<-release // a docker that is slow to go away
		return nil
	}
	u := testStreamer(t, source)
	stopped := make(chan struct{})
	u.OnStop = func() { close(stopped) }

	u.Start()
	waitSignal(t, started, "the stream")
	returned := make(chan struct{})
	go func() {
		u.Stop()
		u.Stop() // the second stop does nothing and must not panic
		close(returned)
	}()
	waitSignal(t, returned, "Stop to return")

	select {
	case <-stopped:
		t.Fatal("the stopped notice arrived before the stream ended")
	default:
	}

	close(release)
	waitSignal(t, stopped, "the stopped notice")
}

func TestUsageStreamer_ANewStreamWaitsForThePreviousOne(t *testing.T) {
	var active, most atomic.Int32
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	source := func(ctx context.Context, _ func(compose.UsageSample)) error {
		n := active.Add(1)
		for {
			old := most.Load()
			if n <= old || most.CompareAndSwap(old, n) {
				break
			}
		}
		started <- struct{}{}
		<-ctx.Done()
		<-release
		active.Add(-1)
		return nil
	}
	u := testStreamer(t, source)

	u.Start()
	waitSignal(t, started, "the first stream")
	u.Stop()
	u.Start()

	select {
	case <-started:
		t.Fatal("the second stream started while the first one was still running")
	case <-time.After(50 * time.Millisecond):
	}

	close(release) // the first stream ends and the second one starts
	waitSignal(t, started, "the second stream")
	u.Stop()

	if got := most.Load(); got != 1 {
		t.Errorf("%d streams at once, want 1", got)
	}
}

func TestUsageStreamer_DeliversTheSamples(t *testing.T) {
	want := compose.UsageSample{DockerReachable: true, Containers: 2}
	source := func(ctx context.Context, onSample func(compose.UsageSample)) error {
		onSample(want)
		<-ctx.Done()
		return nil
	}
	u := testStreamer(t, source)
	got := make(chan compose.UsageSample, 1)
	u.OnSample = func(s compose.UsageSample) { got <- s }

	u.Start()
	select {
	case sample := <-got:
		if sample.Containers != want.Containers || !sample.DockerReachable {
			t.Errorf("sample = %+v, want %+v", sample, want)
		}
	case <-time.After(time.Second):
		t.Fatal("no sample arrived")
	}
	u.Stop()
}

func TestUsageStreamer_ASampleAfterTheStopIsDropped(t *testing.T) {
	var mu sync.Mutex
	var got []int
	first := make(chan struct{})
	source := func(ctx context.Context, onSample func(compose.UsageSample)) error {
		onSample(compose.UsageSample{Containers: 1})
		close(first)
		<-ctx.Done()
		onSample(compose.UsageSample{Containers: 2}) // a sample of a stream already stopped
		return nil
	}
	u := testStreamer(t, source)
	stopped := make(chan struct{})
	u.OnSample = func(s compose.UsageSample) {
		mu.Lock()
		got = append(got, s.Containers)
		mu.Unlock()
	}
	u.OnStop = func() { close(stopped) }

	u.Start()
	waitSignal(t, first, "the first sample")
	u.Stop()
	waitSignal(t, stopped, "the stopped notice")

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 || got[0] != 1 {
		t.Errorf("samples = %v, want the first one only", got)
	}
}

func TestUsageStreamer_TheStoppedNoticeComesAfterTheLastSample(t *testing.T) {
	source := func(ctx context.Context, onSample func(compose.UsageSample)) error {
		for ctx.Err() == nil {
			onSample(compose.UsageSample{DockerReachable: true})
			time.Sleep(time.Millisecond)
		}
		return nil
	}
	u := testStreamer(t, source)
	var mu sync.Mutex
	var events []string
	u.OnSample = func(compose.UsageSample) {
		mu.Lock()
		events = append(events, "sample")
		mu.Unlock()
	}
	stopped := make(chan struct{})
	u.OnStop = func() {
		mu.Lock()
		events = append(events, "stop")
		mu.Unlock()
		close(stopped)
	}

	u.Start()
	u.Stop()
	waitSignal(t, stopped, "the stopped notice")

	mu.Lock()
	defer mu.Unlock()
	if len(events) == 0 || events[len(events)-1] != "stop" {
		t.Errorf("events = %v, want the stop last", events)
	}
	var stops int
	for _, e := range events {
		if e == "stop" {
			stops++
		}
	}
	if stops != 1 {
		t.Errorf("%d stopped notices, want 1", stops)
	}
}

func TestUsageStreamer_AStreamThatFailsReportsTheError(t *testing.T) {
	failure := errors.New("docker not reachable")
	source := func(context.Context, func(compose.UsageSample)) error { return failure }
	u := testStreamer(t, source)
	got := make(chan error, 1)
	u.OnError = func(err error) { got <- err }

	u.Start()
	select {
	case err := <-got:
		if !errors.Is(err, failure) {
			t.Errorf("err = %v, want %v", err, failure)
		}
	case <-time.After(time.Second):
		t.Fatal("no error arrived")
	}
}
