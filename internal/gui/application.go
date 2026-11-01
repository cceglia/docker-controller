package gui

import (
	"context"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/settings"
	"dockercontroller/internal/tray"
	"dockercontroller/internal/trayadapter"
)

// Run opens the application and blocks until it quits. The snapshot of the
// Manager is read through Watch, so a Project added or removed appears without
// a restart. activities is the log of the session: the bottom panel reads it
// and the Manager pushes its events into it. store holds the settings: the
// terminal command of "Apri terminale" and the close behavior come from it.
// notifier shows the desktop notifications the Manager's events and the tray
// Actions deserve.
func Run(ctx context.Context, manager *compose.Manager, activities *ActivityLog, store *settings.Store, notifier *Notifier) error {
	a := &application{
		ctx:        ctx,
		manager:    manager,
		activities: activities,
		store:      store,
		notifier:   notifier,
		// Until the first snapshot arrives Docker has not been asked yet, so
		// the first window must not call it unreachable.
		snapshot: compose.Snapshot{Loading: true},
	}

	a.gtk = adw.NewApplication(appID, gio.ApplicationDefaultFlags)
	// The window is destroyed when the user hides it in the tray. Without this
	// hold the application would quit with its last window, so the hold is
	// what keeps the process alive until "Esci".
	a.gtk.Hold()
	a.gtk.ConnectActivate(func() { a.present() })

	a.watch(ctx)
	a.gtk.Run(os.Args)
	return nil
}

// application owns what outlives the window: the GTK application, the settings,
// the Activity log, the notifier and the tray. The window is built on demand and
// destroyed when the user hides it, so the application, and not the window, is
// the unit that lives as long as the process.
type application struct {
	// gtk is the AdwApplication: only "Esci" quits it.
	gtk *adw.Application
	// ctx stops the Activities, the log streams and `docker stats` on quit.
	ctx        context.Context
	manager    *compose.Manager
	activities *ActivityLog
	store      *settings.Store
	notifier   *Notifier
	// tray is the icon in the system tray, nil until it is started and when it
	// could not start. It is created once, never with a window.
	tray *trayadapter.Tray
	// window is the window on screen, nil while the application sits in the
	// tray.
	window *shell
	// prefs is the settings window while it exists, nil otherwise.
	prefs *prefsWindow
	// snapshot is the last one applied: a new window starts from it.
	snapshot compose.Snapshot
	// quitting is true once "Esci" ran, so no window is created again.
	quitting bool
}

// watch follows the Manager and moves every snapshot onto the GTK loop. The
// last one is kept, so a window created later starts from it.
func (a *application) watch(ctx context.Context) {
	go func() {
		for snapshot := range a.manager.Watch(ctx) {
			glib.IdleAdd(func() { a.apply(snapshot) })
		}
	}()
}

// apply keeps the snapshot and shows it everywhere that reads it: the tray on
// any state, the notifications, and the window when there is one.
func (a *application) apply(snapshot compose.Snapshot) {
	a.snapshot = snapshot
	// A notification names a Project as the list does, and the snapshot is the
	// last place that tells the Display name.
	a.notifier.SetProjects(snapshot.Projects)
	if a.tray != nil {
		a.tray.Show(tray.BuildMenu(snapshot), tray.IconFor(snapshot))
	}
	if a.window != nil {
		a.window.apply(snapshot)
	}
}

// present brings the window to the front and creates it when it is gone. A tray
// click and an activation both land here; it returns nil once the application
// is quitting.
func (a *application) present() *shell {
	if a.quitting {
		return nil
	}
	if a.window != nil {
		a.window.show()
		return a.window
	}
	return a.newWindow()
}

// newWindow builds a window from the latest snapshot and shows it. The
// selection starts empty, as it does on the first window. The history of the
// Activities starts empty too: the close dialog says it is lost, so an Activity
// of the hidden window, and one that ran with no window, stay out.
func (a *application) newWindow() *shell {
	a.startTray()
	a.activities.Clear()
	s := newShell(a)
	a.window = s
	s.apply(a.snapshot)
	s.window.Present()
	return s
}

// dropWindow forgets the window that is going away, so the next click on the
// tray builds a new one. The settings window goes with it: it belongs to the
// window that opened it. The window is destroyed after this call, so every
// state that points at it is cleared first.
func (a *application) dropWindow(s *shell) {
	if a.window != s {
		return
	}
	a.window = nil
	a.closePreferences()
}

// closePreferences destroys the settings window when it is on screen. It is
// called before the window that owns it goes away, so the application never
// holds a window that is already destroyed.
func (a *application) closePreferences() {
	if a.prefs == nil {
		return
	}
	p := a.prefs
	a.prefs = nil
	p.window.Destroy()
}

// quit stops the streams of the window on screen and quits the application. The
// quit runs when the loop is idle, so it never destroys a window on the stack of
// its own handler.
func (a *application) quit() {
	if a.quitting {
		return
	}
	a.quitting = true
	if a.window != nil {
		a.window.stopStreams()
	}
	glib.IdleAdd(func() {
		// Give back the hold of Run; Quit ends the loop in any case.
		a.gtk.Release()
		a.gtk.Quit()
	})
}

// startTray puts the icon in the system tray, once. A tray that cannot start
// stays silent: systray logs the failure on its own goroutine and the window
// goes on without it.
func (a *application) startTray() {
	if a.tray != nil {
		return
	}
	a.tray = trayadapter.New(
		// A click arrives on a D-Bus goroutine: the window moves only on the
		// loop, and a missing window is built there.
		func() { glib.IdleAdd(func() { a.present() }) },
		func(node trayadapter.Node) { glib.IdleAdd(func() { a.trayChoose(node) }) },
	)
	a.tray.Start()
	a.tray.Show(tray.BuildMenu(a.snapshot), tray.IconFor(a.snapshot))
}
