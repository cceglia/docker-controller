package gui

import (
	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"

	"dockercontroller/internal/gui/design"
	"dockercontroller/internal/gui/forms"
	"dockercontroller/internal/gui/view"
	"dockercontroller/internal/settings"
)

const (
	// preferencesMinWidth and preferencesMaxWidth bound the dialog: a long
	// terminal command stays readable and the hint of the close behavior wraps
	// on a few lines.
	preferencesMinWidth = 480
	preferencesMaxWidth = 680
)

// preferencesDialog is the window of the settings: the three close behaviors,
// the terminal command, and the two commands of the dialog. It holds the values
// on screen only: nothing reaches the Store until Salva, so Annulla leaves
// nothing behind.
type preferencesDialog struct {
	dialog *qt6.QDialog
	// behavior is the close choice on screen, and choices are its radio
	// buttons, in the order of CloseBehaviorChoices.
	behavior settings.CloseBehavior
	choices  []*qt6.QRadioButton
	// terminal is the field of the command and error the message of a refused
	// save, under it.
	terminal *qt6.QLineEdit
	error    *qt6.QLabel
	// save and cancel are the commands of the dialog.
	save   *qt6.QPushButton
	cancel *qt6.QPushButton
	// busy is true while the Store writes: a second click on Salva saves once.
	busy bool
	// closed is true from the finished event: a save that ends after the dialog
	// went away touches no widget.
	closed bool
	// onSave hands the values to the application, which writes them off the Qt
	// thread; done reports the refusal, nil when the settings are saved.
	onSave func(settings.Settings, func(error))
	// onSaved runs after a save that worked, and onClosed after the dialog
	// closes, whatever the answer.
	onSaved  func()
	onClosed func()
}

// newPreferencesDialog builds the dialog on the saved settings. tray says
// whether the desktop has an icon: without one, hiding the window is no choice.
func newPreferencesDialog(parent *qt6.QWidget, prefs settings.Settings, tray bool, onSave func(settings.Settings, func(error))) *preferencesDialog {
	d := &preferencesDialog{onSave: onSave}

	d.dialog = qt6.NewQDialog(parent)
	d.dialog.SetWindowTitle(forms.PreferencesTitle)
	d.dialog.SetAccessibleName(forms.PreferencesTitle)
	d.dialog.SetModal(true)
	d.dialog.SetMinimumWidth(preferencesMinWidth)
	d.dialog.SetMaximumWidth(preferencesMaxWidth)

	layout := qt6.NewQVBoxLayout(d.dialog.QWidget)
	layout.SetContentsMargins(design.SpaceL, design.SpaceL, design.SpaceL, design.SpaceL)
	layout.SetSpacing(design.SpaceL)

	title := detailLabel(forms.PreferencesTitle, "dialogTitle")
	layout.AddWidget(title.QWidget)

	layout.AddWidget(d.closeGroup(tray, forms.CloseBehaviorShown(prefs.CloseBehavior, tray)))
	layout.AddWidget(d.terminalGroup(prefs.TerminalCommand))
	layout.AddWidget(d.buttons())

	d.linkTabOrder()

	d.dialog.OnFinished(func(int) {
		// The dialog is over: it goes away with the event loop, and the shell
		// that opened it forgets it. A save still in flight finds the flag and
		// touches nothing.
		d.closed = true
		d.dialog.DeleteLater()
		if d.onClosed != nil {
			d.onClosed()
		}
	})

	// Return saves from any field and Escape closes without saving, as the
	// other modal forms of the window do.
	for _, key := range []string{"Return", "Enter"} {
		save := qt6.NewQShortcut2(qt6.NewQKeySequence2(key), d.dialog.QObject)
		save.SetContext(qt6.WindowShortcut)
		save.OnActivated(d.submit)
	}
	escape := qt6.NewQShortcut2(qt6.NewQKeySequence2("Esc"), d.dialog.QObject)
	escape.SetContext(qt6.WindowShortcut)
	escape.OnActivated(func() { d.dialog.Reject() })

	return d
}

// closeGroup is the group of the three choices. A choice the desktop cannot
// honour stays on screen and off: a missing option would read as a bug, while a
// hint under it says why.
func (d *preferencesDialog) closeGroup(tray bool, shown settings.CloseBehavior) *qt6.QWidget {
	group := qt6.NewQGroupBox3(forms.CloseBehaviorGroupTitle)
	setObjectName(group.QWidget, "preferencesGroup")

	box := qt6.NewQVBoxLayout(group.QWidget)
	box.SetContentsMargins(design.SpaceM, design.SpaceM, design.SpaceM, design.SpaceM)
	box.SetSpacing(design.SpaceS)

	box.AddWidget(detailLabel(forms.CloseBehaviorGroupHint, "").QWidget)
	if !tray {
		box.AddWidget(detailLabel(forms.NoTrayHintText, "").QWidget)
	}

	d.behavior = shown
	for _, choice := range forms.CloseBehaviorChoices {
		choice := choice
		radio := qt6.NewQRadioButton3(choice.Label)
		setObjectName(radio.QWidget, "closeChoice")
		radio.SetAccessibleName(choice.Label)
		if !forms.CloseBehaviorAvailable(choice.Behavior, tray) {
			radio.SetEnabled(false)
			radio.SetToolTip(forms.NoTrayHintText)
			radio.SetAccessibleDescription(forms.NoTrayHintText)
		}
		radio.SetChecked(choice.Behavior == shown)
		radio.OnToggled(func(on bool) {
			if !on {
				return
			}
			d.behavior = choice.Behavior
			d.clearError()
		})
		keepTarget(radio.QWidget)
		d.choices = append(d.choices, radio)
		box.AddWidget(radio.QWidget)
	}

	return group.QWidget
}

// terminalGroup is the group of the terminal command: the field, the hint of the
// {dir} placeholder and, under the field, the message of a refused save.
func (d *preferencesDialog) terminalGroup(command string) *qt6.QWidget {
	group := qt6.NewQGroupBox3(forms.TerminalGroupTitle)
	setObjectName(group.QWidget, "preferencesGroup")

	box := qt6.NewQVBoxLayout(group.QWidget)
	box.SetContentsMargins(design.SpaceM, design.SpaceM, design.SpaceM, design.SpaceM)
	box.SetSpacing(design.SpaceS)

	box.AddWidget(detailLabel(forms.TerminalGroupHint, "").QWidget)

	caption := detailLabel(forms.TerminalRowTitle, "fieldLabel")
	d.terminal = qt6.NewQLineEdit2()
	setObjectName(d.terminal.QWidget, "terminalCommand")
	d.terminal.SetText(command)
	d.terminal.SetClearButtonEnabled(true)
	d.terminal.SetAccessibleName(forms.TerminalRowTitle)
	d.terminal.SetAccessibleDescription(forms.TerminalGroupHint)
	caption.SetBuddy(d.terminal.QWidget)
	keepTarget(d.terminal.QWidget)
	// A change of the command takes the old refusal away: the message must
	// describe what is on screen now.
	d.terminal.OnTextChanged(func(string) { d.clearError() })
	box.AddWidget(caption.QWidget)
	box.AddWidget(d.terminal.QWidget)

	d.error = detailLabel("", "fieldError")
	d.error.SetVisible(false)
	box.AddWidget(d.error.QWidget)

	return group.QWidget
}

// buttons are the two commands of the dialog. The box puts them in the order
// the platform expects, and Salva is the default, so Return writes the settings
// from any field.
func (d *preferencesDialog) buttons() *qt6.QWidget {
	box := qt6.NewQDialogButtonBox2()

	d.cancel = answerButton(box, cancelText, "preferencesCancel", "Chiude senza salvare le impostazioni.", qt6.QDialogButtonBox__RejectRole)
	d.cancel.OnClicked(func() { d.dialog.Reject() })

	d.save = answerButton(box, forms.SaveSettingsText, "primaryAction", "Salva le impostazioni nel database.", qt6.QDialogButtonBox__AcceptRole)
	d.save.SetDefault(true)
	d.save.SetAutoDefault(true)
	d.save.OnClicked(d.submit)

	return box.QWidget
}

// linkTabOrder puts the dialog in the reading order: the three choices, the
// command, then the two commands of the dialog in the order the box shows them.
func (d *preferencesDialog) linkTabOrder() {
	prev := d.choices[0].QWidget
	for _, radio := range d.choices[1:] {
		qt6.QWidget_SetTabOrder(prev, radio.QWidget)
		prev = radio.QWidget
	}
	qt6.QWidget_SetTabOrder(prev, d.terminal.QWidget)
	qt6.QWidget_SetTabOrder(d.terminal.QWidget, d.save.QWidget)
	qt6.QWidget_SetTabOrder(d.save.QWidget, d.cancel.QWidget)
}

// open shows the dialog with the focus on the chosen close behavior, so the
// keyboard starts from the setting the window obeys now.
func (d *preferencesDialog) open() {
	d.dialog.AdjustSize()
	d.dialog.Open()
	for i, radio := range d.choices {
		if i == forms.CloseBehaviorIndex(d.behavior) && radio.IsEnabled() {
			radio.SetFocus()
			return
		}
	}
	if len(d.choices) > 0 {
		d.choices[0].SetFocus()
	}
}

// submit validates the command and hands the settings to the application. A
// command that cannot be split shows the Italian message under the field and
// stays in the dialog: nothing is saved.
func (d *preferencesDialog) submit() {
	if d.busy {
		return
	}
	command := d.terminal.Text()
	if message := forms.ValidateTerminalCommand(command); message != "" {
		d.showError(message)
		d.terminal.SetFocus()
		return
	}
	d.busy = true
	d.save.SetEnabled(false)
	values := forms.StoredSettings(settings.Settings{CloseBehavior: d.behavior, TerminalCommand: command})
	if d.onSave == nil {
		d.done(nil)
		return
	}
	d.onSave(values, d.done)
}

// done ends a save. A refusal keeps the dialog open and writes the Italian
// reason under the field; a save that worked closes the dialog and confirms it
// in the window. A dialog that went away while the save ran keeps nothing.
func (d *preferencesDialog) done(err error) {
	if d.closed {
		return
	}
	d.busy = false
	d.save.SetEnabled(true)
	if err != nil {
		d.showError(view.UserError(err).Error())
		return
	}
	d.dialog.Accept()
	if d.onSaved != nil {
		d.onSaved()
	}
}

// showError puts a message under the command field and repeats it in the
// accessible name, so a screen reader reads the field and its error together.
func (d *preferencesDialog) showError(message string) {
	d.error.SetText(message)
	d.error.SetVisible(true)
	d.terminal.SetAccessibleName(fieldAccessibleName(forms.TerminalRowTitle, message))
}

// clearError hides the message of a refusal.
func (d *preferencesDialog) clearError() {
	if d.error == nil || !d.error.IsVisible() {
		return
	}
	d.error.SetText("")
	d.error.SetVisible(false)
	d.terminal.SetAccessibleName(forms.TerminalRowTitle)
}

// openPreferences shows the settings dialog. The application holds one at a
// time: a second click on "Impostazioni" brings the open one to the front.
func (a *application) openPreferences() {
	if a.shell == nil {
		return
	}
	if a.prefsDialog != nil {
		a.prefsDialog.dialog.Raise()
		a.prefsDialog.dialog.ActivateWindow()
		return
	}
	dialog := newPreferencesDialog(a.shell.window.QWidget, a.prefs, a.hasTray(), a.savePreferences)
	dialog.onSaved = func() {
		if a.shell != nil {
			a.shell.showNotice(forms.SettingsSavedText)
		}
	}
	dialog.onClosed = func() {
		if a.prefsDialog == dialog {
			a.prefsDialog = nil
		}
	}
	a.prefsDialog = dialog
	dialog.open()
}

// savePreferences writes the settings off the Qt thread and keeps the copy the
// close reads. The window keeps what the Store holds, so a blank command reads
// back as the default one.
func (a *application) savePreferences(values settings.Settings, done func(error)) {
	store := a.store
	ctx := a.ctx
	go func() {
		var err error
		if store != nil {
			err = store.Save(ctx, values)
		}
		mainthread.Start(func() {
			if err == nil {
				a.prefs = values
			}
			done(err)
		})
	}()
}
