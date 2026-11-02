package main

import (
	"context"
	"log"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/db"
	"dockercontroller/internal/gui"
	"dockercontroller/internal/gui/activity"
	"dockercontroller/internal/notify"
	"dockercontroller/internal/settings"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run wires the database, the compose Manager and the Qt shell, then blocks
// until the application quits.
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

	// The log of the session lives in the activity package, without Qt: the
	// Manager pushes its events into it and never waits for the window. The
	// notifier reads the same events and shows the notifications they deserve.
	activities := activity.NewLog()
	notifier := gui.NewNotifier(notify.NewDBus())

	manager, err := compose.New(compose.Config{
		Docker: compose.ExecDocker{},
		DB:     conn,
		OnActivity: func(e compose.ActivityEvent) {
			activities.Push(e)
			notifier.Observe(e)
		},
	})
	if err != nil {
		return err
	}

	go manager.Run(ctx, func(err error) { log.Printf("compose refresh: %v", err) })

	return gui.Run(ctx, manager, activities, settings.New(conn), notifier)
}
