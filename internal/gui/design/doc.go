// Package design holds the design tokens of the window: the type scale, the
// spacing rhythm, the colours of the semantic roles, the minimum target of an
// interactive control, the style sheet built from them and the colour of one
// entry of the Service palette.
//
// Dependency direction: design is a leaf. The root package gui imports it, and
// design imports no gui package and no Qt binding, so its test runs without a
// display.
package design
