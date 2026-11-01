package gui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/tray"
	"dockercontroller/internal/trayadapter"
)

// startTray puts the icon in the system tray and shows the menu of the state on
// screen. A tray that cannot start stays silent: systray logs the failure on
// its own goroutine and the window goes on without it.
func (s *shell) startTray() {
	if s.tray != nil {
		return
	}
	s.tray = trayadapter.New(
		// A click arrives on a D-Bus goroutine: the window moves only on the loop.
		func() { glib.IdleAdd(s.showWindow) },
		func(node trayadapter.Node) { glib.IdleAdd(func() { s.trayChoose(node) }) },
	)
	s.tray.Start()
	s.tray.Show(tray.BuildMenu(s.snapshot), tray.IconFor(s.snapshot))
}

// showWindow brings the window to the front, restored and focused.
func (s *shell) showWindow() {
	if s.window == nil {
		return
	}
	s.window.Unminimize()
	s.window.Present()
}

// trayChoose runs the entry chosen in the tray menu. It runs on the GTK loop,
// so the question of a destructive Action and the Action itself are safe.
func (s *shell) trayChoose(node trayadapter.Node) {
	switch node.Command {
	case tray.CommandShowWindow:
		s.showWindow()
		return
	case tray.CommandQuit:
		s.quit()
		return
	}
	if node.Action == nil {
		return
	}
	if node.Confirm != "" {
		s.confirmTray(node)
		return
	}
	s.doTray(*node.Action)
}

// doTray launches an Action chosen in the tray. The Notifier announces it before
// the Manager sees it, so the end of the Activity is notified, and withdraws the
// announcement when the Manager refuses the Action. The refusal is also shown as
// a toast, and the reason is the one the old view used.
func (s *shell) doTray(action compose.Action) {
	if s.manager == nil {
		return
	}
	err := s.notifier.launch(action, func(action compose.Action) (compose.ActivityID, error) {
		return s.manager.Do(s.ctx, action)
	})
	if err != nil {
		s.toast(UserError(err).Error())
	}
}

// confirmTray asks the question of an entry that carries Confirm and launches
// its Action only when the answer is "run". Cancelling or closing the dialog
// launches nothing. The dialog stays in the shell while it is open.
func (s *shell) confirmTray(node trayadapter.Node) {
	dialog := adw.NewMessageDialog(&s.window.ApplicationWindow.Window, windowTitle, node.Confirm)
	s.confirmations = append(s.confirmations, dialog)
	dialog.AddResponse("cancel", "Annulla")
	dialog.AddResponse("run", node.Label)
	dialog.SetCloseResponse("cancel")
	dialog.SetDefaultResponse("cancel")
	dialog.SetResponseAppearance("run", adw.ResponseDestructive)
	dialog.ConnectResponse(func(response string) {
		s.dropConfirmation(dialog)
		if response == "run" {
			s.doTray(*node.Action)
		}
	})
	dialog.Present()
}

// dropConfirmation takes a dialog that answered off the shell.
func (s *shell) dropConfirmation(dialog *adw.MessageDialog) {
	for i, open := range s.confirmations {
		if open != dialog {
			continue
		}
		s.confirmations = append(s.confirmations[:i], s.confirmations[i+1:]...)
		return
	}
}

// quit closes the application from "Esci". The loop then returns, and the main
// function cancels the context, so no Activity, log stream or `docker stats`
// stays behind.
func (s *shell) quit() {
	if s.app == nil {
		return
	}
	s.app.Quit()
}
