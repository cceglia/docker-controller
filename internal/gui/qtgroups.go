package gui

import (
	"errors"
	"strconv"

	"github.com/mappu/miqt/qt6"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/model"
	"dockercontroller/internal/gui/design"
	"dockercontroller/internal/gui/forms"
	guiproject "dockercontroller/internal/gui/project"
	"dockercontroller/internal/gui/view"
)

// The texts of the dialog of the Groups.
const (
	// groupsTitle is the title of the dialog and the name of its list.
	groupsTitle = "Gruppi"
	// groupsCommandText opens the dialog from the form of a Project.
	groupsCommandText = "Gruppi…"
	// newGroupText, renameGroupText, deleteGroupText, moveUpText,
	// moveDownText and closeGroupsText are the commands of the dialog, in the
	// reading order.
	newGroupText    = "Nuovo"
	renameGroupText = "Rinomina"
	deleteGroupText = "Elimina"
	moveUpText      = "Sposta su"
	moveDownText    = "Sposta giù"
	closeGroupsText = "Chiudi"
	// groupNameLabelText names the inline field, and confirmGroupText and
	// cancelFieldText are its two commands.
	groupNameLabelText = "Nome del Gruppo"
	confirmGroupText   = "Conferma"
	cancelFieldText    = "Annulla"
	// groupNameHint says what the field does when it is empty or reserved.
	groupNameHint = "Il nome viene ripulito dagli spazi iniziali e finali. Non può essere vuoto, «Nessun Gruppo» o «Senza Gruppo»."
	// unionQuestionText and unionConfirmText name the question of a union;
	// deleteQuestionText names the question of a deletion.
	unionQuestionText  = "Unisci i Gruppi"
	unionConfirmText   = "Unisci"
	deleteQuestionText = "Elimina il Gruppo"
)

const (
	// groupsMinWidth and groupsMaxWidth bound the dialog. The lower bound keeps
	// the rows readable when a name is long, the upper one keeps the text from
	// running across the screen.
	groupsMinWidth = 460
	groupsMaxWidth = 640
)

// groupDialogOperator reaches the Manager for the operations of the dialog.
// Each call runs the operation and reports the refusal through done; the shell
// supplies the functions and runs the work off the Qt thread.
type groupDialogOperator struct {
	Create func(name string, done func(error))
	Rename func(from, to string, done func(error))
	Delete func(name string, done func(error))
	Move   func(name string, delta int, done func(error))
}

// groupsDialog is the modal dialog of the Groups: the list with the number of
// Projects of each Group and the commands of the dialog. Every operation
// applies at once, with no Save. The dialog keeps what
// the user typed across the snapshots that arrive while it is open, and it
// hands the form one record for every operation that worked.
type groupsDialog struct {
	dialog *qt6.QDialog

	// groups are the saved Groups in the snapshot order, and projects are the
	// Projects the counts read. Both follow the last snapshot and the outcome of
	// the operations of the dialog itself.
	groups   []string
	projects []compose.Project
	// selection is the row of the chosen Group, -1 when none is chosen.
	selection int
	// changes collects the records of the operations that worked, in order.
	changes []forms.GroupChange

	// fieldOpen is true while the inline name field is on screen. fieldCreate
	// says whether the field creates or renames, and fieldFrom is the Group a
	// rename started from.
	fieldOpen   bool
	fieldCreate bool
	fieldFrom   string
	// busy is true while the Manager works: a second click starts nothing and
	// the dialog stays open.
	busy bool
	// moveFrom is the row the running move started from. A snapshot can arrive
	// before the outcome, so the dialog swaps the list only while the Group
	// still sits there.
	moveFrom int
	// closed is true from the finished event, so an outcome that arrives after
	// the dialog went away touches no widget.
	closed bool
	// updating guards the programmatic rebuild of the list, so a selection the
	// dialog sets does not read as a click.
	updating bool

	list           *qt6.QListWidget
	errorLine      *qt6.QLabel
	entryRow       *qt6.QWidget
	entry          *qt6.QLineEdit
	fieldMessage   *qt6.QLabel
	confirmField   *qt6.QPushButton
	cancelField    *qt6.QPushButton
	newButton      *qt6.QPushButton
	renameButton   *qt6.QPushButton
	deleteButton   *qt6.QPushButton
	moveUpButton   *qt6.QPushButton
	moveDownButton *qt6.QPushButton
	closeButton    *qt6.QPushButton

	// controls are the widgets of the dialog in the reading order, and
	// fieldControls the ones of the open field. The tab order follows them.
	controls      []panelControl
	fieldControls []panelControl

	// op reaches the Manager. onConfirm asks the two confirmations; a nil
	// function opens the question of the dialog itself. onClose runs when the
	// dialog closes, whatever the way.
	op        groupDialogOperator
	onConfirm func(guiproject.Confirmation, func())
	onClose   func([]forms.GroupChange)
}

// newGroupsDialog builds the dialog on the last snapshot. The dialog reads the
// Groups and the Projects; the operations reach the Manager through op.
func newGroupsDialog(parent *qt6.QWidget, groups []string, projects []compose.Project, op groupDialogOperator) *groupsDialog {
	d := &groupsDialog{op: op, selection: -1}
	d.groups = append([]string{}, groups...)
	d.projects = projects

	d.dialog = qt6.NewQDialog(parent)
	d.dialog.SetWindowTitle(groupsTitle)
	d.dialog.SetAccessibleName(groupsTitle)
	d.dialog.SetModal(true)
	d.dialog.SetMinimumWidth(groupsMinWidth)
	d.dialog.SetMaximumWidth(groupsMaxWidth)

	layout := qt6.NewQVBoxLayout(d.dialog.QWidget)
	layout.SetContentsMargins(design.SpaceL, design.SpaceL, design.SpaceL, design.SpaceL)
	layout.SetSpacing(design.SpaceM)

	layout.AddWidget(detailLabel(groupsTitle, "dialogTitle").QWidget)

	d.list = qt6.NewQListWidget2()
	setObjectName(d.list.QWidget, "groupsList")
	d.list.SetAccessibleName(groupsTitle)
	d.list.SetSelectionMode(qt6.QAbstractItemView__SingleSelection)
	d.list.SetUniformItemSizes(true)
	d.list.OnCurrentRowChanged(d.rowChanged)
	// Alt+Up and Alt+Down move the chosen Group wherever the focus is. The list
	// takes the arrow keys before the dialog does, so it handles them here too.
	d.list.OnKeyPressEvent(func(super func(event *qt6.QKeyEvent), event *qt6.QKeyEvent) {
		if d.moveKey(event) {
			event.Accept()
			return
		}
		super(event)
	})
	layout.AddWidget2(d.list.QWidget, 1)

	d.errorLine = detailLabel("", "fieldError")
	setObjectName(d.errorLine.QWidget, "groupsError")
	d.errorLine.SetVisible(false)
	layout.AddWidget(d.errorLine.QWidget)

	layout.AddWidget(d.fieldRow())
	layout.AddLayout2(d.commands().QLayout, 0)

	d.linkTabOrder()
	d.dialog.OnFinished(func(int) {
		d.closed = true
		d.dialog.DeleteLater()
		if d.onClose != nil {
			d.onClose(d.changes)
		}
	})
	// The dialog stays open while an operation runs, whatever the way the close
	// is asked: Escape, Chiudi or the window itself.
	d.dialog.OnCloseEvent(func(super func(event *qt6.QCloseEvent), event *qt6.QCloseEvent) {
		if d.busy {
			event.Ignore()
			return
		}
		super(event)
	})
	// Escape cancels the open field and otherwise closes the dialog. Return
	// confirms the open field. Alt+Up and Alt+Down move the chosen Group. The
	// keys work wherever the focus is.
	d.dialog.OnKeyPressEvent(func(super func(event *qt6.QKeyEvent), event *qt6.QKeyEvent) {
		switch qt6.Key(event.Key()) {
		case qt6.Key_Escape:
			d.escape()
			event.Accept()
		case qt6.Key_Return, qt6.Key_Enter:
			if d.fieldOpen {
				d.confirmFieldAction()
				event.Accept()
				return
			}
			super(event)
		case qt6.Key_Up, qt6.Key_Down:
			if d.moveKey(event) {
				event.Accept()
				return
			}
			super(event)
		default:
			super(event)
		}
	})

	d.rebuildList()
	return d
}

// fieldRow builds the inline name field with its two commands and the message
// of a refusal. The row stays hidden until Nuovo or Rinomina opens it.
func (d *groupsDialog) fieldRow() *qt6.QWidget {
	d.entryRow = qt6.NewQWidget2()

	column := qt6.NewQVBoxLayout(d.entryRow)
	column.SetContentsMargins(0, 0, 0, 0)
	column.SetSpacing(design.SpaceXS)

	caption := detailLabel(groupNameLabelText, "fieldLabel")
	d.entry = qt6.NewQLineEdit2()
	setObjectName(d.entry.QWidget, "groupsEntry")
	d.entry.SetClearButtonEnabled(true)
	d.entry.SetAccessibleName(groupNameLabelText)
	d.entry.SetAccessibleDescription(groupNameHint)
	caption.SetBuddy(d.entry.QWidget)
	keepTarget(d.entry.QWidget)
	d.entry.OnReturnPressed(d.confirmFieldAction)
	column.AddWidget(caption.QWidget)

	row := qt6.NewQHBoxLayout2()
	row.SetSpacing(design.SpaceS)

	d.confirmField = detailButton(confirmGroupText, "groupsConfirm", "Salva il nome del Gruppo.", true)
	d.confirmField.OnClicked(d.confirmFieldAction)
	row.AddWidget2(d.entry.QWidget, 1)
	row.AddWidget(d.confirmField.QWidget)

	d.cancelField = detailButton(cancelFieldText, "groupsCancel", "Chiude il campo senza salvare.", true)
	d.cancelField.OnClicked(d.cancelFieldAction)
	row.AddWidget(d.cancelField.QWidget)
	column.AddLayout2(row.QLayout, 0)

	d.fieldMessage = detailLabel("", "fieldError")
	setObjectName(d.fieldMessage.QWidget, "groupsFieldError")
	d.fieldMessage.SetVisible(false)
	column.AddWidget(d.fieldMessage.QWidget)

	d.entryRow.SetVisible(false)

	d.fieldControls = []panelControl{
		{widget: d.entry.QWidget, label: groupNameLabelText},
		{widget: d.confirmField.QWidget, label: confirmGroupText},
		{widget: d.cancelField.QWidget, label: cancelFieldText},
	}
	return d.entryRow
}

// commands builds the commands of the dialog. Nuovo and Rinomina open the
// inline field; Rinomina, Elimina and the two moves stay off without a
// selection, and the moves also stop at the two ends.
func (d *groupsDialog) commands() *qt6.QHBoxLayout {
	row := qt6.NewQHBoxLayout2()
	row.SetSpacing(design.SpaceS)

	d.newButton = detailButton(newGroupText, "groupsNew", "Crea un Gruppo nuovo in fondo all'elenco.", true)
	d.newButton.OnClicked(d.openCreate)
	row.AddWidget(d.newButton.QWidget)

	d.renameButton = detailButton(renameGroupText, "groupsRename", "Cambia il nome del Gruppo scelto.", false)
	d.renameButton.OnClicked(d.openRename)
	row.AddWidget(d.renameButton.QWidget)

	d.deleteButton = detailButton(deleteGroupText, "groupsDelete", "Elimina il Gruppo scelto e lascia i suoi Progetti senza Gruppo.", false)
	d.deleteButton.OnClicked(d.askDelete)
	row.AddWidget(d.deleteButton.QWidget)

	d.moveUpButton = detailButton(moveUpText, "groupsMoveUp", "Sposta il Gruppo scelto di un posto verso l'alto. Alt+Su.", false)
	d.moveUpButton.OnClicked(func() { d.moveBy(-1) })
	row.AddWidget(d.moveUpButton.QWidget)

	d.moveDownButton = detailButton(moveDownText, "groupsMoveDown", "Sposta il Gruppo scelto di un posto verso il basso. Alt+Giù.", false)
	d.moveDownButton.OnClicked(func() { d.moveBy(1) })
	row.AddWidget(d.moveDownButton.QWidget)

	row.AddStretch()

	d.closeButton = detailButton(closeGroupsText, "groupsClose", "Chiude la gestione dei Gruppi.", true)
	d.closeButton.OnClicked(d.close)
	row.AddWidget(d.closeButton.QWidget)

	d.controls = []panelControl{
		{widget: d.list.QWidget, label: groupsTitle},
		{widget: d.newButton.QWidget, label: newGroupText},
		{widget: d.renameButton.QWidget, label: renameGroupText},
		{widget: d.deleteButton.QWidget, label: deleteGroupText},
		{widget: d.moveUpButton.QWidget, label: moveUpText},
		{widget: d.moveDownButton.QWidget, label: moveDownText},
		{widget: d.closeButton.QWidget, label: closeGroupsText},
	}
	return row
}

// linkTabOrder puts the two states of the dialog in the reading order: the
// list and the commands, or the field and its two commands.
func (d *groupsDialog) linkTabOrder() {
	for i := 1; i < len(d.controls); i++ {
		qt6.QWidget_SetTabOrder(d.controls[i-1].widget, d.controls[i].widget)
	}
	for i := 1; i < len(d.fieldControls); i++ {
		qt6.QWidget_SetTabOrder(d.fieldControls[i-1].widget, d.fieldControls[i].widget)
	}
}

// open shows the dialog with the focus on Nuovo, the first command, so a fresh
// dialog starts without a selection and Rinomina and Elimina stay off until the
// user chooses a Group.
func (d *groupsDialog) open() {
	d.dialog.AdjustSize()
	d.dialog.Open()
	d.selection = -1
	d.list.SetCurrentRow(-1)
	d.newButton.SetFocus()
	d.refreshCommands()
}

// setSnapshot keeps the Groups and the Projects of the last snapshot. The
// selection stays on the Group it was on, and the typed text of the open field
// stays as the user wrote it.
func (d *groupsDialog) setSnapshot(groups []string, projects []compose.Project) {
	if d.closed {
		return
	}
	keep := d.selectedName()
	d.groups = append([]string{}, groups...)
	d.projects = projects
	d.selection = d.indexOf(keep)
	d.rebuildList()
}

// groupNames returns the Groups the dialog shows now, in their order. The form
// reads them at the close, so its field proposes the updated list at once.
func (d *groupsDialog) groupNames() []string {
	return append([]string{}, d.groups...)
}

// selectedName is the saved spelling of the chosen Group, "" when none is
// chosen.
func (d *groupsDialog) selectedName() string {
	if d.selection < 0 || d.selection >= len(d.groups) {
		return ""
	}
	return d.groups[d.selection]
}

// indexOf finds a Group of the list by name, without regard to case, and
// returns -1 when it is not there.
func (d *groupsDialog) indexOf(name string) int {
	if name == "" {
		return -1
	}
	for i, group := range d.groups {
		if model.SameGroupName(group, name) {
			return i
		}
	}
	return -1
}

// projectCount counts the Projects of a Group.
func (d *groupsDialog) projectCount(name string) int {
	count := 0
	for _, project := range d.projects {
		if project.Group != "" && model.SameGroupName(project.Group, name) {
			count++
		}
	}
	return count
}

// rowChanged follows the choice of the user and turns the commands on or off.
func (d *groupsDialog) rowChanged(row int) {
	if d.updating {
		return
	}
	d.selection = row
	d.refreshCommands()
}

// rebuildList draws the rows again from the Groups and the Projects. The update
// is guarded, so setting the current row reads as no click.
func (d *groupsDialog) rebuildList() {
	d.updating = true
	d.list.Clear()
	for _, row := range forms.GroupRows(d.groups, d.projects) {
		d.list.AddItem(groupRowText(row))
	}
	if d.selection >= len(d.groups) {
		d.selection = len(d.groups) - 1
	}
	d.list.SetCurrentRow(d.selection)
	d.updating = false
	d.refreshCommands()
}

// groupRowText is one row of the list: the name and the number of Projects.
func groupRowText(row forms.GroupRow) string {
	return row.Name + " (" + groupCountText(row.Count) + ")"
}

// groupCountText says how many Projects a Group holds.
func groupCountText(count int) string {
	switch count {
	case 0:
		return "nessun Progetto"
	case 1:
		return "1 Progetto"
	default:
		return strconv.Itoa(count) + " Progetti"
	}
}

// refreshCommands turns the commands on and off from the state of the dialog:
// nothing while an operation runs, the list and Nuovo while the field is
// closed, and the field with its commands while it is open. A move is on only
// when the chosen Group has a neighbour in that direction.
func (d *groupsDialog) refreshCommands() {
	field := d.fieldOpen
	busy := d.busy
	d.list.SetEnabled(!field && !busy)
	d.newButton.SetEnabled(!field && !busy)
	hasSelection := d.selectedName() != ""
	d.renameButton.SetEnabled(!field && !busy && hasSelection)
	d.deleteButton.SetEnabled(!field && !busy && hasSelection)
	d.moveUpButton.SetEnabled(!field && !busy && d.canMove(-1))
	d.moveDownButton.SetEnabled(!field && !busy && d.canMove(1))
	d.closeButton.SetEnabled(!field && !busy)
	d.entry.SetEnabled(!busy)
	d.confirmField.SetEnabled(!busy)
	d.cancelField.SetEnabled(!busy)
}

// canMove reports whether the chosen Group has a neighbour delta places away.
func (d *groupsDialog) canMove(delta int) bool {
	if d.selection < 0 || d.selection >= len(d.groups) {
		return false
	}
	target := d.selection + delta
	return target >= 0 && target < len(d.groups)
}

// moveBy moves the chosen Group by delta places. The commands are off at the
// two ends and without a selection, and so is the key: the move stops there.
func (d *groupsDialog) moveBy(delta int) {
	if d.busy || d.fieldOpen {
		return
	}
	name := d.selectedName()
	if name == "" || !d.canMove(delta) {
		return
	}
	d.showListError("")
	d.moveFrom = d.selection
	d.busy = true
	d.refreshCommands()
	if d.op.Move == nil {
		d.moveDone(name, delta, nil)
		return
	}
	d.op.Move(name, delta, func(err error) { d.moveDone(name, delta, err) })
}

// moveKey runs Alt+Up or Alt+Down on the chosen Group and reports whether the
// key was a move. While the field is open or an operation runs the moves are
// off, so the key changes nothing.
func (d *groupsDialog) moveKey(event *qt6.QKeyEvent) bool {
	if event.Modifiers()&qt6.AltModifier == 0 {
		return false
	}
	switch qt6.Key(event.Key()) {
	case qt6.Key_Up:
		d.moveBy(-1)
		return true
	case qt6.Key_Down:
		d.moveBy(1)
		return true
	}
	return false
}

// moveDone applies the outcome of a move: an error sits under the list and
// changes no order, a move that worked puts the Group in its new place and
// keeps it selected. A snapshot can arrive before the outcome, so the dialog
// swaps the list itself only while the Group still sits where the move began.
func (d *groupsDialog) moveDone(name string, delta int, err error) {
	if d.closed {
		return
	}
	d.busy = false
	if err != nil {
		d.showListError(view.GroupErrorMessage(err))
		d.refreshCommands()
		return
	}
	index := d.indexOf(name)
	if index == d.moveFrom {
		d.groups, index = forms.MoveGroup(d.groups, index, delta)
	}
	if index >= 0 {
		d.selection = index
	}
	d.rebuildList()
	d.list.SetFocus()
}

// openCreate opens the field for a new Group.
func (d *groupsDialog) openCreate() {
	if d.busy || d.fieldOpen {
		return
	}
	d.openField(true, "")
}

// openRename opens the field on the name of the chosen Group.
func (d *groupsDialog) openRename() {
	if d.busy || d.fieldOpen {
		return
	}
	name := d.selectedName()
	if name == "" {
		return
	}
	d.openField(false, name)
}

// openField shows the inline field on the Group the operation starts from.
func (d *groupsDialog) openField(create bool, from string) {
	d.fieldOpen = true
	d.fieldCreate = create
	d.fieldFrom = from
	d.showListError("")
	d.setFieldMessage("")
	if create {
		d.entry.SetText("")
	} else {
		d.entry.SetText(from)
	}
	d.entryRow.SetVisible(true)
	d.refreshCommands()
	d.entry.SetFocus()
	d.entry.SelectAll()
}

// cancelFieldAction closes the inline field and changes nothing. Annulla always
// closes it, and so does Escape.
func (d *groupsDialog) cancelFieldAction() {
	if d.busy || !d.fieldOpen {
		return
	}
	d.fieldOpen = false
	d.fieldCreate = false
	d.fieldFrom = ""
	d.entryRow.SetVisible(false)
	d.setFieldMessage("")
	d.refreshCommands()
	d.list.SetFocus()
}

// confirmFieldAction applies the name on screen. A blank or reserved name is
// refused by the model rule before the Manager sees it; a duplicate is judged
// by the Manager. Conferma closes the field only when the operation works.
func (d *groupsDialog) confirmFieldAction() {
	if d.busy || !d.fieldOpen {
		return
	}
	name := model.NormalizeGroupName(d.entry.Text())
	d.setFieldMessage("")
	d.showListError("")

	if d.fieldCreate {
		check := model.ValidateGroupName(name, d.groups, model.GroupNameCreate, "")
		if check.Status == model.GroupNameInvalid {
			d.setFieldMessage(invalidGroupMessage())
			return
		}
		d.startCreate(check.Name)
		return
	}

	check := model.ValidateGroupName(name, d.groups, model.GroupNameRename, d.fieldFrom)
	switch check.Status {
	case model.GroupNameInvalid:
		d.setFieldMessage(invalidGroupMessage())
	case model.GroupNameDuplicate:
		from := d.fieldFrom
		d.ask(d.unionQuestion(from, check.Existing), func() {
			d.startRename(from, check.Existing, true)
		})
	default:
		d.startRename(d.fieldFrom, check.Name, false)
	}
}

// startCreate runs a creation. The field stays open with the refusal until the
// Manager says the name is saved.
func (d *groupsDialog) startCreate(name string) {
	d.showListError("")
	d.busy = true
	d.refreshCommands()
	if d.op.Create == nil {
		d.creationDone(name, nil)
		return
	}
	d.op.Create(name, func(err error) { d.creationDone(name, err) })
}

// creationDone applies the outcome of a creation: a refusal stays under the
// field, a saved name closes the field, goes to the end of the list and takes
// the selection.
func (d *groupsDialog) creationDone(name string, err error) {
	if d.closed {
		return
	}
	d.busy = false
	if err != nil {
		d.setFieldMessage(view.GroupErrorMessage(err))
		d.refreshCommands()
		d.entry.SetFocus()
		return
	}
	name = model.NormalizeGroupName(name)
	if i := d.indexOf(name); i < 0 {
		d.groups = append(d.groups, name)
	}
	d.changes = append(d.changes, forms.GroupChange{Kind: forms.GroupCreated, To: name})
	d.fieldOpen, d.fieldCreate, d.fieldFrom = false, false, ""
	d.entryRow.SetVisible(false)
	d.setFieldMessage("")
	d.selection = d.indexOf(name)
	d.rebuildList()
	d.list.SetFocus()
}

// startRename runs a rename, with or without a union, on the Group the field
// opened on.
func (d *groupsDialog) startRename(from, to string, union bool) {
	d.showListError("")
	d.busy = true
	d.refreshCommands()
	if d.op.Rename == nil {
		d.renameDone(from, to, union, nil)
		return
	}
	d.op.Rename(from, to, func(err error) { d.renameDone(from, to, union, err) })
}

// renameDone applies the outcome of a rename. A name refusal stays under the
// field; an error of another kind, such as an unknown Group, sits under the
// list. On success the field closes, the selection goes to the Group that
// survived and the record names the departure and the arrival.
func (d *groupsDialog) renameDone(from, to string, union bool, err error) {
	if d.closed {
		return
	}
	d.busy = false
	if err != nil {
		if isFieldRefusal(err) {
			d.setFieldMessage(view.GroupErrorMessage(err))
			d.entry.SetFocus()
		} else {
			d.showListError(view.GroupErrorMessage(err))
		}
		d.refreshCommands()
		return
	}
	d.applyRename(from, to, union)
	kind := forms.GroupRenamed
	if union {
		kind = forms.GroupUnited
	}
	d.changes = append(d.changes, forms.GroupChange{Kind: kind, From: from, To: to})
	d.fieldOpen, d.fieldCreate, d.fieldFrom = false, false, ""
	d.entryRow.SetVisible(false)
	d.setFieldMessage("")
	d.selection = d.indexOf(to)
	d.rebuildList()
	d.list.SetFocus()
}

// applyRename moves the Projects of the source to the arriving Group. A plain
// rename keeps the place; a union removes the source and leaves the Group that
// already had the name where it was.
func (d *groupsDialog) applyRename(from, to string, union bool) {
	for i := range d.projects {
		if model.SameGroupName(d.projects[i].Group, from) {
			d.projects[i].Group = to
		}
	}
	index := d.indexOf(from)
	if index < 0 {
		return
	}
	if union {
		d.groups = append(d.groups[:index], d.groups[index+1:]...)
		return
	}
	d.groups[index] = to
}

// askDelete asks the question that removes the chosen Group, with the count of
// the Projects that will go to Senza Gruppo.
func (d *groupsDialog) askDelete() {
	if d.busy || d.fieldOpen {
		return
	}
	name := d.selectedName()
	if name == "" {
		return
	}
	d.ask(d.deleteQuestion(name), func() { d.startDelete(name) })
}

// startDelete runs the deletion of the Group the question was about.
func (d *groupsDialog) startDelete(name string) {
	d.showListError("")
	d.busy = true
	d.refreshCommands()
	if d.op.Delete == nil {
		d.deleteDone(name, nil)
		return
	}
	d.op.Delete(name, func(err error) { d.deleteDone(name, err) })
}

// deleteDone applies the outcome of a deletion: an error sits under the list, a
// Group that left takes its Projects' Group away and the selection moves to the
// Group that took the empty place.
func (d *groupsDialog) deleteDone(name string, err error) {
	if d.closed {
		return
	}
	d.busy = false
	if err != nil {
		d.showListError(view.GroupErrorMessage(err))
		d.refreshCommands()
		return
	}
	index := d.indexOf(name)
	length := len(d.groups)
	for i := range d.projects {
		if model.SameGroupName(d.projects[i].Group, name) {
			d.projects[i].Group = ""
		}
	}
	if index >= 0 {
		d.groups = append(d.groups[:index], d.groups[index+1:]...)
	}
	d.changes = append(d.changes, forms.GroupChange{Kind: forms.GroupDeleted, From: name})
	d.selection = forms.SelectionAfterDeletion(index, length)
	d.rebuildList()
	d.list.SetFocus()
}

// ask asks a confirmation. The dialog's own question is a modal one over the
// dialog; a test replaces onConfirm to answer without a second window.
func (d *groupsDialog) ask(question guiproject.Confirmation, run func()) {
	if d.onConfirm != nil {
		d.onConfirm(question, run)
		return
	}
	confirm := newConfirmDialog(d.dialog.QWidget, question, run)
	confirm.open()
}

// unionQuestion asks to merge the Group of departure into the one that already
// has the name, and says how many Projects move.
func (d *groupsDialog) unionQuestion(from, to string) guiproject.Confirmation {
	count := d.projectCount(from)
	return guiproject.Confirmation{
		Title:        unionQuestionText,
		Message:      unionMessage(from, to, count),
		ConfirmLabel: unionConfirmText,
	}
}

// unionMessage says what a union does with the Projects of the source Group.
func unionMessage(from, to string, count int) string {
	switch count {
	case 0:
		return "«" + from + "» non ha Progetti e sparirà."
	case 1:
		return "1 Progetto di «" + from + "» passerà a «" + to + "»."
	default:
		return strconv.Itoa(count) + " Progetti di «" + from + "» passeranno a «" + to + "»."
	}
}

// deleteQuestion asks to remove a Group and says how many Projects go to Senza
// Gruppo.
func (d *groupsDialog) deleteQuestion(name string) guiproject.Confirmation {
	count := d.projectCount(name)
	return guiproject.Confirmation{
		Title:        deleteQuestionText + " «" + name + "»?",
		Message:      deleteMessage(name, count),
		ConfirmLabel: deleteGroupText,
		Destructive:  true,
	}
}

// deleteMessage says what the deletion does with the Projects of the Group.
func deleteMessage(name string, count int) string {
	switch count {
	case 0:
		return "Il Gruppo «" + name + "» non ha Progetti."
	case 1:
		return "1 Progetto passerà a Senza Gruppo."
	default:
		return strconv.Itoa(count) + " Progetti passeranno a Senza Gruppo."
	}
}

// showListError writes the message of an operation that does not concern the
// name field under the list, and clears it with an empty message.
func (d *groupsDialog) showListError(message string) {
	d.errorLine.SetText(message)
	d.errorLine.SetVisible(message != "")
}

// setFieldMessage writes the reason of a refused name under the field and in
// its accessible description, so a screen reader reads it with the field.
func (d *groupsDialog) setFieldMessage(message string) {
	d.fieldMessage.SetText(message)
	d.fieldMessage.SetVisible(message != "")
	if message == "" {
		d.entry.SetAccessibleDescription(groupNameHint)
		return
	}
	d.entry.SetAccessibleDescription(message)
}

// invalidGroupMessage is the Italian reason of a blank or reserved Group name.
func invalidGroupMessage() string {
	if message, ok := view.KnownMessage(compose.ErrInvalidGroupName); ok {
		return message
	}
	return view.GroupErrorMessage(compose.ErrInvalidGroupName)
}

// isFieldRefusal reports whether an error of a rename concerns the name field
// rather than the list.
func isFieldRefusal(err error) bool {
	return errors.Is(err, compose.ErrInvalidGroupName)
}

// escape handles Escape: it cancels the open field, or closes the dialog. It
// does nothing while an operation runs.
func (d *groupsDialog) escape() {
	if d.busy {
		return
	}
	if d.fieldOpen {
		d.cancelFieldAction()
		return
	}
	d.close()
}

// close closes the dialog, unless an operation runs or the field is open.
func (d *groupsDialog) close() {
	if d.busy || d.fieldOpen {
		return
	}
	d.dialog.Reject()
}
