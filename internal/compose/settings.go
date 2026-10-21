package compose

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Settings is a partial update of a Project's settings: a nil field is left as
// it is, an empty (or blank) value clears the setting and restores the default
// behaviour (the compose name is shown, no Group). Profiles, when not nil, is
// the complete set of active Profiles: an empty list deactivates them all.
type Settings struct {
	DisplayName *string  `json:"displayName"`
	Group       *string  `json:"group"`
	Profiles    []string `json:"profiles"`
}

// UpdateSettings changes the settings of a saved Project and publishes the new
// snapshot. It never touches Docker, so it also works while Docker is
// unreachable. Values are trimmed. It is refused with ErrUnknownProject.
func (m *Manager) UpdateSettings(ctx context.Context, name string, s Settings) error {
	s = Settings{DisplayName: trimmed(s.DisplayName), Group: trimmed(s.Group), Profiles: cleanProfiles(s.Profiles)}

	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()

	found, err := m.store.updateSettings(ctx, name, s)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("%w: %q", ErrUnknownProject, name)
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
			next.Projects[i].Group = *s.Group
		}
		if s.Profiles != nil {
			next.Projects[i].ActiveProfiles = s.Profiles
		}
	}
	m.publish(next)
	return nil
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

// groupsOf lists the Groups in use, each once, sorted; never nil.
func groupsOf(projects []Project) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, p := range projects {
		if p.Group != "" && !seen[p.Group] {
			seen[p.Group] = true
			out = append(out, p.Group)
		}
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i]) < strings.ToLower(out[j]) })
	return out
}
