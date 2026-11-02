package gui

import (
	"github.com/mappu/miqt/qt6"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/design"
	"dockercontroller/internal/gui/usage"
)

const (
	// usageStripName names the bar of the machine: its figures are not a
	// Project's, and the strip says so.
	usageStripName = usage.Title + " della macchina"
	// usageServiceName begins the name of the indicator of one Service.
	usageServiceName = usage.Title + " del Servizio "
	// usageBarWidth is the length of the bar of a percentage. A figure keeps its
	// length, so a row never moves when a number grows.
	usageBarWidth = 64
)

// usageMeter builds the bar of a percentage. It is a figure, not a control, so
// it takes no focus and no tab stop.
func usageMeter(name string) *qt6.QProgressBar {
	bar := qt6.NewQProgressBar2()
	bar.SetRange(0, usage.MeterSteps)
	bar.SetValue(0)
	bar.SetTextVisible(false)
	bar.SetFixedWidth(usageBarWidth)
	bar.SetFocusPolicy(qt6.NoFocus)
	bar.SetAccessibleName(name)
	return bar
}

// usageCell is one measure of a row of the Utilizzo risorse. The first line is
// the caption, with the bar when the measure has a percentage; the second line
// is the figure, followed by what the figure alone cannot say.
type usageCell struct {
	widget *qt6.QWidget
	label  *qt6.QLabel
	bar    *qt6.QProgressBar
	value  *qt6.QLabel
	note   *qt6.QLabel
}

// newUsageCell builds one measure. withBar gives it the bar of a percentage and
// withNote the explanation after the figure. The figure always sits under the
// caption, so the value of every cell starts on the same line and the bar keeps
// the caption to itself. A cell that may wrap shrinks at a narrow window; a
// cell that may not keeps one line, so the row of a Service never changes
// height with the width of the window.
func newUsageCell(caption string, withBar, withNote, wrap bool) *usageCell {
	c := &usageCell{widget: qt6.NewQWidget2()}
	c.label = qt6.NewQLabel3(caption)

	heading := qt6.NewQHBoxLayout2()
	heading.SetContentsMargins(0, 0, 0, 0)
	heading.SetSpacing(design.SpaceS)
	heading.AddWidget3(c.label.QWidget, 0, qt6.AlignVCenter)
	if withBar {
		c.bar = usageMeter(caption)
		heading.AddWidget3(c.bar.QWidget, 0, qt6.AlignVCenter)
	}
	heading.AddStretch()

	c.value = qt6.NewQLabel3("")
	c.value.SetWordWrap(wrap)
	figures := qt6.NewQHBoxLayout2()
	figures.SetContentsMargins(0, 0, 0, 0)
	figures.SetSpacing(design.SpaceS)
	figures.AddWidget3(c.value.QWidget, 0, qt6.AlignVCenter)
	if withNote {
		c.note = qt6.NewQLabel3("")
		c.note.SetWordWrap(wrap)
		figures.AddWidget3(c.note.QWidget, 0, qt6.AlignVCenter)
	}
	figures.AddStretch()

	box := qt6.NewQVBoxLayout(c.widget)
	box.SetContentsMargins(0, 0, 0, 0)
	box.SetSpacing(0)
	box.AddLayout2(heading.QLayout, 0)
	box.AddLayout2(figures.QLayout, 0)
	c.widget.SetAccessibleName(caption)
	return c
}

// set shows one measure. A measure without a note hides its second line, so a
// cell never keeps room for text it does not have.
func (c *usageCell) set(value, note string, percent float64) {
	c.value.SetText(value)
	if c.note != nil {
		c.note.SetText(note)
		c.note.SetVisible(note != "")
	}
	if c.bar != nil {
		c.bar.SetValue(usage.MeterValue(percent))
	}
	description := value
	if note != "" {
		description += ", " + note
	}
	c.value.SetAccessibleDescription(description)
}

// usageRow is one row of the Utilizzo risorse: the four measures, each in a cell
// of its own. The bar of the machine carries four cells with a bar for the CPU
// and for the memory; a Service carries the bar of the CPU alone.
type usageRow struct {
	widget *qt6.QWidget
	cpu    *usageCell
	memory *usageCell
	net    *usageCell
	disk   *usageCell
}

// usageRowOptions says which bars a row carries and whether its figures may
// wrap. The bar of the machine wraps at a narrow window, so no figure is cut
// off; the row of a Service never wraps, so the height of a Service block does
// not depend on the width of the window.
type usageRowOptions struct {
	cpuBar    bool
	memoryBar bool
	wrap      bool
}

// newUsageRow builds the four cells of a row of the Utilizzo risorse.
func newUsageRow(options usageRowOptions) *usageRow {
	r := &usageRow{
		widget: qt6.NewQWidget2(),
		cpu:    newUsageCell("CPU", options.cpuBar, true, options.wrap),
		memory: newUsageCell("Memoria", options.memoryBar, true, options.wrap),
		net:    newUsageCell("Rete", false, true, options.wrap),
		disk:   newUsageCell("Disco", false, true, options.wrap),
	}

	line := qt6.NewQHBoxLayout(r.widget)
	line.SetContentsMargins(0, 0, 0, 0)
	line.SetSpacing(design.SpaceM)
	for i, cell := range []*usageCell{r.cpu, r.memory, r.net, r.disk} {
		if i > 0 {
			line.AddWidget(usageRule())
		}
		// The cells line up at their caption: one measure without a note must
		// not move the figure of the one beside it.
		line.AddWidget3(cell.widget, 0, qt6.AlignTop)
	}
	line.AddStretch()
	return r
}

// set shows the figures of one row. cpuPercent is the value of the CPU bar and
// memoryShare the value of the memory bar; a row without a bar ignores them.
func (r *usageRow) set(texts usage.Texts, cpuPercent, memoryShare float64) {
	r.cpu.set(texts.CPU.Value, texts.CPU.Note, cpuPercent)
	r.memory.set(texts.Memory.Value, texts.Memory.Note, memoryShare)
	r.net.set(texts.Net.Value, texts.Net.Note, 0)
	r.disk.set(texts.Disk.Value, texts.Disk.Note, 0)
}

// usageStrip is the Utilizzo risorse of the machine, above the detail: one row
// with the CPU, the memory, the network and the disk, or the line that takes
// their place. It never shows the figures of a stream that stopped.
type usageStrip struct {
	widget  *qt6.QWidget
	body    *qt6.QStackedWidget
	message *qt6.QLabel
	row     *usageRow
}

// newUsageStrip builds the bar in its waiting state: the line before the first
// sample, and no figure at all.
func newUsageStrip() *usageStrip {
	s := &usageStrip{row: newUsageRow(usageRowOptions{cpuBar: true, memoryBar: true, wrap: true})}

	s.message = qt6.NewQLabel3(usage.WaitingText)
	s.message.SetWordWrap(true)

	s.body = qt6.NewQStackedWidget2()
	s.body.AddWidget(s.message.QWidget)
	s.body.AddWidget(s.row.widget)
	s.body.SetCurrentWidget(s.message.QWidget)

	padded := qt6.NewQWidget2()
	pad := qt6.NewQHBoxLayout(padded)
	pad.SetContentsMargins(design.SpaceL, design.SpaceM, design.SpaceL, design.SpaceM)
	pad.SetSpacing(design.SpaceM)
	pad.AddWidget2(s.body.QWidget, 1)

	s.widget = qt6.NewQWidget2()
	s.widget.SetAccessibleName(usageStripName)
	box := qt6.NewQVBoxLayout(s.widget)
	box.SetContentsMargins(0, 0, 0, 0)
	box.SetSpacing(0)
	box.AddWidget(padded)
	box.AddWidget(detailSeparator())
	return s
}

// set shows one sample. Without a sample, and with a Docker that does not
// answer, a line takes the place of the figures: the strip never shows the
// numbers of a stream that stopped.
func (s *usageStrip) set(sample *compose.UsageSample) {
	switch usage.StripStateOf(sample) {
	case usage.StripWaiting:
		s.widget.SetToolTip("")
		s.showMessage(usage.WaitingText, false)
	case usage.StripUnreachable:
		s.widget.SetToolTip(sample.Error)
		s.showMessage(usage.UnreachableText, true)
	default:
		s.widget.SetToolTip("")
		s.row.set(usage.TextsOf(usage.Figures{
			Measures:    sample.Total,
			Rate:        sample.Rate,
			CPUPercent:  sample.CPUMachinePercent,
			MemoryTotal: sample.MemoryTotalBytes,
		}), sample.CPUMachinePercent, sample.MemoryMachinePercent)
		s.body.SetCurrentWidget(s.row.widget)
	}
}

// setUnreachable shows the warning of a Docker that does not answer, which the
// snapshot knows before any sample does.
func (s *usageStrip) setUnreachable() {
	s.widget.SetToolTip("")
	s.showMessage(usage.UnreachableText, true)
}

// showMessage puts one line in place of the figures, in the role of a fault when
// Docker does not answer. The colour is a role of the style sheet, so the line
// keeps its contrast in both themes.
func (s *usageStrip) showMessage(text string, fault bool) {
	s.message.SetText(text)
	s.message.SetAccessibleName(usageStripName)
	s.message.SetAccessibleDescription(text)
	role := ""
	if fault {
		role = design.RoleStatusFault
	}
	if s.message.ObjectName() != role {
		setObjectName(s.message.QWidget, role)
		// A role that changes on a widget already on screen needs the sheet
		// applied again, or the line keeps the colour of the line before.
		s.message.Style().Unpolish(s.message.QWidget)
		s.message.Style().Polish(s.message.QWidget)
	}
	s.body.SetCurrentWidget(s.message.QWidget)
}

// serviceUsageRow is the Utilizzo risorse of one Service, inside its block: the
// four measures of its Containers on one row. It stays hidden until a sample
// carries the Service, so a Service without measures reserves no space and a
// Service without a Container shows no indicator at all.
type serviceUsageRow struct {
	widget  *qt6.QWidget
	row     *usageRow
	key     string
	running bool
}

// newServiceUsageRow builds the indicator of a Service, hidden: only a sample
// that carries the Service brings it on screen. running says whether a measure
// can arrive at all.
func newServiceUsageRow(project, service string, running bool) *serviceUsageRow {
	r := &serviceUsageRow{
		row:     newUsageRow(usageRowOptions{cpuBar: true}),
		key:     project + "/" + service,
		running: running,
	}
	r.widget = r.row.widget
	r.widget.SetAccessibleName(usageServiceName + service)
	r.widget.SetVisible(false)
	return r
}

// set shows one sample for the Service. A Service without figures takes its
// indicator off the layout: a row of zeros would read as a measure.
func (r *serviceUsageRow) set(state usage.UsageState, sample compose.ServiceUsage) {
	if !r.running || state != usage.UsageFigures {
		r.widget.SetVisible(false)
		return
	}
	r.row.set(usage.TextsOf(usage.Figures{
		Measures:   sample.Measures,
		Rate:       sample.Rate,
		CPUPercent: sample.Measures.CPUPercent,
	}), sample.Measures.CPUPercent, 0)
	r.widget.SetVisible(true)
}

// usageRule is the hairline between two cells of a row of measures. A QFrame
// draws it with the colours of the desktop palette.
func usageRule() *qt6.QWidget {
	line := qt6.NewQFrame2()
	line.SetFrameShape(qt6.QFrame__VLine)
	line.SetFrameShadow(qt6.QFrame__Sunken)
	line.SetFocusPolicy(qt6.NoFocus)
	line.SetAttribute2(qt6.WA_TransparentForMouseEvents, true)
	return line.QWidget
}
