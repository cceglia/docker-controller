package gui

import (
	"sort"
	"strings"

	"dockercontroller/internal/compose"
)

// NoGroupText is the option of the Group field that clears the Group.
const NoGroupText = "Nessun Gruppo"

// NewGroupText is the last option of the Group field: it asks for the name of a
// new Group, like the old view did.
const NewGroupText = "Nuovo Gruppo…"

// ProjectEdit is the pending form of a Project: the values the user types and
// changes while the detail stays open. Nothing reaches the Manager until the
// form is confirmed, so Annulla has nothing to undo. Profiles holds the
// complete set of active Profiles, so a Profile the compose files no longer
// define stays as it was, like the old view kept it.
type ProjectEdit struct {
	// Project is the compose name of the Project the form belongs to.
	Project     string
	DisplayName string
	Group       string
	// NewGroup says the Group field asks for a new Group instead of one of the
	// choices. NewGroupName is the name typed there, which a new form resets.
	NewGroup     bool
	NewGroupName string
	Profiles     []string
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

// GroupValue is the Group the form sends: the new Group typed when the Group
// field asks for one, else the chosen Group. Both come out trimmed.
func (e ProjectEdit) GroupValue() string {
	if e.NewGroup {
		return strings.TrimSpace(e.NewGroupName)
	}
	return strings.TrimSpace(e.Group)
}

// SaveEnabled says whether the form may be confirmed. A form that asks for a
// new Group needs its name, so the old view disabled Salva the same way.
func (e ProjectEdit) SaveEnabled() bool {
	return !e.NewGroup || strings.TrimSpace(e.NewGroupName) != ""
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

// DisplayNameOf is the name the detail and the list show: the trimmed Display
// name, or the compose name when the Display name is blank.
func DisplayNameOf(displayName, composeName string) string {
	if name := strings.TrimSpace(displayName); name != "" {
		return name
	}
	return composeName
}

// GroupChoices lists the options of the Group field: no Group first, then every
// Group the Projects in the snapshot use, sorted without duplicates. The
// current Group of the form joins them when nothing else uses it, so the field
// never loses a value the user saved before.
func GroupChoices(current string, projects []compose.Project) []string {
	seen := map[string]bool{"": true}
	var names []string
	for _, p := range projects {
		group := strings.TrimSpace(p.Group)
		if group == "" || seen[group] {
			continue
		}
		seen[group] = true
		names = append(names, group)
	}
	if group := strings.TrimSpace(current); group != "" && !seen[group] {
		seen[group] = true
		names = append(names, group)
	}
	sort.Slice(names, func(i, j int) bool {
		left, right := strings.ToLower(names[i]), strings.ToLower(names[j])
		if left != right {
			return left < right
		}
		return names[i] < names[j]
	})
	return append([]string{""}, names...)
}

// GroupIndex is the option of a Group in the list of choices. A Group the list
// does not hold reads as no Group.
func GroupIndex(choices []string, group string) int {
	group = strings.TrimSpace(group)
	for i, choice := range choices {
		if choice == group {
			return i
		}
	}
	return 0
}

// groupAt is the Group of one option of the list. An option outside the list
// reads as no Group, so a stale selection never names a Group of its own.
func groupAt(choices []string, index int) string {
	if index < 0 || index >= len(choices) {
		return ""
	}
	return choices[index]
}

// GroupOption is the option the Group field shows: the option of the chosen
// Group, or the last option when the form asks for a new Group. The last option
// is the one for a new Group, one past the choices of the list.
func GroupOption(choices []string, e ProjectEdit) int {
	if e.NewGroup {
		return len(choices)
	}
	return GroupIndex(choices, e.Group)
}

// SetGroupOption picks an option of the Group field. The last option turns the
// new Group field on and keeps the name already typed. Any other option turns
// it off and makes its Group the chosen one. An option outside the field reads
// as no Group, like groupAt does.
func (e *ProjectEdit) SetGroupOption(choices []string, index int) {
	if index == len(choices) {
		e.NewGroup = true
		return
	}
	e.NewGroup = false
	e.Group = groupAt(choices, index)
}

// groupLabels is the text of each option of the Group field: no Group reads as
// a sentence, every Group keeps its name, and the last option asks for a new
// Group.
func groupLabels(choices []string) []string {
	labels := make([]string, 0, len(choices)+1)
	for _, choice := range choices {
		if choice == "" {
			labels = append(labels, NoGroupText)
			continue
		}
		labels = append(labels, choice)
	}
	if len(labels) == 0 {
		labels = append(labels, NoGroupText)
	}
	return append(labels, NewGroupText)
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
