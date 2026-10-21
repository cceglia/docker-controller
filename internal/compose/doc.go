// Package compose is the deep module holding the domain logic of the app: it
// discovers docker-compose Projects through the Docker CLI (ADR-0002), saves
// them, computes their status and tells subscribers when something changed.
//
// Window and tray only use Manager: Snapshot, Refresh, Watch, AddProject,
// RemoveFromList, Do (Actions, run as Activities) and Logs (log streams).
// Docker is reached through the Docker port (see docker.go), so tests replace
// it with composefake fed with real captured output.
package compose
