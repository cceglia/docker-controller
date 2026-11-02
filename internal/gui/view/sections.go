package view

import (
	"slices"

	"dockercontroller/internal/compose"
)

// UngroupedTitle is the heading of the Projects without a Group.
const UngroupedTitle = "Senza Gruppo"

// Section is a heading of the Projects list and the Projects it holds. An
// empty Title means no heading: the list has no Group at all.
type Section struct {
	Title    string
	Projects []compose.Project
}

// Sections groups the Projects by Group, with the Groups sorted and the
// Projects without a Group last. Without any Group the list stays flat, in one
// unnamed section.
func Sections(projects []compose.Project) []Section {
	byGroup := map[string][]compose.Project{}
	var ungrouped []compose.Project
	for _, p := range projects {
		if p.Group == "" {
			ungrouped = append(ungrouped, p)
			continue
		}
		byGroup[p.Group] = append(byGroup[p.Group], p)
	}

	titles := make([]string, 0, len(byGroup))
	for title := range byGroup {
		titles = append(titles, title)
	}
	slices.Sort(titles)

	sections := make([]Section, 0, len(titles)+1)
	for _, title := range titles {
		sections = append(sections, Section{Title: title, Projects: byGroup[title]})
	}
	if len(sections) == 0 {
		return []Section{{Projects: ungrouped}}
	}
	if len(ungrouped) > 0 {
		sections = append(sections, Section{Title: UngroupedTitle, Projects: ungrouped})
	}
	return sections
}
