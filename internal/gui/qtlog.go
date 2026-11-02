package gui

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/activity"
	"dockercontroller/internal/gui/design"
	"dockercontroller/internal/gui/logs"
)

// stderrPrefix marks a line of compose itself in the text. The colour of the
// line says the same, but a prefix survives a black and white screen, a colour
// blindness and a copy of the text.
const stderrPrefix = "! "

// logFormats holds the formats of the parts of a line, built once from the
// tokens of the theme on screen: one colour per Service, the clock, the output
// of compose itself and the plain text.
type logFormats struct {
	plain     *qt6.QTextCharFormat
	timestamp *qt6.QTextCharFormat
	stderr    *qt6.QTextCharFormat
	service   [design.ServiceHues]*qt6.QTextCharFormat
}

// newLogFormats builds the formats of one theme. The colours are the roles of
// the tokens, so the token test reads the same values.
func newLogFormats(theme design.DesignTheme) *logFormats {
	f := &logFormats{plain: qt6.NewQTextCharFormat()}

	f.timestamp = qt6.NewQTextCharFormat()
	f.timestamp.SetForeground(qt6.NewQBrush3(qt6.NewQColor6(theme.Color(design.RoleLogTimestamp))))

	// A message of compose itself is not a Service: it carries no colour of the
	// palette, and its slant, its weight and its prefix say it too.
	f.stderr = qt6.NewQTextCharFormat()
	f.stderr.SetForeground(qt6.NewQBrush3(qt6.NewQColor6(theme.Color(design.RoleStatusFault))))
	f.stderr.SetFontWeight(int(qt6.QFont__Bold))
	f.stderr.SetFontItalic(true)

	for hue := 0; hue < design.ServiceHues; hue++ {
		format := qt6.NewQTextCharFormat()
		format.SetForeground(qt6.NewQBrush3(qt6.NewQColor6(design.ServiceColor(hue, theme))))
		format.SetFontWeight(int(qt6.QFont__Bold))
		f.service[hue] = format
	}
	return f
}

// format returns the format of a tag of logs.LogSegments. A tag the page does not
// know, and the empty tag of the text of a Container, keep the plain format.
func (f *logFormats) format(tag string) *qt6.QTextCharFormat {
	switch tag {
	case logs.LogTimestampTag:
		return f.timestamp
	case logs.LogStderrTag:
		return f.stderr
	}
	if hue, ok := logs.ServiceHueOf(tag); ok {
		return f.service[hue]
	}
	return f.plain
}

// logPage is the page of the logs of the bottom panel: the logs of a Project,
// or of one of its Services. It owns the stream: opening it starts
// `docker compose logs`, closing the panel terminates the process.
//
// Lines arrive on a goroutine and wait in pending; the Qt thread moves them
// into the history and the view in bursts, so a fast stream does not rewrite
// the text once per line.
type logPage struct {
	widget      *qt6.QWidget
	search      *qt6.QLineEdit
	searchLabel *qt6.QLabel
	pause       *qt6.QPushButton
	clocks      *qt6.QCheckBox
	target      *qt6.QLabel
	count       *qt6.QLabel
	view        *qt6.QPlainTextEdit
	status      *qt6.QLabel
	scroller    *qt6.QTimer

	theme   design.DesignTheme
	formats *logFormats

	// source opens the stream: it is compose.Manager.Logs in the application.
	source logs.LogSource
	// ctx stops every stream when the application quits.
	ctx context.Context
	// streams keeps the running stream, so a close that overtakes an opening
	// still terminates it.
	streams logs.LogStreams

	// project and service are what the page streams now; id is the stream in
	// the registry, "" when none runs. Both belong to the Qt thread.
	project string
	service string
	id      string

	// lines are the lines the page keeps, at most logs.LogLineLimit; shown is how
	// many of them the view already holds. Both belong to the Qt thread.
	lines []compose.LogLine
	shown int
	// term is the search, and paused stops the automatic scrolling only, never
	// the stream.
	term   string
	paused bool
	// end is how the stream ended on its own, nil while it runs.
	end *logs.LogEnd

	// mu guards generation and pending, which the goroutine of the stream
	// touches too.
	mu         sync.Mutex
	generation int
	pending    []compose.LogLine
	scheduled  atomic.Bool
}

// newLogPage builds the page with no stream.
func newLogPage(theme design.DesignTheme) *logPage {
	p := &logPage{theme: theme, formats: newLogFormats(theme)}

	p.target = qt6.NewQLabel2()
	setObjectName(p.target.QWidget, "logTarget")
	p.target.SetAccessibleName("Log di")

	p.search = qt6.NewQLineEdit2()
	setObjectName(p.search.QWidget, "logSearch")
	p.search.SetAccessibleName("Ricerca nei log")
	p.search.SetClearButtonEnabled(true)
	keepTarget(p.search.QWidget)
	p.search.OnTextChanged(p.searchChanged)

	// The visible label names the field, and the field carries the same words
	// for a screen reader: a hint inside the field would be text too, and the
	// placeholder role of the desktop palette is too faint on the dark surface.
	p.searchLabel = qt6.NewQLabel3(activity.LogSearchLabelText)
	p.searchLabel.SetBuddy(p.search.QWidget)

	p.clocks = qt6.NewQCheckBox3(activity.LogTimestampsText)
	setObjectName(p.clocks.QWidget, "logClocks")
	p.clocks.SetAccessibleName("Mostra l'ora di ogni riga")
	p.clocks.SetChecked(true)
	keepTarget(p.clocks.QWidget)
	p.clocks.OnToggled(func(bool) {
		p.rebuild()
		p.updateState()
		p.follow()
	})

	p.pause = qt6.NewQPushButton3(activity.LogPauseText)
	p.pause.SetCheckable(true)
	p.pause.SetAccessibleName("Pausa dello scorrimento")
	p.pause.SetToolTip("Ferma lo scorrimento automatico. Le righe nuove continuano ad arrivare.")
	keepTarget(p.pause.QWidget)
	p.pause.OnToggled(p.togglePause)

	p.count = qt6.NewQLabel2()
	setObjectName(p.count.QWidget, "logCount")
	p.count.SetAccessibleName("Righe del log")

	// The bar stays on one line over the text: the search takes the room the
	// other controls leave, so no column stays empty.
	bar := qt6.NewQHBoxLayout2()
	bar.SetContentsMargins(design.SpaceM, design.SpaceXS, design.SpaceM, design.SpaceXS)
	bar.SetSpacing(design.SpaceS)
	bar.AddWidget3(p.target.QWidget, 0, qt6.AlignVCenter)
	bar.AddWidget3(p.searchLabel.QWidget, 0, qt6.AlignVCenter)
	bar.AddWidget2(p.search.QWidget, 1)
	bar.AddWidget3(p.clocks.QWidget, 0, qt6.AlignVCenter)
	bar.AddWidget3(p.pause.QWidget, 0, qt6.AlignVCenter)
	bar.AddWidget3(p.count.QWidget, 0, qt6.AlignVCenter)

	p.view = qt6.NewQPlainTextEdit2()
	p.view.SetReadOnly(true)
	setObjectName(p.view.QWidget, "logView")
	p.view.SetAccessibleName("Log del Progetto")
	// The view drops the oldest block on its own when a burst arrives faster
	// than the page trims the history, so the text never grows past the cap.
	// One block more than the lines: the last line opens an empty block after
	// its newline.
	p.view.SetMaximumBlockCount(logs.LogLineLimit + 1)
	p.view.SetLineWrapMode(qt6.QPlainTextEdit__WidgetWidth)
	font := qt6.NewQFont2("Monospace")
	font.SetStyleHint(qt6.QFont__Monospace)
	font.SetFixedPitch(true)
	p.view.SetFont(font)

	p.status = qt6.NewQLabel2()
	setObjectName(p.status.QWidget, "logStatus")
	p.status.SetWordWrap(true)
	p.status.SetAccessibleName("Esito del flusso dei log")
	p.status.SetVisible(false)

	p.widget = qt6.NewQWidget2()
	layout := qt6.NewQVBoxLayout(p.widget)
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.SetSpacing(0)
	layout.AddLayout2(bar.QLayout, 0)
	layout.AddWidget2(p.view.QWidget, 1)
	layout.AddWidget(p.status.QWidget)

	p.scroller = qt6.NewQTimer2(p.widget.QObject)
	p.scroller.SetSingleShot(true)
	p.scroller.OnTimeout(p.scrollToEnd)

	return p
}

// Open shows the logs of the whole Project, or of one of its Services, and
// starts the stream: the last 500 lines at once, then the new ones. The same
// stream already open is shown again, not restarted.
func (p *logPage) Open(project compose.Project, service string) {
	p.target.SetText(logs.LogTargetText(project, service))
	if p.source == nil || project.Name == "" {
		p.project, p.service = "", ""
		p.updateState()
		return
	}
	if p.id != "" && p.project == project.Name && p.service == service {
		return
	}
	p.resetView()
	p.project = project.Name
	p.service = service
	p.start()
}

// Stop terminates the running stream. The lines stay on screen.
func (p *logPage) Stop() {
	p.stopStream()
	p.end = nil
	p.updateState()
}

// Clear stops the stream and forgets the lines: the page shows the logs of no
// Project.
func (p *logPage) Clear() {
	p.Stop()
	p.project, p.service = "", ""
	p.lines = nil
	p.shown = 0
	p.view.SetPlainText("")
	p.target.SetText("")
	p.resetView()
	p.updateState()
}

// Running reports whether a stream is open now.
func (p *logPage) Running() bool {
	return p.id != ""
}

// Target is the Project and the Service the page streams, "" for none and for
// the whole Project.
func (p *logPage) Target() (project, service string) {
	return p.project, p.service
}

// resetView puts the controls back to their first value: a page that is opened
// again starts without a search, with the clocks and without a pause.
func (p *logPage) resetView() {
	p.search.SetText("")
	p.term = ""
	p.clocks.SetChecked(true)
	p.paused = false
	p.pause.SetChecked(false)
	p.pause.SetText(activity.LogPauseText)
}

// start terminates the stream on screen, if any, and opens the new one.
func (p *logPage) start() {
	if p.source == nil {
		return
	}
	p.stopStream()
	p.lines = nil
	p.shown = 0
	p.end = nil
	p.view.SetPlainText("")
	p.updateState()

	ctx := p.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	streamCtx, cancel := context.WithCancel(ctx)

	p.mu.Lock()
	p.generation++
	generation := p.generation
	p.pending = nil
	p.mu.Unlock()

	id := "log-" + strconv.Itoa(generation)
	if err := p.streams.Add(id, cancel); err != nil {
		// The page was closed before the stream registered: do not start.
		cancel()
		return
	}
	p.id = id

	source := p.source
	req := compose.LogRequest{Project: p.project, Service: p.service}
	go func() {
		end := logs.StreamLogs(streamCtx, req, source,
			func(line compose.LogLine) { p.push(generation, line) })
		p.streams.Release(id)
		if end.Stopped {
			return
		}
		p.finish(generation, end)
	}()
}

// stopStream terminates the running stream and forgets the lines still in
// flight: a line of a stopped stream must not reach the view.
func (p *logPage) stopStream() {
	p.mu.Lock()
	p.generation++
	p.pending = nil
	p.mu.Unlock()
	if p.id != "" {
		p.streams.Stop(p.id)
		p.id = ""
	}
}

// push queues one line of the stream for the Qt thread.
func (p *logPage) push(generation int, line compose.LogLine) {
	p.mu.Lock()
	if generation != p.generation {
		p.mu.Unlock()
		return
	}
	p.pending = append(p.pending, line)
	p.mu.Unlock()
	p.schedule()
}

// finish records how the stream ended, on the Qt thread.
func (p *logPage) finish(generation int, end logs.LogEnd) {
	mainthread.Start(func() {
		p.mu.Lock()
		current := generation == p.generation
		p.mu.Unlock()
		if !current {
			return
		}
		p.end = &end
		p.id = ""
		p.updateState()
	})
}

// schedule queues one flush on the Qt thread. Lines arrive in bursts, so the
// page writes them together.
func (p *logPage) schedule() {
	if !p.scheduled.CompareAndSwap(false, true) {
		return
	}
	mainthread.Start(func() {
		p.scheduled.Store(false)
		p.flush()
	})
}

// flush moves the waiting lines into the history and the view. When the cap
// dropped lines, or the filter is active, the view is written again from the
// history; otherwise only the new lines are appended.
func (p *logPage) flush() {
	p.mu.Lock()
	incoming := p.pending
	p.pending = nil
	p.mu.Unlock()
	if len(incoming) == 0 {
		return
	}

	before := len(p.lines)
	p.lines = logs.AppendLogLines(p.lines, incoming, logs.LogLineLimit)
	if before+len(incoming) > len(p.lines) || p.shown > len(p.lines) {
		p.rebuild()
	} else {
		p.appendLines(logs.FilterLogLines(p.lines[p.shown:], p.term))
		p.shown = len(p.lines)
	}
	p.updateState()
	p.follow()
}

// rebuild writes every visible line again, without touching the stream. The
// search and the switch of the clocks use it.
func (p *logPage) rebuild() {
	p.view.SetPlainText("")
	cursor := p.view.TextCursor()
	cursor.MovePosition(qt6.QTextCursor__End)
	p.appendLinesAt(cursor, logs.FilterLogLines(p.lines, p.term))
	p.shown = len(p.lines)
}

// appendLines writes lines at the end of the view.
func (p *logPage) appendLines(lines []compose.LogLine) {
	cursor := p.view.TextCursor()
	cursor.MovePosition(qt6.QTextCursor__End)
	p.appendLinesAt(cursor, lines)
}

// appendLinesAt writes lines at the cursor, with the format of each part.
func (p *logPage) appendLinesAt(cursor *qt6.QTextCursor, lines []compose.LogLine) {
	for _, line := range lines {
		p.writeLine(cursor, line)
	}
}

// writeLine writes one line: the segments with their tags, a space between the
// parts, and the prefix of a message of compose itself.
func (p *logPage) writeLine(cursor *qt6.QTextCursor, line compose.LogLine) {
	for i, segment := range logs.LogSegments(line, p.clocks.IsChecked()) {
		if i > 0 {
			cursor.InsertText2(" ", p.formats.plain)
		}
		// The prefix sits before the text of compose, not before its clock: the
		// message must read as one piece whatever the rest of the line holds.
		if segment.Tag == logs.LogStderrTag {
			cursor.InsertText2(stderrPrefix, p.formats.stderr)
		}
		cursor.InsertText2(segment.Text, p.formats.format(segment.Tag))
	}
	cursor.InsertText2("\n", p.formats.plain)
}

// searchChanged filters the lines to the term and counts them.
func (p *logPage) searchChanged(text string) {
	p.term = text
	p.rebuild()
	p.updateState()
	p.follow()
}

// togglePause stops the automatic scrolling, and resuming goes back to the
// last line.
func (p *logPage) togglePause(checked bool) {
	p.paused = checked
	if checked {
		p.pause.SetText(activity.LogResumeText)
		return
	}
	p.pause.SetText(activity.LogPauseText)
	p.follow()
}

// updateState aligns the count and the line under the text with the visible
// lines and with the end of the stream. The line carries every hint, so an
// empty page never depends on the faint placeholder of the theme.
func (p *logPage) updateState() {
	visible := len(logs.FilterLogLines(p.lines, p.term))
	count := logs.LogCountText(visible, len(p.lines))
	p.count.SetText(count)
	p.count.SetToolTip(count)

	status, fault := "", false
	switch {
	case p.end != nil && p.end.Error != "":
		status, fault = p.end.Error, true
	case p.end != nil && !p.end.Stopped:
		status = activity.LogEndedText
	case visible == 0 && p.term != "":
		status = activity.LogNoMatchText
	case p.project == "":
		status = activity.EmptyLogText
	case visible == 0:
		status = activity.LogWaitingText
	}

	p.status.SetText(status)
	p.status.SetVisible(status != "")
	if fault {
		p.setStatusRole("logStatusFault")
	} else {
		p.setStatusRole("logStatus")
	}
}

// setStatusRole changes the role of the status line and repolishes it, so the
// colour of the theme follows the new name.
func (p *logPage) setStatusRole(role string) {
	if p.status.ObjectName() == role {
		return
	}
	setObjectName(p.status.QWidget, role)
	p.status.Style().Unpolish(p.status.QWidget)
	p.status.Style().Polish(p.status.QWidget)
}

// follow scrolls to the last line unless the scrolling is paused.
func (p *logPage) follow() {
	if p.paused {
		return
	}
	p.scroller.Start(0)
}

// scrollToEnd puts the view on the last line once the layout has sized it.
func (p *logPage) scrollToEnd() {
	if bar := p.view.VerticalScrollBar(); bar != nil {
		bar.SetValue(bar.Maximum())
	}
}
