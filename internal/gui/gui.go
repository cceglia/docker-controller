package gui

import (
	"context"
	"fmt"
	"os"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/cairo"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"

	"dockercontroller/internal/compose"
)

const (
	// appID names the application, its desktop entry and its notifications.
	appID = "io.github.cceglia.DockerController"
	// windowTitle is the title of the window.
	windowTitle = "Docker Controller"
	// windowWidth and windowHeight are the size the window opens at.
	windowWidth  = 1120
	windowHeight = 760
	// breakpointWidth is the width under which the window shows one page at a
	// time, in scale independent pixels.
	breakpointWidth = 900
	// emptyListText is the line of the list when no Project is saved.
	emptyListText = "Nessun Progetto."
)

// Run opens the window and blocks until the application quits. The snapshot of
// the Manager is read through Watch, so a Project added or removed appears
// without a restart.
func Run(ctx context.Context, manager *compose.Manager) error {
	app := adw.NewApplication(appID, gio.ApplicationDefaultFlags)
	var window *shell
	app.ConnectActivate(func() {
		if window != nil {
			window.window.Present()
			return
		}
		window = newShell(app)
		window.watch(ctx, manager)
		window.window.Present()
	})
	app.Run(os.Args)
	return nil
}

// shell owns the window and the two sides of its split view.
type shell struct {
	window   *adw.ApplicationWindow
	split    *adw.NavigationSplitView
	list     *gtk.ListBox
	content  *adw.NavigationPage
	names    map[*gtk.ListBoxRow]string
	selected string
}

// newShell builds the window: the list on the left, the content on the right,
// and the breakpoint that collapses the two into pages under 900 sp.
func newShell(app *adw.Application) *shell {
	s := &shell{names: map[*gtk.ListBoxRow]string{}}

	s.list = gtk.NewListBox()
	s.list.SetSelectionMode(gtk.SelectionSingle)
	s.list.SetShowSeparators(false)
	s.list.AddCSSClass("navigation-sidebar")
	s.list.ConnectRowActivated(s.activateRow)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetChild(s.list)
	scrolled.SetVExpand(true)

	sidebar := adw.NewToolbarView()
	sidebar.AddTopBar(sidebarBar())
	sidebar.SetContent(scrolled)
	sidebar.AddBottomBar(sidebarActions())

	s.content = contentPage()

	s.split = adw.NewNavigationSplitView()
	s.split.SetSidebar(adw.NewNavigationPage(sidebar, "Progetti"))
	s.split.SetContent(s.content)
	s.split.SetShowContent(false)
	s.split.SetMinSidebarWidth(260)
	s.split.SetMaxSidebarWidth(360)

	breakpoint := adw.NewBreakpoint(
		adw.NewBreakpointConditionLength(adw.BreakpointConditionMaxWidth, breakpointWidth, adw.LengthUnitSp),
	)
	breakpoint.AddSetter(s.split, "collapsed", true)

	s.window = adw.NewApplicationWindow(&app.Application)
	s.window.SetTitle(windowTitle)
	s.window.SetDefaultSize(windowWidth, windowHeight)
	s.window.AddBreakpoint(breakpoint)
	s.window.SetContent(s.split)
	return s
}

// sidebarBar is the header of the list. It carries only the page title; the
// Project detail will add its own actions in a later ticket.
func sidebarBar() *adw.HeaderBar {
	bar := adw.NewHeaderBar()
	title := gtk.NewLabel("Progetti")
	title.AddCSSClass("heading")
	bar.SetTitleWidget(title)
	return bar
}

// sidebarActions is the bar under the list. Both buttons stay visible with an
// empty list and stay insensitive until their ticket wires them.
func sidebarActions() *gtk.ActionBar {
	bar := gtk.NewActionBar()

	add := gtk.NewButtonWithLabel("Aggiungi progetto")
	add.SetSensitive(false)
	add.SetTooltipText("Aggiungi un Progetto alla lista.")
	bar.PackStart(add)

	settings := gtk.NewButtonWithLabel("Impostazioni")
	settings.SetSensitive(false)
	settings.SetTooltipText("Apri le Impostazioni.")
	bar.PackEnd(settings)

	return bar
}

// contentPage is the right side of the split view. Its header bar carries the
// window controls and, under the breakpoint, the back button to the list.
func contentPage() *adw.NavigationPage {
	view := adw.NewToolbarView()
	view.AddTopBar(adw.NewHeaderBar())
	view.SetContent(contentPlaceholder())
	return adw.NewNavigationPage(view, windowTitle)
}

// contentPlaceholder invites the user to choose a Project. The detail of the
// chosen one arrives in a later ticket.
func contentPlaceholder() *adw.StatusPage {
	page := adw.NewStatusPage()
	page.SetIconName("view-list-symbolic")
	page.SetTitle("Scegli un Progetto")
	page.SetDescription("Seleziona un Progetto dalla lista per vederne i dettagli.")
	return page
}

// watch follows the Manager and moves every snapshot onto the GTK loop.
func (s *shell) watch(ctx context.Context, manager *compose.Manager) {
	go func() {
		for snapshot := range manager.Watch(ctx) {
			glib.IdleAdd(func() { s.apply(snapshot) })
		}
	}()
}

// apply rebuilds the list from a snapshot, keeping the open Project selected.
func (s *shell) apply(snapshot compose.Snapshot) {
	s.names = map[*gtk.ListBoxRow]string{}
	s.list.RemoveAll()

	var selected *gtk.ListBoxRow
	for _, section := range Sections(snapshot.Projects) {
		if len(section.Projects) == 0 {
			continue
		}
		if section.Title != "" {
			s.appendHeading(section.Title)
		}
		for _, project := range section.Projects {
			row := s.appendProject(project)
			if project.Name == s.selected {
				selected = row
			}
		}
	}

	if len(snapshot.Projects) == 0 && !snapshot.Loading {
		s.appendEmpty()
	}

	if selected != nil {
		s.list.SelectRow(selected)
		return
	}
	// The open Project is gone: the content side has nothing to show.
	s.selected = ""
	s.split.SetShowContent(false)
}

// activateRow opens the content side of the split view on the chosen Project.
func (s *shell) activateRow(row *gtk.ListBoxRow) {
	name, ok := s.names[row]
	if !ok {
		return
	}
	s.selected = name
	s.split.SetShowContent(true)
}

// appendHeading adds a Group heading, which the list cannot select.
func (s *shell) appendHeading(title string) {
	row := gtk.NewListBoxRow()
	row.SetSelectable(false)
	row.SetActivatable(false)

	label := gtk.NewLabel(title)
	label.SetXAlign(0)
	label.AddCSSClass("caption-heading")
	label.SetMarginTop(10)
	label.SetMarginBottom(4)
	label.SetMarginStart(12)
	row.SetChild(label)

	s.list.Append(row)
}

// appendEmpty tells the empty list apart from a broken one. The line is not
// selectable and not activatable, like a Group heading.
func (s *shell) appendEmpty() {
	row := gtk.NewListBoxRow()
	row.SetSelectable(false)
	row.SetActivatable(false)

	label := gtk.NewLabel(emptyListText)
	label.SetXAlign(0)
	label.AddCSSClass("dim-label")
	label.SetMarginTop(10)
	label.SetMarginBottom(4)
	label.SetMarginStart(12)
	row.SetChild(label)

	s.list.Append(row)
}

// appendProject adds the row of one Project and returns it.
func (s *shell) appendProject(project compose.Project) *gtk.ListBoxRow {
	status := StatusOf(project.Status)

	row := gtk.NewListBoxRow()

	box := gtk.NewBox(gtk.OrientationHorizontal, 10)
	box.SetMarginTop(8)
	box.SetMarginBottom(8)
	box.SetMarginStart(12)
	box.SetMarginEnd(12)

	marker := gtk.NewLabel(status.Glyph)
	marker.SetVAlign(gtk.AlignCenter)
	marker.AddCSSClass(status.CSSClass)
	marker.SetTooltipText(status.Label)
	marker.UpdateProperty(
		[]gtk.AccessibleProperty{gtk.AccessiblePropertyLabel},
		[]glib.Value{*glib.NewValue(status.Label)},
	)
	box.Append(marker)

	text := gtk.NewBox(gtk.OrientationVertical, 0)
	text.SetHExpand(true)
	name := gtk.NewLabel(ProjectName(project))
	name.SetXAlign(0)
	name.SetEllipsize(pango.EllipsizeEnd)
	text.Append(name)
	if project.DisplayName != "" && project.DisplayName != project.Name {
		composeName := gtk.NewLabel(project.Name)
		composeName.SetXAlign(0)
		composeName.SetEllipsize(pango.EllipsizeEnd)
		composeName.AddCSSClass("dim-label")
		composeName.AddCSSClass("caption")
		text.Append(composeName)
	}
	box.Append(text)

	count := gtk.NewLabel(ProjectCount(project))
	count.AddCSSClass("dim-label")
	count.AddCSSClass("numeric")
	count.SetVAlign(gtk.AlignCenter)
	box.Append(count)

	tacks, more := Tacks(project.Services)
	strip := gtk.NewBox(gtk.OrientationHorizontal, 2)
	strip.SetVAlign(gtk.AlignCenter)
	strip.Append(tackArea(tacks))
	if more > 0 {
		moreLabel := gtk.NewLabel(fmt.Sprintf("+%d", more))
		moreLabel.AddCSSClass("dim-label")
		moreLabel.AddCSSClass("caption")
		strip.Append(moreLabel)
	}
	box.Append(strip)

	row.SetChild(box)
	row.SetTooltipText(fmt.Sprintf("%s, %s, %d Container in esecuzione su %d",
		ProjectName(project), status.Label, project.Running, project.Total))
	s.names[row] = project.Name
	s.list.Append(row)
	return row
}

const (
	// tackWidth, tackHeight and tackGap size the tacks of the strip.
	tackWidth  = 3
	tackHeight = 10
	tackGap    = 2
)

// tackArea draws the tacks with the colours of the theme: the foreground for a
// running Container, a faint one for a stopped Container and the error colour
// for a faulty one.
func tackArea(states []TackState) *gtk.DrawingArea {
	width := 0
	if len(states) > 0 {
		width = len(states)*(tackWidth+tackGap) - tackGap
	}

	area := gtk.NewDrawingArea()
	area.SetContentWidth(width)
	area.SetContentHeight(tackHeight)
	area.SetSizeRequest(width, tackHeight)
	area.SetVAlign(gtk.AlignCenter)

	area.SetDrawFunc(func(area *gtk.DrawingArea, cr *cairo.Context, _, height int) {
		context := area.StyleContext()
		foreground := context.Color()
		faulty, hasFaulty := context.LookupColor("error_color")

		y := float64(height-tackHeight) / 2
		for i, state := range states {
			switch {
			case state == TackFaulty && hasFaulty:
				cr.SetSourceRGBA(float64(faulty.Red()), float64(faulty.Green()), float64(faulty.Blue()), 1)
			case state == TackRunning:
				cr.SetSourceRGBA(float64(foreground.Red()), float64(foreground.Green()), float64(foreground.Blue()), 1)
			default:
				cr.SetSourceRGBA(float64(foreground.Red()), float64(foreground.Green()), float64(foreground.Blue()), 0.25)
			}
			cr.Rectangle(float64(i*(tackWidth+tackGap)), y, tackWidth, tackHeight)
			cr.Fill()
		}
	})
	return area
}
