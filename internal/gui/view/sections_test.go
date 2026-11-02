package view

import (
	"testing"

	"dockercontroller/internal/compose"
)

func TestSections_KeepsTheGroupsSortedAndTheUngroupedLast(t *testing.T) {
	projects := []compose.Project{
		{Name: "solo", Group: ""},
		{Name: "work-b", Group: "Lavoro"},
		{Name: "home", Group: "Casa"},
		{Name: "work-a", Group: "Lavoro"},
	}

	got := Sections(projects)

	if len(got) != 3 {
		t.Fatalf("%d sections, want 3: %+v", len(got), got)
	}
	if got[0].Title != "Casa" || len(got[0].Projects) != 1 || got[0].Projects[0].Name != "home" {
		t.Errorf("first section = %+v, want Casa/home", got[0])
	}
	if got[1].Title != "Lavoro" || len(got[1].Projects) != 2 {
		t.Fatalf("second section = %+v, want Lavoro with 2 Projects", got[1])
	}
	// The order inside a Group is the one of the snapshot.
	if got[1].Projects[0].Name != "work-b" || got[1].Projects[1].Name != "work-a" {
		t.Errorf("Lavoro order = %v, want work-b then work-a", got[1].Projects)
	}
	if got[2].Title != UngroupedTitle || len(got[2].Projects) != 1 || got[2].Projects[0].Name != "solo" {
		t.Errorf("last section = %+v, want %s/solo", got[2], UngroupedTitle)
	}
}

func TestSections_WithoutGroupsStaysOneFlatUnnamedSection(t *testing.T) {
	got := Sections([]compose.Project{{Name: "a"}, {Name: "b"}})

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
	got := Sections(nil)

	if len(got) != 1 || got[0].Title != "" || len(got[0].Projects) != 0 {
		t.Errorf("got %+v, want one empty section without a heading", got)
	}
}

func TestSections_DoesNotChangeTheInput(t *testing.T) {
	projects := []compose.Project{{Name: "a", Group: "B"}, {Name: "b", Group: "A"}}

	_ = Sections(projects)

	if projects[0].Name != "a" || projects[1].Name != "b" {
		t.Errorf("the input order changed: %+v", projects)
	}
}
