//go:build ui

package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mappu/miqt/qt6"

	"dockercontroller/internal/compose"
)

// TestShellOffscreen builds the empty window, shows it with the offscreen
// backend and destroys it. A broken constructor or a missing wiring of the
// list, the buttons or the central widget fails here, without a display. It
// also builds the panel of the detail from a fake Project: the blocks, the
// rebuild rule and the place of the scroll bar.
//
// The test keeps every Qt call on its own goroutine: QApplication locks the OS
// thread, and a second goroutine that touches a widget is undefined behaviour.
// For the same reason the test never calls t.Parallel, and the process creates
// one QApplication only. A second UI test that needs Qt belongs to this one,
// as a step of the same function.
func TestShellOffscreen(t *testing.T) {
	// One QApplication per process, before every widget.
	app := qt6.NewQApplication([]string{"docker-controller-ui-test"})
	defer app.Delete()
	qt6.QCoreApplication_SetApplicationName(appID)
	app.SetStyleSheet(applicationStyleSheet(lightTheme))

	s := newShell(&application{}, lightTheme)
	defer s.window.Delete()

	s.show()
	settle()

	if !s.window.IsVisible() {
		t.Fatal("the window is not visible after show with the offscreen backend")
	}
	if got := s.window.WindowTitle(); got != windowTitle {
		t.Errorf("window title = %q, want %q", got, windowTitle)
	}
	if s.window.CentralWidget() == nil {
		t.Fatal("the window has no central widget")
	}
	if s.list == nil || s.list.widget == nil {
		t.Fatal("the list of the Projects is not wired")
	}
	if s.add == nil || s.settings == nil || s.back == nil {
		t.Fatal("an action button is not wired")
	}
	if got := app.StyleSheet(); got == "" {
		t.Error("the application carries no style sheet")
	}

	checkMinimumTargets(t, s.window)
	checkDetailPanel(t, s)
	if dir := os.Getenv("SHOT_DIR"); dir != "" {
		captureScreens(t, dir)
	}
}

// settle runs the events the layout waits for: a page change and a rebuild
// deliver their geometry over more than one turn of the event loop.
func settle() {
	for i := 0; i < 4; i++ {
		qt6.QCoreApplication_ProcessEvents()
	}
}

// checkDetailPanel builds the detail from a fake Project and checks the wiring
// of the blocks, the rebuild rule and the place of the scroll bar. It is a step
// of the smoke test, so the process still has one QApplication.
func checkDetailPanel(t *testing.T, s *shell) {
	t.Helper()

	if s.panel == nil || s.panel.widget == nil {
		t.Fatal("the panel of the detail is not wired")
	}
	if s.panel.stack.CurrentWidget().UnsafePointer() != s.panel.invite.UnsafePointer() {
		t.Fatal("the panel does not start at the invitation")
	}

	project := fakeProject()
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{project}})
	s.open(project)
	settle()

	if got := s.panel.stack.CurrentWidget(); got == nil || got.UnsafePointer() != s.panel.scroll.UnsafePointer() {
		t.Fatal("the detail does not show after the Project is opened")
	}
	if got := s.panel.shown.project.Name; got != project.Name {
		t.Errorf("the panel shows %q, want %q", got, project.Name)
	}
	for _, name := range []string{
		"Dati del Progetto",
		"Azioni del Progetto",
		"Comandi del Progetto",
		"Servizi del Progetto",
	} {
		if blockNamed(s.panel, name) == nil {
			t.Fatalf("the panel has no block %q", name)
		}
	}
	if !hasWidgetNamed(blockNamed(s.panel, "Servizi del Progetto"), "serviceUsage") {
		t.Error("no Service reserves the place of the Utilizzo risorse")
	}

	// The same data draws the same panel: the blocks are not built again, so
	// the focus and the place of the scroll stay where the user left them.
	first := blockAt(s.panel, 0)
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{project}})
	if got := blockAt(s.panel, 0); got == nil || first == nil || got.UnsafePointer() != first.UnsafePointer() {
		t.Error("the panel rebuilt without a change of the data it draws")
	}

	// A change of the data builds the blocks again.
	changed := project
	changed.Running = project.Running - 1
	changed.Status = compose.StatusPartial
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{changed}})
	settle()
	if got := blockAt(s.panel, 0); got == nil || first == nil || got.UnsafePointer() == first.UnsafePointer() {
		t.Error("the panel kept its blocks after a change of the data it draws")
	}

	// An update does not move the page under the eyes of the user.
	bar := s.panel.scroll.VerticalScrollBar()
	if bar == nil {
		t.Fatal("the detail has no scroll bar")
	}
	bar.SetValue(200)
	grown := changed
	grown.Services = append(append([]compose.Service{}, changed.Services...), compose.Service{
		Name:       "last",
		Containers: []compose.Container{{Name: "shop-last-1", State: "running"}},
	})
	grown.Total = changed.Total + 1
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{grown}})
	settle()
	if got := bar.Value(); got != 200 {
		t.Errorf("scroll = %d after the update, want 200", got)
	}

	// A missing Compose file shows the notice with its command, present and
	// insensitive until the ticket that wires it.
	missing := grown
	missing.ComposeMissing = true
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{missing}})
	settle()

	block := blockNamed(s.panel, "File compose mancante")
	if block == nil {
		t.Fatal("the missing Compose file shows no notice")
	}
	row := block.Layout().ItemAt(1)
	if row == nil || row.Layout() == nil {
		t.Fatal("the notice carries no command")
	}
	remove := row.Layout().ItemAt(0).Widget()
	if remove == nil {
		t.Fatal("the notice carries no remove entry")
	}
	if got := remove.ObjectName(); got != "removeButton" {
		t.Errorf("the remove entry is %q, want %q", got, "removeButton")
	}
	if remove.IsEnabled() {
		t.Error("the remove entry is enabled before the ticket that wires it")
	}

	// Before the first snapshot the row of the Actions claims no reason: the
	// reason arrives with the first snapshot that does not reach Docker.
	if got := blockNamed(s.panel, "Azioni del Progetto").Layout().Count(); got != 1 {
		t.Errorf("the row of the Actions carries %d blocks before the first snapshot, want 1", got)
	}
	s.apply(compose.Snapshot{DockerReachable: false, Projects: []compose.Project{missing}})
	settle()
	if got := blockNamed(s.panel, "Azioni del Progetto").Layout().Count(); got != 2 {
		t.Errorf("the row of the Actions carries %d blocks when Docker does not answer, want 2", got)
	}

	// Docker does not answer: the notice takes the place of the numbers.
	header := blockNamed(s.panel, "Intestazione del Progetto")
	if header == nil {
		t.Fatal("the panel has no header")
	}
	if !hasWidgetNamed(header, "dockerNotice") {
		t.Error("the header does not show the notice of Docker")
	}
	if hasWidgetNamed(header, "detailCount") {
		t.Error("the header counts the Containers while Docker does not answer")
	}

	// Before the first snapshot Docker has not been asked: the header keeps the
	// numbers and claims nothing about Docker.
	s.apply(compose.Snapshot{Loading: true, Projects: []compose.Project{missing}})
	settle()
	header = blockNamed(s.panel, "Intestazione del Progetto")
	if hasWidgetNamed(header, "dockerNotice") {
		t.Error("the header calls Docker unreachable before the first snapshot")
	}
	if !hasWidgetNamed(header, "detailCount") {
		t.Error("the header left out the numbers before the first snapshot")
	}

	// The controls of the detail meet the target too.
	checkMinimumTargets(t, s.window)
}

// hasWidgetNamed reports whether a widget or one of its children carries the
// object name of the style sheet. The blocks nest layouts, so the walk goes
// through them.
func hasWidgetNamed(widget *qt6.QWidget, name string) bool {
	if widget == nil {
		return false
	}
	if widget.ObjectName() == name {
		return true
	}
	if layout := widget.Layout(); layout != nil {
		return layoutHasNamed(layout, name)
	}
	return false
}

// layoutHasNamed reports whether a layout holds a widget with the object name
// of the style sheet, directly or inside one of its own layouts.
func layoutHasNamed(layout *qt6.QLayout, name string) bool {
	for i := 0; layout.ItemAt(i) != nil; i++ {
		item := layout.ItemAt(i)
		if child := item.Widget(); child != nil && hasWidgetNamed(child, name) {
			return true
		}
		if nested := item.Layout(); nested != nil && layoutHasNamed(nested, name) {
			return true
		}
	}
	return false
}

// blockAt is the widget of one block of the detail, in the order the panel
// built the blocks.
func blockAt(p *detailPanel, index int) *qt6.QWidget {
	item := p.layout.ItemAt(index)
	if item == nil {
		return nil
	}
	return item.Widget()
}

// blockNamed finds the block of the detail whose accessible name is name.
func blockNamed(p *detailPanel, name string) *qt6.QWidget {
	for i := 0; blockAt(p, i) != nil; i++ {
		if block := blockAt(p, i); block.AccessibleName() == name {
			return block
		}
	}
	return nil
}

// fakeProject is a Project with enough Services to fill more than one screen of
// the detail, so the smoke test can move the scroll bar.
func fakeProject() compose.Project {
	services := make([]compose.Service, 0, 24)
	for i := 0; i < 24; i++ {
		name := fmt.Sprintf("service-%02d", i)
		services = append(services, compose.Service{
			Name: name,
			Containers: []compose.Container{{
				Name:   name + "-1",
				Image:  "example/" + name,
				State:  "running",
				Health: "healthy",
				Ports:  []compose.Port{{Published: 8080 + i, Target: 80, Protocol: "tcp"}},
			}},
		})
	}
	return compose.Project{
		Name:           "shop",
		DisplayName:    "Negozio",
		Group:          "Casa",
		Folder:         "/srv/shop",
		ComposeFiles:   []string{"/srv/shop/docker-compose.yml"},
		Status:         compose.StatusActive,
		Running:        24,
		Total:          24,
		Services:       services,
		ActiveProfiles: []string{"debug"},
	}
}

// checkMinimumTargets fails for every interactive control of the window whose
// smallest side is under the token. A control without a size hint counts as
// zero, so a widget that loses its layout fails too.
func checkMinimumTargets(t *testing.T, window *qt6.QMainWindow) {
	t.Helper()

	checked := 0
	for _, widget := range qt6.QApplication_AllWidgets() {
		if widget.Window().UnsafePointer() != window.QWidget.UnsafePointer() {
			continue
		}
		if widget.FocusPolicy() == qt6.NoFocus {
			continue
		}
		widget.EnsurePolished()
		hint, minimum := widget.SizeHint(), widget.MinimumSizeHint()
		width := max(hint.Width(), minimum.Width())
		height := max(hint.Height(), minimum.Height())
		if width < targetMin || height < targetMin {
			t.Errorf("control %q: target %dx%d, want at least %dx%d",
				widget.ObjectName(), width, height, targetMin, targetMin)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no interactive control found: the wiring is broken")
	}
	t.Logf("checked %d interactive controls against the %d px target", checked, targetMin)
}

// captureScreens writes the PNGs of the main states of the window, so a change
// of layout can be reviewed without a display. The test skips without SHOT_DIR:
// the files are a review aid, not an assertion.
func captureScreens(t *testing.T, dir string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create %q: %v", dir, err)
	}

	s := newShell(&application{}, lightTheme)
	defer s.window.Delete()
	s.show()
	settle()

	shot := func(name string) {
		t.Helper()
		settle()
		pixmap := s.window.Grab()
		if pixmap == nil {
			t.Fatalf("grab %q: no pixmap", name)
		}
		path := filepath.Join(dir, name+".png")
		if !pixmap.Save(path) {
			t.Fatalf("save %q", path)
		}
		t.Logf("wrote %s", path)
	}

	shot("01-empty")

	projects := screenshotProjects()
	s.apply(compose.Snapshot{DockerReachable: true, Projects: projects})
	shot("02-list")

	s.open(projects[1])
	settle()
	shot("03-detail")

	missing := projects[1]
	missing.ComposeMissing = true
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{missing}})
	shot("04-detail-missing")

	s.apply(compose.Snapshot{DockerReachable: false, Projects: []compose.Project{missing}})
	shot("05-detail-unreachable")

	s.apply(compose.Snapshot{DockerReachable: true, Projects: projects})
	s.window.Resize(800, 620)
	s.applyWidth()
	s.showList()
	shot("06-narrow-list")

	s.open(projects[1])
	settle()
	s.applyWidth()
	shot("07-narrow-detail")
}

// screenshotProjects fills the list with every state and a Project with more
// than twelve Services, so the "+n" and the two-line cases both show.
func screenshotProjects() []compose.Project {
	service := func(name string, state string) compose.Service {
		return compose.Service{
			Name: name,
			Containers: []compose.Container{{
				Name:   name + "-1",
				Image:  "example/" + name,
				State:  state,
				Health: "healthy",
				Ports:  []compose.Port{{Published: 8080, Target: 80, Protocol: "tcp"}},
			}},
		}
	}
	many := make([]compose.Service, 0, 14)
	for i := 0; i < 14; i++ {
		many = append(many, service(fmt.Sprintf("service-%02d", i), "running"))
	}
	return []compose.Project{
		{
			Name: "vakt", DisplayName: "VAKT", Group: "Casa",
			Folder: "/srv/vakt", ComposeFiles: []string{"/srv/vakt/docker-compose.yml"},
			Status: compose.StatusActive, Running: 3, Total: 3,
			Services:       []compose.Service{service("web", "running"), service("db", "running"), service("cache", "running")},
			ActiveProfiles: []string{"debug"},
		},
		{
			Name: "shop", DisplayName: "Negozio", Group: "Casa",
			Folder: "/srv/shop", ComposeFiles: []string{"/srv/shop/compose.yml", "/srv/shop/compose.override.yml"},
			Status: compose.StatusPartial, Running: 1, Total: 2,
			Services: []compose.Service{
				service("web", "running"),
				service("worker", "exited"),
			},
		},
		{
			Name: "blog", Group: "Lavoro",
			Folder: "/srv/blog", ComposeFiles: []string{"/srv/blog/docker-compose.yml"},
			Status: compose.StatusStopped, Running: 0, Total: 2,
			Services: []compose.Service{service("web", "exited"), service("db", "exited")},
		},
		{
			Name: "docs", DisplayName: "Documentazione", Group: "Lavoro",
			Folder: "/srv/docs", ComposeFiles: []string{"/srv/docs/docker-compose.yml"},
			Status: compose.StatusRemoved,
		},
		{
			Name:   "sandbox",
			Folder: "/srv/sandbox", ComposeFiles: []string{"/srv/sandbox/docker-compose.yml"},
			Status: compose.StatusActive, Running: 14, Total: 14,
			Services: many,
		},
	}
}
