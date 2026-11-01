//go:build ui

package gui

import (
	"testing"

	"github.com/mappu/miqt/qt6"
)

// TestShellOffscreen builds the empty window, shows it with the offscreen
// backend and destroys it. A broken constructor or a missing wiring of the
// list, the buttons or the central widget fails here, without a display.
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
	qt6.QCoreApplication_ProcessEvents()

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
