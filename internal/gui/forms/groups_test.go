package forms

import (
	"reflect"
	"testing"

	"dockercontroller/internal/compose"
)

func TestValidateGroupName_AppliesTheRuleOfTheModel(t *testing.T) {
	for _, tc := range []struct {
		name     string
		input    string
		existing []string
		mode     GroupNameMode
		from     string
		wantName string
		want     GroupNameStatus
		spelling string
	}{
		{
			name:     "a new name is valid and trimmed",
			input:    "  Casa  ",
			existing: []string{"Lavoro"},
			mode:     GroupNameCreate,
			wantName: "Casa",
			want:     GroupNameValid,
		},
		{
			name:  "an empty creation is invalid",
			input: "   ",
			mode:  GroupNameCreate,
			want:  GroupNameInvalid,
		},
		{
			name:     "a reserved label is invalid",
			input:    "Senza Gruppo",
			existing: []string{"Casa"},
			mode:     GroupNameCreate,
			wantName: "Senza Gruppo",
			want:     GroupNameInvalid,
		},
		{
			name:     "another Group with the name is a duplicate with its spelling",
			input:    "casa",
			existing: []string{"Casa", "Lavoro"},
			mode:     GroupNameCreate,
			wantName: "casa",
			want:     GroupNameDuplicate,
			spelling: "Casa",
		},
		{
			name:     "a rename of its own spelling stays valid",
			input:    "CASA",
			existing: []string{"Casa", "Lavoro"},
			mode:     GroupNameRename,
			from:     "Casa",
			wantName: "CASA",
			want:     GroupNameValid,
		},
		{
			name:     "an empty assignment clears the Group",
			existing: []string{"Casa"},
			mode:     GroupNameAssign,
			want:     GroupNameValid,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			check := ValidateGroupName(tc.input, tc.existing, tc.mode, tc.from)
			if check.Name != tc.wantName {
				t.Errorf("the name is %q, want %q", check.Name, tc.wantName)
			}
			if check.Status != tc.want {
				t.Errorf("the status is %v, want %v", check.Status, tc.want)
			}
			if check.Existing != tc.spelling {
				t.Errorf("the saved spelling is %q, want %q", check.Existing, tc.spelling)
			}
		})
	}
}

func TestApplyGroupChanges_FollowsTheTouchedGroup(t *testing.T) {
	for _, tc := range []struct {
		name    string
		current string
		changes []GroupChange
		want    string
	}{
		{
			name:    "no operation changes nothing",
			current: "Casa",
			want:    "Casa",
		},
		{
			name:    "a rename follows the new name",
			current: "Casa",
			changes: []GroupChange{{Kind: GroupRenamed, From: "Casa", To: "Lavoro"}},
			want:    "Lavoro",
		},
		{
			name:    "a rename read with other capitalisation follows it",
			current: "CASA",
			changes: []GroupChange{{Kind: GroupRenamed, From: "Casa", To: "Lavoro"}},
			want:    "Lavoro",
		},
		{
			name:    "another Group stays as it is",
			current: "Lavoro",
			changes: []GroupChange{{Kind: GroupRenamed, From: "Casa", To: "Ufficio"}},
			want:    "Lavoro",
		},
		{
			name:    "an unsaved name stays as it is",
			current: "Nuovo",
			changes: []GroupChange{{Kind: GroupRenamed, From: "Casa", To: "Ufficio"}},
			want:    "Nuovo",
		},
		{
			name:    "a union moves the source to the survivor",
			current: "Gamma",
			changes: []GroupChange{{Kind: GroupUnited, From: "Gamma", To: "Alfa"}},
			want:    "Alfa",
		},
		{
			name:    "a union keeps the survivor and its saved spelling",
			current: "alfa",
			changes: []GroupChange{{Kind: GroupUnited, From: "Gamma", To: "Alfa"}},
			want:    "Alfa",
		},
		{
			name:    "a creation shows the saved spelling of the new name",
			current: "lavoro",
			changes: []GroupChange{{Kind: GroupCreated, To: "Lavoro"}},
			want:    "Lavoro",
		},
		{
			name:    "a creation of another name leaves the field",
			current: "Casa",
			changes: []GroupChange{{Kind: GroupCreated, To: "Lavoro"}},
			want:    "Casa",
		},
		{
			name:    "a deletion empties the field",
			current: "Casa",
			changes: []GroupChange{{Kind: GroupDeleted, From: "Casa"}},
			want:    "",
		},
		{
			name:    "a deletion of another Group leaves the field",
			current: "Lavoro",
			changes: []GroupChange{{Kind: GroupDeleted, From: "Casa"}},
			want:    "Lavoro",
		},
		{
			name:    "the records apply in order",
			current: "Casa",
			changes: []GroupChange{
				{Kind: GroupRenamed, From: "Casa", To: "Lavoro"},
				{Kind: GroupUnited, From: "Lavoro", To: "Alfa"},
			},
			want: "Alfa",
		},
		{
			name:    "an empty field stays empty",
			changes: []GroupChange{{Kind: GroupDeleted, From: "Casa"}},
			want:    "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ApplyGroupChanges(tc.current, tc.changes); got != tc.want {
				t.Errorf("ApplyGroupChanges(%q, %v) = %q, want %q", tc.current, tc.changes, got, tc.want)
			}
		})
	}
}

func TestGroupRows_CountsTheProjectsOfEveryGroupInTheChosenOrder(t *testing.T) {
	projects := []compose.Project{
		{Name: "a", Group: "Casa"},
		{Name: "b", Group: "Lavoro"},
		{Name: "c", Group: "casa"},
		{Name: "d"},
		{Name: "e", Group: "Vecchio"},
	}
	for _, tc := range []struct {
		name     string
		groups   []string
		projects []compose.Project
		want     []GroupRow
	}{
		{
			name: "no Group",
			want: []GroupRow{},
		},
		{
			name:     "the chosen order with the count of each Group",
			groups:   []string{"Lavoro", "Casa", "Vuoto"},
			projects: projects,
			want: []GroupRow{
				{Name: "Lavoro", Count: 1},
				{Name: "Casa", Count: 2},
				{Name: "Vuoto", Count: 0},
			},
		},
		{
			name:     "a Project outside the list counts nowhere",
			groups:   []string{"Casa"},
			projects: projects,
			want:     []GroupRow{{Name: "Casa", Count: 2}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := GroupRows(tc.groups, tc.projects)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("GroupRows(%v) = %v, want %v", tc.groups, got, tc.want)
			}
			if got == nil {
				t.Error("GroupRows returned nil, want an empty list")
			}
		})
	}
}

func TestMoveGroup_SwapsWithTheNeighbourAndKeepsTheRow(t *testing.T) {
	groups := []string{"Alfa", "Beta", "Gamma"}
	for _, tc := range []struct {
		name    string
		index   int
		delta   int
		want    []string
		wantRow int
	}{
		{"Sposta su in the middle", 1, -1, []string{"Beta", "Alfa", "Gamma"}, 0},
		{"Sposta giù in the middle", 1, 1, []string{"Alfa", "Gamma", "Beta"}, 2},
		{"Sposta su on the first Group changes nothing", 0, -1, groups, 0},
		{"Sposta giù on the last Group changes nothing", 2, 1, groups, 2},
		{"no selection changes nothing", -1, -1, groups, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, row := MoveGroup(groups, tc.index, tc.delta)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("MoveGroup(%v, %d, %d) = %v, want %v", groups, tc.index, tc.delta, got, tc.want)
			}
			if row != tc.wantRow {
				t.Errorf("the moved Group is at row %d, want %d", row, tc.wantRow)
			}
			if want := []string{"Alfa", "Beta", "Gamma"}; !reflect.DeepEqual(groups, want) {
				t.Errorf("MoveGroup changed its input to %v, want %v", groups, want)
			}
		})
	}
}

func TestSelectionAfterDeletion_PicksTheGroupThatTakesThePlace(t *testing.T) {
	for _, tc := range []struct {
		name    string
		removed int
		length  int
		want    int
	}{
		{"the first of three leaves the second", 0, 3, 0},
		{"the middle of three leaves the third", 1, 3, 1},
		{"the last of three leaves the new last", 2, 3, 1},
		{"the only Group leaves none", 0, 1, -1},
		{"an empty list leaves none", 0, 0, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := SelectionAfterDeletion(tc.removed, tc.length); got != tc.want {
				t.Errorf("SelectionAfterDeletion(%d, %d) = %d, want %d", tc.removed, tc.length, got, tc.want)
			}
		})
	}
}
