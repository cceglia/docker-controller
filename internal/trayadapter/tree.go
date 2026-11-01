// Package trayadapter puts the tray description of internal/tray on the system
// tray of fyne.io/systray. Half of it is a pure mapping to a neutral tree, which
// compares without touching systray; the other half is a thin layer that applies
// that tree. It imports no GTK and decides nothing: a click goes back to the
// caller, which runs it on the GTK loop.
package trayadapter

import (
	"dockercontroller/internal/compose"
	"dockercontroller/internal/tray"
)

// Node is one entry of the menu systray shows, built from a tray.Item.
type Node struct {
	Label     string
	Enabled   bool
	Separator bool
	// Action is the Action a click launches, nil for the other entries.
	Action *compose.Action
	// Confirm, when not empty, is the question to ask before launching Action.
	Confirm  string
	Command  tray.Command
	Children []Node
}

// Tree maps the description of the pure tray package to the neutral tree the
// systray layer applies. Every call returns fresh nodes: a tree never shares
// memory with the description it came from, and an Action is a copy, so a later
// snapshot cannot change a tree that is already on the tray.
func Tree(items []tray.Item) []Node {
	if len(items) == 0 {
		return nil
	}
	nodes := make([]Node, 0, len(items))
	for _, item := range items {
		node := Node{
			Label:     item.Label,
			Enabled:   item.Enabled,
			Separator: item.Separator,
			Confirm:   item.Confirm,
			Command:   item.Command,
		}
		if item.Action != nil {
			action := *item.Action
			node.Action = &action
		}
		node.Children = Tree(item.Children)
		nodes = append(nodes, node)
	}
	return nodes
}

// Equal reports whether two trees describe the same menu: the same entries, in
// the same order, with the same labels, states, questions, Commands and
// Actions. The adapter rebuilds the menu only when this answer is false.
func Equal(a, b []Node) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !sameNode(a[i], b[i]) {
			return false
		}
	}
	return true
}

// sameNode compares one entry and, when it has any, its children.
func sameNode(a, b Node) bool {
	if a.Label != b.Label || a.Enabled != b.Enabled || a.Separator != b.Separator ||
		a.Confirm != b.Confirm || a.Command != b.Command {
		return false
	}
	if (a.Action == nil) != (b.Action == nil) {
		return false
	}
	if a.Action != nil && *a.Action != *b.Action {
		return false
	}
	return Equal(a.Children, b.Children)
}
