package gui

import (
	"testing"

	"dockercontroller/internal/compose"
)

func TestDetailData_SameDrawComparesOnlyWhatThePanelDraws(t *testing.T) {
	busy := &compose.ActivityRef{ID: "act-1", Kind: compose.ActionUp}
	base := compose.Project{
		Name:           "shop",
		DisplayName:    "Negozio",
		Group:          "Casa",
		Folder:         "/srv/shop",
		ComposeFiles:   []string{"/srv/shop/docker-compose.yml"},
		Status:         compose.StatusPartial,
		Running:        1,
		Total:          2,
		Services:       []compose.Service{{Name: "web", Containers: []compose.Container{{Name: "shop-web-1", State: "running"}}}},
		Profiles:       []string{"debug", "metrics"},
		ActiveProfiles: []string{"debug"},
	}
	edited := func(edit func(*compose.Project)) detailData {
		project := base
		edit(&project)
		return detailData{project: project, dockerReachable: true}
	}

	for _, tc := range []struct {
		name  string
		other detailData
		want  bool
	}{
		{"the same Project", edited(func(*compose.Project) {}), true},
		{
			"another Folder only: the detail does not draw it",
			edited(func(p *compose.Project) { p.Folder = "/home/me/shop" }),
			true,
		},
		{
			"another set of defined Profiles only: the detail does not draw it",
			edited(func(p *compose.Project) { p.Profiles = []string{"debug"} }),
			true,
		},
		{
			"another active Profile",
			edited(func(p *compose.Project) { p.ActiveProfiles = []string{"debug", "metrics"} }),
			false,
		},
		{
			"another Compose file",
			edited(func(p *compose.Project) { p.ComposeFiles = []string{"/srv/shop/compose.override.yml"} }),
			false,
		},
		{
			"another Display name",
			edited(func(p *compose.Project) { p.DisplayName = "Bottega" }),
			false,
		},
		{
			"another Group",
			edited(func(p *compose.Project) { p.Group = "Lavoro" }),
			false,
		},
		{
			"another count of running Containers",
			edited(func(p *compose.Project) { p.Running = 2; p.Status = compose.StatusActive }),
			false,
		},
		{
			"the Compose files are gone",
			edited(func(p *compose.Project) { p.ComposeMissing = true }),
			false,
		},
		{
			"another state of a Container",
			edited(func(p *compose.Project) {
				p.Services = []compose.Service{{Name: "web", Containers: []compose.Container{{Name: "shop-web-1", State: "exited"}}}}
			}),
			false,
		},
		{
			"a new published port",
			edited(func(p *compose.Project) {
				p.Services = []compose.Service{{Name: "web", Containers: []compose.Container{{
					Name:  "shop-web-1",
					State: "running",
					Ports: []compose.Port{{Published: 8080, Target: 80, Protocol: "tcp"}},
				}}}}
			}),
			false,
		},
		{
			"a new Activity",
			edited(func(p *compose.Project) { p.Activity = busy }),
			false,
		},
		{
			"Docker stops answering",
			detailData{project: base, dockerReachable: false},
			false,
		},
		{
			"the first snapshot arrives",
			detailData{project: base, dockerReachable: true, loading: true},
			false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			shown := detailData{project: base, dockerReachable: true}
			if got := shown.sameDraw(tc.other); got != tc.want {
				t.Errorf("sameDraw = %v, want %v", got, tc.want)
			}
			// The comparison is symmetric: a panel that would be rebuilt on the
			// next snapshot must also rebuild on the one before.
			if got := tc.other.sameDraw(shown); got != tc.want {
				t.Errorf("sameDraw the other way = %v, want %v", got, tc.want)
			}
		})
	}
}
