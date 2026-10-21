package main

import (
	"context"

	"dockercontroller/internal/compose"
)

// ProjectsService is the thin Wails binding over the compose module. The
// frontend receives snapshots through the "snapshot" event; Snapshot only
// serves the initial load.
type ProjectsService struct {
	manager *compose.Manager
	// ctx bounds the Activities: they are stopped when the app quits.
	ctx context.Context
}

// Snapshot returns the current snapshot of the saved Projects.
func (s *ProjectsService) Snapshot() compose.Snapshot {
	return s.manager.Snapshot()
}

// Do starts an Action on a Project and returns the id of its Activity. Output
// and outcome arrive through the "activity" event.
func (s *ProjectsService) Do(action compose.Action) (compose.ActivityID, error) {
	return s.manager.Do(s.ctx, action)
}
