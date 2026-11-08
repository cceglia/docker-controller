package compose

import (
	"context"

	"dockercontroller/internal/compose/model"
)

// The Group errors and the duplicate report live in model; the root re-exports
// them for gui, tray and main.
var (
	ErrUnknownGroup     = model.ErrUnknownGroup
	ErrDuplicateGroup   = model.ErrDuplicateGroup
	ErrInvalidGroupName = model.ErrInvalidGroupName
)

// DuplicateGroupError reports the saved spelling of the Group that already has
// a name.
type DuplicateGroupError = model.DuplicateGroupError

// CreateGroup saves a new Group at the end of the list and publishes the new
// snapshot at once. It never talks to Docker, so it also works while Docker is
// unreachable. A name that is already saved, even with different capitalisation,
// is refused with model.DuplicateGroupError, which carries the saved spelling; a
// blank or reserved name with model.ErrInvalidGroupName. The store applies the
// single rule of the model.
func (m *Manager) CreateGroup(ctx context.Context, name string) error {
	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()

	if err := m.store.CreateGroup(ctx, name); err != nil {
		return err
	}
	next, err := m.snapshotWithGroups(ctx)
	if err != nil {
		return err
	}
	m.publish(next)
	return nil
}

// RenameGroup renames a saved Group and publishes the new snapshot at once. The
// new name arrives on every Project of the Group. A name that another Group
// already has unites the two: the Projects move to the Group that already had
// the name, which keeps its spelling, and the source leaves the list. A blank or
// reserved name is refused with model.ErrInvalidGroupName, an unknown Group with
// model.ErrUnknownGroup.
func (m *Manager) RenameGroup(ctx context.Context, from, to string) error {
	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()

	if err := m.store.RenameGroup(ctx, from, to); err != nil {
		return err
	}
	next, err := m.snapshotWithGroups(ctx)
	if err != nil {
		return err
	}
	// The saved list names the Group that survived: the renamed Group itself, or
	// the one that already had the name in a union.
	surviving := survivingGroup(next.Groups, to)
	for i := range next.Projects {
		if model.SameGroupName(next.Projects[i].Group, from) {
			next.Projects[i].Group = surviving
		}
	}
	m.publish(next)
	return nil
}

// DeleteGroup removes a saved Group and publishes the new snapshot at once. The
// Projects of the Group stay and lose their Group; Docker is not touched. An
// unknown Group is refused with model.ErrUnknownGroup.
func (m *Manager) DeleteGroup(ctx context.Context, name string) error {
	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()

	if err := m.store.DeleteGroup(ctx, name); err != nil {
		return err
	}
	next, err := m.snapshotWithGroups(ctx)
	if err != nil {
		return err
	}
	for i := range next.Projects {
		if model.SameGroupName(next.Projects[i].Group, name) {
			next.Projects[i].Group = ""
		}
	}
	m.publish(next)
	return nil
}

// snapshotWithGroups copies the base snapshot and replaces the Group list with
// the saved one. The caller changes the Projects it touched before publishing.
// The copy keeps the snapshot already handed to the watchers untouched.
func (m *Manager) snapshotWithGroups(ctx context.Context) (Snapshot, error) {
	groups, err := m.store.ListGroups(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	m.mu.Lock()
	next := m.base
	next.Projects = append([]Project(nil), m.base.Projects...)
	m.mu.Unlock()
	next.Groups = groups
	return next, nil
}

// survivingGroup returns the saved spelling of the Group whose name is asked
// for. A list that does not hold it, which the store never writes, falls back to
// the normalised name so a Project never keeps a spelling that no longer exists.
func survivingGroup(groups []string, name string) string {
	for _, saved := range groups {
		if model.SameGroupName(saved, name) {
			return saved
		}
	}
	return model.NormalizeGroupName(name)
}
