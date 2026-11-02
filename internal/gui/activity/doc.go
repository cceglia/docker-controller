// Package activity holds the Activities of the session: the log the Manager
// feeds, the texts of a row and of its output, and the model and the texts of
// the bottom panel.
//
// Dependency direction: activity imports compose and project and no Qt
// binding, so its tests run without a display. It does not import the root
// package gui, which imports activity.
package activity
