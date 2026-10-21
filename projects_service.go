package main

import (
	"dockercontroller/internal/compose"
)

// ProjectsService is the thin Wails binding over the compose module. The
// frontend receives snapshots through the "snapshot" event; Snapshot only
// serves the initial load.
type ProjectsService struct {
	manager *compose.Manager
}

// Snapshot returns the current snapshot of the saved Projects.
func (s *ProjectsService) Snapshot() compose.Snapshot {
	return s.manager.Snapshot()
}
