package gui

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/glib/v2"

	"dockercontroller/internal/compose"
)

// addProject saves a Project added by hand and reports the outcome on the GTK
// loop: onDone runs with the refusal of the Manager, nil when the Project is
// saved. The write runs off the loop, so the window stays responsive.
func (s *shell) addProject(name string, files []string, onDone func(error)) {
	if s.app.manager == nil {
		return
	}
	go func() {
		err := s.app.manager.AddProject(s.app.ctx, name, files)
		glib.IdleAdd(func() { onDone(err) })
	}()
}

// selectProject opens the content side on a Project added by hand, so the
// snapshot that follows shows its detail. A snapshot that arrived first already
// holds the Project, and then the row is selected at once.
func (s *shell) selectProject(name string) {
	s.selected = name
	s.split.SetShowContent(true)
	for index, project := range s.rows {
		if project.Name != name {
			continue
		}
		if row := s.list.RowAtIndex(index); row != nil {
			s.list.SelectRow(row)
		}
	}
	for _, project := range s.projects {
		if project.Name == name {
			s.detail.set(project, s.reachable, s.loading, GroupChoices(project.Group, s.projects))
			return
		}
	}
}

// showAddDialog asks for the File compose of a new Project, then opens the modal
// form with them. The picker accepts more than one file, and a picker the user
// closed opens nothing.
func (s *shell) showAddDialog() {
	if s.app.manager == nil || s.add != nil {
		return
	}
	pickComposeFiles(s.app.ctx, &s.window.ApplicationWindow.Window, s.addPicked, func(err error) {
		s.toast(UserError(err).Error())
	})
}

// addPicked opens the form of a new Project on the files just chosen, with the
// name compose proposes for the main one. Files already chosen add nothing, so
// an empty pick opens nothing.
func (s *shell) addPicked(paths []string) {
	if s.app.manager == nil || s.add != nil {
		return
	}
	var form AddProjectForm
	added, err := form.Pick(paths, compose.ProposeName)
	if !added {
		return
	}
	dialog := s.openAddForm(form)
	if dialog != nil && err != nil {
		dialog.showError(UserError(err).Error())
	}
}

// openAddForm shows the modal form of a new Project with form, once.
func (s *shell) openAddForm(form AddProjectForm) *addDialog {
	if s.add != nil {
		return nil
	}
	dialog := newAddDialog(s, form)
	s.add = dialog
	dialog.dialog.ConnectClosed(func() {
		if s.add == dialog {
			s.add = nil
		}
	})
	dialog.dialog.Present(&s.window.ApplicationWindow.Window)
	return dialog
}

// openFolder opens the folder of a saved Project in the file manager.
func (s *shell) openFolder(name string) {
	folder, err := ProjectFolder(name, s.projects)
	if err != nil {
		s.toast(UserError(err).Error())
		return
	}
	s.openExternal(folder)
}

// openTerminal opens a terminal in the folder of a saved Project with the
// command of the settings. The settings come from the database, so the command
// is built off the GTK loop and the refusal comes back as the Italian message.
func (s *shell) openTerminal(name string) {
	folder, err := ProjectFolder(name, s.projects)
	if err != nil {
		s.toast(UserError(err).Error())
		return
	}
	if s.app.store == nil {
		return
	}
	go func() {
		cfg, err := s.app.store.Get(s.app.ctx)
		if err != nil {
			glib.IdleAdd(func() { s.toast(UserError(err).Error()) })
			return
		}
		cmd, err := terminalCommand(cfg.TerminalCommand, folder)
		if err != nil {
			glib.IdleAdd(func() { s.toast(UserError(err).Error()) })
			return
		}
		if err := cmd.Start(); err != nil {
			glib.IdleAdd(func() {
				s.toast(UserError(fmt.Errorf("%w: %s: %w", ErrTerminalStart, cmd.Args[0], err)).Error())
			})
			return
		}
		_ = cmd.Wait() // reap the process, so no zombie stays behind
	}()
}

// openPort opens http://localhost:<port> in the default browser. The address is
// built from the number, so a widget never supplies it.
func (s *shell) openPort(port int) {
	address, err := PortURL(port)
	if err != nil {
		s.toast(UserError(err).Error())
		return
	}
	s.openExternal(address)
}

// openExternal hands a path or an address to the desktop, off the GTK loop. A
// failure comes back as the Italian message.
func (s *shell) openExternal(target string) {
	go func() {
		cmd := openCommand(target)
		if err := cmd.Start(); err != nil {
			glib.IdleAdd(func() {
				s.toast(UserError(fmt.Errorf("%w: %w", ErrCannotOpen, err)).Error())
			})
			return
		}
		_ = cmd.Wait() // reap the process, so no zombie stays behind
	}()
}

// removeFromList takes a Project off the saved list. It stops nothing in
// Docker: the Manager asks Docker again and refuses a Project that still has a
// Container.
func (s *shell) removeFromList(name string) {
	if s.app.manager == nil {
		return
	}
	go func() {
		if err := s.app.manager.RemoveFromList(s.app.ctx, name); err != nil {
			glib.IdleAdd(func() { s.toast(UserError(err).Error()) })
		}
	}()
}
