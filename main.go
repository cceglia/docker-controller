package main

import (
	"context"
	"log"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/db"
	"dockercontroller/internal/gui"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run wires the database, the compose Manager and the GTK4 shell, then blocks
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

	// The log of the session lives in the GUI package, without GTK: the Manager
	// pushes its events into it and never waits for the window.
	activities := gui.NewActivityLog()

	manager, err := compose.New(compose.Config{
		Docker:     compose.ExecDocker{},
		DB:         conn,
		OnActivity: activities.Push,
	})
	if err != nil {
		return err
	}

	go manager.Run(ctx, func(err error) { log.Printf("compose refresh: %v", err) })

	return gui.Run(ctx, manager, activities)
}
