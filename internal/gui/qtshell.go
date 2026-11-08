package gui

import (
	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/activity"
	"dockercontroller/internal/gui/design"
	"dockercontroller/internal/gui/usage"
	"dockercontroller/internal/gui/view"
	"dockercontroller/internal/tray"
)

// shell is the window of the application: the column of the Projects on the
// left and the panel of the detail on the right, or one page at a time when the
// window is narrower than breakpointWidth. It reads the snapshot and reaches the
// Manager for the Actions, off the Qt thread. A refused Action, and later a
// failed Activity, reach the user through the notice.
type shell struct {
	app    *application
	window *qt6.QMainWindow
	// theme supplies the palette the style sheet and the painted roles use.
	theme design.DesignTheme

	// pages holds the wide layout and the two pages of the narrow one. Only one
	// page shows at a time: under the breakpoint the window shows the list or
	// the detail, never both.
	pages *qt6.QStackedWidget
	// widePage holds the splitter; narrowListPage and narrowDetailPage hold the
	// two columns alone. The columns move between the three, so the state of
	// the list and of the detail never needs a second copy.
	widePage           *qt6.QWidget
	narrowListPage     *qt6.QWidget
	narrowDetailPage   *qt6.QWidget
	narrowListLayout   *qt6.QVBoxLayout
	narrowDetailLayout *qt6.QVBoxLayout

	splitter *qt6.QSplitter
	sidebar  *qt6.QWidget
	detail   *qt6.QWidget
	// column is the vertical splitter of the detail and of the bottom panel: the
	// two share the right column, and the panel takes no room while it is closed.
	column *qt6.QSplitter

	list   *projectList
	panel  *detailPanel
	bottom *bottomPanel
	// usage runs the stream of the Utilizzo risorse: the figures of the machine,
	// the same for every Project.
	usage    *usage.UsageFeed
	add      *qt6.QPushButton
	settings *qt6.QPushButton
	back     *qt6.QPushButton
	// notice shows a message in sovrimpressione, off the commands. A refused
	// Action shows its Italian message there, a failed Activity the same with
	// the way to its output.
	notice *toast
	// addForm is the form of a new Project on screen, nil when it is closed.
	addForm *addDialog

	// selected is the compose name of the open Project, "" when none is open.
	selected string
	// snapshot is the last one applied: open reads the Project from it, so the
	// detail shows the row the user clicked.
	snapshot compose.Snapshot
	// narrow is true while the window shows one page at a time.
	narrow bool
}

// newShell builds the window. The list on the left marks the Project the user
// chooses; the panel on the right shows it.
func newShell(app *application, theme design.DesignTheme) *shell {
	s := &shell{app: app, theme: theme}

	s.window = qt6.NewQMainWindow2()
	// The window carries the icon of the app, so the title bar and the task
	// manager show it even when the desktop entry is not installed.
	s.window.SetWindowIcon(systemTrayIcon(tray.IconNormal))
	s.window.SetWindowTitle(windowTitle)
	s.window.SetAccessibleName(windowTitle)
	s.window.Resize(windowWidth, windowHeight)
	s.window.SetMinimumSize2(560, 420)

	s.buildSidebar()
	s.buildDetail()
	s.buildPages()
	s.notice = newToast(s.pages.QWidget)
	// The bar of the Utilizzo risorse takes the top of the detail: a message in
	// sovrimpressione goes under it and never covers a figure.
	s.notice.top = s.noticeTop

	s.window.SetCentralWidget(s.pages.QWidget)
	s.window.OnResizeEvent(func(super func(event *qt6.QResizeEvent), event *qt6.QResizeEvent) {
		super(event)
		s.applyWidth()
		s.notice.place()
	})
	// Closing the window is the choice of the settings: hide it in the tray, ask,
	// or quit. The window is never destroyed here, so the close is refused and
	// the answer runs with the window still holding its Project and its history.
	s.window.OnCloseEvent(func(_ func(event *qt6.QCloseEvent), event *qt6.QCloseEvent) {
		event.Ignore()
		if s.app != nil {
			s.app.closeRequested()
		}
	})

	// The tab order follows the reading order: the list, the actions of the
	// column, then the way back of the narrow view. The panel adds its own
	// controls after the entry the width chooses.
	qt6.QWidget_SetTabOrder(s.list.widget.QWidget, s.add.QWidget)
	qt6.QWidget_SetTabOrder(s.add.QWidget, s.settings.QWidget)
	qt6.QWidget_SetTabOrder(s.settings.QWidget, s.back.QWidget)

	return s
}

// keepTarget raises the minimum size of an interactive control to the token, so
// a desktop theme that draws a smaller control still meets the target.
func keepTarget(widget *qt6.QWidget) {
	widget.SetMinimumSize2(design.TargetMin, design.TargetMin)
}

// show puts the window on screen and starts the streams that follow it.
func (s *shell) show() {
	s.window.Show()
	s.applyWidth()
	s.startStreams()
}

// present shows the window and brings it to the front: the left click of the
// tray icon and "Mostra finestra" land here. The window manager may still
// refuse the focus, so the window is at least restored and raised.
func (s *shell) present() {
	s.window.ShowNormal()
	s.window.Raise()
	s.window.ActivateWindow()
	s.startStreams()
}

// hideWindow takes the window off the screen without destroying it: the Project
// on screen, the list and the history of the Activities stay, and the two
// streams that depend on the window stop.
func (s *shell) hideWindow() {
	s.stopStreams()
	s.window.Hide()
}

// stopStreams ends what draws from a process of its own: the logs of the panel
// and `docker stats`. The output of an Activity keeps arriving, because the
// Manager pushes it into the log of the session.
func (s *shell) stopStreams() {
	if s.bottom != nil {
		s.bottom.Suspend()
	}
	if s.usage != nil {
		s.usage.SetHidden(true)
	}
}

// startStreams starts the two streams the window carries again: a window that
// comes back shows current figures and live logs.
func (s *shell) startStreams() {
	if s.bottom != nil {
		s.bottom.Resume()
	}
	if s.usage != nil {
		s.usage.SetHidden(false)
	}
}

// buildSidebar builds the column of the Projects: the title, the list that
// takes the height and the actions of the column, the settings last.
func (s *shell) buildSidebar() {
	s.sidebar = qt6.NewQWidget2()
	s.sidebar.SetMinimumWidth(260)
	s.sidebar.SetAccessibleName("Progetti")

	layout := qt6.NewQVBoxLayout(s.sidebar)
	layout.SetContentsMargins(design.SpaceL, design.SpaceL, design.SpaceM, design.SpaceM)
	layout.SetSpacing(design.SpaceS)

	title := qt6.NewQLabel3("Progetti")
	setObjectName(title.QWidget, "sidebarTitle")
	layout.AddWidget(title.QWidget)

	s.list = newProjectList(s.theme)
	s.list.onOpen = s.open
	// Before the first snapshot the list is not empty: it is waiting. The line
	// says so, so a slow Docker never reads as "no Project".
	s.list.set(nil, nil, "", true)
	keepTarget(s.list.widget.QWidget)
	layout.AddWidget2(s.list.widget.QWidget, 1)

	actions := qt6.NewQHBoxLayout2()
	actions.SetSpacing(design.SpaceS)

	s.add = qt6.NewQPushButton3("Aggiungi progetto")
	s.add.SetAccessibleName("Aggiungi progetto")
	s.add.SetToolTip("Scegli i File compose di un Progetto da aggiungere alla lista.")
	s.add.OnClicked(s.showAddDialog)
	keepTarget(s.add.QWidget)
	actions.AddWidget(s.add.QWidget)
	actions.AddStretch()

	s.settings = qt6.NewQPushButton3("Impostazioni")
	s.settings.SetAccessibleName("Impostazioni")
	s.settings.SetToolTip("Scegli cosa fa la chiusura della finestra e il comando del terminale.")
	s.settings.OnClicked(func() {
		if s.app != nil {
			s.app.openPreferences()
		}
	})
	keepTarget(s.settings.QWidget)
	actions.AddWidget(s.settings.QWidget)

	layout.AddLayout(actions.QLayout)
}

// buildDetail builds the panel of the right. The panel starts at rest, with the
// invitation to choose a Project, and fills when the user chooses one. Its
// Actions reach the Manager through do.
func (s *shell) buildDetail() {
	s.panel = newDetailPanel(s.theme)
	s.panel.onAction = s.do
	s.panel.onSave = s.save
	s.panel.onFolder = s.openFolder
	s.panel.onTerminal = s.openTerminal
	s.panel.onPort = s.openPort
	s.panel.onRemove = s.removeFromList
	s.detail = s.panel.widget
	// The figures of the machine reach the panel on the Qt thread, and only
	// while the Project on screen has a Container to measure.
	s.usage = usage.NewUsageFeed(s.app.ctx, s.app.usageSource(), mainthread.Start)
	s.usage.OnSample = s.panel.setUsage
}

// do launches an Action on the Manager off the Qt thread: Docker runs for
// seconds, and the window must keep answering. The result comes back on the Qt
// thread, where a refusal becomes the Italian message of the notice and the
// panel lets the command be pressed again.
func (s *shell) do(action compose.Action) {
	manager := s.app.manager
	if manager == nil {
		s.panel.actionDone()
		return
	}
	go func() {
		_, err := manager.Do(s.app.ctx, action)
		mainthread.Start(func() {
			if err != nil {
				s.showNotice(view.UserError(err).Error())
			}
			s.panel.actionDone()
		})
	}()
}

// save hands the settings of the form to the Manager off the Qt thread: the
// database write must not freeze the window. The result comes back on the Qt
// thread, where the panel keeps the form open on a refusal or closes it, and the
// snapshot the Manager publishes brings the new values to the list and the
// detail.
func (s *shell) save(name string, settings compose.Settings) {
	manager := s.app.manager
	if manager == nil {
		s.panel.saveDone(nil)
		return
	}
	go func() {
		err := manager.UpdateSettings(s.app.ctx, name, settings)
		mainthread.Start(func() { s.panel.saveDone(err) })
	}()
}

// noticeTop is the space the notice keeps from the top of the pages: the bar of
// the Utilizzo risorse takes the top of the detail, so the notice starts under
// it. A bar that is not on screen leaves the plain margin.
func (s *shell) noticeTop() int {
	if s.panel == nil || s.panel.strip == nil || !s.panel.strip.widget.IsVisible() {
		return 0
	}
	point := s.panel.strip.widget.MapTo2(s.pages.QWidget, qt6.NewQPoint2(0, 0))
	if point == nil {
		return 0
	}
	return point.Y() + s.panel.strip.widget.Height()
}

// showNotice puts a message in the notice of the window. A refused Action shows
// its reason here.
func (s *shell) showNotice(message string) {
	s.showNoticeAction(message, "", nil)
}

// showNoticeAction puts a message in the notice together with a button that
// runs action. A failed Activity uses it to lead to its output. An empty label
// or a nil action shows the message alone.
func (s *shell) showNoticeAction(message, label string, action func()) {
	if s.notice == nil {
		return
	}
	s.notice.showAction(message, label, action)
}

// buildPages puts the two layouts in the stack of the responsive view and
// starts on the wide one.
func (s *shell) buildPages() {
	// The detail and the bottom panel share the right column: the splitter is
	// vertical, so the handle between them sets how much of the output the user
	// wants without leaving the Project.
	s.column = qt6.NewQSplitter3(qt6.Vertical)
	s.column.SetChildrenCollapsible(false)
	s.column.AddWidget(s.detail)
	s.bottom = newBottomPanel(s.app.activities, s.app.logSource(), s.app.ctx, s.theme, s.column)
	s.bottom.onFailure = func(a activity.Activity) {
		s.showNoticeAction(activity.ActivityToastText(a), showOutputText, func() { s.bottom.Show(a.ID) })
	}
	// The two ways to the logs: the head of the detail opens the whole Project,
	// a Service opens its own output.
	s.panel.onLogs = func(project compose.Project, service string) { s.bottom.ShowLogs(project, service) }
	s.column.AddWidget(s.bottom.widget)
	s.column.SetStretchFactor(0, 1)
	// The panel is built closed: Qt shows the children of a splitter, so the
	// state of the panel must win.
	s.bottom.Close()

	s.splitter = qt6.NewQSplitter3(qt6.Horizontal)
	s.splitter.SetChildrenCollapsible(false)
	s.splitter.AddWidget(s.sidebar)
	s.splitter.AddWidget(s.column.QWidget)
	s.splitter.SetStretchFactor(1, 1)
	s.splitter.SetSizes([]int{300, windowWidth - 300})

	s.widePage = qt6.NewQWidget2()
	wide := qt6.NewQVBoxLayout(s.widePage)
	wide.SetContentsMargins(0, 0, 0, 0)
	wide.AddWidget(s.splitter.QWidget)

	s.back = qt6.NewQPushButton3("Indietro")
	s.back.SetAccessibleName("Indietro")
	s.back.SetToolTip("Torna alla lista dei Progetti.")
	s.back.OnClicked(func() { s.showList() })
	s.back.SetVisible(false)
	keepTarget(s.back.QWidget)

	s.narrowDetailPage = qt6.NewQWidget2()
	s.narrowDetailLayout = qt6.NewQVBoxLayout(s.narrowDetailPage)
	s.narrowDetailLayout.SetContentsMargins(design.SpaceS, design.SpaceS, design.SpaceS, 0)
	s.narrowDetailLayout.SetSpacing(0)
	// The way back sits above the detail and never moves: only the column of the
	// detail and of the panel passes from the splitter to this page. It keeps
	// its own width, so the button reads as an action and not as a title bar.
	s.narrowDetailLayout.AddWidget3(s.back.QWidget, 0, qt6.AlignLeft)

	s.narrowListPage = qt6.NewQWidget2()
	s.narrowListLayout = qt6.NewQVBoxLayout(s.narrowListPage)
	s.narrowListLayout.SetContentsMargins(0, 0, 0, 0)
	s.narrowListLayout.SetSpacing(0)

	s.pages = qt6.NewQStackedWidget2()
	s.pages.AddWidget(s.widePage)
	s.pages.AddWidget(s.narrowListPage)
	s.pages.AddWidget(s.narrowDetailPage)
	s.pages.SetCurrentWidget(s.widePage)
}

// applyWidth moves the columns to the layout the width asks for. It runs on
// every resize and changes nothing while the side of the breakpoint is the
// same. The columns are reparented, so the list keeps its rows and the scroll
// position across the change.
func (s *shell) applyWidth() {
	if s.pages == nil {
		return
	}
	narrow := s.window.Width() < breakpointWidth
	if narrow == s.narrow {
		return
	}
	s.narrow = narrow

	if narrow {
		s.narrowListLayout.AddWidget(s.sidebar)
		s.narrowDetailLayout.AddWidget(s.column.QWidget)
		s.back.SetVisible(true)
		s.linkTabChain()
		s.showNarrowPage()
		return
	}

	s.back.SetVisible(false)
	s.splitter.AddWidget(s.sidebar)
	s.splitter.AddWidget(s.column.QWidget)
	s.splitter.SetSizes([]int{300, max(s.window.Width()-300, 300)})
	s.linkTabChain()
	s.pages.SetCurrentWidget(s.widePage)
}

// showNarrowPage shows the page the state asks for: the detail of the open
// Project, or the list.
func (s *shell) showNarrowPage() {
	if s.selected != "" {
		s.pages.SetCurrentWidget(s.narrowDetailPage)
		return
	}
	s.pages.SetCurrentWidget(s.narrowListPage)
}

// showList goes back to the list. The selection stays, so the row of the open
// Project keeps its mark and a second click on it opens the detail again.
func (s *shell) showList() {
	s.pages.SetCurrentWidget(s.narrowListPage)
}

// open marks the Project the user chose and shows it in the detail. In the
// narrow view it also moves to the detail page, which is the only page that can
// show it.
func (s *shell) open(project compose.Project) {
	s.selected = project.Name
	s.panel.setGroups(s.snapshot.Groups)
	s.panel.set(project, s.snapshot.DockerReachable, s.snapshot.Loading)
	s.usage.Follow(s.snapshot.DockerReachable)
	s.bottom.setProject(project)
	s.linkTabChain()
	if s.narrow {
		s.showNarrowPage()
	}
}

// apply rebuilds the list and the detail from a snapshot. The detail keeps the
// Project it shows while the snapshot holds it: the panel rebuilds only when
// the data it draws changes. A Project that left the snapshot leaves the
// selection.
func (s *shell) apply(snapshot compose.Snapshot) {
	s.snapshot = snapshot
	s.usage.Follow(snapshot.DockerReachable)
	s.panel.setDocker(snapshot.DockerReachable, snapshot.Loading)
	s.list.set(snapshot.Projects, snapshot.Groups, s.selected, snapshot.Loading)
	s.panel.setGroups(snapshot.Groups)
	if s.selected == "" {
		s.panel.clear()
		s.bottom.setProject(compose.Project{})
		return
	}
	project, ok := projectOf(snapshot.Projects, s.selected)
	if !ok {
		s.selected = ""
		s.panel.clear()
		s.bottom.setProject(compose.Project{})
		if s.narrow {
			s.showNarrowPage()
		}
		return
	}
	s.panel.set(project, snapshot.DockerReachable, snapshot.Loading)
	s.bottom.setProject(project)
	s.linkTabChain()
}

// tabEntry is the control the tab chain reaches the detail from: the way back
// of the narrow view, which sits above it, or the last control of the column
// in the wide one.
func (s *shell) tabEntry() *qt6.QWidget {
	if s.narrow {
		return s.back.QWidget
	}
	return s.settings.QWidget
}

// linkTabChain puts the detail and the panel after the entry the width chooses,
// so the keyboard reaches them in the reading order and not in the order of
// creation.
func (s *shell) linkTabChain() {
	s.panel.linkAfter(s.tabEntry())
	prev := s.panel.lastControl()
	if prev == nil {
		prev = s.tabEntry()
	}
	s.bottom.linkAfter(prev)
}

// projectOf finds a Project of the snapshot by its compose name.
func projectOf(projects []compose.Project, name string) (compose.Project, bool) {
	for _, project := range projects {
		if project.Name == name {
			return project, true
		}
	}
	return compose.Project{}, false
}
