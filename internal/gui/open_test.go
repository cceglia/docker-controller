package gui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/guitest"
	"dockercontroller/internal/settings"
)

func TestPortURL_BuildsTheAddressFromTheNumber(t *testing.T) {
	for _, tc := range []struct {
		name string
		port int
		want string
		bad  bool
	}{
		{"lowest port", 1, "http://localhost:1", false},
		{"http port", 80, "http://localhost:80", false},
		{"development port", 8080, "http://localhost:8080", false},
		{"highest port", 65535, "http://localhost:65535", false},
		{"zero", 0, "", true},
		{"below zero", -1, "", true},
		{"above the range", 65536, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := PortURL(tc.port)
			if tc.bad {
				if !errors.Is(err, ErrInvalidPort) {
					t.Fatalf("PortURL(%d) error = %v, want ErrInvalidPort", tc.port, err)
				}
				if got != "" {
					t.Errorf("PortURL(%d) = %q, want no address", tc.port, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("PortURL(%d) error = %v", tc.port, err)
			}
			if got != tc.want {
				t.Errorf("PortURL(%d) = %q, want %q", tc.port, got, tc.want)
			}
		})
	}
}

func TestPortInRange_AcceptsOnlyARealPort(t *testing.T) {
	for _, tc := range []struct {
		port int
		want bool
	}{
		{1, true},
		{8080, true},
		{65535, true},
		{0, false},
		{-1, false},
		{65536, false},
	} {
		if got := PortInRange(tc.port); got != tc.want {
			t.Errorf("PortInRange(%d) = %t, want %t", tc.port, got, tc.want)
		}
	}
}

func TestOpenCommand_HandsOneArgumentToXdgOpen(t *testing.T) {
	for _, target := range []string{
		"http://localhost:8080",
		"/home/me/My Project",
	} {
		t.Run(target, func(t *testing.T) {
			cmd := openCommand(target)
			if len(cmd.Args) != 2 {
				t.Fatalf("Args = %q, want xdg-open and one target", cmd.Args)
			}
			if cmd.Args[0] != "xdg-open" {
				t.Errorf("Args[0] = %q, want %q", cmd.Args[0], "xdg-open")
			}
			if cmd.Args[1] != target {
				t.Errorf("Args[1] = %q, want %q", cmd.Args[1], target)
			}
		})
	}
}

func TestTerminalCommand_UsesTheConfiguredVector(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command string
		folder  string
		args    []string
		bad     bool
	}{
		{
			name:    "the default command",
			command: settings.DefaultTerminalCommand,
			folder:  "/home/me/shop",
			args:    []string{"konsole", "--workdir", "/home/me/shop"},
		},
		{
			name:    "a folder with spaces stays one argument",
			command: settings.DefaultTerminalCommand,
			folder:  "/home/me/My Project",
			args:    []string{"konsole", "--workdir", "/home/me/My Project"},
		},
		{
			name:    "a quoted command",
			command: `xterm -e "bash -l"`,
			folder:  "/home/me/shop",
			args:    []string{"xterm", "-e", "bash -l"},
		},
		{
			name:    "an unterminated quote is refused",
			command: `xterm -e "bash -l`,
			folder:  "/home/me/shop",
			bad:     true,
		},
		{
			name:    "an empty command is refused",
			command: "   ",
			folder:  "/home/me/shop",
			bad:     true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := terminalCommand(tc.command, tc.folder)
			if tc.bad {
				if !errors.Is(err, settings.ErrInvalidTerminalCommand) {
					t.Fatalf("terminalCommand error = %v, want ErrInvalidTerminalCommand", err)
				}
				if cmd != nil {
					t.Errorf("terminalCommand = %v, want no command", cmd)
				}
				return
			}
			if err != nil {
				t.Fatalf("terminalCommand error = %v", err)
			}
			guitest.AssertStrings(t, "Args", cmd.Args, tc.args)
			if cmd.Dir != tc.folder {
				t.Errorf("Dir = %q, want %q", cmd.Dir, tc.folder)
			}
		})
	}
}

func TestProjectFolder_FindsTheFolderOnDisk(t *testing.T) {
	dir := t.TempDir()
	gone := filepath.Join(dir, "gone")
	projects := []compose.Project{
		{Name: "shop", Folder: dir},
		{Name: "homeless", Folder: ""},
		{Name: "gone", Folder: gone},
	}

	for _, tc := range []struct {
		name    string
		project string
		want    string
		err     error
	}{
		{"an existing folder", "shop", dir, nil},
		{"a Project without a folder", "homeless", "", ErrNoFolder},
		{"a folder that is gone", "gone", "", ErrFolderMissing},
		{"an unknown Project", "blog", "", compose.ErrUnknownProject},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ProjectFolder(tc.project, projects)
			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Fatalf("ProjectFolder(%q) error = %v, want %v", tc.project, err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ProjectFolder(%q) error = %v", tc.project, err)
			}
			if got != tc.want {
				t.Errorf("ProjectFolder(%q) = %q, want %q", tc.project, got, tc.want)
			}
		})
	}
}

// TestFirstPublishedPort_FindsTheFirstPortInTheOrderOfTheProject checks the
// port "Porta" opens: the first published one, in the order the snapshot lists
// the Services, their Containers and their ports. A number that is not a port is
// not one the browser can reach, so it is skipped.
func TestFirstPublishedPort_FindsTheFirstPortInTheOrderOfTheProject(t *testing.T) {
	for _, tc := range []struct {
		name    string
		project compose.Project
		want    int
		ok      bool
	}{
		{"no Service", compose.Project{}, 0, false},
		{"no port", projectWithPorts(0), 0, false},
		{"one port", projectWithPorts(8080), 8080, true},
		{"two ports of one Container", projectWithPorts(8080, 9090), 8080, true},
		{
			name: "the Services keep their order",
			project: compose.Project{Services: []compose.Service{
				{Name: "a", Containers: []compose.Container{{Name: "a-1", Ports: []compose.Port{{Published: 3000, Target: 3000}}}}},
				{Name: "b", Containers: []compose.Container{{Name: "b-1", Ports: []compose.Port{{Published: 4000, Target: 4000}}}}},
			}},
			want: 3000,
			ok:   true,
		},
		{
			name: "a number that is not a port is skipped",
			project: compose.Project{Services: []compose.Service{{Containers: []compose.Container{{
				Ports: []compose.Port{{Published: 0, Target: 80}, {Published: 70000, Target: 80}, {Published: 8443, Target: 443}},
			}}}}},
			want: 8443,
			ok:   true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := FirstPublishedPort(tc.project)
			if ok != tc.ok {
				t.Fatalf("FirstPublishedPort found = %t, want %t", ok, tc.ok)
			}
			if got != tc.want {
				t.Errorf("FirstPublishedPort = %d, want %d", got, tc.want)
			}
		})
	}
}

// projectWithPorts is a Project of one Service with the published ports given.
func projectWithPorts(ports ...int) compose.Project {
	var list []compose.Port
	for _, port := range ports {
		list = append(list, compose.Port{Published: port, Target: port, Protocol: "tcp"})
	}
	return compose.Project{Services: []compose.Service{{Name: "web", Containers: []compose.Container{{Name: "web-1", Ports: list}}}}}
}

// TestCommandViews_SaysWhichCommandsTheProjectAllows checks the three commands
// of the detail and the reason of the ones that cannot be pressed.
func TestCommandViews_SaysWhichCommandsTheProjectAllows(t *testing.T) {
	complete := projectWithPorts(8080)
	complete.Folder = "/srv/shop"

	for _, tc := range []struct {
		name     string
		project  compose.Project
		folder   CommandView
		terminal CommandView
		port     CommandView
	}{
		{
			name:     "a complete Project opens everything",
			project:  complete,
			folder:   CommandView{Enabled: true},
			terminal: CommandView{Enabled: true},
			port:     CommandView{Enabled: true},
		},
		{
			name:     "a Project without a folder opens neither the folder nor a terminal",
			project:  projectWithPorts(8080),
			folder:   CommandView{Reason: NoFolderCommandText},
			terminal: CommandView{Reason: NoFolderCommandText},
			port:     CommandView{Enabled: true},
		},
		{
			name:     "a Project that publishes no port opens no browser",
			project:  compose.Project{Folder: "/srv/shop"},
			folder:   CommandView{Enabled: true},
			terminal: CommandView{Enabled: true},
			port:     CommandView{Reason: NoPortCommandText},
		},
		{
			name:     "a missing File compose closes every command",
			project:  compose.Project{Folder: "/srv/shop", ComposeMissing: true, Services: complete.Services},
			folder:   CommandView{Reason: MissingComposeText},
			terminal: CommandView{Reason: MissingComposeText},
			port:     CommandView{Reason: MissingComposeText},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			folder, terminal, port := CommandViews(tc.project)
			if folder != tc.folder {
				t.Errorf("folder = %+v, want %+v", folder, tc.folder)
			}
			if terminal != tc.terminal {
				t.Errorf("terminal = %+v, want %+v", terminal, tc.terminal)
			}
			if port != tc.port {
				t.Errorf("port = %+v, want %+v", port, tc.port)
			}
		})
	}
}

func TestProjectFolder_RefusesAFileInPlaceOfAFolder(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "docker-compose.yml")
	if err := os.WriteFile(file, []byte("services: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	projects := []compose.Project{{Name: "shop", Folder: file}}
	if _, err := ProjectFolder("shop", projects); !errors.Is(err, ErrFolderMissing) {
		t.Errorf("ProjectFolder error = %v, want ErrFolderMissing", err)
	}
}
