package gui

import (
	"github.com/mappu/miqt/qt6"

	"dockercontroller/internal/compose"
)

// shell is the window of the application: the column of the Projects on the
// left and the panel of the detail on the right, or one page at a time when the
// window is narrower than breakpointWidth. It reads the snapshot and never
// changes a Project; the actions arrive in later tickets.
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

	list     *projectList
	add      *qt6.QPushButton
	settings *qt6.QPushButton
	back     *qt6.QPushButton

	// selected is the compose name of the open Project, "" when none is open.
	selected string
	// narrow is true while the window shows one page at a time.
	narrow bool
}

// newShell builds the window. The Project chosen in the list only marks the
// selection: the detail arrives in a later ticket, so the right panel invites
// the user to choose one.
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

	s.window.SetCentralWidget(s.pages.QWidget)
	s.window.OnResizeEvent(func(super func(event *qt6.QResizeEvent), event *qt6.QResizeEvent) {
		super(event)
		s.applyWidth()
	})

	// The tab order follows the reading order: the list, the two actions of the
	// column, then the way back of the narrow view.
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

// buildDetail builds the panel of the right. It only invites the user to choose
// a Project: the detail, the panel of the Activities and the Utilizzo risorse
// arrive in later tickets.
func (s *shell) buildDetail() {
	s.detail = qt6.NewQWidget2()
	s.detail.SetAccessibleName("Dettaglio del Progetto")

	layout := qt6.NewQVBoxLayout(s.detail)
	layout.SetContentsMargins(spaceXL, spaceXL, spaceXL, spaceXL)
	layout.AddStretch()

	// The invite is a block of its own with a bounded width: the sentence fits
	// on one line at the body size, so the layout never squeezes it.
	invite := qt6.NewQWidget2()
	invite.SetMaximumWidth(640)
	box := qt6.NewQVBoxLayout(invite)
	box.SetSpacing(spaceS)

	title := qt6.NewQLabel3("Scegli un Progetto")
	setObjectName(title.QWidget, "placeholderTitle")
	title.SetAlignment(qt6.AlignCenter)
	box.AddWidget(title.QWidget)

	body := qt6.NewQLabel3("Seleziona un Progetto dalla lista per vederne i dettagli.")
	setObjectName(body.QWidget, "placeholderBody")
	body.SetAlignment(qt6.AlignCenter)
	box.AddWidget(body.QWidget)

	layout.AddWidget3(invite, 0, qt6.AlignHCenter)
	layout.AddStretch()
}

// buildPages puts the two layouts in the stack of the responsive view and
// starts on the wide one.
func (s *shell) buildPages() {
	s.splitter = qt6.NewQSplitter3(qt6.Horizontal)
	s.splitter.SetChildrenCollapsible(false)
	s.splitter.AddWidget(s.sidebar)
	s.splitter.AddWidget(s.detail)
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
	s.narrowDetailLayout.SetContentsMargins(0, 0, 0, 0)
	s.narrowDetailLayout.SetSpacing(0)
	// The way back sits above the detail and never moves: only the detail
	// passes from the splitter to this page.
	s.narrowDetailLayout.AddWidget(s.back.QWidget)

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
		s.narrowDetailLayout.AddWidget(s.detail)
		s.back.SetVisible(true)
		s.showNarrowPage()
		return
	}

	s.back.SetVisible(false)
	s.splitter.AddWidget(s.sidebar)
	s.splitter.AddWidget(s.detail)
	s.splitter.SetSizes([]int{300, max(s.window.Width()-300, 300)})
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

// open marks the Project the user chose. In the narrow view it also moves to
// the detail page, which is the only page that can show it.
func (s *shell) open(project compose.Project) {
	s.selected = project.Name
	if s.narrow {
		s.showNarrowPage()
	}
}

// apply rebuilds the list from a snapshot. A Project that left the snapshot
// leaves the selection: the list, the detail and the narrow page agree again.
func (s *shell) apply(snapshot compose.Snapshot) {
	s.list.set(snapshot.Projects, s.selected, snapshot.Loading)
	if s.selected == "" {
		return
	}
	if projectNamed(snapshot.Projects, s.selected) {
		return
	}
	s.selected = ""
	if s.narrow {
		s.showNarrowPage()
	}
}

// projectNamed reports whether the snapshot still holds the Project.
func projectNamed(projects []compose.Project, name string) bool {
	for _, project := range projects {
		if project.Name == name {
			return true
		}
	}
	return false
}
