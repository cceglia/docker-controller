package forms

import (
	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/model"
)

// GroupChangeKind is what one successful operation of the Groups dialog did.
type GroupChangeKind int

const (
	// GroupCreated is a new Group: From is empty and To is the saved name.
	GroupCreated GroupChangeKind = iota
	// GroupRenamed is a rename that kept the Group: To is the new name.
	GroupRenamed
	// GroupUnited is a rename onto a saved Group: To is the Group that survived.
	GroupUnited
	// GroupDeleted removed a Group: To is empty.
	GroupDeleted
)

// GroupChange is one successful operation of the Groups dialog. The form reads
// the records in order and keeps its Group field in step.
type GroupChange struct {
	Kind GroupChangeKind
	From string
	To   string
}

// ApplyGroupChanges returns the Group the form should show after the records of
// the dialog. The value is compared after normalisation and without regard to
// case: a Group the dialog touched follows its record, while another Group and
// an unsaved name stay as they are. Display name and Profiles are not part of a
// record, so they keep what the user typed.
func ApplyGroupChanges(current string, changes []GroupChange) string {
	for _, change := range changes {
		name := model.NormalizeGroupName(current)
		if name == "" {
			continue
		}
		switch change.Kind {
		case GroupCreated:
			if model.SameGroupName(name, change.To) {
				current = change.To
			}
		case GroupRenamed, GroupUnited:
			if model.SameGroupName(name, change.From) || model.SameGroupName(name, change.To) {
				current = change.To
			}
		case GroupDeleted:
			if model.SameGroupName(name, change.From) {
				current = ""
			}
		}
	}
	return current
}

// GroupRow is one Group of the dialog: its name and how many Projects it holds.
type GroupRow struct {
	Name  string
	Count int
}

// GroupRows lists the Groups in the snapshot order with the number of Projects
// each one holds. The comparison ignores case, like the name rule. A Project
// whose Group is not in the list counts nowhere, because the dialog only shows
// the saved Groups.
func GroupRows(groups []string, projects []compose.Project) []GroupRow {
	rows := make([]GroupRow, 0, len(groups))
	for _, group := range groups {
		row := GroupRow{Name: group}
		for _, project := range projects {
			if project.Group != "" && model.SameGroupName(project.Group, group) {
				row.Count++
			}
		}
		rows = append(rows, row)
	}
	return rows
}

// SelectionAfterDeletion returns the row to select after the Group at index
// removed leaves a list that held length Groups: the Group that takes the empty
// place, otherwise the last one, otherwise none (-1).
func SelectionAfterDeletion(removed, length int) int {
	if removed < 0 || length <= 1 {
		return -1
	}
	if removed < length-1 {
		return removed
	}
	return length - 2
}
