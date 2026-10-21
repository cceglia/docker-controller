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
	// ComposeMissing (File compose mancante) is true when at least one of the
	// compose files no longer exists on disk.
	ComposeMissing bool   `json:"composeMissing"`
	Status         Status `json:"status"`
	// Running and Total count the Project's Containers.
	Running int `json:"running"`
	Total   int `json:"total"`
	// Services groups the Project's Containers by Service, sorted by Service
	// name. It is empty (not nil) when the Project has no Container.
	Services []Service `json:"services"`
	// Activity is the Activity running on the Project, nil when there is none.
	Activity *ActivityRef `json:"activity"`
}

// Service is a `services:` entry of a Project with the Containers it has; a
// Service with replicas has more than one.
type Service struct {
	Name       string      `json:"name"`
	Containers []Container `json:"containers"`
}

// Container is an instance, running or not, of a Service.
type Container struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	// State is the Docker state ("running", "exited", ...).
	State string `json:"state"`
	// Health is the healthcheck result ("healthy", "unhealthy", "starting"),
	// "" when the Container has no healthcheck.
	Health string `json:"health"`
	// Ports are the published ports only, each listed once (IPv4 and IPv6
	// bindings of the same port are merged); empty (not nil) when none.
	Ports []Port `json:"ports"`
}

// Port is a port a Container publishes on the host.
type Port struct {
	// Published is the host port; Target the port inside the Container.
	Published int    `json:"published"`
	Target    int    `json:"target"`
	Protocol  string `json:"protocol"`
}

// Snapshot is the complete, resolved view of every saved Project.
type Snapshot struct {
	// DockerReachable is false when the Docker daemon does not answer.
	DockerReachable bool      `json:"dockerReachable"`
	Projects        []Project `json:"projects"`
}
