package gui

import (
	"testing"

	"dockercontroller/internal/compose"
)

func TestToneOfContainer_ReadsTheStateAndTheHealth(t *testing.T) {
	for _, tc := range []struct {
		name      string
		container compose.Container
		glyph     string
		cssClass  string
	}{
		{"running", compose.Container{State: "running"}, "●", "success"},
		{"running and healthy", compose.Container{State: "running", Health: "healthy"}, "●", "success"},
		{"running but unhealthy", compose.Container{State: "running", Health: "unhealthy"}, "●", "error"},
		{"dead", compose.Container{State: "dead"}, "●", "error"},
		{"starting healthcheck", compose.Container{State: "running", Health: "starting"}, "◐", "warning"},
		{"restarting", compose.Container{State: "restarting"}, "◐", "warning"},
		{"paused", compose.Container{State: "paused"}, "◐", "warning"},
		{"exited", compose.Container{State: "exited"}, "○", "dim-label"},
		{"created", compose.Container{State: "created"}, "○", "dim-label"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ToneOfContainer(tc.container)
			if got.Glyph != tc.glyph {
				t.Errorf("Glyph = %q, want %q", got.Glyph, tc.glyph)
			}
			if got.CSSClass != tc.cssClass {
				t.Errorf("CSSClass = %q, want %q", got.CSSClass, tc.cssClass)
			}
		})
	}
}

func TestHealthOf_MarksOnlyAProblem(t *testing.T) {
	for _, tc := range []struct {
		name     string
		health   string
		text     string
		cssClass string
	}{
		{"unhealthy", "unhealthy", "unhealthy", "error"},
		{"starting", "starting", "starting", "warning"},
		{"healthy needs no mark", "healthy", "healthy", ""},
		{"no healthcheck stays empty", "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := HealthOf(tc.health)
			if got.Text != tc.text {
				t.Errorf("Text = %q, want %q", got.Text, tc.text)
			}
			if got.CSSClass != tc.cssClass {
				t.Errorf("CSSClass = %q, want %q", got.CSSClass, tc.cssClass)
			}
		})
	}
}

func TestPortText_ShowsHostPortContainerPortAndProtocol(t *testing.T) {
	for _, tc := range []struct {
		name string
		port compose.Port
		want string
	}{
		{"http", compose.Port{Published: 8080, Target: 80, Protocol: "tcp"}, "8080→80/tcp"},
		{"same on both sides", compose.Port{Published: 5432, Target: 5432, Protocol: "tcp"}, "5432→5432/tcp"},
		{"udp", compose.Port{Published: 53, Target: 53, Protocol: "udp"}, "53→53/udp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := PortText(tc.port); got != tc.want {
				t.Errorf("PortText = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestComposeFilesLine_NamesTheFilesInOrderByBaseName(t *testing.T) {
	for _, tc := range []struct {
		name  string
		files []string
		want  string
	}{
		{"no file", nil, ""},
		{"one file", []string{"/home/me/shop/docker-compose.yml"}, "File compose: docker-compose.yml"},
		{
			"one file and its override",
			[]string{"/home/me/shop/docker-compose.yml", "/home/me/shop/compose.override.yml"},
			"File compose: docker-compose.yml, compose.override.yml",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ComposeFilesLine(tc.files); got != tc.want {
				t.Errorf("ComposeFilesLine = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestActiveProfilesLine_ShowsTheProfilesOnlyWhenThereAreAny(t *testing.T) {
	for _, tc := range []struct {
		name     string
		profiles []string
		want     string
	}{
		{"no active profile", nil, ""},
		{"one profile", []string{"debug"}, "Profili attivi: debug"},
		{"two profiles", []string{"debug", "metrics"}, "Profili attivi: debug, metrics"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ActiveProfilesLine(tc.profiles); got != tc.want {
				t.Errorf("ActiveProfilesLine = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestContainerCountLine_CountsTheRunningContainers(t *testing.T) {
	for _, tc := range []struct {
		name    string
		project compose.Project
		want    string
	}{
		{"some running", compose.Project{Running: 3, Total: 5}, "3/5 in esecuzione"},
		{"all running", compose.Project{Running: 2, Total: 2}, "2/2 in esecuzione"},
		{"removed", compose.Project{Running: 0, Total: 0}, "Nessun Container"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ContainerCountLine(tc.project); got != tc.want {
				t.Errorf("ContainerCountLine = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNoServicesText_ExplainsWhyTheProjectHasNoService(t *testing.T) {
	for _, tc := range []struct {
		name    string
		project compose.Project
		want    string
	}{
		{
			"removed",
			compose.Project{Status: compose.StatusRemoved},
			"Nessun Servizio: il Progetto è Rimosso. Avvialo con Up per creare i Container.",
		},
		{
			"missing compose files",
			compose.Project{Status: compose.StatusRemoved, ComposeMissing: true},
			"Nessun Servizio: i file compose del Progetto non esistono più.",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := NoServicesText(tc.project); got != tc.want {
				t.Errorf("NoServicesText = %q, want %q", got, tc.want)
			}
		})
	}
}
