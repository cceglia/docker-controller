package view

import "dockercontroller/internal/compose"

// TackState is how one Container reads in the strip of a Project.
type TackState int

const (
	// TackStopped is a Container that is not running: the tack stays empty.
	TackStopped TackState = iota
	// TackRunning is a running Container: the tack is filled.
	TackRunning
	// TackFaulty is a Container Docker reports unhealthy or dead: the tack
	// takes the error colour.
	TackFaulty
)

// MaxTacks is how many tacks the strip draws before it counts the rest.
const MaxTacks = 12

// ContainerTack reads the tack of one Container. A fault wins over the state,
// because a running but unhealthy Container is the one worth noticing.
func ContainerTack(c compose.Container) TackState {
	if c.Health == "unhealthy" || c.State == "dead" {
		return TackFaulty
	}
	if c.State == "running" {
		return TackRunning
	}
	return TackStopped
}

// Tacks flattens the Containers of the Services in order and returns at most
// MaxTacks tacks, with how many Containers they leave out.
func Tacks(services []compose.Service) ([]TackState, int) {
	var all []TackState
	for _, s := range services {
		for _, c := range s.Containers {
			all = append(all, ContainerTack(c))
		}
	}
	more := len(all) - MaxTacks
	if more < 0 {
		more = 0
	}
	if len(all) > MaxTacks {
		all = all[:MaxTacks]
	}
	return all, more
}
