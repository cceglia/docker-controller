package view

import (
	"fmt"

	"dockercontroller/internal/compose"
)

// StatusView is how a Project status reads in the list: the Italian label, the
// glyph of the state and the Adwaita class that tints it. The label is never
// drawn in the row; it is its tooltip and its accessible name.
type StatusView struct {
	Label    string
	Glyph    string
	CSSClass string
}

// StatusOf returns the view of a Project status. A status outside the four
// known ones reads as Fermo and keeps its own label, so nothing is hidden.
// The glyphs are the ones of the tray: ●, ◐ and ○.
func StatusOf(status compose.Status) StatusView {
	switch status {
	case compose.StatusActive:
		return StatusView{Label: "Attivo", Glyph: "●", CSSClass: "success"}
	case compose.StatusPartial:
		return StatusView{Label: "Parziale", Glyph: "◐", CSSClass: "warning"}
	case compose.StatusStopped:
		return StatusView{Label: "Fermo", Glyph: "○", CSSClass: "dim-label"}
	case compose.StatusRemoved:
		return StatusView{Label: "Rimosso", Glyph: "○", CSSClass: "dim-label"}
	}
	return StatusView{Label: string(status), Glyph: "○", CSSClass: "dim-label"}
}

// ProjectCount is the count of the running Containers over the total, as the
// row shows it ("3/5").
func ProjectCount(p compose.Project) string {
	return fmt.Sprintf("%d/%d", p.Running, p.Total)
}
