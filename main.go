package main

import (
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/db"
)

// The frontend build output is embedded in the binary.
//
//go:embed all:frontend/dist
var assets embed.FS

// snapshotEvent carries a compose.Snapshot to the frontend.
const snapshotEvent = "snapshot"

// activityEvent carries a compose.ActivityEvent to the frontend.
const activityEvent = "activity"

func init() {
	application.RegisterEvent[LogEvent](logEvent)
	// Lets the binding generator emit a typed event for the frontend.
	application.RegisterEvent[compose.Snapshot](snapshotEvent)
	application.RegisterEvent[compose.ActivityEvent](activityEvent)
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
	app = application.New(application.Options{
		Name:        "Docker Controller",
		Description: "Gestione dei Progetti docker-compose",
		Services: []application.Service{
			application.NewService(&ProjectsService{manager: manager, ctx: ctx, emit: func(name string, data any) { app.Event.Emit(name, data) }}),
			application.NewService(notifs),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Docker Controller",
		Width:  1000,
		Height: 700,
		URL:    "/",
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
