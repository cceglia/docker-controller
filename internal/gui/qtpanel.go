package gui

import (
	"github.com/mappu/miqt/qt6"

	"dockercontroller/internal/compose"
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
// stays compact until an Activity starts.
type bottomPanel struct {
	widget *qt6.QWidget
	tabs   *qt6.QTabWidget
	page   *activityPage
	close  *qt6.QPushButton
	// splitter is the vertical splitter the panel shares with the detail: it is
	// what gives the panel its height when it opens.
	splitter *qt6.QSplitter
	// fitTimer sizes the panel once the layout of the splitter has run.
	fitTimer *qt6.QTimer
	// open is true while the panel is on screen.
	open bool

	// onFailure, when set, is called once for each failed Activity: the shell
	// turns it into the notice with the link to the output.
	onFailure func(Activity)
}

// newBottomPanel builds the panel closed. splitter is the vertical splitter the
// panel shares with the detail. A nil log builds a panel that never opens by
// itself.
func newBottomPanel(log *ActivityLog, splitter *qt6.QSplitter) *bottomPanel {
	p := &bottomPanel{splitter: splitter}

	p.page = newActivityPage(log)
	// A new Activity opens the panel on it, as the old drawer did.
	p.page.onNew = p.Show
	p.page.onFailure = func(a Activity) {
		if p.onFailure != nil {
			p.onFailure(a)
		}
	}

	p.tabs = qt6.NewQTabWidget2()
	p.tabs.SetAccessibleName(panelAccessibleName)
	// The two tabs always fit, so the bar never scrolls and the two arrows of
	// Qt stay out of the way.
	if bar := p.tabs.TabBar(); bar != nil {
		bar.SetUsesScrollButtons(false)
	}
	// The two tabs carry the content alone: the panel is compact, and a frame
	// around a frame would take a line away from the output.
	p.tabs.SetDocumentMode(true)
	p.tabs.AddTab(p.page.widget, activityPageTitle)
	p.tabs.AddTab(newLogPage(), logPageTitle)

	p.close = qt6.NewQPushButton3(panelCloseText)
	p.close.SetAccessibleName(panelCloseText)
	p.close.SetToolTip("Nascondi il pannello. L'Attività in corso continua.")
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

// newLogPage is the tab of the Log until the ticket that streams it: an empty
// page with the line that says so.
func newLogPage() *qt6.QWidget {
	page := qt6.NewQWidget2()
	layout := qt6.NewQVBoxLayout(page)
	layout.SetContentsMargins(spaceM, spaceM, spaceM, spaceM)
	layout.SetSpacing(spaceS)

	label := qt6.NewQLabel3(emptyLogText)
	setObjectName(label.QWidget, "logEmpty")
	label.SetWordWrap(true)
	layout.AddWidget(label.QWidget)
	layout.AddStretch()

	return page
}

// Show opens the panel on an Activity of the history: its tab and its output.
func (p *bottomPanel) Show(id compose.ActivityID) {
	p.Open()
	p.tabs.SetCurrentIndex(activityTab)
	p.page.Select(id)
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
}

// Close hides the panel. It stops nothing and keeps the history.
func (p *bottomPanel) Close() {
	p.open = false
	p.fitTimer.Stop()
	p.widget.SetVisible(false)
}

// showing reports whether the panel is on screen.
func (p *bottomPanel) showing() bool {
	return p.open
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
// history, then the output.
func (p *bottomPanel) linkAfter(prev *qt6.QWidget) {
	if prev == nil {
		return
	}
	qt6.QWidget_SetTabOrder(prev, p.close.QWidget)
	qt6.QWidget_SetTabOrder(p.close.QWidget, p.page.list.QWidget)
	qt6.QWidget_SetTabOrder(p.page.list.QWidget, p.page.output.QWidget)
}
