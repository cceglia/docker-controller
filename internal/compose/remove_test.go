package compose_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/composefake"
)

// removedProjects returns a Manager (on dbPath) whose Docker has no Container
// at all: every saved Project is Rimosso.
func removedProjects(t *testing.T, dbPath string) (*compose.Manager, *composefake.Fake) {
	t.Helper()
	first := newManager(t, realDocker(t), dbPath)
	if err := first.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	f := composefake.New()
	f.Respond([]byte("[]"), lsArgs...)
	m := newManager(t, f, dbPath)
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	return m, f
}

func listed(s compose.Snapshot, name string) bool {
	for _, p := range s.Projects {
		if p.Name == name {
			return true
		}
	}
	return false
}

func TestRemoveFromList_DropsARemovedProjectForGoodWithoutTouchingDocker(t *testing.T) {
	path := tempDB(t)
	m, f := removedProjects(t, path)
	calls := len(f.Calls())

	if err := m.RemoveFromList(context.Background(), "stellantisvpn"); err != nil {
		t.Fatal(err)
	}
	if listed(m.Snapshot(), "stellantisvpn") {
		t.Error("the Project is still in the snapshot")
	}
	if n := len(m.Snapshot().Projects); n != 4 {
		t.Errorf("%d Projects left, want 4", n)
	}
	for _, call := range f.Calls()[calls:] {
		if len(call) < 2 || call[1] != "ls" {
			t.Errorf("RemoveFromList may only re-read Docker, ran %v", call)
		}
	}

	// A new process on the same DB no longer knows it either.
	f2 := composefake.New()
	f2.Respond([]byte("[]"), lsArgs...)
	again := newManager(t, f2, path)
	if err := again.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if listed(again.Snapshot(), "stellantisvpn") {
		t.Error("the Project came back after reopening the DB")
	}
}

func TestRemoveFromList_RefusesAProjectWithContainers(t *testing.T) {
	for _, name := range []string{"sunlinkdocker", "stellantisvpn"} { // running, and stopped but with Containers
		t.Run(name, func(t *testing.T) {
			m := newManager(t, realDocker(t), tempDB(t))
			if err := m.Refresh(context.Background()); err != nil {
				t.Fatal(err)
			}
			err := m.RemoveFromList(context.Background(), name)
			if !errors.Is(err, compose.ErrProjectNotRemoved) {
				t.Errorf("err = %v, want ErrProjectNotRemoved", err)
			}
			if !listed(m.Snapshot(), name) {
				t.Error("a refused removal must keep the Project")
			}
		})
	}
}

func TestRemoveFromList_RefusesAnUnknownProject(t *testing.T) {
	m, _ := removedProjects(t, tempDB(t))
	if err := m.RemoveFromList(context.Background(), "nope"); !errors.Is(err, compose.ErrUnknownProject) {
		t.Errorf("err = %v, want ErrUnknownProject", err)
	}
}

func TestRemoveFromList_RefusesWhileAnActivityRuns(t *testing.T) {
	m, f := removedProjects(t, tempDB(t))
	args := composeArgs(t, m, "stellantisvpn", "up", "-d")
	release := f.Hold(args...)
	f.RespondStream(nil, nil, nil, args...)
	if _, err := m.Do(context.Background(), compose.Action{Project: "stellantisvpn", Kind: compose.ActionUp}); err != nil {
		t.Fatal(err)
	}
	err := m.RemoveFromList(context.Background(), "stellantisvpn")
	if !errors.Is(err, compose.ErrActivityRunning) {
		t.Errorf("err = %v, want ErrActivityRunning", err)
	}

	release()
	for deadline := time.Now().Add(2 * time.Second); find(t, m.Snapshot(), "stellantisvpn").Activity != nil; {
		if time.Now().After(deadline) {
			t.Fatal("the Activity did not finish")
		}
		time.Sleep(2 * time.Millisecond)
	}
}

func TestRemoveFromList_RefusesWhenDockerIsUnreachable(t *testing.T) {
	// With Docker silent a Project cannot be known to have no Containers.
	m, f := removedProjects(t, tempDB(t))
	f.RespondErr(nil, errors.New("Cannot connect to the Docker daemon"), lsArgs...)
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	err := m.RemoveFromList(context.Background(), "stellantisvpn")
	if !errors.Is(err, compose.ErrDockerUnreachable) {
		t.Errorf("err = %v, want ErrDockerUnreachable", err)
	}
	if !listed(m.Snapshot(), "stellantisvpn") {
		t.Error("a refused removal must keep the Project")
	}
}
