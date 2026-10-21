package compose

import (
	"context"
	"errors"
	"fmt"
)

var (
	// ErrProjectNotRemoved is returned by RemoveFromList for a Project that is
	// not Rimosso or still has Containers.
	ErrProjectNotRemoved = errors.New("compose: only a Rimosso Project can be removed from the list")
	// ErrDockerUnreachable is returned when the answer depends on what Docker
	// reports and Docker does not answer.
	ErrDockerUnreachable = errors.New("compose: Docker is not reachable")
)

// RemoveFromList deletes a saved Project (and its settings) from the app. It
// never touches Docker beyond re-reading its state. It is refused with
// ErrUnknownProject, ErrActivityRunning, ErrDockerUnreachable (the Project
// cannot be known to have no Containers) or ErrProjectNotRemoved.
func (m *Manager) RemoveFromList(ctx context.Context, name string) error {
	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()

	m.mu.Lock()
	var project *Project
	for i := range m.base.Projects {
		if m.base.Projects[i].Name == name {
			p := m.base.Projects[i]
			project = &p
			break
		}
	}
	reachable := m.base.DockerReachable
	_, busy := m.running[name]
	m.mu.Unlock()

	switch {
	case project == nil:
		return fmt.Errorf("%w: %q", ErrUnknownProject, name)
	case busy:
		return fmt.Errorf("%w: %s", ErrActivityRunning, name)
	case !reachable:
		return fmt.Errorf("%w: cannot tell whether %s has Containers", ErrDockerUnreachable, name)
	case project.Status != StatusRemoved || project.Total != 0:
		return fmt.Errorf("%w: %s is %s with %d Containers", ErrProjectNotRemoved, name, project.Status, project.Total)
	}

	if err := m.store.remove(ctx, name); err != nil {
		return err
	}
	next, err := m.build(ctx)
	if err != nil {
		return err
	}
	m.publish(next)
	return nil
}
