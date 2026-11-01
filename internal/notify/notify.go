// Package notify shows desktop notifications through the notification service
// of the desktop. The transport is org.freedesktop.Notifications on the session
// bus, the service every desktop understands; nothing here depends on GTK, so
// the app keeps working when the service is missing.
package notify

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/godbus/dbus/v5"
)

// Sender shows a desktop notification. Send must return at once: the transport
// does the work in the background, because the D-Bus call blocks and would
// otherwise stop the GTK loop and the Activity that produced the notification.
type Sender interface {
	Send(title, body string)
}

const (
	// appName is the name the notification server shows as the source.
	appName = "Docker Controller"
	// destination, path and method are the service of the desktop.
	destination = "org.freedesktop.Notifications"
	path        = dbus.ObjectPath("/org/freedesktop/Notifications")
	method      = "org.freedesktop.Notifications.Notify"
	// callTimeout bounds one call, so a service that does not answer leaves no
	// goroutine behind.
	callTimeout = 5 * time.Second
)

// DBus is the Sender of the session bus.
type DBus struct{}

// NewDBus returns a Sender on org.freedesktop.Notifications of the session bus.
func NewDBus() *DBus { return &DBus{} }

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

// notify makes the D-Bus call.
//
// It passes no icon: the desktop entry that names one is installed together
// with this app by the packaging step, and an icon name no theme resolves would
// leave an empty image in the notification. It passes a replaces id of 0, so the
// server shows a new notification each time; the Notifier already guarantees
// one notification for each Activity.
func (d *DBus) notify(title, body string) error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("session bus: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	call := conn.Object(destination, path).CallWithContext(ctx, method, 0,
		appName,
		uint32(0), // replaces id: 0 shows a new notification
		"",        // icon: none, see above
		title,
		body,
		[]string{},                // actions
		map[string]dbus.Variant{}, // hints
		int32(-1),                 // expire timeout: the server default
	)
	var id uint32
	if err := call.Store(&id); err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	return nil
}
