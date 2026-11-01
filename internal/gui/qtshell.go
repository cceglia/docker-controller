package gui

import (
	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"

	"dockercontroller/internal/compose"
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
	theme designTheme

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

	list     *projectList
	panel    *detailPanel
	bottom   *bottomPanel
	add      *qt6.QPushButton
	settings *qt6.QPushButton
	back     *qt6.QPushButton
	// notice shows a message in sovrimpressione, off the commands. A refused
	// Action shows its Italian message there, a failed Activity the same with
	// the way to its output.
	notice *toast

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
func newShell(app *application, theme designTheme) *shell {
	s := &shell{app: app, theme: theme}

	s.window = qt6.NewQMainWindow2()
	s.window.SetWindowTitle(windowTitle)
	s.window.SetAccessibleName(windowTitle)
	s.window.Resize(windowWidth, windowHeight)
	s.window.SetMinimumSize2(560, 420)

	s.buildSidebar()
	s.buildDetail()
	s.buildPages()
	s.notice = newToast(s.pages.QWidget)

	s.window.SetCentralWidget(s.pages.QWidget)
	s.window.OnResizeEvent(func(super func(event *qt6.QResizeEvent), event *qt6.QResizeEvent) {
		super(event)
		s.applyWidth()
		s.notice.place()
	})

	// The tab order follows the reading order: the list, the two actions of the
	// column, the detail, then the way back of the narrow view. The panel adds
	// its own controls after the entry the width chooses.
	qt6.QWidget_SetTabOrder(s.list.widget.QWidget, s.add.QWidget)
	qt6.QWidget_SetTabOrder(s.add.QWidget, s.settings.QWidget)
	qt6.QWidget_SetTabOrder(s.settings.QWidget, s.back.QWidget)

	return s
}

// keepTarget raises the minimum size of an interactive control to the token, so
// a desktop theme that draws a smaller control still meets the target.
func keepTarget(widget *qt6.QWidget) {
	widget.SetMinimumSize2(targetMin, targetMin)
}

// show puts the window on screen.
func (s *shell) show() {
	s.window.Show()
	s.applyWidth()
}

// buildSidebar builds the column of the Projects: the title, the list that
// takes the height and the two actions of later tickets.
func (s *shell) buildSidebar() {
	s.sidebar = qt6.NewQWidget2()
	s.sidebar.SetMinimumWidth(260)
	s.sidebar.SetAccessibleName("Progetti")

	layout := qt6.NewQVBoxLayout(s.sidebar)
	layout.SetContentsMargins(spaceL, spaceL, spaceM, spaceM)
	layout.SetSpacing(spaceS)

	title := qt6.NewQLabel3("Progetti")
	setObjectName(title.QWidget, "sidebarTitle")
	layout.AddWidget(title.QWidget)

	s.list = newProjectList(s.theme)
	s.list.onOpen = s.open
	// Before the first snapshot the list is not empty: it is waiting. The line
	// says so, so a slow Docker never reads as "no Project".
	s.list.set(nil, "", true)
	keepTarget(s.list.widget.QWidget)
	layout.AddWidget2(s.list.widget.QWidget, 1)

	actions := qt6.NewQHBoxLayout2()
	actions.SetSpacing(spaceS)

	s.add = qt6.NewQPushButton3("Aggiungi progetto")
	s.add.SetAccessibleName("Aggiungi progetto")
	s.add.SetEnabled(false)
	keepTarget(s.add.QWidget)
	actions.AddWidget(s.add.QWidget)
	actions.AddStretch()

	s.settings = qt6.NewQPushButton3("Impostazioni")
	s.settings.SetAccessibleName("Impostazioni")
	s.settings.SetEnabled(false)
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
	s.detail = s.panel.widget
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
				s.showNotice(UserError(err).Error())
			}
			s.panel.actionDone()
		})
	}()
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
	s.bottom.onFailure = func(a Activity) {
		s.showNoticeAction(ActivityToastText(a), showOutputText, func() { s.bottom.Show(a.ID) })
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
	s.narrowDetailLayout.SetContentsMargins(spaceS, spaceS, spaceS, 0)
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
	s.panel.set(project, s.snapshot.DockerReachable, s.snapshot.Loading)
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
	s.list.set(snapshot.Projects, s.selected, snapshot.Loading)
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
