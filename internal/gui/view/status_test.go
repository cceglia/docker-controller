package view

import (
	"testing"

	"dockercontroller/internal/compose"
)

func TestStatusOf_ReadsTheFourStatusesInItalian(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   compose.Status
		label    string
		glyph    string
		cssClass string
	}{
		{"active", compose.StatusActive, "Attivo", "●", "success"},
		{"partial", compose.StatusPartial, "Parziale", "◐", "warning"},
		{"stopped", compose.StatusStopped, "Fermo", "○", "dim-label"},
		{"removed", compose.StatusRemoved, "Rimosso", "○", "dim-label"},
		{"unknown falls back to stopped", compose.Status("weird"), "weird", "○", "dim-label"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := StatusOf(tc.status)
			if got.Label != tc.label {
				t.Errorf("Label = %q, want %q", got.Label, tc.label)
			}
			if got.Glyph != tc.glyph {
				t.Errorf("Glyph = %q, want %q", got.Glyph, tc.glyph)
			}
			if got.CSSClass != tc.cssClass {
				t.Errorf("CSSClass = %q, want %q", got.CSSClass, tc.cssClass)
			}
		})
	}
}
