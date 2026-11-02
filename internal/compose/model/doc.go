// Package model holds the compose domain: Projects, Services, Containers,
// Snapshots, Actions, Activities, Settings, Logs, the pure helpers StatusOf
// and Folder, and the sentinel errors the subpackages share.
//
// Dependency direction: model is the leaf. The docker, store and parse
// subpackages and the root package compose import model. model imports none of
// them, so it never creates a cycle.
package model
