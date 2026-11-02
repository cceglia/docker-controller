// Package traymenu holds the tree of the tray menu: it maps the description of
// internal/tray to the nodes the Qt adapter applies, and it answers the
// questions the adapter asks about them, the tooltip of the icon and the
// confirmation before a destructive Action.
//
// Dependency direction: traymenu imports compose, view, project and tray, and
// it never imports the root package gui. It uses no Qt binding, so its tests run
// without a display.
package traymenu
