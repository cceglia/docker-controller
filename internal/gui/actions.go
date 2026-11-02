package gui

import (
	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/view"
)

// DockerUnreachableText is the notice of a view whose Actions are disabled
// because the Docker daemon does not answer.
const DockerUnreachableText = "Docker non è raggiungibile: le Azioni sono disabilitate."

// ActivityRunningText explains an Action disabled by a running Activity.
const ActivityRunningText = "Un'Attività è in corso su questo Progetto."

// Confirmation is the question asked before a destructive Action runs. The
// texts are the ones of the old view, so the same Action asks the same thing.
type Confirmation struct {
	Title        string
	Message      string
	ConfirmLabel string
	// TypeToConfirm, when not empty, is the text the user must type before the
	// confirmation can run: the name of the Project for Down con volumi.
	TypeToConfirm string
	// Destructive marks the Action that deletes data.
	Destructive bool
}

// ActionView is one Action of the detail: how the button reads, the Action to
// launch and whether the button can be pressed. Confirmation is nil when the
// Action starts at once.
type ActionView struct {
	Label        string
	Action       compose.Action
	Enabled      bool
	Confirmation *Confirmation
}

// ActionLabel is the label of an ActionKind, the one the old view and the tray
// show. An unknown kind keeps its own name, so nothing is hidden.
func ActionLabel(kind compose.ActionKind) string {
	switch kind {
	case compose.ActionUp:
		return "Up"
	case compose.ActionStart:
		return "Start"
	case compose.ActionStop:
		return "Stop"
	case compose.ActionRestart:
		return "Restart"
	case compose.ActionPull:
		return "Pull"
	case compose.ActionBuild:
		return "Build"
	case compose.ActionDown:
		return "Down"
	case compose.ActionDownVolumes:
		return "Down con volumi"
	case compose.ActionDownService:
		return "Down"
	}
	return string(kind)
}

// projectActionKinds are the Actions a Project offers, in the order the detail
// lists them after the one the status suggests.
var projectActionKinds = []compose.ActionKind{
	compose.ActionUp,
	compose.ActionStop,
	compose.ActionRestart,
	compose.ActionPull,
	compose.ActionBuild,
}

// primaryKind is the Action a status suggests: Stop for a running Project,
// Restart for a partial one and Up for everything else.
func primaryKind(status compose.Status) compose.ActionKind {
	switch status {
	case compose.StatusActive:
		return compose.ActionStop
	case compose.StatusPartial:
		return compose.ActionRestart
	}
	return compose.ActionUp
}

// ProjectActions lists the Actions of a Project: the one the status suggests
// first, then the others, then Down and Down con volumi. Every Action is
// disabled while the Project has an Activity or Docker does not answer.
func ProjectActions(project compose.Project, dockerReachable bool) []ActionView {
	enabled := dockerReachable && project.Activity == nil
	primary := primaryKind(project.Status)

	views := make([]ActionView, 0, len(projectActionKinds)+2)
	views = append(views, projectAction(project, primary, enabled))
	for _, kind := range projectActionKinds {
		if kind == primary {
			continue
		}
		views = append(views, projectAction(project, kind, enabled))
	}
	return append(views,
		ActionView{
			Label:   ActionLabel(compose.ActionDown),
			Action:  compose.Action{Project: project.Name, Kind: compose.ActionDown},
			Enabled: enabled,
			Confirmation: &Confirmation{
				Title:        "Down di " + view.ProjectName(project),
				Message:      "Ferma e rimuove i Container del Progetto. I volumi restano. Il Progetto resta in lista come Rimosso.",
				ConfirmLabel: "Down",
			},
		},
		ActionView{
			Label:   ActionLabel(compose.ActionDownVolumes),
			Action:  compose.Action{Project: project.Name, Kind: compose.ActionDownVolumes},
			Enabled: enabled,
			Confirmation: &Confirmation{
				Title:         "Down con volumi di " + view.ProjectName(project),
				Message:       "Rimuove i Container e ELIMINA i volumi del Progetto: i dati nei volumi andranno persi per sempre.",
				ConfirmLabel:  "Down con volumi",
				TypeToConfirm: project.Name,
				Destructive:   true,
			},
		},
	)
}

// projectAction is one Action of the whole Project.
func projectAction(project compose.Project, kind compose.ActionKind, enabled bool) ActionView {
	return ActionView{
		Label:   ActionLabel(kind),
		Action:  compose.Action{Project: project.Name, Kind: kind},
		Enabled: enabled,
	}
}

// serviceActionKinds are the Actions a Service offers before Down.
var serviceActionKinds = []compose.ActionKind{
	compose.ActionStart,
	compose.ActionStop,
	compose.ActionRestart,
}

// ServiceActions lists the Actions of one Service: Start, Stop, Restart and
// Down. Down is the only one that asks for a confirmation. Every Action is
// disabled while the Project has an Activity or Docker does not answer.
func ServiceActions(project compose.Project, service compose.Service, dockerReachable bool) []ActionView {
	enabled := dockerReachable && project.Activity == nil

	views := make([]ActionView, 0, len(serviceActionKinds)+1)
	for _, kind := range serviceActionKinds {
		views = append(views, ActionView{
			Label:   ActionLabel(kind),
			Action:  compose.Action{Project: project.Name, Service: service.Name, Kind: kind},
			Enabled: enabled,
		})
	}
	return append(views, ActionView{
		Label:   ActionLabel(compose.ActionDownService),
		Action:  compose.Action{Project: project.Name, Service: service.Name, Kind: compose.ActionDownService},
		Enabled: enabled,
		Confirmation: &Confirmation{
			Title:        "Down di " + service.Name,
			Message:      "Ferma e rimuove i Container del Servizio. I volumi restano.",
			ConfirmLabel: "Down",
		},
	})
}

// DockerNotice is what the detail says about Docker: nothing before the first
// snapshot, nothing when Docker answers, and the reason the Actions are off
// when it does not.
func DockerNotice(loading, dockerReachable bool) string {
	if loading || dockerReachable {
		return ""
	}
	return DockerUnreachableText
}

// ActionDisabledReason explains an insensitive button: the missing Docker
// first, then the running Activity. It is "" while the Action is available.
func ActionDisabledReason(dockerReachable, busy bool) string {
	if !dockerReachable {
		return DockerUnreachableText
	}
	if busy {
		return ActivityRunningText
	}
	return ""
}

// actionReasonText is what the detail writes beside the row of Actions. Before
// the first snapshot Docker has not been asked yet, so the row claims no reason
// and never says Docker is unreachable. After it, the reason is the one of the
// insensitive button.
func actionReasonText(loading, dockerReachable, busy bool) string {
	if loading {
		return ""
	}
	return ActionDisabledReason(dockerReachable, busy)
}
