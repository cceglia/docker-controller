package gui

import (
	"reflect"
	"testing"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/tray"
)

// trayAction copies a, so a table can spell an Action out without sharing it.
func trayAction(a compose.Action) *compose.Action {
	return &a
}

func TestTrayTree(t *testing.T) {
	tests := []struct {
		name  string
		items []tray.Item
		want  []trayNode
	}{
		{
			name: "an empty description gives no node",
		},
		{
			name:  "a separator",
			items: []tray.Item{{Separator: true}},
			want:  []trayNode{{Separator: true}},
		},
		{
			name:  "a notice is a disabled leaf",
			items: []tray.Item{{Label: "Docker non raggiungibile"}},
			want:  []trayNode{{Label: "Docker non raggiungibile"}},
		},
		{
			name: "an Action with its question",
			items: []tray.Item{{
				Label:   "Down",
				Enabled: true,
				Action:  &compose.Action{Project: "shop", Kind: compose.ActionDown},
				Confirm: "Fermare e rimuovere i Container di shop? I volumi restano.",
			}},
			want: []trayNode{{
				Label:   "Down",
				Enabled: true,
				Action:  trayAction(compose.Action{Project: "shop", Kind: compose.ActionDown}),
				Confirm: "Fermare e rimuovere i Container di shop? I volumi restano.",
			}},
		},
		{
			name:  "a Command",
			items: []tray.Item{{Label: "Esci", Enabled: true, Command: tray.CommandQuit}},
			want:  []trayNode{{Label: "Esci", Enabled: true, Command: tray.CommandQuit}},
		},
		{
			name: "a submenu keeps the order of its children",
			items: []tray.Item{{
				Label:   "Lavoro",
				Enabled: true,
				Children: []tray.Item{
					{Label: "● shop", Enabled: true},
					{Label: "○ blog", Enabled: true},
				},
			}},
			want: []trayNode{{
				Label:   "Lavoro",
				Enabled: true,
				Children: []trayNode{
					{Label: "● shop", Enabled: true},
					{Label: "○ blog", Enabled: true},
				},
			}},
		},
		{
			name: "project, service and action nest three deep",
			items: []tray.Item{{
				Label:   "● shop",
				Enabled: true,
				Children: []tray.Item{
					{Separator: true},
					{Label: "Servizi", Enabled: true, Children: []tray.Item{
						{Label: "web", Enabled: true, Children: []tray.Item{
							{Label: "Start", Enabled: true, Action: &compose.Action{Project: "shop", Service: "web", Kind: compose.ActionStart}},
						}},
					}},
				},
			}},
			want: []trayNode{{
				Label:   "● shop",
				Enabled: true,
				Children: []trayNode{
					{Separator: true},
					{Label: "Servizi", Enabled: true, Children: []trayNode{
						{Label: "web", Enabled: true, Children: []trayNode{
							{Label: "Start", Enabled: true, Action: trayAction(compose.Action{Project: "shop", Service: "web", Kind: compose.ActionStart})},
						}},
					}},
				},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trayTree(tt.items)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("trayTree() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// A tree lasts across snapshots, so a later change to the description it came
// from must not reach it.
func TestTrayTree_DoesNotShareMemoryWithTheDescription(t *testing.T) {
	up := compose.Action{Project: "shop", Kind: compose.ActionUp}
	items := []tray.Item{{
		Label:   "● shop",
		Enabled: true,
		Children: []tray.Item{
			{Label: "Up", Enabled: true, Action: &up},
		},
	}}

	nodes := trayTree(items)
	items[0].Label = "cambiato"
	items[0].Children[0].Label = "cambiato"
	up.Kind = compose.ActionDown

	want := []trayNode{{
		Label:   "● shop",
		Enabled: true,
		Children: []trayNode{
			{Label: "Up", Enabled: true, Action: trayAction(compose.Action{Project: "shop", Kind: compose.ActionUp})},
		},
	}}
	if !reflect.DeepEqual(nodes, want) {
		t.Errorf("trayTree() = %+v, want %+v", nodes, want)
	}
}

// The real description of the menu keeps its hierarchy in the tree the adapter
// applies: a Group, a Project, its Services and their Actions.
func TestTrayTree_BuildMenuKeepsTheHierarchy(t *testing.T) {
	snapshot := compose.Snapshot{
		DockerReachable: true,
		Groups:          []string{"Lavoro"},
		Projects: []compose.Project{{
			Name:        "shop",
			DisplayName: "Negozio",
			Group:       "Lavoro",
			Status:      compose.StatusActive,
			Services:    []compose.Service{{Name: "web"}},
		}},
	}

	nodes := trayTree(tray.BuildMenu(snapshot))
	if len(nodes) != 4 {
		t.Fatalf("top level = %d entries, want the Group, a separator, Mostra finestra and Esci", len(nodes))
	}
	if got := nodes[0].Label; got != "Lavoro" {
		t.Fatalf("first Group = %q, want %q", got, "Lavoro")
	}
	projects := nodes[0].Children
	if len(projects) != 1 || projects[0].Label != "● Negozio" {
		t.Fatalf("Group children = %+v, want the Project", projects)
	}
	children := projects[0].Children
	labels := make([]string, 0, len(children))
	for _, child := range children {
		labels = append(labels, child.Label)
	}
	want := []string{"Up", "Stop", "Restart", "Down", "", "Servizi"}
	if !reflect.DeepEqual(labels, want) {
		t.Fatalf("Project children = %q, want %q", labels, want)
	}
	services := children[len(children)-1].Children
	if len(services) != 1 || services[0].Label != "web" {
		t.Fatalf("Servizi children = %+v, want the Service", services)
	}
	if got := services[0].Children[0].Action; got == nil || *got != (compose.Action{Project: "shop", Service: "web", Kind: compose.ActionStart}) {
		t.Fatalf("first Action of the Service = %+v, want Start", got)
	}
	if !nodes[1].Separator {
		t.Fatalf("second entry = %+v, want a separator", nodes[1])
	}
	if got := nodes[2].Command; got != tray.CommandShowWindow {
		t.Fatalf("third entry = Command %d, want Mostra finestra", got)
	}
	if got := nodes[3].Command; got != tray.CommandQuit {
		t.Fatalf("last entry = Command %d, want Esci", got)
	}
}

func TestSameTrayTree(t *testing.T) {
	base := func() []trayNode {
		return []trayNode{{
			Label:   "● shop",
			Enabled: true,
			Children: []trayNode{
				{Label: "Up", Enabled: true, Action: trayAction(compose.Action{Project: "shop", Kind: compose.ActionUp})},
				{Label: "Down", Enabled: false, Confirm: "Fermare?", Action: trayAction(compose.Action{Project: "shop", Kind: compose.ActionDown})},
				{Separator: true},
				{Label: "Esci", Enabled: true, Command: tray.CommandQuit},
			},
		}}
	}

	tests := []struct {
		name string
		b    []trayNode
		want bool
	}{
		{
			name: "the same tree",
			b:    base(),
			want: true,
		},
		{
			name: "an empty tree against the full one",
			b:    nil,
			want: false,
		},
		{
			name: "a different label",
			b: func() []trayNode {
				tree := base()
				tree[0].Children[0].Label = "Start"
				return tree
			}(),
			want: false,
		},
		{
			name: "a different state",
			b: func() []trayNode {
				tree := base()
				tree[0].Children[0].Enabled = false
				return tree
			}(),
			want: false,
		},
		{
			name: "a different question",
			b: func() []trayNode {
				tree := base()
				tree[0].Children[1].Confirm = "Fermare e rimuovere?"
				return tree
			}(),
			want: false,
		},
		{
			name: "a different Command",
			b: func() []trayNode {
				tree := base()
				tree[0].Children[3].Command = tray.CommandShowWindow
				return tree
			}(),
			want: false,
		},
		{
			name: "an Action against no Action",
			b: func() []trayNode {
				tree := base()
				tree[0].Children[0].Action = nil
				return tree
			}(),
			want: false,
		},
		{
			name: "a different Action",
			b: func() []trayNode {
				tree := base()
				tree[0].Children[0].Action = trayAction(compose.Action{Project: "shop", Kind: compose.ActionStop})
				return tree
			}(),
			want: false,
		},
		{
			name: "a different Service of the same Action",
			b: func() []trayNode {
				tree := base()
				tree[0].Children[0].Action = trayAction(compose.Action{Project: "shop", Service: "web", Kind: compose.ActionUp})
				return tree
			}(),
			want: false,
		},
		{
			name: "a separator against an entry",
			b: func() []trayNode {
				tree := base()
				tree[0].Children[2] = trayNode{Label: "Servizi", Enabled: true}
				return tree
			}(),
			want: false,
		},
		{
			name: "a missing child",
			b: func() []trayNode {
				tree := base()
				tree[0].Children = tree[0].Children[:2]
				return tree
			}(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sameTrayTree(base(), tt.b); got != tt.want {
				t.Errorf("sameTrayTree() = %v, want %v", got, tt.want)
			}
			if got := sameTrayTree(tt.b, base()); got != tt.want {
				t.Errorf("sameTrayTree() with the operands swapped = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrayTooltip_NamesTheApplicationAndTheStateOfDocker(t *testing.T) {
	for _, tc := range []struct {
		name     string
		snapshot compose.Snapshot
		want     string
	}{
		{"before the first snapshot", compose.Snapshot{Loading: true}, "Docker Controller · Controllo di Docker…"},
		{"Docker answers", compose.Snapshot{DockerReachable: true}, "Docker Controller · Docker risponde"},
		{"Docker does not answer", compose.Snapshot{DockerReachable: false}, "Docker Controller · Docker non risponde"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := trayTooltip(tc.snapshot); got != tc.want {
				t.Errorf("trayTooltip() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTrayQuestion(t *testing.T) {
	for _, tc := range []struct {
		name string
		node trayNode
		want *Confirmation
	}{
		{
			name: "an Action without a question launches at once",
			node: trayNode{Label: "Up", Enabled: true, Action: trayAction(compose.Action{Project: "shop", Kind: compose.ActionUp})},
		},
		{
			name: "a notice asks nothing",
			node: trayNode{Label: "Docker non raggiungibile"},
		},
		{
			name: "Down of a Project",
			node: trayNode{
				Label:   "Down",
				Enabled: true,
				Confirm: "Fermare e rimuovere i Container di shop? I volumi restano.",
				Action:  trayAction(compose.Action{Project: "shop", Kind: compose.ActionDown}),
			},
			want: &Confirmation{
				Title:        "Docker Controller",
				Message:      "Fermare e rimuovere i Container di shop? I volumi restano.",
				ConfirmLabel: "Down",
			},
		},
		{
			name: "Down of a Service",
			node: trayNode{
				Label:   "Down",
				Enabled: true,
				Confirm: "Fermare e rimuovere i Container del Servizio web di shop? I volumi restano.",
				Action:  trayAction(compose.Action{Project: "shop", Service: "web", Kind: compose.ActionDownService}),
			},
			want: &Confirmation{
				Title:        "Docker Controller",
				Message:      "Fermare e rimuovere i Container del Servizio web di shop? I volumi restano.",
				ConfirmLabel: "Down",
			},
		},
		{
			name: "Down con volumi asks for the name of the Project",
			node: trayNode{
				Label:   "Down con volumi",
				Enabled: true,
				Confirm: "Rimuove i Container e ELIMINA i volumi del Progetto: i dati nei volumi andranno persi per sempre.",
				Action:  trayAction(compose.Action{Project: "shop", Kind: compose.ActionDownVolumes}),
			},
			want: &Confirmation{
				Title:         "Docker Controller",
				Message:       "Rimuove i Container e ELIMINA i volumi del Progetto: i dati nei volumi andranno persi per sempre.",
				ConfirmLabel:  "Down con volumi",
				TypeToConfirm: "shop",
				Destructive:   true,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := trayQuestion(tc.node)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("trayQuestion() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
