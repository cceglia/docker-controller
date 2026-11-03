// Package logs holds the lifecycle of the log streams and the model of the Log
// view: the lines kept and filtered, the segments of a line, the palette entry
// of a Service, the end of a stream and its Italian message.
//
// Dependency direction: design and view are leaves. logs imports them and
// compose, and it never imports the root package gui, which imports logs. It
// uses no Qt binding, so its tests run without a display.
package logs
