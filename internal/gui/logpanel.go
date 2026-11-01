package gui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"

	"dockercontroller/internal/compose"
)

const (
	// allServicesText is the first choice of the selector: the whole Project.
	allServicesText = "Tutti"
	// searchPlaceholder is the hint of the search entry.
	searchPlaceholder = "Cerca nel testo"
	// timestampsText is the label of the switch of the clocks.
	timestampsText = "Timestamp"
	// pauseText and resumeText are the two faces of the pause button.
	pauseText  = "Pausa"
	resumeText = "Riprendi"
	// closeLogText is the button that ends the stream.
	closeLogText = "Chiudi log"
	// waitingLogText is the line before the first line arrives.
	waitingLogText = "In attesa di righe…"
	// noMatchLogText is the line when the search hides every line.
	noMatchLogText = "Nessuna riga corrisponde alla ricerca."
	// endedLogText is the line under the text when compose closed the stream.
	endedLogText = "Il flusso dei log è terminato."
	// barSpacing is the space between the controls of the bar.
	barSpacing = 8
)

// logPanel is the page of the logs of the bottom panel: the logs of a Project,
// or of one of its Services. It owns the stream: opening it starts
// `docker compose logs`, closing it terminates the process.
//
// Lines arrive on a goroutine and wait in pending; the main loop moves them
// into the buffer in bursts, so a fast stream does not rebuild the text once
// per line.
type logPanel struct {
	// widget is the page, page its place in the stack of the panel.
	widget *gtk.Box
	page   *gtk.StackPage

	selector   *gtk.DropDown
	search     *gtk.SearchEntry
	timestamps *gtk.CheckButton
	pause      *gtk.ToggleButton
	count      *gtk.Label
	content    *gtk.Stack
	empty      *gtk.Label
	view       *gtk.TextView
	buffer     *gtk.TextBuffer
	scrolled   *gtk.ScrolledWindow
	status     *gtk.Label

	// source opens the stream: it is compose.Manager.Logs in the application.
	source LogSource
	// ctx stops every stream when the application quits.
	ctx context.Context
	// streams keeps the running stream, so a close that overtakes an opening
	// still terminates it.
	streams LogStreams

	// project and service are what the page streams now; services are the
	// names of the selector, in its order, and filling is true while the
	// selector is rebuilt, so its signals do not restart the stream.
	project  string
	service  string
	services []string
	filling  bool
	// id is the stream in the registry, "" when none runs.
	id string
	// lines are the lines the panel keeps, at most LogLineLimit; shown is how
	// many of them the buffer already holds. Both belong to the main loop.
	lines []compose.LogLine
	shown int
	// term is the search, lower case as the filter wants it; paused stops the
	// automatic scrolling only, never the stream.
	term   string
	paused bool
	// end is how the stream ended on its own, nil while it runs.
	end *LogEnd

	// mu guards generation and pending, which the goroutine of the stream
	// touches too.
	mu         sync.Mutex
	generation int
	pending    []compose.LogLine
	scheduled  atomic.Bool

	// onClose ends the page: the button of the page asks the panel.
	onClose func()
}

// newLogPanel builds the page. It does not start a stream: Open does.
func newLogPanel() *logPanel {
	p := &logPanel{}

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

	p.empty = gtk.NewLabel(waitingLogText)
	p.empty.SetXAlign(0)
	p.empty.SetYAlign(0)
	p.empty.SetWrap(true)
	p.empty.AddCSSClass("dim-label")
	p.empty.SetMarginTop(8)
	p.empty.SetMarginStart(8)

	p.content = gtk.NewStack()
	p.content.AddNamed(p.empty, "empty")
	p.content.AddNamed(p.scrolled, "lines")
	p.content.SetVisibleChildName("lines")
	p.content.SetVExpand(true)

	p.status = gtk.NewLabel("")
	p.status.SetXAlign(0)
	p.status.SetWrap(true)
	p.status.SetMarginTop(2)
	p.status.SetMarginBottom(4)
	p.status.SetMarginStart(8)
	p.status.AddCSSClass("dim-label")
	p.status.SetVisible(false)

	p.installTags()

	p.selector = gtk.NewDropDownFromStrings([]string{allServicesText})
	p.selector.SetTooltipText("Scegli il Servizio. Con Tutti il pannello segue l'intero Progetto.")
	p.selector.NotifyProperty("selected", p.serviceChanged)

	p.search = gtk.NewSearchEntry()
	p.search.SetPlaceholderText(searchPlaceholder)
	p.search.SetHExpand(true)
	p.search.ConnectSearchChanged(func() {
		p.term = strings.TrimSpace(p.search.Text())
		p.rebuild()
		p.updateState()
		p.follow()
	})

	p.timestamps = gtk.NewCheckButtonWithLabel(timestampsText)
	p.timestamps.SetActive(true)
	p.timestamps.ConnectToggled(func() {
		p.rebuild()
		p.updateState()
		p.follow()
	})

	p.pause = gtk.NewToggleButtonWithLabel(pauseText)
	p.pause.SetTooltipText("Ferma lo scorrimento automatico. Le righe nuove continuano ad arrivare.")
	p.pause.ConnectToggled(func() {
		p.paused = p.pause.Active()
		if p.paused {
			p.pause.SetLabel(resumeText)
			return
		}
		p.pause.SetLabel(pauseText)
		p.scrollToEnd()
	})

	close := gtk.NewButtonWithLabel(closeLogText)
	close.AddCSSClass("flat")
	close.SetTooltipText("Termina il processo di docker compose logs.")
	close.ConnectClicked(func() {
		if p.onClose != nil {
			p.onClose()
		}
	})

	p.count = gtk.NewLabel(LogCountText(0, 0))
	p.count.AddCSSClass("dim-label")
	p.count.SetHAlign(gtk.AlignEnd)
	p.count.SetEllipsize(pango.EllipsizeEnd)
	p.count.SetTooltipText(LogCountText(0, 0))

	serviceLabel := gtk.NewLabel("Servizio")

	bar := gtk.NewBox(gtk.OrientationHorizontal, barSpacing)
	bar.SetMarginTop(4)
	bar.SetMarginBottom(4)
	bar.SetMarginStart(8)
	bar.SetMarginEnd(8)
	bar.Append(serviceLabel)
	bar.Append(p.selector)
	bar.Append(p.search)
	bar.Append(p.timestamps)
	bar.Append(p.pause)
	bar.Append(close)
	bar.Append(p.count)

	p.widget = gtk.NewBox(gtk.OrientationVertical, 0)
	p.widget.SetMarginBottom(4)
	p.widget.Append(bar)
	p.widget.Append(p.content)
	p.widget.Append(p.status)

	// The colours of the tags follow the theme: they are set when the view is
	// mapped and again when Adwaita switches between light and dark.
	p.view.ConnectMap(p.applyTheme)
	if manager := adw.StyleManagerGetDefault(); manager != nil {
		manager.NotifyProperty("dark", p.applyTheme)
	}

	return p
}

// Open shows the logs of the whole Project, or of one of its Services, and
// starts the stream: the last 500 lines at once, then the new ones.
func (p *logPanel) Open(project compose.Project, service string) {
	if p.source == nil {
		return // the shell has not wired the Manager yet
	}
	// The same stream already open is shown again, not restarted: the last
	// lines are already on screen.
	if p.id != "" && p.project == project.Name && p.service == service {
		return
	}
	p.resetView()
	p.project = project.Name
	p.fillServices(project, service)
	p.start(project.Name, service)
}

// Stop terminates the running stream. The lines stay on screen.
func (p *logPanel) Stop() {
	p.stopStream()
	p.end = nil
	p.updateStatus()
}

// StopAll terminates every stream: the window is going away.
func (p *logPanel) StopAll() {
	p.stopStream()
	p.streams.StopAll()
}

// resetView puts the controls back to their first value: a page that is opened
// again starts without a search, with the clocks and without a pause.
func (p *logPanel) resetView() {
	p.search.SetText("")
	p.term = ""
	p.timestamps.SetActive(true)
	p.paused = false
	p.pause.SetActive(false)
	p.pause.SetLabel(pauseText)
}

// fillServices puts Tutti and the Services of the Project in the selector.
func (p *logPanel) fillServices(project compose.Project, service string) {
	p.services = make([]string, 0, len(project.Services)+1)
	p.services = append(p.services, allServicesText)
	selected := uint(0)
	for i, s := range project.Services {
		p.services = append(p.services, s.Name)
		if s.Name == service {
			selected = uint(i + 1)
		}
	}
	p.filling = true
	p.selector.SetModel(gtk.NewStringList(p.services))
	p.selector.SetSelected(selected)
	p.filling = false
}

// serviceChanged restarts the stream when the user chooses another Service.
func (p *logPanel) serviceChanged() {
	if p.source == nil || p.filling {
		return
	}
	service := p.selectedService()
	if service == p.service {
		return
	}
	p.start(p.project, service)
}

// selectedService is the Service of the selector, "" for the whole Project.
func (p *logPanel) selectedService() string {
	selected := p.selector.Selected()
	if selected == gtk.InvalidListPosition || int(selected) >= len(p.services) {
		return ""
	}
	if p.services[selected] == allServicesText {
		return ""
	}
	return p.services[selected]
}

// start terminates the stream on screen, if any, and opens the new one.
func (p *logPanel) start(project, service string) {
	if p.source == nil {
		return
	}
	p.stopStream()
	p.project, p.service = project, service
	p.lines = nil
	p.shown = 0
	p.end = nil
	p.buffer.SetText("")
	p.updateStatus()
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

	id := strconv.Itoa(generation)
	if err := p.streams.Add(id, cancel); err != nil {
		// The page was closed before the stream registered: do not start.
		cancel()
		return
	}
	p.id = id

	source := p.source
	go func() {
		end := StreamLogs(streamCtx, compose.LogRequest{Project: project, Service: service}, source,
			func(line compose.LogLine) { p.push(generation, line) })
		p.streams.Release(id)
		if end.Stopped {
			return
		}
		p.finish(generation, end)
	}()
}

// stopStream terminates the running stream and forgets the lines still in
// flight: a line of a stopped stream must not reach the buffer.
func (p *logPanel) stopStream() {
	p.mu.Lock()
	p.generation++
	p.pending = nil
	p.mu.Unlock()
	if p.id != "" {
		p.streams.Stop(p.id)
		p.id = ""
	}
}

// push queues one line of the stream for the main loop.
func (p *logPanel) push(generation int, line compose.LogLine) {
	p.mu.Lock()
	if generation != p.generation {
		p.mu.Unlock()
		return
	}
	p.pending = append(p.pending, line)
	p.mu.Unlock()
	p.schedule()
}

// finish records how the stream ended, on the main loop.
func (p *logPanel) finish(generation int, end LogEnd) {
	glib.IdleAdd(func() {
		p.mu.Lock()
		current := generation == p.generation
		p.mu.Unlock()
		if !current {
			return
		}
		p.end = &end
		p.id = ""
		p.updateStatus()
		p.updateState()
	})
}

// schedule queues one flush on the main loop. Lines arrive in bursts, so the
// panel writes them together.
func (p *logPanel) schedule() {
	if !p.scheduled.CompareAndSwap(false, true) {
		return
	}
	glib.IdleAdd(func() {
		p.scheduled.Store(false)
		p.flush()
	})
}

// flush moves the waiting lines into the history and the buffer. When the cap
// dropped lines, or the filter is active, the buffer is written again from the
// history; otherwise only the new lines are appended.
func (p *logPanel) flush() {
	p.mu.Lock()
	incoming := p.pending
	p.pending = nil
	p.mu.Unlock()
	if len(incoming) == 0 {
		return
	}

	before := len(p.lines)
	p.lines = AppendLogLines(p.lines, incoming, LogLineLimit)
	if before+len(incoming) > len(p.lines) || p.shown > len(p.lines) {
		p.rebuild()
	} else {
		p.appendLines(FilterLogLines(p.lines[p.shown:], p.term))
		p.shown = len(p.lines)
	}
	p.updateState()
	p.follow()
}

// rebuild writes every visible line again, without touching the stream. The
// search and the switch of the timestamps use it.
func (p *logPanel) rebuild() {
	p.buffer.SetText("")
	p.appendLines(FilterLogLines(p.lines, p.term))
	p.shown = len(p.lines)
}

// appendLines writes lines at the end of the buffer, with the tag of each part.
func (p *logPanel) appendLines(lines []compose.LogLine) {
	for _, line := range lines {
		p.writeLine(line)
	}
}

// writeLine writes one line: the segments with their tags, a space between the
// parts, and the newline.
func (p *logPanel) writeLine(line compose.LogLine) {
	for i, segment := range LogSegments(line, p.timestamps.Active()) {
		if i > 0 {
			p.buffer.Insert(p.buffer.EndIter(), " ")
		}
		start := p.buffer.CharCount()
		p.buffer.Insert(p.buffer.EndIter(), segment.Text)
		if segment.Tag == "" {
			continue
		}
		tag := p.buffer.TagTable().Lookup(segment.Tag)
		if tag == nil {
			continue
		}
		p.buffer.ApplyTag(tag, p.buffer.IterAtOffset(start), p.buffer.IterAtOffset(p.buffer.CharCount()))
	}
	p.buffer.Insert(p.buffer.EndIter(), "\n")
}

// updateState aligns the count and the empty notice with the visible lines.
func (p *logPanel) updateState() {
	visible := len(FilterLogLines(p.lines, p.term))
	count := LogCountText(visible, len(p.lines))
	p.count.SetText(count)
	p.count.SetTooltipText(count)
	switch {
	case visible > 0:
		p.content.SetVisibleChildName("lines")
	case p.term != "":
		p.empty.SetText(noMatchLogText)
		p.content.SetVisibleChildName("empty")
	case p.end == nil:
		p.empty.SetText(waitingLogText)
		p.content.SetVisibleChildName("empty")
	default:
		p.content.SetVisibleChildName("lines")
	}
}

// updateStatus shows how the stream ended, or hides the line while it runs.
func (p *logPanel) updateStatus() {
	switch {
	case p.end == nil || p.end.Stopped:
		p.status.SetVisible(false)
	case p.end.Error != "":
		p.status.SetText(p.end.Error)
		p.status.RemoveCSSClass("dim-label")
		p.status.AddCSSClass("error")
		p.status.SetVisible(true)
	default:
		p.status.SetText(endedLogText)
		p.status.RemoveCSSClass("error")
		p.status.AddCSSClass("dim-label")
		p.status.SetVisible(true)
	}
}

// follow scrolls to the last line unless the scrolling is paused.
func (p *logPanel) follow() {
	if p.paused {
		return
	}
	p.scrollToEnd()
}

// scrollToEnd puts the view on the last line when the main loop is idle.
func (p *logPanel) scrollToEnd() {
	scrolled := p.scrolled
	glib.IdleAdd(func() {
		adjustment := scrolled.VAdjustment()
		adjustment.SetValue(adjustment.Upper())
	})
}

// installTags creates the tags of the buffer, once: one colour per palette
// entry, the clock and the messages of compose.
func (p *logPanel) installTags() {
	table := p.buffer.TagTable()
	for hue := 0; hue < ServiceHues; hue++ {
		tag := gtk.NewTextTag(serviceTag(hue))
		tag.SetObjectProperty("weight", int(pango.WeightBold))
		table.Add(tag)
	}
	table.Add(gtk.NewTextTag(LogTimestampTag))
	table.Add(gtk.NewTextTag(LogStderrTag))
}

// applyTheme gives the tags their colour for the theme on screen: the Service
// colours, the clock dimmed, the messages of compose in the error colour.
func (p *logPanel) applyTheme() {
	dark := false
	if manager := adw.StyleManagerGetDefault(); manager != nil {
		dark = manager.Dark()
	}
	table := p.buffer.TagTable()
	for hue := 0; hue < ServiceHues; hue++ {
		if tag := table.Lookup(serviceTag(hue)); tag != nil {
			tag.SetObjectProperty("foreground", ServiceColor(hue, dark))
		}
	}
	style := p.view.StyleContext()
	if color, ok := style.LookupColor("window_fg_color"); ok {
		if tag := table.Lookup(LogTimestampTag); tag != nil {
			tag.SetObjectProperty("foreground", rgbaText(color, 0.6))
		}
	}
	if color, ok := style.LookupColor("error_color"); ok {
		if tag := table.Lookup(LogStderrTag); tag != nil {
			tag.SetObjectProperty("foreground", rgbaText(color, 1))
		}
	}
}

// rgbaText is a colour of the theme as the text of an RGBA colour.
func rgbaText(color *gdk.RGBA, alpha float64) string {
	return fmt.Sprintf("rgba(%d,%d,%d,%.2f)",
		int(color.Red()*255+0.5), int(color.Green()*255+0.5), int(color.Blue()*255+0.5), alpha)
}
