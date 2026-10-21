package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/wailsapp/wails/v3/pkg/application"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/settings"
)

// ProjectsService is the thin Wails binding over the compose module. The
// frontend receives snapshots through the "snapshot" event; Snapshot only
// serves the initial load.
type ProjectsService struct {
	manager *compose.Manager
	// settings holds the terminal command of "Apri terminale".
	settings *settings.Store
	// ctx bounds the Activities: they are stopped when the app quits.
	ctx context.Context
	// emit sends an event to the frontend.
	emit func(name string, data any)
	// logs maps the open log streams to their cancellation.
	logs logStreams
}

// Snapshot returns the current snapshot of the saved Projects.
func (s *ProjectsService) Snapshot() compose.Snapshot {
	return s.manager.Snapshot()
}

// Do starts an Action on a Project and returns the id of its Activity. Output
// and outcome arrive through the "activity" event.
func (s *ProjectsService) Do(action compose.Action) (compose.ActivityID, error) {
	id, err := s.manager.Do(s.ctx, action)
	return id, userError(err)
}

// RemoveFromList removes a Rimosso Project from the saved list. It never
// touches Docker; a Project with Containers is refused.
func (s *ProjectsService) RemoveFromList(project string) error {
	return userError(s.manager.RemoveFromList(s.ctx, project))
}

// UpdateSettings changes the Display name, Group and/or active Profiles of a
// Project: a null field is left as it is, an empty one is cleared. It never
// launches an Action.
func (s *ProjectsService) UpdateSettings(project string, settings compose.Settings) error {
	return userError(s.manager.UpdateSettings(s.ctx, project, settings))
}

// PickComposeFile shows the native file picker and returns the chosen compose
// file, or "" when the user cancels.
func (s *ProjectsService) PickComposeFile() (string, error) {
	return application.Get().Dialog.OpenFile().
		SetTitle("Scegli un File compose").
		AddFilter("File compose (*.yml, *.yaml)", "*.yml;*.yaml").
		PromptForSingleSelection()
}

// ProposeName returns the compose name compose would give to a Project whose
// main File compose is mainFile.
func (s *ProjectsService) ProposeName(mainFile string) (string, error) {
	name, err := compose.ProposeName(mainFile)
	return name, userError(err)
}

// AddProject saves a Project added by hand: its compose name and its File
// compose in -f order (the main one first). A name already saved is refused.
func (s *ProjectsService) AddProject(name string, files []string) error {
	return userError(s.manager.AddProject(s.ctx, name, files))
}

// OpenPort opens http://localhost:<port> in the default browser. The URL is
// built here from the number, never taken from the frontend.
func (s *ProjectsService) OpenPort(port int) error {
	if port < 1 || port > 65535 {
		return userError(fmt.Errorf("%w: %d", errInvalidPort, port))
	}
	return userError(xdgOpen(fmt.Sprintf("http://localhost:%d", port)))
}

// OpenFolder opens the folder of a saved Project in the file manager.
func (s *ProjectsService) OpenFolder(project string) error {
	folder, err := s.projectFolder(project)
	if err != nil {
		return userError(err)
	}
	return userError(xdgOpen(folder))
}

// OpenTerminal opens a terminal in the folder of a saved Project, running the
// terminal command of the settings with {dir} replaced by the folder. No shell
// is involved: the command is split into arguments before the substitution.
func (s *ProjectsService) OpenTerminal(project string) error {
	folder, err := s.projectFolder(project)
	if err != nil {
		return userError(err)
	}
	cfg, err := s.settings.Get(s.ctx)
	if err != nil {
		return userError(err)
	}
	argv, err := settings.TerminalArgv(cfg.TerminalCommand, folder)
	if err != nil {
		return userError(err)
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = folder
	if err := cmd.Start(); err != nil {
		return userError(fmt.Errorf("%w: %s: %w", errTerminalStart, argv[0], err))
	}
	go func() { _ = cmd.Wait() }() // reap the process
	return nil
}

// projectFolder is the existing folder of a saved Project.
func (s *ProjectsService) projectFolder(project string) (string, error) {
	for _, p := range s.manager.Snapshot().Projects {
		if p.Name != project {
			continue
		}
		if p.Folder == "" {
			return "", fmt.Errorf("%w: %s", errNoFolder, project)
		}
		if info, err := os.Stat(p.Folder); err != nil || !info.IsDir() {
			return "", fmt.Errorf("%w: %s", errFolderMissing, p.Folder)
		}
		return p.Folder, nil
	}
	return "", fmt.Errorf("%w: %q", compose.ErrUnknownProject, project)
}

// xdgOpen hands a URL or path to the desktop's default handler.
func xdgOpen(target string) error {
	cmd := exec.Command("xdg-open", target)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%w: %w", errCannotOpen, err)
	}
	go func() { _ = cmd.Wait() }() // reap the process
	return nil
}
