package tray

import (
	"strings"
	"testing"

	"dockercontroller/internal/compose"
)

func project(name, display, group string, status compose.Status, services ...string) compose.Project {
	p := compose.Project{Name: name, DisplayName: display, Group: group, Status: status, Services: []compose.Service{}}
	for _, s := range services {
		p.Services = append(p.Services, compose.Service{Name: s})
	}
	return p
}

func snapshot(reachable bool, projects ...compose.Project) compose.Snapshot {
	s := compose.Snapshot{DockerReachable: reachable, Projects: projects, Groups: []string{}}
	seen := map[string]bool{}
	for _, p := range projects {
		if p.Group != "" && !seen[p.Group] {
			seen[p.Group] = true
			s.Groups = append(s.Groups, p.Group)
		}
	}
	return s
}

func find(items []Item, label string) *Item {
	for i := range items {
		if items[i].Label == label {
			return &items[i]
		}
	}
	return nil
}

func labels(items []Item) []string {
	var out []string
	for _, it := range items {
		if !it.Separator {
			out = append(out, it.Label)
		}
	}
	return out
}

func TestBuildMenu_AlwaysEndsWithShowWindowAndQuit(t *testing.T) {
	for _, reachable := range []bool{true, false} {
		m := BuildMenu(snapshot(reachable))
		if len(m) < 2 {
			t.Fatalf("menu too short: %v", labels(m))
		}
		show, quit := m[len(m)-2], m[len(m)-1]
		if show.Label != "Mostra finestra" || show.Command != CommandShowWindow || !show.Enabled {
			t.Errorf("reachable=%v: show = %+v", reachable, show)
		}
		if quit.Label != "Esci" || quit.Command != CommandQuit || !quit.Enabled {
			t.Errorf("reachable=%v: quit = %+v", reachable, quit)
		}
	}
}

func TestBuildMenu_StateSymbolsAndDisplayName(t *testing.T) {
	m := BuildMenu(snapshot(true,
		project("a", "", "", compose.StatusActive),
		project("b", "Bravo", "", compose.StatusPartial),
		project("c", "", "", compose.StatusStopped),
		project("d", "", "", compose.StatusRemoved),
	))
	for _, label := range []string{"● a", "◐ Bravo", "○ c", "○ d"} {
		if find(m, label) == nil {
			t.Errorf("missing %q in %v", label, labels(m))
		}
	}
}

func TestBuildMenu_UngroupedAtTopLevelGroupedUnderTheirGroup(t *testing.T) {
	m := BuildMenu(snapshot(true,
		project("solo", "", "", compose.StatusActive),
		project("w1", "", "Lavoro", compose.StatusActive),
		project("w2", "", "Lavoro", compose.StatusStopped),
		project("h1", "", "Casa", compose.StatusStopped),
	))
	if find(m, "● solo") == nil {
		t.Fatalf("ungrouped project not at top level: %v", labels(m))
	}
	if find(m, "● w1") != nil {
		t.Errorf("grouped project leaked to top level: %v", labels(m))
	}
	lavoro := find(m, "Lavoro")
	if lavoro == nil || len(lavoro.Children) != 2 || !lavoro.Enabled {
		t.Fatalf("Lavoro group = %+v", lavoro)
	}
	if find(lavoro.Children, "● w1") == nil || find(lavoro.Children, "○ w2") == nil {
		t.Errorf("Lavoro children = %v", labels(lavoro.Children))
	}
	if find(m, "Casa") == nil {
		t.Errorf("Casa group missing: %v", labels(m))
	}
}

func TestBuildMenu_ProjectActionsAndServices(t *testing.T) {
	m := BuildMenu(snapshot(true, project("vpn", "VPN", "", compose.StatusActive, "web", "db")))
	p := find(m, "● VPN")
	if p == nil {
		t.Fatalf("no project: %v", labels(m))
	}
	want := map[string]compose.ActionKind{"Up": compose.ActionUp, "Stop": compose.ActionStop, "Restart": compose.ActionRestart, "Down": compose.ActionDown}
	for label, kind := range want {
		it := find(p.Children, label)
		if it == nil || !it.Enabled || it.Action == nil || *it.Action != (compose.Action{Project: "vpn", Kind: kind}) {
			t.Errorf("%s = %+v", label, it)
		}
	}
	services := find(p.Children, "Servizi")
	if services == nil || len(services.Children) != 2 {
		t.Fatalf("Servizi = %+v", services)
	}
	web := find(services.Children, "web")
	if web == nil {
		t.Fatalf("services = %v", labels(services.Children))
	}
	wantSvc := map[string]compose.ActionKind{"Start": compose.ActionStart, "Stop": compose.ActionStop, "Restart": compose.ActionRestart, "Down": compose.ActionDownService}
	if len(web.Children) != len(wantSvc) {
		t.Errorf("service entries = %v", labels(web.Children))
	}
	for label, kind := range wantSvc {
		it := find(web.Children, label)
		if it == nil || !it.Enabled || it.Action == nil || *it.Action != (compose.Action{Project: "vpn", Service: "web", Kind: kind}) {
			t.Errorf("web %s = %+v", label, it)
		}
	}
}

func TestBuildMenu_ProjectWithoutServicesHasNoServicesSubmenu(t *testing.T) {
	m := BuildMenu(snapshot(true, project("x", "", "", compose.StatusRemoved)))
	if find(find(m, "○ x").Children, "Servizi") != nil {
		t.Error("Servizi shown for a Project without Services")
	}
}

func walk(items []Item, f func(Item)) {
	for _, it := range items {
		f(it)
		walk(it.Children, f)
	}
}

func TestBuildMenu_ActivityDisablesOnlyThatProject(t *testing.T) {
	busy := project("busy", "", "", compose.StatusActive, "web")
	busy.Activity = &compose.ActivityRef{ID: "act-1", Kind: compose.ActionUp}
	idle := project("idle", "", "", compose.StatusActive, "web")
	m := BuildMenu(snapshot(true, busy, idle))

	walk(find(m, "● busy").Children, func(it Item) {
		if it.Action != nil && it.Enabled {
			t.Errorf("busy: %q is enabled", it.Label)
		}
	})
	enabled := 0
	walk(find(m, "● idle").Children, func(it Item) {
		if it.Action != nil && it.Enabled {
			enabled++
		}
	})
	if enabled != 8 {
		t.Errorf("idle: %d enabled actions, want 8", enabled)
	}
}

func TestBuildMenu_DockerUnreachableDisablesEveryAction(t *testing.T) {
	m := BuildMenu(snapshot(false, project("a", "", "", compose.StatusRemoved, "web"), project("g", "", "G", compose.StatusStopped, "web")))
	actions := 0
	walk(m, func(it Item) {
		if it.Action != nil {
			actions++
			if it.Enabled {
				t.Errorf("%q enabled while Docker is unreachable", it.Label)
			}
		}
	})
	if actions == 0 {
		t.Error("projects should still be listed, with disabled actions")
	}
	if m[0].Label != "Docker non raggiungibile" || m[0].Enabled {
		t.Errorf("first item = %+v", m[0])
	}
}

func TestBuildMenu_NeverOffersDownWithVolumesPullOrBuild(t *testing.T) {
	p := project("p", "", "G", compose.StatusActive, "web")
	m := BuildMenu(snapshot(true, p, project("q", "", "", compose.StatusStopped, "db")))
	banned := map[compose.ActionKind]bool{compose.ActionDownVolumes: true, compose.ActionPull: true, compose.ActionBuild: true}
	walk(m, func(it Item) {
		if it.Action != nil && banned[it.Action.Kind] {
			t.Errorf("%q offers %s", it.Label, it.Action.Kind)
		}
		l := strings.ToLower(it.Label)
		for _, w := range []string{"pull", "build", "volum"} {
			if strings.Contains(l, w) {
				t.Errorf("label %q mentions %s", it.Label, w)
			}
		}
	})
}

func TestBuildMenu_DownOfAProjectAndOfAServiceAsksForConfirmation(t *testing.T) {
	m := BuildMenu(snapshot(true, project("p", "", "", compose.StatusActive, "web")))
	downs := map[compose.ActionKind]Item{}
	walk(m, func(it Item) {
		if it.Action != nil && (it.Action.Kind == compose.ActionDown || it.Action.Kind == compose.ActionDownService) {
			downs[it.Action.Kind] = it
		}
	})
	for _, kind := range []compose.ActionKind{compose.ActionDown, compose.ActionDownService} {
		it, ok := downs[kind]
		if !ok {
			t.Errorf("no %s entry in the menu", kind)
			continue
		}
		if it.Label != "Down" || it.Confirm == "" {
			t.Errorf("%s entry = %+v, want label Down and a confirmation", kind, it)
		}
	}
	if downs[compose.ActionDownService].Action.Service != "web" || downs[compose.ActionDown].Action.Service != "" {
		t.Errorf("Down targets wrong: %+v", downs)
	}
}

func TestIconFor(t *testing.T) {
	if IconFor(snapshot(true)) != IconNormal {
		t.Error("reachable Docker should use the normal icon")
	}
	if IconFor(snapshot(false)) != IconOffline {
		t.Error("unreachable Docker should use the offline icon")
	}
}

func TestIconFor_NoOfflineIconBeforeTheFirstRefresh(t *testing.T) {
	loading := compose.Snapshot{Loading: true}
	if IconFor(loading) != IconNormal {
		t.Error("a Loading snapshot must not show the offline icon")
	}
	if IconFor(compose.Snapshot{}) != IconOffline {
		t.Error("an unreachable Docker must show the offline icon")
	}
	if find(BuildMenu(loading), "Docker non raggiungibile") != nil {
		t.Error("a Loading snapshot must not say Docker is unreachable")
	}
}
