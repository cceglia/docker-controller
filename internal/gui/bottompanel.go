package gui

import (
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"dockercontroller/internal/compose"
)

const (
	// panelHeight is the height the bottom panel opens at, in pixels.
	panelHeight = 260
	// panelCloseText is the button that hides the panel.
	panelCloseText = "Chiudi"
	// activityPageTitle is the title of the page of the Activities.
	activityPageTitle = "Attività"
	// logPageTitle is the title of the page of the logs.
	logPageTitle = "Log"
)

// bottomPanel is the bottom panel of the detail: a stack with its switcher and
// the two pages, the Activities and the Log. The Log page exists only while a
// stream is open, like the old drawer: closing it terminates the process of
// `docker compose logs`.
type bottomPanel struct {
	// body is the end child of the paned, paned the widget it sits in.
	body  *gtk.Box
	paned *gtk.Paned
	stack *gtk.Stack
	// activities and logs are the two pages.
	activities *activityPage
	logs       *logPanel
	// onFailure, when set, is called once for each failed Activity: the shell
	// turns it into a toast with a link to the output.
	onFailure func(Activity)
}

// newBottomPanel builds the panel. It stays hidden until an Activity starts or
// the user asks for the logs: the first Activity opens it.
func newBottomPanel(log *ActivityLog) *bottomPanel {
	p := &bottomPanel{}

	p.activities = newActivityPage(log)
	p.activities.onFailure = func(a Activity) {
		if p.onFailure != nil {
			p.onFailure(a)
		}
	}
	// A new Activity opens the panel on its output, as the old drawer did.
	p.activities.onNew = p.Show

	p.logs = newLogPanel()
	p.logs.onClose = p.closeLogs

	p.stack = gtk.NewStack()
	p.stack.SetTransitionType(gtk.StackTransitionTypeCrossfade)
	p.stack.SetVExpand(true)
	p.stack.AddTitled(p.activities.widget, "attivita", activityPageTitle)
	p.logs.page = p.stack.AddTitled(p.logs.widget, "log", logPageTitle)
	p.logs.page.SetVisible(false)

	switcher := gtk.NewStackSwitcher()
	switcher.SetStack(p.stack)
	switcher.SetHExpand(true)
	switcher.SetHAlign(gtk.AlignStart)

	close := gtk.NewButtonWithLabel(panelCloseText)
	close.AddCSSClass("flat")
	close.SetTooltipText("Nascondi il pannello. L'Attività e i log in corso continuano.")
	close.ConnectClicked(p.Close)

	head := gtk.NewBox(gtk.OrientationHorizontal, 6)
	head.SetMarginTop(4)
	head.SetMarginBottom(2)
	head.SetMarginStart(8)
	head.SetMarginEnd(8)
	head.Append(switcher)
	head.Append(close)

	p.body = gtk.NewBox(gtk.OrientationVertical, 0)
	p.body.Append(head)
	p.body.Append(p.stack)
	p.body.SetVisible(false)

	return p
}

// panelActivity returns the Activity the bottom panel must open on after a
// change: the newest Activity of the history it has not shown yet. It reports
// false when every Activity of the history is already shown, so an output line
// never reopens the panel. It reads no widget, so a test drives it.
func panelActivity(shown map[compose.ActivityID]bool, activities []Activity) (compose.ActivityID, bool) {
	for _, a := range activities {
		if !shown[a.ID] {
			return a.ID, true
		}
	}
	return "", false
}

// Show selects an Activity of the history and opens the panel on it.
func (p *bottomPanel) Show(id compose.ActivityID) {
	p.Open()
	p.stack.SetVisibleChildName("attivita")
	p.activities.Select(id)
}

// ShowLogs opens the Log page on the logs of a Project, whole or of one of its
// Services, and starts the stream.
func (p *bottomPanel) ShowLogs(project compose.Project, service string) {
	p.Open()
	p.logs.page.SetVisible(true)
	p.stack.SetVisibleChildName("log")
	p.logs.Open(project, service)
}

// closeLogs terminates the stream, hides the Log page and goes back to the
// Activities. It is the button of the page and the click of its tab's close.
func (p *bottomPanel) closeLogs() {
	p.logs.Stop()
	p.stack.SetVisibleChildName("attivita")
	p.logs.page.SetVisible(false)
}

// Open shows the panel at its height. The Activity and the logs on screen keep
// running.
func (p *bottomPanel) Open() {
	if p.body.Visible() {
		return
	}
	p.body.SetVisible(true)
	p.fit()
}

// Close hides the panel. It stops nothing and keeps the history.
func (p *bottomPanel) Close() {
	p.body.SetVisible(false)
}

// fit puts the handle of the paned panelHeight above the bottom.
func (p *bottomPanel) fit() {
	paned := p.paned
	if paned == nil {
		return
	}
	// The new child is laid out after this call, so the position is set when
	// the main loop is idle.
	glib.IdleAdd(func() {
		position := paned.Height() - panelHeight
		if position < 0 {
			position = 0
		}
		paned.SetPosition(position)
	})
}
