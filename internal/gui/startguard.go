package gui

// startGuard stops one click from starting two Activities. The interval it
// covers is the one between the click on a command and the result of the
// Manager: the snapshot that reports the Activity arrives with that result, and
// the buttons of the panel are insensitive from then on.
//
// The panel holds one guard. Every command of the panel goes through it, so a
// second click on any command of the same Project is refused while the first
// launch waits for its result.
type startGuard struct {
	// closed says a launch waits for its result.
	closed bool
}

// begin reports whether a click may reach the Manager, and closes the guard.
// The second click of a double click gets false, so it reaches nothing.
func (g *startGuard) begin() bool {
	if g.closed {
		return false
	}
	g.closed = true
	return true
}

// release opens the guard: the Manager answered, or the snapshot reports the
// Activity, so the commands follow the data again.
func (g *startGuard) release() {
	g.closed = false
}

// snapshot opens the guard when the snapshot reports an Activity on the
// Project: the Manager holds the Project busy, and the commands are insensitive
// from the data itself.
func (g *startGuard) snapshot(busy bool) {
	if busy {
		g.closed = false
	}
}

// waiting reports whether a launch waits for its result.
func (g *startGuard) waiting() bool {
	return g.closed
}
