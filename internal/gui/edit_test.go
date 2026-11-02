package gui

import (
	"reflect"
	"testing"

	"dockercontroller/internal/compose"
)

func TestNewProjectEdit_StartsFromTheSavedValues(t *testing.T) {
	project := compose.Project{
		Name:           "shop",
		DisplayName:    "Negozio",
		Group:          "Casa",
		Profiles:       []string{"debug", "metrics"},
		ActiveProfiles: []string{"debug", "gone"},
	}

	got := NewProjectEdit(project)

	want := ProjectEdit{
		Project:     "shop",
		DisplayName: "Negozio",
		Group:       "Casa",
		Profiles:    []string{"debug", "gone"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewProjectEdit = %+v, want %+v", got, want)
	}

	// The form holds its own set: a change of the snapshot must not edit it.
	project.ActiveProfiles[0] = "changed"
	if got.Profiles[0] != "debug" {
		t.Errorf("Profiles = %v, the form shares the slice of the Project", got.Profiles)
	}
}

func TestNewProjectEdit_StartsEmptyWithoutASavedValue(t *testing.T) {
	got := NewProjectEdit(compose.Project{Name: "shop"})
	if got.DisplayName != "" || got.Group != "" {
		t.Errorf("NewProjectEdit = %+v, want an empty Display name and Group", got)
	}
	if len(got.Profiles) != 0 {
		t.Errorf("Profiles = %v, want none", got.Profiles)
	}
}

func TestNewProjectEdit_ResetsTheNewGroupField(t *testing.T) {
	got := NewProjectEdit(compose.Project{Name: "shop", Group: "Casa"})
	if got.NewGroup {
		t.Error("NewGroup = true, want the saved Group chosen")
	}
	if got.NewGroupName != "" {
		t.Errorf("NewGroupName = %q, want an empty field", got.NewGroupName)
	}
}

func TestProjectEdit_SetProfile_KeepsTheOtherProfiles(t *testing.T) {
	for _, tc := range []struct {
		name  string
		start []string
		set   string
		on    bool
		want  []string
	}{
		{"turns one on", []string{"debug"}, "metrics", true, []string{"debug", "metrics"}},
		{"keeps the set sorted", []string{"metrics"}, "debug", true, []string{"debug", "metrics"}},
		{"turns one off", []string{"debug", "metrics"}, "debug", false, []string{"metrics"}},
		{"turning off a Profile that is off changes nothing", []string{"debug"}, "metrics", false, []string{"debug"}},
		{"turning on a Profile that is on changes nothing", []string{"debug"}, "debug", true, []string{"debug"}},
		{"surrounding spaces go", []string{"debug"}, "  metrics  ", true, []string{"debug", "metrics"}},
		{"turning off a name with spaces finds it", []string{"debug", "metrics"}, " debug ", false, []string{"metrics"}},
		{"a blank name changes nothing", []string{"debug"}, "   ", true, []string{"debug"}},
		{"a blank name does not turn off a Profile", []string{"debug"}, "", false, []string{"debug"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			edit := ProjectEdit{Profiles: append([]string{}, tc.start...)}
			edit.SetProfile(tc.set, tc.on)
			if !reflect.DeepEqual(edit.Profiles, tc.want) {
				t.Errorf("Profiles = %v, want %v", edit.Profiles, tc.want)
			}
		})
	}
}

func TestProjectEdit_ProfileOn_ReadsTheSet(t *testing.T) {
	edit := ProjectEdit{Profiles: []string{"debug"}}
	if !edit.ProfileOn("debug") {
		t.Error("ProfileOn(debug) = false, want true")
	}
	if edit.ProfileOn("metrics") {
		t.Error("ProfileOn(metrics) = true, want false")
	}
}

func TestHasProfile_ReadsASetOfProfiles(t *testing.T) {
	for _, tc := range []struct {
		name     string
		profiles []string
		profile  string
		want     bool
	}{
		{"a Profile of the set", []string{"debug", "metrics"}, "metrics", true},
		{"a Profile outside the set", []string{"debug"}, "metrics", false},
		{"an empty set", nil, "debug", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := HasProfile(tc.profiles, tc.profile); got != tc.want {
				t.Errorf("HasProfile(%q) = %t, want %t", tc.profile, got, tc.want)
			}
		})
	}
}

func TestProjectEdit_GroupValue_ReadsTheGroupToSend(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit ProjectEdit
		want string
	}{
		{"the chosen Group", ProjectEdit{Group: "Casa"}, "Casa"},
		{"surrounding spaces go", ProjectEdit{Group: "  Casa  "}, "Casa"},
		{"no Group", ProjectEdit{}, ""},
		{"the new Group typed", ProjectEdit{NewGroup: true, Group: "Casa", NewGroupName: "  Nuovo "}, "Nuovo"},
		{"the new Group field left empty", ProjectEdit{NewGroup: true, Group: "Casa"}, ""},
		{"the new Group field with spaces only", ProjectEdit{NewGroup: true, NewGroupName: "   "}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.edit.GroupValue(); got != tc.want {
				t.Errorf("GroupValue = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestProjectEdit_SaveEnabled_AsksForTheNewGroupName(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit ProjectEdit
		want bool
	}{
		{"a chosen Group may be saved", ProjectEdit{Group: "Casa"}, true},
		{"no Group may be saved", ProjectEdit{}, true},
		{"the new Group named may be saved", ProjectEdit{NewGroup: true, NewGroupName: "Nuovo"}, true},
		{"the new Group named with spaces around", ProjectEdit{NewGroup: true, NewGroupName: " Nuovo "}, true},
		{"the new Group field left empty may not be saved", ProjectEdit{NewGroup: true}, false},
		{"the new Group field with spaces only may not be saved", ProjectEdit{NewGroup: true, NewGroupName: "   "}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.edit.SaveEnabled(); got != tc.want {
				t.Errorf("SaveEnabled = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestProjectEdit_Settings_TrimsTheNamesAndSendsTheCompleteSet(t *testing.T) {
	for _, tc := range []struct {
		name        string
		edit        ProjectEdit
		displayName string
		group       string
		profiles    []string
	}{
		{
			name:        "values as typed",
			edit:        ProjectEdit{DisplayName: "Negozio", Group: "Casa", Profiles: []string{"metrics", "debug"}},
			displayName: "Negozio",
			group:       "Casa",
			profiles:    []string{"debug", "metrics"},
		},
		{
			name:        "surrounding spaces go",
			edit:        ProjectEdit{DisplayName: "  Negozio  ", Group: " Casa ", Profiles: []string{" debug "}},
			displayName: "Negozio",
			group:       "Casa",
			profiles:    []string{"debug"},
		},
		{
			name:        "a blank Display name clears it",
			edit:        ProjectEdit{DisplayName: "   ", Group: "Casa"},
			displayName: "",
			group:       "Casa",
			profiles:    []string{},
		},
		{
			name:        "a blank Group clears it",
			edit:        ProjectEdit{DisplayName: "Negozio", Group: "  "},
			displayName: "Negozio",
			group:       "",
			profiles:    []string{},
		},
		{
			name:        "the new Group typed is sent",
			edit:        ProjectEdit{DisplayName: "Negozio", Group: "Casa", NewGroup: true, NewGroupName: "  Nuovo "},
			displayName: "Negozio",
			group:       "Nuovo",
			profiles:    []string{},
		},
		{
			name:        "the Profiles are deactivated with an empty set",
			edit:        ProjectEdit{DisplayName: "Negozio"},
			displayName: "Negozio",
			group:       "",
			profiles:    []string{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.edit.Settings()
			if got.DisplayName == nil || *got.DisplayName != tc.displayName {
				t.Errorf("DisplayName = %v, want %q", got.DisplayName, tc.displayName)
			}
			if got.Group == nil || *got.Group != tc.group {
				t.Errorf("Group = %v, want %q", got.Group, tc.group)
			}
			if !reflect.DeepEqual(got.Profiles, tc.profiles) {
				t.Errorf("Profiles = %v, want %v", got.Profiles, tc.profiles)
			}
		})
	}
}

func TestProjectEdit_Settings_DropsBlankAndDuplicateProfiles(t *testing.T) {
	edit := ProjectEdit{DisplayName: "Negozio", Profiles: []string{"debug", " ", "debug", "metrics"}}
	got := edit.Settings()
	if want := []string{"debug", "metrics"}; !reflect.DeepEqual(got.Profiles, want) {
		t.Errorf("Profiles = %v, want %v", got.Profiles, want)
	}
}

// TestProjectEdit_Settings_KeepsAProfileTheComposeFilesNoLongerDefine checks the
// complete set: a Profile the compose files stopped defining stays as it was,
// like the old view kept it.
func TestProjectEdit_Settings_KeepsAProfileTheComposeFilesNoLongerDefine(t *testing.T) {
	edit := ProjectEdit{DisplayName: "Negozio", Profiles: []string{"debug", "gone"}}
	got := edit.Settings()
	if want := []string{"debug", "gone"}; !reflect.DeepEqual(got.Profiles, want) {
		t.Errorf("Profiles = %v, want %v", got.Profiles, want)
	}
}

func TestGroupChoices_OffersNoGroupThenTheGroupsInUse(t *testing.T) {
	for _, tc := range []struct {
		name     string
		current  string
		projects []compose.Project
		want     []string
	}{
		{
			name: "no Group at all",
			want: []string{""},
		},
		{
			name:     "the Groups in use, sorted",
			projects: []compose.Project{{Group: "Zeta"}, {Group: "Alfa"}, {Group: "Beta"}},
			want:     []string{"", "Alfa", "Beta", "Zeta"},
		},
		{
			name:     "sorted without regard to case",
			projects: []compose.Project{{Group: "beta"}, {Group: "Alfa"}},
			want:     []string{"", "Alfa", "beta"},
		},
		{
			name:     "a Group in use appears once",
			projects: []compose.Project{{Group: "Casa"}, {Group: "Casa"}, {Group: "Lavoro"}},
			want:     []string{"", "Casa", "Lavoro"},
		},
		{
			name:     "a Project without Group offers nothing",
			projects: []compose.Project{{Group: ""}, {Group: "Casa"}},
			want:     []string{"", "Casa"},
		},
		{
			name:     "the spaces around a Group go",
			projects: []compose.Project{{Group: " Casa "}, {Group: "Casa"}},
			want:     []string{"", "Casa"},
		},
		{
			name:     "the current Group joins the ones in use",
			current:  "Vecchio",
			projects: []compose.Project{{Group: "Casa"}},
			want:     []string{"", "Casa", "Vecchio"},
		},
		{
			name:     "the current Group is not listed twice",
			current:  "Casa",
			projects: []compose.Project{{Group: "Casa"}, {Group: "Lavoro"}},
			want:     []string{"", "Casa", "Lavoro"},
		},
		{
			name:    "a blank current Group adds nothing",
			current: "   ",
			want:    []string{""},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := GroupChoices(tc.current, tc.projects); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("GroupChoices = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGroupIndex_FindsTheOptionOfAGroup(t *testing.T) {
	choices := []string{"", "Casa", "Lavoro"}
	for _, tc := range []struct {
		name  string
		group string
		want  int
	}{
		{"no Group", "", 0},
		{"a Group of the list", "Lavoro", 2},
		{"the spaces around the Group go", " Lavoro ", 2},
		{"a Group outside the list falls back to no Group", "Vecchio", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := GroupIndex(choices, tc.group); got != tc.want {
				t.Errorf("GroupIndex(%q) = %d, want %d", tc.group, got, tc.want)
			}
		})
	}
}

func TestGroupAt_ReadsTheOptionOfAGroup(t *testing.T) {
	choices := []string{"", "Casa", "Lavoro"}
	for _, tc := range []struct {
		name  string
		index int
		want  string
	}{
		{"no Group", 0, ""},
		{"a Group of the list", 2, "Lavoro"},
		{"an option past the list falls back to no Group", 7, ""},
		{"a negative option falls back to no Group", -1, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := groupAt(choices, tc.index); got != tc.want {
				t.Errorf("groupAt(%d) = %q, want %q", tc.index, got, tc.want)
			}
		})
	}
}

func TestGroupOption_ReadsTheOptionOfTheForm(t *testing.T) {
	choices := []string{"", "Casa", "Lavoro"}
	for _, tc := range []struct {
		name string
		edit ProjectEdit
		want int
	}{
		{"no Group", ProjectEdit{}, 0},
		{"the chosen Group", ProjectEdit{Group: "Lavoro"}, 2},
		{"a Group outside the list reads as no Group", ProjectEdit{Group: "Vecchio"}, 0},
		{"a new Group asks for the last option", ProjectEdit{Group: "Lavoro", NewGroup: true}, 3},
		{"a new Group asks for the last option without a Group", ProjectEdit{NewGroup: true}, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := GroupOption(choices, tc.edit); got != tc.want {
				t.Errorf("GroupOption = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestProjectEdit_SetGroupOption_PicksTheOptionAndTheNewGroupField(t *testing.T) {
	choices := []string{"", "Casa", "Lavoro"}
	for _, tc := range []struct {
		name         string
		edit         ProjectEdit
		index        int
		group        string
		newGroup     bool
		newGroupName string
	}{
		{
			name:         "no Group",
			edit:         ProjectEdit{Group: "Lavoro", NewGroup: true, NewGroupName: "Nuovo"},
			index:        0,
			group:        "",
			newGroup:     false,
			newGroupName: "Nuovo",
		},
		{
			name:         "a Group of the list",
			edit:         ProjectEdit{Group: "Casa"},
			index:        2,
			group:        "Lavoro",
			newGroup:     false,
			newGroupName: "",
		},
		{
			name:         "the last option asks for a new Group",
			edit:         ProjectEdit{Group: "Casa"},
			index:        3,
			group:        "Casa",
			newGroup:     true,
			newGroupName: "",
		},
		{
			name:         "the last option keeps the name already typed",
			edit:         ProjectEdit{Group: "Casa", NewGroup: true, NewGroupName: "Nuovo"},
			index:        3,
			group:        "Casa",
			newGroup:     true,
			newGroupName: "Nuovo",
		},
		{
			name:         "an option past the field falls back to no Group",
			edit:         ProjectEdit{Group: "Casa", NewGroup: true, NewGroupName: "Nuovo"},
			index:        9,
			group:        "",
			newGroup:     false,
			newGroupName: "Nuovo",
		},
		{
			name:         "a negative option falls back to no Group",
			edit:         ProjectEdit{Group: "Casa"},
			index:        -1,
			group:        "",
			newGroup:     false,
			newGroupName: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			edit := tc.edit
			edit.SetGroupOption(choices, tc.index)
			if edit.Group != tc.group {
				t.Errorf("Group = %q, want %q", edit.Group, tc.group)
			}
			if edit.NewGroup != tc.newGroup {
				t.Errorf("NewGroup = %t, want %t", edit.NewGroup, tc.newGroup)
			}
			if edit.NewGroupName != tc.newGroupName {
				t.Errorf("NewGroupName = %q, want %q", edit.NewGroupName, tc.newGroupName)
			}
		})
	}
}

func TestProfileChoices_OffersTheProfilesOfTheProject(t *testing.T) {
	for _, tc := range []struct {
		name     string
		profiles []string
		want     []string
	}{
		{"none", nil, []string{}},
		{"the defined order", []string{"debug", "metrics"}, []string{"debug", "metrics"}},
		{"a Profile appears once", []string{"debug", "debug"}, []string{"debug"}},
		{"blank names go", []string{"debug", " "}, []string{"debug"}},
		{"the spaces around a name go", []string{" debug ", "metrics"}, []string{"debug", "metrics"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ProfileChoices(compose.Project{Profiles: tc.profiles})
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ProfileChoices = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestGroupSuggestions_DropsTheEmptyOption checks the options of the editable
// box of the Group: every Group of the choices, never the empty one, which the
// box shows as its hint instead.
func TestGroupSuggestions_DropsTheEmptyOption(t *testing.T) {
	for _, tc := range []struct {
		name    string
		choices []string
		want    []string
	}{
		{"no Group to propose", []string{""}, []string{}},
		{"no choice at all", nil, []string{}},
		{"the Groups of the choices keep their order", []string{"", "Casa", "Lavoro"}, []string{"Casa", "Lavoro"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := GroupSuggestions(tc.choices)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("GroupSuggestions(%v) = %v, want %v", tc.choices, got, tc.want)
			}
			if got == nil {
				t.Error("GroupSuggestions returned nil, want an empty list")
			}
		})
	}
}
