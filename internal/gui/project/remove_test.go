package project

import (
	"testing"

	"dockercontroller/internal/compose"
)

func TestCanRemoveFromList_NeedsARemovedProjectWithoutContainers(t *testing.T) {
	activity := &compose.ActivityRef{}
	for _, tc := range []struct {
		name      string
		project   compose.Project
		reachable bool
		want      bool
	}{
		{"a Rimosso Project and Docker answers", compose.Project{Status: compose.StatusRemoved}, true, true},
		{"a Rimosso Project and Docker is down", compose.Project{Status: compose.StatusRemoved}, false, false},
		{"a Project with Containers", compose.Project{Status: compose.StatusRemoved, Total: 2}, true, false},
		{"a running Project", compose.Project{Status: compose.StatusActive, Running: 2, Total: 2}, true, false},
		{"a stopped Project", compose.Project{Status: compose.StatusStopped, Total: 2}, true, false},
		{"a Project with an Activity", compose.Project{Status: compose.StatusRemoved, Activity: activity}, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := CanRemoveFromList(tc.project, tc.reachable); got != tc.want {
				t.Errorf("CanRemoveFromList = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestRemoveDisabledReason_ExplainsTheRefusal(t *testing.T) {
	activity := &compose.ActivityRef{}
	for _, tc := range []struct {
		name      string
		project   compose.Project
		reachable bool
		want      string
	}{
		{"the entry is available", compose.Project{Status: compose.StatusRemoved}, true, ""},
		{"Docker does not answer", compose.Project{Status: compose.StatusRemoved}, false, DockerUnreachableText},
		{"an Activity runs", compose.Project{Status: compose.StatusRemoved, Activity: activity}, true, ActivityRunningText},
		{"the Project has Containers", compose.Project{Status: compose.StatusRemoved, Total: 3}, true, "Il Progetto ha Container: fai Down prima di toglierlo dalla lista."},
		{"the Project is not Rimosso", compose.Project{Status: compose.StatusStopped}, true, "Si può togliere dalla lista solo un Progetto Rimosso, senza Container."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := RemoveDisabledReason(tc.project, tc.reachable); got != tc.want {
				t.Errorf("RemoveDisabledReason = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRemoveConfirmation_AsksWithTheOldText(t *testing.T) {
	confirmation := RemoveConfirmation(compose.Project{Name: "shop"})
	if confirmation.Title != "Rimuovi shop dalla lista" {
		t.Errorf("Title = %q", confirmation.Title)
	}
	if confirmation.Message != "Il Progetto sparisce dalla lista dell'app. Docker e i file non vengono toccati." {
		t.Errorf("Message = %q", confirmation.Message)
	}
	if confirmation.ConfirmLabel != "Rimuovi dalla lista" {
		t.Errorf("ConfirmLabel = %q", confirmation.ConfirmLabel)
	}
	if !confirmation.Destructive {
		t.Error("Destructive = false, want true")
	}
	if confirmation.TypeToConfirm != "" {
		t.Errorf("TypeToConfirm = %q, want no text to type", confirmation.TypeToConfirm)
	}
}

func TestRemoveConfirmation_UsesTheDisplayName(t *testing.T) {
	confirmation := RemoveConfirmation(compose.Project{Name: "shop", DisplayName: "Negozio"})
	if confirmation.Title != "Rimuovi Negozio dalla lista" {
		t.Errorf("Title = %q", confirmation.Title)
	}
}
