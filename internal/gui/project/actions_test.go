package project

import (
	"testing"

	"dockercontroller/internal/compose"
)

func TestActionLabel_NamesEveryKind(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind compose.ActionKind
		want string
	}{
		{"up", compose.ActionUp, "Up"},
		{"start", compose.ActionStart, "Start"},
		{"stop", compose.ActionStop, "Stop"},
		{"restart", compose.ActionRestart, "Restart"},
		{"pull", compose.ActionPull, "Pull"},
		{"build", compose.ActionBuild, "Build"},
		{"down", compose.ActionDown, "Down"},
		{"down with volumes", compose.ActionDownVolumes, "Down con volumi"},
		{"down of a service", compose.ActionDownService, "Down"},
		{"unknown keeps its own name", compose.ActionKind("inspect"), "inspect"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ActionLabel(tc.kind); got != tc.want {
				t.Errorf("ActionLabel(%q) = %q, want %q", tc.kind, got, tc.want)
			}
		})
	}
}

func TestProjectActions_OrdersTheActionsByTheStatus(t *testing.T) {
	for _, tc := range []struct {
		name    string
		project compose.Project
		want    []compose.ActionKind
	}{
		{
			"stopped starts with Up",
			compose.Project{Status: compose.StatusStopped},
			[]compose.ActionKind{compose.ActionUp, compose.ActionStop, compose.ActionRestart, compose.ActionPull, compose.ActionBuild, compose.ActionDown, compose.ActionDownVolumes},
		},
		{
			"removed starts with Up",
			compose.Project{Status: compose.StatusRemoved},
			[]compose.ActionKind{compose.ActionUp, compose.ActionStop, compose.ActionRestart, compose.ActionPull, compose.ActionBuild, compose.ActionDown, compose.ActionDownVolumes},
		},
		{
			"active starts with Stop",
			compose.Project{Status: compose.StatusActive},
			[]compose.ActionKind{compose.ActionStop, compose.ActionUp, compose.ActionRestart, compose.ActionPull, compose.ActionBuild, compose.ActionDown, compose.ActionDownVolumes},
		},
		{
			"partial starts with Restart",
			compose.Project{Status: compose.StatusPartial},
			[]compose.ActionKind{compose.ActionRestart, compose.ActionUp, compose.ActionStop, compose.ActionPull, compose.ActionBuild, compose.ActionDown, compose.ActionDownVolumes},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ProjectActions(tc.project, true)
			if len(got) != len(tc.want) {
				t.Fatalf("len = %d, want %d", len(got), len(tc.want))
			}
			for i, want := range tc.want {
				if got[i].Action.Kind != want {
					t.Errorf("Action %d = %q, want %q", i, got[i].Action.Kind, want)
				}
				if got[i].Label != ActionLabel(want) {
					t.Errorf("Label %d = %q, want %q", i, got[i].Label, ActionLabel(want))
				}
			}
		})
	}
}

func TestProjectActions_TargetsTheWholeProject(t *testing.T) {
	for _, view := range ProjectActions(compose.Project{Name: "shop"}, true) {
		t.Run(string(view.Action.Kind), func(t *testing.T) {
			if view.Action.Project != "shop" {
				t.Errorf("Project = %q, want %q", view.Action.Project, "shop")
			}
			if view.Action.Service != "" {
				t.Errorf("Service = %q, want no Service", view.Action.Service)
			}
		})
	}
}

func TestProjectActions_EnablesTheActionsOnlyWhenTheProjectIsFree(t *testing.T) {
	busy := &compose.ActivityRef{ID: "act-1", Kind: compose.ActionUp}
	for _, tc := range []struct {
		name            string
		project         compose.Project
		dockerReachable bool
		want            bool
	}{
		{"docker answers and no Activity", compose.Project{}, true, true},
		{"an Activity runs", compose.Project{Activity: busy}, true, false},
		{"docker does not answer", compose.Project{}, false, false},
		{"an Activity runs and docker does not answer", compose.Project{Activity: busy}, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, view := range ProjectActions(tc.project, tc.dockerReachable) {
				if view.Enabled != tc.want {
					t.Errorf("%s Enabled = %v, want %v", view.Action.Kind, view.Enabled, tc.want)
				}
			}
		})
	}
}

func TestProjectActions_ConfirmsOnlyTheDestructiveActions(t *testing.T) {
	project := compose.Project{Name: "shop", DisplayName: "Negozio"}
	var down, volumes bool
	for _, view := range ProjectActions(project, true) {
		switch view.Action.Kind {
		case compose.ActionDown:
			down = true
			assertConfirmation(t, view, Confirmation{
				Title:        "Down di Negozio",
				Message:      "Ferma e rimuove i Container del Progetto. I volumi restano. Il Progetto resta in lista come Rimosso.",
				ConfirmLabel: "Down",
			})
		case compose.ActionDownVolumes:
			volumes = true
			assertConfirmation(t, view, Confirmation{
				Title:         "Down con volumi di Negozio",
				Message:       "Rimuove i Container e ELIMINA i volumi del Progetto: i dati nei volumi andranno persi per sempre.",
				ConfirmLabel:  "Down con volumi",
				TypeToConfirm: "shop",
				Destructive:   true,
			})
		default:
			if view.Confirmation != nil {
				t.Errorf("%s asks for a confirmation, want none", view.Action.Kind)
			}
		}
	}
	if !down || !volumes {
		t.Errorf("ProjectActions offers Down %v and Down con volumi %v, want both", down, volumes)
	}
}

func assertConfirmation(t *testing.T, view ActionView, want Confirmation) {
	t.Helper()
	if view.Confirmation == nil {
		t.Fatalf("%s asks for no confirmation, want %+v", view.Action.Kind, want)
	}
	if *view.Confirmation != want {
		t.Errorf("%s Confirmation = %+v, want %+v", view.Action.Kind, *view.Confirmation, want)
	}
}

func TestServiceActions_OffersStartStopRestartAndDown(t *testing.T) {
	project := compose.Project{Name: "shop"}
	service := compose.Service{Name: "web"}

	got := ServiceActions(project, service, true)
	want := []compose.ActionKind{compose.ActionStart, compose.ActionStop, compose.ActionRestart, compose.ActionDownService}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i, kind := range want {
		if got[i].Action.Kind != kind {
			t.Errorf("Action %d = %q, want %q", i, got[i].Action.Kind, kind)
		}
		if got[i].Label != ActionLabel(kind) {
			t.Errorf("Label %d = %q, want %q", i, got[i].Label, ActionLabel(kind))
		}
		if got[i].Action.Project != "shop" || got[i].Action.Service != "web" {
			t.Errorf("Action %d targets %q/%q, want %q/%q", i, got[i].Action.Project, got[i].Action.Service, "shop", "web")
		}
	}
}

func TestServiceActions_EnablesTheActionsOnlyWhenTheProjectIsFree(t *testing.T) {
	busy := &compose.ActivityRef{ID: "act-1", Kind: compose.ActionUp}
	for _, tc := range []struct {
		name            string
		project         compose.Project
		dockerReachable bool
		want            bool
	}{
		{"docker answers and no Activity", compose.Project{}, true, true},
		{"an Activity runs", compose.Project{Activity: busy}, true, false},
		{"docker does not answer", compose.Project{}, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, view := range ServiceActions(tc.project, compose.Service{Name: "web"}, tc.dockerReachable) {
				if view.Enabled != tc.want {
					t.Errorf("%s Enabled = %v, want %v", view.Action.Kind, view.Enabled, tc.want)
				}
			}
		})
	}
}

func TestServiceActions_ConfirmsOnlyDown(t *testing.T) {
	var down bool
	for _, view := range ServiceActions(compose.Project{Name: "shop"}, compose.Service{Name: "web"}, true) {
		if view.Action.Kind != compose.ActionDownService {
			if view.Confirmation != nil {
				t.Errorf("%s asks for a confirmation, want none", view.Action.Kind)
			}
			continue
		}
		down = true
		assertConfirmation(t, view, Confirmation{
			Title:        "Down di web",
			Message:      "Ferma e rimuove i Container del Servizio. I volumi restano.",
			ConfirmLabel: "Down",
		})
	}
	if !down {
		t.Error("ServiceActions offers no Down")
	}
}

func TestDockerNotice_SaysNothingBeforeTheFirstSnapshot(t *testing.T) {
	for _, tc := range []struct {
		name      string
		loading   bool
		reachable bool
		want      string
	}{
		{"first snapshot pending", true, false, ""},
		{"docker answers", false, true, ""},
		{"docker does not answer", false, false, DockerUnreachableText},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := DockerNotice(tc.loading, tc.reachable); got != tc.want {
				t.Errorf("DockerNotice(%v, %v) = %q, want %q", tc.loading, tc.reachable, got, tc.want)
			}
		})
	}
}

func TestActionDisabledReason_ExplainsWhyTheButtonIsOff(t *testing.T) {
	for _, tc := range []struct {
		name            string
		dockerReachable bool
		busy            bool
		want            string
	}{
		{"available", true, false, ""},
		{"docker does not answer", false, false, DockerUnreachableText},
		{"an Activity runs", true, true, ActivityRunningText},
		{"an Activity runs and docker does not answer", false, true, DockerUnreachableText},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ActionDisabledReason(tc.dockerReachable, tc.busy); got != tc.want {
				t.Errorf("ActionDisabledReason(%v, %v) = %q, want %q", tc.dockerReachable, tc.busy, got, tc.want)
			}
		})
	}
}

func TestActionReasonText_SaysNothingBeforeTheFirstSnapshot(t *testing.T) {
	for _, tc := range []struct {
		name      string
		loading   bool
		reachable bool
		busy      bool
		want      string
	}{
		{"first snapshot pending", true, false, false, ""},
		{"first snapshot pending while an Activity runs", true, false, true, ""},
		{"docker answers and the Project is free", false, true, false, ""},
		{"docker does not answer", false, false, false, DockerUnreachableText},
		{"an Activity runs", false, true, true, ActivityRunningText},
		{"an Activity runs and docker does not answer", false, false, true, DockerUnreachableText},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ActionReasonText(tc.loading, tc.reachable, tc.busy); got != tc.want {
				t.Errorf("ActionReasonText(%v, %v, %v) = %q, want %q",
					tc.loading, tc.reachable, tc.busy, got, tc.want)
			}
		})
	}
}
