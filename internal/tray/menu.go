// Package tray is the pure part of the system tray adapter: it turns a
// compose.Snapshot into a plain menu description and decides which Activity
// outcomes deserve a desktop notification. It holds no domain logic, does no
// polling and does not import Wails; the Wails glue lives in the main package.
package tray

import "dockercontroller/internal/compose"

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
	Action   *compose.Action
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
// "Esci". Only non-destructive Actions are ever offered: Up, Stop and Restart
// on a Project, Start, Stop and Restart on a Service. Every Action is disabled
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
	for _, g := range s.Groups {
		if children := byGroup[g]; len(children) > 0 {
			menu = append(menu, Item{Label: g, Enabled: true, Children: children})
		}
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

	children := []Item{
		act("Up", "", compose.ActionUp),
		act("Stop", "", compose.ActionStop),
		act("Restart", "", compose.ActionRestart),
	}
	if len(p.Services) > 0 {
		services := make([]Item, 0, len(p.Services))
		for _, svc := range p.Services {
			services = append(services, Item{Label: svc.Name, Enabled: usable, Children: []Item{
				act("Start", svc.Name, compose.ActionStart),
				act("Stop", svc.Name, compose.ActionStop),
				act("Restart", svc.Name, compose.ActionRestart),
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
