// Package notify shows desktop notifications through the notification service
// of the desktop and reports the click on them. The transport is
// org.freedesktop.Notifications on the session bus, the service every desktop
// understands; nothing here depends on GTK, so the app keeps working when the
// service is missing.
package notify

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
)

// Sender shows a desktop notification. Send must return at once: the transport
// does the work in the background, because the D-Bus call blocks and would
// otherwise stop the Qt loop and the Activity that produced the notification.
type Sender interface {
	Send(title, body string)
}

// Clickable is a Sender that also reports the click on a notification it sent.
// A Sender that cannot is still usable: the notification shows and no click
// arrives.
type Clickable interface {
	Sender
	// SetOnClick registers the function that runs on a click. The transport
	// calls it on its own goroutine, so a callback that touches a window posts
	// the work with mainthread.Start.
	SetOnClick(fn func())
}

const (
	// appName is the name the notification server shows as the source.
	appName = "Docker Controller"
	// destination, path and method are the service of the desktop.
	destination = "org.freedesktop.Notifications"
	path        = dbus.ObjectPath("/org/freedesktop/Notifications")
	method      = "org.freedesktop.Notifications.Notify"
	// actionInvoked is the signal the server emits when the user clicks a
	// notification that carries an action.
	actionInvoked = "org.freedesktop.Notifications.ActionInvoked"
	// member is the name of that signal in the match rule.
	member = "ActionInvoked"
	// defaultAction is the action every notification carries: the click on the
	// body of the notification. The label is empty, so the server shows no
	// button.
	defaultAction = "default"
	// callTimeout bounds one call, so a service that does not answer leaves no
	// goroutine behind.
	callTimeout = 5 * time.Second
)

// DBus is the Sender of the session bus. It sends the notifications and listens
// for the click on them, once for the whole application: the match rule and the
// goroutine of the signals start with the first notification. It is safe for
// concurrent use.
type DBus struct {
	// mu guards onClick and listening, which the sender and the signal
	// goroutine share.
	mu sync.Mutex
	// onClick runs when the user clicks one of the notifications below. A nil
	// callback drops the click and keeps the notification on screen.
	onClick func()
	// pending is the ids of this application still waiting for a click.
	pending pending
	// listening says the match rule and the goroutine of the signals exist.
	listening bool
}

// NewDBus returns a Sender on org.freedesktop.Notifications of the session bus.
func NewDBus() *DBus { return &DBus{} }

// SetOnClick registers the function that runs when the user clicks one of the
// notifications of this Sender. The function runs on the D-Bus goroutine.
func (d *DBus) SetOnClick(fn func()) {
	d.mu.Lock()
	d.onClick = fn
	d.mu.Unlock()
}

// Send hands the notification to the service on its own goroutine. A service
// that does not answer is not a failure of the app: the error is logged and the
// app goes on.
func (d *DBus) Send(title, body string) {
	go func() {
		if err := d.notify(title, body); err != nil {
			log.Printf("desktop notification: %v", err)
		}
	}()
}

// pendingLimit bounds the ids that wait for a click. A notification the user
// ignores never leaves the set, and the set must not grow for a whole session.
const pendingLimit = 64

// pending is the set of notification ids this application sent and that wait
// for a click. It is the rule that separates a click on our notification from a
// click on the notification of another application: only an id in the set is
// ours, and taking it makes a second click on the same id a stranger. It has no
// bus and no lock of the Sender, so a test drives it on its own.
type pending struct {
	mu  sync.Mutex
	ids map[uint32]struct{}
	// order is the ids oldest first, so the oldest one leaves when the set is
	// full.
	order []uint32
}

// add records an id that waits for a click and forgets the oldest one when the
// set is full.
func (p *pending) add(id uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ids == nil {
		p.ids = make(map[uint32]struct{})
	}
	if _, ok := p.ids[id]; ok {
		return
	}
	if len(p.order) >= pendingLimit {
		delete(p.ids, p.order[0])
		p.order = p.order[1:]
	}
	p.ids[id] = struct{}{}
	p.order = append(p.order, id)
}

// take reports whether id waits for a click and forgets it, so a second click
// on the same notification does nothing.
func (p *pending) take(id uint32) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.ids[id]; !ok {
		return false
	}
	delete(p.ids, id)
	for i, kept := range p.order {
		if kept == id {
			p.order = append(p.order[:i], p.order[i+1:]...)
			break
		}
	}
	return true
}

// notify makes the D-Bus call and remembers the id the server returns, so a
// click on that notification reaches the callback.
//
// The listener starts before the call: a click that arrives right after it must
// find a channel already open. A bus that refuses the match rule still shows
// the notification; only the click is lost.
func (d *DBus) notify(title, body string) error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("session bus: %w", err)
	}
	if err := d.listen(conn); err != nil {
		log.Printf("desktop notification: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	call := conn.Object(destination, path).CallWithContext(ctx, method, 0, notifyArgs(title, body)...)
	var id uint32
	if err := call.Store(&id); err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	d.pending.add(id)
	return nil
}

// notifyArgs is the argument list of Notify, in the order of the interface.
//
// It passes no icon: the desktop entry that names one is installed together
// with this app by the packaging step, and an icon name no theme resolves would
// leave an empty image in the notification. It passes a replaces id of 0, so the
// server shows a new notification each time; the Notifier already guarantees
// one notification for each Activity.
//
// The action default is what makes the server report the click. An empty list
// gives the server nothing to report back: that is why the ShowMessage of the
// tray never did.
func notifyArgs(title, body string) []any {
	return []any{
		appName,
		uint32(0), // replaces id: 0 shows a new notification
		"",        // icon: none, see above
		title,
		body,
		[]string{defaultAction, ""}, // actions: the click on the body, no label
		map[string]dbus.Variant{},   // hints
		int32(-1),                   // expire timeout: the server default
	}
}

// listen opens the channel of ActionInvoked once. It never blocks the caller:
// the signals are read on their own goroutine, and a second notification adds
// neither a match rule nor a goroutine.
func (d *DBus) listen(conn *dbus.Conn) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.listening {
		return nil
	}
	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface(destination),
		dbus.WithMatchMember(member),
		dbus.WithMatchObjectPath(path),
	); err != nil {
		return fmt.Errorf("%s: %w", actionInvoked, err)
	}
	d.listening = true
	signals := make(chan *dbus.Signal, 16)
	conn.Signal(signals)
	go func() {
		for signal := range signals {
			d.handle(signal)
		}
	}()
	return nil
}

// handle reports the click of one ActionInvoked signal. It does nothing for
// another signal, for an action other than default, and for an id that is not
// waiting: that id belongs to another application, or to a click already
// reported.
func (d *DBus) handle(signal *dbus.Signal) {
	if signal == nil || signal.Name != actionInvoked || len(signal.Body) != 2 {
		return
	}
	id, ok := signal.Body[0].(uint32)
	if !ok {
		return
	}
	action, ok := signal.Body[1].(string)
	if !ok || action != defaultAction {
		return
	}
	if !d.pending.take(id) {
		return
	}

	d.mu.Lock()
	onClick := d.onClick
	d.mu.Unlock()
	if onClick != nil {
		onClick()
	}
}
