package gui

import (
	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/tray"
)

// trayAdapter puts the menu of internal/tray on the system tray. It is the only
// file of the package that knows QSystemTrayIcon: the description of the menu
// stays pure, and the tree it applies has a test without Qt (traytree.go).
//
// On Wayland the icon is a StatusNotifierItem and the desktop draws the menu
// through com.canonical.dbusmenu, so the context menu is a QMenu exported to the
// session bus and never a popup of our own.
type trayAdapter struct {
	// icon is the item on the panel. Nothing else owns a parentless object, so
	// the wrapper stays in the field for the life of the application.
	icon *qt6.QSystemTrayIcon
	// menu is the context menu. A rebuild clears it and refills it, so the icon
	// keeps the same object and SetContextMenu runs once.
	menu *qt6.QMenu
	// tree is the menu on the tray, the last one applied.
	tree []trayNode
	// submenus are the menus this adapter created, so a rebuild can delete the
	// ones of the previous tree: Clear removes their actions but not the menus.
	submenus []*qt6.QMenu
	// iconKind is the icon on the tray, and hasIcon says it was set once.
	iconKind tray.Icon
	hasIcon  bool

	onShow   func()
	onQuit   func()
	onChoose func(trayNode)
}

// startTray creates the tray when the desktop has one. Without a tray it
// returns nil and the application runs as before: the window keeps "Esci" and
// "Impostazioni", so the user never depends on an icon that is not there. The
// first snapshot goes on the icon before it shows, so the icon never appears
// without a picture.
func startTray(initial compose.Snapshot, onShow, onQuit func(), onChoose func(trayNode)) *trayAdapter {
	if !qt6.QSystemTrayIcon_IsSystemTrayAvailable() {
		return nil
	}
	t := newTrayAdapter(onShow, onQuit, onChoose)
	t.apply(initial)
	t.icon.Show()
	return t
}

// newTrayAdapter builds the icon and the menu without showing them, so a test
// drives the adapter with no desktop. Only startTray asks the platform for a
// tray and shows the icon.
func newTrayAdapter(onShow, onQuit func(), onChoose func(trayNode)) *trayAdapter {
	t := &trayAdapter{
		onShow:   onShow,
		onQuit:   onQuit,
		onChoose: onChoose,
		iconKind: tray.IconNormal,
	}
	t.menu = qt6.NewQMenu2()
	t.menu.SetAccessibleName(windowTitle)
	t.icon = qt6.NewQSystemTrayIcon()
	t.icon.SetContextMenu(t.menu)
	t.icon.SetToolTip(trayTooltip(compose.Snapshot{Loading: true}))
	t.icon.OnActivated(func(reason qt6.QSystemTrayIcon__ActivationReason) {
		// A left click, and the double click of a desktop that uses it, show the
		// window. The right click belongs to the context menu the desktop draws,
		// so it runs nothing here.
		switch reason {
		case qt6.QSystemTrayIcon__Trigger, qt6.QSystemTrayIcon__DoubleClick:
			t.activate()
		}
	})
	return t
}

// apply shows a snapshot on the tray: the icon, the tooltip and, only when it
// changed, the menu. An unchanged tree touches nothing, so an entry the user has
// under the pointer is not taken away for nothing.
func (t *trayAdapter) apply(snapshot compose.Snapshot) {
	kind := tray.IconFor(snapshot)
	if !t.hasIcon || kind != t.iconKind {
		t.icon.SetIcon(systemTrayIcon(kind))
		t.iconKind, t.hasIcon = kind, true
	}
	t.icon.SetToolTip(trayTooltip(snapshot))

	nodes := trayTree(tray.BuildMenu(snapshot))
	if sameTrayTree(nodes, t.tree) {
		return
	}
	t.tree = nodes
	t.rebuild()
}

// rebuild replaces the whole menu with the tree. Clear removes the entries of
// the menu before it, and the submenus of the previous tree are deleted: Clear
// deletes an action but not the menu it opens, so without this loop one menu
// would stay behind on every rebuild. They are deleted deepest first, because a
// submenu is a child of the menu that holds it and the wrapper of a child would
// dangle after the parent is gone.
func (t *trayAdapter) rebuild() {
	t.menu.Clear()
	for i := len(t.submenus) - 1; i >= 0; i-- {
		t.submenus[i].Delete()
	}
	t.submenus = nil
	for _, node := range t.tree {
		t.add(t.menu, node)
	}
}

// add appends one node to menu, or to the submenu it builds. A leaf with an
// Action or a Command gets a callback; a disabled entry cannot fire, and a
// notice has nothing to trigger.
func (t *trayAdapter) add(menu *qt6.QMenu, node trayNode) {
	if node.Separator {
		menu.AddSeparator()
		return
	}
	if len(node.Children) > 0 {
		sub := qt6.NewQMenu4(node.Label, menu.QWidget)
		t.submenus = append(t.submenus, sub)
		action := menu.AddMenu(sub)
		action.SetEnabled(node.Enabled)
		for _, child := range node.Children {
			t.add(sub, child)
		}
		return
	}
	action := menu.AddActionWithText(node.Label)
	action.SetEnabled(node.Enabled)
	if node.Action == nil && node.Command == tray.CommandNone {
		// A notice: it says why the Actions are off and has nothing to trigger.
		return
	}
	action.OnTriggered(func() { t.choose(node) })
}

// choose runs the entry the user clicked. It runs on the Qt thread: the
// confirmation opens there, and only the Manager call moves to a goroutine.
func (t *trayAdapter) choose(node trayNode) {
	switch node.Command {
	case tray.CommandShowWindow:
		t.activate()
		return
	case tray.CommandQuit:
		t.quit()
		return
	}
	if node.Action == nil || t.onChoose == nil {
		return
	}
	t.onChoose(node)
}

// activate shows the window through the caller.
func (t *trayAdapter) activate() {
	if t.onShow != nil {
		t.onShow()
	}
}

// quit ends the application through the caller.
func (t *trayAdapter) quit() {
	if t.onQuit != nil {
		t.onQuit()
	}
}

// systemTrayIcon builds the icon of the application from the PNG the tray
// package embeds. No theme lookup and no installed desktop entry are involved,
// so the icon is the same in a build that runs from the checkout, and the
// offline variant already exists beside the normal one.
func systemTrayIcon(kind tray.Icon) *qt6.QIcon {
	pixmap := qt6.NewQPixmap()
	if !pixmap.LoadFromDataWithData(kind.PNG()) {
		return qt6.NewQIcon()
	}
	return qt6.NewQIcon2(pixmap)
}

// present shows the window and brings it to the front: the left click of the
// icon and "Mostra finestra" land here. The window manager may still refuse the
// focus, so the window is at least restored and raised.
func (a *application) present() {
	if a.shell == nil {
		return
	}
	a.shell.present()
}

// quit ends the application: "Esci" of the tray and of the window both land
// here, so the two ways out behave the same. QApplication inherits quit from
// QCoreApplication, which is the binding miqt exposes.
func (a *application) quit() {
	qt6.QCoreApplication_Quit()
}

// chooseTray runs an entry chosen in the tray menu. A destructive Action asks
// its question first, and the answer runs it.
func (a *application) chooseTray(node trayNode) {
	if node.Action == nil {
		return
	}
	if node.Confirm != "" {
		a.askTray(node)
		return
	}
	a.launchTray(*node.Action)
}

// askTray opens the question of a destructive Action chosen in the tray. The
// window comes first, so the modal question has a parent the user can see.
func (a *application) askTray(node trayNode) {
	question := trayQuestion(node)
	if question == nil || a.shell == nil || a.question != nil {
		return
	}
	a.present()
	dialog := newConfirmDialog(a.shell.window.QWidget, *question, func() {
		a.launchTray(*node.Action)
	})
	dialog.onClosed = func() { a.question = nil }
	a.question = dialog
	dialog.open()
}

// launchTray hands an Action chosen in the tray to the Manager off the Qt
// thread: Docker runs for seconds, and the tray must keep answering. The
// Notifier announces the launch, so the end of the Activity is notified, and a
// refusal reaches the window as the Italian message of the notice.
func (a *application) launchTray(action compose.Action) {
	if a.manager == nil {
		return
	}
	go func() {
		var err error
		if a.notifier != nil {
			err = a.notifier.launch(action, func(action compose.Action) (compose.ActivityID, error) {
				return a.manager.Do(a.ctx, action)
			})
		} else {
			_, err = a.manager.Do(a.ctx, action)
		}
		if err == nil {
			return
		}
		mainthread.Start(func() {
			if a.shell != nil {
				a.shell.showNotice(UserError(err).Error())
			}
		})
	}()
}
