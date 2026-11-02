package gui

import (
	"context"
	"sync"

	"dockercontroller/internal/compose"
)

// usageFeed runs the stream of the Utilizzo risorse. The figures are the
// machine's, not a Project's: the stream starts while the window is on screen
// and Docker answers, and no choice of Project starts or ends it. Every sample
// reaches OnSample on the thread that draw posts to.
//
// Every sample carries the generation of the stream that delivered it. A sample
// that was queued before a stop reaches OnSample only while its generation is
// still the current one, so a late figure never comes back on screen.
type usageFeed struct {
	// OnSample receives one sample of the stream on screen, or nil when the
	// stream stops and the figures must leave the screen. It runs on the thread
	// draw posts to, never on the goroutine of the stream.
	OnSample func(*compose.UsageSample)

	// source is the stream of the Manager. A nil source is a Manager that is
	// not wired yet: the feed never starts `docker stats`.
	source UsageSource
	// stream is the single stream of the feed: it never runs two at once.
	stream *UsageStreamer
	// draw moves a function onto the thread that touches the widgets.
	draw func(func())

	mu sync.Mutex
	// reachable is what the last snapshot said of Docker: it decides the stream
	// together with hidden.
	reachable bool
	// hidden is true while the window is off screen: `docker stats` runs only
	// while the figures can be seen.
	hidden bool
	// gen bumps at every change of stream: it is what tells a sample of a stream
	// that ended from the sample of the one on screen.
	gen uint64
	// on says whether a stream is wanted now.
	on bool
}

// newUsageFeed builds a feed that starts no stream. ctx is the application and
// ends every stream. draw posts to the thread that draws; a nil draw runs the
// sample where it arrives, which is what a test wants and a window never does.
func newUsageFeed(ctx context.Context, source UsageSource, draw func(func())) *usageFeed {
	if draw == nil {
		draw = func(run func()) { run() }
	}
	f := &usageFeed{source: source, draw: draw}
	f.stream = NewUsageStreamer(ctx, func(ctx context.Context, _ func(compose.UsageSample)) error {
		if f.source == nil {
			return nil
		}
		// The generation is read when the stream starts, not when a sample
		// arrives: a figure of the stream that ends must stay out of the window
		// even when it reaches the feed after the stop.
		gen := f.generation()
		return f.source(ctx, func(sample compose.UsageSample) {
			f.deliver(gen, &sample)
		})
	})
	return f
}

// Follow keeps the state of Docker the snapshot reports and starts or stops the
// stream. The stream runs while the window is on screen and Docker answers; no
// Project takes part, because the figures of the machine do not belong to one.
func (f *usageFeed) Follow(dockerReachable bool) {
	f.mu.Lock()
	f.reachable = dockerReachable
	f.mu.Unlock()
	f.refresh()
}

// SetHidden stops the stream when the window leaves the screen and starts it
// again when the window comes back. The Docker state stays, so the return reads
// no snapshot to know what to measure.
func (f *usageFeed) SetHidden(hidden bool) {
	f.mu.Lock()
	changed := f.hidden != hidden
	f.hidden = hidden
	f.mu.Unlock()
	if changed {
		f.refresh()
	}
}

// refresh starts or stops the stream the current state asks for. Only a change
// of the answer touches the stream, so a new snapshot that changes nothing keeps
// `docker stats` running.
func (f *usageFeed) refresh() {
	f.mu.Lock()
	want := !f.hidden && f.reachable && f.source != nil
	was := f.on
	start := want && !was
	stop := !want && was
	f.on = want
	if !start && !stop {
		f.mu.Unlock()
		return
	}
	f.gen++
	gen := f.gen
	f.mu.Unlock()

	if stop {
		f.stream.Stop()
		// The figures of the stream that ended leave the screen at once: the
		// next stream may take a while to deliver.
		f.deliver(gen, nil)
		return
	}
	f.stream.Start()
}

// generation returns the generation of the stream wanted now.
func (f *usageFeed) generation() uint64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.gen
}

// current reports whether gen is still the generation of the stream wanted now.
func (f *usageFeed) current(gen uint64) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.gen == gen
}

// deliver hands one sample of the stream gen to the thread that draws, unless
// the feed moved on in the meantime: the generation is what keeps a figure of a
// stopped stream off the screen. A nil sample takes the figures away.
func (f *usageFeed) deliver(gen uint64, sample *compose.UsageSample) {
	f.draw(func() {
		if !f.current(gen) {
			return
		}
		if f.OnSample != nil {
			f.OnSample(sample)
		}
	})
}
