//go:build ui

package gui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mappu/miqt/qt6"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/model"
	"dockercontroller/internal/gui/activity"
	"dockercontroller/internal/gui/design"
	"dockercontroller/internal/gui/forms"
	"dockercontroller/internal/gui/guitest"
	"dockercontroller/internal/gui/logs"
	guiproject "dockercontroller/internal/gui/project"
	"dockercontroller/internal/gui/traymenu"
	"dockercontroller/internal/gui/usage"
	"dockercontroller/internal/gui/view"
	"dockercontroller/internal/settings"
	"dockercontroller/internal/tray"
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
	app.SetStyleSheet(design.ApplicationStyleSheet(design.LightTheme, ""))

	s := newShell(&application{}, design.LightTheme)
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
	checkList(t, s)
	checkEditForm(t, s)
	checkGroupsDialog(t, s)
	checkActions(t, s)
	checkConfirmDialog(t, s)
	checkCommands(t, s)
	checkPicker(t, s)
	checkAddForm(t, s)
	checkBottomPanel(t)
	checkLogs(t)
	checkUsage(t, s)
	checkTray(t)
	checkPreferences(t, s)
	checkCloseDialog(t, s)
	checkHiddenWindow(t)
	if dir := os.Getenv("SHOT_DIR"); dir != "" {
		captureScreens(t, dir, app)
	}
	// The quit comes last: it ends the application of the process.
	checkQuit(t)
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
	if got := s.panel.shown.Project.Name; got != project.Name {
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
	if !hasWidgetNamed(blockNamed(s.panel, "Servizi del Progetto"), "serviceState") {
		t.Error("no Service shows the state of its Containers")
	}
	if !hasWidgetNamed(s.panel.content, "removeCommandButton") &&
		!hasWidgetNamed(s.panel.content, "removeButton") {
		t.Error("the detail carries no way to take the Project off the list")
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
	remove := qt6.UnsafeNewQPushButton(row.Layout().ItemAt(0).Widget().UnsafePointer())
	if got := remove.ObjectName(); got != "removeButton" {
		t.Errorf("the remove entry is %q, want %q", got, "removeButton")
	}
	if remove.IsEnabled() {
		t.Error("the remove entry is enabled on a Project that still has Containers")
	}
	if got := remove.ToolTip(); got != guiproject.ContainerRefusalText {
		t.Errorf("the remove entry says %q, want %q", got, guiproject.ContainerRefusalText)
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

// checkActions checks the wiring of the Actions: the commands of the Project
// and of every Service are there and follow the snapshot, an insensitive
// command shows its reason, and Down opens its question without showing it in a
// blocking way.
func checkActions(t *testing.T, s *shell) {
	t.Helper()

	project := fakeProject()
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{project}})
	s.open(project)
	settle()

	actions := blockNamed(s.panel, "Azioni del Progetto")
	if actions == nil {
		t.Fatal("the panel has no row of Actions")
	}
	commands := projectActionButtons(actions)
	if got, want := len(commands), len(guiproject.ProjectActions(project, true)); got != want {
		t.Fatalf("the row of Actions carries %d commands, want %d", got, want)
	}
	primary := widgetsNamed(actions, "primaryAction")
	if len(primary) != 1 {
		t.Fatalf("the row of Actions carries %d primary commands, want 1", len(primary))
	}
	if got, want := primary[0].AccessibleName(), guiproject.ProjectActions(project, true)[0].Label; got != want {
		t.Errorf("the primary command is %q, want the one the state suggests: %q", got, want)
	}
	for _, button := range commands {
		if !button.IsEnabled() {
			t.Errorf("the command %q is insensitive with Docker answering", button.AccessibleName())
		}
	}

	services := blockNamed(s.panel, "Servizi del Progetto")
	if services == nil {
		t.Fatal("the panel has no block of Services")
	}
	perService := len(guiproject.ServiceActions(project, project.Services[0], true))
	if got, want := len(widgetsNamed(services, "serviceActionButton")), len(project.Services)*perService; got != want {
		t.Errorf("the Services carry %d commands, want %d", got, want)
	}

	// An Activity disables every command and writes the reason beside the rows.
	busy := project
	busy.Activity = &compose.ActivityRef{ID: "act-1", Kind: compose.ActionUp}
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{busy}})
	settle()

	actions = blockNamed(s.panel, "Azioni del Progetto")
	for _, button := range projectActionButtons(actions) {
		if button.IsEnabled() {
			t.Errorf("the command %q is enabled during an Activity", button.AccessibleName())
		}
		if got := button.AccessibleDescription(); got != guiproject.ActivityRunningText {
			t.Errorf("the command %q says %q, want %q", button.AccessibleName(), got, guiproject.ActivityRunningText)
		}
	}
	if !hasWidgetNamed(actions, "actionsReason") {
		t.Error("the row of Actions hides the reason of an insensitive command")
	}
	for _, button := range widgetsNamed(blockNamed(s.panel, "Servizi del Progetto"), "serviceActionButton") {
		if got := button.AccessibleDescription(); got != guiproject.ActivityRunningText {
			t.Errorf("the command %q says %q, want %q", button.AccessibleName(), got, guiproject.ActivityRunningText)
		}
	}

	// Down opens its question, and a rejection launches nothing.
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{project}})
	settle()
	down, ok := actionKind(guiproject.ProjectActions(project, true), compose.ActionDown)
	if !ok {
		t.Fatal("the Project offers no Down")
	}
	s.panel.choose(down, nil, down.Label)
	settle()
	if s.panel.question == nil {
		t.Fatal("Down opens no question")
	}
	if !s.panel.question.confirm.IsEnabled() {
		t.Error("the affirmative button of Down is off without a text to type")
	}
	s.panel.question.dialog.Reject()
	settle()
	if s.panel.question != nil {
		t.Error("the question stays open after a rejection")
	}
	if s.panel.guard.waiting() {
		t.Error("a rejected question took the guard: it launched something")
	}
}

// actionKind finds the view of an ActionKind in a row.
func actionKind(views []guiproject.ActionView, kind compose.ActionKind) (guiproject.ActionView, bool) {
	for _, view := range views {
		if view.Action.Kind == kind {
			return view, true
		}
	}
	return guiproject.ActionView{}, false
}

// checkList drives the Projects list: the headings follow the order of the
// snapshot, an empty Group produces no heading, an unknown Group goes after the
// known ones and before "Senza Gruppo", and a list with no Group at all stays
// flat. It is a step of the smoke test, so the process still has one
// QApplication.
func checkList(t *testing.T, s *shell) {
	t.Helper()

	projects := []compose.Project{
		{Name: "solo"},
		{Name: "work", Group: "Lavoro"},
		{Name: "home", Group: "Casa"},
		{Name: "zeta", Group: "Zeta"},
	}
	// Lavoro before Casa is the chosen order; Zeta is not in the snapshot.
	s.apply(compose.Snapshot{DockerReachable: true, Projects: projects, Groups: []string{"Lavoro", "Casa", "Vuoto"}})
	settle()

	got := listRows(s.list.widget)
	want := []string{"Lavoro", "work", "Casa", "home", "Zeta", "zeta", "Senza Gruppo", "solo"}
	if !slices.Equal(got, want) {
		t.Errorf("the list shows %v, want %v", got, want)
	}

	// A saved Group with no Project produces no heading, and without a Project
	// that carries a Group the list stays flat and unnamed.
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{{Name: "a"}, {Name: "b"}}, Groups: []string{"Lavoro", "Casa"}})
	settle()
	got = listRows(s.list.widget)
	if want := []string{"a", "b"}; !slices.Equal(got, want) {
		t.Errorf("the flat list shows %v, want %v", got, want)
	}
}

// listRows lists the text of every row of the list, headings included.
func listRows(list *qt6.QListWidget) []string {
	var rows []string
	for i := 0; ; i++ {
		item := list.Item(i)
		if item == nil {
			return rows
		}
		widget := list.ItemWidget(item)
		for _, label := range widgetsNamed(widget, "listHeading") {
			rows = append(rows, qt6.UnsafeNewQLabel(label.UnsafePointer()).Text())
		}
		for _, label := range widgetsNamed(widget, "projectName") {
			rows = append(rows, qt6.UnsafeNewQLabel(label.UnsafePointer()).Text())
		}
	}
}

// checkEditForm drives the form of the settings: the command of the data opens
// it with the saved values, the Group field proposes the Groups in use, a
// snapshot that arrives while it is open keeps the text the user typed, a
// refusal writes its message under the field, and Salva and Annulla leave the
// form. It is a step of the smoke test, so the process still has one
// QApplication.
func checkEditForm(t *testing.T, s *shell) {
	t.Helper()

	shop := fakeProject()
	shop.Profiles = []string{"debug", "metrics"}
	other := shop
	other.Name = "vakt"
	other.DisplayName = ""
	other.Group = "Lavoro"
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{shop, other}, Groups: []string{"Lavoro"}})
	s.open(shop)
	settle()

	// The command of the data opens the form in its place.
	data := blockNamed(s.panel, dataTitle)
	if data == nil {
		t.Fatal("the detail shows no data of the Project")
	}
	commands := widgetsNamed(data, "editButton")
	if len(commands) != 1 {
		t.Fatalf("the data carries %d commands to open the form, want 1", len(commands))
	}
	qt6.UnsafeNewQPushButton(commands[0].UnsafePointer()).Click()
	settle()

	if blockNamed(s.panel, editTitle) == nil {
		t.Fatal("the command opens no form")
	}
	if blockNamed(s.panel, dataTitle) != nil {
		t.Error("the form does not replace the data of the Project")
	}

	// The form starts on the saved values: the Display name, the Group and the
	// active Profiles.
	if got := s.panel.edit.name.Text(); got != shop.DisplayName {
		t.Errorf("the Display name field holds %q, want the saved %q", got, shop.DisplayName)
	}
	if got := s.panel.edit.group.CurrentText(); got != shop.Group {
		t.Errorf("the Group field holds %q, want the saved %q", got, shop.Group)
	}
	// The box proposes the saved Groups in their order, without the empty
	// option: an empty box is no Group, and the hint of the box says so.
	group := s.panel.edit.group
	if got, want := itemsOf(group), forms.GroupSuggestions(forms.GroupChoices(shop.Group, []string{"Lavoro"})); !slices.Equal(got, want) {
		t.Errorf("the Group field proposes %v, want %v", got, want)
	}
	if !slices.Contains(itemsOf(group), other.Group) {
		t.Errorf("the Group field does not propose the saved Group %q", other.Group)
	}
	if got := group.PlaceholderText(); got != forms.NoGroupText {
		t.Errorf("an empty Group box says %q, want %q", got, forms.NoGroupText)
	}
	if got := group.AccessibleDescription(); !strings.Contains(got, "crea un Gruppo") {
		t.Errorf("the Group hint reads %q, want it to say a new name creates a Group", got)
	}
	if !s.panel.edit.checks["debug"].IsChecked() {
		t.Error("the form leaves the active Profile debug off")
	}
	if s.panel.edit.checks["metrics"].IsChecked() {
		t.Error("the form turns the idle Profile metrics on")
	}

	// The form is one step of the keyboard: the tab order follows the reading
	// order, and the focus starts on its first field.
	if got, want := formOrder(s.panel, shop), []string{
		displayNameLabelText, groupLabelText, groupsCommandText, "debug", "metrics", cancelText, saveText,
	}; !slices.Equal(got, want) {
		t.Errorf("the tab order of the form is %v, want %v", got, want)
	}
	if got := s.window.FocusWidget(); got == nil || got.UnsafePointer() != s.panel.edit.name.QWidget.UnsafePointer() {
		t.Error("the form does not start with the focus on its first field")
	}

	// A snapshot that arrives while the form is open never overwrites the text
	// the user typed: the form stays and the data waits.
	s.panel.edit.name.SetText("Negozio nuovo")
	s.panel.edit.checks["metrics"].SetChecked(true)
	group.SetCurrentText("Lavoro")
	settle()
	moved := shop
	moved.Group = "Magazzino"
	moved.Running = shop.Running - 1
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{moved, other}})
	settle()
	if blockNamed(s.panel, editTitle) == nil {
		t.Fatal("the snapshot closes the form")
	}
	if got := s.panel.edit.name.Text(); got != "Negozio nuovo" {
		t.Errorf("the snapshot overwrote the Display name with %q", got)
	}
	if got := s.panel.edit.group.CurrentText(); got != "Lavoro" {
		t.Errorf("the snapshot overwrote the Group with %q", got)
	}
	if !s.panel.edit.edit.ProfileOn("metrics") {
		t.Error("the snapshot turned the Profile the user checked off")
	}

	// A refused save keeps the form open and writes the Italian reason under the
	// Display name field, which repeats it in its accessible name.
	var savedName string
	var saved compose.Settings
	s.panel.onSave = func(name string, settings compose.Settings) {
		savedName, saved = name, settings
		s.panel.saveDone(compose.ErrUnknownProject)
	}
	s.panel.edit.save.Click()
	settle()

	if savedName != shop.Name {
		t.Errorf("the save reaches %q, want %q", savedName, shop.Name)
	}
	if saved.DisplayName == nil || *saved.DisplayName != "Negozio nuovo" {
		t.Errorf("the save sends %v, want %q", saved.DisplayName, "Negozio nuovo")
	}
	if saved.Group == nil || *saved.Group != "Lavoro" {
		t.Errorf("the save sends %v, want %q", saved.Group, "Lavoro")
	}
	if want := []string{"debug", "metrics"}; !slices.Equal(saved.Profiles, want) {
		t.Errorf("the save sends the Profiles %v, want %v", saved.Profiles, want)
	}
	form := blockNamed(s.panel, editTitle)
	if form == nil {
		t.Fatal("a refused save closes the form")
	}
	message := view.UserError(compose.ErrUnknownProject).Error()
	if got := errorUnder(form); got != message {
		t.Errorf("the refusal reads %q under the field, want %q", got, message)
	}
	if got := s.panel.edit.name.AccessibleName(); !strings.Contains(got, message) {
		t.Errorf("the field is named %q, want it to repeat %q", got, message)
	}

	// A reserved Group name is refused under the Group field, not under the
	// Display name, and the Group field repeats the reason in its accessible
	// name.
	s.panel.edit.group.SetCurrentText(forms.NoGroupText)
	settle()
	s.panel.onSave = func(string, compose.Settings) { s.panel.saveDone(compose.ErrInvalidGroupName) }
	s.panel.edit.save.Click()
	settle()
	form = blockNamed(s.panel, editTitle)
	if form == nil {
		t.Fatal("a refused Group name closes the form")
	}
	message = view.UserError(compose.ErrInvalidGroupName).Error()
	if got := errorUnder(form); got != message {
		t.Errorf("the refusal reads %q under the Group field, want %q", got, message)
	}
	if got := s.panel.edit.group.AccessibleName(); !strings.Contains(got, message) {
		t.Errorf("the Group field is named %q, want it to repeat %q", got, message)
	}
	if got := s.panel.edit.name.AccessibleName(); strings.Contains(got, message) {
		t.Errorf("the Display name field is named %q, want no Group error", got)
	}

	// A save that works closes the form and gives the focus back to the command
	// that opened it.
	s.panel.onSave = func(string, compose.Settings) { s.panel.saveDone(nil) }
	s.panel.edit.save.Click()
	settle()
	if blockNamed(s.panel, editTitle) != nil {
		t.Error("a save that worked leaves the form open")
	}
	modify := widgetsNamed(blockNamed(s.panel, dataTitle), "editButton")
	if len(modify) != 1 {
		t.Fatal("the data carries no command to open the form")
	}
	if got := s.window.FocusWidget(); got == nil || got.UnsafePointer() != modify[0].UnsafePointer() {
		t.Error("the closing form does not give the focus back to the command that opened it")
	}

	// The snapshot after the save brings the new values, so the list and the
	// detail never need a restart.
	after := moved
	after.DisplayName = "Negozio nuovo"
	after.Group = "Lavoro"
	after.ActiveProfiles = []string{"debug", "metrics"}
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{after, other}})
	settle()
	if got := s.panel.shown.Project.DisplayName; got != "Negozio nuovo" {
		t.Errorf("the detail shows the Display name %q after the save, want %q", got, "Negozio nuovo")
	}
	if !slices.Contains(listNames(s.list.widget), "Negozio nuovo") {
		t.Error("the list does not show the new Display name after the save")
	}

	// Annulla leaves the form and the saved values as they are.
	s.panel.startEdit()
	settle()
	if blockNamed(s.panel, editTitle) == nil {
		t.Fatal("the form does not open on the saved values")
	}
	s.panel.edit.name.SetText("Scartato")
	s.panel.cancelEdit()
	settle()
	if blockNamed(s.panel, editTitle) != nil {
		t.Error("Annulla leaves the form open")
	}
	if got := s.panel.shown.Project.DisplayName; got != "Negozio nuovo" {
		t.Errorf("Annulla changed the Project to %q", got)
	}

	// A blank Display name clears the setting, so the compose name comes back to
	// the list and to the detail, as DisplayNameOf says.
	s.panel.onSave = func(_ string, settings compose.Settings) {
		saved = settings
		s.panel.saveDone(nil)
	}
	s.panel.startEdit()
	settle()
	s.panel.edit.name.SetText("   ")
	s.panel.edit.save.Click()
	settle()
	if saved.DisplayName == nil || *saved.DisplayName != "" {
		t.Errorf("a blank Display name sends %v, want an empty value", saved.DisplayName)
	}
	cleared := after
	cleared.DisplayName = ""
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{cleared, other}})
	settle()
	if got := view.ProjectName(s.panel.shown.Project); got != cleared.Name {
		t.Errorf("the detail names the Project %q, want the compose name %q", got, cleared.Name)
	}
	if !slices.Contains(listNames(s.list.widget), cleared.Name) {
		t.Errorf("the list does not fall back to the compose name %q", cleared.Name)
	}

	// The form answers the keyboard: Return saves and Escape cancels from its
	// fields, without the pointer.
	s.panel.onSave = func(string, compose.Settings) { s.panel.saveDone(nil) }
	s.panel.startEdit()
	settle()
	s.panel.edit.group.SetFocus()
	pressKey(s, qt6.Key_Return)
	settle()
	if blockNamed(s.panel, editTitle) != nil {
		t.Error("Return in the Group field does not save the form")
	}
	if blockNamed(s.panel, dataTitle) == nil {
		t.Error("Return leaves no data of the Project")
	}

	s.panel.startEdit()
	settle()
	pressKey(s, qt6.Key_Escape)
	settle()
	if blockNamed(s.panel, editTitle) != nil {
		t.Error("Escape does not cancel the form")
	}

	s.panel.onSave = s.save
	checkMinimumTargets(t, s.window)
}

// checkGroupsDialog drives the dialog of the Groups from the form of a Project:
// creation, rename, union and deletion through its widgets, the counts of the
// list, the disabled commands, the inline name field, the snapshot that keeps
// what the user typed, the focus, the keyboard and the records the form applies
// at the close. It is a step of the smoke test.
func checkGroupsDialog(t *testing.T, s *shell) {
	t.Helper()

	shop := fakeProject()
	shop.Profiles = []string{"debug", "metrics"}
	other := shop
	other.Name = "vakt"
	other.DisplayName = ""
	other.Group = "Lavoro"
	projects := []compose.Project{shop, other}
	groups := []string{"Casa", "Lavoro", "Vuoto"}
	s.apply(compose.Snapshot{DockerReachable: true, Projects: projects, Groups: groups})
	s.open(shop)
	settle()
	s.panel.startEdit()
	settle()

	form := blockNamed(s.panel, editTitle)
	if form == nil {
		t.Fatal("the form of the Project does not open")
	}
	commands := widgetsNamed(form, "editGroups")
	if len(commands) != 1 {
		t.Fatalf("the Group field carries %d commands to open the Groups, want 1", len(commands))
	}
	groupsCommand := qt6.UnsafeNewQPushButton(commands[0].UnsafePointer())
	if got := groupsCommand.Text(); got != groupsCommandText {
		t.Errorf("the command reads %q, want %q", got, groupsCommandText)
	}
	if !groupsCommand.IsEnabled() {
		t.Error("the command of the Groups is insensitive on the form")
	}

	groupsCommand.Click()
	settle()
	d := s.groups
	if d == nil {
		t.Fatal("the command opens no dialog of the Groups")
	}
	if !d.dialog.IsModal() {
		t.Error("the dialog of the Groups is not modal")
	}
	if got := d.dialog.WindowTitle(); got != groupsTitle {
		t.Errorf("the dialog is titled %q, want %q", got, groupsTitle)
	}
	if got := d.list.AccessibleName(); got != groupsTitle {
		t.Errorf("the list is named %q, want %q", got, groupsTitle)
	}
	if got := d.entry.AccessibleName(); got != groupNameLabelText {
		t.Errorf("the field is named %q, want %q", got, groupNameLabelText)
	}

	// The list follows the order of the snapshot and counts the Projects of each
	// Group, saying when there is none.
	wantRows := []string{
		"Casa (1 Progetto)",
		"Lavoro (1 Progetto)",
		"Vuoto (nessun Progetto)",
	}
	if got := listTexts(d.list); !slices.Equal(got, wantRows) {
		t.Errorf("the list shows %v, want %v", got, wantRows)
	}

	// Without a selection Rinomina and Elimina are off, Nuovo and Chiudi stay on.
	if d.renameButton.IsEnabled() || d.deleteButton.IsEnabled() {
		t.Error("Rinomina or Elimina is on without a selection")
	}
	if !d.newButton.IsEnabled() || !d.closeButton.IsEnabled() {
		t.Error("Nuovo or Chiudi is off without a selection")
	}
	d.list.SetCurrentRow(1)
	settle()
	if !d.renameButton.IsEnabled() || !d.deleteButton.IsEnabled() {
		t.Error("Rinomina or Elimina stays off with a selection")
	}

	// The tab order follows the reading order, and a snapshot keeps the
	// selection.
	if got, want := controlLabels(d.controls), []string{
		groupsTitle, newGroupText, renameGroupText, deleteGroupText, closeGroupsText,
	}; !slices.Equal(got, want) {
		t.Errorf("the tab order of the dialog is %v, want %v", got, want)
	}
	s.apply(compose.Snapshot{DockerReachable: true, Projects: projects, Groups: groups})
	settle()
	if got := d.selectedName(); got != "Lavoro" {
		t.Errorf("the snapshot moved the selection to %q, want Lavoro", got)
	}

	// Nuovo opens the inline field with Conferma and Annulla and shuts every
	// other command, the list included.
	d.newButton.Click()
	settle()
	if !d.fieldOpen {
		t.Fatal("Nuovo opens no field")
	}
	if got := d.entry.Text(); got != "" {
		t.Errorf("the new name starts as %q, want empty", got)
	}
	if d.list.IsEnabled() || d.newButton.IsEnabled() || d.renameButton.IsEnabled() ||
		d.deleteButton.IsEnabled() || d.closeButton.IsEnabled() {
		t.Error("a command or the list stays on while the field is open")
	}
	if got, want := controlLabels(d.fieldControls), []string{
		groupNameLabelText, confirmGroupText, cancelFieldText,
	}; !slices.Equal(got, want) {
		t.Errorf("the tab order of the field is %v, want %v", got, want)
	}
	if got := d.dialog.FocusWidget(); got == nil || got.UnsafePointer() != d.entry.QWidget.UnsafePointer() {
		t.Error("the open field does not take the focus")
	}

	// A snapshot keeps the typed text.
	d.entry.SetText("lavoro")
	s.apply(compose.Snapshot{DockerReachable: true, Projects: projects, Groups: groups})
	settle()
	if got := d.entry.Text(); got != "lavoro" {
		t.Errorf("a snapshot changed the typed name to %q", got)
	}

	// A reserved name is refused under the field and in its accessible
	// description, and the field stays open.
	invalid, _ := view.KnownMessage(compose.ErrInvalidGroupName)
	d.entry.SetText(forms.NoGroupText)
	d.confirmField.Click()
	settle()
	if !d.fieldOpen {
		t.Error("a refused name closes the field")
	}
	if got := d.fieldMessage.Text(); got != invalid {
		t.Errorf("the refusal reads %q under the field, want %q", got, invalid)
	}
	if got := d.entry.AccessibleDescription(); !strings.Contains(got, invalid) {
		t.Errorf("the field says %q, want it to repeat %q", got, invalid)
	}

	// A duplicate is judged by the Manager and the message names the saved
	// Group; a duplicate without the name falls back to the generic message.
	d.op.Create = func(string, func(error)) {
		d.creationDone("lavoro", &model.DuplicateGroupError{Existing: "Lavoro"})
	}
	d.entry.SetText("lavoro")
	d.confirmField.Click()
	settle()
	if got, want := d.fieldMessage.Text(), "Esiste già il Gruppo «Lavoro»."; got != want {
		t.Errorf("the duplicate reads %q, want %q", got, want)
	}
	d.op.Create = func(string, func(error)) {
		d.creationDone("lavoro", &model.DuplicateGroupError{})
	}
	d.confirmField.Click()
	settle()
	generic, _ := view.KnownMessage(compose.ErrDuplicateGroup)
	if got := d.fieldMessage.Text(); got != generic {
		t.Errorf("a duplicate without a name reads %q, want the generic %q", got, generic)
	}

	// A creation that works closes the field, selects the new Group at the end
	// and goes on the records of the form.
	d.op.Create = func(string, func(error)) { d.creationDone("Nuovo", nil) }
	d.entry.SetText("  Nuovo  ")
	d.confirmField.Click()
	settle()
	if d.fieldOpen {
		t.Error("Conferma leaves the field open after a creation")
	}
	if got := d.selectedName(); got != "Nuovo" {
		t.Errorf("after the creation the selection is %q, want Nuovo", got)
	}
	if got := listTexts(d.list); !slices.Equal(got, append(wantRows, "Nuovo (nessun Progetto)")) {
		t.Errorf("the list shows %v after the creation, want the new Group at the end", got)
	}
	if got := d.changes[len(d.changes)-1]; got.Kind != forms.GroupCreated || got.From != "" || got.To != "Nuovo" {
		t.Errorf("the creation record is %+v, want an empty departure and Nuovo", got)
	}

	// Annulla always closes the field; Escape cancels it and keeps the dialog.
	d.list.SetCurrentRow(0)
	settle()
	d.renameButton.Click()
	settle()
	if got := d.entry.Text(); got != "Casa" {
		t.Errorf("Rinomina starts from %q, want the selected Group Casa", got)
	}
	d.entry.SetText("Scartato")
	d.cancelField.Click()
	settle()
	if d.fieldOpen {
		t.Error("Annulla leaves the field open")
	}
	if !d.list.IsEnabled() || !d.newButton.IsEnabled() {
		t.Error("Annulla leaves the commands off")
	}
	d.renameButton.Click()
	settle()
	pressDialogKey(d, qt6.Key_Escape)
	settle()
	if d.fieldOpen {
		t.Error("Escape does not cancel the open field")
	}
	if d.closed {
		t.Error("Escape with the field open closes the dialog")
	}

	// A plain rename keeps the Group in its place and the selection on it.
	d.op.Rename = func(string, string, func(error)) { d.renameDone("Casa", "Ufficio", false, nil) }
	d.renameButton.Click()
	settle()
	d.entry.SetText("Ufficio")
	d.confirmField.Click()
	settle()
	if got := d.selectedName(); got != "Ufficio" {
		t.Errorf("after the rename the selection is %q, want Ufficio", got)
	}
	if got := listTexts(d.list); !slices.Equal(got, []string{
		"Ufficio (1 Progetto)", "Lavoro (1 Progetto)", "Vuoto (nessun Progetto)", "Nuovo (nessun Progetto)",
	}) {
		t.Errorf("the list shows %v after the rename", got)
	}
	if got := d.changes[len(d.changes)-1]; got.Kind != forms.GroupRenamed || got.From != "Casa" || got.To != "Ufficio" {
		t.Errorf("the rename record is %+v, want Casa to Ufficio", got)
	}

	// A rename onto a saved Group asks for a union with the count of the
	// Projects that move, and the selection goes to the survivor.
	var question guiproject.Confirmation
	asked := false
	d.onConfirm = func(q guiproject.Confirmation, run func()) { asked = true; question = q; run() }
	d.op.Rename = func(string, string, func(error)) { d.renameDone("Ufficio", "Lavoro", true, nil) }
	d.renameButton.Click()
	settle()
	d.entry.SetText("Lavoro")
	d.confirmField.Click()
	settle()
	if !asked {
		t.Error("the union asks no confirmation")
	}
	if !strings.Contains(question.Message, "1 Progetto") {
		t.Errorf("the union says %q, want the count of the Projects", question.Message)
	}
	if got := d.selectedName(); got != "Lavoro" {
		t.Errorf("after the union the selection is %q, want the survivor Lavoro", got)
	}
	if got := listTexts(d.list); !slices.Equal(got, []string{
		"Lavoro (2 Progetti)", "Vuoto (nessun Progetto)", "Nuovo (nessun Progetto)",
	}) {
		t.Errorf("the list shows %v after the union", got)
	}
	if got := d.changes[len(d.changes)-1]; got.Kind != forms.GroupUnited || got.From != "Ufficio" || got.To != "Lavoro" {
		t.Errorf("the union record is %+v, want Ufficio to Lavoro", got)
	}

	// An error that does not concern the field shows under the list and keeps
	// the field open.
	d.op.Rename = func(string, string, func(error)) { d.renameDone("Vorrei", "X", false, compose.ErrUnknownGroup) }
	d.list.SetCurrentRow(1)
	settle()
	d.renameButton.Click()
	settle()
	d.entry.SetText("X")
	d.confirmField.Click()
	settle()
	if got, want := d.errorLine.Text(), "Gruppo sconosciuto."; got != want {
		t.Errorf("the unknown Group reads %q under the list, want %q", got, want)
	}
	if !d.fieldOpen {
		t.Error("a refused rename closes the field")
	}
	d.cancelField.Click()
	settle()

	// Chiudi gives the form its field back with the new Group, and the focus
	// returns to the command that opened the dialog.
	d.closeButton.Click()
	settle()
	if !d.closed {
		t.Fatal("Chiudi leaves the dialog open")
	}
	if s.groups != nil {
		t.Error("the shell keeps the closed dialog")
	}
	if blockNamed(s.panel, editTitle) == nil {
		t.Fatal("closing the dialog closes the form")
	}
	if got := s.panel.edit.group.CurrentText(); got != "Lavoro" {
		t.Errorf("the Group field shows %q after the close, want the survivor Lavoro", got)
	}
	if got := s.panel.edit.name.Text(); got != shop.DisplayName {
		t.Errorf("the closing dialog changed the Display name to %q", got)
	}
	if got, want := itemsOf(s.panel.edit.group), forms.GroupSuggestions([]string{"", "Lavoro", "Vuoto", "Nuovo"}); !slices.Equal(got, want) {
		t.Errorf("the Group field proposes %v after the close, want %v", got, want)
	}
	if !s.panel.edit.edit.ProfileOn("debug") {
		t.Error("the closing dialog turned the Profile the user checked off")
	}
	form = blockNamed(s.panel, editTitle)
	commands = widgetsNamed(form, "editGroups")
	if len(commands) != 1 {
		t.Fatal("the form carries no command of the Groups after the close")
	}
	if got := s.window.FocusWidget(); got == nil || got.UnsafePointer() != commands[0].UnsafePointer() {
		t.Error("the close does not give the focus back to the command of the Groups")
	}

	// The snapshot that follows the operations brings the same list to the
	// window, as the Manager publishes it.
	shopAfter, otherAfter := shop, other
	shopAfter.Group, otherAfter.Group = "Lavoro", "Lavoro"
	s.apply(compose.Snapshot{
		DockerReachable: true,
		Projects:        []compose.Project{shopAfter, otherAfter},
		Groups:          []string{"Lavoro", "Vuoto", "Nuovo"},
	})
	settle()

	// The list of the Projects follows the same order, and a Group without
	// Projects leaves no heading.
	rows := listRows(s.list.widget)
	if len(rows) == 0 || rows[0] != "Lavoro" {
		t.Errorf("the list headings after the operations are %v, want Lavoro first", rows)
	}
	for _, empty := range []string{"Vuoto", "Nuovo"} {
		if slices.Contains(rows, empty) {
			t.Errorf("the list shows the heading %q of a Group without Projects: %v", empty, rows)
		}
	}

	// Deletion: the question says how many Projects go to Senza Gruppo, the
	// selection takes the empty place, and the field of the form empties.
	qt6.UnsafeNewQPushButton(commands[0].UnsafePointer()).Click()
	settle()
	d = s.groups
	if d == nil {
		t.Fatal("the command opens no dialog the second time")
	}
	if got, want := listTexts(d.list), []string{
		"Lavoro (2 Progetti)", "Vuoto (nessun Progetto)", "Nuovo (nessun Progetto)",
	}; !slices.Equal(got, want) {
		t.Errorf("the second dialog shows %v, want %v", got, want)
	}
	d.onConfirm = func(q guiproject.Confirmation, run func()) { question = q; run() }
	d.op.Delete = func(string, func(error)) { d.deleteDone("Lavoro", nil) }
	d.list.SetCurrentRow(0)
	settle()
	d.deleteButton.Click()
	settle()
	if !strings.Contains(question.Message, "2 Progetti") {
		t.Errorf("the deletion says %q, want the count of the Projects", question.Message)
	}
	if got := d.selectedName(); got != "Vuoto" {
		t.Errorf("after the deletion the selection is %q, want the Group that took the place", got)
	}
	if got := listTexts(d.list); !slices.Equal(got, []string{
		"Vuoto (nessun Progetto)", "Nuovo (nessun Progetto)",
	}) {
		t.Errorf("the list shows %v after the deletion", got)
	}
	if got := d.changes[len(d.changes)-1]; got.Kind != forms.GroupDeleted || got.From != "Lavoro" || got.To != "" {
		t.Errorf("the deletion record is %+v, want Lavoro with an empty arrival", got)
	}

	// During an operation every command is off and a second click starts
	// nothing. Escape and Chiudi are ignored until the outcome arrives.
	var blocked func(error)
	calls := 0
	d.op.Delete = func(string, func(error)) { calls++; blocked = func(err error) { d.deleteDone("Vuoto", err) } }
	d.list.SetCurrentRow(0)
	settle()
	d.deleteButton.Click()
	settle()
	if !d.busy {
		t.Fatal("the dialog is not busy during the operation")
	}
	for _, button := range []*qt6.QPushButton{d.newButton, d.renameButton, d.deleteButton, d.closeButton} {
		if button.IsEnabled() {
			t.Errorf("the command %q is on during the operation", button.Text())
		}
	}
	if d.list.IsEnabled() {
		t.Error("the list accepts clicks during the operation")
	}
	d.deleteButton.Click()
	settle()
	if calls != 1 {
		t.Errorf("a second click started %d operations, want 1", calls)
	}
	pressDialogKey(d, qt6.Key_Escape)
	settle()
	if d.closed {
		t.Error("Escape closes the dialog during an operation")
	}
	d.closeButton.Click()
	settle()
	if d.closed {
		t.Error("Chiudi closes the dialog during an operation")
	}
	blocked(nil)
	settle()
	if d.busy {
		t.Error("the dialog stays busy after the outcome")
	}
	if !d.newButton.IsEnabled() || !d.closeButton.IsEnabled() {
		t.Error("the commands stay off after the outcome")
	}

	// The dialog meets the target of 24 px on every control.
	checkGroupDialogTargets(t, d.dialog)

	d.closeButton.Click()
	settle()
	if got := s.panel.edit.group.CurrentText(); got != "" {
		t.Errorf("the Group field shows %q after the deletion, want empty", got)
	}
	shopAfter.Group, otherAfter.Group = "", ""
	s.apply(compose.Snapshot{
		DockerReachable: true,
		Projects:        []compose.Project{shopAfter, otherAfter},
		Groups:          []string{"Vuoto", "Nuovo"},
	})
	settle()

	// Escape without an open field closes the dialog, and the focus returns to
	// the command of the Groups.
	commands = widgetsNamed(blockNamed(s.panel, editTitle), "editGroups")
	d = nil
	qt6.UnsafeNewQPushButton(commands[0].UnsafePointer()).Click()
	settle()
	d = s.groups
	if d == nil {
		t.Fatal("the command opens no dialog the third time")
	}
	pressDialogKey(d, qt6.Key_Escape)
	settle()
	if !d.closed {
		t.Error("Escape without an open field leaves the dialog open")
	}
	if s.groups != nil {
		t.Error("the shell keeps the dialog Escape closed")
	}
	commands = widgetsNamed(blockNamed(s.panel, editTitle), "editGroups")
	if got := s.window.FocusWidget(); got == nil || got.UnsafePointer() != commands[0].UnsafePointer() {
		t.Error("Escape does not give the focus back to the command of the Groups")
	}

	// The step of the form leaves the panel as it found it.
	s.panel.cancelEdit()
	settle()
}

// listTexts lists the text of every row of a list.
func listTexts(list *qt6.QListWidget) []string {
	out := make([]string, 0, list.Count())
	for i := 0; i < list.Count(); i++ {
		if item := list.Item(i); item != nil {
			out = append(out, item.Text())
		}
	}
	return out
}

// controlLabels lists the labels of a list of controls, in order.
func controlLabels(controls []panelControl) []string {
	out := make([]string, 0, len(controls))
	for _, control := range controls {
		out = append(out, control.label)
	}
	return out
}

// pressDialogKey sends a key press to a dialog, as the keyboard would.
func pressDialogKey(d *groupsDialog, key qt6.Key) {
	handle := d.dialog.WindowHandle()
	if handle == nil {
		return
	}
	event := qt6.NewQKeyEvent(qt6.QEvent__KeyPress, int(key), qt6.NoModifier)
	qt6.QCoreApplication_SendEvent(handle.QObject, event.QInputEvent.QEvent)
	event.Delete()
}

// checkGroupDialogTargets checks the 24 px target on every control of a dialog.
func checkGroupDialogTargets(t *testing.T, dialog *qt6.QDialog) {
	t.Helper()

	checked := 0
	for _, widget := range qt6.QApplication_AllWidgets() {
		if widget.Window().UnsafePointer() != dialog.QWidget.UnsafePointer() {
			continue
		}
		if widget.FocusPolicy() == qt6.NoFocus {
			continue
		}
		widget.EnsurePolished()
		hint, minimum := widget.SizeHint(), widget.MinimumSizeHint()
		width := max(hint.Width(), minimum.Width())
		height := max(hint.Height(), minimum.Height())
		if width < design.TargetMin || height < design.TargetMin {
			t.Errorf("control %q: target %dx%d, want at least %dx%d",
				widget.ObjectName(), width, height, design.TargetMin, design.TargetMin)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no interactive control found in the dialog: the wiring is broken")
	}
}

// pressKey sends a key press to the window, as the keyboard would.
func pressKey(s *shell, key qt6.Key) {
	handle := s.window.WindowHandle()
	if handle == nil {
		return
	}
	event := qt6.NewQKeyEvent(qt6.QEvent__KeyPress, int(key), qt6.NoModifier)
	qt6.QCoreApplication_SendEvent(handle.QObject, event.QInputEvent.QEvent)
	event.Delete()
}

// listNames lists the names the rows of the list show, in their order.
func listNames(list *qt6.QListWidget) []string {
	var names []string
	for i := 0; ; i++ {
		item := list.Item(i)
		if item == nil {
			return names
		}
		row := list.ItemWidget(item)
		for _, label := range widgetsNamed(row, "projectName") {
			names = append(names, qt6.UnsafeNewQLabel(label.UnsafePointer()).Text())
		}
	}
}

// itemsOf lists the options of an editable box, in their order.
func itemsOf(box *qt6.QComboBox) []string {
	items := make([]string, 0, box.Count())
	for i := 0; i < box.Count(); i++ {
		items = append(items, box.ItemText(i))
	}
	return items
}

// formOrder is the order the keyboard walks the form in: the labels of the
// controls that belong to the form, in the order the tab chain follows.
func formOrder(p *detailPanel, project compose.Project) []string {
	fields := map[string]bool{
		displayNameLabelText: true,
		groupLabelText:       true,
		groupsCommandText:    true,
		cancelText:           true,
		saveText:             true,
	}
	var order []string
	for _, control := range p.controls {
		if fields[control.label] || forms.HasProfile(project.Profiles, control.label) {
			order = append(order, control.label)
		}
	}
	return order
}

// errorUnder is the message the form shows under the field it concerns.
func errorUnder(form *qt6.QWidget) string {
	for _, label := range widgetsNamed(form, "fieldError") {
		if text := qt6.UnsafeNewQLabel(label.UnsafePointer()).Text(); text != "" {
			return text
		}
	}
	return ""
}

// checkConfirmDialog builds the question of Down con volumi and checks the
// rules of the dialog: modal, affirmative button as default, text to type, and
// a rejection that runs nothing. The question never opens in a blocking way.
func checkConfirmDialog(t *testing.T, s *shell) {
	t.Helper()

	question := guiproject.Confirmation{
		Title:         "Down con volumi di shop",
		Message:       "Rimuove i Container e ELIMINA i volumi del Progetto: i dati nei volumi andranno persi per sempre.",
		ConfirmLabel:  "Down con volumi",
		TypeToConfirm: "shop",
		Destructive:   true,
	}
	ran := false
	dialog := newConfirmDialog(s.window.QWidget, question, func() { ran = true })

	if !dialog.dialog.IsModal() {
		t.Error("the question is not modal")
	}
	if got := dialog.dialog.WindowTitle(); got != question.Title {
		t.Errorf("the question is titled %q, want %q", got, question.Title)
	}
	if !dialog.confirm.IsDefault() {
		t.Error("the affirmative button is not the default of the question")
	}
	if dialog.confirm.IsEnabled() {
		t.Error("the affirmative button is on before the name is typed")
	}
	if dialog.entry == nil {
		t.Fatal("the question asks for no text to type")
	}
	dialog.entry.SetText("sho")
	if dialog.confirm.IsEnabled() {
		t.Error("the affirmative button is on with a partial name")
	}
	dialog.entry.SetText(question.TypeToConfirm)
	if !dialog.confirm.IsEnabled() {
		t.Error("the affirmative button stays off with the name typed")
	}

	dialog.dialog.Reject()
	if ran {
		t.Error("a rejected question ran the Action")
	}
	settle()

	// The affirmative answer runs the Action.
	dialog = newConfirmDialog(s.window.QWidget, question, func() { ran = true })
	dialog.entry.SetText(question.TypeToConfirm)
	dialog.dialog.Accept()
	if !ran {
		t.Error("the affirmative answer ran nothing")
	}
	settle()
}

// checkCommands drives the commands of the detail and the entry that takes a
// Project off the list: the three commands follow the Project, the reason of an
// insensitive one reaches the tooltip and the screen reader, and the removal asks
// its question and touches no Docker. It is a step of the smoke test.
func checkCommands(t *testing.T, s *shell) {
	t.Helper()

	project := fakeProject()
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{project}})
	s.open(project)
	settle()

	block := blockNamed(s.panel, "Comandi del Progetto")
	if block == nil {
		t.Fatal("the panel has no block of commands")
	}
	for _, name := range []string{"folderButton", "terminalButton", "portButton"} {
		button := commandButton(t, block, name)
		if !button.IsEnabled() {
			t.Errorf("the command %q is insensitive on a complete Project", name)
		}
		if got := button.AccessibleDescription(); got != "" {
			t.Errorf("the command %q says %q on a complete Project", name, got)
		}
	}

	// A press reaches the shell with the name of the Project and its first
	// published port. The callbacks are replaced, so nothing leaves the window.
	var folder, terminal string
	var port int
	folderCallback, terminalCallback, portCallback := s.panel.onFolder, s.panel.onTerminal, s.panel.onPort
	s.panel.onFolder = func(name string) { folder = name }
	s.panel.onTerminal = func(name string) { terminal = name }
	s.panel.onPort = func(number int) { port = number }
	commandButton(t, block, "folderButton").Click()
	commandButton(t, block, "terminalButton").Click()
	commandButton(t, block, "portButton").Click()
	settle()
	if folder != project.Name || terminal != project.Name {
		t.Errorf("the commands opened %q and %q, want %q", folder, terminal, project.Name)
	}
	if want, _ := view.FirstPublishedPort(project); port != want {
		t.Errorf("the port command opened %d, want %d", port, want)
	}
	s.panel.onFolder, s.panel.onTerminal, s.panel.onPort = folderCallback, terminalCallback, portCallback

	// A Project without a folder opens neither the folder nor a terminal, and a
	// Project without a published port opens no browser.
	noFolder := project
	noFolder.Folder = ""
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{noFolder}})
	s.open(noFolder)
	settle()
	block = blockNamed(s.panel, "Comandi del Progetto")
	for _, name := range []string{"folderButton", "terminalButton"} {
		button := commandButton(t, block, name)
		if button.IsEnabled() {
			t.Errorf("the command %q is sensitive on a Project without a folder", name)
		}
		if got := button.AccessibleDescription(); got != view.NoFolderCommandText {
			t.Errorf("the command %q says %q, want %q", name, got, view.NoFolderCommandText)
		}
	}
	if !commandButton(t, block, "portButton").IsEnabled() {
		t.Error("the port command is insensitive on a Project without a folder")
	}

	noPort := project
	noPort.Services = []compose.Service{{
		Name:       "web",
		Containers: []compose.Container{{Name: "shop-web-1", State: "running"}},
	}}
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{noPort}})
	s.open(noPort)
	settle()
	block = blockNamed(s.panel, "Comandi del Progetto")
	if got := commandButton(t, block, "portButton").AccessibleDescription(); got != view.NoPortCommandText {
		t.Errorf("the port command says %q, want %q", got, view.NoPortCommandText)
	}

	// A missing File compose closes every command.
	missing := noPort
	missing.ComposeMissing = true
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{missing}})
	s.open(missing)
	settle()
	block = blockNamed(s.panel, "Comandi del Progetto")
	for _, name := range []string{"folderButton", "terminalButton", "portButton"} {
		button := commandButton(t, block, name)
		if button.IsEnabled() {
			t.Errorf("the command %q is sensitive while the File compose is missing", name)
		}
		if got := button.AccessibleDescription(); got != view.MissingComposeText {
			t.Errorf("the command %q says %q, want %q", name, got, view.MissingComposeText)
		}
	}

	// The removal asks its question, and only a yes hands the name to the shell.
	removable := compose.Project{
		Name: "docs", DisplayName: "Documentazione",
		Folder: "/srv/docs", ComposeFiles: []string{"/srv/docs/docker-compose.yml"},
		ComposeMissing: true, Status: compose.StatusRemoved,
	}
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{removable}})
	s.open(removable)
	settle()

	var removed string
	previousRemove := s.panel.onRemove
	s.panel.onRemove = func(name string) { removed = name }
	defer func() { s.panel.onRemove = previousRemove }()

	remove := removeCommand(t, s)
	if !remove.IsEnabled() {
		t.Fatal("the remove entry is insensitive on a Rimosso Project without Container")
	}
	remove.Click()
	settle()
	if s.panel.question == nil {
		t.Fatal("the remove entry opens no question")
	}
	want := guiproject.RemoveConfirmation(removable)
	if got := s.panel.question.dialog.WindowTitle(); got != want.Title {
		t.Errorf("the question is titled %q, want %q", got, want.Title)
	}
	s.panel.question.dialog.Reject()
	settle()
	if removed != "" {
		t.Errorf("a rejected question removed %q", removed)
	}
	if s.panel.guard.waiting() {
		t.Error("the removal took the guard of the Actions: it launched one")
	}

	removeCommand(t, s).Click()
	settle()
	if s.panel.question == nil {
		t.Fatal("the remove entry opens no question on the second press")
	}
	s.panel.question.dialog.Accept()
	settle()
	if removed != removable.Name {
		t.Errorf("the confirmed removal handed %q, want %q", removed, removable.Name)
	}
}

// commandButton is the command of a block with the object name of the style
// sheet. A block carries one of each.
func commandButton(t *testing.T, block *qt6.QWidget, name string) *qt6.QPushButton {
	t.Helper()
	found := widgetsNamed(block, name)
	if len(found) != 1 {
		t.Fatalf("the block carries %d commands named %q, want 1", len(found), name)
	}
	return qt6.UnsafeNewQPushButton(found[0].UnsafePointer())
}

// removeCommand is the entry that takes the Project off the list, from the notice
// of its missing File compose.
func removeCommand(t *testing.T, s *shell) *qt6.QPushButton {
	t.Helper()
	block := blockNamed(s.panel, "File compose mancante")
	if block == nil {
		t.Fatal("the missing File compose shows no notice")
	}
	found := widgetsNamed(block, "removeButton")
	if len(found) != 1 {
		t.Fatalf("the notice carries %d remove commands, want 1", len(found))
	}
	return qt6.UnsafeNewQPushButton(found[0].UnsafePointer())
}

// checkPicker opens the picker of the File compose and checks the shape the
// ticket asks for: more than one file, the filter of the compose files and the
// widget dialog of Qt, which carries the action "Show hidden files". It is a
// step of the smoke test, so the process still has one QApplication.
func checkPicker(t *testing.T, s *shell) {
	t.Helper()

	picked := false
	pickComposeFiles(s.window.QWidget, func([]string) { picked = true })
	settle()

	active := qt6.QApplication_ActiveModalWidget()
	if active == nil {
		t.Fatal("the picker did not open as a modal dialog")
	}
	picker := qt6.UnsafeNewQFileDialog(active.UnsafePointer())
	if got := picker.WindowTitle(); got != pickFilesText {
		t.Errorf("the picker is titled %q, want %q", got, pickFilesText)
	}
	if got := picker.FileMode(); got != qt6.QFileDialog__ExistingFiles {
		t.Errorf("the picker accepts %d, want more than one existing file", got)
	}
	if !slices.Contains(picker.NameFilters(), composeFilterText) {
		t.Errorf("the picker offers %q, want %q", picker.NameFilters(), composeFilterText)
	}
	if !picker.TestOption(qt6.QFileDialog__DontUseNativeDialog) {
		t.Error("the picker is the native dialog, which has no action for the hidden files")
	}

	picker.Reject()
	settle()
	if picked {
		t.Error("a closed picker handed a choice")
	}
}

// checkAddForm drives the form of a new Project: the chosen File compose keep the
// order of the picker, the first one is the main one, a refusal keeps the form
// open with its Italian text, and a save closes it and names the Project. It is a
// step of the smoke test.
func checkAddForm(t *testing.T, s *shell) {
	t.Helper()

	dir := t.TempDir()
	main := filepath.Join(dir, "docker-compose.yml")
	override := filepath.Join(dir, "compose.override.yml")
	for _, file := range []string{main, override} {
		if err := os.WriteFile(file, []byte("name: shop\nservices: {}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var form forms.AddProjectForm
	if _, err := form.Pick([]string{override, main}, compose.ProposeName); err != nil {
		t.Fatalf("Pick error = %v", err)
	}

	refusal := error(compose.ErrDuplicateProject)
	var saved string
	dialog := newAddDialog(s.window.QWidget, form,
		func(name string, files []string, done func(error)) {
			if name != "shop" || !slices.Equal(files, form.Files) {
				t.Errorf("the form sent %q and %q, want %q and %q", name, files, "shop", form.Files)
			}
			done(refusal)
		},
		func(name string) { saved = name })
	dialog.open()
	settle()

	if got := dialog.dialog.WindowTitle(); got != addProjectTitle {
		t.Errorf("the form is titled %q, want %q", got, addProjectTitle)
	}
	if !dialog.dialog.IsModal() {
		t.Error("the form of a new Project is not modal")
	}
	if got := dialog.name.Text(); got != "shop" {
		t.Errorf("the name field reads %q, want the proposal %q", got, "shop")
	}
	if !dialog.add.IsEnabled() {
		t.Error("the confirm command is off with a File compose and a name")
	}

	// The rows follow the order of the picker, and the first file is the main
	// one whatever its name: compose reads -f in that order.
	var rows, roles []string
	for _, label := range widgetsNamed(dialog.list, "fileComposeName") {
		rows = append(rows, qt6.UnsafeNewQLabel(label.UnsafePointer()).Text())
	}
	for _, label := range widgetsNamed(dialog.list, "fileComposeRole") {
		roles = append(roles, qt6.UnsafeNewQLabel(label.UnsafePointer()).Text())
	}
	if !slices.Equal(rows, []string{"compose.override.yml", "docker-compose.yml"}) {
		t.Errorf("the rows read %q, want the order of the picker", rows)
	}
	if !slices.Equal(roles, []string{mainFileRoleText, extraFileRoleText}) {
		t.Errorf("the roles read %q, want %q then %q", roles, mainFileRoleText, extraFileRoleText)
	}

	// A name already saved keeps the form open with its text.
	dialog.submit()
	settle()
	if got := errorUnder(dialog.dialog.QWidget); got != "Esiste già un Progetto con questo nome." {
		t.Errorf("the refusal reads %q, want the text of the duplicate name", got)
	}
	if !dialog.add.IsEnabled() {
		t.Error("the confirm command stays off after a refusal")
	}
	if saved != "" {
		t.Errorf("a refused form saved %q", saved)
	}

	// A save that compose refuses keeps the form open with its text too.
	refusal = compose.ErrInvalidProject
	dialog.submit()
	settle()
	if got := errorUnder(dialog.dialog.QWidget); got != "Progetto non valido: i File compose indicati non sono utilizzabili." {
		t.Errorf("the refusal reads %q, want the text of an unusable Project", got)
	}

	// A save that works closes the form and names the Project for the shell.
	refusal = nil
	dialog.submit()
	settle()
	if saved != "shop" {
		t.Errorf("the saved Project is %q, want %q", saved, "shop")
	}

	// A form without a File compose asks for one first, and its name field is
	// insensitive.
	empty := newAddDialog(s.window.QWidget, forms.AddProjectForm{},
		func(string, []string, func(error)) { t.Error("an empty form sent a request") },
		func(string) { t.Error("an empty form saved a Project") })
	defer empty.dialog.Delete()
	empty.open()
	settle()
	if empty.name.IsEnabled() {
		t.Error("the name field is sensitive before a File compose is chosen")
	}
	if empty.add.IsEnabled() {
		t.Error("the confirm command is on with no File compose and no name")
	}

	// Dropping the main file moves the first place down the order and proposes
	// the name again.
	var second forms.AddProjectForm
	if _, err := second.Pick([]string{main, override}, compose.ProposeName); err != nil {
		t.Fatalf("Pick error = %v", err)
	}
	again := newAddDialog(s.window.QWidget, second,
		func(string, []string, func(error)) {},
		func(string) {})
	defer again.dialog.Delete()
	again.open()
	settle()
	again.removeFile(0)
	settle()
	rows = nil
	for _, label := range widgetsNamed(again.list, "fileComposeName") {
		rows = append(rows, qt6.UnsafeNewQLabel(label.UnsafePointer()).Text())
	}
	if !slices.Equal(rows, []string{"compose.override.yml"}) {
		t.Errorf("the rows read %q after dropping the main file", rows)
	}
	roles = nil
	for _, label := range widgetsNamed(again.list, "fileComposeRole") {
		roles = append(roles, qt6.UnsafeNewQLabel(label.UnsafePointer()).Text())
	}
	if !slices.Equal(roles, []string{mainFileRoleText}) {
		t.Errorf("the remaining file is %q, want the main role", roles)
	}
	if got := again.name.Text(); got != "shop" {
		t.Errorf("the name field reads %q after the main file changed, want %q", got, "shop")
	}
}

// checkBottomPanel builds the panel of the Activities over a log of fake
// Activities and checks the two tabs, the history, the output that grows while
// the command runs, the scroll the history keeps, and the notice of a failure
// with its way to the output.
func checkBottomPanel(t *testing.T) {
	t.Helper()

	log := activity.NewLog()
	s := newShell(&application{activities: log}, design.LightTheme)
	defer s.window.Delete()
	s.show()
	settle()

	if s.bottom == nil || s.bottom.widget == nil {
		t.Fatal("the bottom panel is not wired")
	}
	if got := s.column.Orientation(); got != qt6.Vertical {
		t.Fatalf("the column is %d, want a vertical splitter", got)
	}
	if got := s.column.Count(); got != 2 {
		t.Fatalf("the column carries %d widgets, want the detail and the panel", got)
	}
	if got := s.column.Widget(0); got == nil || got.UnsafePointer() != s.detail.UnsafePointer() {
		t.Error("the column does not start with the detail")
	}
	if got := s.column.Widget(1); got == nil || got.UnsafePointer() != s.bottom.widget.UnsafePointer() {
		t.Error("the column does not hold the panel under the detail")
	}
	if got := s.bottom.tabs.Count(); got != 2 {
		t.Fatalf("the panel carries %d tabs, want 2", got)
	}
	for i, want := range []string{activity.ActivityPageTitle, activity.LogPageTitle} {
		if got := s.bottom.tabs.TabText(i); got != want {
			t.Errorf("tab %d is %q, want %q", i, got, want)
		}
	}
	// The Log tab carries the page of the logs, with its bar over the text.
	page := s.bottom.tabs.Widget(logTab)
	if page == nil || page.UnsafePointer() != s.bottom.logs.widget.UnsafePointer() {
		t.Fatal("the Log tab does not carry the page of the logs")
	}
	if !s.bottom.logs.view.IsReadOnly() {
		t.Error("the text of the logs is editable")
	}
	if got := s.bottom.logs.searchLabel.Text(); got != activity.LogSearchLabelText {
		t.Errorf("the search label reads %q, want %q", got, activity.LogSearchLabelText)
	}
	if got := s.bottom.logs.status.Text(); got != activity.EmptyLogText {
		t.Errorf("the empty log reads %q, want %q", got, activity.EmptyLogText)
	}
	if s.bottom.showing() || s.bottom.widget.IsVisible() {
		t.Fatal("the panel takes room before the first Activity")
	}
	if got := s.bottom.page.output.PlaceholderText(); got != activity.EmptyText {
		t.Errorf("the empty output reads %q, want %q", got, activity.EmptyText)
	}

	up := activity.Activity{ID: "act-1", Project: "shop", Kind: compose.ActionUp, State: activity.ActivityRunning}
	log.Push(guitest.Started(up.ID, up.Project, "", up.Kind))
	for i := 0; i < 3; i++ {
		log.Push(guitest.Output(up.ID, up.Project, "", up.Kind, fmt.Sprintf("step %d", i), false))
	}
	settle()

	if !s.bottom.showing() || !s.bottom.widget.IsVisible() {
		t.Fatal("the first Activity does not open the panel")
	}
	if got := s.bottom.tabs.CurrentIndex(); got != activityTab {
		t.Errorf("the panel opens on tab %d, want the Activities", got)
	}
	if got := s.bottom.page.list.Count(); got != 1 {
		t.Fatalf("the history carries %d rows, want 1", got)
	}
	if got := s.bottom.page.list.Item(0).Text(); got != activity.ActivityHeader(up) {
		t.Errorf("the row reads %q, want %q", got, activity.ActivityHeader(up))
	}
	if got := s.bottom.page.list.Item(0).ToolTip(); got != activity.RunningText {
		t.Errorf("the row says %q, want %q", got, activity.RunningText)
	}
	if got := s.bottom.page.selected; got != up.ID {
		t.Errorf("the panel shows %q, want the running Activity", got)
	}
	if out := s.bottom.page.output.ToPlainText(); !strings.Contains(out, "step 2") {
		t.Errorf("the output is %q, want the lines of the Activity", out)
	}

	// The output grows while the command runs: the row stays, the text grows.
	log.Push(guitest.Output(up.ID, up.Project, "", up.Kind, "step 3", false))
	settle()
	if got := s.bottom.page.list.Count(); got != 1 {
		t.Errorf("a new line rebuilt the history: %d rows, want 1", got)
	}
	if out := s.bottom.page.output.ToPlainText(); !strings.Contains(out, "step 3") {
		t.Errorf("the output is %q, want the line that arrived while the command runs", out)
	}

	// The history keeps its scroll while the output changes.
	for i := 0; i < 40; i++ {
		log.Push(guitest.Started(compose.ActivityID(fmt.Sprintf("bulk-%02d", i)), "shop", "", compose.ActionStop))
	}
	settle()
	bar := s.bottom.page.list.VerticalScrollBar()
	if bar == nil || bar.Maximum() == 0 {
		t.Fatal("the history of forty Activities does not scroll")
	}
	bar.SetValue(bar.Maximum() / 2)
	value := bar.Value()
	log.Push(guitest.Output("bulk-39", "shop", "", compose.ActionStop, "late line", false))
	settle()
	if got := bar.Value(); got != value {
		t.Errorf("scroll = %d after a new line, want %d", got, value)
	}

	// A failed Activity shows its message in the notice, with the way to its
	// output.
	log.Push(guitest.Started("act-bad", "blog", "db", compose.ActionStart))
	log.Push(guitest.Output("act-bad", "blog", "db", compose.ActionStart, "creating", false))
	log.Push(guitest.Finished("act-bad", "blog", "db", compose.ActionStart, "error: no such image\n", "error: no such image"))
	settle()

	bad := activity.Activity{
		ID:      "act-bad",
		Project: "blog",
		Service: "db",
		Kind:    compose.ActionStart,
		State:   activity.ActivityFailed,
		Error:   "error: no such image\n",
		Summary: "error: no such image",
	}

	if s.notice == nil || !s.notice.frame.IsVisible() {
		t.Fatal("a failed Activity shows no notice")
	}
	if got, want := s.notice.label.Text(), activity.ActivityToastText(bad); got != want {
		t.Errorf("the notice reads %q, want %q", got, want)
	}
	if !s.notice.action.IsVisible() {
		t.Fatal("the notice of a failure carries no way to the output")
	}
	if got := s.notice.action.Text(); got != showOutputText {
		t.Errorf("the action reads %q, want %q", got, showOutputText)
	}

	// The action leads to the output: it opens the panel on its tab and on the
	// Activity that failed.
	s.bottom.Close()
	settle()
	s.notice.action.Click()
	settle()
	if !s.bottom.showing() {
		t.Error("the action of the notice leaves the panel closed")
	}
	if got := s.bottom.tabs.CurrentIndex(); got != activityTab {
		t.Errorf("the action opens tab %d, want the Activities", got)
	}
	if got := s.bottom.page.selected; got != bad.ID {
		t.Errorf("the action selects %q, want %q", got, bad.ID)
	}
	if out := s.bottom.page.output.ToPlainText(); !strings.Contains(out, "error: no such image") {
		t.Errorf("the output is %q, want the error of the Activity", out)
	}

	// A message without an action keeps the notice as it was: no button.
	s.showNotice(guiproject.ActivityRunningText)
	settle()
	if s.notice.action.IsVisible() {
		t.Error("a notice without an action shows the button of the previous one")
	}
	if got := s.notice.label.Text(); got != guiproject.ActivityRunningText {
		t.Errorf("the notice reads %q, want %q", got, guiproject.ActivityRunningText)
	}

	checkMinimumTargets(t, s.window)
}

// checkLogs drives the Log tab over a fake stream: the tab of the selected
// Project, a Service of the detail, the colours of the roles, the search, the
// pause, the timestamps, the cap and the end of the process when the panel
// closes or the Project changes.
func checkLogs(t *testing.T) {
	t.Helper()

	log := activity.NewLog()
	s := newShell(&application{activities: log}, design.LightTheme)
	defer s.window.Delete()
	s.show()
	settle()

	projects := screenshotProjects()
	project := projects[0] // vakt: web, db and cache

	lines := make(chan compose.LogLine, 2*logs.LogLineLimit)
	stopped := make(chan struct{}, 8)
	s.bottom.logs.ctx = context.Background()
	s.bottom.logs.source = func(ctx context.Context, _ compose.LogRequest, onLine func(compose.LogLine)) error {
		for {
			select {
			case <-ctx.Done():
				stopped <- struct{}{}
				return nil
			case line := <-lines:
				onLine(line)
			}
		}
	}

	s.apply(compose.Snapshot{DockerReachable: true, Projects: projects})
	s.open(project)
	settle()

	// A closed panel starts nothing: the stream begins with the Log tab.
	if s.bottom.logs.Running() {
		t.Error("the Log page streams while the panel is closed")
	}

	if !hasWidgetNamed(blockNamed(s.panel, "Intestazione del Progetto"), "logButton") {
		t.Error("the head of the Project carries no way to its logs")
	}
	services := blockNamed(s.panel, "Servizi del Progetto")
	if got, want := len(widgetsNamed(services, "serviceLogButton")), len(project.Services); got != want {
		t.Errorf("the Services carry %d ways to their logs, want %d", got, want)
	}
	for _, button := range widgetsNamed(services, "serviceLogButton") {
		if button.AccessibleName() == "" {
			t.Error("a way to the logs of a Service has no accessible name")
		}
	}

	// The tab names the selected Project and starts the stream.
	s.bottom.Open()
	s.bottom.tabs.SetCurrentIndex(logTab)
	waitFor(t, "the stream of the Project", func() bool { return s.bottom.logs.Running() })
	if got, want := s.bottom.logs.target.Text(), logs.LogTargetText(project, ""); got != want {
		t.Errorf("the Log tab names %q, want %q", got, want)
	}

	clock := logs.LogTime("2024-03-01T12:34:56Z")
	for _, line := range []compose.LogLine{
		{Service: "web", Timestamp: "2024-03-01T12:34:56Z", Text: "GET /index.html 200"},
		{Service: "db", Text: "database system is ready"},
		{Text: "no configuration file provided", Stderr: true},
	} {
		lines <- line
	}
	waitFor(t, "the first lines", func() bool { return len(s.bottom.logs.lines) == 3 })
	settle()

	text := s.bottom.logs.view.ToPlainText()
	for _, want := range []string{"web", "GET /index.html 200", "db", "database system is ready", clock} {
		if !strings.Contains(text, want) {
			t.Errorf("the text is %q, want %q", text, want)
		}
	}
	// A message of compose itself carries a prefix, so it reads without the
	// colour too.
	if !strings.Contains(text, stderrPrefix+"no configuration file provided") {
		t.Errorf("the text is %q, want the prefix of a message of compose", text)
	}

	// The format of a Service comes from the role of the palette.
	hue := logs.ServiceHue("web")
	if got, want := s.bottom.logs.formats.service[hue].Foreground().Color().Name(), design.ServiceColor(hue, design.LightTheme); got != want {
		t.Errorf("the colour of the name %q is %s, want %s", "web", got, want)
	}

	// The search keeps the lines that match and counts them.
	s.bottom.logs.search.SetText("ready")
	settle()
	text = s.bottom.logs.view.ToPlainText()
	if strings.Contains(text, "GET /index.html 200") || !strings.Contains(text, "database system is ready") {
		t.Errorf("the search shows %q, want the matching line alone", text)
	}
	if got, want := s.bottom.logs.count.Text(), logs.LogCountText(1, 3); got != want {
		t.Errorf("the count reads %q, want %q", got, want)
	}

	// A search with no match keeps the text empty and says so.
	s.bottom.logs.search.SetText("missing")
	settle()
	if got := s.bottom.logs.status.Text(); got != activity.LogNoMatchText {
		t.Errorf("the empty search reads %q, want %q", got, activity.LogNoMatchText)
	}
	s.bottom.logs.search.SetText("")
	settle()

	// The timestamps come and go.
	s.bottom.logs.clocks.SetChecked(false)
	settle()
	if strings.Contains(s.bottom.logs.view.ToPlainText(), clock) {
		t.Error("the clock stays on screen with the timestamps off")
	}
	s.bottom.logs.clocks.SetChecked(true)
	settle()

	// The pause stops the scrolling; resuming goes back to the last line.
	s.bottom.logs.pause.Click()
	settle()
	if !s.bottom.logs.paused {
		t.Fatal("the pause does not stop the scrolling")
	}
	if got := s.bottom.logs.pause.Text(); got != activity.LogResumeText {
		t.Errorf("the pause button reads %q, want %q", got, activity.LogResumeText)
	}
	for i := 0; i < 400; i++ {
		lines <- compose.LogLine{Service: "web", Text: fmt.Sprintf("filler %d", i)}
	}
	waitFor(t, "the filler lines", func() bool { return len(s.bottom.logs.lines) == 403 })
	settle()
	bar := s.bottom.logs.view.VerticalScrollBar()
	if bar == nil || bar.Maximum() == 0 {
		t.Fatal("the text of the logs does not scroll")
	}
	if got, max := bar.Value(), bar.Maximum(); got >= max {
		t.Errorf("the scroll is at %d of %d while paused, want the view left behind", got, max)
	}
	s.bottom.logs.pause.Click()
	settle()
	waitFor(t, "the last line after the resume", func() bool { return bar.Value() == bar.Maximum() })

	// The panel keeps at most logs.LogLineLimit lines and drops the oldest.
	for i := 0; i < logs.LogLineLimit; i++ {
		lines <- compose.LogLine{Service: "db", Text: fmt.Sprintf("overflow %d", i)}
	}
	waitFor(t, "the cap of the lines", func() bool { return len(s.bottom.logs.lines) == logs.LogLineLimit })
	if got := len(s.bottom.logs.lines); got != logs.LogLineLimit {
		t.Errorf("the panel keeps %d lines, want %d", got, logs.LogLineLimit)
	}
	if text := s.bottom.logs.view.ToPlainText(); strings.Contains(text, "GET /index.html 200") ||
		!strings.Contains(text, "overflow 0\n") || !strings.Contains(text, "overflow 4999") {
		t.Error("the text does not hold the last lines of the burst alone")
	}
	if got, limit := s.bottom.logs.view.Document().BlockCount(), logs.LogLineLimit+1; got > limit {
		t.Errorf("the text holds %d blocks, want at most %d", got, limit)
	}

	// A Service of the detail opens the logs of that Service alone.
	s.bottom.ShowLogs(project, "db")
	waitFor(t, "the stream of the Service", func() bool { return s.bottom.logs.Running() })
	if got, want := s.bottom.logs.target.Text(), logs.LogTargetText(project, "db"); got != want {
		t.Errorf("the Log tab names %q, want %q", got, want)
	}
	if name, service := s.bottom.logs.Target(); name != project.Name || service != "db" {
		t.Errorf("the page streams %q/%q, want %q/%q", name, service, project.Name, "db")
	}

	// The way of a Service in the detail reaches the same page.
	s.panel.onLogs(project, "cache")
	waitFor(t, "the stream of the Service button", func() bool { return s.bottom.logs.Running() })
	if name, service := s.bottom.logs.Target(); name != project.Name || service != "cache" {
		t.Errorf("the Service button streams %q/%q, want %q/%q", name, service, project.Name, "cache")
	}

	// Closing the panel terminates the process of the logs.
	s.bottom.Close()
	guitest.WaitSignal(t, stopped, "the end of the stream on close")
	if s.bottom.logs.Running() {
		t.Error("the stream runs after the panel closed")
	}

	// A change of Project terminates the stream of the old one, and the Log
	// tab on screen starts again on the new one.
	s.bottom.ShowLogs(project, "")
	waitFor(t, "the stream again", func() bool { return s.bottom.logs.Running() })
	s.open(projects[2]) // blog
	guitest.WaitSignal(t, stopped, "the end of the stream on a change of Project")
	waitFor(t, "the stream of the new Project", func() bool { return s.bottom.logs.Running() })
	if name, service := s.bottom.logs.Target(); name != projects[2].Name || service != "" {
		t.Errorf("the page streams %q/%q, want %q alone", name, service, projects[2].Name)
	}
	s.bottom.Close()
	guitest.WaitSignal(t, stopped, "the last stream")
}

// checkUsage checks the Utilizzo risorse: the bar of the machine above the
// detail, the indicator of a Service that has a measure, the Service without a
// Container, the height of the rows, and the figures of a stream that stopped.
func checkUsage(t *testing.T, s *shell) {
	t.Helper()

	project := compose.Project{
		Name: "usage", Group: "Casa", Status: compose.StatusActive,
		Folder: "/srv/usage", ComposeFiles: []string{"/srv/usage/compose.yml"},
		Running: 1, Total: 2,
		Services: []compose.Service{
			{Name: "web", Containers: []compose.Container{
				{Name: "usage-web-1", State: "running"},
				{Name: "usage-web-2", State: "exited"},
			}},
			{Name: "worker", Containers: []compose.Container{{Name: "usage-worker-1", State: "exited"}}},
			{Name: "idle"},
		},
	}
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{project}})
	s.open(project)
	settle()

	// Before the first sample the bar waits and claims no figure, and no
	// Service holds room for an indicator.
	if !s.panel.strip.widget.IsVisible() {
		t.Fatal("the bar of the machine is not on screen with a Project open")
	}
	if s.panel.strip.body.CurrentWidget().UnsafePointer() != s.panel.strip.message.QWidget.UnsafePointer() {
		t.Error("the bar shows figures before the first sample")
	}
	if got := s.panel.strip.message.Text(); got != usage.WaitingText {
		t.Errorf("the bar says %q before the first sample, want %q", got, usage.WaitingText)
	}
	for key, row := range s.panel.usageRows {
		if row.widget.IsVisible() {
			t.Errorf("the Service %q shows an indicator before the first sample", key)
		}
	}

	sample := compose.UsageSample{
		DockerReachable: true, Containers: 2,
		CPUs: 4, MemoryTotalBytes: 8 << 30,
		CPUMachinePercent: 12.3, MemoryMachinePercent: 21.3,
		Total: compose.Measures{CPUPercent: 49.2, MemoryBytes: 1 << 30},
		Rate: compose.Rates{
			NetInBytesPerSec: 1234.5, NetOutBytesPerSec: 340,
			BlockReadBytesPerSec: 1234.5, BlockWriteBytesPerSec: 340,
		},
		Services: []compose.ServiceUsage{{
			Project:  "usage",
			Service:  "web",
			Measures: compose.Measures{CPUPercent: 49.2, MemoryBytes: 512 << 20},
			Rate:     compose.Rates{NetInBytesPerSec: 1234.5, NetOutBytesPerSec: 340},
		}},
	}
	s.panel.setUsage(&sample)
	settle()

	// The first sample brings the figures of the machine and the indicator of
	// the Service it carries, and nothing for the others.
	if s.panel.strip.body.CurrentWidget().UnsafePointer() != s.panel.strip.row.widget.UnsafePointer() {
		t.Fatal("the bar does not show the figures of the first sample")
	}
	if got := s.panel.strip.row.cpu.value.Text(); got != "12,3%" {
		t.Errorf("the CPU of the machine reads %q, want %q", got, "12,3%")
	}
	if got := s.panel.strip.row.memory.note.Text(); got != "di 8,00 GiB" {
		t.Errorf("the memory of the machine sits in %q, want %q", got, "di 8,00 GiB")
	}

	web := s.panel.usageRows["usage/web"]
	if web == nil {
		t.Fatal("the Service web has no indicator")
	}
	if !web.widget.IsVisible() {
		t.Fatal("the indicator of web is not on screen with its sample")
	}
	if got := web.row.cpu.bar.Value(); got != 492 {
		t.Errorf("the bar of the CPU of web is at %d steps, want 492", got)
	}
	if got := web.row.cpu.value.Text(); got != "49,2%" {
		t.Errorf("the CPU of web reads %q, want %q", got, "49,2%")
	}
	if got := web.row.cpu.note.Text(); got != "0,49 core" {
		t.Errorf("the CPU of web reads %q in cores, want %q", got, "0,49 core")
	}
	for _, key := range []string{"usage/worker", "usage/idle"} {
		if row := s.panel.usageRows[key]; row != nil && row.widget.IsVisible() {
			t.Errorf("the Service %q shows an indicator without a running Container", key)
		}
	}

	// The figures keep one line: neither the width of the window nor the number
	// of Containers moves their height.
	height := web.widget.Height()
	if height <= 0 {
		t.Fatal("the indicator of web has no height on screen")
	}
	s.window.Resize(940, 700)
	s.applyWidth()
	settle()
	if got := web.widget.Height(); got != height {
		t.Errorf("the indicator is %d px tall at 940 px, want %d", got, height)
	}
	s.window.Resize(windowWidth, windowHeight)
	s.applyWidth()
	settle()

	grown := project
	grown.Services = append([]compose.Service{}, project.Services...)
	grown.Services[0].Containers = append(grown.Services[0].Containers, compose.Container{Name: "usage-web-3", State: "running"})
	grown.Running = 2
	grown.Total = 3
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{grown}})
	settle()

	web = s.panel.usageRows["usage/web"]
	if web == nil || !web.widget.IsVisible() {
		t.Fatal("the indicator of web left the screen after a Container was added")
	}
	if got := web.widget.Height(); got != height {
		t.Errorf("the indicator is %d px tall with three Containers, want %d", got, height)
	}

	// A stream that stops takes every figure away: nothing of the old sample
	// stays on screen as if it were current.
	s.panel.setUsage(nil)
	settle()
	if s.panel.strip.body.CurrentWidget().UnsafePointer() != s.panel.strip.message.QWidget.UnsafePointer() {
		t.Error("the bar keeps the figures after the stream stopped")
	}
	for key, row := range s.panel.usageRows {
		if row.widget.IsVisible() {
			t.Errorf("the Service %q keeps its indicator after the stream stopped", key)
		}
	}

	// A Docker that does not answer shows the warning in place of the figures.
	unreachable := compose.UsageSample{Error: "Cannot connect to the Docker daemon"}
	s.panel.setUsage(&unreachable)
	settle()
	if got := s.panel.strip.message.Text(); got != usage.UnreachableText {
		t.Errorf("the bar says %q with Docker unreachable, want %q", got, usage.UnreachableText)
	}
	if got := s.panel.strip.message.ObjectName(); got != design.RoleStatusFault {
		t.Errorf("the warning carries the role %q, want %q", got, design.RoleStatusFault)
	}
	if s.panel.strip.body.CurrentWidget().UnsafePointer() != s.panel.strip.message.QWidget.UnsafePointer() {
		t.Error("the bar shows figures while Docker does not answer")
	}

	// A Docker that does not answer with no sample at all shows the same
	// warning: the snapshot knows it before the stream does.
	s.panel.setUsage(nil)
	s.apply(compose.Snapshot{DockerReachable: false, Projects: []compose.Project{grown}})
	settle()
	if got := s.panel.strip.message.Text(); got != usage.UnreachableText {
		t.Errorf("the bar says %q from an unreachable snapshot, want %q", got, usage.UnreachableText)
	}

	// The bar belongs to Docker, not to the Project on screen: with no Project
	// open it stays, at the figures or the waiting line the machine state asks
	// for.
	s.apply(compose.Snapshot{DockerReachable: true})
	settle()
	if !s.panel.strip.widget.IsVisible() {
		t.Error("the bar of the machine leaves the screen with no Project open")
	}
	if got := s.panel.strip.message.Text(); got != usage.WaitingText {
		t.Errorf("the bar says %q with no Project open, want %q", got, usage.WaitingText)
	}
}

// waitFor processes the events of the Qt thread until cond holds or the
// deadline passes. A stream runs on its own goroutine, so the test cannot wait
// for it with a plain settle.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		for i := 0; i < 4; i++ {
			qt6.QCoreApplication_ProcessEvents()
		}
		if cond() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// hasWidgetNamed reports whether a widget or one of its children carries the
// object name of the style sheet. The blocks nest layouts, so the walk goes
// through them.
func hasWidgetNamed(widget *qt6.QWidget, name string) bool {
	return len(widgetsNamed(widget, name)) > 0
}

// projectActionButtons returns the commands of the row of the Actions of a
// Project: the ordinary ones and the primary one the state suggests.
func projectActionButtons(block *qt6.QWidget) []*qt6.QWidget {
	buttons := widgetsNamed(block, "actionButton")
	return append(buttons, widgetsNamed(block, "primaryAction")...)
}

// widgetsNamed collects the widgets that carry an object name of the style
// sheet, in the order the layouts hold them.
func widgetsNamed(widget *qt6.QWidget, name string) []*qt6.QWidget {
	var found []*qt6.QWidget
	walkWidget(widget, func(w *qt6.QWidget) {
		if w.ObjectName() == name {
			found = append(found, w)
		}
	})
	return found
}

// walkWidget walks a widget and its children through their layouts.
func walkWidget(widget *qt6.QWidget, visit func(*qt6.QWidget)) {
	if widget == nil {
		return
	}
	visit(widget)
	if layout := widget.Layout(); layout != nil {
		walkLayout(layout, visit)
	}
}

// walkLayout walks the widgets and the nested layouts of a layout.
func walkLayout(layout *qt6.QLayout, visit func(*qt6.QWidget)) {
	for i := 0; layout.ItemAt(i) != nil; i++ {
		item := layout.ItemAt(i)
		if child := item.Widget(); child != nil {
			walkWidget(child, visit)
		}
		if nested := item.Layout(); nested != nil {
			walkLayout(nested, visit)
		}
	}
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
// zero, so a widget that loses its layout fails too. A hidden control is
// checked as well, because the window can show it later: the way back of the
// narrow view waits hidden at wide width. The chrome Qt hides inside its own
// containers is not a control of this window, so it stays out.
// checkTray builds the adapter of the tray without showing the icon, so the
// menu, the tooltip, the icon and the two commands run with no desktop. It also
// checks the guard of the availability and the rule that rebuilds the menu only
// when the tree changes.
func checkTray(t *testing.T) {
	t.Helper()

	// The offscreen backend has no tray: the application must start without one.
	if got := startTray(compose.Snapshot{Loading: true}, nil, nil, nil); got != nil {
		got.icon.Hide()
		t.Error("startTray() found a tray with no desktop")
	}

	var shown, quit int
	var chosen traymenu.TrayNode
	adapter := newTrayAdapter(
		func() { shown++ },
		func() { quit++ },
		func(node traymenu.TrayNode) { chosen = node },
	)
	defer adapter.icon.Delete()
	defer adapter.menu.Delete()

	snapshot := compose.Snapshot{
		DockerReachable: true,
		Projects: []compose.Project{{
			Name:     "shop",
			Status:   compose.StatusActive,
			Services: []compose.Service{{Name: "web"}},
		}},
	}
	adapter.apply(snapshot)

	if got := adapter.icon.ToolTip(); got != "Docker Controller · Docker risponde" {
		t.Errorf("tooltip = %q, want the state of Docker in words", got)
	}
	if got := adapter.icon.Icon(); got == nil || got.IsNull() {
		t.Error("the tray has no icon")
	}

	// The Project, a separator, "Mostra finestra" and "Esci".
	actions := adapter.menu.Actions()
	if len(actions) != 4 {
		t.Fatalf("menu = %d entries, want 4", len(actions))
	}
	if got := actions[0].Text(); got != "● shop" {
		t.Errorf("first entry = %q, want the Project", got)
	}
	if !actions[1].IsSeparator() {
		t.Error("second entry is not a separator")
	}
	if got := actions[2].Text(); got != "Mostra finestra" {
		t.Errorf("third entry = %q, want Mostra finestra", got)
	}
	if got := actions[3].Text(); got != "Esci" {
		t.Errorf("last entry = %q, want Esci", got)
	}

	actions[2].Trigger()
	if shown != 1 {
		t.Errorf("Mostra finestra ran %d times, want 1", shown)
	}
	actions[3].Trigger()
	if quit != 1 {
		t.Errorf("Esci ran %d times, want 1", quit)
	}

	// A leaf of the tree reaches its handler the same way the menu does.
	adapter.choose(traymenu.TrayNode{
		Label:   "Up",
		Enabled: true,
		Action:  &compose.Action{Project: "shop", Kind: compose.ActionUp},
	})
	if chosen.Action == nil || *chosen.Action != (compose.Action{Project: "shop", Kind: compose.ActionUp}) {
		t.Errorf("chosen entry = %+v, want Up on shop", chosen)
	}

	// The same snapshot keeps the menu: a rebuild would replace its entries.
	first := actions[0].UnsafePointer()
	adapter.apply(snapshot)
	if got := adapter.menu.Actions()[0].UnsafePointer(); got != first {
		t.Error("an unchanged snapshot rebuilt the menu")
	}

	// A Docker that stops answering changes the icon, the tooltip and the menu.
	adapter.apply(compose.Snapshot{DockerReachable: false})
	if adapter.iconKind != tray.IconOffline {
		t.Errorf("icon = %d, want the offline variant", adapter.iconKind)
	}
	if got := adapter.icon.ToolTip(); got != "Docker Controller · Docker non risponde" {
		t.Errorf("tooltip = %q, want the offline state", got)
	}
	if got := adapter.menu.Actions()[0].UnsafePointer(); got == first {
		t.Error("a changed snapshot kept the old menu")
	}
	if len(adapter.submenus) != 0 {
		t.Errorf("an offline tree left %d submenus behind, want none", len(adapter.submenus))
	}
}

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
		if parent := widget.Parent(); parent != nil && !widget.IsVisible() &&
			strings.HasPrefix(parent.ObjectName(), "qt_") {
			continue
		}
		widget.EnsurePolished()
		hint, minimum := widget.SizeHint(), widget.MinimumSizeHint()
		width := max(hint.Width(), minimum.Width())
		height := max(hint.Height(), minimum.Height())
		if width < design.TargetMin || height < design.TargetMin {
			t.Errorf("control %q: target %dx%d, want at least %dx%d",
				widget.ObjectName(), width, height, design.TargetMin, design.TargetMin)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no interactive control found: the wiring is broken")
	}
	t.Logf("checked %d interactive controls against the %d px target", checked, design.TargetMin)
}

// captureScreens writes the PNGs of the main states of the window, so a change
// of layout can be reviewed without a display. The test skips without SHOT_DIR:
// the files are a review aid, not an assertion.
func captureScreens(t *testing.T, dir string, app *qt6.QApplication) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create %q: %v", dir, err)
	}

	s := newShell(&application{}, design.LightTheme)
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
	shot("03a-waiting")
	// The figures of the machine and of the Services of the open Project, as
	// the first sample brings them.
	s.panel.setUsage(screenshotUsage(projects[1]))
	shot("03-detail")

	// The form of the settings, with the Italian reason of a refused save under
	// its first field: the light theme and the dark one have a shot each.
	s.panel.startEdit()
	s.panel.edit.message = view.UserError(compose.ErrUnknownProject).Error()
	s.panel.render()
	shot("18-form-light")
	s.panel.cancelEdit()
	settle()

	// A Rimosso Project with its files still on disk: the way off the list sits
	// in the row of the commands, because no notice carries it.
	s.open(projects[3])
	settle()
	shot("03b-detail-removed")

	s.open(projects[1])
	settle()
	s.showNotice(guiproject.ActivityRunningText)
	shot("04-notice")
	s.notice.hide()

	missing := projects[1]
	missing.ComposeMissing = true
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{missing}})
	s.open(missing)
	shot("05-detail-missing")

	// A Rimosso Project may leave the list: the notice of its missing File
	// compose carries the removal command, and it can be pressed.
	removable := compose.Project{
		Name: "docs", DisplayName: "Documentazione", Group: "Lavoro",
		Folder: "/srv/docs", ComposeFiles: []string{"/srv/docs/docker-compose.yml"},
		ComposeMissing: true, Status: compose.StatusRemoved,
	}
	s.apply(compose.Snapshot{DockerReachable: true, Projects: []compose.Project{removable}})
	s.open(removable)
	shot("05b-detail-missing-remove")

	s.apply(compose.Snapshot{DockerReachable: false, Projects: []compose.Project{missing}})
	s.open(missing)
	shot("06-detail-unreachable")

	s.apply(compose.Snapshot{DockerReachable: true, Projects: projects})
	s.window.Resize(800, 620)
	s.applyWidth()
	s.showList()
	shot("07-narrow-list")

	s.open(projects[1])
	settle()
	s.applyWidth()
	shot("08-narrow-detail")

	// The modal question is grabbed on its own: the window behind it would hide
	// the size of the dialog, which is what the check is about.
	confirm := guiproject.Confirmation{
		Title:        "Down di " + view.ProjectName(projects[0]),
		Message:      "Ferma e rimuove i Container del Progetto. I volumi restano. Il Progetto resta in lista come Rimosso.",
		ConfirmLabel: "Down",
	}
	captureDialog(t, s.window.QWidget, dir, "09-confirm", confirm)

	typed := confirm
	typed.Title = "Down con volumi di " + view.ProjectName(projects[0])
	typed.Message = "Rimuove i Container e ELIMINA i volumi del Progetto: i dati nei volumi andranno persi per sempre."
	typed.ConfirmLabel = "Down con volumi"
	typed.TypeToConfirm = projects[0].Name
	captureDialog(t, s.window.QWidget, dir, "10-confirm-type", typed)

	// The question of "Rimuovi dalla lista" and the form of a new Project, with
	// two File compose in the order the picker returned them.
	captureDialog(t, s.window.QWidget, dir, "21-confirm-remove", guiproject.RemoveConfirmation(projects[3]))
	captureAddForm(t, s.window.QWidget, dir, "20-add-form")

	// The settings and the question of a close, with a tray: the three choices
	// are on, and "Nascondi" is one of the answers.
	capturePreferences(t, s.window.QWidget, dir, "22-preferences-light", settings.Defaults(), true, false)
	capturePreferences(t, s.window.QWidget, dir, "22b-preferences-error", settings.Defaults(), true, true)
	captureCloseQuestion(t, s.window.QWidget, dir, "23-close-light", true)

	// The panel of the Activities: one that runs, one that worked, and one that
	// failed with the notice that leads to its output.
	captureActivities(t, dir, design.LightTheme, "13-activities-light")
	// The tab of the logs, with more Services and a message of compose itself.
	captureLogs(t, dir, design.LightTheme, "16-log-light")

	// The dark desktop: the same blocks with the palette of Breeze Dark, so a
	// role that reads on white and disappears on black shows here.
	qt6.QApplication_SetPalette(testDarkPalette())
	if got := selectionForPalette(design.LightTheme, qt6.QGuiApplication_Palette()); got != "" {
		t.Errorf("light selection = %q, want the highlight of the sheet", got)
	}
	if got := selectionForPalette(design.DarkTheme, qt6.QGuiApplication_Palette()); got != "#1b4155" {
		t.Errorf("dark selection = %q, want the inactive highlight of the palette", got)
	}
	app.SetStyleSheet(design.ApplicationStyleSheet(design.DarkTheme, selectionForPalette(design.DarkTheme, qt6.QGuiApplication_Palette())))
	dark := newShell(&application{}, design.DarkTheme)
	defer dark.window.Delete()
	dark.show()
	settle()
	dark.apply(compose.Snapshot{DockerReachable: true, Projects: projects})
	dark.open(projects[1])
	// The click that opens a Project also selects its row: the screenshot shows
	// the band of the selection of the dark palette.
	dark.list.selectProject(projects[1].Name)
	settle()
	dark.panel.setUsage(screenshotUsage(projects[1]))
	saveGrab(t, dark.window, dir, "11-detail-dark")
	// The form of the settings in the dark palette: the message of a refused
	// save must read on the dark surface too.
	dark.panel.startEdit()
	dark.panel.edit.message = view.UserError(compose.ErrUnknownProject).Error()
	dark.panel.render()
	saveGrab(t, dark.window, dir, "19-form-dark")
	dark.panel.cancelEdit()
	captureDialog(t, dark.window.QWidget, dir, "12-confirm-dark", confirm)
	// The settings and the question of a close on the dark surface: the dialog
	// shows the desktop without a tray, so the disabled "Nascondi" and its
	// reason must read here too.
	darkPrefs := settings.Settings{CloseBehavior: settings.CloseAsk, TerminalCommand: "alacritty --working-directory {dir}"}
	capturePreferences(t, dark.window.QWidget, dir, "24-preferences-dark", darkPrefs, false, false)
	captureCloseQuestion(t, dark.window.QWidget, dir, "25-close-dark", false)
	captureActivities(t, dir, design.DarkTheme, "15-activities-dark")
	captureLogs(t, dir, design.DarkTheme, "17-log-dark")
}

// captureActivities writes the PNGs of the panel of the Activities in one
// theme: an Activity that runs with its output growing, one that worked, and
// one that failed with the notice over the window.
func captureActivities(t *testing.T, dir string, theme design.DesignTheme, name string) {
	t.Helper()

	log := activity.NewLog()
	s := newShell(&application{activities: log}, theme)
	defer s.window.Delete()
	s.show()
	settle()

	projects := screenshotProjects()
	s.apply(compose.Snapshot{DockerReachable: true, Projects: projects})
	s.open(projects[1])
	settle()

	log.Push(guitest.Started("cap-run", "shop", "", compose.ActionUp))
	for _, line := range []string{
		"[+] Running 2/2",
		" ✔ Container shop-db-1  Started",
		" ✔ Container shop-web-1  Started",
	} {
		log.Push(guitest.Output("cap-run", "shop", "", compose.ActionUp, line, false))
	}
	log.Push(guitest.Started("cap-ok", "vakt", "web", compose.ActionRestart))
	log.Push(guitest.Output("cap-ok", "vakt", "web", compose.ActionRestart, " ✔ Container vakt-web-1  Restarted", false))
	log.Push(guitest.Finished("cap-ok", "vakt", "web", compose.ActionRestart, "", ""))
	// The running Activity is the one on screen: it is what the user follows
	// while the command runs. The pushes queue their refresh, so the page must
	// have its rows before the choice.
	settle()
	s.bottom.Show("cap-run")
	saveGrab(t, s.window, dir, name)

	log.Push(guitest.Started("cap-bad", "blog", "", compose.ActionPull))
	log.Push(guitest.Output("cap-bad", "blog", "", compose.ActionPull, "Pulling web...", false))
	log.Push(guitest.Output("cap-bad", "blog", "", compose.ActionPull, "error: no such image", true))
	log.Push(guitest.Finished("cap-bad", "blog", "", compose.ActionPull, "error: no such image\n", "error: no such image"))
	saveGrab(t, s.window, dir, name+"-failure")
}

// captureLogs writes the PNG of the Log tab in one theme: several Services,
// the clock of each line and a message of compose itself.
func captureLogs(t *testing.T, dir string, theme design.DesignTheme, name string) {
	t.Helper()

	log := activity.NewLog()
	s := newShell(&application{activities: log}, theme)
	defer s.window.Delete()
	s.show()
	settle()

	projects := screenshotProjects()
	project := projects[0]
	s.apply(compose.Snapshot{DockerReachable: true, Projects: projects})
	s.open(project)
	settle()

	const base = "2024-03-01T12:34:5"
	lines := []compose.LogLine{
		{Service: "web", Timestamp: base + "1Z", Text: "GET /index.html 200 3ms"},
		{Service: "db", Timestamp: base + "2Z", Text: "database system is ready to accept connections"},
		{Service: "web", Timestamp: base + "3Z", Text: "GET /api/health 200 1ms"},
		{Service: "cache", Timestamp: base + "4Z", Text: "Ready to accept connections tcp"},
		{Service: "db", Timestamp: base + "5Z", Text: "checkpoint starting: time"},
		{Service: "web", Timestamp: base + "6Z", Text: "GET /api/items 500 12ms"},
		{Service: "cache", Timestamp: base + "7Z", Text: "1 clients connected"},
		{Timestamp: base + "8Z", Text: "warning: no dependency for service cache", Stderr: true},
		{Service: "web", Timestamp: base + "9Z", Text: "GET /static/app.css 200 2ms"},
	}

	done := make(chan struct{})
	s.bottom.logs.ctx = context.Background()
	s.bottom.logs.source = func(ctx context.Context, _ compose.LogRequest, onLine func(compose.LogLine)) error {
		for _, line := range lines {
			onLine(line)
		}
		<-ctx.Done()
		close(done)
		return nil
	}
	s.bottom.ShowLogs(project, "")
	waitFor(t, "the lines of the capture", func() bool { return len(s.bottom.logs.lines) == len(lines) })
	settle()
	saveGrab(t, s.window, dir, name)

	s.bottom.Close()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the stream of the capture did not end")
	}
}

// testDarkPalette is the palette of Breeze Dark: the offscreen backend has no
// desktop theme, so a test that reviews the dark mode brings its own.
func testDarkPalette() *qt6.QPalette {
	palette := qt6.NewQPalette()
	set := func(role qt6.QPalette__ColorRole, hex string) {
		palette.SetColor2(role, qt6.NewQColor6(hex))
	}
	set(qt6.QPalette__Window, "#232629")
	set(qt6.QPalette__WindowText, "#eff0f1")
	set(qt6.QPalette__Base, "#1b1e20")
	set(qt6.QPalette__AlternateBase, "#232629")
	set(qt6.QPalette__Text, "#eff0f1")
	set(qt6.QPalette__Button, "#31363b")
	set(qt6.QPalette__ButtonText, "#eff0f1")
	set(qt6.QPalette__BrightText, "#ffffff")
	set(qt6.QPalette__Highlight, "#3daee9")
	set(qt6.QPalette__HighlightedText, "#ffffff")
	// A view without focus takes the darker highlight of the desktop.
	palette.SetColor(qt6.QPalette__Inactive, qt6.QPalette__Highlight, qt6.NewQColor6("#1b4155"))
	set(qt6.QPalette__Mid, "#4a4f55")
	set(qt6.QPalette__Dark, "#191c1e")
	set(qt6.QPalette__Light, "#5a6068")
	return palette
}

// saveGrab writes the PNG of one window.
func saveGrab(t *testing.T, window *qt6.QMainWindow, dir, name string) {
	t.Helper()

	settle()
	pixmap := window.Grab()
	if pixmap == nil {
		t.Fatalf("grab %q: no pixmap", name)
	}
	path := filepath.Join(dir, name+".png")
	if !pixmap.Save(path) {
		t.Fatalf("save %q", path)
	}
	t.Logf("wrote %s", path)
}

// captureDialog writes the PNG of a modal question, so its size and its text
// can be reviewed without a desktop.
func captureDialog(t *testing.T, parent *qt6.QWidget, dir, name string, question guiproject.Confirmation) {
	t.Helper()

	dialog := newConfirmDialog(parent, question, func() {})
	dialog.dialog.Show()
	settle()
	pixmap := dialog.dialog.Grab()
	if pixmap == nil {
		t.Fatalf("grab %q: no pixmap", name)
	}
	path := filepath.Join(dir, name+".png")
	if !pixmap.Save(path) {
		t.Fatalf("save %q", path)
	}
	t.Logf("wrote %s", path)
	dialog.dialog.Delete()
}

// checkPreferences drives the settings dialog: the three choices, the command
// that cannot run and its message under the field, the save that closes it and
// the choice the desktop cannot honour. It is a step of the smoke test, so the
// process still has one QApplication.
func checkPreferences(t *testing.T, s *shell) {
	t.Helper()

	var saved settings.Settings
	saves := 0
	dialog := newPreferencesDialog(s.window.QWidget, settings.Defaults(), true,
		func(values settings.Settings, done func(error)) {
			saves++
			saved = values
			done(nil)
		})
	dialog.open()
	settle()

	// The three choices are on screen and on, and the saved one is the checked
	// one.
	if len(dialog.choices) != len(forms.CloseBehaviorChoices) {
		t.Fatalf("%d choices, want %d", len(dialog.choices), len(forms.CloseBehaviorChoices))
	}
	for i, radio := range dialog.choices {
		if radio.Text() != forms.CloseBehaviorChoices[i].Label {
			t.Errorf("choice %d = %q, want %q", i, radio.Text(), forms.CloseBehaviorChoices[i].Label)
		}
		if !radio.IsEnabled() {
			t.Errorf("choice %d is off with a tray", i)
		}
	}
	if !dialog.choices[forms.CloseBehaviorIndex(settings.CloseHide)].IsChecked() {
		t.Error("the dialog does not show the saved behavior")
	}

	// A command that cannot run shows its message under the field, keeps the
	// dialog open and saves nothing.
	dialog.terminal.SetText(`konsole --workdir "shop`)
	dialog.save.Click()
	settle()
	if saves != 0 {
		t.Fatal("an invalid command reached the save")
	}
	if !dialog.error.IsVisible() {
		t.Fatal("the message of an invalid command is not on screen")
	}
	if got := dialog.error.Text(); !strings.Contains(got, "Comando del terminale non valido") {
		t.Errorf("message = %q, want the Italian refusal of the terminal command", got)
	}
	if !strings.Contains(dialog.terminal.AccessibleName(), "Comando del terminale non valido") {
		t.Errorf("accessible name = %q, want the message with the field", dialog.terminal.AccessibleName())
	}
	if !dialog.dialog.IsVisible() {
		t.Fatal("the dialog closed on an invalid command")
	}

	// The change of the field takes the message away.
	dialog.terminal.SetText(settings.DefaultTerminalCommand)
	if dialog.error.IsVisible() {
		t.Error("the message of the refusal stays after the command changed")
	}

	// A valid command and another behavior save both values and close the
	// dialog.
	dialog.choices[forms.CloseBehaviorIndex(settings.CloseQuit)].Click()
	dialog.save.Click()
	settle()
	if saves != 1 {
		t.Fatalf("saves = %d, want 1", saves)
	}
	want := settings.Settings{CloseBehavior: settings.CloseQuit, TerminalCommand: settings.DefaultTerminalCommand}
	if saved != want {
		t.Errorf("saved = %+v, want %+v", saved, want)
	}
	if dialog.dialog.IsVisible() {
		t.Error("the dialog stays on screen after a save that worked")
	}

	// Without a tray the hide choice is off, with the reason in its tooltip, and
	// the dialog shows the quit the close really does.
	noTray := newPreferencesDialog(s.window.QWidget, settings.Defaults(), false, nil)
	defer noTray.dialog.Delete()
	hide := noTray.choices[forms.CloseBehaviorIndex(settings.CloseHide)]
	if hide.IsEnabled() {
		t.Error("hiding the window is a choice without a tray")
	}
	if got := hide.ToolTip(); got != forms.NoTrayHintText {
		t.Errorf("tooltip = %q, want the reason hiding is not a choice", got)
	}
	if !noTray.choices[forms.CloseBehaviorIndex(settings.CloseQuit)].IsChecked() {
		t.Error("without a tray the dialog does not show Esci")
	}

	// A save that ends after the dialog closed touches no widget: Escape during
	// a write must not break the window.
	late := newPreferencesDialog(s.window.QWidget, settings.Defaults(), true,
		func(settings.Settings, func(error)) {})
	late.open()
	settle()
	late.dialog.Reject()
	settle()
	late.done(nil)
}

// checkCloseDialog drives the "Chiedi ogni volta" question: the answers, the
// option to remember, and the answer that comes back after the dialog closed.
func checkCloseDialog(t *testing.T, s *shell) {
	t.Helper()

	var choice settings.CloseBehavior
	kept, calls := false, 0
	answer := func(c settings.CloseBehavior, k bool) {
		calls++
		choice, kept = c, k
	}
	question := newCloseDialog(s.window.QWidget, true, answer)
	question.open()
	settle()

	if !question.dialog.IsModal() {
		t.Error("the question is not modal")
	}
	if len(question.answers) != len(forms.AskCloseAnswers) {
		t.Fatalf("%d answers, want %d", len(question.answers), len(forms.AskCloseAnswers))
	}

	// "Annulla" is the default and answers nothing, even with the option on.
	question.remember.SetChecked(true)
	question.answers[len(question.answers)-1].Click()
	settle()
	if calls != 0 {
		t.Fatal("Annulla answered the question")
	}

	// The next question remembers the answer and reports it after it closed.
	second := newCloseDialog(s.window.QWidget, true, answer)
	second.open()
	second.remember.SetChecked(true)
	second.answers[0].Click()
	settle()
	if calls != 1 {
		t.Fatalf("answers = %d, want 1", calls)
	}
	if choice != settings.CloseHide || !kept {
		t.Errorf("answer = %q with remember %v, want hide and the option on", choice, kept)
	}

	// Without a tray the question offers no way to a window that could not come
	// back.
	open := newCloseDialog(s.window.QWidget, false, nil)
	defer open.dialog.Delete()
	if len(open.answers) != 2 {
		t.Fatalf("%d answers without a tray, want Esci and Annulla", len(open.answers))
	}
	if got := open.answers[0].Text(); got != "Esci" {
		t.Errorf("first answer = %q, want Esci", got)
	}
}

// checkHiddenWindow hides the window the way the tray does and brings it back:
// the window is not destroyed, the Project on screen and the history stay, and
// the stream of the logs stops and starts again. It is a step of the smoke test.
func checkHiddenWindow(t *testing.T) {
	t.Helper()

	log := activity.NewLog()
	// The application answers the close with the saved behavior. An adapter
	// without an icon is enough: the close asks only whether a tray exists.
	app := &application{
		activities: log,
		prefs:      settings.Settings{CloseBehavior: settings.CloseHide, TerminalCommand: settings.DefaultTerminalCommand},
		tray:       &trayAdapter{},
	}
	s := newShell(app, design.LightTheme)
	app.shell = s
	defer s.window.Delete()
	s.show()
	settle()

	projects := screenshotProjects()
	s.apply(compose.Snapshot{DockerReachable: true, Projects: projects})
	s.open(projects[1])
	settle()

	log.Push(guitest.Started("hide-1", projects[1].Name, "", compose.ActionUp))
	log.Push(guitest.Finished("hide-1", projects[1].Name, "", compose.ActionUp, "", ""))
	waitFor(t, "the history of the Activities", func() bool { return s.bottom.page.list.Count() == 1 })
	history := s.bottom.page.list.Count()

	// A stream of logs that runs while the window is on screen.
	done := make(chan struct{})
	var once sync.Once
	s.bottom.logs.ctx = context.Background()
	s.bottom.logs.source = func(ctx context.Context, _ compose.LogRequest, onLine func(compose.LogLine)) error {
		onLine(compose.LogLine{Service: "web", Text: "GET /index.html 200"})
		<-ctx.Done()
		once.Do(func() { close(done) })
		return nil
	}
	s.bottom.ShowLogs(projects[1], "")
	waitFor(t, "the stream of the logs", func() bool { return s.bottom.logs.Running() })

	// The close of the window follows the saved behavior: the window goes off
	// the screen, it is not destroyed, and the stream of the logs stops.
	s.window.Close()
	settle()
	if s.window.IsVisible() {
		t.Fatal("the window is still on screen after the close")
	}
	if s.window.CentralWidget() == nil {
		t.Fatal("the close destroyed the window")
	}
	if s.selected != projects[1].Name {
		t.Errorf("selection = %q after hiding, want %q", s.selected, projects[1].Name)
	}
	if s.bottom.logs.Running() {
		t.Error("the stream of the logs runs with the window hidden")
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the stream of the logs did not end with the window hidden")
	}

	// An Activity that runs while the window is hidden keeps its output: the
	// Manager pushes it into the log of the session, which no window holds.
	log.Push(guitest.Started("hide-2", projects[1].Name, "", compose.ActionPull))
	log.Push(guitest.Output("hide-2", projects[1].Name, "", compose.ActionPull, "✔ Container shop-web-1  Started", false))
	wantHistory := history + 1
	waitFor(t, "the output of the Activity on the hidden window", func() bool {
		return s.bottom.page.list.Count() == wantHistory &&
			strings.Contains(s.bottom.page.output.ToPlainText(), "shop-web-1")
	})

	// The window comes back: the same Project and the same history, and the
	// stream of the logs starts again.
	s.present()
	settle()
	if !s.window.IsVisible() {
		t.Fatal("the window is not on screen after present")
	}
	if s.selected != projects[1].Name {
		t.Errorf("selection = %q after the return, want %q", s.selected, projects[1].Name)
	}
	if got := s.panel.shown.Project.Name; got != projects[1].Name {
		t.Errorf("the detail shows %q after the return, want %q", got, projects[1].Name)
	}
	if got := s.bottom.page.list.Count(); got != wantHistory {
		t.Errorf("history = %d rows after the return, want %d", got, wantHistory)
	}
	// The new Activity opened the panel on its tab, so the logs start again when
	// the user goes back to their tab: a stopped stream resumes nothing by
	// itself.
	s.bottom.tabs.SetCurrentIndex(logTab)
	waitFor(t, "the stream of the logs on the return", func() bool { return s.bottom.logs.Running() })
	s.bottom.Close()
}

// checkQuit ends the application the way the tray does, and the close of the
// window on a desktop without one: the streams stop before the event loop does,
// so no `docker compose logs` and no `docker stats` outlives the window. It is
// the last step of the smoke test, because Quit leaves the application of the
// process closing.
func checkQuit(t *testing.T) {
	t.Helper()

	log := activity.NewLog()
	app := &application{activities: log, prefs: settings.Defaults()}
	s := newShell(app, design.LightTheme)
	app.shell = s
	defer s.window.Delete()
	s.show()
	settle()

	projects := screenshotProjects()
	s.apply(compose.Snapshot{DockerReachable: true, Projects: projects})
	s.open(projects[1])
	settle()

	stopped := make(chan struct{})
	s.bottom.logs.ctx = context.Background()
	s.bottom.logs.source = func(ctx context.Context, _ compose.LogRequest, _ func(compose.LogLine)) error {
		<-ctx.Done()
		close(stopped)
		return nil
	}
	s.bottom.ShowLogs(projects[1], "")
	waitFor(t, "the stream of the logs", func() bool { return s.bottom.logs.Running() })

	app.quit()
	if !app.quitting {
		t.Error("quit did not mark the application as leaving")
	}
	if s.bottom.logs.Running() {
		t.Error("the stream of the logs runs after quit")
	}
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("the stream of the logs did not end with the application")
	}

	// A second quit, and the close of a window that would follow, start nothing.
	app.quit()
	app.closeRequested()
}

// capturePreferences writes the PNG of the settings dialog, so its layout and
// its message under the command can be reviewed without a desktop. withError
// puts the refusal of an invalid command under the field.
func capturePreferences(t *testing.T, parent *qt6.QWidget, dir, name string, prefs settings.Settings, tray, withError bool) {
	t.Helper()

	dialog := newPreferencesDialog(parent, prefs, tray, nil)
	if withError {
		invalid := `konsole --workdir "shop`
		dialog.terminal.SetText(invalid)
		dialog.showError(forms.ValidateTerminalCommand(invalid))
	}
	dialog.dialog.Show()
	settle()
	pixmap := dialog.dialog.Grab()
	if pixmap == nil {
		t.Fatalf("grab %q: no pixmap", name)
	}
	path := filepath.Join(dir, name+".png")
	if !pixmap.Save(path) {
		t.Fatalf("save %q", path)
	}
	t.Logf("wrote %s", path)
	dialog.dialog.Delete()
}

// captureCloseQuestion writes the PNG of the "Chiedi ogni volta" question of a
// close.
func captureCloseQuestion(t *testing.T, parent *qt6.QWidget, dir, name string, tray bool) {
	t.Helper()

	dialog := newCloseDialog(parent, tray, nil)
	dialog.dialog.Show()
	settle()
	pixmap := dialog.dialog.Grab()
	if pixmap == nil {
		t.Fatalf("grab %q: no pixmap", name)
	}
	path := filepath.Join(dir, name+".png")
	if !pixmap.Save(path) {
		t.Fatalf("save %q", path)
	}
	t.Logf("wrote %s", path)
	dialog.dialog.Delete()
}

// captureAddForm writes the PNG of the form of a new Project with two File
// compose, so its layout and the order of the files can be reviewed without a
// desktop.
func captureAddForm(t *testing.T, parent *qt6.QWidget, dir, name string) {
	t.Helper()

	form := forms.AddProjectForm{
		Files: []string{"/srv/shop/compose.override.yml", "/srv/shop/docker-compose.yml"},
		Name:  "shop",
	}
	dialog := newAddDialog(parent, form,
		func(string, []string, func(error)) {},
		func(string) {})
	dialog.dialog.Show()
	settle()
	pixmap := dialog.dialog.Grab()
	if pixmap == nil {
		t.Fatalf("grab %q: no pixmap", name)
	}
	path := filepath.Join(dir, name+".png")
	if !pixmap.Save(path) {
		t.Fatalf("save %q", path)
	}
	t.Logf("wrote %s", path)
	dialog.dialog.Delete()
}

// screenshotUsage is a sample of the Utilizzo risorse for a Project: the figures
// of the machine and those of every Service that runs, so the bar and the
// indicators both show numbers in the capture.
func screenshotUsage(project compose.Project) *compose.UsageSample {
	sample := &compose.UsageSample{
		DockerReachable:      true,
		CPUs:                 8,
		MemoryTotalBytes:     16 << 30,
		CPUMachinePercent:    18.7,
		MemoryMachinePercent: 41.2,
		Total:                compose.Measures{CPUPercent: 149.6, MemoryBytes: 6<<30 + 400<<20},
		Rate: compose.Rates{
			NetInBytesPerSec:      1_200_000,
			NetOutBytesPerSec:     340_000,
			BlockReadBytesPerSec:  12_000_000,
			BlockWriteBytesPerSec: 2_400_000,
		},
	}
	for i, service := range project.Services {
		if !usage.ServiceRuns(service) {
			continue
		}
		sample.Containers += len(service.Containers)
		sample.Services = append(sample.Services, compose.ServiceUsage{
			Project:  project.Name,
			Service:  service.Name,
			Measures: compose.Measures{CPUPercent: 62.4 + float64(i)*7.5, MemoryBytes: 1<<30 + uint64(i)*256<<20},
			Rate: compose.Rates{
				NetInBytesPerSec:      620_000 + float64(i)*40_000,
				NetOutBytesPerSec:     180_000 + float64(i)*20_000,
				BlockReadBytesPerSec:  9_000_000 + float64(i)*500_000,
				BlockWriteBytesPerSec: 1_200_000 + float64(i)*100_000,
			},
		})
	}
	return sample
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
			Profiles:       []string{"debug", "metrics", "trace"},
			ActiveProfiles: []string{"debug", "metrics"},
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
