// Package project holds the presentation of a Project: the Actions and their
// labels, the question before a destructive Action, the entry that takes a
// Project off the list, the lines of the detail and the state that decides when
// to draw the panel again.
//
// Dependency direction: project imports compose and view. The root package gui
// imports project, and project imports no gui package and no Qt binding, so its
// tests run without a display.
package project
