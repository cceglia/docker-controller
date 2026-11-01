package trayadapter

import (
	"reflect"
	"testing"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/tray"
)

// action copies a, so a table can spell an Action out without sharing it.
func action(a compose.Action) *compose.Action {
	return &a
}

func TestTree(t *testing.T) {
	tests := []struct {
		name  string
		items []tray.Item
		want  []Node
	}{
		{
			name: "an empty description gives no node",
		},
		{
			name:  "a separator",
			items: []tray.Item{{Separator: true}},
			want:  []Node{{Separator: true}},
		},
		{
			name:  "a notice is a disabled leaf",
			items: []tray.Item{{Label: "Docker non raggiungibile"}},
			want:  []Node{{Label: "Docker non raggiungibile"}},
		},
		{
			name: "an Action with its question",
			items: []tray.Item{{
				Label:   "Down",
				Enabled: true,
				Action:  &compose.Action{Project: "shop", Kind: compose.ActionDown},
				Confirm: "Fermare e rimuovere i Container di shop? I volumi restano.",
			}},
			want: []Node{{
				Label:   "Down",
				Enabled: true,
				Action:  action(compose.Action{Project: "shop", Kind: compose.ActionDown}),
				Confirm: "Fermare e rimuovere i Container di shop? I volumi restano.",
			}},
		},
		{
			name:  "a Command",
			items: []tray.Item{{Label: "Esci", Enabled: true, Command: tray.CommandQuit}},
			want:  []Node{{Label: "Esci", Enabled: true, Command: tray.CommandQuit}},
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
			want: []Node{{
				Label:   "Lavoro",
				Enabled: true,
				Children: []Node{
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
			want: []Node{{
				Label:   "● shop",
				Enabled: true,
				Children: []Node{
					{Separator: true},
					{Label: "Servizi", Enabled: true, Children: []Node{
						{Label: "web", Enabled: true, Children: []Node{
							{Label: "Start", Enabled: true, Action: action(compose.Action{Project: "shop", Service: "web", Kind: compose.ActionStart})},
						}},
					}},
				},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Tree(tt.items)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Tree() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// A tree lasts across snapshots, so a later change to the description it came
// from must not reach it.
func TestTree_DoesNotShareMemoryWithTheDescription(t *testing.T) {
	up := compose.Action{Project: "shop", Kind: compose.ActionUp}
	items := []tray.Item{{
		Label:   "● shop",
		Enabled: true,
		Children: []tray.Item{
			{Label: "Up", Enabled: true, Action: &up},
		},
	}}

	nodes := Tree(items)
	items[0].Label = "cambiato"
	items[0].Children[0].Label = "cambiato"
	up.Kind = compose.ActionDown

	want := []Node{{
		Label:   "● shop",
		Enabled: true,
		Children: []Node{
			{Label: "Up", Enabled: true, Action: action(compose.Action{Project: "shop", Kind: compose.ActionUp})},
		},
	}}
	if !reflect.DeepEqual(nodes, want) {
		t.Errorf("Tree() = %+v, want %+v", nodes, want)
	}
}

func TestEqual(t *testing.T) {
	menu := func() []Node {
		return []Node{
			{Label: "● shop", Enabled: true, Children: []Node{
				{Label: "Up", Enabled: true, Action: action(compose.Action{Project: "shop", Kind: compose.ActionUp})},
				{Label: "Down", Enabled: true, Action: action(compose.Action{Project: "shop", Kind: compose.ActionDown}), Confirm: "Down?"},
			}},
			{Separator: true},
			{Label: "Esci", Enabled: true, Command: tray.CommandQuit},
		}
	}

	tests := []struct {
		name string
		a    []Node
		b    []Node
		want bool
	}{
		{
			name: "two empty trees",
			want: true,
		},
		{
			name: "an empty tree and one entry",
			a:    []Node{{Label: "Esci", Enabled: true}},
		},
		{
			name: "the same menu",
			a:    menu(),
			b:    menu(),
			want: true,
		},
		{
			name: "a different label",
			a:    menu(),
			b:    []Node{{Label: "○ shop", Enabled: true}},
		},
		{
			name: "a different state",
			a:    []Node{{Label: "Esci", Enabled: true}},
			b:    []Node{{Label: "Esci"}},
		},
		{
			name: "a question appeared",
			a:    []Node{{Label: "Down", Enabled: true}},
			b:    []Node{{Label: "Down", Enabled: true, Confirm: "Down?"}},
		},
		{
			name: "a different Command",
			a:    []Node{{Label: "Finestra", Enabled: true, Command: tray.CommandShowWindow}},
			b:    []Node{{Label: "Finestra", Enabled: true, Command: tray.CommandQuit}},
		},
		{
			name: "a different Action",
			a:    []Node{{Label: "Azione", Enabled: true, Action: action(compose.Action{Project: "shop", Kind: compose.ActionUp})}},
			b:    []Node{{Label: "Azione", Enabled: true, Action: action(compose.Action{Project: "shop", Kind: compose.ActionRestart})}},
		},
		{
			name: "an Action appeared",
			a:    []Node{{Label: "Azione", Enabled: true}},
			b:    []Node{{Label: "Azione", Enabled: true, Action: action(compose.Action{Project: "shop", Kind: compose.ActionUp})}},
		},
		{
			name: "a child changed",
			a:    menu(),
			b:    append(menu()[:1], []Node{{Separator: true}, {Label: "Esci", Enabled: true, Command: tray.CommandShowWindow}}...),
		},
		{
			name: "a child was added",
			a:    []Node{{Label: "shop", Enabled: true, Children: []Node{{Label: "Up", Enabled: true}}}},
			b:    []Node{{Label: "shop", Enabled: true, Children: []Node{{Label: "Up", Enabled: true}, {Label: "Stop", Enabled: true}}}},
		},
		{
			name: "the children are in another order",
			a:    []Node{{Label: "shop", Enabled: true, Children: []Node{{Label: "Up", Enabled: true}, {Label: "Stop", Enabled: true}}}},
			b:    []Node{{Label: "shop", Enabled: true, Children: []Node{{Label: "Stop", Enabled: true}, {Label: "Up", Enabled: true}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Equal(tt.a, tt.b); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
			if got := Equal(tt.b, tt.a); got != tt.want {
				t.Errorf("Equal() with the arguments swapped = %v, want %v", got, tt.want)
			}
		})
	}
}

// The adapter rebuilds only when the tree changes, so the tree of a snapshot
// must stay equal to itself and differ when Docker goes away.
func TestTree_FollowsTheSnapshot(t *testing.T) {
	reachable := compose.Snapshot{
		DockerReachable: true,
		Projects:        []compose.Project{{Name: "shop", Status: compose.StatusActive}},
		Groups:          []string{},
	}
	offline := reachable
	offline.DockerReachable = false

	if !Equal(Tree(tray.BuildMenu(reachable)), Tree(tray.BuildMenu(reachable))) {
		t.Error("the same snapshot gave two different trees")
	}
	if Equal(Tree(tray.BuildMenu(reachable)), Tree(tray.BuildMenu(offline))) {
		t.Error("a snapshot with Docker unreachable left the tree unchanged")
	}
}
