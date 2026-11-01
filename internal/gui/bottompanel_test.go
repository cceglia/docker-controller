package gui

import (
	"testing"

	"dockercontroller/internal/compose"
)

func TestPanelActivity(t *testing.T) {
	up := Activity{ID: "a1", State: ActivityRunning, Project: "shop", Service: "db", Kind: compose.ActionUp}
	down := Activity{ID: "a2", State: ActivityRunning, Project: "shop", Kind: compose.ActionDown}

	for _, tc := range []struct {
		name       string
		shown      map[compose.ActivityID]bool
		activities []Activity
		wantID     compose.ActivityID
		wantOpen   bool
	}{
		{
			name: "empty history",
		},
		{
			name:       "the first Activity opens the panel",
			activities: []Activity{up},
			wantID:     "a1",
			wantOpen:   true,
		},
		{
			name:       "an output line of a shown Activity leaves the panel closed",
			shown:      map[compose.ActivityID]bool{"a1": true},
			activities: []Activity{up},
		},
		{
			name:       "a new Activity opens on the newest of the history",
			shown:      map[compose.ActivityID]bool{"a1": true},
			activities: []Activity{down, up},
			wantID:     "a2",
			wantOpen:   true,
		},
		{
			name:       "two new Activities open once, on the newest",
			activities: []Activity{down, up},
			wantID:     "a2",
			wantOpen:   true,
		},
		{
			name:       "a shown Activity newer than the new one does not hide it",
			shown:      map[compose.ActivityID]bool{"a2": true},
			activities: []Activity{down, up},
			wantID:     "a1",
			wantOpen:   true,
		},
		{
			name:       "a failed Activity already shown keeps the panel as it is",
			shown:      map[compose.ActivityID]bool{"a1": true},
			activities: []Activity{{ID: "a1", State: ActivityFailed, Project: "shop", Kind: compose.ActionUp}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id, open := panelActivity(tc.shown, tc.activities)
			if id != tc.wantID || open != tc.wantOpen {
				t.Errorf("panelActivity() = (%q, %t), want (%q, %t)", id, open, tc.wantID, tc.wantOpen)
			}
		})
	}
}
