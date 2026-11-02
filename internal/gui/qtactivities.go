package gui

import (
	"strings"
	"sync/atomic"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/activity"
	"dockercontroller/internal/gui/design"
)

const (
	// historyWidth is the width of the column of the history.
	historyWidth = 220
	// activityListName and activityOutputName are what a screen reader calls the
	// two halves of the page.
	activityListName   = "Attività della sessione"
	activityOutputName = "Output dell'Attività"
)

// activityPage is the tab of the Activities: the history on the left and the
// output of the chosen Activity on the right. It reads the log of the session,
// so it shows the history of the session, and it never stops an Activity.
//
// A change of the log arrives on the goroutine of the Manager: the page reads
// it once per burst on the Qt thread, so a thousand lines cost one refresh.
type activityPage struct {
	widget *qt6.QWidget
	list   *qt6.QListWidget
	output *qt6.QPlainTextEdit
	// scroller follows the end of a running output once the layout ran.
	scroller *qt6.QTimer

	log *activity.Log

	// rows is the Activity of each row, by index, and rowOf its index: the two
	// keep the history of the log and the history on screen in the same order.
	rows    []compose.ActivityID
	rowOf   map[compose.ActivityID]int
	shownID map[compose.ActivityID]bool
	// notified remembers the Activities that already raised the notice.
	notified map[compose.ActivityID]bool

	// selected is the Activity of the output; shown, shownLines, shownDrop and
	// shownDone are what the output holds of it, so a new line is appended
	// without rebuilding the text.
	selected   compose.ActivityID
	shown      compose.ActivityID
	shownLines int
	shownDrop  int
	shownDone  bool

	// updating guards the selection the page sets itself: a programmatic
	// SetCurrentRow must not re-enter the render of the output.
	updating bool
	// pending is a refresh already queued on the Qt thread.
	pending atomic.Bool

	// onNew, when set, is called with an Activity the page has not shown yet:
	// the panel opens on it.
	onNew func(compose.ActivityID)
	// onFailure, when set, is called once for each failed Activity: the panel
	// turns it into the notice with the link to the output.
	onFailure func(activity.Activity)
}

// newActivityPage builds the page of the Activities. A nil log builds a page
// that never shows anything, so a shell without the log of the session still
// builds.
func newActivityPage(log *activity.Log) *activityPage {
	p := &activityPage{
		log:      log,
		rowOf:    map[compose.ActivityID]int{},
		shownID:  map[compose.ActivityID]bool{},
		notified: map[compose.ActivityID]bool{},
	}

	p.list = qt6.NewQListWidget2()
	p.list.SetAccessibleName(activityListName)
	p.list.SetSelectionMode(qt6.QAbstractItemView__SingleSelection)
	// Every row is one line of the same height, so the view can size the rows
	// without measuring each one.
	p.list.SetUniformItemSizes(true)
	p.list.SetMinimumWidth(historyWidth)
	p.list.OnCurrentRowChanged(p.rowSelected)

	p.output = qt6.NewQPlainTextEdit2()
	p.output.SetReadOnly(true)
	setObjectName(p.output.QWidget, "activityOutput")
	p.output.SetAccessibleName(activityOutputName)
	p.output.SetPlaceholderText(activity.EmptyText)
	p.output.SetLineWrapMode(qt6.QPlainTextEdit__WidgetWidth)
	// The output of compose is text to read line by line, so it keeps the fixed
	// font of the old view.
	font := qt6.NewQFont2("Monospace")
	font.SetStyleHint(qt6.QFont__Monospace)
	font.SetFixedPitch(true)
	p.output.SetFont(font)

	p.widget = qt6.NewQWidget2()
	layout := qt6.NewQHBoxLayout(p.widget)
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.SetSpacing(design.SpaceS)
	layout.AddWidget(p.list.QWidget)
	layout.AddWidget2(p.output.QWidget, 1)

	p.scroller = qt6.NewQTimer2(p.widget.QObject)
	p.scroller.SetSingleShot(true)
	p.scroller.OnTimeout(p.scrollToEnd)

	if log != nil {
		log.OnChange(p.schedule)
	}

	return p
}

// schedule queues one refresh on the Qt thread. Lines arrive in bursts, so the
// page reads the log once per burst, not once per line.
func (p *activityPage) schedule() {
	if !p.pending.CompareAndSwap(false, true) {
		return
	}
	mainthread.Start(func() {
		p.pending.Store(false)
		p.refresh()
	})
}

// refresh moves the log into the widgets: new Activities as rows, the state of
// the known ones, the notice of a failure, and the output of the selected
// Activity. A new Activity opens the panel on its output, as the old view did.
func (p *activityPage) refresh() {
	if p.log == nil {
		return
	}
	activities := p.log.Activities()

	// The decision comes before the rows go in, so the Activities already shown
	// are the rows on screen.
	opened, open := activity.PanelActivity(p.shownID, activities)

	// The rows of a refresh do not move the view: the scroll of the history is
	// the one the user left, not the one of a history that grew.
	bar := p.list.VerticalScrollBar()
	offset := 0
	if bar != nil {
		offset = bar.Value()
	}

	// A new Activity goes to the front of the history, so the rows are inserted
	// oldest first and each one lands above the one before.
	for i := len(activities) - 1; i >= 0; i-- {
		a := activities[i]
		if !p.shownID[a.ID] {
			p.insertRow(a)
			continue
		}
		p.updateRow(a)
	}

	if bar != nil {
		bar.SetValue(offset)
	}

	for _, a := range activities {
		if a.State == activity.ActivityFailed && !p.notified[a.ID] {
			p.notified[a.ID] = true
			if p.onFailure != nil {
				p.onFailure(a)
			}
		}
	}

	switch {
	case open:
		// An Action started: show its output without being asked. The panel
		// shows the Activity and renders the output itself.
		if p.onNew != nil {
			p.onNew(opened)
			return
		}
		p.Select(opened)
	case p.selected == "" && len(activities) > 0:
		p.Select(activities[0].ID)
	default:
		p.renderOutput()
	}
}

// insertRow adds the row of one Activity at the top of the history.
func (p *activityPage) insertRow(a activity.Activity) {
	item := qt6.NewQListWidgetItem2(activity.ActivityHeader(a))
	item.SetToolTip(activity.Tooltip(a))

	p.list.InsertItem(0, item)
	p.shownID[a.ID] = true
	p.rows = append([]compose.ActivityID{a.ID}, p.rows...)
	p.reindex()
}

// updateRow writes the state of a known Activity on its row. The row itself
// stays, so the history keeps the focus and the scroll.
func (p *activityPage) updateRow(a activity.Activity) {
	row, known := p.rowOf[a.ID]
	if !known {
		return
	}
	item := p.list.Item(row)
	if item == nil {
		return
	}
	item.SetText(activity.ActivityHeader(a))
	item.SetToolTip(activity.Tooltip(a))
}

// rowSelected shows the output of the Activity of the chosen row.
func (p *activityPage) rowSelected(row int) {
	if p.updating || row < 0 || row >= len(p.rows) {
		return
	}
	p.selected = p.rows[row]
	p.renderOutput()
}

// Select shows the output of an Activity of the history.
func (p *activityPage) Select(id compose.ActivityID) {
	p.selected = id
	if row, known := p.rowOf[id]; known {
		// Selecting the row emits the signal, which would set selected again.
		p.updating = true
		p.list.SetCurrentRow(row)
		p.updating = false
	}
	p.renderOutput()
}

// renderOutput puts the output of the selected Activity in the text. Only the
// new lines are appended: the text is rebuilt when the Activity or the lines
// the cap dropped changed.
func (p *activityPage) renderOutput() {
	if p.log == nil {
		return
	}
	a, ok := p.log.Activity(p.selected)
	if !ok {
		p.output.SetPlainText("")
		p.shown = ""
		p.shownLines, p.shownDrop, p.shownDone = 0, 0, false
		return
	}

	if a.ID != p.shown || a.Dropped != p.shownDrop {
		p.output.SetPlainText(activity.ActivityOutput(a))
		p.shown, p.shownDrop = a.ID, a.Dropped
		p.shownLines = len(a.Lines)
		p.shownDone = a.State != activity.ActivityRunning
		p.scroller.Start(0)
		return
	}

	changed := false
	if len(a.Lines) > p.shownLines {
		// AppendPlainText keeps one paragraph per line, so the trailing newline
		// of the block would leave an empty line at the end.
		if lines := strings.TrimRight(activity.LinesText(a.Lines[p.shownLines:]), "\n"); lines != "" {
			p.output.AppendPlainText(lines)
		}
		p.shownLines = len(a.Lines)
		changed = true
	}
	if !p.shownDone && a.State != activity.ActivityRunning {
		p.output.AppendPlainText(strings.TrimRight(activity.ActivityOutcome(a), "\n"))
		p.shownDone = true
		changed = true
	}
	if changed {
		p.scroller.Start(0)
	}
}

// scrollToEnd follows the live output, as the old view did. It runs once the
// layout has sized the text, which is why the caller starts a timer.
func (p *activityPage) scrollToEnd() {
	if bar := p.output.VerticalScrollBar(); bar != nil {
		bar.SetValue(bar.Maximum())
	}
}

// reindex rebuilds the position of each row.
func (p *activityPage) reindex() {
	p.rowOf = make(map[compose.ActivityID]int, len(p.rows))
	for i, id := range p.rows {
		p.rowOf[id] = i
	}
}
