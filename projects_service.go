package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/wailsapp/wails/v3/pkg/application"

	"dockercontroller/internal/compose"
)

// ProjectsService is the thin Wails binding over the compose module. The
// frontend receives snapshots through the "snapshot" event; Snapshot only
// serves the initial load.
type ProjectsService struct {
	manager *compose.Manager
	// ctx bounds the Activities: they are stopped when the app quits.
	ctx context.Context
}

// Snapshot returns the current snapshot of the saved Projects.
func (s *ProjectsService) Snapshot() compose.Snapshot {
	return s.manager.Snapshot()
}

// Do starts an Action on a Project and returns the id of its Activity. Output
// and outcome arrive through the "activity" event.
func (s *ProjectsService) Do(action compose.Action) (compose.ActivityID, error) {
	return s.manager.Do(s.ctx, action)
}

// RemoveFromList removes a Rimosso Project from the saved list. It never
// touches Docker; a Project with Containers is refused.
func (s *ProjectsService) RemoveFromList(project string) error {
	return s.manager.RemoveFromList(s.ctx, project)
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
	return compose.ProposeName(mainFile)
}

// AddProject saves a Project added by hand: its compose name and its File
// compose in -f order (the main one first). A name already saved is refused.
func (s *ProjectsService) AddProject(name string, files []string) error {
	return s.manager.AddProject(s.ctx, name, files)
}

// OpenPort opens http://localhost:<port> in the default browser. The URL is
// built here from the number, never taken from the frontend.
func (s *ProjectsService) OpenPort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("porta non valida: %d", port)
	}
	return xdgOpen(fmt.Sprintf("http://localhost:%d", port))
}

// OpenFolder opens the folder of a saved Project in the file manager.
func (s *ProjectsService) OpenFolder(project string) error {
	for _, p := range s.manager.Snapshot().Projects {
		if p.Name != project {
			continue
		}
		if p.Folder == "" {
			return fmt.Errorf("il Progetto %s non ha una cartella", project)
		}
		if info, err := os.Stat(p.Folder); err != nil || !info.IsDir() {
			return fmt.Errorf("la cartella %s non esiste", p.Folder)
		}
		return xdgOpen(p.Folder)
	}
	return fmt.Errorf("Progetto sconosciuto: %s", project)
}

// xdgOpen hands a URL or path to the desktop's default handler.
func xdgOpen(target string) error {
	cmd := exec.Command("xdg-open", target)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("xdg-open: %w", err)
	}
	go func() { _ = cmd.Wait() }() // reap the process
	return nil
}
