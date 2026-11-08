package forms

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
		{"spaces only is no Group", ProjectEdit{Group: "   "}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.edit.GroupValue(); got != tc.want {
				t.Errorf("GroupValue = %q, want %q", got, tc.want)
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
			name:        "a new name outside the list is sent as typed",
			edit:        ProjectEdit{DisplayName: "Negozio", Group: "  Nuovo "},
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

func TestGroupChoices_OffersNoGroupThenTheSavedGroupsInTheirOrder(t *testing.T) {
	for _, tc := range []struct {
		name    string
		current string
		groups  []string
		want    []string
	}{
		{
			name: "no Group at all",
			want: []string{""},
		},
		{
			name:   "the saved Groups in the chosen order",
			groups: []string{"Zeta", "Alfa", "Beta"},
			want:   []string{"", "Zeta", "Alfa", "Beta"},
		},
		{
			name:   "an empty saved Group offers nothing",
			groups: []string{"", "Casa"},
			want:   []string{"", "Casa"},
		},
		{
			name:   "the spaces around a Group go",
			groups: []string{" Casa ", "Casa"},
			want:   []string{"", "Casa"},
		},
		{
			name:    "the current Group outside the list joins at the end",
			current: "Vecchio",
			groups:  []string{"Casa", "Lavoro"},
			want:    []string{"", "Casa", "Lavoro", "Vecchio"},
		},
		{
			name:    "the current Group of the list is not listed twice",
			current: "Casa",
			groups:  []string{"Casa", "Lavoro"},
			want:    []string{"", "Casa", "Lavoro"},
		},
		{
			name:    "the current Group differing only in case is the saved one",
			current: "CASA",
			groups:  []string{"Casa", "Lavoro"},
			want:    []string{"", "Casa", "Lavoro"},
		},
		{
			name:    "a blank current Group adds nothing",
			current: "   ",
			groups:  []string{"Casa"},
			want:    []string{"", "Casa"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := GroupChoices(tc.current, tc.groups); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("GroupChoices = %v, want %v", got, tc.want)
			}
		})
	}
}

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
