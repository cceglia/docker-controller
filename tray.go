package main

import (
	"context"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/tray"
)

// trayAdapter is the thin Wails adapter over internal/tray: it rebuilds the
// system tray menu on every compose.Watch snapshot (no polling of its own) and
// shows the desktop notifications tray.Notifier asks for.
type trayAdapter struct {
	app           *application.App
	manager       *compose.Manager
	window        application.Window
	notifications *notifications.NotificationService
	ctx           context.Context // bounds the Activities and the Watch

	systemTray *application.SystemTray
	notifier   tray.Notifier

	mu       sync.Mutex
	snapshot compose.Snapshot // last one seen, for display names
	icon     tray.Icon
	menus    [2]*application.Menu // the menu on screen and the one before it
}

// newTrayAdapter creates the tray with an empty menu. The tray starts
// following the manager once the application is running.
func newTrayAdapter(ctx context.Context, app *application.App, manager *compose.Manager, window application.Window, notifs *notifications.NotificationService) *trayAdapter {
	t := &trayAdapter{app: app, manager: manager, window: window, notifications: notifs, ctx: ctx}
	t.systemTray = app.SystemTray.New()
	t.systemTray.SetTooltip("Docker Controller")
	t.systemTray.OnDoubleClick(t.showWindow)
	t.apply(manager.Snapshot())

	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		go func() {
			for s := range manager.Watch(ctx) {
				t.apply(s)
			}
		}()
	})
	return t
}

// apply rebuilds the icon and the menu from a snapshot.
func (t *trayAdapter) apply(s compose.Snapshot) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.snapshot = s

	if icon := tray.IconFor(s); icon != t.icon || t.menus[0] == nil {
		t.icon = icon
		t.systemTray.SetIcon(icon.PNG())
	}

	menu := application.NewMenu()
	t.fill(menu, tray.BuildMenu(s))
	t.systemTray.SetMenu(menu)
	// Keep the previous menu alive for one more generation, so a click on a row
	// of the menu that was open while the snapshot changed still resolves.
	if t.menus[1] != nil {
		t.menus[1].Destroy()
	}
	t.menus[1], t.menus[0] = t.menus[0], menu
}

func (t *trayAdapter) fill(menu *application.Menu, items []tray.Item) {
	for _, it := range items {
		switch {
		case it.Separator:
			menu.AddSeparator()
		case len(it.Children) > 0:
			sub := menu.AddSubmenu(it.Label)
			t.fill(sub, it.Children)
		default:
			mi := menu.Add(it.Label).SetEnabled(it.Enabled)
			if it.Enabled {
				mi.OnClick(func(*application.Context) { t.run(it) })
			}
		}
	}
}

// showWindow brings the application window to the front.
func (t *trayAdapter) showWindow() {
	t.window.Show()
	t.window.UnMinimise()
	t.window.Focus()
}

// run executes the click on an entry.
func (t *trayAdapter) run(it tray.Item) {
	switch it.Command {
	case tray.CommandShowWindow:
		t.showWindow()
		return
	case tray.CommandQuit:
		t.app.Quit()
		return
	}
	if it.Action == nil {
		return
	}
	if it.Confirm != "" {
		t.confirm(it)
		return
	}
	t.launch(*it.Action)
}

// confirm asks the question of a destructive entry and launches its Action
// only when the user accepts.
func (t *trayAdapter) confirm(it tray.Item) {
	d := t.app.Dialog.Question().SetTitle("Docker Controller").SetMessage(it.Confirm)
	ok := d.AddButton(it.Label)
	ok.OnClick(func() { t.launch(*it.Action) })
	cancel := d.AddButton("Annulla")
	d.SetDefaultButton(cancel).SetCancelButton(cancel)
	d.Show()
}

// launch starts an Action from the tray.
func (t *trayAdapter) launch(a compose.Action) {
	t.notifier.Launching(a)
	if _, err := t.manager.Do(t.ctx, a); err != nil {
		t.notifier.Aborted(a)
		t.app.Logger.Error("tray action", "project", a.Project, "kind", a.Kind, "error", err)
		t.notify(string(a.Kind)+a.Project, tray.Notification{Title: "Docker Controller", Body: userError(err).Error()})
	}
}

// onActivity receives every Activity event (see compose.Config.OnActivity).
func (t *trayAdapter) onActivity(e compose.ActivityEvent) {
	n, ok := t.notifier.Observe(e, t.label(e.Project))
	if ok {
		t.notify(string(e.ID), n)
	}
}

// label is the Display name of a Project, or its name.
func (t *trayAdapter) label(project string) string {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, p := range t.snapshot.Projects {
		if p.Name == project && p.DisplayName != "" {
			return p.DisplayName
		}
	}
	return project
}

func (t *trayAdapter) notify(id string, n tray.Notification) {
	err := t.notifications.SendNotification(notifications.NotificationOptions{ID: id, Title: n.Title, Body: n.Body})
	if err != nil {
		t.app.Logger.Error("desktop notification", "error", err)
	}
}
