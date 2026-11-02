package model

import "errors"

var (
	// ErrActivityRunning is returned when the Project already has an Activity.
	ErrActivityRunning = errors.New("compose: the Project already has an Activity in progress")
	// ErrUnknownProject is returned for a Project that is not saved.
	ErrUnknownProject = errors.New("compose: unknown Project")
	// ErrUnknownAction is returned for an ActionKind that does not exist.
	ErrUnknownAction = errors.New("compose: unknown Action")
	// ErrInvalidService is returned for a Service name compose could take for
	// an option.
	ErrInvalidService = errors.New("compose: invalid Service name")
	// ErrServiceOnly is returned for Down of a Service on the whole Project.
	ErrServiceOnly = errors.New("compose: this Action applies to a Service, not to the whole Project")
	// ErrProjectOnly is returned for Down or Down with volumes on a Service.
	ErrProjectOnly = errors.New("compose: this Action applies to the whole Project, not to a Service")

	// ErrDuplicateProject is returned by AddProject when a Project with that
	// name is already saved.
	ErrDuplicateProject = errors.New("compose: a Project with this name already exists")
	// ErrInvalidProject is returned for a Project that compose could not use: a
	// name it would reject, no compose file, or a file that cannot be read.
	ErrInvalidProject = errors.New("compose: invalid Project")
	// The specific reasons of an ErrInvalidProject; each error wraps both.
	ErrInvalidName         = errors.New("name is not a valid compose name")
	ErrNoComposeFile       = errors.New("no compose file")
	ErrComposeFileUnusable = errors.New("compose file is not an absolute path to a readable file")

	// ErrProjectNotRemoved is returned by RemoveFromList for a Project that is
	// not Rimosso or still has Containers.
	ErrProjectNotRemoved = errors.New("compose: only a Rimosso Project can be removed from the list")
	// ErrDockerUnreachable is returned when the answer depends on what Docker
	// reports and Docker does not answer.
	ErrDockerUnreachable = errors.New("compose: Docker is not reachable")
)
