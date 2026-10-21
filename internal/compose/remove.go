package compose

import (
	"context"
	"errors"
	"fmt"
	"slices"
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
// never touches Docker beyond re-reading its state: the Containers of the
// Project are asked to Docker right before deleting, so a stale snapshot cannot
// let a Project with Containers through. For the whole check and delete the
// Project is reserved, so no Activity can start on it in between. It is
// refused with ErrUnknownProject, ErrActivityRunning, ErrDockerUnreachable (the
// Project cannot be known to have no Containers) or ErrProjectNotRemoved.
func (m *Manager) RemoveFromList(ctx context.Context, name string) error {
	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()

	if err := m.reserveForRemoval(name); err != nil {
		return err
	}
	defer func() {
		m.mu.Lock()
		delete(m.removing, name)
		m.mu.Unlock()
	}()

	out, err := m.docker.Run(ctx, "compose", "-p", name, "ps", "-a", "--format", "json")
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("%w: cannot tell whether %s has Containers: %v", ErrDockerUnreachable, name, err)
	}
	containers, err := parsePs(out)
	if err != nil {
		return err
	}
	if len(containers) > 0 {
		return fmt.Errorf("%w: %s has %d Containers", ErrProjectNotRemoved, name, len(containers))
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

// reserveForRemoval checks that name is a saved Project with no running
// Activity, with Docker reachable, and marks it as being removed.
func (m *Manager) reserveForRemoval(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	known := slices.ContainsFunc(m.base.Projects, func(p Project) bool { return p.Name == name })
	_, busy := m.running[name]
	switch {
	case !known:
		return fmt.Errorf("%w: %q", ErrUnknownProject, name)
	case busy || m.removing[name]:
		return fmt.Errorf("%w: %s", ErrActivityRunning, name)
	case !m.base.DockerReachable:
		return fmt.Errorf("%w: cannot tell whether %s has Containers", ErrDockerUnreachable, name)
	}
	m.removing[name] = true
	return nil
}
