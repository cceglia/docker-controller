package traymenu

import (
	"dockercontroller/internal/compose"
	guiproject "dockercontroller/internal/gui/project"
	"dockercontroller/internal/gui/view"
	"dockercontroller/internal/tray"
)

// The tooltip of the icon names the application and then the state of Docker in
// words, so the state is readable without looking at the icon.
const (
	trayTooltipLoading = "Controllo di Docker…"
	trayTooltipOnline  = "Docker risponde"
	trayTooltipOffline = "Docker non risponde"
)

// TrayNode is one entry of the menu on the tray, built from a tray.Item.
type TrayNode struct {
	Label     string
	Enabled   bool
	Separator bool
	// Action is the Action a click launches, nil for the other entries.
	Action *compose.Action
	// Confirm, when not empty, is the question to ask before launching Action.
	Confirm  string
	Command  tray.Command
	Children []TrayNode
}

// Tree maps the description of the pure tray package to the tree the Qt
// adapter applies. Every call returns fresh nodes: a tree never shares memory
// with the description it came from, and an Action is a copy, so a later
// snapshot cannot change a tree that is already on the tray.
func Tree(items []tray.Item) []TrayNode {
	if len(items) == 0 {
		return nil
	}
	nodes := make([]TrayNode, 0, len(items))
	for _, item := range items {
		node := TrayNode{
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

// SameTree reports whether two trees describe the same menu: the same entries,
// in the same order, with the same labels, states, questions, Commands and
// Actions. The adapter rebuilds the menu only when this answer is false.
func SameTree(a, b []TrayNode) bool {
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
func sameNode(a, b TrayNode) bool {
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
	return SameTree(a.Children, b.Children)
}

// Tooltip is the text of the tooltip of the icon: the name of the application
// and the state of Docker. Before the first snapshot Docker has not been asked
// yet, so the tooltip says the application is checking.
func Tooltip(s compose.Snapshot) string {
	switch {
	case s.Loading:
		return view.AppTitle + " · " + trayTooltipLoading
	case !s.DockerReachable:
		return view.AppTitle + " · " + trayTooltipOffline
	}
	return view.AppTitle + " · " + trayTooltipOnline
}

// Question is the confirmation an entry asks before launching its Action, nil
// when the entry asks nothing. The message is the one the tray description
// carries, and the label of the entry is the affirmative answer. Down con volumi
// also asks for the name of the Project, because it deletes data.
func Question(node TrayNode) *guiproject.Confirmation {
	if node.Action == nil || node.Confirm == "" {
		return nil
	}
	question := &guiproject.Confirmation{
		Title:        view.AppTitle,
		Message:      node.Confirm,
		ConfirmLabel: node.Label,
	}
	if node.Action.Kind == compose.ActionDownVolumes {
		question.TypeToConfirm = node.Action.Project
		question.Destructive = true
	}
	return question
}
