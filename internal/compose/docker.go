package compose

import "dockercontroller/internal/compose/docker"

// Docker is the port towards the Docker CLI (ADR-0002). It lives in the docker
// subpackage, and the alias keeps the root interface that gui, tray and main
// already use.
type Docker = docker.Docker

// ExecDocker is the production adapter of the Docker port (ADR-0002). The
// alias keeps the wiring in main.go unchanged.
type ExecDocker = docker.ExecDocker
