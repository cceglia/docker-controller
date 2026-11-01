package gui

import (
	"strings"
	"sync/atomic"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"

	"dockercontroller/internal/compose"
)

const (
	// historyWidth is the width of the column of the history.
	historyWidth = 220
	// emptyActivityText is the line of the page before the first Activity.
	emptyActivityText = "Nessuna Attività in questa sessione."
)

// activityPage is the page of the Activities of the bottom panel: the history
// on the left and the output of the chosen Activity on the right. It reads the
// ActivityLog, so it shows the history of the session; it never stops an
// Activity.
type activityPage struct {
	widget *gtk.Box
	list   *gtk.ListBox
	// output shows the text of the selected Activity, or the empty notice.
	output   *gtk.Stack
	view     *gtk.TextView
	buffer   *gtk.TextBuffer
	scrolled *gtk.ScrolledWindow
	log      *ActivityLog
	// onFailure, when set, is called once for each failed Activity: the shell
	// turns it into a toast with a link to the output.
	onFailure func(Activity)
	// onNew, when set, is called with an Activity the page has not shown yet:
	// the bottom panel opens on it. It is not called for an output line of an
	// Activity that already has a row.
	onNew func(compose.ActivityID)
	// shownIDs is the Activity of each row. It is what decides whether a change
	// opens the panel.
	shownIDs map[compose.ActivityID]bool
	// notified remembers the Activities that already raised the toast.
	notified map[compose.ActivityID]bool
	// rowByID and rowText give the row and its label of each Activity; rowIDs
	// is the Activity of each row, by index, and changes only when a row is
	// inserted, so a click always reads the list as it is on screen.
	rowByID map[compose.ActivityID]*gtk.ListBoxRow
	rowText map[compose.ActivityID]*gtk.Label
	rowIDs  []compose.ActivityID
	// selected is the Activity of the output; shown, shownLines, shownDrop and
	// shownDone are what the buffer holds of it, so a new line is appended
	// without rebuilding the text.
	selected   compose.ActivityID
	shown      compose.ActivityID
	shownLines int
	shownDrop  int
	shownDone  bool
	pending    atomic.Bool // a refresh is queued on the main loop
}

// newActivityPage builds the page of the Activities.
func newActivityPage(log *ActivityLog) *activityPage {
	p := &activityPage{
		log:      log,
		shownIDs: map[compose.ActivityID]bool{},
		notified: map[compose.ActivityID]bool{},
		rowByID:  map[compose.ActivityID]*gtk.ListBoxRow{},
		rowText:  map[compose.ActivityID]*gtk.Label{},
	}

	p.list = gtk.NewListBox()
	p.list.SetSelectionMode(gtk.SelectionSingle)
	p.list.SetShowSeparators(false)
	p.list.AddCSSClass("navigation-sidebar")
	p.list.ConnectRowSelected(p.rowSelected)

	history := gtk.NewScrolledWindow()
	history.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	history.SetChild(p.list)
	history.SetSizeRequest(historyWidth, -1)
	history.SetVExpand(true)

	p.buffer = gtk.NewTextBuffer(nil)
	p.view = gtk.NewTextView()
	p.view.SetBuffer(p.buffer)
	p.view.SetEditable(false)
	p.view.SetCursorVisible(false)
	p.view.SetMonospace(true)
	p.view.SetWrapMode(gtk.WrapWordChar)
	p.view.SetTopMargin(6)
	p.view.SetBottomMargin(6)
	p.view.SetLeftMargin(8)
	p.view.SetRightMargin(8)

	p.scrolled = gtk.NewScrolledWindow()
	p.scrolled.SetPolicy(gtk.PolicyAutomatic, gtk.PolicyAutomatic)
	p.scrolled.SetChild(p.view)
	p.scrolled.SetHExpand(true)
	p.scrolled.SetVExpand(true)

	empty := gtk.NewLabel(emptyActivityText)
	empty.SetXAlign(0)
	empty.SetYAlign(0)
	empty.SetWrap(true)
	empty.AddCSSClass("dim-label")
	empty.SetMarginTop(8)
	empty.SetMarginStart(8)

	p.output = gtk.NewStack()
	p.output.AddNamed(empty, "empty")
	p.output.AddNamed(p.scrolled, "activity")
	p.output.SetVisibleChildName("empty")

	p.widget = gtk.NewBox(gtk.OrientationHorizontal, 6)
	p.widget.SetMarginTop(6)
	p.widget.SetMarginBottom(6)
	p.widget.SetMarginStart(6)
	p.widget.SetMarginEnd(6)
	p.widget.Append(history)
	p.widget.Append(p.output)
	p.output.SetHExpand(true)

	log.OnChange(p.schedule)

	return p
}

// schedule queues one refresh on the main loop. Lines arrive in bursts, so the
// page reads the log once per burst, not once per line.
func (p *activityPage) schedule() {
	if !p.pending.CompareAndSwap(false, true) {
		return
	}
	glib.IdleAdd(func() {
		p.pending.Store(false)
		p.refresh()
	})
}

// refresh moves the log into the widgets: new Activities as rows, the state of
// the known ones, the failure toast, and the output of the selected Activity. A
// new Activity also opens the panel on its output, as the old drawer did.
func (p *activityPage) refresh() {
	activities := p.log.Activities()

	// The decision comes before the rows go in, so the Activities already shown
	// are the rows on screen.
	newID, open := panelActivity(p.shownIDs, activities)

	// A new Activity goes to the front of the history, so the rows are inserted
	// oldest first and each one lands above the one before.
	for i := len(activities) - 1; i >= 0; i-- {
		if !p.shownIDs[activities[i].ID] {
			p.insertRow(activities[i])
		}
	}

	for _, a := range activities {
		if label, known := p.rowText[a.ID]; known {
			label.SetText(ActivityHeader(a))
			p.rowByID[a.ID].SetTooltipText(activityTooltip(a))
		}
		if a.State == ActivityFailed && !p.notified[a.ID] {
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
			p.onNew(newID)
			return
		}
		p.Select(newID)
	case p.selected == "" && len(activities) > 0:
		p.Select(activities[0].ID)
	default:
		p.renderOutput()
	}
}

// insertRow adds the row of one Activity at the top of the history.
func (p *activityPage) insertRow(a Activity) {
	label := gtk.NewLabel(ActivityHeader(a))
	label.SetXAlign(0)
	label.SetEllipsize(pango.EllipsizeEnd)

	row := gtk.NewListBoxRow()
	row.SetChild(label)
	row.SetTooltipText(activityTooltip(a))

	p.list.Insert(row, 0)
	p.rowByID[a.ID] = row
	p.rowText[a.ID] = label
	p.rowIDs = append([]compose.ActivityID{a.ID}, p.rowIDs...)
	p.shownIDs[a.ID] = true
}

// rowSelected shows the output of the Activity of the chosen row.
func (p *activityPage) rowSelected(row *gtk.ListBoxRow) {
	if row == nil {
		return
	}
	i := row.Index()
	if i < 0 || i >= len(p.rowIDs) {
		return
	}
	p.selected = p.rowIDs[i]
	p.renderOutput()
}

// Select shows the output of an Activity of the history.
func (p *activityPage) Select(id compose.ActivityID) {
	p.selected = id
	if row, known := p.rowByID[id]; known {
		// Selecting a row emits row-selected, which sets selected again.
		p.list.SelectRow(row)
	}
	p.renderOutput()
}

// renderOutput puts the output of the selected Activity in the buffer. Only the
// new lines are appended: the text is rebuilt when the Activity or the lines
// the cap dropped changed.
func (p *activityPage) renderOutput() {
	a, ok := p.log.Activity(p.selected)
	if !ok {
		p.output.SetVisibleChildName("empty")
		p.buffer.SetText("")
		p.shown = ""
		p.shownLines, p.shownDrop, p.shownDone = 0, 0, false
		return
	}
	p.output.SetVisibleChildName("activity")

	if a.ID != p.shown || a.Dropped != p.shownDrop {
		p.buffer.SetText(ActivityOutput(a))
		p.shown, p.shownDrop = a.ID, a.Dropped
		p.shownLines = len(a.Lines)
		p.shownDone = a.State != ActivityRunning
		p.scrollToEnd()
		return
	}

	changed := false
	if len(a.Lines) > p.shownLines {
		p.buffer.Insert(p.buffer.EndIter(), linesText(a.Lines[p.shownLines:]))
		p.shownLines = len(a.Lines)
		changed = true
	}
	if !p.shownDone && a.State != ActivityRunning {
		p.buffer.Insert(p.buffer.EndIter(), strings.TrimRight(ActivityOutcome(a), "\n")+"\n")
		p.shownDone = true
		changed = true
	}
	if changed {
		p.scrollToEnd()
	}
}

// scrollToEnd follows the live output, as the old view did.
func (p *activityPage) scrollToEnd() {
	scrolled := p.scrolled
	glib.IdleAdd(func() {
		adjustment := scrolled.VAdjustment()
		adjustment.SetValue(adjustment.Upper())
	})
}

// activityTooltip is what the row says on hover: the outcome of a finished
// Activity, nothing more than its state while it runs.
func activityTooltip(a Activity) string {
	switch a.State {
	case ActivityOK:
		return ActivitySuccessText(a)
	case ActivityFailed:
		return ActivityToastText(a)
	}
	return "Attività in corso."
}
