package gui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"dockercontroller/internal/compose"
)

const (
	// dataTitle is the heading of the block of the identity of the Project.
	dataTitle = "Dati del Progetto"
	// profilesTitle is the heading of the block of the Profiles.
	profilesTitle = "Profili"
	// editText opens the form of the settings, saveText confirms it and
	// cancelText leaves it untouched.
	editText   = "Modifica"
	saveText   = "Salva"
	cancelText = "Annulla"
	// activeProfileText and idleProfileText read the state of a Profile at
	// rest.
	activeProfileText = "Attivo"
	idleProfileText   = "Non attivo"
	// formHint explains what an empty Display name does, and newGroupTitle names
	// the field of a Group the user is creating.
	formHint      = "Un Nome visualizzato vuoto mostra il nome compose."
	newGroupTitle = "Nome del nuovo Gruppo"
	// profilesHint says what the Profiles drive, as the old view did.
	profilesHint = "Vengono usati da Up, Stop, Restart, Pull e Build. " +
		"Cambiarli non avvia né ferma nulla. Down agisce su tutti i Profili."
	// noFolderText is the value of a Project without a folder.
	noFolderText = "Nessuna cartella"
)

// projectData is the block of the detail at rest: the Display name, the compose
// name, the Group and the folder of the Project. onEdit opens the form.
func projectData(project compose.Project, onEdit func()) *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle(dataTitle)

	edit := gtk.NewButtonWithLabel(editText)
	edit.AddCSSClass("flat")
	edit.SetTooltipText("Modifica Nome visualizzato, Gruppo e Profili attivi.")
	edit.ConnectClicked(onEdit)
	group.SetHeaderSuffix(edit)

	displayRow := adw.NewActionRow()
	displayRow.SetTitle("Nome visualizzato")
	displayRow.AddSuffix(valueLabel(ProjectName(project), ""))
	if project.DisplayName == "" {
		displayRow.SetTooltipText("Il Progetto mostra il nome compose.")
	}
	group.Add(displayRow)

	composeRow := adw.NewActionRow()
	composeRow.SetTitle("Nome compose")
	composeRow.AddSuffix(valueLabel(project.Name, "monospace"))
	group.Add(composeRow)

	groupRow := adw.NewActionRow()
	groupRow.SetTitle("Gruppo")
	if project.Group == "" {
		groupRow.AddSuffix(valueLabel(NoGroupText, "dim-label"))
	} else {
		groupRow.AddSuffix(valueLabel(project.Group, ""))
	}
	group.Add(groupRow)

	group.Add(folderRow(project.Folder))
	return group
}

// projectProfiles is the block of the Profiles at rest: one row per Profile the
// compose files define, with its state. A Project without Profiles has no block.
func projectProfiles(project compose.Project) *adw.PreferencesGroup {
	names := ProfileChoices(project)
	if len(names) == 0 {
		return nil
	}
	group := adw.NewPreferencesGroup()
	group.SetTitle(profilesTitle)
	for _, name := range names {
		row := adw.NewActionRow()
		row.SetTitle(name)
		if HasProfile(project.ActiveProfiles, name) {
			row.AddSuffix(valueLabel(activeProfileText, ""))
		} else {
			row.AddSuffix(valueLabel(idleProfileText, "dim-label"))
		}
		group.Add(row)
	}
	return group
}

// projectForm is the form of the settings while the user edits them: the
// Display name, the Group, the field of a new Group, the active Profiles and the
// two buttons. It changes the pending edit only, so Annulla has nothing to undo
// on the backend.
func projectForm(project compose.Project, groups []string, edit *ProjectEdit, onSave, onCancel func()) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationVertical, detailSpacing)

	data := adw.NewPreferencesGroup()
	data.SetTitle(dataTitle)
	data.SetDescription(formHint)

	nameRow := adw.NewActionRow()
	nameRow.SetTitle("Nome visualizzato")
	entry := gtk.NewEntry()
	entry.SetText(edit.DisplayName)
	entry.SetPlaceholderText(project.Name)
	entry.SetWidthChars(24)
	entry.SetHExpand(true)
	entry.SetTooltipText(formHint)
	entry.ConnectChanged(func() { edit.DisplayName = entry.Text() })
	nameRow.AddSuffix(entry)
	nameRow.SetActivatableWidget(entry)
	data.Add(nameRow)

	groupRow := adw.NewActionRow()
	groupRow.SetTitle("Gruppo")
	dropdown := gtk.NewDropDownFromStrings(groupLabels(groups))
	dropdown.SetSelected(uint(GroupOption(groups, *edit)))
	dropdown.SetTooltipText("Il Gruppo che raggruppa il Progetto nella lista.")
	groupRow.AddSuffix(dropdown)
	groupRow.SetActivatableWidget(dropdown)
	data.Add(groupRow)

	newGroupRow := adw.NewActionRow()
	newGroupRow.SetTitle(newGroupTitle)
	newGroupEntry := gtk.NewEntry()
	newGroupEntry.SetText(edit.NewGroupName)
	newGroupEntry.SetWidthChars(24)
	newGroupEntry.SetHExpand(true)
	newGroupEntry.SetTooltipText("Il nome del Gruppo da creare.")
	newGroupRow.AddSuffix(newGroupEntry)
	newGroupRow.SetActivatableWidget(newGroupEntry)
	newGroupRow.SetVisible(edit.NewGroup)
	data.Add(newGroupRow)

	buttons := gtk.NewBox(gtk.OrientationHorizontal, 6)
	buttons.SetVAlign(gtk.AlignCenter)

	cancel := gtk.NewButtonWithLabel(cancelText)
	cancel.AddCSSClass("flat")
	cancel.SetTooltipText("Lascia i valori di prima.")
	cancel.ConnectClicked(onCancel)
	buttons.Append(cancel)

	save := gtk.NewButtonWithLabel(saveText)
	save.AddCSSClass("suggested-action")
	save.SetTooltipText("Salva Nome visualizzato, Gruppo e Profili attivi.")
	save.SetSensitive(edit.SaveEnabled())
	save.ConnectClicked(func() {
		// The choice of the Group is read here too, so a selection that did
		// not emit a change still reaches the form.
		edit.SetGroupOption(groups, int(dropdown.Selected()))
		onSave()
	})
	buttons.Append(save)

	// The new Group field follows the choice of the Group, and Salva follows the
	// name typed there, so the old view behaves the same way.
	dropdown.NotifyProperty("selected", func() {
		edit.SetGroupOption(groups, int(dropdown.Selected()))
		newGroupRow.SetVisible(edit.NewGroup)
		save.SetSensitive(edit.SaveEnabled())
	})
	newGroupEntry.ConnectChanged(func() {
		edit.NewGroupName = newGroupEntry.Text()
		save.SetSensitive(edit.SaveEnabled())
	})

	data.SetHeaderSuffix(buttons)
	box.Append(data)

	if profiles := profileChecks(project, edit); profiles != nil {
		box.Append(profiles)
	}
	return box
}

// profileChecks is the block of the Profiles in the form: one check per Profile
// the compose files define. A Project without Profiles has no block.
func profileChecks(project compose.Project, edit *ProjectEdit) *adw.PreferencesGroup {
	names := ProfileChoices(project)
	if len(names) == 0 {
		return nil
	}
	group := adw.NewPreferencesGroup()
	group.SetTitle(profilesTitle + " attivi")
	group.SetDescription(profilesHint)
	for _, name := range names {
		check := gtk.NewCheckButton()
		check.SetActive(edit.ProfileOn(name))
		check.SetVAlign(gtk.AlignCenter)
		check.ConnectToggled(func() { edit.SetProfile(name, check.Active()) })

		row := adw.NewActionRow()
		row.SetTitle(name)
		row.AddSuffix(check)
		row.SetActivatableWidget(check)
		group.Add(row)
	}
	return group
}

// groupLabels is the text of each option of the Group field: no Group reads as
// a sentence, every Group keeps its name, and the last option asks for a new
// Group.
func groupLabels(choices []string) []string {
	labels := make([]string, 0, len(choices)+1)
	for _, choice := range choices {
		if choice == "" {
			labels = append(labels, NoGroupText)
			continue
		}
		labels = append(labels, choice)
	}
	if len(labels) == 0 {
		labels = append(labels, NoGroupText)
	}
	return append(labels, NewGroupText)
}

// valueLabel is a short value at the end of a row, with an optional Adwaita
// class that dims it or sets it in the mono font.
func valueLabel(text, cssClass string) *gtk.Label {
	label := gtk.NewLabel(text)
	label.SetVAlign(gtk.AlignCenter)
	if cssClass != "" {
		label.AddCSSClass(cssClass)
	}
	return label
}

// folderRow is the row of the folder of the Project. The path is the subtitle,
// so it ellipsizes instead of widening the detail; the tooltip holds it whole.
func folderRow(folder string) *adw.ActionRow {
	row := adw.NewActionRow()
	row.SetTitle("Cartella")
	if folder == "" {
		row.AddSuffix(valueLabel(noFolderText, "dim-label"))
		return row
	}
	row.SetSubtitle(folder)
	row.SetSubtitleLines(1)
	row.SetTooltipText(folder)
	return row
}
