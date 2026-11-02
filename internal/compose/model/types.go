package model

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
	// Profiles are the Profiles the compose files define, sorted; empty (not
	// nil) when there are none.
	Profiles []string `json:"profiles"`
	// ActiveProfiles are the Profiles the user chose, sorted; empty (not nil)
	// when there are none. They may include Profiles no longer defined.
	ActiveProfiles []string `json:"activeProfiles"`
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
	// Loading is true until the first Refresh completes: nothing has been asked
	// to Docker yet, so DockerReachable and Projects mean nothing and must not
	// be shown as "Docker offline".
	Loading bool `json:"loading"`
	// DockerReachable is false when the Docker daemon does not answer.
	DockerReachable bool      `json:"dockerReachable"`
	Projects        []Project `json:"projects"`
	// Groups lists the Groups in use, each once, sorted; empty (not nil) when
	// no Project has a Group.
	Groups []string `json:"groups"`
}

// ActionKind is what an Action does.
type ActionKind string

const (
	ActionUp ActionKind = "up"
	// ActionStart starts existing Containers; it is meant for a Service.
	ActionStart   ActionKind = "start"
	ActionStop    ActionKind = "stop"
	ActionRestart ActionKind = "restart"
	ActionPull    ActionKind = "pull"
	ActionBuild   ActionKind = "build"
	// ActionDown removes the Containers of the whole Project (Project only).
	ActionDown ActionKind = "down"
	// ActionDownVolumes is Down that also deletes the Project's volumes, and so
	// its data (Project only). It is the only kind that ever passes -v.
	ActionDownVolumes ActionKind = "down-volumes"
	// ActionDownService stops and removes the Containers of one Service, keeping
	// volumes (Service only): `compose rm -s -f`, never -v.
	ActionDownService ActionKind = "down-service"
)

// ActivityID identifies one Activity (one execution of an Action).
type ActivityID string

// ActivityRef is the running Activity of a Project, as shown in the snapshot.
type ActivityRef struct {
	ID ActivityID `json:"id"`
	// Service is the Service the Activity targets, "" for the whole Project.
	Service string     `json:"service"`
	Kind    ActionKind `json:"kind"`
}

// Settings is a partial update of a Project's settings: a nil field is left as
// it is, an empty (or blank) value clears the setting and restores the default
// behaviour (the compose name is shown, no Group). Profiles, when not nil, is
// the complete set of active Profiles: an empty list deactivates them all.
type Settings struct {
	DisplayName *string  `json:"displayName"`
	Group       *string  `json:"group"`
	Profiles    []string `json:"profiles"`
}

// Discovered is a Project as reported by `docker compose ls -a`.
type Discovered struct {
	Name         string
	ComposeFiles []string
}

// LogRequest asks for the log stream of a Project, or of one of its Services.
type LogRequest struct {
	// Project is the compose name of the Project.
	Project string `json:"project"`
	// Service, when not empty, restricts the stream to that Service; compose
	// is asked for it, the others are not read at all.
	Service string `json:"service"`
}

// LogLine is one line of a log stream, already attributed to its Service.
type LogLine struct {
	// Service is the Service whose Container wrote the line. It is empty for
	// a line that did not come from a Container (compose's own messages) and
	// holds the raw compose prefix when it matches no known Service.
	Service string `json:"service"`
	// Timestamp is the time Docker recorded for the line (RFC 3339, as
	// compose prints it); empty when the line carries none. It is always
	// requested: whether to show it is up to the caller.
	Timestamp string `json:"timestamp"`
	// Text is the line without prefix, timestamp or terminal colour codes.
	Text string `json:"text"`
	// Stderr marks a message of compose itself (for example a Service that
	// does not exist); Container output always arrives as not Stderr.
	Stderr bool `json:"stderr"`
}
