package view

import (
	"testing"

	"dockercontroller/internal/compose"
)

func TestProjectName_PrefersTheDisplayName(t *testing.T) {
	for _, tc := range []struct {
		name    string
		project compose.Project
		want    string
	}{
		{"display name", compose.Project{Name: "shop", DisplayName: "Negozio"}, "Negozio"},
		{"compose name when the display name is empty", compose.Project{Name: "shop"}, "shop"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ProjectName(tc.project); got != tc.want {
				t.Errorf("ProjectName = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDisplayNameOf_FallsBackToTheComposeName(t *testing.T) {
	for _, tc := range []struct {
		name        string
		displayName string
		composeName string
		want        string
	}{
		{"a Display name wins", "Negozio", "shop", "Negozio"},
		{"a blank Display name falls back", "", "shop", "shop"},
		{"spaces fall back", "   ", "shop", "shop"},
		{"surrounding spaces go", "  Negozio ", "shop", "Negozio"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := DisplayNameOf(tc.displayName, tc.composeName); got != tc.want {
				t.Errorf("DisplayNameOf = %q, want %q", got, tc.want)
			}
		})
	}
}
