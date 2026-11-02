package compose

import "dockercontroller/internal/compose/model"

// The domain types below live in model. The root re-exports the ones that gui,
// tray and main use, so those callers keep the compose.Project spelling.
type (
	Project   = model.Project
	Service   = model.Service
	Container = model.Container
	Port      = model.Port
	Snapshot  = model.Snapshot
	Status    = model.Status
)

const (
	StatusActive  = model.StatusActive
	StatusPartial = model.StatusPartial
	StatusStopped = model.StatusStopped
	StatusRemoved = model.StatusRemoved
)
