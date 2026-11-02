package gui

import (
	"github.com/mappu/miqt/qt6"

	"dockercontroller/internal/gui/design"
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
)

// tackStrip draws the tacks of a Project. It carries the selection itself: the
// tacks are painted, so the highlighted text of a selected row does not reach
// them through the palette, which the desktop style owns.
type tackStrip struct {
	*qt6.QWidget
	selected bool
}

// newTackStrip draws the tacks of a Project: filled for a running Container,
// faint for a stopped one and in the fault role for an unhealthy or dead one.
// A strip without tacks is empty but keeps its place, so the rows line up.
func newTackStrip(states []TackState, theme design.DesignTheme) *tackStrip {
	width := 0
	if len(states) > 0 {
		width = len(states)*(tackWidth+tackGap) - tackGap
	}

	strip := &tackStrip{QWidget: qt6.NewQWidget2()}
	strip.SetFixedSize2(width, tackHeight)
	strip.SetAttribute2(qt6.WA_TransparentForMouseEvents, true)

	strip.OnPaintEvent(func(super func(event *qt6.QPaintEvent), event *qt6.QPaintEvent) {
		painter := qt6.NewQPainter2(strip.QPaintDevice)
		defer painter.End()

		role := qt6.QPalette__WindowText
		if strip.selected {
			role = qt6.QPalette__HighlightedText
		}
		foreground := strip.Palette().ColorWithCr(role)
		fault := qt6.NewQColor6(theme.Color(design.RoleTackFault))
		if strip.selected {
			// The fault colour keeps its meaning, but it must read on the
			// highlight: the highlighted text is the one colour the desktop
			// guarantees there.
			fault = qt6.NewQColor10(foreground)
			fault.SetAlpha(192)
		}

		for i, state := range states {
			color := qt6.NewQColor10(foreground)
			switch state {
			case TackFaulty:
				color = qt6.NewQColor10(fault)
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

// setSelected repaints the tacks with the colour of the selected row.
func (s *tackStrip) setSelected(selected bool) {
	if s == nil || s.selected == selected {
		return
	}
	s.selected = selected
	s.Update()
}
