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
// Project and its Services. It rebuilds from the snapshot it is given, so it
// never holds a copy that is out of date.
type detailView struct {
	page     *adw.NavigationPage
	stack    *gtk.Stack
	scrolled *gtk.ScrolledWindow
	content  *gtk.Box
	// shown is the Project the detail is built from, and visible says whether
	// the detail is on screen at all.
	shown   compose.Project
	visible bool
}

// newDetailView builds the content side: the invitation until a Project is
// chosen, and the detail in its place after.
func newDetailView() *detailView {
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

	view := adw.NewToolbarView()
	view.AddTopBar(adw.NewHeaderBar())
	view.SetContent(d.stack)

	d.page = adw.NewNavigationPage(view, windowTitle)
	return d
}

// set shows one Project. A snapshot that does not change it is a no-op, so the
// detail keeps its place while the user reads it.
func (d *detailView) set(project compose.Project) {
	if d.visible && reflect.DeepEqual(d.shown, project) {
		return
	}
	d.visible = true
	d.shown = project
	d.page.SetTitle(ProjectName(project))
	d.stack.SetVisibleChildName("detail")

	offset := d.scrolled.VAdjustment().Value()
	removeAll(d.content)
	d.content.Append(projectHeader(project))
	if project.ComposeMissing {
		d.content.Append(missingBanner())
	}
	d.content.Append(servicesHeading(len(project.Services)))
	if len(project.Services) == 0 {
		d.content.Append(emptyServices(project))
	}
	for _, service := range project.Services {
		d.content.Append(serviceGroup(service))
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

// removeAll empties a box. GTK4 removes one child at a time, so it takes the
// first child until none is left.
func removeAll(box *gtk.Box) {
	for child := box.FirstChild(); child != nil; child = box.FirstChild() {
		box.Remove(child)
	}
}

// projectHeader is the head of the detail: the name of the Project, its status
// with the count of its Containers, its Compose files and its active Profiles.
func projectHeader(project compose.Project) *gtk.Box {
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

// missingBanner flags the Project whose Compose files are gone. It carries no
// button: Rimuovi dalla lista belongs to a later ticket.
func missingBanner() *adw.Banner {
	banner := adw.NewBanner(MissingComposeText)
	banner.SetRevealed(true)
	return banner
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

// serviceGroup is one block of the detail: the name of a Service and one row
// per Container, in the order the Project reports them.
func serviceGroup(service compose.Service) *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle(service.Name)
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
