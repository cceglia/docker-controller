package forms

import (
	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/model"
)

// The name rule of a Group lives in the model: one function and one set of
// outcomes. The form model re-exports them for the dialog, so the dialog reads
// the rule through the form and never repeats it.
type (
	// GroupNameMode selects the rule a validation applies.
	GroupNameMode = model.GroupNameMode
	// GroupNameStatus is the outcome of a validation.
	GroupNameStatus = model.GroupNameStatus
	// GroupNameCheck is the result of ValidateGroupName.
	GroupNameCheck = model.GroupNameCheck
)

const (
	// GroupNameAssign is the rule of the Project form, GroupNameCreate the rule
	// of a new Group and GroupNameRename the rule of a rename.
	GroupNameAssign = model.GroupNameAssign
	GroupNameCreate = model.GroupNameCreate
	GroupNameRename = model.GroupNameRename
	// GroupNameValid, GroupNameInvalid and GroupNameDuplicate are the outcomes
	// of a validation.
	GroupNameValid     = model.GroupNameValid
	GroupNameInvalid   = model.GroupNameInvalid
	GroupNameDuplicate = model.GroupNameDuplicate
)

// NormalizeGroupName returns the name of a Group as it is saved: without the
// spaces at its ends.
func NormalizeGroupName(name string) string {
	return model.NormalizeGroupName(name)
}

// ValidateGroupName applies the single rule of the Group names. The name is
// trimmed first. A blank name is valid only when the form assigns a Group, and
// a reserved label is never valid. A name another Group has is a duplicate with
// its spelling, while a rename of the same Group stays valid.
func ValidateGroupName(name string, existing []string, mode GroupNameMode, from string) GroupNameCheck {
	return model.ValidateGroupName(name, existing, mode, from)
}

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

// MoveGroup returns the list with the Group at index swapped with the neighbour
// delta places away, and the row of the moved Group. A move past either end
// changes nothing: the list keeps its order and the row stays. The input list
// is not changed, so a caller can keep the snapshot it received.
func MoveGroup(groups []string, index, delta int) ([]string, int) {
	moved := append([]string(nil), groups...)
	if index < 0 || index >= len(moved) {
		return moved, index
	}
	target := index + delta
	if target < 0 || target >= len(moved) {
		return moved, index
	}
	moved[index], moved[target] = moved[target], moved[index]
	return moved, target
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
