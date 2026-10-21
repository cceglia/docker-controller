package compose_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"dockercontroller/internal/compose"
)

func ptr(s string) *string { return &s }

func refreshed(t *testing.T, dbPath string) *compose.Manager {
	t.Helper()
	m := newManager(t, realDocker(t), dbPath)
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestUpdateSettings_SetsDisplayNameAndGroupAndListsTheGroups(t *testing.T) {
	m := refreshed(t, tempDB(t))
	ctx := context.Background()

	if err := m.UpdateSettings(ctx, "sunlinkdocker", compose.Settings{DisplayName: ptr("Sunlink"), Group: ptr("Lavoro")}); err != nil {
		t.Fatal(err)
	}
	if err := m.UpdateSettings(ctx, "personalvpn", compose.Settings{Group: ptr("Casa")}); err != nil {
		t.Fatal(err)
	}
	if err := m.UpdateSettings(ctx, "stellantisvpn", compose.Settings{Group: ptr("Lavoro")}); err != nil {
		t.Fatal(err)
	}

	s := m.Snapshot()
	p := find(t, s, "sunlinkdocker")
	if p.DisplayName != "Sunlink" || p.Group != "Lavoro" {
		t.Errorf("sunlinkdocker = %q / %q, want Sunlink / Lavoro", p.DisplayName, p.Group)
	}
	if q := find(t, s, "personalvpn"); q.DisplayName != "" || q.Group != "Casa" {
		t.Errorf("personalvpn = %q / %q, want no display name, Casa", q.DisplayName, q.Group)
	}
	if want := []string{"Casa", "Lavoro"}; !reflect.DeepEqual(s.Groups, want) {
		t.Errorf("Groups = %v, want %v (each once, sorted)", s.Groups, want)
	}
}

func TestUpdateSettings_AFieldLeftOutIsUntouched(t *testing.T) {
	m := refreshed(t, tempDB(t))
	ctx := context.Background()
	if err := m.UpdateSettings(ctx, "sunlinkdocker", compose.Settings{DisplayName: ptr("Sunlink"), Group: ptr("Lavoro")}); err != nil {
		t.Fatal(err)
	}
	if err := m.UpdateSettings(ctx, "sunlinkdocker", compose.Settings{Group: ptr("Casa")}); err != nil {
		t.Fatal(err)
	}
	if p := find(t, m.Snapshot(), "sunlinkdocker"); p.DisplayName != "Sunlink" || p.Group != "Casa" {
		t.Errorf("got %q / %q, want Sunlink / Casa", p.DisplayName, p.Group)
	}
}

func TestUpdateSettings_ClearingRestoresTheDefaults(t *testing.T) {
	path := tempDB(t)
	m := refreshed(t, path)
	ctx := context.Background()
	if err := m.UpdateSettings(ctx, "sunlinkdocker", compose.Settings{DisplayName: ptr("Sunlink"), Group: ptr("Lavoro")}); err != nil {
		t.Fatal(err)
	}

	// An empty or blank value clears the setting.
	if err := m.UpdateSettings(ctx, "sunlinkdocker", compose.Settings{DisplayName: ptr(""), Group: ptr("  ")}); err != nil {
		t.Fatal(err)
	}
	s := m.Snapshot()
	if p := find(t, s, "sunlinkdocker"); p.DisplayName != "" || p.Group != "" {
		t.Errorf("after clearing: %q / %q, want both empty", p.DisplayName, p.Group)
	}
	if len(s.Groups) != 0 {
		t.Errorf("Groups = %v, want none once the last Project left", s.Groups)
	}

	// Also for good: a new process sees no setting.
	again := refreshed(t, path)
	if p := find(t, again.Snapshot(), "sunlinkdocker"); p.DisplayName != "" || p.Group != "" {
		t.Errorf("after reopening: %q / %q, want both empty", p.DisplayName, p.Group)
	}
}

func TestUpdateSettings_ValuesAreTrimmed(t *testing.T) {
	m := refreshed(t, tempDB(t))
	if err := m.UpdateSettings(context.Background(), "sunlinkdocker", compose.Settings{DisplayName: ptr("  Sunlink "), Group: ptr(" Lavoro\t")}); err != nil {
		t.Fatal(err)
	}
	if p := find(t, m.Snapshot(), "sunlinkdocker"); p.DisplayName != "Sunlink" || p.Group != "Lavoro" {
		t.Errorf("got %q / %q, want trimmed values", p.DisplayName, p.Group)
	}
}

func TestUpdateSettings_SurviveReopeningTheDBAndRediscovery(t *testing.T) {
	path := tempDB(t)
	m := refreshed(t, path)
	if err := m.UpdateSettings(context.Background(), "sunlinkdocker", compose.Settings{DisplayName: ptr("Sunlink"), Group: ptr("Lavoro")}); err != nil {
		t.Fatal(err)
	}

	// Docker re-discovers the Project on every Refresh: it must not wipe them.
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p := find(t, m.Snapshot(), "sunlinkdocker"); p.DisplayName != "Sunlink" || p.Group != "Lavoro" {
		t.Errorf("after re-discovery: %q / %q", p.DisplayName, p.Group)
	}

	again := refreshed(t, path)
	if p := find(t, again.Snapshot(), "sunlinkdocker"); p.DisplayName != "Sunlink" || p.Group != "Lavoro" {
		t.Errorf("after reopening: %q / %q", p.DisplayName, p.Group)
	}
}

func TestUpdateSettings_SurviveADown(t *testing.T) {
	path := tempDB(t)
	first := refreshed(t, path)
	if err := first.UpdateSettings(context.Background(), "stellantisvpn", compose.Settings{DisplayName: ptr("VPN"), Group: ptr("Lavoro")}); err != nil {
		t.Fatal(err)
	}

	// Down: Docker no longer has any Container, every Project is Rimosso.
	m, _ := removedProjects(t, path)
	p := find(t, m.Snapshot(), "stellantisvpn")
	if p.Status != compose.StatusRemoved {
		t.Fatalf("status = %s, want removed", p.Status)
	}
	if p.DisplayName != "VPN" || p.Group != "Lavoro" {
		t.Errorf("after Down: %q / %q, want VPN / Lavoro", p.DisplayName, p.Group)
	}
}

func TestUpdateSettings_WorksWhileDockerIsUnreachable(t *testing.T) {
	f := realDocker(t)
	m := newManager(t, f, tempDB(t))
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	f.RespondErr(nil, errors.New("Cannot connect to the Docker daemon"), lsArgs...)
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := m.UpdateSettings(context.Background(), "sunlinkdocker", compose.Settings{Group: ptr("Lavoro")}); err != nil {
		t.Fatal(err)
	}
	s := m.Snapshot()
	if s.DockerReachable {
		t.Error("Docker should still be unreachable")
	}
	if p := find(t, s, "sunlinkdocker"); p.Group != "Lavoro" {
		t.Errorf("Group = %q, want Lavoro", p.Group)
	}
}

func TestUpdateSettings_RefusesAnUnknownProject(t *testing.T) {
	m := refreshed(t, tempDB(t))
	err := m.UpdateSettings(context.Background(), "nonexistent", compose.Settings{Group: ptr("X")})
	if !errors.Is(err, compose.ErrUnknownProject) {
		t.Errorf("err = %v, want ErrUnknownProject", err)
	}
}
