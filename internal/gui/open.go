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
