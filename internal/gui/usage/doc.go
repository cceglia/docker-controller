// Package usage holds the Utilizzo risorse: the feed that follows Docker, the
// stream of `docker stats`, the state of the bar and of the block of a
// Service, the texts of a row and the steps of a bar.
//
// Dependency direction: design and view are leaves. The root package gui
// imports usage, and usage imports view and no internal/gui package and no Qt
// binding, so its tests run without a display.
package usage
