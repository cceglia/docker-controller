package gui

import (
	"path/filepath"

	"github.com/mappu/miqt/qt6"

	"dockercontroller/internal/compose"
)

// The texts of the form of a new Project and of its picker.
const (
	// addProjectTitle is the title of the question and its accessible name.
	addProjectTitle = "Aggiungi Progetto"
	// addDialogWidth and addDialogMaxWidth bound the form: a long path stays on
	// one line and the sentence of a refusal still reads.
	addDialogWidth    = 520
	addDialogMaxWidth = 720
	// addFilesTitle names the block of the chosen File compose and addFilesHint
	// explains the order while the block is empty.
	addFilesTitle = "File compose"
	addFilesHint  = "Scegli uno o più File compose: il primo è quello principale."
	// pickFirstFileText invites the user to choose the File compose;
	// pickMoreFilesText asks for another one. pickFilesText is the title of the
	// picker and composeFilterText the name of its filter.
	pickFirstFileText  = "Scegli i File compose…"
	pickMoreFilesText  = "Aggiungi un altro File compose…"
	pickFilesText      = "Scegli i File compose"
	composeFilterText  = "File compose (*.yml *.yaml)"
	allFilesFilterText = "Tutti i file (*)"
	// mainFileRoleText and extraFileRoleText tell the main File compose from the
	// others, as the old view did.
	mainFileRoleText  = "principale"
	extraFileRoleText = "aggiuntivo"
	// removeFileText is the command that drops one chosen File compose.
	removeFileText = "Rimuovi"
	// addNameLabelText names the field of the compose name, addNamePlaceholder
	// the example in it and addNameHint what compose accepts.
	addNameLabelText   = "Nome compose"
	addNamePlaceholder = "nome-compose"
	addNameHint        = "Il nome compose del Progetto: minuscole, cifre, - e _."
	// addText confirms the form.
	addText = "Aggiungi"
)

// addDialog is the modal form of a new Project: the chosen File compose in -f
// order, the name compose proposes and the two answers. It holds the pending
// form only, so Annulla leaves nothing behind: nothing reaches the Manager until
// Aggiungi.
type addDialog struct {
	dialog *qt6.QDialog
	form   AddProjectForm
	// list holds the rows of the chosen File compose, which refresh builds
	// again; drops are the commands of those rows. name is the field of the
	// compose name and error the message of a refusal.
	list  *qt6.QWidget
	drops []*qt6.QPushButton
	name  *qt6.QLineEdit
	error *qt6.QLabel
	// pick, cancel and add are the commands of the form.
	pick   *qt6.QPushButton
	cancel *qt6.QPushButton
	add    *qt6.QPushButton
	// busy keeps the commands off while the Manager saves, and syncing stops a
	// proposed name from counting as a name the user typed.
	busy    bool
	syncing bool
	// onSubmit hands the request to the shell, which reaches the Manager off the
	// Qt thread; done reports the refusal, nil when the Project is saved.
	onSubmit func(name string, files []string, done func(error))
	// onSaved runs after a Project is saved, so the shell opens its detail.
	onSaved func(name string)
	// onClosed runs after the dialog closes, whatever the answer.
	onClosed func()
}

// newAddDialog builds the form on the File compose the picker returned. Nothing
// is saved until Aggiungi, and a refusal keeps the form open with what the user
// chose.
func newAddDialog(parent *qt6.QWidget, form AddProjectForm, onSubmit func(string, []string, func(error)), onSaved func(string)) *addDialog {
	d := &addDialog{form: form, onSubmit: onSubmit, onSaved: onSaved}

	d.dialog = qt6.NewQDialog(parent)
	d.dialog.SetWindowTitle(addProjectTitle)
	d.dialog.SetAccessibleName(addProjectTitle)
	d.dialog.SetModal(true)
	d.dialog.SetMinimumWidth(addDialogWidth)
	d.dialog.SetMaximumWidth(addDialogMaxWidth)

	layout := qt6.NewQVBoxLayout(d.dialog.QWidget)
	layout.SetContentsMargins(spaceL, spaceL, spaceL, spaceL)
	layout.SetSpacing(spaceM)

	title := detailLabel(addProjectTitle, "dialogTitle")
	layout.AddWidget(title.QWidget)

	// The chosen File compose, in the order the picker returned them.
	files := qt6.NewQWidget2()
	box := qt6.NewQVBoxLayout(files)
	box.SetContentsMargins(0, 0, 0, 0)
	box.SetSpacing(spaceXS)
	box.AddWidget(detailLabel(addFilesTitle, "detailHeading").QWidget)

	d.list = qt6.NewQWidget2()
	rows := qt6.NewQVBoxLayout(d.list)
	rows.SetContentsMargins(0, 0, 0, 0)
	rows.SetSpacing(spaceXS)
	box.AddWidget(d.list)

	d.pick = detailButton(pickFirstFileText, "pickFilesButton", "Scegli uno o più File compose.", true)
	d.pick.OnClicked(d.pickFiles)
	box.AddWidget3(d.pick.QWidget, 0, qt6.AlignLeft)
	layout.AddWidget(files)

	// The compose name: the proposal of the main file, which the user may
	// change. The message of a refusal sits under the field.
	field := qt6.NewQWidget2()
	fieldBox := qt6.NewQVBoxLayout(field)
	fieldBox.SetContentsMargins(0, 0, 0, 0)
	fieldBox.SetSpacing(spaceXS)

	caption := detailLabel(addNameLabelText, "fieldLabel")
	d.name = qt6.NewQLineEdit2()
	setObjectName(d.name.QWidget, "addName")
	d.name.SetPlaceholderText(addNamePlaceholder)
	d.name.SetClearButtonEnabled(true)
	caption.SetBuddy(d.name.QWidget)
	keepTarget(d.name.QWidget)
	d.name.OnTextChanged(d.nameChanged)
	fieldBox.AddWidget(caption.QWidget)
	fieldBox.AddWidget(d.name.QWidget)

	d.error = detailLabel("", "fieldError")
	d.error.SetVisible(false)
	fieldBox.AddWidget(d.error.QWidget)
	fieldBox.AddWidget(detailLabel(addNameHint, "").QWidget)
	layout.AddWidget(field)

	// The two answers: Annulla closes the form, Aggiungi saves it. Aggiungi is
	// the default of the dialog, so Enter confirms it while the form is valid.
	answers := qt6.NewQDialogButtonBox2()

	d.cancel = answerButton(answers, cancelText, "addCancel", "Chiudi senza aggiungere il Progetto.", qt6.QDialogButtonBox__RejectRole)
	d.cancel.OnClicked(d.close)

	d.add = answerButton(answers, addText, "primaryAction", "Aggiungi il Progetto alla lista.", qt6.QDialogButtonBox__AcceptRole)
	d.add.SetEnabled(false)
	d.add.SetDefault(true)
	d.add.OnClicked(d.submit)
	layout.AddWidget(answers.QWidget)

	d.dialog.OnFinished(func(int) {
		// The form is over: it goes away with the event loop, and the shell that
		// opened it forgets it.
		d.dialog.DeleteLater()
		if d.onClosed != nil {
			d.onClosed()
		}
	})

	// Return saves the form from any field and Escape closes it, as a modal form
	// of this window does.
	for _, key := range []string{"Return", "Enter"} {
		save := qt6.NewQShortcut2(qt6.NewQKeySequence2(key), d.dialog.QObject)
		save.SetContext(qt6.WindowShortcut)
		save.OnActivated(d.submit)
	}
	escape := qt6.NewQShortcut2(qt6.NewQKeySequence2("Esc"), d.dialog.QObject)
	escape.SetContext(qt6.WindowShortcut)
	escape.OnActivated(d.close)

	d.refresh()
	d.showName()
	return d
}

// open shows the modal form. The name field takes the focus when a File compose
// is already chosen, otherwise the picker does, so the keyboard starts from the
// first thing to do.
func (d *addDialog) open() {
	d.dialog.AdjustSize()
	d.dialog.Open()
	if len(d.form.Files) > 0 {
		d.name.SetFocus()
		return
	}
	d.pick.SetFocus()
}

// close closes the form without saving. A save closes it on its own.
func (d *addDialog) close() {
	d.dialog.Reject()
}

// pickFiles shows the picker and adds every chosen File compose after the ones
// already chosen: the order of the choice is the -f order. A picker the user
// closed changes nothing.
func (d *addDialog) pickFiles() {
	pickComposeFiles(d.dialog.QWidget, func(paths []string) {
		added, err := d.form.Pick(paths, compose.ProposeName)
		if !added {
			return
		}
		d.refresh()
		d.showName()
		if err != nil {
			d.showError(UserError(err).Error())
		}
	})
}

// refresh rebuilds the rows of the chosen File compose and follows them with the
// state of the form: the main file keeps the first place, and the name field
// stays insensitive until a File compose is chosen.
func (d *addDialog) refresh() {
	clearDetailLayout(d.list.Layout())
	d.drops = d.drops[:0]

	if len(d.form.Files) == 0 {
		hint := detailLabel(addFilesHint, "")
		d.list.Layout().AddWidget(hint.QWidget)
	} else {
		for i, file := range d.form.Files {
			d.list.Layout().AddWidget(d.fileRow(i, file))
		}
	}

	d.pick.SetText(pickFirstFileText)
	if len(d.form.Files) > 0 {
		d.pick.SetText(pickMoreFilesText)
	}
	d.name.SetEnabled(len(d.form.Files) > 0)
	d.updateAdd()
	d.linkTabOrder()
}

// fileRow is one chosen File compose: its name, its role and the command that
// drops it. The whole path sits in the tooltip, so the row stays on one line.
func (d *addDialog) fileRow(index int, file string) *qt6.QWidget {
	row := qt6.NewQWidget2()

	line := qt6.NewQHBoxLayout(row)
	line.SetContentsMargins(0, 0, 0, 0)
	line.SetSpacing(spaceS)

	name := detailLabel(filepath.Base(file), "fileComposeName")
	name.SetToolTip(file)
	line.AddWidget3(name.QWidget, 0, qt6.AlignVCenter)

	role := extraFileRoleText
	if index == 0 {
		role = mainFileRoleText
	}
	line.AddWidget3(detailLabel(role, "fileComposeRole").QWidget, 0, qt6.AlignVCenter)
	line.AddStretch()

	drop := detailButton(removeFileText, "removeFileButton", "Toglie "+file+" dalla scelta.", true)
	drop.SetAccessibleName(removeFileText + " " + filepath.Base(file))
	drop.OnClicked(func() { d.removeFile(index) })
	line.AddWidget3(drop.QWidget, 0, qt6.AlignVCenter)
	d.drops = append(d.drops, drop)

	return row
}

// removeFile drops a chosen File compose. When the main file changed, the new
// one proposes its name again, unless the user typed one.
func (d *addDialog) removeFile(index int) {
	before := d.form.MainFile()
	d.form.RemoveFile(index)
	if d.form.MainFile() != before {
		d.proposeMain()
	}
	d.refresh()
	d.showName()
}

// proposeMain asks compose for the name of the main file and shows it. An error
// of the file stays on screen and leaves the name as it is.
func (d *addDialog) proposeMain() {
	main := d.form.MainFile()
	if main == "" {
		d.form.Propose("")
		return
	}
	name, err := compose.ProposeName(main)
	if err != nil {
		d.showError(UserError(err).Error())
		return
	}
	d.form.Propose(name)
}

// showName puts the name of the form in the field, without counting it as a name
// the user typed.
func (d *addDialog) showName() {
	if d.syncing || d.name.Text() == d.form.Name {
		return
	}
	d.syncing = true
	d.name.SetText(d.form.Name)
	d.syncing = false
	d.updateAdd()
}

// nameChanged stores the name the user typed. From now on the main file no
// longer replaces it, and the message of a refusal goes away with the change.
func (d *addDialog) nameChanged(text string) {
	if d.syncing {
		return
	}
	d.form.Rename(text)
	d.clearError()
	d.updateAdd()
}

// submit sends the form: the shell refuses a name already saved before the
// Manager runs, and the Manager refuses the rest. One click starts one save.
func (d *addDialog) submit() {
	if d.busy || !d.form.CanAdd() {
		return
	}
	name := d.form.CleanName()
	files := append([]string{}, d.form.Files...)
	d.busy = true
	d.updateAdd()
	d.onSubmit(name, files, func(err error) { d.done(name, err) })
}

// done ends a save. A refusal keeps the form open with the chosen files and
// writes the Italian reason under the field. A save that worked closes the form,
// and the snapshot that follows brings the new Project to the list.
func (d *addDialog) done(name string, err error) {
	d.busy = false
	if err != nil {
		d.showError(UserError(err).Error())
		d.updateAdd()
		return
	}
	d.dialog.Accept()
	if d.onSaved != nil {
		d.onSaved(name)
	}
}

// updateAdd follows the state of the form: the confirm command is on only with a
// File compose, a name and no save in flight.
func (d *addDialog) updateAdd() {
	d.add.SetEnabled(!d.busy && d.form.CanAdd())
}

// showError puts a message under the name field.
func (d *addDialog) showError(message string) {
	d.error.SetText(message)
	d.error.SetVisible(true)
	d.name.SetAccessibleName(fieldAccessibleName(addNameLabelText, message))
}

// clearError hides the message of a refusal.
func (d *addDialog) clearError() {
	d.error.SetText("")
	d.error.SetVisible(false)
	d.name.SetAccessibleName(addNameLabelText)
}

// linkTabOrder puts the form in the reading order: the picker, the commands that
// drop a File compose, the name field and the two answers, in the order the box
// shows them. The rows are new at every refresh, so the chain is linked again.
func (d *addDialog) linkTabOrder() {
	prev := d.pick.QWidget
	for _, drop := range d.drops {
		qt6.QWidget_SetTabOrder(prev, drop.QWidget)
		prev = drop.QWidget
	}
	qt6.QWidget_SetTabOrder(prev, d.name.QWidget)
	qt6.QWidget_SetTabOrder(d.name.QWidget, d.add.QWidget)
	qt6.QWidget_SetTabOrder(d.add.QWidget, d.cancel.QWidget)
}

// pickComposeFiles shows the picker of one or more File compose over parent and
// hands the chosen paths to onPicked in the order the picker returns them. A
// picker the user closed calls nothing.
//
// The widget dialog of Qt carries the action "Show hidden files", so a File
// compose in a hidden folder is reachable when the user asks for it.
func pickComposeFiles(parent *qt6.QWidget, onPicked func([]string)) {
	picker := qt6.NewQFileDialog4(parent, pickFilesText)
	picker.SetOption2(qt6.QFileDialog__DontUseNativeDialog, true)
	picker.SetFileMode(qt6.QFileDialog__ExistingFiles)
	picker.SetAcceptMode(qt6.QFileDialog__AcceptOpen)
	picker.SetNameFilters([]string{composeFilterText, allFilesFilterText})
	picker.SetModal(true)
	picker.OnFilesSelected(onPicked)
	// The picker goes away with the event loop: a closed one leaves nothing
	// behind and no second picker stacks on the window.
	picker.OnFinished(func(int) { picker.DeleteLater() })
	picker.Open()
}
