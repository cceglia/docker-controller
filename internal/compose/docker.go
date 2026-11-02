package compose

import "dockercontroller/internal/compose/docker"

// ExecDocker is the production adapter of the Docker port (ADR-0002). The
// alias keeps the wiring in main.go unchanged.
type ExecDocker = docker.ExecDocker
