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
	// panelHeight is the height the bottom panel opens at, in pixels.
	panelHeight = 260
	// historyWidth is the width of the column of the history.
	historyWidth = 220
	// emptyActivityText is the line of the panel before the first Activity.
	emptyActivityText = "Nessuna Attività in questa sessione."
	// panelCloseText is the button that hides the panel.
	panelCloseText = "Chiudi"
	// panelTabText is the title of the one tab this ticket adds.
	panelTabText = "Attività"
)

// activityPanel is the bottom panel of the detail: a stack with its switcher
// and, for now, the one page of the Activities. It reads the ActivityLog, so it
// shows the history of the session; it never stops an Activity.
type activityPanel struct {
	// body is the end child of the paned, paned the widget it sits in.
	body  *gtk.Box
	paned *gtk.Paned
	list  *gtk.ListBox
	// output shows the text of the selected Activity, or the empty notice.
	output   *gtk.Stack
	view     *gtk.TextView
	buffer   *gtk.TextBuffer
	scrolled *gtk.ScrolledWindow
	log      *ActivityLog
	// onFailure, when set, is called once for each failed Activity: the shell
	// turns it into a toast with a link to the output.
	onFailure func(Activity)
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

// newActivityPanel builds the panel. It stays hidden until an Activity starts:
// the first one opens it.
func newActivityPanel(log *ActivityLog) *activityPanel {
	p := &activityPanel{
		log:      log,
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

	page := gtk.NewBox(gtk.OrientationHorizontal, 6)
	page.SetMarginTop(6)
	page.SetMarginBottom(6)
	page.SetMarginStart(6)
	page.SetMarginEnd(6)
	page.Append(history)
	page.Append(p.output)
	p.output.SetHExpand(true)

	stack := gtk.NewStack()
	stack.SetTransitionType(gtk.StackTransitionTypeCrossfade)
	stack.AddTitled(page, "attivita", panelTabText)
	stack.SetVExpand(true)

	switcher := gtk.NewStackSwitcher()
	switcher.SetStack(stack)
	switcher.SetHExpand(true)
	switcher.SetHAlign(gtk.AlignStart)

	close := gtk.NewButtonWithLabel(panelCloseText)
	close.AddCSSClass("flat")
	close.SetTooltipText("Nascondi il pannello. L'Attività in corso continua.")
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
	p.body.Append(stack)
	p.body.SetVisible(false)

	log.OnChange(p.schedule)

	return p
}

// schedule queues one refresh on the main loop. Lines arrive in bursts, so the
// panel reads the log once per burst, not once per line.
func (p *activityPanel) schedule() {
	if !p.pending.CompareAndSwap(false, true) {
		return
	}
	glib.IdleAdd(func() {
		p.pending.Store(false)
		p.refresh()
	})
}

// refresh moves the log into the widgets: new Activities as rows, the state of
// the known ones, the failure toast, and the output of the selected Activity.
func (p *activityPanel) refresh() {
	activities := p.log.Activities()

	added := 0
	for _, a := range activities {
		if _, known := p.rowByID[a.ID]; !known {
			added++
		}
	}
	// A new Activity goes to the front of the history, so the rows are inserted
	// oldest first and each one lands above the one before.
	for i := len(activities) - 1; i >= 0; i-- {
		if _, known := p.rowByID[activities[i].ID]; !known {
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
	case added > 0:
		// An Action started: show its output without being asked.
		p.Show(activities[0].ID)
	case p.selected == "" && len(activities) > 0:
		p.Show(activities[0].ID)
	default:
		p.renderOutput()
	}
}

// insertRow adds the row of one Activity at the top of the history.
func (p *activityPanel) insertRow(a Activity) {
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
}

// rowSelected shows the output of the Activity of the chosen row.
func (p *activityPanel) rowSelected(row *gtk.ListBoxRow) {
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

// Show selects an Activity of the history and opens the panel on it.
func (p *activityPanel) Show(id compose.ActivityID) {
	p.Open()
	p.selected = id
	if row, known := p.rowByID[id]; known {
		// Selecting a row emits row-selected, which sets selected again.
		p.list.SelectRow(row)
	}
	p.renderOutput()
}

// Open shows the panel at its height. The Activity on screen keeps running.
func (p *activityPanel) Open() {
	if p.body.Visible() {
		return
	}
	p.body.SetVisible(true)
	p.fit()
}

// Close hides the panel. It stops no Activity and keeps the history.
func (p *activityPanel) Close() {
	p.body.SetVisible(false)
}

// fit puts the handle of the paned panelHeight above the bottom.
func (p *activityPanel) fit() {
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

// renderOutput puts the output of the selected Activity in the buffer. Only the
// new lines are appended: the text is rebuilt when the Activity or the lines
// the cap dropped changed.
func (p *activityPanel) renderOutput() {
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
func (p *activityPanel) scrollToEnd() {
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
