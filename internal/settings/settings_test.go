package settings_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"dockercontroller/internal/db"
	"dockercontroller/internal/settings"
)

func openStore(t *testing.T, path string) *settings.Store {
	t.Helper()
	conn, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return settings.New(conn)
}

func tempDB(t *testing.T) string { return filepath.Join(t.TempDir(), "test.db") }

func TestStore_Get_ReturnsTheDefaultsOnAFreshDatabase(t *testing.T) {
	got, err := openStore(t, tempDB(t)).Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := settings.Settings{CloseBehavior: settings.CloseHide, TerminalCommand: "konsole --workdir {dir}"}
	if got != want {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestStore_Save_PersistsAcrossReopening(t *testing.T) {
	ctx := context.Background()
	path := tempDB(t)
	want := settings.Settings{CloseBehavior: settings.CloseAsk, TerminalCommand: "xterm -e 'cd {dir}; bash'"}
	if err := openStore(t, path).Save(ctx, want); err != nil {
		t.Fatal(err)
	}

	got, err := openStore(t, path).Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("after reopening Get() = %+v, want %+v", got, want)
	}
}

func TestStore_Save_OverwritesAPreviousValue(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, tempDB(t))
	for _, b := range []settings.CloseBehavior{settings.CloseQuit, settings.CloseHide} {
		if err := s.Save(ctx, settings.Settings{CloseBehavior: b, TerminalCommand: "konsole --workdir {dir}"}); err != nil {
			t.Fatal(err)
		}
		got, err := s.Get(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if got.CloseBehavior != b {
			t.Errorf("CloseBehavior = %q, want %q", got.CloseBehavior, b)
		}
	}
}

func TestStore_Save_RefusesAnUnknownCloseBehaviorAndKeepsTheOldSettings(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, tempDB(t))
	good := settings.Settings{CloseBehavior: settings.CloseQuit, TerminalCommand: "xterm"}
	if err := s.Save(ctx, good); err != nil {
		t.Fatal(err)
	}

	err := s.Save(ctx, settings.Settings{CloseBehavior: "boh", TerminalCommand: "other"})
	if !errors.Is(err, settings.ErrInvalidCloseBehavior) {
		t.Fatalf("Save error = %v, want ErrInvalidCloseBehavior", err)
	}
	got, err := s.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != good {
		t.Errorf("Get() = %+v, want unchanged %+v", got, good)
	}
}

func TestStore_Save_BlankTerminalCommandRestoresTheDefault(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, tempDB(t))
	if err := s.Save(ctx, settings.Settings{CloseBehavior: settings.CloseHide, TerminalCommand: "xterm"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(ctx, settings.Settings{CloseBehavior: settings.CloseHide, TerminalCommand: "   "}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.TerminalCommand != settings.DefaultTerminalCommand {
		t.Errorf("TerminalCommand = %q, want the default", got.TerminalCommand)
	}
}

func TestStore_Save_RefusesATerminalCommandThatCannotBeSplit(t *testing.T) {
	err := openStore(t, tempDB(t)).Save(context.Background(), settings.Settings{CloseBehavior: settings.CloseHide, TerminalCommand: `xterm "unclosed`})
	if err == nil {
		t.Fatal("Save accepted an unterminated quote")
	}
}
