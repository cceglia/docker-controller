// Package parse reads the output of the Docker and Compose CLI: the Project
// list, the Container state, Log lines, Utilizzo risorse, Profiles and the size
// of the machine. It holds the parsing support types Machine and StatsRow.
//
// Dependency direction: model ← parse ← root package compose. parse imports
// model and never imports the root package.
package parse
