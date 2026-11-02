package gui

import (
	"context"

	"github.com/mappu/miqt/qt6"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/activity"
	"dockercontroller/internal/gui/design"
	"dockercontroller/internal/gui/logs"
)

// panelHeight is the height the bottom panel opens at, in logical pixels, as
// the panel of the old view.
const panelHeight = 260

// The positions of the two tabs of the panel.
const (
	activityTab = 0
	logTab      = 1
)

// bottomPanel is the panel at the bottom of the detail: the tabs Attività and
// Log and the button that hides it. It sits in the vertical splitter of the
// shell, under the detail, and takes no room while it is closed: the window
// stays compact until an Activity starts or the logs open.
//
// The panel owns the target of the logs: the Project the shell selected and the
// Service the detail asked for, "" for the whole Project. It stops the stream
// when it closes and when the target changes, so no `docker compose logs`
// process outlives what it shows.
type bottomPanel struct {
	widget *qt6.QWidget
	tabs   *qt6.QTabWidget
	page   *activityPage
	logs   *logPage
	close  *qt6.QPushButton
	// splitter is the vertical splitter the panel shares with the detail: it is
	// what gives the panel its height when it opens.
	splitter *qt6.QSplitter
	// fitTimer sizes the panel once the layout of the splitter has run.
	fitTimer *qt6.QTimer
	// open is true while the panel is on screen.
	open bool
	// suspended is true while the window is off screen: the stream of the logs
	// stays closed until the window comes back.
	suspended bool

	// project is the target of the logs, service is the Service of that
	// Project, "" for the whole Project. hasProject says whether a Project is
	// open at all.
	project    compose.Project
	service    string
	hasProject bool

	// onFailure, when set, is called once for each failed Activity: the shell
	// turns it into the notice with the link to the output.
	onFailure func(activity.Activity)
}

// newBottomPanel builds the panel closed. splitter is the vertical splitter the
// panel shares with the detail. A nil log builds a panel that never opens by
// itself; a nil source builds a Log page that streams nothing.
func newBottomPanel(log *activity.Log, source logs.LogSource, ctx context.Context, theme design.DesignTheme, splitter *qt6.QSplitter) *bottomPanel {
	p := &bottomPanel{splitter: splitter}

	p.page = newActivityPage(log)
	// A new Activity opens the panel on it, as the old drawer did.
	p.page.onNew = p.Show
	p.page.onFailure = func(a activity.Activity) {
		if p.onFailure != nil {
			p.onFailure(a)
		}
	}

	p.logs = newLogPage(theme)
	p.logs.source = source
	p.logs.ctx = ctx

	p.tabs = qt6.NewQTabWidget2()
	p.tabs.SetAccessibleName(activity.PanelAccessibleName)
	// The two tabs always fit, so the bar never scrolls and the two arrows of
	// Qt stay out of the way.
	if bar := p.tabs.TabBar(); bar != nil {
		bar.SetUsesScrollButtons(false)
	}
	// The two tabs carry the content alone: the panel is compact, and a frame
	// around a frame would take a line away from the output.
	p.tabs.SetDocumentMode(true)
	p.tabs.AddTab(p.page.widget, activity.ActivityPageTitle)
	p.tabs.AddTab(p.logs.widget, activity.LogPageTitle)
	// The Log tab shows the logs of the Project on screen when the user chooses
	// it, so a tab that was never opened still streams what it names.
	p.tabs.OnCurrentChanged(func(index int) {
		if p.open && index == logTab {
			p.openLogs()
		}
	})

	p.close = qt6.NewQPushButton3(activity.PanelCloseText)
	p.close.SetAccessibleName(activity.PanelCloseText)
	p.close.SetToolTip("Nascondi il pannello e ferma il flusso dei log. L'Attività in corso continua.")
	p.close.OnClicked(func() { p.Close() })
	keepTarget(p.close.QWidget)
	// The button sits at the right of the tabs, clear of the two of them.
	p.tabs.SetCornerWidget2(p.close.QWidget, qt6.TopRightCorner)

	p.widget = qt6.NewQWidget2()
	layout := qt6.NewQVBoxLayout(p.widget)
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.SetSpacing(0)
	layout.AddWidget(p.tabs.QWidget)

	p.fitTimer = qt6.NewQTimer2(p.widget.QObject)
	p.fitTimer.SetSingleShot(true)
	p.fitTimer.OnTimeout(p.fit)

	p.widget.SetVisible(false)

	return p
}

// Show opens the panel on an Activity of the history: its tab and its output.
func (p *bottomPanel) Show(id compose.ActivityID) {
	p.Open()
	p.tabs.SetCurrentIndex(activityTab)
	p.page.Select(id)
}

// ShowLogs opens the panel on the logs of a Project, whole or of one of its
// Services, and starts the stream.
func (p *bottomPanel) ShowLogs(project compose.Project, service string) {
	p.setTarget(project, service)
	p.Open()
	p.tabs.SetCurrentIndex(logTab)
	// The choice of the tab opens the stream already; the call covers the case
	// of a tab that did not change.
	p.openLogs()
}

// setProject follows the Project the shell has open. A change of Project
// terminates the stream of the old one; the Log tab, when it is on screen,
// starts again on the new Project.
func (p *bottomPanel) setProject(project compose.Project) {
	if p.hasProject && p.project.Name == project.Name {
		p.project = project
		return
	}
	p.setTarget(project, "")
	if p.hasProject && p.open && p.tabs.CurrentIndex() == logTab {
		p.openLogs()
	}
}

// setTarget keeps the target of the logs and stops the stream of the old one.
// It never starts a stream: the panel opens one only when the Log tab is on
// screen, so a closed panel holds no `docker compose logs` process.
func (p *bottomPanel) setTarget(project compose.Project, service string) {
	changed := !p.hasProject || p.project.Name != project.Name
	p.project, p.hasProject = project, project.Name != ""
	if p.hasProject {
		p.service = service
	} else {
		p.service = ""
	}
	if changed {
		p.logs.Clear()
	}
}

// openLogs starts the logs of the target on screen, if any. A hidden window
// starts nothing: the stream waits for the window to come back.
func (p *bottomPanel) openLogs() {
	if p.suspended || !p.hasProject {
		return
	}
	p.logs.Open(p.project, p.service)
}

// Open shows the panel at its height. The Activity on screen keeps running.
func (p *bottomPanel) Open() {
	if p.open {
		return
	}
	p.open = true
	p.widget.SetVisible(true)
	// The height of the splitter arrives with the layout, so the size waits for
	// the next turn of the event loop.
	p.fitTimer.Start(0)
	// The Log tab may be the one on screen: a panel that opens on it starts the
	// stream of the Project the shell has open.
	if p.tabs.CurrentIndex() == logTab {
		p.openLogs()
	}
}

// Close hides the panel and terminates the stream of the logs. It keeps the
// history of the Activities.
func (p *bottomPanel) Close() {
	p.open = false
	p.fitTimer.Stop()
	p.widget.SetVisible(false)
	p.logs.Stop()
}

// showing reports whether the panel is on screen.
func (p *bottomPanel) showing() bool {
	return p.open
}

// Suspend stops the stream of the logs while the window is off screen and keeps
// the target and the panel, so a window that comes back shows the same logs. The
// history of the Activities stays: it lives in memory and runs no process.
func (p *bottomPanel) Suspend() {
	if p.suspended {
		return
	}
	p.suspended = true
	p.logs.Stop()
}

// Resume starts the logs again when the window comes back. A closed panel, a tab
// other than the Log one and a panel with no Project start nothing.
func (p *bottomPanel) Resume() {
	if !p.suspended {
		return
	}
	p.suspended = false
	if p.open && p.tabs.CurrentIndex() == logTab {
		p.openLogs()
	}
}

// fit gives the panel its height in the vertical splitter and leaves the rest
// to the detail. A short window gives the panel at most half of it, so the
// detail never loses the room it needs.
func (p *bottomPanel) fit() {
	if p.splitter == nil || !p.open {
		return
	}
	height := p.splitter.Height()
	if height <= 0 {
		return
	}
	size := min(panelHeight, height/2)
	p.splitter.SetSizes([]int{height - size, size})
}

// linkAfter puts the controls of the panel after a control of the shell, so the
// tab chain reads the panel after the detail: the button that closes it, the
// history, the output, then the bar and the text of the logs.
func (p *bottomPanel) linkAfter(prev *qt6.QWidget) {
	if prev == nil {
		return
	}
	qt6.QWidget_SetTabOrder(prev, p.close.QWidget)
	qt6.QWidget_SetTabOrder(p.close.QWidget, p.page.list.QWidget)
	qt6.QWidget_SetTabOrder(p.page.list.QWidget, p.page.output.QWidget)
	qt6.QWidget_SetTabOrder(p.page.output.QWidget, p.logs.search.QWidget)
	qt6.QWidget_SetTabOrder(p.logs.search.QWidget, p.logs.clocks.QWidget)
	qt6.QWidget_SetTabOrder(p.logs.clocks.QWidget, p.logs.pause.QWidget)
	qt6.QWidget_SetTabOrder(p.logs.pause.QWidget, p.logs.view.QWidget)
}
