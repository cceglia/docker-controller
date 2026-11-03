// Package model holds the compose domain: Projects, Services, Containers,
// Snapshots, Actions, Activities, Settings, Logs, the Measures of the Utilizzo
// risorse, the pure helpers StatusOf and Folder, and the sentinel errors the
// subpackages share.
//
// Dependency direction: model is the leaf. The store and parse subpackages and
// the root package compose import model. The docker subpackage is independent
// of it. model imports none of them, so it never creates a cycle.
package model
