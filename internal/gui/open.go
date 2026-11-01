package gui

import (
	"fmt"
	"os"
	"os/exec"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/settings"
)

// PortURL is the address the browser opens for a published port. The URL is
// built here from the number, never read from the text of a widget, and only a
// real port gets one.
func PortURL(port int) (string, error) {
	if !PortInRange(port) {
		return "", fmt.Errorf("%w: %d", ErrInvalidPort, port)
	}
	return fmt.Sprintf("http://localhost:%d", port), nil
}

// PortInRange reports whether a number is a port the browser can reach.
func PortInRange(port int) bool {
	return port >= 1 && port <= 65535
}

// The reasons of a command of the detail that cannot be pressed. They are the
// sentences the notice and the old view use, so one fact has one text.
const (
	// NoFolderCommandText explains "Cartella" and "Terminale" on a Project
	// without a folder.
	NoFolderCommandText = "Il Progetto non ha una cartella."
	// NoPortCommandText explains "Porta" on a Project that publishes no port.
	NoPortCommandText = "Il Progetto non pubblica porte."
)

// CommandView is one command of the detail: whether it can be pressed and, when
// it cannot, the reason the user reads in the tooltip and hears from the screen
// reader.
type CommandView struct {
	Enabled bool
	Reason  string
}

// CommandViews reports the three commands of the detail: "Cartella",
// "Terminale" and "Porta". A Project whose File compose are missing opens
// nothing, because the folder and the ports it would open no longer describe it.
// Without a folder the first two stay off, and without a published port the
// third does.
func CommandViews(project compose.Project) (folder, terminal, port CommandView) {
	if project.ComposeMissing {
		missing := CommandView{Reason: MissingComposeText}
		return missing, missing, missing
	}
	folder, terminal = CommandView{Enabled: true}, CommandView{Enabled: true}
	if project.Folder == "" {
		noFolder := CommandView{Reason: NoFolderCommandText}
		folder, terminal = noFolder, noFolder
	}
	port = CommandView{Enabled: true}
	if _, ok := FirstPublishedPort(project); !ok {
		port = CommandView{Reason: NoPortCommandText}
	}
	return folder, terminal, port
}

// FirstPublishedPort is the port "Porta" opens: the first published port of the
// Project, in the order the snapshot lists its Services, their Containers and
// their ports. A number outside the ports the browser can reach is skipped, so a
// stale row never builds a dead address.
func FirstPublishedPort(project compose.Project) (int, bool) {
	for _, service := range project.Services {
		for _, container := range service.Containers {
			for _, port := range container.Ports {
				if PortInRange(port.Published) {
					return port.Published, true
				}
			}
		}
	}
	return 0, false
}

// openCommand hands a path or an address to the desktop's default handler. The
// target is one argument, so the shell never reads a space or a quote inside it.
func openCommand(target string) *exec.Cmd {
	return exec.Command("xdg-open", target)
}

// terminalCommand builds the command of "Apri terminale": the argument vector
// of the saved terminal command, with the folder of the Project in place of
// {dir}. No shell is involved, so the command runs only if it splits. The
// command is built and not started, so its shape is what the tests check.
func terminalCommand(command, folder string) (*exec.Cmd, error) {
	argv, err := settings.TerminalArgv(command, folder)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = folder
	return cmd, nil
}

// ProjectFolder is the folder of a saved Project, when it is still a folder on
// disk. A Project the list does not hold, a Project without a folder and a
// folder that is gone are three different errors, so the user reads the reason.
func ProjectFolder(name string, projects []compose.Project) (string, error) {
	for _, project := range projects {
		if project.Name != name {
			continue
		}
		if project.Folder == "" {
			return "", fmt.Errorf("%w: %s", ErrNoFolder, name)
		}
		if info, err := os.Stat(project.Folder); err != nil || !info.IsDir() {
			return "", fmt.Errorf("%w: %s", ErrFolderMissing, project.Folder)
		}
		return project.Folder, nil
	}
	return "", fmt.Errorf("%w: %q", compose.ErrUnknownProject, name)
}
