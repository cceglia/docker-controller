// Package compose is the deep module holding the domain logic of the app: it
// discovers docker-compose Projects through the Docker CLI (ADR-0002), saves
// them, computes their status and tells subscribers when something changed.
//
// The root holds Manager, its Config and New, and the methods of the module:
// Snapshot, Refresh, Watch, AddProject, RemoveFromList, UpdateSettings, Do
// (Actions, run as Activities), Logs (log streams) and Usage (Utilizzo risorse
// samples, a flow separate from Watch). Window and tray drive the module
// through Manager.
//
// Docker is reached through the port in the docker subpackage, so tests replace
// it with the fake under docker/fake, fed with real captured output. The domain
// types, the pure helpers and the sentinel errors live in model, the SQLite
// persistence in store, and the parsing of the CLI output in parse.
//
// The root re-exports with aliases the surface that gui, tray and main use.
// Dependency direction: model ← docker/store/parse ← root. No subpackage
// imports the root.
package compose
