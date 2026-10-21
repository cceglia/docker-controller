package compose

// Status is the state of a Project, derived from the state of its Containers.
type Status string

const (
	// StatusActive (Attivo): every Container is running.
	StatusActive Status = "active"
	// StatusPartial (Parziale): some Containers run, others do not.
	StatusPartial Status = "partial"
	// StatusStopped (Fermo): it has Containers but none is running.
	StatusStopped Status = "stopped"
	// StatusRemoved (Rimosso): it has no Container at all.
	StatusRemoved Status = "removed"
)

// Project is a saved Project, already resolved against what Docker reports.
type Project struct {
	// Name is the compose project name; it identifies the Project.
	Name string `json:"name"`
	// DisplayName is the user's optional display name ("" when unset).
	DisplayName string `json:"displayName"`
	// Group is the user's optional group ("" when unset).
	Group string `json:"group"`
	// Folder is the directory of the Project.
	Folder string `json:"folder"`
	// ComposeFiles are the compose files, in -f order.
	ComposeFiles []string `json:"composeFiles"`
	Status       Status   `json:"status"`
	// Running and Total count the Project's Containers.
	Running int `json:"running"`
	Total   int `json:"total"`
	// Activity is the Activity running on the Project, nil when there is none.
	Activity *ActivityRef `json:"activity"`
}

// Snapshot is the complete, resolved view of every saved Project.
type Snapshot struct {
	// DockerReachable is false when the Docker daemon does not answer.
	DockerReachable bool      `json:"dockerReachable"`
	Projects        []Project `json:"projects"`
}
