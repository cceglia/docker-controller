package compose

import "dockercontroller/internal/compose/docker"

// Docker is the port towards the Docker CLI (ADR-0002). The alias keeps the
// surface of the root package unchanged.
type Docker = docker.Docker

// ExecDocker is the production adapter of the Docker port (ADR-0002). The
// alias keeps the wiring in main.go unchanged.
type ExecDocker = docker.ExecDocker
