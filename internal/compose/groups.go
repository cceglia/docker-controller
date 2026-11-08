package compose

import "dockercontroller/internal/compose/model"

// The Group errors and the duplicate report live in model; the root re-exports
// them for gui, tray and main.
var (
	ErrUnknownGroup     = model.ErrUnknownGroup
	ErrDuplicateGroup   = model.ErrDuplicateGroup
	ErrInvalidGroupName = model.ErrInvalidGroupName
)

// DuplicateGroupError reports the saved spelling of the Group that already has
// a name.
type DuplicateGroupError = model.DuplicateGroupError
