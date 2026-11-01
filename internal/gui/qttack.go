package gui

import (
	"github.com/mappu/miqt/qt6"
)

const (
	// tackWidth, tackHeight and tackGap size the tacks of the strip, in
	// logical pixels.
	tackWidth  = 3
	tackHeight = 12
	tackGap    = 2
	// tackStoppedAlpha is how faint a stopped Container reads: the tack is
	// still there, so the count of the Project stays readable.
	tackStoppedAlpha = 64
	// tackFaultColor is the fallback colour of a faulty Container: the palette
	// has no error role, so the colour is declared next to its role.
	tackFaultColor = "#da4453"
)

// newTackStrip draws the tacks of a Project: filled for a running Container,
// faint for a stopped one and in the fault colour for an unhealthy or dead one.
// A strip without tacks is empty but keeps its place, so the rows line up.
func newTackStrip(states []TackState) *qt6.QWidget {
	width := 0
	if len(states) > 0 {
		width = len(states)*(tackWidth+tackGap) - tackGap
	}

	strip := qt6.NewQWidget2()
	strip.SetFixedSize2(width, tackHeight)
	strip.SetAttribute2(qt6.WA_TransparentForMouseEvents, true)

	strip.OnPaintEvent(func(super func(event *qt6.QPaintEvent), event *qt6.QPaintEvent) {
		painter := qt6.NewQPainter2(strip.QPaintDevice)
		defer painter.End()

		foreground := strip.Palette().ColorWithCr(qt6.QPalette__WindowText)
		fault := qt6.NewQColor6(tackFaultColor)

		for i, state := range states {
			color := qt6.NewQColor10(foreground)
			switch state {
			case TackFaulty:
				color = fault
			case TackRunning:
				// The full foreground colour needs no change.
			default:
				color.SetAlpha(tackStoppedAlpha)
			}
			x := i * (tackWidth + tackGap)
			painter.FillRect2(x, 0, tackWidth, tackHeight, qt6.NewQBrush3(color))
		}
	})

	return strip
}
