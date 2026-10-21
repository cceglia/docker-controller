// Package compose is the deep module holding the domain logic of the app: it
// discovers docker-compose Projects through the Docker CLI (ADR-0002), saves
// them, computes their status and tells subscribers when something changed.
//
// Window and tray only use Manager: Snapshot, Refresh, Watch, AddProject, RemoveFromList and Do (Actions, run as
// Activities) today; Logs, project settings and manual add will join it in later
// tickets. Docker is reached through the Docker port (see docker.go), so tests
// replace it with composefake fed with real captured output.
package compose
