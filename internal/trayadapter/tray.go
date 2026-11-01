package trayadapter

import (
	"sync"

	"fyne.io/systray"

	"dockercontroller/internal/tray"
)

// Title is the text of the tooltip of the icon: the name the user knows the
// app by, the same one the window and the notifications show.
const Title = "Docker Controller"

// Tray shows the icon and the menu on the system tray. It is the only place of
// the package that speaks to systray.
//
// A click arrives on the goroutine of D-Bus, never on the GTK loop. Tray only
// reports it to the callbacks of New: the caller moves every choice, and every
// confirmation, onto the loop with glib.IdleAdd.
type Tray struct {
	onActivate func()
	onChoose   func(Node)

	mu      sync.Mutex
	nodes   []Node
	icon    tray.Icon
	hasIcon bool
	started bool
}

// New creates a tray. onActivate runs on a left click of the icon, onChoose on
// a click on an enabled entry that launches something. Either may be nil.
func New(onActivate func(), onChoose func(Node)) *Tray {
	return &Tray{onActivate: onActivate, onChoose: onChoose}
}

// Start puts the icon on the session bus and returns at once: systray connects
// on its own goroutine and logs a failure there, so the window never waits for
// the tray and never stops because of it. A later call does nothing.
func (t *Tray) Start() {
	t.mu.Lock()
	if t.started {
		t.mu.Unlock()
		return
	}
	t.started = true
	t.mu.Unlock()

	systray.SetTooltip(Title)
	systray.SetOnTapped(t.activate)
	// On Linux Register only stores the callbacks, so the item starts with the
	// start function of RunWithExternalLoop. The GTK loop stays the event loop
	// of the app, and the stop the call returns is not kept: the end of the
	// process already takes the item away.
	start, _ := systray.RunWithExternalLoop(nil, nil)
	go start()
}

// Show puts a description on the tray. The menu is rebuilt only when the tree
// changes: an unchanged snapshot touches neither the menu nor the icon, so no
// entry of the menu on screen is taken away for nothing.
func (t *Tray) Show(items []tray.Item, icon tray.Icon) {
	nodes := Tree(items)

	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.hasIcon || icon != t.icon {
		systray.SetIcon(icon.PNG())
		t.icon, t.hasIcon = icon, true
	}
	if Equal(nodes, t.nodes) {
		return
	}
	t.nodes = nodes
	t.build(nodes)
}

// activate reports a left click of the icon.
func (t *Tray) activate() {
	if t.onActivate != nil {
		t.onActivate()
	}
}

// build replaces the whole menu with nodes. ResetMenu removes the entries of
// the menu before it, so none of them stays on the tray and their listeners end.
func (t *Tray) build(nodes []Node) {
	systray.ResetMenu()
	for _, node := range nodes {
		t.add(nil, node)
	}
}

// add appends one node at the top of the menu when parent is nil, or under it.
// Only an enabled entry without children gets a listener: the other entries
// launch nothing.
func (t *Tray) add(parent *systray.MenuItem, node Node) {
	if node.Separator {
		if parent == nil {
			systray.AddSeparator()
			return
		}
		parent.AddSeparator()
		return
	}

	var item *systray.MenuItem
	if parent == nil {
		item = systray.AddMenuItem(node.Label, "")
	} else {
		item = parent.AddSubMenuItem(node.Label, "")
	}
	if !node.Enabled {
		item.Disable()
	}
	if len(node.Children) > 0 {
		for _, child := range node.Children {
			t.add(item, child)
		}
		return
	}
	if !node.Enabled || t.onChoose == nil {
		return
	}
	go t.listen(item, node)
}

// listen reports every click on item until the item leaves the menu. Remove
// closes ClickedCh, so the listener ends with its entry and never leaks.
func (t *Tray) listen(item *systray.MenuItem, node Node) {
	for range item.ClickedCh {
		t.onChoose(node)
	}
}
