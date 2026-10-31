package main

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"dockercontroller/internal/settings"
)

const (
	windowTitle  = "Docker Controller"
	windowWidth  = 1120
	windowHeight = 760
	windowURL    = "/"
)

// mainWindow owns the application window. It creates the window with the same
// options every time, installs the closing hook, and destroys the window when
// the user closes it. It is the only place that knows the window options and
// its life cycle.
type mainWindow struct {
	app      *application.App
	settings *settings.Store
	usage    *usageStreamer
	logs     *ProjectsService

	// ops serializes creation, showing and destruction. The tray, the closing
	// hook and the shutdown can run at the same time.
	ops    sync.Mutex
	window *application.WebviewWindow
	// current is the id of the window, or 0 when the application is in the
	// tray. The closing hook reads it without ops, so it never blocks the
	// main thread.
	current atomic.Uint64
	// destroying is the id of the window that closeToTray is destroying: its
	// closing hook must let the event through.
	destroying atomic.Uint64
	// quitting is set once the application is shutting down, so no window is
	// created again and the closing of the windows done by Wails is let
	// through.
	quitting atomic.Bool
}

// newMainWindow returns the owner of the application window.
func newMainWindow(app *application.App, store *settings.Store, usage *usageStreamer, logs *ProjectsService) *mainWindow {
	return &mainWindow{app: app, settings: store, usage: usage, logs: logs}
}

// install creates the first window and hooks the application shutdown.
func (w *mainWindow) install() {
	w.app.OnShutdown(func() { w.quitting.Store(true) })
	w.ops.Lock()
	defer w.ops.Unlock()
	w.window = w.newWindow()
	w.current.Store(uint64(w.window.ID()))
}

// newWindow creates a window with the application options.
func (w *mainWindow) newWindow() *application.WebviewWindow {
	win := w.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  windowTitle,
		Width:  windowWidth,
		Height: windowHeight,
		URL:    windowURL,
	})
	// A hook (not a listener) is used so the default close can be cancelled.
	win.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		w.onClosing(win, e)
	})
	return win
}

// show brings the window to the front. It creates the window again when it was
// destroyed, then starts the usage stream.
func (w *mainWindow) show() {
	w.ops.Lock()
	defer w.ops.Unlock()
	if w.quitting.Load() {
		return
	}
	win := w.window
	if win == nil {
		win = w.newWindow()
		w.window = win
		w.current.Store(uint64(win.ID()))
	}
	w.logs.resumeLogs()

	win.Show()
	win.UnMinimise()
	win.Focus()
	w.usage.Start()
}

// closeToTray destroys the window and leaves the application in the tray. It
// stops the usage stream and every open log stream first.
//
// Destroying the window frees the memory of the webview. The cost is that the
// next click on the tray creates a new webview and loses the session state,
// such as the history of the Activities and the open panels. That trade is
// deliberate: a window in the tray must not keep the webview alive.
func (w *mainWindow) closeToTray() {
	w.ops.Lock()
	if w.quitting.Load() {
		w.ops.Unlock()
		return
	}
	win := w.window
	if win == nil {
		w.ops.Unlock()
		return
	}
	// Stop the streams before the window goes away. The lock keeps a tray
	// click from creating a new window, and starting its stream, in between.
	w.usage.Stop()
	w.logs.StopAllLogs()
	w.window = nil
	w.current.Store(0)
	w.destroying.Store(uint64(win.ID()))
	w.ops.Unlock()

	win.Close()

	w.destroying.CompareAndSwap(uint64(win.ID()), 0)
}

// onClosing decides what closing win does. A window that closeToTray is
// destroying, or one that is no longer current, is let through. The choice
// runs in its own goroutine: the hook runs on the main thread, and the choice
// can wait for it.
func (w *mainWindow) onClosing(win *application.WebviewWindow, e *application.WindowEvent) {
	id := uint64(win.ID())
	if w.quitting.Load() || w.destroying.Load() == id || w.current.Load() != id {
		return
	}

	e.Cancel()
	go func() {
		cfg, err := w.settings.Get(context.Background())
		if err != nil {
			// Never trap the user: fall back to the default behaviour.
			w.app.Logger.Error("close behavior", "error", err)
			cfg = settings.Defaults()
		}
		w.perform(cfg.CloseBehavior)
	}()
}

// perform carries out a choice made by the user or read from the settings.
func (w *mainWindow) perform(b settings.CloseBehavior) {
	switch b {
	case settings.CloseQuit:
		go w.app.Quit() // never from inside the closing hook
	case settings.CloseAsk:
		w.app.Event.Emit(closeRequestedEvent)
	default:
		w.closeToTray()
	}
}
