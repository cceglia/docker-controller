package compose_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/model"
)

// TestCreateGroup_AddsTheGroupAtTheEndAndPublishesAtOnce checks a creation: the
// Group goes after the saved ones, the snapshot carries it without a Refresh,
// and a new process reads it back.
func TestCreateGroup_AddsTheGroupAtTheEndAndPublishesAtOnce(t *testing.T) {
	path := tempDB(t)
	m := refreshed(t, path)
	ctx := context.Background()

	if err := m.CreateGroup(ctx, "  Lavoro  "); err != nil {
		t.Fatal(err)
	}
	if err := m.CreateGroup(ctx, "Casa"); err != nil {
		t.Fatal(err)
	}

	if want := []string{"Lavoro", "Casa"}; !reflect.DeepEqual(m.Snapshot().Groups, want) {
		t.Errorf("Groups = %v, want %v after the creation, without a Refresh", m.Snapshot().Groups, want)
	}

	again := refreshed(t, path)
	if want := []string{"Lavoro", "Casa"}; !reflect.DeepEqual(again.Snapshot().Groups, want) {
		t.Errorf("Groups = %v after reopening, want %v", again.Snapshot().Groups, want)
	}
}

// TestCreateGroup_RefusesADuplicateWithTheSavedSpelling checks the refusal of a
// name that already exists, even with different capitalisation: the error names
// the saved Group so the dialog can say which one it is.
func TestCreateGroup_RefusesADuplicateWithTheSavedSpelling(t *testing.T) {
	m := refreshed(t, tempDB(t))
	ctx := context.Background()
	if err := m.CreateGroup(ctx, "Lavoro"); err != nil {
		t.Fatal(err)
	}

	err := m.CreateGroup(ctx, "LAVORO")
	var duplicate *model.DuplicateGroupError
	if !errors.As(err, &duplicate) {
		t.Fatalf("CreateGroup(LAVORO) = %v, want a DuplicateGroupError", err)
	}
	if duplicate.Existing != "Lavoro" {
		t.Errorf("Existing = %q, want the saved spelling %q", duplicate.Existing, "Lavoro")
	}
	if !errors.Is(err, compose.ErrDuplicateGroup) {
		t.Errorf("the error does not wrap ErrDuplicateGroup: %v", err)
	}
	if want := []string{"Lavoro"}; !reflect.DeepEqual(m.Snapshot().Groups, want) {
		t.Errorf("Groups = %v after a refused creation, want %v", m.Snapshot().Groups, want)
	}
}

// TestCreateGroup_RefusesABlankOrReservedName checks the names that can never
// be a Group.
func TestCreateGroup_RefusesABlankOrReservedName(t *testing.T) {
	m := refreshed(t, tempDB(t))
	ctx := context.Background()

	for _, name := range []string{"", "   ", "Nessun Gruppo", "nessun gruppo", "Senza Gruppo", "SENZA GRUPPO"} {
		t.Run(name, func(t *testing.T) {
			if err := m.CreateGroup(ctx, name); !errors.Is(err, compose.ErrInvalidGroupName) {
				t.Errorf("CreateGroup(%q) = %v, want ErrInvalidGroupName", name, err)
			}
		})
	}
	if len(m.Snapshot().Groups) != 0 {
		t.Errorf("Groups = %v after refused names, want none", m.Snapshot().Groups)
	}
}

// TestRenameGroup_ReachesEveryProject checks a rename: the new name arrives on
// every Project of the Group and the Group keeps its place.
func TestRenameGroup_ReachesEveryProject(t *testing.T) {
	m := refreshed(t, tempDB(t))
	ctx := context.Background()
	if err := m.CreateGroup(ctx, "Zeta"); err != nil {
		t.Fatal(err)
	}
	if err := m.CreateGroup(ctx, "Lavoro"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"sunlinkdocker", "personalvpn"} {
		if err := m.UpdateSettings(ctx, name, compose.Settings{Group: ptr("Lavoro")}); err != nil {
			t.Fatal(err)
		}
	}

	if err := m.RenameGroup(ctx, "Lavoro", "  Ufficio "); err != nil {
		t.Fatal(err)
	}

	s := m.Snapshot()
	if want := []string{"Zeta", "Ufficio"}; !reflect.DeepEqual(s.Groups, want) {
		t.Errorf("Groups = %v, want %v, the Group keeps its place", s.Groups, want)
	}
	for _, name := range []string{"sunlinkdocker", "personalvpn"} {
		if p := find(t, s, name); p.Group != "Ufficio" {
			t.Errorf("%s.Group = %q after the rename, want %q", name, p.Group, "Ufficio")
		}
	}
}

// TestRenameGroup_ACaseChangeKeepsTheSameGroup checks that changing only the
// spelling of a Group neither creates nor removes a Group.
func TestRenameGroup_ACaseChangeKeepsTheSameGroup(t *testing.T) {
	m := refreshed(t, tempDB(t))
	ctx := context.Background()
	if err := m.UpdateSettings(ctx, "sunlinkdocker", compose.Settings{Group: ptr("Città")}); err != nil {
		t.Fatal(err)
	}

	if err := m.RenameGroup(ctx, "Città", "CITTÀ"); err != nil {
		t.Fatal(err)
	}

	s := m.Snapshot()
	if want := []string{"CITTÀ"}; !reflect.DeepEqual(s.Groups, want) {
		t.Errorf("Groups = %v, want %v", s.Groups, want)
	}
	if p := find(t, s, "sunlinkdocker"); p.Group != "CITTÀ" {
		t.Errorf("Group = %q after the case change, want %q", p.Group, "CITTÀ")
	}
}

// TestRenameGroup_UnitesTheTwoGroups checks the union: the source leaves, the
// Group that already had the name keeps its spelling and its relative place,
// and the Projects of the source join it.
func TestRenameGroup_UnitesTheTwoGroups(t *testing.T) {
	m := refreshed(t, tempDB(t))
	ctx := context.Background()
	for _, name := range []string{"Alfa", "Beta", "Gamma"} {
		if err := m.CreateGroup(ctx, name); err != nil {
			t.Fatal(err)
		}
	}
	for project, group := range map[string]string{
		"sunlinkdocker": "Gamma",
		"personalvpn":   "Alfa",
		"stellantisvpn": "Beta",
	} {
		if err := m.UpdateSettings(ctx, project, compose.Settings{Group: ptr(group)}); err != nil {
			t.Fatal(err)
		}
	}

	if err := m.RenameGroup(ctx, "Gamma", "Alfa"); err != nil {
		t.Fatal(err)
	}

	s := m.Snapshot()
	if want := []string{"Alfa", "Beta"}; !reflect.DeepEqual(s.Groups, want) {
		t.Errorf("Groups = %v after the union, want %v", s.Groups, want)
	}
	for _, name := range []string{"sunlinkdocker", "personalvpn"} {
		if p := find(t, s, name); p.Group != "Alfa" {
			t.Errorf("%s.Group = %q after the union, want the survivor %q", name, p.Group, "Alfa")
		}
	}
	if p := find(t, s, "stellantisvpn"); p.Group != "Beta" {
		t.Errorf("stellantisvpn.Group = %q, the union touched another Group", p.Group)
	}
}

// TestRenameGroup_RefusesAnUnknownGroup checks a rename of a Group that is not
// saved.
func TestRenameGroup_RefusesAnUnknownGroup(t *testing.T) {
	m := refreshed(t, tempDB(t))
	if err := m.RenameGroup(context.Background(), "Nonexistent", "Altro"); !errors.Is(err, compose.ErrUnknownGroup) {
		t.Errorf("RenameGroup of an unknown Group = %v, want ErrUnknownGroup", err)
	}
}

// TestDeleteGroup_LeavesItsProjectsWithoutAGroup checks a deletion: the Group
// leaves the list, its Projects stay without one, and Docker is not touched.
func TestDeleteGroup_LeavesItsProjectsWithoutAGroup(t *testing.T) {
	m := refreshed(t, tempDB(t))
	ctx := context.Background()
	for _, name := range []string{"Zeta", "Lavoro"} {
		if err := m.CreateGroup(ctx, name); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"sunlinkdocker", "personalvpn"} {
		if err := m.UpdateSettings(ctx, name, compose.Settings{Group: ptr("Lavoro")}); err != nil {
			t.Fatal(err)
		}
	}
	before := find(t, m.Snapshot(), "sunlinkdocker")

	if err := m.DeleteGroup(ctx, "Lavoro"); err != nil {
		t.Fatal(err)
	}

	s := m.Snapshot()
	if want := []string{"Zeta"}; !reflect.DeepEqual(s.Groups, want) {
		t.Errorf("Groups = %v after the deletion, want %v", s.Groups, want)
	}
	for _, name := range []string{"sunlinkdocker", "personalvpn"} {
		if p := find(t, s, name); p.Group != "" {
			t.Errorf("%s.Group = %q after the deletion, want no Group", name, p.Group)
		}
	}
	// The Project itself stays, with the state Docker reported.
	if p := find(t, s, "sunlinkdocker"); p.Name != before.Name || p.Status != before.Status {
		t.Errorf("the deletion changed the Project to %+v", p)
	}
}

// TestDeleteGroup_RefusesAnUnknownGroup checks a deletion of a Group that is
// not saved.
func TestDeleteGroup_RefusesAnUnknownGroup(t *testing.T) {
	m := refreshed(t, tempDB(t))
	if err := m.DeleteGroup(context.Background(), "Nonexistent"); !errors.Is(err, compose.ErrUnknownGroup) {
		t.Errorf("DeleteGroup of an unknown Group = %v, want ErrUnknownGroup", err)
	}
}

// TestGroupOperations_WorkWhileDockerIsUnreachable checks that the three
// operations publish a snapshot without asking Docker.
func TestGroupOperations_WorkWhileDockerIsUnreachable(t *testing.T) {
	f := realDocker(t)
	m := newManager(t, f, tempDB(t))
	ctx := context.Background()
	if err := m.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	f.RespondErr(nil, errors.New("Cannot connect to the Docker daemon"), lsArgs...)
	if err := m.Refresh(ctx); err != nil {
		t.Fatal(err)
	}

	if err := m.CreateGroup(ctx, "Lavoro"); err != nil {
		t.Fatalf("CreateGroup with Docker down: %v", err)
	}
	if err := m.UpdateSettings(ctx, "sunlinkdocker", compose.Settings{Group: ptr("Lavoro")}); err != nil {
		t.Fatal(err)
	}
	if err := m.RenameGroup(ctx, "Lavoro", "Ufficio"); err != nil {
		t.Fatalf("RenameGroup with Docker down: %v", err)
	}
	s := m.Snapshot()
	if s.DockerReachable {
		t.Error("the operations asked Docker for a state it could not give")
	}
	if p := find(t, s, "sunlinkdocker"); p.Group != "Ufficio" {
		t.Errorf("Group = %q after the rename, want %q", p.Group, "Ufficio")
	}

	if err := m.DeleteGroup(ctx, "Ufficio"); err != nil {
		t.Fatalf("DeleteGroup with Docker down: %v", err)
	}
	s = m.Snapshot()
	if len(s.Groups) != 0 {
		t.Errorf("Groups = %v after the deletion, want none", s.Groups)
	}
	if p := find(t, s, "sunlinkdocker"); p.Group != "" {
		t.Errorf("Group = %q after the deletion, want none", p.Group)
	}
}
