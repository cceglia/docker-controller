// Package docker holds the only place of compose that runs a process: the
// Docker port and its production adapter (ADR-0002). The test adapter lives
// under docker/fake and satisfies the same port.
//
// Dependency direction: model ← docker ← root. This package does not import
// the compose root, so tests and the fake need no cgo and no Docker daemon.
package docker
