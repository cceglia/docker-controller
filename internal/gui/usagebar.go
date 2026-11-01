package gui

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"dockercontroller/internal/compose"
)

const (
	// waitingUsageText is the line before the first sample arrives.
	waitingUsageText = "In attesa della prima misura…"
	// unreachableUsageText is the warning that takes the place of the figures.
	unreachableUsageText = "Docker non raggiungibile"
	// usageTitle names the figures: they are the machine's, not the Project's.
	usageTitle = "Utilizzo risorse"
	// usageStripSpacing is the space between the cells of a row of measures.
	usageStripSpacing = 14
	// usageMeterWidth is the length of a meter of the global strip.
	usageMeterWidth = 120
	// serviceMeterWidth is the length of the meter of a Service.
	serviceMeterWidth = 90
)

// UsageState is what a Service block shows for the Utilizzo risorse.
type UsageState int

const (
	// UsageFigures are the measures of the Service.
	UsageFigures UsageState = iota
	// UsageWaiting is the text before the first sample, while the Service runs.
	UsageWaiting
	// UsageNone is the dash: the Service has no measures, which is not zero.
	UsageNone
)

// UsageStateOf tells what the block of a Service shows from the sample and the
// entry of the Service in it. A Service without a running Container, or a
// sample that carries no figures, shows the dash and never zero.
func UsageStateOf(sample *compose.UsageSample, hasEntry, running bool) UsageState {
	switch {
	case sample == nil:
		if running {
			return UsageWaiting
		}
		return UsageNone
	case !sample.DockerReachable:
		return UsageNone
	case !hasEntry:
		return UsageNone
	}
	return UsageFigures
}

// ServiceRuns reports whether a Service has a Container in execution. Only
// then can it have measures: a Service with no running Container shows the dash.
func ServiceRuns(service compose.Service) bool {
	for _, container := range service.Containers {
		if container.State == "running" {
			return true
		}
	}
	return false
}

// StripState is what the global bar shows.
type StripState int

const (
	// StripWaiting is the text before the first sample.
	StripWaiting StripState = iota
	// StripUnreachable is the warning of a Docker that does not answer.
	StripUnreachable
	// StripFigures are the figures of the machine.
	StripFigures
)

// StripStateOf tells what the global bar shows from the sample. A nil sample is
// a stream that has not delivered yet or that the view stopped.
func StripStateOf(sample *compose.UsageSample) StripState {
	switch {
	case sample == nil:
		return StripWaiting
	case !sample.DockerReachable:
		return StripUnreachable
	}
	return StripFigures
}

// usageStrip is the global Utilizzo risorse: one quiet row above the detail
// with the CPU, the memory, the network and the disk of the machine. It never
// shows the last figures of a stopped stream: without a sample it shows the
// waiting text, and it shows a warning when Docker does not answer.
type usageStrip struct {
	widget  *gtk.Box
	body    *gtk.Stack
	message *gtk.Label

	cpu         *gtk.Label
	cpuNote     *gtk.Label
	cpuMeter    *gtk.LevelBar
	memory      *gtk.Label
	memoryNote  *gtk.Label
	memoryMeter *gtk.LevelBar
	net         *rateCell
	disk        *rateCell
}

// newUsageStrip builds the strip in its waiting state.
func newUsageStrip() *usageStrip {
	s := &usageStrip{}

	owner := gtk.NewBox(gtk.OrientationVertical, 0)
	owner.SetVAlign(gtk.AlignCenter)
	name := gtk.NewLabel("Docker")
	name.SetXAlign(0)
	scope := gtk.NewLabel(usageTitle)
	scope.SetXAlign(0)
	scope.AddCSSClass("caption")
	scope.AddCSSClass("dim-label")
	owner.Append(name)
	owner.Append(scope)

	s.cpu = valueLine()
	s.cpuNote = noteLine()
	s.cpuMeter = usageMeter(usageMeterWidth)
	s.memory = valueLine()
	s.memoryNote = noteLine()
	s.memoryMeter = usageMeter(usageMeterWidth)
	s.net = newRateCell("in", "out")
	s.disk = newRateCell("lettura", "scrittura")

	// The cells of the machine: one per measure, each as wide as the others.
	cells := gtk.NewBox(gtk.OrientationHorizontal, usageStripSpacing)
	cells.SetVAlign(gtk.AlignCenter)
	for i, cell := range []*gtk.Box{
		metricCell("CPU", s.cpu, s.cpuNote, s.cpuMeter),
		metricCell("Memoria", s.memory, s.memoryNote, s.memoryMeter),
		metricCell("Rete", s.net.widget),
		metricCell("Disco", s.disk.widget),
	} {
		if i > 0 {
			cells.Append(separator())
		}
		cell.SetHExpand(true)
		cells.Append(cell)
	}

	// The owner stays on screen in every state, so the message says whose
	// figures are missing.
	s.message = gtk.NewLabel(waitingUsageText)
	s.message.SetXAlign(0)
	s.message.SetVAlign(gtk.AlignCenter)
	s.message.AddCSSClass("dim-label")

	s.body = gtk.NewStack()
	s.body.AddNamed(s.message, "message")
	s.body.AddNamed(cells, "figures")
	s.body.SetVisibleChildName("message")
	s.body.SetHExpand(true)
	s.body.SetVAlign(gtk.AlignCenter)

	row := gtk.NewBox(gtk.OrientationHorizontal, usageStripSpacing)
	row.SetVAlign(gtk.AlignCenter)
	row.Append(owner)
	row.Append(separator())
	row.Append(s.body)

	padded := gtk.NewBox(gtk.OrientationHorizontal, 0)
	padded.SetMarginTop(8)
	padded.SetMarginBottom(8)
	padded.SetMarginStart(18)
	padded.SetMarginEnd(18)
	padded.Append(row)

	s.widget = gtk.NewBox(gtk.OrientationVertical, 0)
	s.widget.Append(padded)
	s.widget.Append(gtk.NewSeparator(gtk.OrientationHorizontal))
	return s
}

// set shows one sample. A nil sample is a stopped stream: the strip shows the
// waiting text and never the figures of the sample before.
func (s *usageStrip) set(sample *compose.UsageSample) {
	switch StripStateOf(sample) {
	case StripWaiting:
		s.widget.SetTooltipText("")
		s.showMessage(waitingUsageText, false)
	case StripUnreachable:
		s.showMessage(unreachableUsageText, true)
		s.widget.SetTooltipText(sample.Error)
	default:
		s.widget.SetTooltipText("")
		s.cpu.SetText(FormatCores(sample.Total.CPUPercent))
		s.cpuNote.SetText(fmt.Sprintf("%s della macchina", FormatPercent(sample.CPUMachinePercent)))
		s.cpuMeter.SetValue(levelValue(sample.CPUMachinePercent))
		s.memory.SetText(FormatBytes(float64(sample.Total.MemoryBytes), ByteBaseBinary))
		s.memoryNote.SetText(fmt.Sprintf("%s di %s",
			FormatPercent(sample.MemoryMachinePercent),
			FormatBytes(float64(sample.MemoryTotalBytes), ByteBaseBinary)))
		s.memoryMeter.SetValue(levelValue(sample.MemoryMachinePercent))
		s.net.set(sample.Rate.NetInBytesPerSec, sample.Rate.NetOutBytesPerSec)
		s.disk.set(sample.Rate.BlockReadBytesPerSec, sample.Rate.BlockWriteBytesPerSec)
		s.body.SetVisibleChildName("figures")
	}
}

// showMessage puts one line in place of the figures.
func (s *usageStrip) showMessage(text string, warning bool) {
	s.message.SetText(text)
	if warning {
		s.message.RemoveCSSClass("dim-label")
		s.message.AddCSSClass("warning")
	} else {
		s.message.RemoveCSSClass("warning")
		s.message.AddCSSClass("dim-label")
	}
	s.body.SetVisibleChildName("message")
}

// serviceUsageBlock is the Utilizzo risorse of one Service, inside its block:
// the CPU with its meter and its cores, the memory and the two rates. A Service
// without running Containers shows a dash, never zero.
type serviceUsageBlock struct {
	widget  *gtk.Box
	body    *gtk.Stack
	message *gtk.Label
	figures *gtk.Box
	// running says whether the Service has a Container in execution, so the
	// block waits for a measure only while one can arrive.
	running bool

	cpu      *gtk.Label
	cpuNote  *gtk.Label
	cpuMeter *gtk.LevelBar
	memory   *gtk.Label
	net      *rateCell
	disk     *rateCell
}

// newServiceUsageBlock builds the block of one Service. running says whether
// the Service has a Container in execution, so the waiting text is shown only
// while a measure can still arrive.
func newServiceUsageBlock(running bool) *serviceUsageBlock {
	b := &serviceUsageBlock{running: running}

	title := gtk.NewLabel(usageTitle)
	title.SetXAlign(0)
	title.AddCSSClass("caption")
	title.AddCSSClass("dim-label")

	b.cpu = valueLine()
	b.cpuNote = noteLine()
	b.cpuMeter = usageMeter(serviceMeterWidth)
	b.memory = valueLine()
	b.net = newRateCell("in", "out")
	b.disk = newRateCell("lettura", "scrittura")

	b.figures = gtk.NewBox(gtk.OrientationHorizontal, usageStripSpacing)
	b.figures.SetMarginTop(4)
	b.figures.SetMarginBottom(4)
	for _, cell := range []*gtk.Box{
		metricCell("CPU", b.cpu, b.cpuNote, b.cpuMeter),
		metricCell("Memoria", b.memory),
		metricCell("Rete", b.net.widget),
		metricCell("Disco", b.disk.widget),
	} {
		cell.SetHExpand(true)
		b.figures.Append(cell)
	}

	b.message = gtk.NewLabel("")
	b.message.SetXAlign(0)
	b.message.AddCSSClass("dim-label")
	b.message.SetMarginTop(4)
	b.message.SetMarginBottom(4)

	b.body = gtk.NewStack()
	b.body.AddNamed(b.message, "message")
	b.body.AddNamed(b.figures, "figures")

	b.widget = gtk.NewBox(gtk.OrientationVertical, 0)
	b.widget.Append(title)
	b.widget.Append(b.body)

	b.set(nil, nil)
	return b
}

// set shows one sample and the entry of this Service, if the sample has one.
// Without a sample the block waits when the Service runs, and shows a dash
// otherwise: a Service with no running Container has no measures, which is not
// zero.
func (b *serviceUsageBlock) set(sample *compose.UsageSample, usage *compose.ServiceUsage) {
	switch UsageStateOf(sample, usage != nil, b.running) {
	case UsageWaiting:
		b.showMessage(waitingUsageText)
	case UsageFigures:
		b.cpu.SetText(FormatPercent(usage.Measures.CPUPercent))
		b.cpuNote.SetText(FormatCores(usage.Measures.CPUPercent))
		b.cpuMeter.SetValue(levelValue(usage.Measures.CPUPercent))
		b.memory.SetText(FormatBytes(float64(usage.Measures.MemoryBytes), ByteBaseBinary))
		b.net.set(usage.Rate.NetInBytesPerSec, usage.Rate.NetOutBytesPerSec)
		b.disk.set(usage.Rate.BlockReadBytesPerSec, usage.Rate.BlockWriteBytesPerSec)
		b.body.SetVisibleChildName("figures")
	default:
		b.showMessage("—")
	}
}

// showMessage puts one line in place of the figures.
func (b *serviceUsageBlock) showMessage(text string) {
	b.message.SetText(text)
	b.body.SetVisibleChildName("message")
}

// rateCell is a pair of rates, in and out or read and written: one row per
// reading, the labels in a column of their own, so the values line up.
type rateCell struct {
	widget *gtk.Grid
	first  *gtk.Label
	second *gtk.Label
}

// newRateCell builds the pair with its two directions.
func newRateCell(first, second string) *rateCell {
	c := &rateCell{widget: gtk.NewGrid()}
	c.widget.SetRowSpacing(0)
	c.widget.SetColumnSpacing(6)
	c.first = rateValue(c.widget, 0, first)
	c.second = rateValue(c.widget, 1, second)
	return c
}

// rateValue adds one reading at row of the grid and returns its value label.
func rateValue(grid *gtk.Grid, row int, direction string) *gtk.Label {
	label := gtk.NewLabel(direction)
	label.SetXAlign(0)
	label.AddCSSClass("caption")
	label.AddCSSClass("dim-label")
	label.SetVAlign(gtk.AlignCenter)
	grid.Attach(label, 0, row, 1, 1)

	value := gtk.NewLabel("")
	value.SetXAlign(0)
	value.AddCSSClass("numeric")
	value.AddCSSClass("caption")
	value.SetVAlign(gtk.AlignCenter)
	grid.Attach(value, 1, row, 1, 1)
	return value
}

// set shows the two rates of the pair, in bytes per second.
func (c *rateCell) set(first, second float64) {
	c.first.SetText(FormatRate(first))
	c.second.SetText(FormatRate(second))
}

// metricCell is one cell of a row of measures: a small title and its lines.
func metricCell(title string, lines ...gtk.Widgetter) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationVertical, 2)
	box.SetVAlign(gtk.AlignStart)

	label := gtk.NewLabel(title)
	label.SetXAlign(0)
	label.AddCSSClass("caption")
	label.AddCSSClass("dim-label")
	box.Append(label)

	for _, line := range lines {
		box.Append(line)
	}
	return box
}

// valueLine is a figure of a cell: tabular digits, aligned to the start.
func valueLine() *gtk.Label {
	label := gtk.NewLabel("")
	label.SetXAlign(0)
	label.AddCSSClass("numeric")
	return label
}

// noteLine is the second line of a cell: a share, a size or a count.
func noteLine() *gtk.Label {
	label := gtk.NewLabel("")
	label.SetXAlign(0)
	label.AddCSSClass("caption")
	label.AddCSSClass("dim-label")
	label.AddCSSClass("numeric")
	return label
}

// usageMeter is the thin GtkLevelBar of a percentage.
func usageMeter(width int) *gtk.LevelBar {
	meter := gtk.NewLevelBarForInterval(0, 1)
	meter.SetValue(0)
	meter.SetSizeRequest(width, -1)
	meter.SetHAlign(gtk.AlignStart)
	return meter
}

// separator is the hairline between two cells of a row of measures.
func separator() *gtk.Separator {
	line := gtk.NewSeparator(gtk.OrientationVertical)
	line.SetMarginTop(2)
	line.SetMarginBottom(2)
	return line
}
