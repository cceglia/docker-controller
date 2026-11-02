package gui

import (
	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/view"
)

// removeText is the label of the entry that takes a Project off the list.
const removeText = "Rimuovi dalla lista"

// removeTooltip says what the entry does while it can be pressed: it leaves
// Docker and the files of the Project alone.
const removeTooltip = "Togli il Progetto dalla lista: Docker e i file non vengono toccati."

// ContainerRefusalText explains the entry that cannot be pressed because the
// Project still has Containers.
const ContainerRefusalText = "Il Progetto ha Container: fai Down prima di toglierlo dalla lista."

// NotRemovedRefusalText explains the entry that cannot be pressed because the
// Project is not Rimosso.
const NotRemovedRefusalText = "Si può togliere dalla lista solo un Progetto Rimosso, senza Container."

// CanRemoveFromList reports whether a Project may leave the list: only a
// Rimosso Project with no Container, and only while Docker answers and no
// Activity runs, because the Manager asks Docker again before it deletes.
func CanRemoveFromList(project compose.Project, dockerReachable bool) bool {
	return dockerReachable && project.Activity == nil &&
		project.Status == compose.StatusRemoved && project.Total == 0
}

// RemoveDisabledReason explains an entry that cannot be pressed. It is "" while
// the entry is available, so a button with no message is the one that works.
func RemoveDisabledReason(project compose.Project, dockerReachable bool) string {
	switch {
	case !dockerReachable:
		return DockerUnreachableText
	case project.Activity != nil:
		return ActivityRunningText
	case project.Total > 0:
		return ContainerRefusalText
	case project.Status != compose.StatusRemoved:
		return NotRemovedRefusalText
	}
	return ""
}

// RemoveConfirmation is the question asked before a Project leaves the list.
// The texts are the ones of the old view, and the answer deletes nothing in
// Docker.
func RemoveConfirmation(project compose.Project) Confirmation {
	return Confirmation{
		Title:        "Rimuovi " + view.ProjectName(project) + " dalla lista",
		Message:      "Il Progetto sparisce dalla lista dell'app. Docker e i file non vengono toccati.",
		ConfirmLabel: removeText,
		Destructive:  true,
	}
}
