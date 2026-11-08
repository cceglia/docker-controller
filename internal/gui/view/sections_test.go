package view

import (
	"testing"

	"dockercontroller/internal/compose"
)

func TestSections_KeepsTheSnapshotOrderAndTheUngroupedLast(t *testing.T) {
	projects := []compose.Project{
		{Name: "solo", Group: ""},
		{Name: "work-b", Group: "Lavoro"},
		{Name: "home", Group: "Casa"},
		{Name: "work-a", Group: "Lavoro"},
	}

	// The user chose Lavoro before Casa: the sections follow that order, not
	// the alphabetical one.
	got := Sections(projects, []string{"Lavoro", "Casa"})

	if len(got) != 3 {
		t.Fatalf("%d sections, want 3: %+v", len(got), got)
	}
	if got[0].Title != "Lavoro" || len(got[0].Projects) != 2 {
		t.Fatalf("first section = %+v, want Lavoro with 2 Projects", got[0])
	}
	// The order inside a Group is the one of the snapshot.
	if got[0].Projects[0].Name != "work-b" || got[0].Projects[1].Name != "work-a" {
		t.Errorf("Lavoro order = %v, want work-b then work-a", got[0].Projects)
	}
	if got[1].Title != "Casa" || len(got[1].Projects) != 1 || got[1].Projects[0].Name != "home" {
		t.Errorf("second section = %+v, want Casa/home", got[1])
	}
	if got[2].Title != UngroupedTitle || len(got[2].Projects) != 1 || got[2].Projects[0].Name != "solo" {
		t.Errorf("last section = %+v, want %s/solo", got[2], UngroupedTitle)
	}
}

func TestSections_PutsAnUnknownGroupAfterTheKnownOnesAndBeforeTheUngrouped(t *testing.T) {
	projects := []compose.Project{
		{Name: "solo", Group: ""},
		{Name: "z", Group: "Zeta"},
		{Name: "known", Group: "Lavoro"},
		{Name: "alfa", Group: "Alfa"},
	}

	got := Sections(projects, []string{"Lavoro"})

	var titles []string
	for _, section := range got {
		titles = append(titles, section.Title)
	}
	want := []string{"Lavoro", "Alfa", "Zeta", UngroupedTitle}
	if len(titles) != len(want) {
		t.Fatalf("sections = %v, want %v", titles, want)
	}
	for i := range want {
		if titles[i] != want[i] {
			t.Fatalf("sections = %v, want %v", titles, want)
		}
	}
	if got[1].Projects[0].Name != "alfa" || got[2].Projects[0].Name != "z" {
		t.Errorf("unknown groups lost their Projects: %+v", got)
	}
}

func TestSections_SavedEmptyGroupProducesNoHeading(t *testing.T) {
	got := Sections([]compose.Project{{Name: "a"}, {Name: "b"}}, []string{"Lavoro", "Casa"})

	if len(got) != 1 {
		t.Fatalf("%d sections, want 1: %+v", len(got), got)
	}
	if got[0].Title != "" {
		t.Errorf("title = %q, want no heading", got[0].Title)
	}
	if len(got[0].Projects) != 2 {
		t.Errorf("%d Projects, want 2", len(got[0].Projects))
	}
}

func TestSections_EmptyListIsOneEmptySection(t *testing.T) {
	got := Sections(nil, nil)

	if len(got) != 1 || got[0].Title != "" || len(got[0].Projects) != 0 {
		t.Errorf("got %+v, want one empty section without a heading", got)
	}
}

func TestSections_DoesNotChangeTheInput(t *testing.T) {
	projects := []compose.Project{{Name: "a", Group: "B"}, {Name: "b", Group: "A"}}

	_ = Sections(projects, []string{"A", "B"})

	if projects[0].Name != "a" || projects[1].Name != "b" {
		t.Errorf("the input order changed: %+v", projects)
	}
}
