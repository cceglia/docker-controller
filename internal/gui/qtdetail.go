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

// panelControl is one command of the detail: its widget and the key that names
// it. The key finds the command again after a render replaced its widget, so a
// closed question gives the focus back to the same command. A command of a
// Service is keyed "servizio/Azione", so it does not collide with the "Down" of
// the Project.
type panelControl struct {
	widget *qt6.QWidget
	label  string
}

// detailPanel is the panel of the right: the Project the user chose, its data,
// its Services and the rows of the Actions and of the commands. It reads the
// snapshot and never changes a Project. The Actions of the Project and of a
// Service reach the Manager through onAction, which the shell supplies. The
// commands and the entry that takes a Project off the list stay insensitive
// until the tickets that open them.
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
	controls []panelControl
	// pendingOffset is the value of the scroll bar to give back after a render,
	// 0 when the bar needs no attention.
	pendingOffset int

	// onAction hands an Action to the shell, which reaches the Manager off the
	// Qt thread. A nil function launches nothing.
	onAction func(compose.Action)
	// onSave hands the values of the form to the shell, which reaches the
	// Manager off the Qt thread. A nil function closes the form at once.
	onSave func(string, compose.Settings)
	// projects are the Projects of the last snapshot: the Group field proposes
	// the Groups they use.
	projects []compose.Project
	// edit is the form of the settings on screen, nil while the detail shows the
	// data of the Project. A snapshot that arrives while it is open waits, so the
	// form keeps what the user typed.
	edit *editForm
	// onLogs asks the shell to show the logs of a Project, whole or of one of
	// its Services. A nil function opens nothing.
	onLogs func(compose.Project, string)
	// guard stops a second click between a launch and its result.
	guard startGuard
	// launching is the command the last click disabled. A render destroys the
	// widget, so it clears this before building the blocks again.
	launching *qt6.QPushButton
	// question is the confirmation on screen, nil when there is none. While it
	// is open the panel keeps the commands the question belongs to, and a
	// snapshot that arrives draws the panel when the question closes.
	question    *confirmDialog
	waitingDraw bool

	// shown is the data the panel draws now, and visible says whether it draws
	// a Project at all.
	shown   detailData
	visible bool

	// strip is the Utilizzo risorse of the machine, above the detail: it is the
	// machine's, so it stays out of the blocks a Project rebuilds.
	strip *usageStrip
	// usage is the sample the panel draws, nil until the first one arrives and
	// after the stream stops. usageRows holds the indicator of each Service of
	// the Project on screen, by "project/service": a render builds them again
	// from the sample it kept.
	usage     *compose.UsageSample
	usageRows map[string]*serviceUsageRow
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

	// The bar of the machine sits above the stack, so it keeps its place while
	// the detail scrolls. It stays out of the blocks: a render of a Project
	// never rebuilds a figure of the machine.
	p.strip = newUsageStrip()
	p.strip.widget.SetVisible(false)
	p.usageRows = make(map[string]*serviceUsageRow)

	p.widget = qt6.NewQWidget2()
	column := qt6.NewQVBoxLayout(p.widget)
	column.SetContentsMargins(0, 0, 0, 0)
	column.SetSpacing(0)
	column.AddWidget(p.strip.widget)
	column.AddWidget2(p.stack.QWidget, 1)

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
	p.guard.snapshot(project.Activity != nil)
	p.stack.SetCurrentWidget(p.scroll.QWidget)
	// The bar of the machine belongs to the detail: with no Project open there
	// is nothing to measure it for.
	p.strip.widget.SetVisible(true)
	if p.question != nil || p.edit != nil {
		// The question belongs to the command on screen and the form holds what
		// the user typed: the panel keeps both, then draws the data that arrived.
		p.waitingDraw = true
		return
	}
	p.render()
}

// clear goes back to the invitation, for when no Project is open or the open
// one left the snapshot. A question still on screen goes away with the Project.
func (p *detailPanel) clear() {
	if !p.visible {
		return
	}
	p.visible = false
	p.shown = detailData{}
	p.pendingOffset = 0
	p.waitingDraw = false
	p.controls = p.controls[:0]
	p.launching = nil
	// The form belongs to the Project on screen: it goes with it, so a render of
	// the next Project never builds it again.
	p.edit = nil
	// The figures of a Project that left the screen go with it: the stream
	// stops, and nothing may stay behind as if it were current.
	p.usage = nil
	p.usageRows = make(map[string]*serviceUsageRow)
	p.strip.set(nil)
	p.strip.widget.SetVisible(false)
	clearDetailLayout(p.layout.QLayout)
	p.stack.SetCurrentWidget(p.invite)
	if p.question != nil {
		p.question.dialog.Reject()
	}
}

// linkAfter puts the controls of the panel after a control of the shell, so the
// tab chain of the window reaches the detail in the reading order.
func (p *detailPanel) linkAfter(prev *qt6.QWidget) {
	if prev == nil || len(p.controls) == 0 {
		return
	}
	qt6.QWidget_SetTabOrder(prev, p.controls[0].widget)
}

// lastControl is the control the tab chain leaves the detail from, nil while
// the panel shows nothing.
func (p *detailPanel) lastControl() *qt6.QWidget {
	if len(p.controls) == 0 {
		return nil
	}
	return p.controls[len(p.controls)-1].widget
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
	p.launching = nil
	p.waitingDraw = false
	// The blocks are new, so their indicators are new too: the sample the panel
	// kept comes back on them at once, without waiting for the next frame.
	p.usageRows = make(map[string]*serviceUsageRow)

	p.layout.AddWidget(p.header())
	// The form of the settings takes the place of the data of the Project: the
	// rest of the detail stays where the user left it.
	if p.edit != nil {
		p.layout.AddWidget(p.editBlock())
	} else {
		p.layout.AddWidget(p.data())
	}
	if p.shown.project.ComposeMissing {
		p.layout.AddWidget(p.missing())
	}
	p.layout.AddWidget(p.actions())
	p.layout.AddWidget(p.commands())
	p.layout.AddWidget(p.services())
	p.layout.AddStretch()

	p.applyUsage()
	p.showUsage()
	p.linkTabOrder()
	p.pendingOffset = offset
}

// setUsage shows one sample of the Utilizzo risorse. A nil sample is a stream
// that stopped: the bar goes back to its waiting line and the indicators of the
// Services leave the page, so an old figure never stands as the current one.
func (p *detailPanel) setUsage(sample *compose.UsageSample) {
	p.usage = sample
	p.showUsage()
	p.applyUsage()
}

// showUsage puts the bar in the state the panel knows: a Docker that does not
// answer shows the warning in place of the figures, whatever the last sample
// said, and a sample otherwise says what it saw at its instant.
func (p *detailPanel) showUsage() {
	if !p.shown.loading && !p.shown.dockerReachable {
		p.strip.setUnreachable()
		return
	}
	p.strip.set(p.usage)
}

// applyUsage puts the sample on screen: the four measures of every Service the
// sample carries. A Service without an entry, without a running Container or
// without figures shows no indicator at all, and a Docker that does not answer
// takes every indicator away.
func (p *detailPanel) applyUsage() {
	unreachable := !p.shown.loading && !p.shown.dockerReachable
	byService := make(map[string]compose.ServiceUsage, len(p.usageRows))
	if p.usage != nil {
		for _, usage := range p.usage.Services {
			byService[usage.Project+"/"+usage.Service] = usage
		}
	}
	for key, row := range p.usageRows {
		entry, ok := byService[key]
		state := UsageNone
		if !unreachable {
			state = UsageStateOf(p.usage, ok, row.running)
		}
		row.set(state, entry)
	}
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
		qt6.QWidget_SetTabOrder(p.controls[i-1].widget, p.controls[i].widget)
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

	// The logs of the whole Project sit under the head, where the old view put
	// them: one press opens the panel on the tab of the logs.
	row := qt6.NewQHBoxLayout2()
	row.SetSpacing(spaceS)
	logs := detailButton(showLogsText, "logButton", "Mostra i log del Progetto.", true)
	logs.OnClicked(func() {
		if p.onLogs != nil {
			p.onLogs(project, "")
		}
	})
	row.AddWidget(logs.QWidget)
	row.AddStretch()
	box.AddLayout2(row.QLayout, 0)
	p.controls = append(p.controls, panelControl{widget: logs.QWidget, label: showLogsText})

	return block
}

// data is the block of the data of the Project: the Group, the active Profiles
// and the Compose files, with the command that opens the form of the settings. A
// line without a value stays hidden, so the block never shows a fact with
// nothing after it.
func (p *detailPanel) data() *qt6.QWidget {
	project := p.shown.project
	block, box := detailBlock()
	block.SetAccessibleName(dataTitle)

	heading := qt6.NewQHBoxLayout2()
	heading.SetSpacing(spaceS)
	heading.AddWidget3(detailLabel(dataTitle, "detailHeading").QWidget, 0, qt6.AlignVCenter)
	heading.AddStretch()
	edit := detailButton(editText, "editButton", "Modifica Nome visualizzato, Gruppo e Profili attivi.", true)
	edit.OnClicked(func() { p.startEdit() })
	heading.AddWidget3(edit.QWidget, 0, qt6.AlignVCenter)
	box.AddLayout2(heading.QLayout, 0)
	p.controls = append(p.controls, panelControl{widget: edit.QWidget, label: editText})

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
	remove := detailButton(removeText, "removeButton", reason, false)
	row.AddWidget(remove.QWidget)
	row.AddStretch()
	box.AddLayout2(row.QLayout, 0)
	p.controls = append(p.controls, panelControl{widget: remove.QWidget, label: removeText})

	return block
}

// actions is the row of the Actions of the Project. A destructive Action asks
// its question first; the others reach the shell at once. The reason of an
// insensitive Action, when there is one, sits beside the row: the tooltip alone
// would hide it from the keyboard and from the screen reader.
func (p *detailPanel) actions() *qt6.QWidget {
	project := p.shown.project
	block, box := detailBlock()
	block.SetAccessibleName("Azioni del Progetto")

	reason := actionReasonText(p.shown.loading, p.shown.dockerReachable, project.Activity != nil)

	row := qt6.NewQHBoxLayout2()
	row.SetSpacing(spaceS)
	for i, view := range ProjectActions(project, p.shown.dockerReachable) {
		// The Action the state suggests comes first: it carries the role that
		// makes it the one the eye finds, without a colour of its own.
		name := "actionButton"
		if i == 0 {
			name = "primaryAction"
		}
		button := detailButton(view.Label, name, "", view.Enabled)
		if reason != "" {
			button.SetAccessibleDescription(reason)
			button.SetToolTip(reason)
		}
		button.OnClicked(func() { p.choose(view, button, view.Label) })
		row.AddWidget(button.QWidget)
		p.controls = append(p.controls, panelControl{widget: button.QWidget, label: view.Label})
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
		button := detailButton(command.text, command.name, command.tooltip, false)
		row.AddWidget(button.QWidget)
		p.controls = append(p.controls, panelControl{widget: button.QWidget, label: command.text})
	}
	row.AddStretch()
	box.AddLayout2(row.QLayout, 0)

	return block
}

// services is the block of the Services: the heading with how many there are,
// then one block per Service. A Project without Service says why. The reason of
// an insensitive Action sits on the commands as description and tooltip: the
// row of the Actions of the Project already writes it in full, and the detail
// does not repeat it for every Service.
func (p *detailPanel) services() *qt6.QWidget {
	project := p.shown.project
	block, box := detailBlock()
	block.SetAccessibleName("Servizi del Progetto")

	reason := actionReasonText(p.shown.loading, p.shown.dockerReachable, project.Activity != nil)

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

	for i, service := range project.Services {
		if i > 0 {
			box.AddWidget(detailSeparator())
		}
		box.AddWidget(p.service(service, reason))
	}
	return block
}

// detailSeparator is the line between two blocks of the detail: two Services,
// and the Utilizzo risorse of the machine above them. A QFrame draws it with the
// colours of the desktop palette: a plain widget would not paint the background
// of the style sheet, and a colour of the sheet reads as the window itself.
func detailSeparator() *qt6.QWidget {
	line := qt6.NewQFrame2()
	line.SetFrameShape(qt6.QFrame__HLine)
	line.SetFrameShadow(qt6.QFrame__Sunken)
	line.SetFocusPolicy(qt6.NoFocus)
	line.SetAttribute2(qt6.WA_TransparentForMouseEvents, true)
	return line.QWidget
}

// service is one block of the detail: the name, the state of its Containers,
// their tacks, the Actions of the Service, one row per Container and the place
// of the Utilizzo risorse. The indicator of the Utilizzo risorse arrives with
// the first sample and takes no room before it: a Service without a measure, and
// a Service without a Container, show no indicator.
func (p *detailPanel) service(service compose.Service, reason string) *qt6.QWidget {
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

	// The logs of this Service alone, next to its name: the panel follows one
	// Container's output instead of the whole Project.
	logs := detailButton(serviceLogText, "serviceLogButton", "Mostra i log del Servizio.", true)
	logs.SetAccessibleName(serviceLogText + " di " + service.Name)
	logs.OnClicked(func() {
		if p.onLogs != nil {
			p.onLogs(p.shown.project, service.Name)
		}
	})
	line.AddWidget3(logs.QWidget, 0, qt6.AlignVCenter)
	p.controls = append(p.controls, panelControl{widget: logs.QWidget, label: service.Name + "/" + serviceLogText})
	box.AddLayout2(line.QLayout, 0)

	for _, container := range service.Containers {
		box.AddWidget(newContainerRow(container))
	}

	// The Utilizzo risorse of the Service sits under its Containers, where the
	// old view put it. It takes no room until a sample carries the Service.
	usage := newServiceUsageRow(p.shown.project.Name, service.Name, ServiceRuns(service))
	p.usageRows[usage.key] = usage
	box.AddWidget(usage.widget)

	// The Actions close the block, so the Containers stay under their own
	// Service and the row of the next Service starts after a divider.
	box.AddLayout2(p.serviceActions(service, reason).QLayout, 0)

	return block
}

// serviceActions is the row of the Actions of one Service: Start, Stop, Restart
// and Down. Down asks its question first, and every Action follows the same
// rules of the Actions of the Project.
func (p *detailPanel) serviceActions(service compose.Service, reason string) *qt6.QHBoxLayout {
	row := qt6.NewQHBoxLayout2()
	row.SetSpacing(spaceS)
	for _, view := range ServiceActions(p.shown.project, service, p.shown.dockerReachable) {
		button := detailButton(view.Label, "serviceActionButton", "", view.Enabled)
		button.SetAccessibleName(view.Label + " di " + service.Name)
		if reason != "" {
			button.SetAccessibleDescription(reason)
			button.SetToolTip(reason)
		}
		key := service.Name + "/" + view.Label
		button.OnClicked(func() { p.choose(view, button, key) })
		row.AddWidget(button.QWidget)
		p.controls = append(p.controls, panelControl{widget: button.QWidget, label: key})
	}
	row.AddStretch()
	return row
}

// choose runs an Action of the detail. A destructive Action asks its question
// first, and an insensitive one does nothing, so a stale button never starts a
// command.
func (p *detailPanel) choose(view ActionView, button *qt6.QPushButton, key string) {
	if !view.Enabled {
		return
	}
	if view.Confirmation == nil {
		p.launch(view, button)
		return
	}
	p.ask(view, button, key)
}

// ask opens the question of a destructive Action. Annulling, Escape and closing
// run nothing. When the question closes, the panel draws a snapshot that waited
// and gives the focus back to the command that asked.
func (p *detailPanel) ask(view ActionView, button *qt6.QPushButton, key string) {
	if p.question != nil {
		return
	}
	question := newConfirmDialog(p.parent(), *view.Confirmation, func() { p.launch(view, button) })
	question.onClosed = func() {
		p.question = nil
		p.settle()
		p.focusControl(key)
	}
	p.question = question
	question.open()
}

// launch hands the Action to the shell. The command that launched it stays
// insensitive until the result comes back, and the guard refuses a second
// click, so one click starts one Activity.
func (p *detailPanel) launch(view ActionView, button *qt6.QPushButton) {
	if !p.guard.begin() {
		return
	}
	if button != nil {
		p.launching = button
		button.SetEnabled(false)
	}
	if p.onAction == nil {
		p.guard.release()
		p.commandLaunched(true)
		return
	}
	p.onAction(view.Action)
}

// actionDone ends a launch: the shell reached the Manager, so the guard opens
// again and the command the click disabled comes back. A refusal carries its
// own message, which the shell shows in the notice.
func (p *detailPanel) actionDone() {
	p.guard.release()
	p.commandLaunched(true)
}

// commandLaunched gives the command of the last launch its state back. A render
// replaced the widget and the new one comes from the data, so there is nothing
// to restore.
func (p *detailPanel) commandLaunched(enabled bool) {
	if p.launching == nil {
		return
	}
	p.launching.SetEnabled(enabled)
	p.launching = nil
}

// settle draws a snapshot that arrived while a question was open. A form of the
// settings keeps its place: closing it draws the data that arrived, so a render
// never takes the form away with the text the user typed.
func (p *detailPanel) settle() {
	if !p.visible || !p.waitingDraw || p.edit != nil {
		return
	}
	p.render()
}

// focusControl gives the focus back to a command of the detail by its key: a
// question closes and the user goes on from where it started. A command that
// the new snapshot disabled cannot take the focus, so the first command that
// can takes it.
func (p *detailPanel) focusControl(key string) {
	var first *qt6.QWidget
	for _, control := range p.controls {
		if !control.widget.IsEnabled() {
			continue
		}
		if control.label == key {
			control.widget.SetFocus()
			return
		}
		if first == nil {
			first = control.widget
		}
	}
	if first != nil {
		first.SetFocus()
	}
}

// parent is the window a question opens over.
func (p *detailPanel) parent() *qt6.QWidget {
	return p.widget.Window()
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

// detailButton builds a command of the detail. enabled says whether the command
// can be pressed: a command of a later ticket starts off. Every command meets
// the target; a command of a Service sets its own accessible name, so it reads
// as "Start di web".
func detailButton(text, name, tooltip string, enabled bool) *qt6.QPushButton {
	button := qt6.NewQPushButton3(text)
	setObjectName(button.QWidget, name)
	button.SetAccessibleName(text)
	if tooltip != "" {
		button.SetToolTip(tooltip)
	}
	button.SetEnabled(enabled)
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
