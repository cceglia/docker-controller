package usage

import (
	"context"
	"sync"

	"dockercontroller/internal/compose"
)

// UsageSource streams the Utilizzo risorse: one sample per frame, delivered to
// onSample, which is never called concurrently. It blocks until ctx is
// cancelled. compose.Manager.Usage satisfies it.
type UsageSource func(ctx context.Context, onSample func(compose.UsageSample)) error

var _ UsageSource = (*compose.Manager)(nil).Usage

// UsageStreamer runs a UsageSource while the window is visible. Start and Stop
// are idempotent and never block the caller: a stream that is stopping finishes
// in the background, a new one waits for it, and OnStop is called only once no
// sample can follow and no new stream has started.
type UsageStreamer struct {
	ctx    context.Context // the application: ends the stream on quit
	source UsageSource

	// OnSample receives every sample of the stream, on the goroutine of the
	// source. OnStop says the stream ended because the view stopped it; OnError
	// says it ended by itself and why. Both run once, after the last sample.
	OnSample func(compose.UsageSample)
	OnStop   func()
	OnError  func(error)

	mu     sync.Mutex
	gen    uint64             // bumps with every Start and every Stop
	cancel context.CancelFunc // of the stream wanted now; nil when stopped
	last   chan struct{}      // closed when the latest stream is over
}

// NewUsageStreamer builds a stopped streamer. ctx is the application: without
// it the stream never starts.
func NewUsageStreamer(ctx context.Context, source UsageSource) *UsageStreamer {
	if ctx == nil {
		ctx = context.Background()
	}
	return &UsageStreamer{ctx: ctx, source: source}
}

// Start begins the stream unless it is already wanted.
func (u *UsageStreamer) Start() {
	u.mu.Lock()
	if u.cancel != nil {
		u.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(u.ctx)
	prev, done := u.last, make(chan struct{})
	u.cancel, u.last = cancel, done
	u.gen++
	gen := u.gen
	u.mu.Unlock()

	go func() {
		defer close(done)
		if prev != nil {
			<-prev // never two `docker stats` at once
		}
		if ctx.Err() != nil {
			return
		}
		err := u.source(ctx, func(sample compose.UsageSample) {
			if !u.current(gen) {
				return // a sample of a stream the view stopped
			}
			if u.OnSample != nil {
				u.OnSample(sample)
			}
		})
		if ctx.Err() != nil {
			return
		}
		if err != nil && u.OnError != nil {
			u.OnError(err)
		}
	}()
}

// Stop ends the stream, which terminates the `docker stats` process, and then
// tells the view, unless a new stream has started in the meantime. It does
// nothing when the stream is not running.
func (u *UsageStreamer) Stop() {
	u.mu.Lock()
	cancel, done := u.cancel, u.last
	if cancel == nil {
		u.mu.Unlock()
		return
	}
	u.cancel = nil
	u.gen++
	gen := u.gen
	u.mu.Unlock()

	cancel()
	go func() {
		<-done
		if u.current(gen) && u.OnStop != nil {
			u.OnStop()
		}
	}()
}

// current reports whether gen is still the generation of the stream wanted now.
func (u *UsageStreamer) current(gen uint64) bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.gen == gen
}
