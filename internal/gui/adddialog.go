package gui

import (
	"path/filepath"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"dockercontroller/internal/compose"
)

const (
	// addProjectTitle is the title of the dialog of a new Project.
	addProjectTitle = "Aggiungi Progetto"
	// addDialogWidth is the width the dialog asks for, so a long path still
	// reads without widening it.
	addDialogWidth = 520
	// addFilesTitle is the heading of the chosen File compose; addFilesHint
	// invites the user to choose the main one.
	addFilesTitle = "File compose"
	addFilesHint  = "Scegli il File compose principale del Progetto."
	// addFirstFileText invites the user to choose the main File compose;
	// addMoreFilesText asks for another one.
	addFirstFileText = "Scegli il File compose…"
	addMoreFilesText = "Aggiungi un altro File compose…"
	// addNameTitle is the heading of the compose name; addNameTitleRow is the
	// label of its field and addNamePlaceholder the example in it.
	addNameTitle       = "Nome del Progetto"
	addNameTitleRow    = "Nome compose"
	addNamePlaceholder = "nome-compose"
	addText            = "Aggiungi"
	// mainFileRoleText and extraFileRoleText tell the main File compose from the
	// others, as the old view did.
	mainFileRoleText  = "principale"
	extraFileRoleText = "aggiuntivo"
)

// addDialog is the modal dialog of a new Project: the chosen File compose, the
// name compose proposes and the two answers. It holds the pending form only, so
// Annulla leaves nothing behind.
type addDialog struct {
	shell  *shell
	form   AddProjectForm
	dialog *adw.Dialog
	// files holds the block the chosen files rebuild; entry is the name field,
	// error the message of a refusal and add the confirm button.
	files *gtk.Box
	entry *gtk.Entry
	error *gtk.Label
	add   *gtk.Button
	// busy keeps the confirm button off while the Manager saves, and syncing
	// stops a proposed name from counting as a name the user typed.
	busy    bool
	syncing bool
}

// newAddDialog builds the dialog on the form the picker filled: the chosen File
// compose, and the name compose proposed for the main one. The user may change
// the name, and once changed the form keeps it.
func newAddDialog(s *shell, form AddProjectForm) *addDialog {
	a := &addDialog{shell: s, form: form}

	a.dialog = adw.NewDialog()
	a.dialog.SetTitle(addProjectTitle)
	a.dialog.SetContentWidth(addDialogWidth)

	content := gtk.NewBox(gtk.OrientationVertical, 18)
	content.SetMarginTop(18)
	content.SetMarginBottom(18)
	content.SetMarginStart(18)
	content.SetMarginEnd(18)
	a.files = gtk.NewBox(gtk.OrientationVertical, 0)
	content.Append(a.files)

	a.entry = gtk.NewEntry()
	a.entry.SetPlaceholderText(addNamePlaceholder)
	a.entry.SetWidthChars(24)
	a.entry.SetHExpand(true)
	a.entry.SetTooltipText("Il nome compose del Progetto: minuscole, cifre, - e _.")
	a.entry.ConnectChanged(a.nameChanged)

	name := adw.NewPreferencesGroup()
	name.SetTitle(addNameTitle)
	row := adw.NewActionRow()
	row.SetTitle(addNameTitleRow)
	row.AddSuffix(a.entry)
	row.SetActivatableWidget(a.entry)
	name.Add(row)
	content.Append(name)

	a.error = gtk.NewLabel("")
	a.error.SetXAlign(0)
	a.error.SetWrap(true)
	a.error.AddCSSClass("error")
	a.error.SetVisible(false)
	content.Append(a.error)

	buttons := gtk.NewBox(gtk.OrientationHorizontal, 6)
	buttons.SetHAlign(gtk.AlignEnd)

	cancel := gtk.NewButtonWithLabel(cancelText)
	cancel.AddCSSClass("flat")
	cancel.SetTooltipText("Chiudi senza aggiungere il Progetto.")
	cancel.ConnectClicked(func() { a.dialog.Close() })
	buttons.Append(cancel)

	a.add = gtk.NewButtonWithLabel(addText)
	a.add.AddCSSClass("suggested-action")
	a.add.SetSensitive(false)
	a.add.SetTooltipText("Aggiungi il Progetto alla lista.")
	a.add.ConnectClicked(a.submit)
	buttons.Append(a.add)

	content.Append(buttons)

	// The dialog has no bar of its own: the header names it and carries the
	// close button, as a GNOME dialog does.
	view := adw.NewToolbarView()
	bar := adw.NewHeaderBar()
	title := gtk.NewLabel(addProjectTitle)
	title.AddCSSClass("heading")
	bar.SetTitleWidget(title)
	view.AddTopBar(bar)
	view.SetContent(content)
	a.dialog.SetChild(view)
	a.refresh()
	a.showName()
	return a
}

// refresh rebuilds the block of the chosen files and follows them with the
// name field and the confirm button. It runs after every change of the files.
func (a *addDialog) refresh() {
	removeAll(a.files)

	group := adw.NewPreferencesGroup()
	group.SetTitle(addFilesTitle)
	if len(a.form.Files) == 0 {
		group.SetDescription(addFilesHint)
	}

	pick := gtk.NewButtonWithLabel(addFirstFileText)
	if len(a.form.Files) > 0 {
		pick.SetLabel(addMoreFilesText)
	}
	pick.AddCSSClass("flat")
	pick.SetTooltipText("Scegli uno o più File compose.")
	pick.ConnectClicked(a.pick)
	group.SetHeaderSuffix(pick)

	for i, file := range a.form.Files {
		group.Add(a.fileRow(i, file))
	}
	a.files.Append(group)

	a.entry.SetSensitive(len(a.form.Files) > 0)
	a.updateAdd()
}

// fileRow is one chosen File compose: its name, its whole path, its role and the
// button that drops it. removeFile learns which one by its place.
func (a *addDialog) fileRow(index int, file string) *adw.ActionRow {
	row := adw.NewActionRow()
	row.SetTitle(filepath.Base(file))
	row.SetSubtitle(file)
	row.SetSubtitleLines(1)
	row.SetTooltipText(file)

	role := extraFileRoleText
	if index == 0 {
		role = mainFileRoleText
	}
	row.AddSuffix(valueLabel(role, "dim-label"))

	remove := gtk.NewButtonFromIconName("list-remove-symbolic")
	remove.AddCSSClass("flat")
	remove.SetTooltipText("Rimuovi " + file)
	remove.SetVAlign(gtk.AlignCenter)
	remove.ConnectClicked(func() { a.removeFile(index) })
	row.AddSuffix(remove)
	return row
}

// pick shows the file picker and adds every chosen File compose. The picker
// accepts more than one file, so a Project with an override arrives in one go.
// A picker the user closed changes nothing.
func (a *addDialog) pick() {
	pickComposeFiles(a.shell.app.ctx, &a.shell.window.ApplicationWindow.Window, func(paths []string) {
		if _, err := a.form.Pick(paths, compose.ProposeName); err != nil {
			a.showError(UserError(err).Error())
		}
		a.refresh()
		a.showName()
	}, func(err error) {
		a.showError(UserError(err).Error())
	})
}

// removeFile drops a chosen File compose. When the main file changed, the new
// one proposes its name again, unless the user typed one.
func (a *addDialog) removeFile(index int) {
	before := a.form.MainFile()
	a.form.RemoveFile(index)
	if a.form.MainFile() != before {
		a.proposeMain()
	}
	a.refresh()
}

// proposeMain asks compose for the name of the main file and shows it. An error
// of the file stays on screen and leaves the name as it is.
func (a *addDialog) proposeMain() {
	main := a.form.MainFile()
	if main == "" {
		a.form.Propose("")
		a.showName()
		return
	}
	name, err := compose.ProposeName(main)
	if err != nil {
		a.showError(UserError(err).Error())
		return
	}
	a.form.Propose(name)
	a.showName()
}

// showName puts the name of the form in the field, without counting it as a
// name the user typed.
func (a *addDialog) showName() {
	if a.entry.Text() == a.form.Name {
		return
	}
	a.syncing = true
	a.entry.SetText(a.form.Name)
	a.syncing = false
	a.updateAdd()
}

// nameChanged stores the name the user typed. From now on the main file no
// longer replaces it.
func (a *addDialog) nameChanged() {
	if a.syncing {
		return
	}
	a.form.Rename(a.entry.Text())
	a.clearError()
	a.updateAdd()
}

// submit sends the form: a name already saved is refused with its message, and
// the Manager gets the rest of the refusals.
func (a *addDialog) submit() {
	if !a.form.CanAdd() || a.busy {
		return
	}
	name := a.form.CleanName()
	if NameTaken(name, a.shell.projects) {
		a.showError(UserError(compose.ErrDuplicateProject).Error())
		return
	}

	a.busy = true
	a.updateAdd()
	files := append([]string{}, a.form.Files...)
	a.shell.addProject(name, files, func(err error) {
		a.busy = false
		if err != nil {
			a.showError(UserError(err).Error())
			a.updateAdd()
			return
		}
		a.dialog.Close()
		a.shell.selectProject(name)
	})
}

// updateAdd follows the state of the form: the confirm button is on only with a
// File compose, a name and no save in flight.
func (a *addDialog) updateAdd() {
	a.add.SetSensitive(!a.busy && a.form.CanAdd())
}

// showError puts a message under the fields.
func (a *addDialog) showError(message string) {
	a.error.SetText(message)
	a.error.SetVisible(true)
}

// clearError hides the message of a refusal.
func (a *addDialog) clearError() {
	a.error.SetText("")
	a.error.SetVisible(false)
}
