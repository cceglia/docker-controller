package gui

import (
	"fmt"

	"github.com/mappu/miqt/qt6/mainthread"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/view"
)

// showAddDialog opens the picker of the File compose of a new Project, then the
// form with the files the user chose. A picker the user closed opens nothing, and
// a second click on "Aggiungi progetto" opens no second form.
func (s *shell) showAddDialog() {
	if s.addForm != nil || s.app.manager == nil {
		return
	}
	pickComposeFiles(s.window.QWidget, func(paths []string) {
		var form AddProjectForm
		added, err := form.Pick(paths, compose.ProposeName)
		if !added {
			return
		}
		dialog := s.openAddForm(form)
		if dialog != nil && err != nil {
			dialog.showError(view.UserError(err).Error())
		}
	})
}

// openAddForm shows the form of a new Project with form, once.
func (s *shell) openAddForm(form AddProjectForm) *addDialog {
	if s.addForm != nil {
		return nil
	}
	dialog := newAddDialog(s.window.QWidget, form, s.addProject, s.selectProject)
	s.addForm = dialog
	dialog.onClosed = func() {
		if s.addForm == dialog {
			s.addForm = nil
		}
	}
	dialog.open()
	return dialog
}

// addProject saves a Project added by hand, off the Qt thread: the database write
// must not freeze the window. The window refuses what it already knows, and the
// Manager judges the files and the saved list. done runs on the Qt thread with
// the refusal, nil when the Project is saved.
func (s *shell) addProject(name string, files []string, done func(error)) {
	if err := AddRefusal(name, files, s.snapshot.Projects); err != nil {
		done(err)
		return
	}
	manager := s.app.manager
	if manager == nil {
		done(nil)
		return
	}
	go func() {
		err := manager.AddProject(s.app.ctx, name, files)
		mainthread.Start(func() { done(err) })
	}()
}

// selectProject opens the detail of the Project just added. The snapshot that
// follows carries its row; until then the selection waits for the snapshot.
func (s *shell) selectProject(name string) {
	s.selected = name
	if project, ok := projectOf(s.snapshot.Projects, name); ok {
		s.open(project)
		return
	}
	if s.narrow {
		s.showNarrowPage()
	}
}

// openFolder opens the folder of a saved Project in the file manager.
func (s *shell) openFolder(name string) {
	folder, err := view.ProjectFolder(name, s.snapshot.Projects)
	if err != nil {
		s.showNotice(view.UserError(err).Error())
		return
	}
	s.openExternal(folder)
}

// openTerminal opens a terminal in the folder of a saved Project with the command
// of the settings. The settings come from the database, so the command is built
// off the Qt thread and a refusal comes back as the Italian message.
func (s *shell) openTerminal(name string) {
	folder, err := view.ProjectFolder(name, s.snapshot.Projects)
	if err != nil {
		s.showNotice(view.UserError(err).Error())
		return
	}
	store := s.app.store
	if store == nil {
		return
	}
	go func() {
		cfg, err := store.Get(s.app.ctx)
		if err != nil {
			mainthread.Start(func() { s.showNotice(view.UserError(err).Error()) })
			return
		}
		cmd, err := view.TerminalCommand(cfg.TerminalCommand, folder)
		if err != nil {
			mainthread.Start(func() { s.showNotice(view.UserError(err).Error()) })
			return
		}
		if err := cmd.Start(); err != nil {
			program := cmd.Args[0]
			mainthread.Start(func() {
				s.showNotice(view.UserError(fmt.Errorf("%w: %s: %w", view.ErrTerminalStart, program, err)).Error())
			})
			return
		}
		_ = cmd.Wait() // reap the process, so no zombie stays behind
	}()
}

// openPort opens a published port of the Project in the default browser. The
// address is built from the number, so a widget never supplies it.
func (s *shell) openPort(port int) {
	address, err := view.PortURL(port)
	if err != nil {
		s.showNotice(view.UserError(err).Error())
		return
	}
	s.openExternal(address)
}

// openExternal hands a path or an address to the desktop, off the Qt thread. A
// failure comes back on the Qt thread as the Italian message of the notice.
func (s *shell) openExternal(target string) {
	go func() {
		cmd := view.OpenCommand(target)
		if err := cmd.Start(); err != nil {
			mainthread.Start(func() {
				s.showNotice(view.UserError(fmt.Errorf("%w: %w", view.ErrCannotOpen, err)).Error())
			})
			return
		}
		_ = cmd.Wait() // reap the process, so no zombie stays behind
	}()
}

// removeFromList takes a Project off the saved list, off the Qt thread. It stops
// nothing in Docker: the Manager asks Docker again and refuses a Project that
// still has a Container. The snapshot that follows drops the row and the detail.
func (s *shell) removeFromList(name string) {
	manager := s.app.manager
	if manager == nil {
		return
	}
	go func() {
		if err := manager.RemoveFromList(s.app.ctx, name); err != nil {
			mainthread.Start(func() { s.showNotice(view.UserError(err).Error()) })
		}
	}()
}
