package compose_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/model"
	"dockercontroller/internal/gui/view"
	"dockercontroller/internal/tray"
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

// TestMoveGroup_SwapsTheGroupWithItsNeighbour checks a move in both
// directions: the Group changes place with its neighbour and the snapshot
// follows at once, without a Refresh.
func TestMoveGroup_SwapsTheGroupWithItsNeighbour(t *testing.T) {
	m := refreshed(t, tempDB(t))
	ctx := context.Background()
	for _, name := range []string{"Alfa", "Beta", "Gamma"} {
		if err := m.CreateGroup(ctx, name); err != nil {
			t.Fatal(err)
		}
	}

	if err := m.MoveGroup(ctx, "Beta", -1); err != nil {
		t.Fatal(err)
	}
	if want := []string{"Beta", "Alfa", "Gamma"}; !reflect.DeepEqual(m.Snapshot().Groups, want) {
		t.Errorf("Groups = %v after Sposta su, want %v", m.Snapshot().Groups, want)
	}

	if err := m.MoveGroup(ctx, "Beta", 1); err != nil {
		t.Fatal(err)
	}
	if want := []string{"Alfa", "Beta", "Gamma"}; !reflect.DeepEqual(m.Snapshot().Groups, want) {
		t.Errorf("Groups = %v after Sposta giù, want %v", m.Snapshot().Groups, want)
	}
}

// TestMoveGroup_PastEitherEndChangesNothing checks the two edges: Sposta su on
// the first Group and Sposta giù on the last one do nothing and are not an
// error.
func TestMoveGroup_PastEitherEndChangesNothing(t *testing.T) {
	m := refreshed(t, tempDB(t))
	ctx := context.Background()
	for _, name := range []string{"Alfa", "Beta", "Gamma"} {
		if err := m.CreateGroup(ctx, name); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{"Alfa", "Beta", "Gamma"}

	if err := m.MoveGroup(ctx, "Alfa", -1); err != nil {
		t.Fatalf("Sposta su on the first Group = %v, want no error", err)
	}
	if got := m.Snapshot().Groups; !reflect.DeepEqual(got, want) {
		t.Errorf("Groups = %v after Sposta su on the first Group, want %v", got, want)
	}

	if err := m.MoveGroup(ctx, "Gamma", 1); err != nil {
		t.Fatalf("Sposta giù on the last Group = %v, want no error", err)
	}
	if got := m.Snapshot().Groups; !reflect.DeepEqual(got, want) {
		t.Errorf("Groups = %v after Sposta giù on the last Group, want %v", got, want)
	}

	if err := m.MoveGroup(ctx, "Beta", 0); err != nil {
		t.Fatalf("a move of zero places = %v, want no error", err)
	}
	if got := m.Snapshot().Groups; !reflect.DeepEqual(got, want) {
		t.Errorf("Groups = %v after a move of zero places, want %v", got, want)
	}
}

// TestMoveGroup_RefusesAnUnknownGroup checks a move of a Group that a stale
// dialog still shows: the Manager refuses it and the saved order stays.
func TestMoveGroup_RefusesAnUnknownGroup(t *testing.T) {
	m := refreshed(t, tempDB(t))
	ctx := context.Background()
	if err := m.CreateGroup(ctx, "Alfa"); err != nil {
		t.Fatal(err)
	}

	if err := m.MoveGroup(ctx, "Fantasma", -1); !errors.Is(err, compose.ErrUnknownGroup) {
		t.Errorf("MoveGroup of an unknown Group = %v, want ErrUnknownGroup", err)
	}
	if err := m.MoveGroup(ctx, "Fantasma", 0); !errors.Is(err, compose.ErrUnknownGroup) {
		t.Errorf("a zero move of an unknown Group = %v, want ErrUnknownGroup", err)
	}
	if want := []string{"Alfa"}; !reflect.DeepEqual(m.Snapshot().Groups, want) {
		t.Errorf("Groups = %v after a refused move, want %v", m.Snapshot().Groups, want)
	}
}

// TestMoveGroup_KeepsTheOrderAfterTheDatabaseReopens checks that the chosen
// order survives the end of one process and the reopening of the database.
func TestMoveGroup_KeepsTheOrderAfterTheDatabaseReopens(t *testing.T) {
	path := tempDB(t)
	m := refreshed(t, path)
	ctx := context.Background()
	for _, name := range []string{"Alfa", "Beta", "Gamma"} {
		if err := m.CreateGroup(ctx, name); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.MoveGroup(ctx, "Gamma", -1); err != nil {
		t.Fatal(err)
	}

	again := refreshed(t, path)
	if want := []string{"Alfa", "Gamma", "Beta"}; !reflect.DeepEqual(again.Snapshot().Groups, want) {
		t.Errorf("Groups = %v after reopening, want %v", again.Snapshot().Groups, want)
	}
}

// TestMoveGroup_MovesTheNeighboursAfterADeletion checks that a deletion closes
// its gap: the Groups left still swap with their neighbours and keep the
// expected order.
func TestMoveGroup_MovesTheNeighboursAfterADeletion(t *testing.T) {
	m := refreshed(t, tempDB(t))
	ctx := context.Background()
	for _, name := range []string{"Alfa", "Beta", "Gamma"} {
		if err := m.CreateGroup(ctx, name); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.DeleteGroup(ctx, "Beta"); err != nil {
		t.Fatal(err)
	}
	if err := m.MoveGroup(ctx, "Gamma", -1); err != nil {
		t.Fatal(err)
	}
	if want := []string{"Gamma", "Alfa"}; !reflect.DeepEqual(m.Snapshot().Groups, want) {
		t.Errorf("Groups = %v after the move, want %v", m.Snapshot().Groups, want)
	}
	if err := m.MoveGroup(ctx, "Alfa", -1); err != nil {
		t.Fatal(err)
	}
	if want := []string{"Alfa", "Gamma"}; !reflect.DeepEqual(m.Snapshot().Groups, want) {
		t.Errorf("Groups = %v after the second move, want %v", m.Snapshot().Groups, want)
	}
}

// TestMoveGroup_TheChosenOrderReachesSectionsAndTray checks that the list and
// the tray read the order the Manager publishes after a move.
func TestMoveGroup_TheChosenOrderReachesSectionsAndTray(t *testing.T) {
	m := refreshed(t, tempDB(t))
	ctx := context.Background()
	for _, name := range []string{"Lavoro", "Casa"} {
		if err := m.CreateGroup(ctx, name); err != nil {
			t.Fatal(err)
		}
	}
	for project, group := range map[string]string{
		"sunlinkdocker": "Lavoro",
		"personalvpn":   "Casa",
	} {
		if err := m.UpdateSettings(ctx, project, compose.Settings{Group: ptr(group)}); err != nil {
			t.Fatal(err)
		}
	}
	if err := m.MoveGroup(ctx, "Casa", -1); err != nil {
		t.Fatal(err)
	}

	s := m.Snapshot()
	var titles []string
	for _, section := range view.Sections(s.Projects, s.Groups) {
		titles = append(titles, section.Title)
	}
	if want := []string{"Casa", "Lavoro", view.UngroupedTitle}; !reflect.DeepEqual(titles, want) {
		t.Errorf("the sections are %v after the move, want %v", titles, want)
	}

	var menu []string
	for _, item := range tray.BuildMenu(s) {
		if item.Label == "" || item.Separator {
			continue
		}
		menu = append(menu, item.Label)
	}
	if len(menu) < 2 || menu[0] != "Casa" || menu[1] != "Lavoro" {
		t.Errorf("the tray menu starts with %v after the move, want Casa then Lavoro", menu)
	}
}

// TestGroupOperations_WorkWhileDockerIsUnreachable checks that the operations
// publish a snapshot without asking Docker.
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

	if err := m.CreateGroup(ctx, "Casa"); err != nil {
		t.Fatalf("CreateGroup with Docker down: %v", err)
	}
	if err := m.MoveGroup(ctx, "Casa", -1); err != nil {
		t.Fatalf("MoveGroup with Docker down: %v", err)
	}
	s = m.Snapshot()
	if s.DockerReachable {
		t.Error("the move asked Docker for a state it could not give")
	}
	if want := []string{"Casa", "Ufficio"}; !reflect.DeepEqual(s.Groups, want) {
		t.Errorf("Groups = %v after the move, want %v", s.Groups, want)
	}

	if err := m.DeleteGroup(ctx, "Ufficio"); err != nil {
		t.Fatalf("DeleteGroup with Docker down: %v", err)
	}
	s = m.Snapshot()
	if want := []string{"Casa"}; !reflect.DeepEqual(s.Groups, want) {
		t.Errorf("Groups = %v after the deletion, want %v", s.Groups, want)
	}
	if p := find(t, s, "sunlinkdocker"); p.Group != "" {
		t.Errorf("Group = %q after the deletion, want none", p.Group)
	}
}
