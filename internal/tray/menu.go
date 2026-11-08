// Package tray describes the menu of the system tray and decides which Activity
// outcomes deserve a desktop notification. It turns a compose.Snapshot into a
// plain menu description and holds no domain logic and does no polling. It has
// no Qt and no D-Bus: the GUI applies the menu to QSystemTrayIcon, and notify
// carries the notifications.
package tray

import (
	"sort"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/model"
)

// Command is what a menu entry that is not an Action does.
type Command int

const (
	CommandNone Command = iota
	CommandShowWindow
	CommandQuit
)

// Item is one entry of the tray menu. An Item with Children is a submenu.
type Item struct {
	Label     string
	Enabled   bool
	Separator bool
	// Action is the compose Action the entry launches, nil for other entries.
	Action *compose.Action
	// Confirm, when not empty, is the question to ask before launching Action.
	Confirm  string
	Command  Command
	Children []Item
}

// Icon selects the tray icon.
type Icon int

const (
	IconNormal Icon = iota
	IconOffline
)

// IconFor returns the icon that reflects whether Docker is reachable. Before
// the first refresh nothing is known, so it is not the offline one.
func IconFor(s compose.Snapshot) Icon {
	if s.Loading || s.DockerReachable {
		return IconNormal
	}
	return IconOffline
}

// BuildMenu describes the whole tray menu for a snapshot: Groups as submenus
// of Projects, ungrouped Projects at the top level, then "Mostra finestra" and
// "Esci". The Groups follow the order of the snapshot, an empty Group shows no
// submenu, and a Group the snapshot does not list goes after the known ones, in
// alphabetical order, and before the ungrouped Projects. Up, Stop, Restart and
// Down are offered on a Project; Start, Stop,
// Restart and Down on a Service. Down asks for confirmation (Item.Confirm) and
// never removes volumes: Down con volumi, Pull and Build are not in the tray.
// Every Action is disabled
// while Docker is unreachable and, for a Project, while it has an Activity.
func BuildMenu(s compose.Snapshot) []Item {
	var menu []Item
	if !s.Loading && !s.DockerReachable {
		menu = append(menu, Item{Label: "Docker non raggiungibile"}, Item{Separator: true})
	}

	byGroup := map[string][]Item{}
	var ungrouped []Item
	for _, p := range s.Projects {
		item := projectItem(p, s.DockerReachable)
		if p.Group == "" {
			ungrouped = append(ungrouped, item)
		} else {
			byGroup[p.Group] = append(byGroup[p.Group], item)
		}
	}
	known := make(map[string]bool, len(s.Groups))
	for _, g := range s.Groups {
		known[g] = true
		if children := byGroup[g]; len(children) > 0 {
			menu = append(menu, Item{Label: g, Enabled: true, Children: children})
		}
	}
	var unknown []string
	for name := range byGroup {
		if !known[name] {
			unknown = append(unknown, name)
		}
	}
	sort.Slice(unknown, func(i, j int) bool { return model.LessGroupName(unknown[i], unknown[j]) })
	for _, name := range unknown {
		menu = append(menu, Item{Label: name, Enabled: true, Children: byGroup[name]})
	}
	menu = append(menu, ungrouped...)

	if len(s.Projects) > 0 {
		menu = append(menu, Item{Separator: true})
	}
	return append(menu,
		Item{Label: "Mostra finestra", Enabled: true, Command: CommandShowWindow},
		Item{Label: "Esci", Enabled: true, Command: CommandQuit},
	)
}

func projectItem(p compose.Project, dockerReachable bool) Item {
	usable := dockerReachable && p.Activity == nil
	act := func(label string, service string, kind compose.ActionKind) Item {
		return Item{Label: label, Enabled: usable, Action: &compose.Action{Project: p.Name, Service: service, Kind: kind}}
	}
	down := func(label, service string, kind compose.ActionKind, confirm string) Item {
		it := act(label, service, kind)
		it.Confirm = confirm
		return it
	}

	children := []Item{
		act("Up", "", compose.ActionUp),
		act("Stop", "", compose.ActionStop),
		act("Restart", "", compose.ActionRestart),
		down("Down", "", compose.ActionDown, "Fermare e rimuovere i Container di "+displayName(p)+"? I volumi restano."),
	}
	if len(p.Services) > 0 {
		services := make([]Item, 0, len(p.Services))
		for _, svc := range p.Services {
			services = append(services, Item{Label: svc.Name, Enabled: usable, Children: []Item{
				act("Start", svc.Name, compose.ActionStart),
				act("Stop", svc.Name, compose.ActionStop),
				act("Restart", svc.Name, compose.ActionRestart),
				down("Down", svc.Name, compose.ActionDownService, "Fermare e rimuovere i Container del Servizio "+svc.Name+" di "+displayName(p)+"? I volumi restano."),
			}})
		}
		children = append(children, Item{Separator: true}, Item{Label: "Servizi", Enabled: usable, Children: services})
	}
	return Item{Label: symbol(p.Status) + " " + displayName(p), Enabled: true, Children: children}
}

func displayName(p compose.Project) string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	return p.Name
}

// symbol is the state marker: ● Attivo, ◐ Parziale, ○ Fermo or Rimosso.
func symbol(s compose.Status) string {
	switch s {
	case compose.StatusActive:
		return "●"
	case compose.StatusPartial:
		return "◐"
	default:
		return "○"
	}
}
