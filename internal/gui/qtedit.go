package gui

import (
	"github.com/mappu/miqt/qt6"

	"dockercontroller/internal/compose"
)

// The texts of the form of the settings.
const (
	// editText opens the form in place of the data of the Project, saveText
	// keeps the values and cancelText leaves them as they are.
	editText   = "Modifica"
	saveText   = "Salva"
	cancelText = "Annulla"
	// editTitle is the heading of the form and its accessible name. dataTitle
	// is the heading of the data of the Project, the block the form replaces.
	editTitle = "Modifica del Progetto"
	dataTitle = "Dati del Progetto"
	// displayNameLabelText, groupLabelText and profilesLabelText name the three
	// fields of the form, in the reading order.
	displayNameLabelText = "Nome visualizzato"
	groupLabelText       = "Gruppo"
	profilesLabelText    = "Profili attivi"
	// displayNameHint says what an empty Display name does, groupHint what the
	// Group does and profilesHint what the Profiles drive.
	displayNameHint = "Un Nome visualizzato vuoto mostra il nome compose."
	groupHint       = "Raggruppa il Progetto nella lista. Vuoto per nessun Gruppo."
	profilesHint    = "Vengono usati da Up, Stop, Restart, Pull e Build. " +
		"Cambiarli non avvia né ferma nulla. Down agisce su tutti i Profili."
	// noProfilesText explains a form with no Profile to choose.
	noProfilesText = "I File compose del Progetto non definiscono Profili."
)

// editForm is the form of the settings while it is open: the values the user
// typed, the message of a refused save, and the widgets of the last render. The
// values live here and not in the widgets, so a render builds the form again
// without losing what the user typed.
type editForm struct {
	edit ProjectEdit
	// message is the Italian reason of a refused save, "" while there is none.
	// It sits under the Display name field, the first field of the form.
	message string
	// saving is true while the shell reaches the Manager: a second click on
	// Salva starts nothing.
	saving bool
	// name, group, checks and save are the widgets of the last render. A render
	// replaces them, so a form reads and touches only the current ones.
	name   *qt6.QLineEdit
	group  *qt6.QComboBox
	checks map[string]*qt6.QCheckBox
	save   *qt6.QPushButton
}

// startEdit opens the form on the saved values of the Project. Nothing reaches
// the Manager until Salva, so Annulla has nothing to undo.
func (p *detailPanel) startEdit() {
	if p.edit != nil || !p.visible {
		return
	}
	p.edit = &editForm{edit: NewProjectEdit(p.shown.project)}
	p.render()
	p.focusEdit()
}

// cancelEdit leaves the form and shows the saved values again.
func (p *detailPanel) cancelEdit() {
	if p.edit == nil {
		return
	}
	p.closeEdit()
}

// saveEdit hands the values of the form to the shell, which reaches the Manager
// off the Qt thread. Salva stays insensitive until the answer, so one click
// starts one save.
func (p *detailPanel) saveEdit() {
	if p.edit == nil || p.edit.saving {
		return
	}
	p.edit.saving = true
	if p.edit.save != nil {
		p.edit.save.SetEnabled(false)
	}
	name, settings := p.edit.edit.Project, p.edit.edit.Settings()
	if p.onSave == nil {
		p.saveDone(nil)
		return
	}
	p.onSave(name, settings)
}

// saveDone ends a save. A refusal keeps the form open with the typed values and
// writes the Italian reason under the field it concerns, so the user corrects
// it in place. A save that worked closes the form, and the snapshot that follows
// brings the new values to the list and to the detail.
func (p *detailPanel) saveDone(err error) {
	if p.edit == nil {
		return
	}
	p.edit.saving = false
	if err != nil {
		p.edit.message = UserError(err).Error()
		p.render()
		p.focusEdit()
		return
	}
	p.closeEdit()
}

// closeEdit leaves the form, draws the detail again and gives the focus back to
// the command that opened it, so the keyboard goes on from where it started.
func (p *detailPanel) closeEdit() {
	if p.edit == nil {
		return
	}
	p.edit = nil
	p.waitingDraw = false
	p.render()
	p.focusControl(editText)
}

// focusEdit puts the focus on the first field of the form, so the keyboard
// reaches the form at its first value.
func (p *detailPanel) focusEdit() {
	if p.edit != nil && p.edit.name != nil {
		p.edit.name.SetFocus()
	}
}

// setProjects keeps the Projects of the snapshot. The Group field proposes the
// Groups the Projects in use carry, so it reads them here.
func (p *detailPanel) setProjects(projects []compose.Project) {
	p.projects = projects
}

// fieldAccessibleName is the accessible name of a field of the form: its
// caption, then the message that says what is wrong with it, so a screen reader
// repeats the error with the field.
func fieldAccessibleName(caption, message string) string {
	if message == "" {
		return caption
	}
	return caption + ". " + message
}

// editField is one field of the form: its caption, the widget the user works in
// and, under it, the message of a refused save and the hint. The labels wrap, so
// a narrow panel keeps the whole text and the field its height.
func editField(caption string, field *qt6.QWidget, message, hint string) *qt6.QWidget {
	widget := qt6.NewQWidget2()

	box := qt6.NewQVBoxLayout(widget)
	box.SetContentsMargins(0, 0, 0, 0)
	box.SetSpacing(spaceXS)

	label := detailLabel(caption, "fieldLabel")
	label.SetBuddy(field)
	box.AddWidget(label.QWidget)
	box.AddWidget(field)
	if message != "" {
		box.AddWidget(detailLabel(message, "fieldError").QWidget)
	}
	if hint != "" {
		box.AddWidget(detailLabel(hint, "").QWidget)
	}
	return widget
}

// editBlock builds the form of the settings in place of the data of the
// Project: the Display name, the Group with the Groups in use to choose from,
// the active Profiles, and the two commands. Every change goes to the pending
// form, so a render builds it again from the same values.
func (p *detailPanel) editBlock() *qt6.QWidget {
	form, project := p.edit, p.shown.project

	// The form is a card of its own. The border separates it from the row of the
	// Project's commands below, so Salva and Annulla never read as two more
	// commands in that row.
	frame := qt6.NewQFrame2()
	frame.SetFrameShape(qt6.QFrame__NoFrame)
	setObjectName(frame.QWidget, "editPanel")
	block := frame.QWidget
	block.SetAccessibleName(editTitle)
	box := qt6.NewQVBoxLayout(block)
	box.SetContentsMargins(spaceM, spaceM, spaceM, spaceM)
	// The fields sit a step apart, further than the lines inside one field, so
	// the eye reads where one field ends and the next starts.
	box.SetSpacing(spaceM)

	box.AddWidget(detailLabel(editTitle, "detailHeading").QWidget)

	// Nome visualizzato: the saved value, the compose name as its hint, and
	// under the field the reason of a refused save.
	name := qt6.NewQLineEdit2()
	setObjectName(name.QWidget, "editName")
	name.SetText(form.edit.DisplayName)
	name.SetPlaceholderText(project.Name)
	name.SetClearButtonEnabled(true)
	name.SetAccessibleName(fieldAccessibleName(displayNameLabelText, form.message))
	name.SetAccessibleDescription(displayNameHint)
	keepTarget(name.QWidget)
	name.OnTextChanged(func(text string) { form.edit.DisplayName = text })
	form.name = name
	box.AddWidget(editField(displayNameLabelText, name.QWidget, form.message, displayNameHint))

	// Gruppo: a box the user types in, which proposes the Groups already in use.
	// An empty box is no Group, and the hint of the box says so.
	group := qt6.NewQComboBox2()
	setObjectName(group.QWidget, "editGroup")
	group.SetEditable(true)
	group.SetInsertPolicy(qt6.QComboBox__NoInsert)
	group.SetPlaceholderText(NoGroupText)
	group.AddItems(GroupSuggestions(GroupChoices(form.edit.Group, p.projects)))
	group.SetCurrentText(form.edit.Group)
	group.SetAccessibleName(groupLabelText)
	group.SetAccessibleDescription(groupHint)
	keepTarget(group.QWidget)
	group.OnCurrentTextChanged(func(text string) {
		form.edit.Group = text
		form.edit.NewGroup = false
	})
	form.group = group
	box.AddWidget(editField(groupLabelText, group.QWidget, "", groupHint))

	// The two fields join the tab chain in the reading order, before the
	// Profiles and the commands.
	p.controls = append(p.controls,
		panelControl{widget: name.QWidget, label: displayNameLabelText},
		panelControl{widget: group.QWidget, label: groupLabelText},
	)

	// Profili attivi: one check per Profile the compose files define. A Project
	// without Profile says so, so the block is never silently empty.
	box.AddWidget(detailLabel(profilesLabelText, "detailHeading").QWidget)
	box.AddWidget(detailLabel(profilesHint, "").QWidget)

	names := ProfileChoices(project)
	form.checks = make(map[string]*qt6.QCheckBox, len(names))
	if len(names) == 0 {
		box.AddWidget(detailLabel(noProfilesText, "detailEmpty").QWidget)
	}
	for _, profile := range names {
		check := qt6.NewQCheckBox3(profile)
		setObjectName(check.QWidget, "editProfile")
		check.SetChecked(form.edit.ProfileOn(profile))
		check.SetAccessibleName(profile)
		check.SetAccessibleDescription(profilesHint)
		keepTarget(check.QWidget)
		check.OnToggled(func(on bool) { form.edit.SetProfile(profile, on) })
		form.checks[profile] = check
		box.AddWidget(check.QWidget)
		p.controls = append(p.controls, panelControl{widget: check.QWidget, label: profile})
	}

	// The two commands close the form. Annulla comes first, so the reading order
	// ends on Salva, the command that keeps the values.
	buttons := qt6.NewQHBoxLayout2()
	buttons.SetSpacing(spaceS)

	cancel := detailButton(cancelText, "editCancel", "Lascia i valori di prima.", true)
	cancel.OnClicked(p.cancelEdit)
	buttons.AddWidget(cancel.QWidget)

	save := detailButton(saveText, "primaryAction", "Salva Nome visualizzato, Gruppo e Profili attivi.", !form.saving)
	save.OnClicked(p.saveEdit)
	buttons.AddWidget(save.QWidget)
	buttons.AddStretch()
	box.AddLayout2(buttons.QLayout, 0)
	form.save = save

	// The tab chain of the detail follows p.controls, so the commands join it
	// after the Profiles and close the reading order of the form.
	p.controls = append(p.controls,
		panelControl{widget: cancel.QWidget, label: cancelText},
		panelControl{widget: save.QWidget, label: saveText},
	)

	// The form is fully usable from the keyboard: Return saves and Escape
	// cancels from any of its fields. The context is the form itself, so the two
	// keys keep their meaning elsewhere in the window.
	for _, key := range []string{"Return", "Enter"} {
		confirm := qt6.NewQShortcut2(qt6.NewQKeySequence2(key), block.QObject)
		confirm.SetContext(qt6.WidgetWithChildrenShortcut)
		confirm.OnActivated(p.saveEdit)
	}
	escape := qt6.NewQShortcut2(qt6.NewQKeySequence2("Esc"), block.QObject)
	escape.SetContext(qt6.WidgetWithChildrenShortcut)
	escape.OnActivated(p.cancelEdit)

	return block
}
