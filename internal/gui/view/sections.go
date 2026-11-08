package view

import (
	"sort"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/model"
)

// UngroupedTitle is the heading of the Projects without a Group. The model
// holds the single definition, so form, list and migration agree.
const UngroupedTitle = model.UngroupedTitle

// Section is a heading of the Projects list and the Projects it holds. An
// empty Title means no heading: the list has no Group at all.
type Section struct {
	Title    string
	Projects []compose.Project
}

// Sections groups the Projects by Group. The Groups of the snapshot keep the
// order the user chose, empty Groups produce no heading, and a Group the
// snapshot does not list goes after the known ones, in alphabetical order.
// The Projects without a Group come last under UngroupedTitle. Without any
// Project with a Group the list stays flat, in one unnamed section.
func Sections(projects []compose.Project, groups []string) []Section {
	byGroup := map[string][]compose.Project{}
	var ungrouped []compose.Project
	for _, p := range projects {
		if p.Group == "" {
			ungrouped = append(ungrouped, p)
			continue
		}
		byGroup[p.Group] = append(byGroup[p.Group], p)
	}
	if len(byGroup) == 0 {
		return []Section{{Projects: ungrouped}}
	}

	known := make(map[string]bool, len(groups))
	sections := make([]Section, 0, len(groups)+1)
	for _, title := range groups {
		known[title] = true
		if len(byGroup[title]) > 0 {
			sections = append(sections, Section{Title: title, Projects: byGroup[title]})
		}
	}

	var unknown []string
	for title := range byGroup {
		if !known[title] {
			unknown = append(unknown, title)
		}
	}
	sort.Slice(unknown, func(i, j int) bool { return model.LessGroupName(unknown[i], unknown[j]) })
	for _, title := range unknown {
		sections = append(sections, Section{Title: title, Projects: byGroup[title]})
	}

	if len(ungrouped) > 0 {
		sections = append(sections, Section{Title: UngroupedTitle, Projects: ungrouped})
	}
	return sections
}
