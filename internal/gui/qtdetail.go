package gui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mappu/miqt/qt6"

	"dockercontroller/internal/compose"
)

// The texts of the panel. They live with the widgets, so the tickets that wire
// the commands change no text.
const (
	// servicesTitle is the heading of the list of the Services.
	servicesTitle = "Servizi"
	// folderCommandText, terminalCommandText and portCommandText are the
	// commands of the Project. They stay insensitive until the ticket that
	// opens them.
	folderCommandText   = "Cartella"
	terminalCommandText = "Terminale"
	portCommandText     = "Porta"
)

// The tooltips of the commands, the ones of the old view.
const (
	folderTooltip   = "Apri la cartella del Progetto nel gestore file."
	terminalTooltip = "Apri un terminale nella cartella del Progetto."
	portTooltip     = "Apri la prima porta pubblicata del Progetto nel browser."
)

// serviceUsageHeight is the height a Service keeps for the Utilizzo risorse.
// The indicator arrives in a later ticket: the place is reserved now, so the
// rows do not change height when it arrives.
const serviceUsageHeight = 28

// detailPanel is the panel of the right: the Project the user chose, its data,
// its Services and the rows of the Actions and of the commands. It reads the
// snapshot and never changes a Project. The Actions, the commands and the entry
// that takes a Project off the list stay insensitive until the tickets that
// reach the Manager.
type detailPanel struct {
	theme designTheme

	// widget is what the shell puts in the splitter and in the narrow page. It
	// holds the invitation and the detail, one at a time.
	widget *qt6.QWidget
	stack  *qt6.QStackedWidget
	invite *qt6.QWidget
	scroll *qt6.QScrollArea
	// content holds the blocks of the detail. A render empties it and builds it
	// again only when the data on screen changes.
	content *qt6.QWidget
	layout  *qt6.QVBoxLayout
	// controls are the interactive widgets in the reading order, so the tab
	// order is explicit and not the order of creation.
	controls []*qt6.QWidget
	// pendingOffset is the value of the scroll bar to give back after a render,
	// 0 when the bar needs no attention.
	pendingOffset int

	// shown is the data the panel draws now, and visible says whether it draws
	// a Project at all.
	shown   detailData
	visible bool
}

// newDetailPanel builds the panel at rest: the invitation to choose a Project,
// and the scroll area that will hold the detail.
func newDetailPanel(theme designTheme) *detailPanel {
	p := &detailPanel{theme: theme}

	p.invite = newDetailInvite()

	p.content = qt6.NewQWidget2()
	p.layout = qt6.NewQVBoxLayout(p.content)
	p.layout.SetContentsMargins(spaceL, spaceL, spaceL, spaceL)
	p.layout.SetSpacing(spaceL)

	p.scroll = qt6.NewQScrollArea2()
	p.scroll.SetWidgetResizable(true)
	p.scroll.SetFrameShape(qt6.QFrame__NoFrame)
	// The scroll area is a container, not a control: the focus stays on the
	// controls of the blocks.
	p.scroll.SetFocusPolicy(qt6.NoFocus)
	p.scroll.SetAccessibleName("Dettaglio del Progetto")
	p.scroll.SetWidget(p.content)

	p.stack = qt6.NewQStackedWidget2()
	p.stack.SetAccessibleName("Dettaglio del Progetto")
	p.stack.AddWidget(p.invite)
	p.stack.AddWidget(p.scroll.QWidget)
	p.stack.SetCurrentWidget(p.invite)
	p.widget = p.stack.QWidget

	// The rebuilt blocks change the range of the scroll bar, and the new range
	// is what gives the old value back: a render cannot set the value itself,
	// because the layout has not run yet.
	if bar := p.scroll.VerticalScrollBar(); bar != nil {
		bar.OnRangeChanged(p.keepScroll)
	}

	return p
}

// newDetailInvite is the panel at rest: the invitation to choose a Project. It
// is what the user sees before the first choice and after the open Project
// leaves the snapshot.
func newDetailInvite() *qt6.QWidget {
	invite := qt6.NewQWidget2()

	layout := qt6.NewQVBoxLayout(invite)
	layout.SetContentsMargins(0, 0, 0, 0)
	layout.SetSpacing(spaceS)
	layout.AddStretch()

	title := qt6.NewQLabel3("Scegli un Progetto")
	setObjectName(title.QWidget, "placeholderTitle")
	title.SetAlignment(qt6.AlignCenter)
	layout.AddWidget(title.QWidget)

	body := qt6.NewQLabel3("Seleziona un Progetto dalla lista per vederne i dettagli.")
	setObjectName(body.QWidget, "placeholderBody")
	body.SetAlignment(qt6.AlignCenter)
	body.SetWordWrap(true)
	layout.AddWidget(body.QWidget)

	layout.AddStretch()
	return invite
}

// set shows a Project in the state of Docker the snapshot reports. A snapshot
// that draws the same panel is a no-op: the panel keeps its blocks, the value
// of the scroll bar and the focus, so a refresh never moves the page under the
// eyes of the user.
func (p *detailPanel) set(project compose.Project, dockerReachable, loading bool) {
	data := detailData{project: project, dockerReachable: dockerReachable, loading: loading}
	if p.visible && p.shown.sameDraw(data) {
		return
	}
	p.visible = true
	p.shown = data
	p.stack.SetCurrentWidget(p.scroll.QWidget)
	p.render()
}

// clear goes back to the invitation, for when no Project is open or the open
// one left the snapshot.
func (p *detailPanel) clear() {
	if !p.visible {
		return
	}
	p.visible = false
	p.shown = detailData{}
	p.pendingOffset = 0
	p.controls = p.controls[:0]
	clearDetailLayout(p.layout.QLayout)
	p.stack.SetCurrentWidget(p.invite)
}

// linkAfter puts the controls of the panel after a control of the shell, so the
// tab chain of the window reaches the detail in the reading order.
func (p *detailPanel) linkAfter(prev *qt6.QWidget) {
	if prev == nil || len(p.controls) == 0 {
		return
	}
	qt6.QWidget_SetTabOrder(prev, p.controls[0])
}

// render builds the blocks of the detail from the data on screen. It reads the
// value of the scroll bar first, so the panel can give it back once the new
// range is known.
func (p *detailPanel) render() {
	offset := 0
	if bar := p.scroll.VerticalScrollBar(); bar != nil {
		offset = bar.Value()
	}

	clearDetailLayout(p.layout.QLayout)
	p.controls = p.controls[:0]

	p.layout.AddWidget(p.header())
	p.layout.AddWidget(p.data())
	if p.shown.project.ComposeMissing {
		p.layout.AddWidget(p.missing())
	}
	p.layout.AddWidget(p.actions())
	p.layout.AddWidget(p.commands())
	p.layout.AddWidget(p.services())
	p.layout.AddStretch()

	p.linkTabOrder()
	p.pendingOffset = offset
}

// keepScroll gives the scroll bar its value back when the rebuilt content has
// its new range. The range drops to nothing while the old blocks go, so a range
// that cannot hold the value leaves it pending: the range of the new blocks
// arrives after and takes it. A content that shrank under the old value keeps
// the view at its end, which is where the user already is.
func (p *detailPanel) keepScroll(_, maximum int) {
	if !p.visible || p.pendingOffset == 0 || maximum < p.pendingOffset {
		return
	}
	offset := p.pendingOffset
	p.pendingOffset = 0
	if bar := p.scroll.VerticalScrollBar(); bar != nil {
		bar.SetValue(offset)
	}
}

// linkTabOrder puts the controls of the detail in the reading order. The tab
// order of a window follows creation, and the blocks are built out of order.
func (p *detailPanel) linkTabOrder() {
	for i := 1; i < len(p.controls); i++ {
		qt6.QWidget_SetTabOrder(p.controls[i-1], p.controls[i])
	}
}

// header is the head of the detail: the name of the Project, the compose name
// when it differs, the state with its glyph and the tacks. When Docker does not
// answer, the notice takes the place of the numbers: the count and the tacks of
// the last snapshot would read as the current ones.
func (p *detailPanel) header() *qt6.QWidget {
	project := p.shown.project
	block, box := detailBlock()
	block.SetAccessibleName("Intestazione del Progetto")

	box.AddWidget(detailLabel(ProjectName(project), "detailTitle").QWidget)
	if project.DisplayName != "" && project.DisplayName != project.Name {
		box.AddWidget(detailLabel(project.Name, "detailComposeName").QWidget)
	}

	line := qt6.NewQHBoxLayout2()
	line.SetSpacing(spaceXS)

	status := StatusOf(project.Status)
	role := stateRole(status.CSSClass)

	marker := detailLabel(status.Glyph, role)
	marker.SetMinimumWidth(14)
	marker.SetAccessibleName(status.Label)
	line.AddWidget3(marker.QWidget, 0, qt6.AlignVCenter)
	line.AddWidget3(detailLabel(status.Label, role).QWidget, 0, qt6.AlignVCenter)

	notice := DockerNotice(p.shown.loading, p.shown.dockerReachable)
	if notice == "" {
		line.AddWidget3(detailLabel(ContainerCountLine(project), "detailCount").QWidget, 0, qt6.AlignVCenter)

		tacks, more := Tacks(project.Services)
		line.AddWidget3(newTackStrip(tacks, p.theme).QWidget, 0, qt6.AlignVCenter)
		if more > 0 {
			moreLabel := detailLabel(fmt.Sprintf("+%d", more), "tacksMore")
			line.AddWidget3(moreLabel.QWidget, 0, qt6.AlignVCenter)
		}
	}
	line.AddStretch()
	box.AddLayout2(line.QLayout, 0)

	// The notice takes the place of the numbers on a line of its own: on the
	// line of the state it would read as part of the state.
	if notice != "" {
		label := detailLabel(notice, "dockerNotice")
		label.SetWordWrap(true)
		box.AddWidget(label.QWidget)
	}

	return block
}

// data is the block of the data of the Project: the Group, the active Profiles
// and the Compose files. A line without a value stays hidden, so the block
// never shows a fact with nothing after it.
func (p *detailPanel) data() *qt6.QWidget {
	project := p.shown.project
	block, box := detailBlock()
	block.SetAccessibleName("Dati del Progetto")

	box.AddWidget(detailLabel(GroupLine(project.Group), "detailData").QWidget)

	if line := ActiveProfilesLine(project.ActiveProfiles); line != "" {
		box.AddWidget(detailLabel(line, "detailData").QWidget)
	}

	if line := ComposeFilesLine(project.ComposeFiles); line != "" {
		files := detailLabel(line, "detailData")
		// The line names the files by their base name; the tooltip carries the
		// whole paths, so the block keeps one line per fact.
		files.SetToolTip(strings.Join(project.ComposeFiles, "\n"))
		box.AddWidget(files.QWidget)
	}

	return block
}

// missing is the notice of a Project whose Compose files are gone, with the
// entry that takes it off the list. The entry is present and insensitive: the
// ticket that asks the confirmation and reaches the Manager enables it.
func (p *detailPanel) missing() *qt6.QWidget {
	block, box := detailBlock()
	block.SetAccessibleName("File compose mancante")

	box.AddWidget(detailLabel(MissingComposeText, "missingComposeNotice").QWidget)

	row := qt6.NewQHBoxLayout2()
	row.SetSpacing(spaceS)
	reason := RemoveDisabledReason(p.shown.project, p.shown.dockerReachable)
	remove := detailButton(removeText, "removeButton", reason)
	row.AddWidget(remove.QWidget)
	row.AddStretch()
	box.AddLayout2(row.QLayout, 0)
	p.controls = append(p.controls, remove.QWidget)

	return block
}

// actions is the row of the Actions of the Project. The buttons stay
// insensitive until the ticket that reaches the Manager. The reason of an
// insensitive Action, when there is one, sits beside the row: the tooltip alone
// would hide it from the keyboard and from the screen reader.
func (p *detailPanel) actions() *qt6.QWidget {
	project := p.shown.project
	block, box := detailBlock()
	block.SetAccessibleName("Azioni del Progetto")

	reason := actionReasonText(p.shown.loading, p.shown.dockerReachable, project.Activity != nil)

	row := qt6.NewQHBoxLayout2()
	row.SetSpacing(spaceS)
	for _, view := range ProjectActions(project, p.shown.dockerReachable) {
		button := detailButton(view.Label, "actionButton", reason)
		if reason != "" {
			button.SetAccessibleDescription(reason)
		}
		row.AddWidget(button.QWidget)
		p.controls = append(p.controls, button.QWidget)
	}
	row.AddStretch()
	box.AddLayout2(row.QLayout, 0)

	if reason != "" {
		box.AddWidget(detailLabel(reason, "actionsReason").QWidget)
	}

	return block
}

// commands is the block of the commands of the Project: its folder, a terminal
// in it and its first published port. They stay insensitive until the ticket
// that opens them.
func (p *detailPanel) commands() *qt6.QWidget {
	block, box := detailBlock()
	block.SetAccessibleName("Comandi del Progetto")

	row := qt6.NewQHBoxLayout2()
	row.SetSpacing(spaceS)
	for _, command := range []struct {
		text    string
		name    string
		tooltip string
	}{
		{folderCommandText, "folderButton", folderTooltip},
		{terminalCommandText, "terminalButton", terminalTooltip},
		{portCommandText, "portButton", portTooltip},
	} {
		button := detailButton(command.text, command.name, command.tooltip)
		row.AddWidget(button.QWidget)
		p.controls = append(p.controls, button.QWidget)
	}
	row.AddStretch()
	box.AddLayout2(row.QLayout, 0)

	return block
}

// services is the block of the Services: the heading with how many there are,
// then one block per Service. A Project without Service says why.
func (p *detailPanel) services() *qt6.QWidget {
	project := p.shown.project
	block, box := detailBlock()
	block.SetAccessibleName("Servizi del Progetto")

	heading := qt6.NewQHBoxLayout2()
	heading.SetSpacing(spaceS)
	heading.AddWidget3(detailLabel(servicesTitle, "detailHeading").QWidget, 0, qt6.AlignVCenter)
	if count := len(project.Services); count > 0 {
		number := detailLabel(strconv.Itoa(count), "detailCount")
		heading.AddWidget3(number.QWidget, 0, qt6.AlignVCenter)
	}
	heading.AddStretch()
	box.AddLayout2(heading.QLayout, 0)

	if len(project.Services) == 0 {
		box.AddWidget(detailLabel(NoServicesText(project), "detailEmpty").QWidget)
		return block
	}

	for _, service := range project.Services {
		box.AddWidget(p.service(service))
	}
	return block
}

// service is one block of the detail: the name, the state of its Containers,
// their tacks, the place of the Utilizzo risorse and one row per Container. The
// place of the indicator is reserved and empty: the indicator arrives in a
// later ticket and must not change the height of the block.
func (p *detailPanel) service(service compose.Service) *qt6.QWidget {
	block, box := detailBlock()
	block.SetAccessibleName(service.Name)

	line := qt6.NewQHBoxLayout2()
	line.SetSpacing(spaceS)
	line.AddWidget3(detailLabel(service.Name, "serviceName").QWidget, 0, qt6.AlignVCenter)
	line.AddWidget3(detailLabel(ServiceStateLine(service), "serviceState").QWidget, 0, qt6.AlignVCenter)
	line.AddStretch()

	tacks, more := Tacks([]compose.Service{service})
	line.AddWidget3(newTackStrip(tacks, p.theme).QWidget, 0, qt6.AlignVCenter)
	if more > 0 {
		moreLabel := detailLabel(fmt.Sprintf("+%d", more), "tacksMore")
		line.AddWidget3(moreLabel.QWidget, 0, qt6.AlignVCenter)
	}
	box.AddLayout2(line.QLayout, 0)

	usage := qt6.NewQWidget2()
	setObjectName(usage, "serviceUsage")
	usage.SetAccessibleName("Utilizzo risorse di " + service.Name)
	usage.SetFocusPolicy(qt6.NoFocus)
	usage.SetFixedHeight(serviceUsageHeight)
	box.AddWidget(usage)

	for _, container := range service.Containers {
		box.AddWidget(newContainerRow(container))
	}

	return block
}

// newContainerRow is one Container: the glyph and the colour of its tone, its
// name, its state, its health and its published ports. The state is never only
// a colour: the glyph, the word and the health label say it too.
func newContainerRow(container compose.Container) *qt6.QWidget {
	row := qt6.NewQWidget2()

	line := qt6.NewQHBoxLayout(row)
	line.SetContentsMargins(spaceXS, 0, 0, 0)
	line.SetSpacing(spaceS)

	tone := ToneOfContainer(container)
	marker := detailLabel(tone.Glyph, stateRole(tone.CSSClass))
	marker.SetMinimumWidth(14)
	marker.SetAccessibleName(container.State)
	line.AddWidget3(marker.QWidget, 0, qt6.AlignVCenter)

	line.AddWidget3(detailLabel(container.Name, "containerName").QWidget, 1, qt6.AlignVCenter)
	line.AddWidget3(detailLabel(container.State, "containerState").QWidget, 0, qt6.AlignVCenter)

	if health := HealthOf(container.Health); health.Text != "" {
		label := detailLabel(health.Text, "containerHealth")
		if role := stateRole(health.CSSClass); role != "" {
			setObjectName(label.QWidget, role)
		}
		line.AddWidget3(label.QWidget, 0, qt6.AlignVCenter)
	}

	for _, port := range container.Ports {
		line.AddWidget3(detailLabel(PortText(port), "containerPort").QWidget, 0, qt6.AlignVCenter)
	}

	return row
}

// detailBlock is one block of the detail: a widget with a vertical layout with
// no margin, so the spacing of the page is the only space between blocks.
func detailBlock() (*qt6.QWidget, *qt6.QVBoxLayout) {
	block := qt6.NewQWidget2()
	box := qt6.NewQVBoxLayout(block)
	box.SetContentsMargins(0, 0, 0, 0)
	box.SetSpacing(spaceS)
	return block, box
}

// detailLabel builds a label of the detail. It wraps, so a long name or a long
// notice needs no fixed width. A role of the style sheet colours it; an empty
// role keeps the colour of the theme.
func detailLabel(text, role string) *qt6.QLabel {
	label := qt6.NewQLabel3(text)
	label.SetWordWrap(true)
	if role != "" {
		setObjectName(label.QWidget, role)
	}
	return label
}

// detailButton builds a command of the detail. Every command of this ticket is
// present and insensitive: the ticket that wires it enables it, so the row
// keeps its shape and its place now.
func detailButton(text, name, tooltip string) *qt6.QPushButton {
	button := qt6.NewQPushButton3(text)
	setObjectName(button.QWidget, name)
	button.SetAccessibleName(text)
	if tooltip != "" {
		button.SetToolTip(tooltip)
	}
	button.SetEnabled(false)
	keepTarget(button.QWidget)
	return button
}

// clearDetailLayout empties a layout and destroys its widgets and its items, so
// no widget and no item holds the data of a Project the user left. The layout
// of a block belongs to the block, so it goes with it.
func clearDetailLayout(layout *qt6.QLayout) {
	for layout.Count() > 0 {
		item := layout.TakeAt(0)
		if item == nil {
			return
		}
		if widget := item.Widget(); widget != nil {
			widget.Delete()
		}
		// TakeAt hands the item over: without this the item of every block of
		// every render stays allocated.
		item.Delete()
	}
}
