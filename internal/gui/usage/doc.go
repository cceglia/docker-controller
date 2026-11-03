// Package usage holds the Utilizzo risorse: the feed that follows Docker, the
// stream of `docker stats`, the state of the bar and of the block of a
// Service, the texts of a row and the steps of a bar.
//
// Dependency direction: view is a leaf. usage imports view and compose, and it
// never imports the root package gui, which imports usage. It uses no Qt
// binding, so its tests run without a display.
package usage
