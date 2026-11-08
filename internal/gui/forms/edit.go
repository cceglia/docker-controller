package forms

import (
	"sort"
	"strings"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/model"
)

// NoGroupText is the hint of the empty Group field. The model holds the single
// definition, so form, list and migration agree.
const NoGroupText = model.NoGroupText

// ProjectEdit is the pending form of a Project: the values the user types and
// changes while the detail stays open. Nothing reaches the Manager until the
// form is confirmed, so Annulla has nothing to undo. Profiles holds the
// complete set of active Profiles, so a Profile the compose files no longer
// define stays as it was, like the old view kept it.
type ProjectEdit struct {
	// Project is the compose name of the Project the form belongs to.
	Project     string
	DisplayName string
	// Group is the Group the form sends; an empty value clears it. A name
	// outside the saved Groups creates the Group when the form is saved.
	Group    string
	Profiles []string
}

// NewProjectEdit opens the form on the saved values of a Project.
func NewProjectEdit(p compose.Project) ProjectEdit {
	return ProjectEdit{
		Project:     p.Name,
		DisplayName: p.DisplayName,
		Group:       p.Group,
		Profiles:    append([]string{}, p.ActiveProfiles...),
	}
}

// SetProfile turns one Profile on or off and keeps the set sorted. The name
// comes trimmed, and a blank name changes nothing: a Profile always has a name.
func (e *ProjectEdit) SetProfile(name string, on bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	kept := make([]string, 0, len(e.Profiles)+1)
	for _, p := range e.Profiles {
		if p != name {
			kept = append(kept, p)
		}
	}
	if on {
		kept = append(kept, name)
	}
	sort.Strings(kept)
	e.Profiles = kept
}

// ProfileOn says whether the form has a Profile on.
func (e ProjectEdit) ProfileOn(name string) bool {
	return HasProfile(e.Profiles, name)
}

// HasProfile says whether a set of Profiles holds a Profile.
func HasProfile(profiles []string, name string) bool {
	for _, p := range profiles {
		if p == name {
			return true
		}
	}
	return false
}

// GroupValue is the Group the form sends, trimmed. An empty value clears the
// Group of the Project.
func (e ProjectEdit) GroupValue() string {
	return model.NormalizeGroupName(e.Group)
}

// Settings is the update the form asks for: the trimmed Display name, the
// trimmed Group and the complete set of Profiles. A blank name clears the
// setting, so the compose name comes back.
func (e ProjectEdit) Settings() compose.Settings {
	displayName := strings.TrimSpace(e.DisplayName)
	group := e.GroupValue()
	return compose.Settings{
		DisplayName: &displayName,
		Group:       &group,
		Profiles:    cleanEditProfiles(e.Profiles),
	}
}

// cleanEditProfiles trims the names, drops the blanks and the duplicates and
// sorts them; it never returns nil, so the complete set is always sent.
func cleanEditProfiles(in []string) []string {
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

// GroupChoices lists the options of the Group field: no Group first, then the
// saved Groups in the order of the snapshot. A saved Group appears once. The
// current Group of the form joins them at the end, in alphabetical order after
// the known ones, when it matches no saved Group, so the field never loses a
// value the user saved before. An empty choice is no Group.
func GroupChoices(current string, groups []string) []string {
	out := []string{""}
	seen := map[string]bool{"": true}
	for _, g := range groups {
		g = model.NormalizeGroupName(g)
		if g == "" || seen[g] {
			continue
		}
		seen[g] = true
		out = append(out, g)
	}
	current = model.NormalizeGroupName(current)
	if current == "" {
		return out
	}
	for _, g := range out {
		if model.SameGroupName(g, current) {
			return out
		}
	}
	return append(out, current)
}

// GroupSuggestions lists the options of the editable box of the Group: the
// Groups of the choices, without the empty option, which the box shows as its
// hint instead of as a value. It never returns nil.
func GroupSuggestions(choices []string) []string {
	out := []string{}
	for _, choice := range choices {
		if choice != "" {
			out = append(out, choice)
		}
	}
	return out
}

// ProfileChoices lists the Profiles the form offers: the ones the compose files
// of the Project define, in their order, without blanks and duplicates. It
// never returns nil.
func ProfileChoices(p compose.Project) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, profile := range p.Profiles {
		if profile = strings.TrimSpace(profile); profile != "" && !seen[profile] {
			seen[profile] = true
			out = append(out, profile)
		}
	}
	return out
}
