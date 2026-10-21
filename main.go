package main

import (
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/db"
)

// The frontend build output is embedded in the binary.
//
//go:embed all:frontend/dist
var assets embed.FS

// snapshotEvent carries a compose.Snapshot to the frontend.
const snapshotEvent = "snapshot"

func init() {
	// Lets the binding generator emit a typed event for the frontend.
	application.RegisterEvent[compose.Snapshot](snapshotEvent)
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

	manager, err := compose.New(compose.Config{Docker: compose.ExecDocker{}, DB: conn})
	if err != nil {
		return err
	}

	app := application.New(application.Options{
		Name:        "Docker Controller",
		Description: "Gestione dei Progetti docker-compose",
		Services: []application.Service{
			application.NewService(&ProjectsService{manager: manager}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Docker Controller",
		Width:  1000,
		Height: 700,
		URL:    "/",
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Snapshot -> frontend event. The first value is the current snapshot.
	go func() {
		for s := range manager.Watch(ctx) {
			app.Event.Emit(snapshotEvent, s)
		}
	}()
	go manager.Run(ctx, func(err error) { app.Logger.Error("compose refresh", "error", err) })

	return app.Run()
}
