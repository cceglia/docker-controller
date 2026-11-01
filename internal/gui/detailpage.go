package gui

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"dockercontroller/internal/compose"
)

const (
	// detailWidth is the width the content side clamps the detail to, so the
	// lines stay readable in a wide window.
	detailWidth = 700
	// detailSpacing is the space between the blocks of the detail.
	detailSpacing = 18
	// showLogsText is the button of the head that opens the Log page.
	showLogsText = "Mostra log"
)

// detailView is the content side of the split view: the header of the chosen
// Project, its Actions and its Services, with the bottom panel of the
// Activities under them. It rebuilds from the snapshot it is given, so it never
// holds a copy that is out of date.
type detailView struct {
	page     *adw.NavigationPage
	stack    *gtk.Stack
	scrolled *gtk.ScrolledWindow
	content  *gtk.Box
	// paned splits the detail from the bottom panel; panel is the panel itself.
	paned *gtk.Paned
	panel *bottomPanel
	// strip is the global Utilizzo risorse above the detail; blocks holds the
	// Utilizzo risorse of each Service, by "project/service", and usage is the
	// sample on screen, nil until one arrives and after a stop.
	strip  *usageStrip
	blocks map[string]*serviceUsageBlock
	usage  *compose.UsageSample
	// shown is the Project the detail is built from, shownDocker and
	// shownLoading the state of Docker it was built for, and visible says
	// whether the detail is on screen at all.
	shown        compose.Project
	shownDocker  bool
	shownLoading bool
	visible      bool
	// edit is the form of the settings, nil when the detail is at rest. It
	// holds only what the user typed, so the snapshot stays the truth of the
	// saved values. groups are the Groups the form offers, from the whole
	// snapshot.
	edit   *ProjectEdit
	groups []string
	// onAction launches an Action; onSave hands the settings of the form to the
	// Manager. parent hosts the confirmation dialog. GTK must keep the last
	// dialog alive while it is on screen.
	onAction func(compose.Action)
	onSave   func(name string, settings compose.Settings)
	parent   *gtk.Window
	dialog   *adw.MessageDialog
}

// newDetailView builds the content side: the invitation until a Project is
// chosen, the detail in its place after, and the bottom panel of the Activities
// under both. activities is the log of the session, so the panel shows the
// history of every Project.
func newDetailView(activities *ActivityLog) *detailView {
	d := &detailView{content: gtk.NewBox(gtk.OrientationVertical, detailSpacing)}
	d.content.SetMarginTop(18)
	d.content.SetMarginBottom(24)
	d.content.SetMarginStart(18)
	d.content.SetMarginEnd(18)

	clamp := adw.NewClamp()
	clamp.SetMaximumSize(detailWidth)
	clamp.SetTighteningThreshold(detailWidth / 2)
	clamp.SetChild(d.content)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetChild(clamp)
	scrolled.SetVExpand(true)
	d.scrolled = scrolled

	d.stack = gtk.NewStack()
	d.stack.SetTransitionType(gtk.StackTransitionTypeCrossfade)
	d.stack.AddNamed(contentPlaceholder(), "empty")
	d.stack.AddNamed(scrolled, "detail")
	d.stack.SetVisibleChildName("empty")

	d.panel = newBottomPanel(activities)
	d.paned = gtk.NewPaned(gtk.OrientationVertical)
	d.paned.SetStartChild(d.stack)
	d.paned.SetEndChild(d.panel.body)
	d.paned.SetResizeStartChild(true)
	d.paned.SetShrinkStartChild(true)
	d.paned.SetShrinkEndChild(true)
	d.paned.SetVExpand(true)
	d.panel.paned = d.paned

	// The global Utilizzo risorse stays one quiet row above the detail, under
	// the header bar of the page.
	d.strip = newUsageStrip()
	body := gtk.NewBox(gtk.OrientationVertical, 0)
	body.Append(d.strip.widget)
	body.Append(d.paned)

	view := adw.NewToolbarView()
	view.AddTopBar(adw.NewHeaderBar())
	view.SetContent(body)

	d.page = adw.NewNavigationPage(view, windowTitle)
	return d
}

// set shows one Project in the state of Docker the snapshot reports. A
// snapshot that changes none of them is a no-op, so the detail keeps its place
// while the user reads it and the form keeps what the user typed. groups are the
// Groups of the whole snapshot, which the form of the settings offers.
func (d *detailView) set(project compose.Project, dockerReachable, loading bool, groups []string) {
	if d.visible && d.shownDocker == dockerReachable && d.shownLoading == loading &&
		reflect.DeepEqual(d.shown, project) && slices.Equal(d.groups, groups) {
		return
	}
	// Another Project: the stream of the one before must not write into it, and
	// the form of the one before must not save on it.
	if d.visible && d.shown.Name != project.Name {
		d.panel.closeLogs()
		d.edit = nil
	}
	d.visible = true
	d.shown = project
	d.shownDocker = dockerReachable
	d.shownLoading = loading
	d.groups = groups
	d.page.SetTitle(ProjectName(project))
	d.stack.SetVisibleChildName("detail")
	// The form of the settings keeps what the user typed and its place: a
	// snapshot that arrives while the user edits waits for the save or the
	// cancel, which render the latest state.
	if d.edit != nil {
		return
	}
	d.render()
}

// render builds the content of the detail from the Project it holds. The offset
// of the scroll comes back, so a refresh does not move the text under the eyes
// of the user.
func (d *detailView) render() {
	project := d.shown
	reason := ActionDisabledReason(d.shownDocker, project.Activity != nil)

	offset := d.scrolled.VAdjustment().Value()
	removeAll(d.content)
	if notice := DockerNotice(d.shownLoading, d.shownDocker); notice != "" {
		d.content.Append(noticeBanner(notice))
	}
	d.content.Append(projectHeader(project, actionBar(ProjectActions(project, d.shownDocker), reason, true, d.choose), func() { d.showLogs(project) }))
	if project.ComposeMissing {
		d.content.Append(missingBanner())
	}
	if d.edit != nil {
		d.content.Append(projectForm(project, d.groups, d.edit, d.saveEdit, d.cancelEdit))
	} else {
		d.content.Append(projectData(project, func() { d.startEdit(project) }))
		if profiles := projectProfiles(project); profiles != nil {
			d.content.Append(profiles)
		}
	}
	d.content.Append(servicesHeading(len(project.Services)))
	if len(project.Services) == 0 {
		d.content.Append(emptyServices(project))
	}
	d.blocks = make(map[string]*serviceUsageBlock, len(project.Services))
	for _, service := range project.Services {
		group, block := serviceGroup(project, service, d.shownDocker, d.choose)
		d.content.Append(group)
		d.blocks[project.Name+"/"+service.Name] = block
	}
	d.applyUsage()
	// The new rows are laid out after this call, so the scroll comes back
	// when the main loop is idle.
	scrolled := d.scrolled
	glib.IdleAdd(func() { scrolled.VAdjustment().SetValue(offset) })
}

// startEdit opens the form on the saved values of the Project. The list and the
// header keep the saved values until the form is confirmed.
func (d *detailView) startEdit(project compose.Project) {
	edit := NewProjectEdit(project)
	d.edit = &edit
	d.render()
}

// cancelEdit leaves the form and shows the saved values again. Nothing was sent
// to the Manager, so there is nothing to undo.
func (d *detailView) cancelEdit() {
	if d.edit == nil {
		return
	}
	d.edit = nil
	d.render()
}

// saveEdit hands the values of the form to the Manager and goes back to rest.
// A refusal leaves the saved values on screen and shows its message, so a save
// that did not happen is never shown as done.
func (d *detailView) saveEdit() {
	if d.edit == nil {
		return
	}
	name, settings := d.edit.Project, d.edit.Settings()
	d.edit = nil
	d.render()
	if d.onSave != nil {
		d.onSave(name, settings)
	}
}

// setUsage shows a sample in the global bar and in the blocks of the Services.
// A nil sample is a stopped stream: every figure goes away at once.
func (d *detailView) setUsage(sample *compose.UsageSample) {
	d.usage = sample
	d.applyUsage()
}

// applyUsage puts the sample on screen: on the strip and on each Service block.
// A Service the sample does not list shows the dash.
func (d *detailView) applyUsage() {
	d.strip.set(d.usage)
	byService := map[string]compose.ServiceUsage{}
	if d.usage != nil {
		for _, usage := range d.usage.Services {
			byService[usage.Project+"/"+usage.Service] = usage
		}
	}
	for key, block := range d.blocks {
		usage, ok := byService[key]
		if !ok {
			block.set(d.usage, nil)
			continue
		}
		block.set(d.usage, &usage)
	}
}

// clear goes back to the invitation, for when the open Project is gone.
func (d *detailView) clear() {
	d.visible = false
	d.edit = nil
	d.panel.closeLogs()
	d.blocks = nil
	removeAll(d.content)
	d.page.SetTitle(windowTitle)
	d.stack.SetVisibleChildName("empty")
}

// showLogs opens the Log page on the logs of the Project, whole or of one of
// its Services.
func (d *detailView) showLogs(project compose.Project) {
	d.panel.ShowLogs(project, "")
}

// choose runs an Action, or asks for its confirmation first. A disabled Action
// does nothing, so a stale button never starts a command.
func (d *detailView) choose(view ActionView) {
	if !view.Enabled {
		return
	}
	if view.Confirmation == nil {
		d.launch(view.Action)
		return
	}
	d.confirm(view)
}

// confirm asks the question of a destructive Action. Cancelling or closing the
// dialog launches nothing. When the question names a text to type, the confirm
// response stays off until the entry holds exactly that text.
func (d *detailView) confirm(view ActionView) {
	question := view.Confirmation
	dialog := adw.NewMessageDialog(d.parent, question.Title, question.Message)
	d.dialog = dialog
	dialog.AddResponse("cancel", "Annulla")
	dialog.AddResponse("run", question.ConfirmLabel)
	dialog.SetCloseResponse("cancel")
	dialog.SetDefaultResponse("cancel")
	if question.Destructive {
		dialog.SetResponseAppearance("run", adw.ResponseDestructive)
	}
	if question.TypeToConfirm != "" {
		hint := gtk.NewLabel("Per confermare digita " + question.TypeToConfirm)
		hint.SetXAlign(0)
		hint.AddCSSClass("dim-label")

		entry := gtk.NewEntry()
		entry.SetPlaceholderText(question.TypeToConfirm)
		entry.ConnectChanged(func() {
			dialog.SetResponseEnabled("run", entry.Text() == question.TypeToConfirm)
		})

		child := gtk.NewBox(gtk.OrientationVertical, 6)
		child.Append(hint)
		child.Append(entry)
		dialog.SetExtraChild(child)
		dialog.SetResponseEnabled("run", false)
		dialog.SetFocus(entry)
	}
	dialog.ConnectResponse(func(response string) {
		if d.dialog == dialog {
			d.dialog = nil
		}
		if response == "run" {
			d.launch(view.Action)
		}
	})
	dialog.Present()
}

// launch hands the Action to the Manager through the shell.
func (d *detailView) launch(action compose.Action) {
	if d.onAction != nil {
		d.onAction(action)
	}
}

// removeAll empties a box. GTK4 removes one child at a time, so it takes the
// first child until none is left.
func removeAll(box *gtk.Box) {
	for child := box.FirstChild(); child != nil; child = box.FirstChild() {
		box.Remove(child)
	}
}

// projectHeader is the head of the detail: the name of the Project, its status
// with the count of its Containers, its Actions, its Compose files and its
// active Profiles. onLogs, when set, opens the Log page on the Project.
func projectHeader(project compose.Project, actions *gtk.Box, onLogs func()) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationVertical, 6)

	name := gtk.NewLabel(ProjectName(project))
	name.SetXAlign(0)
	name.SetWrap(true)
	name.AddCSSClass("title-1")
	box.Append(name)

	status := StatusOf(project.Status)
	line := gtk.NewBox(gtk.OrientationHorizontal, 6)
	marker := gtk.NewLabel(status.Glyph)
	marker.AddCSSClass(status.CSSClass)
	marker.SetTooltipText(status.Label)
	line.Append(marker)
	state := gtk.NewLabel(status.Label)
	line.Append(state)
	count := gtk.NewLabel(ContainerCountLine(project))
	count.AddCSSClass("dim-label")
	line.Append(count)
	box.Append(line)

	if onLogs != nil {
		logs := gtk.NewButtonWithLabel(showLogsText)
		logs.AddCSSClass("flat")
		logs.SetTooltipText("Mostra i log del Progetto.")
		logs.ConnectClicked(onLogs)
		actions.Append(logs)
	}
	box.Append(actions)

	if files := ComposeFilesLine(project.ComposeFiles); files != "" {
		label := gtk.NewLabel(files)
		label.SetXAlign(0)
		label.SetWrap(true)
		label.AddCSSClass("dim-label")
		label.SetTooltipText(strings.Join(project.ComposeFiles, "\n"))
		box.Append(label)
	}

	if profiles := ActiveProfilesLine(project.ActiveProfiles); profiles != "" {
		label := gtk.NewLabel(profiles)
		label.SetXAlign(0)
		label.SetWrap(true)
		label.AddCSSClass("dim-label")
		box.Append(label)
	}

	return box
}

// noticeBanner is the strip that tells why the view cannot act.
func noticeBanner(text string) *adw.Banner {
	banner := adw.NewBanner(text)
	banner.SetRevealed(true)
	return banner
}

// missingBanner flags the Project whose Compose files are gone. It carries no
// button: Rimuovi dalla lista belongs to a later ticket.
func missingBanner() *adw.Banner {
	return noticeBanner(MissingComposeText)
}

// actionBar is the row of Action buttons of a Project or of a Service. When
// accent is true the first button carries it, because the status suggests it;
// the others stay flat, so a destructive Action is discreet at rest.
func actionBar(views []ActionView, disabledReason string, accent bool, choose func(ActionView)) *gtk.Box {
	bar := gtk.NewBox(gtk.OrientationHorizontal, 6)
	bar.SetHAlign(gtk.AlignStart)
	for i, view := range views {
		button := gtk.NewButtonWithLabel(view.Label)
		button.SetSensitive(view.Enabled)
		if i == 0 && accent {
			button.AddCSSClass("suggested-action")
		} else {
			button.AddCSSClass("flat")
		}
		switch {
		case !view.Enabled && disabledReason != "":
			button.SetTooltipText(disabledReason)
		case view.Confirmation != nil:
			button.SetTooltipText(view.Confirmation.Message)
		}
		button.ConnectClicked(func() { choose(view) })
		bar.Append(button)
	}
	return bar
}

// servicesHeading is the heading of the list of Services, with how many there
// are.
func servicesHeading(count int) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationHorizontal, 6)
	title := gtk.NewLabel("Servizi")
	title.AddCSSClass("heading")
	box.Append(title)
	if count > 0 {
		number := gtk.NewLabel(strconv.Itoa(count))
		number.AddCSSClass("dim-label")
		box.Append(number)
	}
	return box
}

// emptyServices explains a Project without Service. A Project that has none is
// Rimosso, which is a state and not an error.
func emptyServices(project compose.Project) *gtk.Label {
	label := gtk.NewLabel(NoServicesText(project))
	label.SetXAlign(0)
	label.SetWrap(true)
	label.AddCSSClass("dim-label")
	return label
}

// serviceGroup is one block of the detail: the Actions of a Service, its name,
// its Utilizzo risorse and one row per Container, in the order the Project
// reports them. It returns the block of the measures, so the detail updates it
// without rebuilding the group.
func serviceGroup(project compose.Project, service compose.Service, dockerReachable bool, choose func(ActionView)) (*adw.PreferencesGroup, *serviceUsageBlock) {
	group := adw.NewPreferencesGroup()
	group.SetTitle(service.Name)
	reason := ActionDisabledReason(dockerReachable, project.Activity != nil)
	group.SetHeaderSuffix(actionBar(ServiceActions(project, service, dockerReachable), reason, false, choose))

	block := newServiceUsageBlock(ServiceRuns(service))
	row := gtk.NewListBoxRow()
	row.SetActivatable(false)
	row.SetSelectable(false)
	usageBox := gtk.NewBox(gtk.OrientationVertical, 0)
	usageBox.SetMarginTop(6)
	usageBox.SetMarginBottom(8)
	usageBox.SetMarginStart(12)
	usageBox.SetMarginEnd(12)
	usageBox.Append(block.widget)
	row.SetChild(usageBox)
	group.Add(row)

	for _, container := range service.Containers {
		group.Add(containerRow(container))
	}
	return group, block
}

// containerRow is one Container: the marker of its tone, its name, its image,
// its state, its health and its published ports.
func containerRow(container compose.Container) *adw.ActionRow {
	row := adw.NewActionRow()
	row.SetTitle(container.Name)
	row.SetSubtitle(container.Image)
	row.SetSubtitleLines(1)

	tone := ToneOfContainer(container)
	marker := gtk.NewLabel(tone.Glyph)
	marker.AddCSSClass(tone.CSSClass)
	marker.SetTooltipText(container.State)
	row.AddPrefix(marker)

	state := gtk.NewLabel(container.State)
	state.AddCSSClass("dim-label")
	state.AddCSSClass("caption")
	row.AddSuffix(state)

	if health := HealthOf(container.Health); health.Text != "" {
		label := gtk.NewLabel(health.Text)
		label.AddCSSClass("caption")
		if health.CSSClass == "" {
			label.AddCSSClass("dim-label")
		} else {
			label.AddCSSClass(health.CSSClass)
		}
		row.AddSuffix(label)
	}

	if len(container.Ports) > 0 {
		row.AddSuffix(portBox(container.Ports))
	}

	return row
}

// portBox lists the published ports of one Container. Opening a port in the
// browser belongs to a later ticket, so the ports are text only.
func portBox(ports []compose.Port) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationHorizontal, 6)
	box.SetVAlign(gtk.AlignCenter)
	for _, port := range ports {
		label := gtk.NewLabel(PortText(port))
		label.AddCSSClass("monospace")
		label.AddCSSClass("caption")
		label.SetTooltipText(fmt.Sprintf("Porta host %d, porta del Container %d, protocollo %s",
			port.Published, port.Target, port.Protocol))
		box.Append(label)
	}
	return box
}
