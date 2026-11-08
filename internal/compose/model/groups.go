package model

import "strings"

// The two labels the Group field and the Projects list share. Both are
// reserved: neither can be the name of a saved Group. They are defined here so
// the form, the list and the migration read one spelling.
const (
	// NoGroupText is the hint of an empty Group field.
	NoGroupText = "Nessun Gruppo"
	// UngroupedTitle is the heading of the Projects without a Group.
	UngroupedTitle = "Senza Gruppo"
)

// NormalizeGroupName returns the name of a Group as it is saved: without the
// spaces at its ends. Inner spaces stay, because they are part of the name.
func NormalizeGroupName(name string) string {
	return strings.TrimSpace(name)
}

// SameGroupName reports whether two names identify the same Group. The
// comparison ignores the difference between upper and lower case with the
// simple folding of strings.EqualFold, which knows Unicode: "Città" and "CITTÀ"
// are one Group, "Straße" and "STRASSE" are two.
func SameGroupName(a, b string) bool {
	return strings.EqualFold(NormalizeGroupName(a), NormalizeGroupName(b))
}

// LessGroupName reports whether a comes before b in the alphabetical order of
// the Groups: increasing, without regard to case, then by the exact spelling.
func LessGroupName(a, b string) bool {
	left, right := strings.ToLower(a), strings.ToLower(b)
	if left != right {
		return left < right
	}
	return a < b
}

// GroupNameMode selects the rule a validation applies: the form assigns a
// Group, the dialog creates one, or the dialog renames one.
type GroupNameMode int

const (
	// GroupNameAssign is the rule of the Project form: an empty name clears the
	// Group, and a name that already exists is the Group to assign.
	GroupNameAssign GroupNameMode = iota
	// GroupNameCreate is the rule of a new Group: the name cannot be empty and
	// a duplicate is refused.
	GroupNameCreate
	// GroupNameRename is the rule of a rename: the name cannot be empty, a
	// duplicate of the Group being renamed is valid, and another duplicate
	// asks for a union.
	GroupNameRename
)

// GroupNameStatus is the outcome of a validation.
type GroupNameStatus int

const (
	// GroupNameValid means the name can be used as it is.
	GroupNameValid GroupNameStatus = iota
	// GroupNameInvalid means the name is empty where it cannot be, or it is
	// one of the reserved labels.
	GroupNameInvalid
	// GroupNameDuplicate means a Group already has the name.
	GroupNameDuplicate
)

// GroupNameCheck is the result of ValidateGroupName. A duplicate carries the
// saved spelling of the Group, so the caller can ask for a union or assign it.
type GroupNameCheck struct {
	// Name is the name after normalisation.
	Name string
	// Status is the outcome of the validation.
	Status GroupNameStatus
	// Existing is the saved spelling of the Group that already has the name; it
	// is set only when Status is GroupNameDuplicate.
	Existing string
}

// ValidateGroupName applies the single rule of the Group names. The name is
// trimmed first. An empty name is valid only when the form assigns a Group,
// where it clears the Group. A reserved label is never valid. When another
// Group already has the name, the result is a duplicate with its spelling; a
// rename of the same Group, which only changes the spelling, stays valid.
func ValidateGroupName(name string, existing []string, mode GroupNameMode, from string) GroupNameCheck {
	check := GroupNameCheck{Name: NormalizeGroupName(name)}
	if check.Name == "" {
		if mode == GroupNameAssign {
			check.Status = GroupNameValid
			return check
		}
		check.Status = GroupNameInvalid
		return check
	}
	if SameGroupName(check.Name, NoGroupText) || SameGroupName(check.Name, UngroupedTitle) {
		check.Status = GroupNameInvalid
		return check
	}
	for _, saved := range existing {
		if !SameGroupName(saved, check.Name) {
			continue
		}
		if mode == GroupNameRename && SameGroupName(saved, from) {
			check.Status = GroupNameValid
			return check
		}
		check.Status = GroupNameDuplicate
		check.Existing = saved
		return check
	}
	check.Status = GroupNameValid
	return check
}
