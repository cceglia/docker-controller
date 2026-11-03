// Package view holds the presentation the gui shares: the Italian number
// formats, the state and the tacks of a Project, the sections of the list, the
// names and the notices of a Project, the Italian text of a known error and the
// commands that open a folder, a terminal or an address.
//
// Dependency direction: view is a leaf. The root package gui and the feature
// packages import it, and view imports no root package `internal/gui` and no Qt
// binding, so its tests run without a display.
package view
