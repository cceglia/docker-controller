package gui

import (
	"time"

	"github.com/mappu/miqt/qt6"

	"dockercontroller/internal/gui/design"
)

// noticeAccessibleName is what a screen reader calls the notice before it reads
// the message.
const noticeAccessibleName = "Avviso"

// showOutputText is the button of the notice of a failure: it opens the output
// of the failed Activity.
const showOutputText = "Mostra output"

// toastDuration is how long a message stays on screen: the six seconds of the
// old view, long enough to read it without a click.
const toastDuration = 6 * time.Second

// toastWidthLimit keeps a long message on a few lines instead of one line
// across the window.
const toastWidthLimit = 520

// toastMargin is the space the notice keeps from the edges of its host.
const toastMargin = design.SpaceL

// toast is the notice in sovrimpressione of the shell: one text over the
// content, at the top right of the window. It sits clear of the commands at the
// bottom of the column and of the way back at the top left.
//
// A message alone takes no focus and lets a click through, so it never takes a
// command away. A message with an action carries a button: the button takes the
// focus and the click, because it is the one control of the notice.
//
// The shell owns one toast. A refused Action shows its Italian message here,
// and a failed Activity shows the message with the way to its output.
type toast struct {
	host   *qt6.QWidget
	frame  *qt6.QFrame
	label  *qt6.QLabel
	action *qt6.QPushButton
	timer  *qt6.QTimer
	// run is what the button of the notice does, nil for a message alone.
	run func()
	// top is the extra space the notice keeps from the top of its host, beyond
	// its margin: the bar of the Utilizzo risorse sits there, and a message must
	// not cover a figure. A nil function keeps the plain margin.
	top func() int
}

// newToast builds the notice over host. The notice stays hidden until a message
// arrives.
func newToast(host *qt6.QWidget) *toast {
	t := &toast{host: host}

	t.frame = qt6.NewQFrame(host)
	setObjectName(t.frame.QWidget, "toast")
	t.frame.SetAccessibleName(noticeAccessibleName)
	t.frame.SetFocusPolicy(qt6.NoFocus)
	// A click on a message alone reaches the content under it: the notice must
	// not take a click away from a command.
	t.frame.SetAttribute(qt6.WA_TransparentForMouseEvents)
	t.frame.SetMaximumWidth(toastWidthLimit)
	t.frame.SetVisible(false)

	layout := qt6.NewQHBoxLayout(t.frame.QWidget)
	layout.SetContentsMargins(design.SpaceXS, design.SpaceXS, design.SpaceXS, design.SpaceXS)
	layout.SetSpacing(design.SpaceS)

	t.label = qt6.NewQLabel3("")
	setObjectName(t.label.QWidget, "toastText")
	t.label.SetWordWrap(true)
	layout.AddWidget2(t.label.QWidget, 1)

	t.action = qt6.NewQPushButton3("")
	setObjectName(t.action.QWidget, "toastAction")
	t.action.SetVisible(false)
	// The handler is registered once: a message that arrives while the previous
	// one is on screen must not stack a second handler on the button.
	t.action.OnClicked(func() {
		run := t.run
		t.hide()
		if run != nil {
			run()
		}
	})
	keepTarget(t.action.QWidget)
	layout.AddWidget3(t.action.QWidget, 0, qt6.AlignVCenter)

	t.timer = qt6.NewQTimer2(t.frame.QWidget.QObject)
	t.timer.SetSingleShot(true)
	t.timer.OnTimeout(func() { t.frame.SetVisible(false) })

	return t
}

// show puts message in the notice for the duration, without an action.
func (t *toast) show(message string) {
	t.showAction(message, "", nil)
}

// showAction puts message in the notice for the duration together with a button
// that runs action. An empty label or a nil action shows the message alone. An
// empty message shows nothing.
func (t *toast) showAction(message, label string, action func()) {
	if t.frame == nil || message == "" {
		return
	}

	t.run = nil
	if label != "" && action != nil {
		t.run = action
		t.action.SetText(label)
		t.action.SetAccessibleName(label)
		t.action.SetVisible(true)
		t.frame.SetAttribute2(qt6.WA_TransparentForMouseEvents, false)
	} else {
		t.action.SetVisible(false)
		t.frame.SetAttribute2(qt6.WA_TransparentForMouseEvents, true)
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

// place puts the notice at the top right of its host, below the bar of the
// Utilizzo risorse when the bar is on screen. The placement runs on every
// resize, so the notice stays where the window expects it.
func (t *toast) place() {
	if t.host == nil || t.frame == nil {
		return
	}
	x := t.host.Width() - t.frame.Width() - toastMargin
	if x < toastMargin {
		x = toastMargin
	}
	y := toastMargin
	if t.top != nil {
		y += t.top()
	}
	t.frame.Move(x, y)
}
