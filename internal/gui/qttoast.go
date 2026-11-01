package gui

import (
	"time"

	"github.com/mappu/miqt/qt6"
)

// noticeAccessibleName is what a screen reader calls the notice before it reads
// the message.
const noticeAccessibleName = "Avviso"

// toastDuration is how long a message stays on screen: the six seconds of the
// old view, long enough to read it without a click.
const toastDuration = 6 * time.Second

// toastWidthLimit keeps a long message on a few lines instead of one line
// across the window.
const toastWidthLimit = 520

// toastMargin is the space the notice keeps from the edges of its host.
const toastMargin = spaceL

// toast is the notice in sovrimpressione of the shell: one text over the
// content, at the top right of the window. It sits clear of the commands at the
// bottom of the column and of the way back at the top left. It takes no focus,
// and a click passes through it, so it never takes a command away.
//
// The shell owns one toast. A refused Action shows its Italian message here,
// and the failed Activity of the next ticket uses the same notice.
type toast struct {
	host  *qt6.QWidget
	frame *qt6.QFrame
	label *qt6.QLabel
	timer *qt6.QTimer
}

// newToast builds the notice over host. The notice stays hidden until a message
// arrives.
func newToast(host *qt6.QWidget) *toast {
	t := &toast{host: host}

	t.frame = qt6.NewQFrame(host)
	setObjectName(t.frame.QWidget, "toast")
	t.frame.SetAccessibleName(noticeAccessibleName)
	t.frame.SetFocusPolicy(qt6.NoFocus)
	// A click on the notice reaches the content under it: the notice must not
	// take a click away from a command.
	t.frame.SetAttribute(qt6.WA_TransparentForMouseEvents)
	t.frame.SetMaximumWidth(toastWidthLimit)
	t.frame.SetVisible(false)

	layout := qt6.NewQHBoxLayout(t.frame.QWidget)
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.SetSpacing(0)

	t.label = qt6.NewQLabel3("")
	setObjectName(t.label.QWidget, "toastText")
	t.label.SetWordWrap(true)
	layout.AddWidget(t.label.QWidget)

	t.timer = qt6.NewQTimer2(t.frame.QWidget.QObject)
	t.timer.SetSingleShot(true)
	t.timer.OnTimeout(func() { t.frame.SetVisible(false) })

	return t
}

// show puts message in the notice for the duration and moves it out of the way
// of the commands. An empty message shows nothing.
func (t *toast) show(message string) {
	if t.frame == nil || message == "" {
		return
	}
	t.label.SetText(message)
	t.frame.SetAccessibleDescription(message)
	t.frame.AdjustSize()
	t.place()
	t.frame.Raise()
	t.frame.SetVisible(true)
	t.timer.Start(int(toastDuration / time.Millisecond))
}

// hide takes the notice off the screen and stops its timer.
func (t *toast) hide() {
	if t.frame == nil {
		return
	}
	t.timer.Stop()
	t.frame.SetVisible(false)
}

// place puts the notice at the top right of its host. The placement runs on
// every resize, so the notice stays where the window expects it.
func (t *toast) place() {
	if t.host == nil || t.frame == nil {
		return
	}
	x := t.host.Width() - t.frame.Width() - toastMargin
	if x < toastMargin {
		x = toastMargin
	}
	t.frame.Move(x, toastMargin)
}
