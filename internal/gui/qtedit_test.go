package gui

import "testing"

// TestFieldAccessibleName_RepeatsTheMessageOfTheField checks the accessible
// name of a field of the form: the caption alone while the field is right, the
// caption and the message when it is not, so a screen reader reads the error
// with the field.
func TestFieldAccessibleName_RepeatsTheMessageOfTheField(t *testing.T) {
	for _, tc := range []struct {
		name    string
		caption string
		message string
		want    string
	}{
		{"a field with no message", "Nome visualizzato", "", "Nome visualizzato"},
		{"a field with a message", "Nome visualizzato", "Progetto sconosciuto.", "Nome visualizzato. Progetto sconosciuto."},
		{"a Group field with a message", "Gruppo", "Gruppo non valido.", "Gruppo. Gruppo non valido."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := fieldAccessibleName(tc.caption, tc.message); got != tc.want {
				t.Errorf("fieldAccessibleName = %q, want %q", got, tc.want)
			}
		})
	}
}
