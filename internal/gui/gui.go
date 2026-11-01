package gui

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/cairo"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/settings"
	"dockercontroller/internal/tray"
	"dockercontroller/internal/trayadapter"
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
	// toastTimeout is how long a message stays on screen, in seconds.
	toastTimeout = 6
	// showOutputText is the button of the toast of a failure: it opens the
	// output of the failed Activity.
	showOutputText = "Vedi output"
)

// Run opens the window and blocks until the application quits. The snapshot of
// the Manager is read through Watch, so a Project added or removed appears
// without a restart. activities is the log of the session: the bottom panel
// reads it and the Manager pushes its events into it. store holds the settings:
// the terminal command of "Apri terminale" comes from it. notifier shows the
// desktop notifications the Manager's events and the tray Actions deserve.
func Run(ctx context.Context, manager *compose.Manager, activities *ActivityLog, store *settings.Store, notifier *Notifier) error {
	app := adw.NewApplication(appID, gio.ApplicationDefaultFlags)
	var window *shell
	app.ConnectActivate(func() {
		if window != nil {
			window.window.Present()
			return
		}
		window = newShell(app, activities, store, notifier)
		window.watch(ctx, manager)
		window.window.Present()
	})
	app.Run(os.Args)
	return nil
}

// shell owns the window and the two sides of its split view.
type shell struct {
	// app is the application: "Esci" quits it, so the loop returns.
	app     *adw.Application
	window  *adw.ApplicationWindow
	toasts  *adw.ToastOverlay
	split   *adw.NavigationSplitView
	list    *gtk.ListBox
	content *adw.NavigationPage
	detail  *detailView
	// tray is the icon in the system tray, nil when it could not start.
	tray *trayadapter.Tray
	// confirmations are the questions of the tray that are on screen.
	confirmations []*adw.MessageDialog
	// manager runs the Actions and ctx stops them when the application quits.
	manager *compose.Manager
	ctx     context.Context
	// store holds the settings: the terminal command of "Apri terminale".
	store *settings.Store
	// notifier shows the desktop notifications of the Activities and of the
	// tray Actions the Manager refuses.
	notifier *Notifier
	// snapshot is the last one applied: the tray builds its menu from it.
	snapshot compose.Snapshot
	// add is the dialog of a new Project while it is on screen, nil otherwise.
	add *addDialog
	// rows maps the index of a row of the list to its Project. GTK hands the
	// handler a wrapper of the row, not the pointer it was built with, so the
	// pointer cannot be the key.
	rows     map[int]compose.Project
	selected string
	// projects is the whole snapshot: the form of the settings reads the Groups
	// in use from it, so a Group of another Project is one click away.
	projects []compose.Project
	// reachable and loading are the state of Docker of the last snapshot, so
	// the detail keeps it while no new snapshot arrives.
	reachable bool
	loading   bool
	// usage runs `docker stats` while the window is visible. usageEpoch changes
	// with every start and stop, so a sample still in flight is dropped and
	// never puts old figures back on screen.
	usage      *UsageStreamer
	usageEpoch atomic.Uint64
}

// newShell builds the window: the list on the left, the content on the right,
// and the breakpoint that collapses the two into pages under 900 sp.
func newShell(app *adw.Application, activities *ActivityLog, store *settings.Store, notifier *Notifier) *shell {
	// Until the first snapshot arrives Docker has not been asked yet, so the
	// detail must not call it unreachable and the tray must not show the notice.
	s := &shell{
		app:      app,
		rows:     map[int]compose.Project{},
		loading:  true,
		store:    store,
		notifier: notifier,
		snapshot: compose.Snapshot{Loading: true},
	}

	s.list = gtk.NewListBox()
	s.list.SetSelectionMode(gtk.SelectionSingle)
	s.list.SetShowSeparators(false)
	s.list.SetActivateOnSingleClick(true)
	s.list.AddCSSClass("navigation-sidebar")
	s.list.ConnectRowActivated(s.activateRow)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetChild(s.list)
	scrolled.SetVExpand(true)

	sidebar := adw.NewToolbarView()
	sidebar.AddTopBar(sidebarBar())
	sidebar.SetContent(scrolled)
	sidebar.AddBottomBar(s.sidebarActions())

	s.detail = newDetailView(activities)
	s.detail.onAction = s.do
	s.detail.onSave = s.saveSettings
	s.detail.onFolder = s.openFolder
	s.detail.onTerminal = s.openTerminal
	s.detail.onRemove = s.removeFromList
	s.detail.onPort = s.openPort
	s.detail.panel.onFailure = s.activityFailed
	s.content = s.detail.page

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

	s.toasts = adw.NewToastOverlay()
	s.toasts.SetChild(s.split)

	s.window = adw.NewApplicationWindow(&app.Application)
	s.window.SetTitle(windowTitle)
	s.window.SetDefaultSize(windowWidth, windowHeight)
	s.window.AddBreakpoint(breakpoint)
	s.window.SetContent(s.toasts)
	s.detail.parent = &s.window.ApplicationWindow.Window
	// A window that goes away must not leave `docker compose logs` or
	// `docker stats` running.
	s.window.ConnectDestroy(func() {
		s.detail.panel.logs.StopAll()
		s.stopUsage()
	})
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

// sidebarActions is the bar under the list. The button of a new Project asks for
// its File compose; Impostazioni stays insensitive until its ticket wires it.
func (s *shell) sidebarActions() *gtk.ActionBar {
	bar := gtk.NewActionBar()

	add := gtk.NewButtonWithLabel("Aggiungi progetto")
	add.SetTooltipText("Aggiungi un Progetto alla lista.")
	add.ConnectClicked(s.showAddDialog)
	bar.PackStart(add)

	prefs := gtk.NewButtonWithLabel("Impostazioni")
	prefs.SetSensitive(false)
	prefs.SetTooltipText("Apri le Impostazioni.")
	bar.PackEnd(prefs)

	return bar
}

// contentPlaceholder invites the user to choose a Project, until one is open.
func contentPlaceholder() *adw.StatusPage {
	page := adw.NewStatusPage()
	page.SetIconName("view-list-symbolic")
	page.SetTitle("Scegli un Progetto")
	page.SetDescription("Seleziona un Progetto dalla lista per vederne i dettagli.")
	return page
}

// watch follows the Manager and moves every snapshot onto the GTK loop.
func (s *shell) watch(ctx context.Context, manager *compose.Manager) {
	s.ctx = ctx
	s.manager = manager
	// The Log page streams through the Manager, and stops with the application.
	s.detail.panel.logs.source = manager.Logs
	s.detail.panel.logs.ctx = ctx
	s.watchUsage(manager)
	// The tray shows the state before the first snapshot, so the icon is on the
	// panel as soon as the window is.
	s.startTray()
	go func() {
		for snapshot := range manager.Watch(ctx) {
			glib.IdleAdd(func() { s.apply(snapshot) })
		}
	}()
}

// watchUsage wires the Utilizzo risorse: the stream runs while the window is
// visible and stops when the window hides, so no `docker stats` outlives the
// figures on screen.
func (s *shell) watchUsage(manager *compose.Manager) {
	s.usage = NewUsageStreamer(s.ctx, manager.Usage)
	s.usage.OnSample = func(sample compose.UsageSample) {
		epoch := s.usageEpoch.Load()
		glib.IdleAdd(func() {
			if s.usageEpoch.Load() != epoch {
				return // a sample of a stream the window stopped
			}
			s.detail.setUsage(&sample)
		})
	}
	s.usage.OnStop = func() {
		epoch := s.usageEpoch.Load()
		glib.IdleAdd(func() {
			if s.usageEpoch.Load() == epoch {
				s.detail.setUsage(nil)
			}
		})
	}
	s.usage.OnError = func(error) {
		glib.IdleAdd(func() { s.detail.setUsage(nil) })
	}
	s.window.NotifyProperty("visible", s.visibilityChanged)
}

// visibilityChanged starts the stream when the window shows and stops it when
// the window hides in the tray.
func (s *shell) visibilityChanged() {
	if s.window.Visible() {
		s.startUsage()
		return
	}
	s.stopUsage()
}

// startUsage begins the stream, once.
func (s *shell) startUsage() {
	if s.usage == nil {
		return
	}
	s.usageEpoch.Add(1)
	s.usage.Start()
}

// stopUsage ends the stream and clears the figures at once: a stopped stream
// leaves no stale number on screen.
func (s *shell) stopUsage() {
	if s.usage == nil {
		return
	}
	s.usageEpoch.Add(1)
	s.usage.Stop()
	s.detail.setUsage(nil)
}

// do launches an Action on the Manager and tells the user why it did not start.
// The outcome of the Activity belongs to a later ticket.
func (s *shell) do(action compose.Action) {
	if s.manager == nil {
		return
	}
	if _, err := s.manager.Do(s.ctx, action); err != nil {
		s.toast(UserError(err).Error())
	}
}

// saveSettings writes the settings of a Project and publishes the new snapshot,
// so the detail and the list refresh by themselves. The write runs off the GTK
// loop; a refusal comes back as the Italian message. The form already went back
// to rest, so a refusal leaves the saved values on screen.
func (s *shell) saveSettings(name string, settings compose.Settings) {
	if s.manager == nil {
		return
	}
	go func() {
		if err := s.manager.UpdateSettings(s.ctx, name, settings); err != nil {
			glib.IdleAdd(func() { s.toast(UserError(err).Error()) })
		}
	}()
}

// toast shows a short message over the content.
func (s *shell) toast(message string) {
	toast := adw.NewToast(message)
	toast.SetTimeout(toastTimeout)
	s.toasts.AddToast(toast)
}

// activityFailed tells the user a failed Activity with the Italian message and
// a link to its output, as the old view did.
func (s *shell) activityFailed(a Activity) {
	toast := adw.NewToast(ActivityToastText(a))
	toast.SetTimeout(toastTimeout)
	toast.SetButtonLabel(showOutputText)
	toast.ConnectButtonClicked(func() { s.detail.panel.Show(a.ID) })
	s.toasts.AddToast(toast)
}

// apply rebuilds the list from a snapshot, keeping the open Project selected
// and refreshing its detail in place.
func (s *shell) apply(snapshot compose.Snapshot) {
	s.snapshot = snapshot
	// A notification names a Project as the list does, and the snapshot is the
	// last place that tells the Display name.
	s.notifier.SetProjects(snapshot.Projects)
	if s.tray != nil {
		s.tray.Show(tray.BuildMenu(snapshot), tray.IconFor(snapshot))
	}
	s.reachable = snapshot.DockerReachable
	s.loading = snapshot.Loading
	s.projects = snapshot.Projects
	s.rows = map[int]compose.Project{}
	s.list.RemoveAll()

	var selected *gtk.ListBoxRow
	var opened compose.Project
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
				opened = project
			}
		}
	}

	if len(snapshot.Projects) == 0 && !snapshot.Loading {
		s.appendEmpty()
	}

	if selected != nil {
		s.list.SelectRow(selected)
		s.detail.set(opened, s.reachable, s.loading, GroupChoices(opened.Group, s.projects))
		return
	}
	// The open Project is gone: the content side has nothing to show.
	s.selected = ""
	s.detail.clear()
	s.split.SetShowContent(false)
}

// activateRow opens the content side of the split view on the chosen Project.
// A single click is enough: the list activates on one, like the old view.
func (s *shell) activateRow(row *gtk.ListBoxRow) {
	// GTK hands the handler a wrapper of the row, not the pointer it was
	// built with, so the row is found by its place in the list.
	project, ok := s.rows[row.Index()]
	if !ok {
		return
	}
	s.selected = project.Name
	s.detail.set(project, s.reachable, s.loading, GroupChoices(project.Group, s.projects))
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
	s.list.Append(row)
	s.rows[row.Index()] = project
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
