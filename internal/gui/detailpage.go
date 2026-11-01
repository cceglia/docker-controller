package gui

import (
	"fmt"
	"reflect"
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
	panel *activityPanel
	// shown is the Project the detail is built from, shownDocker and
	// shownLoading the state of Docker it was built for, and visible says
	// whether the detail is on screen at all.
	shown        compose.Project
	shownDocker  bool
	shownLoading bool
	visible      bool
	// onAction launches an Action; parent hosts the confirmation dialog. GTK
	// must keep the last dialog alive while it is on screen.
	onAction func(compose.Action)
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

	d.panel = newActivityPanel(activities)
	d.paned = gtk.NewPaned(gtk.OrientationVertical)
	d.paned.SetStartChild(d.stack)
	d.paned.SetEndChild(d.panel.body)
	d.paned.SetResizeStartChild(true)
	d.paned.SetShrinkStartChild(true)
	d.paned.SetShrinkEndChild(true)
	d.panel.paned = d.paned

	view := adw.NewToolbarView()
	view.AddTopBar(adw.NewHeaderBar())
	view.SetContent(d.paned)

	d.page = adw.NewNavigationPage(view, windowTitle)
	return d
}

// set shows one Project in the state of Docker the snapshot reports. A
// snapshot that changes neither is a no-op, so the detail keeps its place while
// the user reads it.
func (d *detailView) set(project compose.Project, dockerReachable, loading bool) {
	if d.visible && d.shownDocker == dockerReachable && d.shownLoading == loading &&
		reflect.DeepEqual(d.shown, project) {
		return
	}
	d.visible = true
	d.shown = project
	d.shownDocker = dockerReachable
	d.shownLoading = loading
	d.page.SetTitle(ProjectName(project))
	d.stack.SetVisibleChildName("detail")

	reason := ActionDisabledReason(dockerReachable, project.Activity != nil)

	offset := d.scrolled.VAdjustment().Value()
	removeAll(d.content)
	if notice := DockerNotice(loading, dockerReachable); notice != "" {
		d.content.Append(noticeBanner(notice))
	}
	d.content.Append(projectHeader(project, actionBar(ProjectActions(project, dockerReachable), reason, true, d.choose)))
	if project.ComposeMissing {
		d.content.Append(missingBanner())
	}
	d.content.Append(servicesHeading(len(project.Services)))
	if len(project.Services) == 0 {
		d.content.Append(emptyServices(project))
	}
	for _, service := range project.Services {
		d.content.Append(serviceGroup(project, service, dockerReachable, d.choose))
	}
	// The new rows are laid out after this call, so the scroll comes back
	// when the main loop is idle.
	scrolled := d.scrolled
	glib.IdleAdd(func() { scrolled.VAdjustment().SetValue(offset) })
}

// clear goes back to the invitation, for when the open Project is gone.
func (d *detailView) clear() {
	d.visible = false
	removeAll(d.content)
	d.page.SetTitle(windowTitle)
	d.stack.SetVisibleChildName("empty")
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
// active Profiles.
func projectHeader(project compose.Project, actions *gtk.Box) *gtk.Box {
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

// serviceGroup is one block of the detail: the Actions of a Service, its name
// and one row per Container, in the order the Project reports them.
func serviceGroup(project compose.Project, service compose.Service, dockerReachable bool, choose func(ActionView)) *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle(service.Name)
	reason := ActionDisabledReason(dockerReachable, project.Activity != nil)
	group.SetHeaderSuffix(actionBar(ServiceActions(project, service, dockerReachable), reason, false, choose))
	for _, container := range service.Containers {
		group.Add(containerRow(container))
	}
	return group
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
