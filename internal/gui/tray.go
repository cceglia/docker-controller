package gui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/tray"
	"dockercontroller/internal/trayadapter"
)

// show brings the window to the front, restored and focused.
func (s *shell) show() {
	s.window.Unminimize()
	s.window.Present()
}

// trayChoose runs the entry chosen in the tray menu. It runs on the GTK loop,
// so the question of a destructive Action and the Action itself are safe. The
// window the menu needs is created when it is gone: the tray outlives it.
func (a *application) trayChoose(node trayadapter.Node) {
	switch node.Command {
	case tray.CommandShowWindow:
		a.present()
		return
	case tray.CommandQuit:
		a.quit()
		return
	}
	if node.Action == nil {
		return
	}
	if node.Confirm != "" {
		a.confirmTray(node)
		return
	}
	a.doTray(*node.Action)
}

// doTray launches an Action chosen in the tray. The Notifier announces it before
// the Manager sees it, so the end of the Activity is notified, and withdraws the
// announcement when the Manager refuses the Action. The refusal is also shown as
// a toast when a window is on screen; the notification covers the case of the
// tray alone, so no window is built just for the message.
func (a *application) doTray(action compose.Action) {
	if a.manager == nil {
		return
	}
	err := a.notifier.launch(action, func(action compose.Action) (compose.ActivityID, error) {
		return a.manager.Do(a.ctx, action)
	})
	if err != nil && a.window != nil {
		a.window.toast(UserError(err).Error())
	}
}

// confirmTray asks the question of an entry that carries Confirm and launches
// its Action only when the answer is "run". Cancelling or closing the dialog
// launches nothing. The question goes on the window when there is one. With the
// application in the tray it has no parent, because a tray command builds no
// window and so loses no Activity history.
func (a *application) confirmTray(node trayadapter.Node) {
	if a.quitting {
		return
	}
	parent := (*gtk.Window)(nil)
	host := &a.confirmations
	if a.window != nil {
		a.window.show()
		parent = &a.window.window.ApplicationWindow.Window
		host = &a.window.confirmations
	}

	dialog := adw.NewMessageDialog(parent, windowTitle, node.Confirm)
	host.add(dialog)
	dialog.AddResponse("cancel", "Annulla")
	dialog.AddResponse("run", node.Label)
	dialog.SetCloseResponse("cancel")
	dialog.SetDefaultResponse("cancel")
	dialog.SetResponseAppearance("run", adw.ResponseDestructive)
	dialog.ConnectResponse(func(response string) {
		host.drop(dialog)
		if response == "run" {
			a.doTray(*node.Action)
		}
	})
	dialog.Present()
}

// dialogList keeps the questions of the tray that are on screen, so GTK does
// not free a dialog before it answers. A window keeps the ones it hosts; the
// application keeps the ones with no parent.
type dialogList []*adw.MessageDialog

// add keeps a dialog that just went on screen.
func (list *dialogList) add(dialog *adw.MessageDialog) {
	*list = append(*list, dialog)
}

// drop takes a dialog that answered off the list.
func (list *dialogList) drop(dialog *adw.MessageDialog) {
	for i, open := range *list {
		if open == dialog {
			*list = append((*list)[:i], (*list)[i+1:]...)
			return
		}
	}
}
