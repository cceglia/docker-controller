package gui

import (
	"context"
	"os"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/design"
	"dockercontroller/internal/gui/logs"
	"dockercontroller/internal/settings"
)

// Run opens the application and blocks until it quits. The snapshot of the
// Manager is read through Watch, so a Project added or removed appears without
// a restart. activities is the log of the session and store holds the settings:
// the list, the bottom panel and the settings window read them in later
// tickets. notifier names a Project as the list does and shows the desktop
// notifications the Manager's events deserve.
func Run(ctx context.Context, manager *compose.Manager, activities *ActivityLog, store *settings.Store, notifier *Notifier) error {
	a := &application{
		ctx:        ctx,
		manager:    manager,
		activities: activities,
		store:      store,
		notifier:   notifier,
		// Until the first snapshot arrives Docker has not been asked yet, so
		// the list must not call it empty.
		snapshot: compose.Snapshot{Loading: true},
	}

	// The settings come before the first window and before the Qt thread: a
	// close reads them at once, and a database that cannot be read applies the
	// defaults, so the user is never locked in the window.
	a.prefs = settings.Defaults()
	if store != nil {
		if saved, err := store.Get(ctx); err == nil {
			a.prefs = saved
		}
	}

	// QApplication calls runtime.LockOSThread, so the Qt main thread stays the
	// thread of Run for the life of the process.
	a.qt = qt6.NewQApplication(os.Args)

	// The application name is the id of the desktop entry: Qt derives the class
	// of the window from it, and the entry declares the same value in
	// StartupWMClass. Without the match KDE shows the wrong icon and does not
	// group the window under the menu entry.
	qt6.QCoreApplication_SetApplicationName(appID)
	qt6.QGuiApplication_SetDesktopFileName(appID)
	// The display name is what the window list of the desktop reads.
	qt6.QGuiApplication_SetApplicationDisplayName(windowTitle)
	// Closing the window hides it, so the application does not end with its last
	// window: the close event decides, and the tray's "Esci" is what quits.
	qt6.QGuiApplication_SetQuitOnLastWindowClosed(false)

	// The window colour of the desktop palette picks the palette of the roles
	// the sheet carries: a dark window needs lighter colours to keep the
	// contrast of the labels of state.
	theme := themeForPalette(qt6.QGuiApplication_Palette())

	a.qt.SetStyleSheet(design.ApplicationStyleSheet(theme))

	a.shell = newShell(a, theme)
	// The click on a notification shows the window. The signal arrives on a
	// D-Bus goroutine, so the window is touched on the Qt thread.
	if notifier != nil {
		notifier.OnClick(func() { mainthread.Start(a.present) })
	}
	// The tray comes before the first snapshot and before the window shows, so
	// the icon is on the panel while the list is still loading. Without a tray
	// the field stays nil, the window keeps "Impostazioni" and closing it quits.
	a.tray = startTray(a.snapshot, a.present, a.quit, a.chooseTray)
	a.watch(ctx)
	a.shell.show()

	qt6.QApplication_Exec()
	return nil
}

// themeForPalette reads the window colour of the desktop palette. A missing
// palette or colour falls back to the light theme.
func themeForPalette(palette *qt6.QPalette) design.DesignTheme {
	if palette == nil {
		return design.LightTheme
	}
	window := palette.ColorWithCr(qt6.QPalette__Window)
	if window == nil {
		return design.LightTheme
	}
	return design.ThemeForLightness(window.Lightness())
}

// application owns what outlives the window: the Qt application, the Manager,
// the Activity log, the settings and the notifier. A goroutine that arrives
// with a snapshot moves it onto the Qt thread with mainthread.Start, so no
// widget is touched from there.
type application struct {
	// qt is the QApplication: it must exist before every widget.
	qt         *qt6.QApplication
	ctx        context.Context
	manager    *compose.Manager
	activities *ActivityLog
	store      *settings.Store
	notifier   *Notifier
	// prefs is the saved settings: what a close does and the terminal command.
	// It is read once at startup and after every save, so a close answers
	// without a query.
	prefs settings.Settings
	// shell is the window of the application.
	shell *shell
	// tray is the icon on the panel, nil when the desktop has no tray.
	tray *trayAdapter
	// question is the modal question of a destructive Action chosen in the tray
	// while it is open, so a second click cannot open a second one.
	question *confirmDialog
	// prefsDialog is the settings dialog on screen and closeQuestion the
	// "Chiedi ogni volta" question of a close: each one opens once.
	prefsDialog   *preferencesDialog
	closeQuestion *closeDialog
	// quitting is true from the first quit: the close event that follows starts
	// nothing.
	quitting bool
	// snapshot is the last one applied.
	snapshot compose.Snapshot
}

// logSource is the stream the Log tab opens. A Manager that is not wired yet
// leaves the page with no source, so the window still builds.
func (a *application) logSource() logs.LogSource {
	if a.manager == nil {
		return nil
	}
	return a.manager.Logs
}

// usageSource is the stream the detail opens for the Utilizzo risorse. A
// Manager that is not wired yet leaves it with no source, so no `docker stats`
// starts and the window still builds.
func (a *application) usageSource() UsageSource {
	if a.manager == nil {
		return nil
	}
	return a.manager.Usage
}

// watch follows the Manager and moves every snapshot onto the Qt thread.
func (a *application) watch(ctx context.Context) {
	go func() {
		for snapshot := range a.manager.Watch(ctx) {
			snapshot := snapshot
			mainthread.Start(func() { a.apply(snapshot) })
		}
	}()
}

// apply keeps the snapshot and shows it in the window. The notification keeps
// the Display names, so a notification names a Project as the list does.
func (a *application) apply(snapshot compose.Snapshot) {
	a.snapshot = snapshot
	if a.notifier != nil {
		a.notifier.SetProjects(snapshot.Projects)
	}
	a.shell.apply(snapshot)
	// The tray follows the same snapshot as the window: the icon, the tooltip
	// and, only when the tree changed, the menu.
	if a.tray != nil {
		a.tray.apply(snapshot)
	}
}
