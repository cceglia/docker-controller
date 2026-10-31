package main

import (
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/db"
	"dockercontroller/internal/settings"
)

// The frontend build output is embedded in the binary.
//
//go:embed all:frontend/dist
var assets embed.FS

// appIcon is the icon of the window (title bar, taskbar, about box).
//
//go:embed build/appicon.png
var appIcon []byte

// snapshotEvent carries a compose.Snapshot to the frontend.
const snapshotEvent = "snapshot"

// activityEvent carries a compose.ActivityEvent to the frontend.
const activityEvent = "activity"

func init() {
	application.RegisterEvent[LogEvent](logEvent)
	// Lets the binding generator emit a typed event for the frontend.
	application.RegisterEvent[compose.Snapshot](snapshotEvent)
	application.RegisterEvent[compose.ActivityEvent](activityEvent)
	application.RegisterEvent[compose.UsageSample](usageEvent)
	application.RegisterEvent[application.Void](usageStoppedEvent)
	application.RegisterEvent[application.Void](closeRequestedEvent)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dbPath, err := db.DefaultPath()
	if err != nil {
		return err
	}
	conn, err := db.Open(dbPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var app *application.App
	var systemTray *trayAdapter
	manager, err := compose.New(compose.Config{
		Docker: compose.ExecDocker{},
		DB:     conn,
		OnActivity: func(e compose.ActivityEvent) {
			app.Event.Emit(activityEvent, e)
			if systemTray != nil {
				systemTray.onActivity(e)
			}
		},
	})
	if err != nil {
		return err
	}

	notifs := notifications.New()
	settingsStore := settings.New(conn)
	projects := &ProjectsService{
		manager:  manager,
		settings: settingsStore,
		ctx:      ctx,
		emit:     func(name string, data any) { app.Event.Emit(name, data) },
	}
	settingsService := &SettingsService{store: settingsStore, ctx: ctx}
	app = application.New(application.Options{
		Name: "Docker Controller",
		Icon: appIcon,
		Linux: application.LinuxOptions{
			// The window's app id: it must match docker-controller.desktop so the
			// taskbar and title bar use its icon, whatever the executable is called.
			ProgramName: "docker-controller",
			// The application stays in the tray when the last window is destroyed.
			DisableQuitOnLastWindowClosed: true,
		},
		Description: "Gestione dei Progetti docker-compose",
		Services: []application.Service{
			application.NewService(projects),
			application.NewService(settingsService),
			application.NewService(notifs),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	// The usage stream runs while the window is visible: it starts with the
	// application (the window is created shown) and follows the window life
	// cycle.
	usage := &usageStreamer{
		manager: manager,
		ctx:     ctx,
		emit:    func(name string, data any) { app.Event.Emit(name, data) },
		onError: func(err error) { app.Logger.Error("usage stream", "error", err) },
	}
	window := newMainWindow(app, settingsStore, usage, projects)
	settingsService.closer = window
	window.install()
	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		usage.Start()
	})

	systemTray = newTrayAdapter(ctx, app, manager, window, notifs)

	// Snapshot -> frontend event. The first value is the current snapshot.
	go func() {
		for s := range manager.Watch(ctx) {
			app.Event.Emit(snapshotEvent, s)
		}
	}()
	go manager.Run(ctx, func(err error) { app.Logger.Error("compose refresh", "error", err) })

	return app.Run()
}
