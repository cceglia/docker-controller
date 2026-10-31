package main

import (
	"context"
	"sync"

	"dockercontroller/internal/compose"
)

const (
	// usageEvent carries a compose.UsageSample to the frontend.
	usageEvent = "usage"
	// usageStoppedEvent tells the frontend the usage stream stopped (window
	// closed), so it drops the figures instead of showing stale ones.
	usageStoppedEvent = "usage-stopped"
)

// usageStreamer is the thin Wails adapter that runs compose.Manager.Usage
// while the window is visible. Start and Stop are idempotent and never block
// the caller: a stream that is stopping finishes in the background, a new one
// waits for it, and the frontend is told about the stop only once no sample can
// follow.
type usageStreamer struct {
	manager *compose.Manager
	ctx     context.Context // the application: ends the stream on quit
	emit    func(name string, data any)
	onError func(error)

	mu     sync.Mutex
	cancel context.CancelFunc // of the stream wanted now; nil when stopped
	last   chan struct{}      // closed when the latest stream is over
}

// Start begins the stream unless it is already wanted.
func (u *usageStreamer) Start() {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(u.ctx)
	prev, done := u.last, make(chan struct{})
	u.cancel, u.last = cancel, done
	go func() {
		defer close(done)
		if prev != nil {
			<-prev // never two docker stats at once
		}
		if ctx.Err() != nil {
			return
		}
		if err := u.manager.Usage(ctx, func(s compose.UsageSample) { u.emit(usageEvent, s) }); err != nil && u.onError != nil {
			u.onError(err)
		}
	}()
}

// Stop ends the stream, which terminates the `docker stats` process, and then
// tells the frontend. It does nothing when the stream is not running.
func (u *usageStreamer) Stop() {
	u.mu.Lock()
	cancel, done := u.cancel, u.last
	u.cancel = nil
	u.mu.Unlock()
	if cancel == nil {
		return
	}
	cancel()
	go func() {
		<-done
		u.mu.Lock()
		restarted := u.cancel != nil
		u.mu.Unlock()
		if !restarted {
			u.emit(usageStoppedEvent, nil)
		}
	}()
}
