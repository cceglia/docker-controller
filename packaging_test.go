package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The values below come from the spec, not from the build files, so a drift
// between the application and its packaging makes a test fail.
const (
	packagingAppID        = "io.github.cceglia.DockerController"
	packagingBinary       = "docker-controller"
	packagingEntry        = "build/linux/docker-controller.desktop.in"
	packagingReadme       = "README.md"
	packagingMakefile     = "Makefile"
	packagingWindowConsts = "internal/gui/windowconsts.go"
)

// entryValue returns the value of a key of a desktop entry, "" when the entry
// has no such key.
func entryValue(body, key string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, key+"=") {
			continue
		}
		return strings.TrimPrefix(line, key+"=")
	}
	return ""
}

// declaredString returns the value of the string constant name declared in a
// Go source file. The packaging test reads the constant instead of importing
// it, because it lives in another package.
func declaredString(t *testing.T, path, name string) string {
	t.Helper()
	pattern := regexp.MustCompile(`(?m)^\s*(?:const\s+)?` + regexp.QuoteMeta(name) + `\s*=\s*"([^"]*)"`)
	match := pattern.FindStringSubmatch(readTestFile(t, path))
	if match == nil {
		t.Fatalf("%s declares no string constant %s", path, name)
	}
	return match[1]
}

// hasCommentMentioning reports whether body has a comment line that names word.
func hasCommentMentioning(body, word string) bool {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") && strings.Contains(line, word) {
			return true
		}
	}
	return false
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(body)
}

// TestDesktopEntryNamesTheApplication checks that the installed entry carries
// the application id as icon name and as window class, so the launcher, the
// icon and the notifications of GApplication point to the same name.
func TestDesktopEntryNamesTheApplication(t *testing.T) {
	body := readTestFile(t, packagingEntry)
	for _, want := range []struct{ key, value string }{
		{"Type", "Application"},
		{"Name", "Docker Controller"},
		{"Exec", "@BINDIR@/" + packagingBinary},
		{"Icon", packagingAppID},
		{"Terminal", "false"},
		{"Categories", "Development;"},
		{"StartupWMClass", packagingAppID},
	} {
		if got := entryValue(body, want.key); got != want.value {
			t.Errorf("%s = %q, want %q", want.key, got, want.value)
		}
	}
}

// TestApplicationNameMatchesStartupWMClass checks that the name the application
// declares to Qt is the class the desktop entry starts with. Without the match
// KDE shows the wrong icon and does not group the window under the menu entry.
func TestApplicationNameMatchesStartupWMClass(t *testing.T) {
	declared := declaredString(t, packagingWindowConsts, "appID")
	class := entryValue(readTestFile(t, packagingEntry), "StartupWMClass")
	if declared != class {
		t.Errorf("application name %q, StartupWMClass %q: they must be equal", declared, class)
	}
	if declared != packagingAppID {
		t.Errorf("application name %q, want %q", declared, packagingAppID)
	}
}

// TestReadmeNamesTheQtPrerequisites checks that the build documentation asks for
// the packages of a Qt 6 build, explains the C++ shim and the cost of the first
// build, and names none of the GTK packages the project dropped.
func TestReadmeNamesTheQtPrerequisites(t *testing.T) {
	body := readTestFile(t, packagingReadme)
	lower := strings.ToLower(body)
	for _, want := range []string{"build-essential", "pkg-config", "qt6-base-dev", "c++", "quattro minuti"} {
		if !strings.Contains(lower, want) {
			t.Errorf("README does not name %q", want)
		}
	}
	for _, banned := range []string{"gtk", "adwaita", "gobject-introspection", "gotk4"} {
		if strings.Contains(lower, banned) {
			t.Errorf("README still names %q", banned)
		}
	}
}

// TestMakefileBuildsWithQtAndCgo checks that the build uses cgo and Qt, that no
// GTK package is a dependency, and that the icon cache tool stays documented.
func TestMakefileBuildsWithQtAndCgo(t *testing.T) {
	body := readTestFile(t, packagingMakefile)
	for _, want := range []string{"CGO_ENABLED=1", "gtk-update-icon-cache"} {
		if !strings.Contains(body, want) {
			t.Errorf("Makefile does not contain %q", want)
		}
	}
	// The icon cache tool is a program of the desktop, not a library of the
	// build: a comment explains that, and the recipe tolerates its absence.
	if !hasCommentMentioning(body, "gtk-update-icon-cache") {
		t.Error("Makefile does not explain gtk-update-icon-cache in a comment")
	}
	for _, banned := range []string{"libgtk", "libadwaita", "gobject-introspection", "gotk4"} {
		if strings.Contains(strings.ToLower(body), banned) {
			t.Errorf("Makefile still names %q", banned)
		}
	}
}

// TestDesktopIconsExist checks that both hicolor sizes have an icon to install
// under the application id.
func TestDesktopIconsExist(t *testing.T) {
	for _, size := range []string{"256x256", "512x512"} {
		path := filepath.Join("build/linux/icons", size, "docker-controller.png")
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if info.Size() == 0 {
			t.Errorf("%s is empty", path)
		}
	}
}

// TestUninstallRemovesInstalledAndLegacyFiles runs the real uninstall recipe on
// a temporary prefix: the files of this version and the entry and icon of the
// previous version must go, the database must stay.
func TestUninstallRemovesInstalledAndLegacyFiles(t *testing.T) {
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make is not installed")
	}
	prefix := t.TempDir()
	files := []string{
		filepath.Join(prefix, "bin", packagingBinary),
		filepath.Join(prefix, "share/applications", packagingAppID+".desktop"),
		filepath.Join(prefix, "share/applications", "docker-controller.desktop"),
		filepath.Join(prefix, "share/icons/hicolor/256x256/apps", packagingAppID+".png"),
		filepath.Join(prefix, "share/icons/hicolor/512x512/apps", packagingAppID+".png"),
		filepath.Join(prefix, "share/icons/hicolor/256x256/apps", "docker-controller.png"),
		filepath.Join(prefix, "share/icons/hicolor/512x512/apps", "docker-controller.png"),
	}
	for _, path := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte("fake"), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	database := filepath.Join(prefix, "share", "docker-controller", "docker-controller.db")
	if err := os.MkdirAll(filepath.Dir(database), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(database), err)
	}
	if err := os.WriteFile(database, []byte("projects"), 0o644); err != nil {
		t.Fatalf("write %s: %v", database, err)
	}

	cmd := exec.Command("make", "uninstall", "PREFIX="+prefix)
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("make uninstall: %v\n%s", err, out)
	}

	for _, path := range files {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s still exists after uninstall", path)
		}
	}
	if _, err := os.Stat(database); err != nil {
		t.Errorf("uninstall removed the database: %v", err)
	}
}
