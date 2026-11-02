package activity

import (
	"testing"

	"dockercontroller/internal/compose"
)

// TestActivityTooltip is the outcome of an Activity as the row of the history
// shows it: the result of a finished Activity, the state of a running one. The
// toast of a failure carries the same text, so the two cannot disagree.
func TestActivityTooltip(t *testing.T) {
	for _, tc := range []struct {
		name     string
		activity Activity
		want     string
	}{
		{
			"a running Activity says so",
			Activity{State: ActivityRunning, Project: "shop", Kind: compose.ActionUp},
			RunningText,
		},
		{
			"a success names the Action and its target",
			Activity{State: ActivityOK, Project: "shop", Kind: compose.ActionUp},
			"Up su shop riuscito.",
		},
		{
			"a success on a Service names the Service",
			Activity{State: ActivityOK, Project: "shop", Service: "db", Kind: compose.ActionStart},
			"Start su shop/db riuscito.",
		},
		{
			"a failure carries the message of the notice",
			Activity{State: ActivityFailed, Project: "shop", Kind: compose.ActionUp, Summary: "no such image"},
			"Up su shop non riuscito: no such image",
		},
		{
			"a failure without a summary keeps the label",
			Activity{State: ActivityFailed, Project: "shop", Kind: compose.ActionUp, Error: "boom\n"},
			"Up su shop non riuscito",
		},
		{
			"a state the view does not know reads as running",
			Activity{State: ActivityState("queued"), Project: "shop", Kind: compose.ActionUp},
			RunningText,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Tooltip(tc.activity); got != tc.want {
				t.Errorf("Tooltip() = %q, want %q", got, tc.want)
			}
		})
	}
}
