package notify

import (
	"reflect"
	"testing"

	"github.com/godbus/dbus/v5"
)

// clicker counts the clicks the Sender reports to its callback.
type clicker struct {
	clicks int
}

func (c *clicker) on() { c.clicks++ }

// actionSignal is the signal the notification server emits when the user
// clicks a notification that carries an action.
func actionSignal(id uint32, action string) *dbus.Signal {
	return &dbus.Signal{Name: actionInvoked, Body: []any{id, action}}
}

func TestPending(t *testing.T) {
	var p pending

	if p.take(7) {
		t.Fatal("an id never sent must not be taken")
	}
	p.add(7)
	if !p.take(7) {
		t.Fatal("an id sent must be taken")
	}
	if p.take(7) {
		t.Fatal("a consumed id must not be taken again")
	}
}

func TestPending_KeepsTheSetBounded(t *testing.T) {
	var p pending
	for id := uint32(1); id <= pendingLimit+1; id++ {
		p.add(id)
	}
	if p.take(1) {
		t.Fatal("the oldest id must leave when the set is full")
	}
	if !p.take(pendingLimit + 1) {
		t.Fatal("the newest id must stay")
	}
}

func TestDBus_Handle(t *testing.T) {
	for _, tc := range []struct {
		name    string
		sent    []uint32
		signals []*dbus.Signal
		want    int
	}{
		{
			name:    "a click on a notification we sent reports the click",
			sent:    []uint32{7},
			signals: []*dbus.Signal{actionSignal(7, defaultAction)},
			want:    1,
		},
		{
			name:    "a click on an id of another application does nothing",
			sent:    []uint32{7},
			signals: []*dbus.Signal{actionSignal(8, defaultAction)},
			want:    0,
		},
		{
			name:    "a second click on the same notification reports one click",
			sent:    []uint32{7},
			signals: []*dbus.Signal{actionSignal(7, defaultAction), actionSignal(7, defaultAction)},
			want:    1,
		},
		{
			name:    "an action other than default does nothing",
			sent:    []uint32{7},
			signals: []*dbus.Signal{actionSignal(7, "reply")},
			want:    0,
		},
		{
			name: "another signal does nothing",
			sent: []uint32{7},
			signals: []*dbus.Signal{{
				Name: "org.freedesktop.Notifications.NotificationClosed",
				Body: []any{uint32(7), uint32(2)},
			}},
			want: 0,
		},
		{
			name:    "a signal without the two arguments does nothing",
			sent:    []uint32{7},
			signals: []*dbus.Signal{{Name: actionInvoked, Body: []any{uint32(7)}}},
			want:    0,
		},
		{
			name: "a signal with an id of another type does nothing",
			sent: []uint32{7},
			signals: []*dbus.Signal{{
				Name: actionInvoked,
				Body: []any{int32(7), defaultAction},
			}},
			want: 0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := NewDBus()
			var got clicker
			d.SetOnClick(got.on)
			for _, id := range tc.sent {
				d.pending.add(id)
			}
			for _, signal := range tc.signals {
				d.handle(signal)
			}
			if got.clicks != tc.want {
				t.Errorf("clicks = %d, want %d", got.clicks, tc.want)
			}
		})
	}
}

func TestNotifyArgs(t *testing.T) {
	args := notifyArgs("Docker Controller", "Up di Negozio completato")

	if len(args) != 8 {
		t.Fatalf("arguments = %d, want the 8 of Notify", len(args))
	}
	actions, ok := args[5].([]string)
	if !ok {
		t.Fatalf("actions = %T, want []string", args[5])
	}
	if want := []string{"default", ""}; !reflect.DeepEqual(actions, want) {
		t.Errorf("actions = %q, want %q", actions, want)
	}
	if args[3] != "Docker Controller" || args[4] != "Up di Negozio completato" {
		t.Errorf("title and body = %v, %v, want the two strings", args[3], args[4])
	}
}
