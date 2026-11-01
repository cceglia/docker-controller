package gui

import (
	"context"
	"sync"
	"testing"
	"time"

	"dockercontroller/internal/compose"
)

// usageQueue is the dispatch of a feed in a test: it holds the work a sample
// queues, so the test decides when the sample arrives and can stop the stream
// in between.
type usageQueue struct {
	mu     sync.Mutex
	queued []func()
}

// post queues one piece of work for the thread that draws.
func (q *usageQueue) post(run func()) {
	q.mu.Lock()
	q.queued = append(q.queued, run)
	q.mu.Unlock()
}

// drain runs the work queued so far, in the order it arrived.
func (q *usageQueue) drain() {
	q.mu.Lock()
	queued := q.queued
	q.queued = nil
	q.mu.Unlock()
	for _, run := range queued {
		run()
	}
}

// idleSource is the source of a feed in a test: it reports every start on
// starts and ends with the context.
func idleSource(starts chan<- struct{}) UsageSource {
	return func(ctx context.Context, _ func(compose.UsageSample)) error {
		if starts != nil {
			starts <- struct{}{}
		}
		<-ctx.Done()
		return nil
	}
}

// testFeed builds a feed that ends with the test.
func testFeed(t *testing.T, source UsageSource, dispatch func(func())) *usageFeed {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return newUsageFeed(ctx, source, dispatch)
}

func TestUsageFeed_WithoutDockerOrARunningContainerThereIsNoStream(t *testing.T) {
	for _, tc := range []struct {
		name      string
		running   int
		reachable bool
		want      bool
	}{
		{"Docker answers and a Container runs", 2, true, true},
		{"Docker does not answer", 2, false, false},
		{"no Container runs", 0, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			starts := make(chan struct{}, 1)
			f := testFeed(t, idleSource(starts), nil)
			f.Follow(compose.Project{Name: "shop", Running: tc.running}, tc.reachable)
			select {
			case <-starts:
				if !tc.want {
					t.Fatal("the stream started without a running Container or without Docker")
				}
			case <-time.After(50 * time.Millisecond):
				if tc.want {
					t.Fatal("the stream did not start with Docker and a running Container")
				}
			}
		})
	}
}

func TestUsageFeed_DeliversTheSamples(t *testing.T) {
	want := compose.UsageSample{DockerReachable: true, Containers: 3}
	reported := make(chan struct{}, 1)
	source := func(ctx context.Context, onSample func(compose.UsageSample)) error {
		onSample(want)
		reported <- struct{}{}
		<-ctx.Done()
		return nil
	}
	f := testFeed(t, source, nil)
	var got []*compose.UsageSample
	f.OnSample = func(sample *compose.UsageSample) { got = append(got, sample) }

	f.Follow(compose.Project{Name: "shop", Running: 1}, true)
	waitSignal(t, reported, "the sample")

	if len(got) != 1 || got[0] == nil || got[0].Containers != want.Containers {
		t.Fatalf("samples = %v, want the one with %d Containers", got, want.Containers)
	}
}

func TestUsageFeed_TheSameProjectKeepsOneStream(t *testing.T) {
	starts := make(chan struct{}, 3)
	f := testFeed(t, idleSource(starts), nil)
	project := compose.Project{Name: "shop", Running: 1}

	f.Follow(project, true)
	waitSignal(t, starts, "the stream")

	// A new snapshot of the same Project draws the same figures: it must not
	// terminate and restart `docker stats`.
	f.Follow(project, true)
	f.Follow(project, true)

	select {
	case <-starts:
		t.Fatal("a new snapshot of the same Project started a second stream")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestUsageFeed_AChangeOfProjectStartsANewStream(t *testing.T) {
	starts := make(chan struct{}, 2)
	f := testFeed(t, idleSource(starts), nil)

	f.Follow(compose.Project{Name: "shop", Running: 1}, true)
	waitSignal(t, starts, "the first stream")
	f.Follow(compose.Project{Name: "blog", Running: 1}, true)
	waitSignal(t, starts, "the second stream")
}

func TestUsageFeed_AStopTakesTheFiguresOffTheScreen(t *testing.T) {
	f := testFeed(t, idleSource(nil), nil)
	var got []*compose.UsageSample
	f.OnSample = func(sample *compose.UsageSample) { got = append(got, sample) }

	f.Follow(compose.Project{Name: "shop", Running: 1}, true)
	f.Follow(compose.Project{}, false)

	if len(got) != 1 || got[0] != nil {
		t.Fatalf("samples = %v, want the nil sample of the stop", got)
	}
}

func TestUsageFeed_AHiddenWindowStopsTheStreamAndTheReturnStartsIt(t *testing.T) {
	starts := make(chan struct{}, 2)
	f := testFeed(t, idleSource(starts), nil)
	var got []*compose.UsageSample
	f.OnSample = func(sample *compose.UsageSample) { got = append(got, sample) }

	project := compose.Project{Name: "shop", Running: 1}
	f.Follow(project, true)
	waitSignal(t, starts, "the stream")

	// The window leaves the screen: `docker stats` stops and the figures go, so
	// no number stays on screen as if it were current.
	f.SetHidden(true)
	if len(got) != 1 || got[0] != nil {
		t.Fatalf("samples = %v, want the nil sample of the stop", got)
	}

	// A snapshot of the hidden window starts nothing.
	f.Follow(project, true)
	select {
	case <-starts:
		t.Fatal("a hidden window started docker stats")
	case <-time.After(50 * time.Millisecond):
	}

	// The window comes back: the stream of the same Project starts again.
	f.SetHidden(false)
	waitSignal(t, starts, "the stream of the window back")
}

func TestUsageFeed_ASampleQueuedBeforeTheStopNeverArrives(t *testing.T) {
	queue := &usageQueue{}
	reported := make(chan struct{}, 1)
	source := func(ctx context.Context, onSample func(compose.UsageSample)) error {
		onSample(compose.UsageSample{DockerReachable: true, Containers: 1})
		reported <- struct{}{}
		<-ctx.Done()
		return nil
	}
	f := testFeed(t, source, queue.post)
	var got []*compose.UsageSample
	f.OnSample = func(sample *compose.UsageSample) { got = append(got, sample) }

	f.Follow(compose.Project{Name: "shop", Running: 1}, true)
	waitSignal(t, reported, "the sample")

	// The sample waits for the thread that draws; the stop arrives first.
	f.Follow(compose.Project{}, false)
	queue.drain()

	if len(got) != 1 || got[0] != nil {
		t.Fatalf("samples = %v, want the nil sample of the stop alone", got)
	}
}

func TestUsageFeed_WithoutASourceThereIsNothingToStream(t *testing.T) {
	f := testFeed(t, nil, nil)
	var got []*compose.UsageSample
	f.OnSample = func(sample *compose.UsageSample) { got = append(got, sample) }

	f.Follow(compose.Project{Name: "shop", Running: 1}, true)

	if len(got) != 0 {
		t.Fatalf("samples = %v, want none from a feed without a source", got)
	}
}
