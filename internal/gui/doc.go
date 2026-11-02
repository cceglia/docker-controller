// Package gui is the composition of the window: the application, the shell,
// the Qt views of the list and of the detail, the dialogs, the bottom panel and
// the shared widgets, which are the notifier, the start guard and the window
// constants. The Qt adapters live here, so this is the only gui package that
// imports the Qt binding.
//
// Dependency direction: gui imports design, view and the feature packages
// (activity, forms, logs, project, traymenu, usage), and no subpackage imports
// gui. Run is the entry point of the UI.
package gui
