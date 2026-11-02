package gui

import (
	"fmt"
	"reflect"
	"testing"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/guitest"
)

// sent is one notification the fake Sender received.
type sent struct {
	Title string
	Body  string
}

// fakeSender records the notifications without D-Bus.
type fakeSender struct {
	notifications []sent
}

func (f *fakeSender) Send(title, body string) {
	f.notifications = append(f.notifications, sent{Title: title, Body: body})
}

// clickableSender is a fake Sender that also reports a click, like the D-Bus
// sender does.
type clickableSender struct {
	fakeSender
	onClick func()
}

func (c *clickableSender) SetOnClick(fn func()) { c.onClick = fn }

// shop is the Project of the fixture: the notifications show its DisplayName.
var shop = compose.Project{Name: "shop", DisplayName: "Negozio"}

// launchOK is a Manager that accepts the Action.
func launchOK(compose.Action) (compose.ActivityID, error) { return "act-1", nil }

func TestNotifier_OnClick(t *testing.T) {
	t.Parallel()

	sender := &clickableSender{}
	n := NewNotifier(sender)
	var clicks int
	n.OnClick(func() { clicks++ })

	if sender.onClick == nil {
		t.Fatal("the click of a notification must reach a Sender that reports it")
	}
	sender.onClick()
	if clicks != 1 {
		t.Errorf("clicks = %d, want 1", clicks)
	}
}

// A Sender that cannot report a click still shows the notification, so the
// registration must be a no-op and not a failure.
func TestNotifier_OnClickWithoutAClickableSender(t *testing.T) {
	t.Parallel()

	n := NewNotifier(&fakeSender{})
	n.OnClick(func() { t.Error("a Sender without a click must not run the callback") })
}

func TestNotifier(t *testing.T) {
	for _, tc := range []struct {
		name  string
		steps func(n *Notifier)
		want  []sent
	}{
		{
			name: "a tray Activity notifies once when it succeeds",
			steps: func(n *Notifier) {
				n.launch(compose.Action{Project: "shop", Kind: compose.ActionUp}, launchOK)
				n.Observe(guitest.Started("act-1", "shop", "", compose.ActionUp))
				n.Observe(guitest.Finished("act-1", "shop", "", compose.ActionUp, "", ""))
				// The Manager sends one finished event for each Activity; the
				// rules must not notify again if a second one arrives.
				n.Observe(guitest.Finished("act-1", "shop", "", compose.ActionUp, "", ""))
			},
			want: []sent{{Title: windowTitle, Body: "Up di Negozio completato"}},
		},
		{
			name: "a tray Activity notifies when it fails",
			steps: func(n *Notifier) {
				n.launch(compose.Action{Project: "shop", Kind: compose.ActionRestart}, launchOK)
				n.Observe(guitest.Started("act-4", "shop", "", compose.ActionRestart))
				n.Observe(guitest.Finished("act-4", "shop", "", compose.ActionRestart,
					"Error response from daemon: restart failed",
					"Error response from daemon: restart failed"))
			},
			want: []sent{{Title: windowTitle, Body: "Restart di Negozio fallito: Error response from daemon: restart failed"}},
		},
		{
			name: "a launch the Manager refuses notifies and is forgotten",
			steps: func(n *Notifier) {
				n.launch(compose.Action{Project: "shop", Kind: compose.ActionUp}, func(compose.Action) (compose.ActivityID, error) {
					return "", fmt.Errorf("%w: shop", compose.ErrActivityRunning)
				})
				// The refusal withdrew the launch: the success of a later
				// window Activity must stay silent.
				n.Observe(guitest.Started("act-2", "shop", "", compose.ActionUp))
				n.Observe(guitest.Finished("act-2", "shop", "", compose.ActionUp, "", ""))
			},
			want: []sent{{Title: windowTitle, Body: "C'è già un'Attività in corso su questo Progetto."}},
		},
		{
			name: "a failed window Activity notifies",
			steps: func(n *Notifier) {
				n.Observe(guitest.Started("act-7", "shop", "", compose.ActionPull))
				n.Observe(guitest.Finished("act-7", "shop", "", compose.ActionPull,
					"progress\nError response from daemon: pull access denied",
					"Error response from daemon: pull access denied"))
			},
			want: []sent{{Title: windowTitle, Body: "Pull di Negozio fallito: Error response from daemon: pull access denied"}},
		},
		{
			name: "a Project outside the snapshot keeps its compose name",
			steps: func(n *Notifier) {
				n.Observe(guitest.Started("act-8", "raw", "", compose.ActionUp))
				n.Observe(guitest.Finished("act-8", "raw", "", compose.ActionUp, "boom", "boom"))
			},
			want: []sent{{Title: windowTitle, Body: "Up di raw fallito: boom"}},
		},
		{
			name: "a successful window Activity is silent",
			steps: func(n *Notifier) {
				n.Observe(guitest.Started("act-3", "shop", "", compose.ActionUp))
				n.Observe(guitest.Finished("act-3", "shop", "", compose.ActionUp, "", ""))
			},
			want: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fake := &fakeSender{}
			n := NewNotifier(fake)
			n.SetProjects([]compose.Project{shop})
			tc.steps(n)

			if !reflect.DeepEqual(fake.notifications, tc.want) {
				t.Errorf("notifications = %+v, want %+v", fake.notifications, tc.want)
			}
		})
	}
}
