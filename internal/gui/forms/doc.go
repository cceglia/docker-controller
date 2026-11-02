// Package forms holds the models of the forms of the window: the new Project,
// the settings of a Project, the settings of the application and the close
// behavior. A form keeps what the user typed and hands it to the Manager only
// when the user confirms it.
//
// Dependency direction: forms imports compose, settings and view. The root
// package gui imports forms, and forms imports no gui package and no Qt binding,
// so its tests run without a display.
package forms
