package compose

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"dockercontroller/internal/compose/model"
)

// Settings lives in model; the root re-exports it for gui, tray and main.
type Settings = model.Settings

// UpdateSettings changes the settings of a saved Project and publishes the new
// snapshot. It never touches Docker, so it also works while Docker is
// unreachable. Values are trimmed. It is refused with ErrUnknownProject, and a
// reserved Group name with ErrInvalidGroupName.
func (m *Manager) UpdateSettings(ctx context.Context, name string, s Settings) error {
	s = Settings{DisplayName: trimmed(s.DisplayName), Group: cleanGroup(s.Group), Profiles: cleanProfiles(s.Profiles)}

	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()

	found, err := m.store.UpdateSettings(ctx, name, s)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("%w: %q", ErrUnknownProject, name)
	}

	// The Groups are read back so a Group typed in the form arrives with its
	// place and a name that differs only in capitalisation shows the saved
	// spelling. A Group the form left empty does not create anything.
	groups, err := m.store.ListGroups(ctx)
	if err != nil {
		return err
	}

	m.mu.Lock()
	next := m.base
	next.Projects = append([]Project(nil), m.base.Projects...)
	m.mu.Unlock()
	for i := range next.Projects {
		if next.Projects[i].Name != name {
			continue
		}
		if s.DisplayName != nil {
			next.Projects[i].DisplayName = *s.DisplayName
		}
		if s.Group != nil {
			next.Projects[i].Group = savedGroupName(*s.Group, groups)
		}
		if s.Profiles != nil {
			next.Projects[i].ActiveProfiles = s.Profiles
		}
	}
	next.Groups = groups
	m.publish(next)
	return nil
}

// savedGroupName returns the spelling of the Group whose name is asked for. A
// name that matches no saved Group stays as it is, and an empty name is no
// Group.
func savedGroupName(name string, groups []string) string {
	if name == "" {
		return ""
	}
	for _, saved := range groups {
		if model.SameGroupName(saved, name) {
			return saved
		}
	}
	return name
}

// cleanGroup trims the Group name; nil stays nil (unchanged).
func cleanGroup(s *string) *string {
	if s == nil {
		return nil
	}
	t := model.NormalizeGroupName(*s)
	return &t
}

// cleanProfiles trims the names, drops blanks and duplicates and sorts them;
// nil stays nil (unchanged), anything else is never nil.
func cleanProfiles(in []string) []string {
	if in == nil {
		return nil
	}
	seen := map[string]bool{}
	out := []string{}
	for _, p := range in {
		if p = strings.TrimSpace(p); p != "" && !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

func trimmed(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	return &t
}
