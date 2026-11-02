package gui

import (
	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"

	"dockercontroller/internal/gui/design"
	"dockercontroller/internal/settings"
)

// closeDialog is the "Chiedi ogni volta" question of a close: the answers the
// desktop offers, the option to remember the answer, and nothing else. It never
// blocks the caller: the answer runs once the dialog has closed, so hiding or
// quitting never leaves a dialog behind.
type closeDialog struct {
	dialog *qt6.QDialog
	// remember is the option that saves the answer as the setting.
	remember *qt6.QCheckBox
	// answers are the buttons, in the order the answers rule gives them.
	answers []*qt6.QPushButton
	// cancel is the "Annulla" answer: the default of the dialog, so Enter
	// without a choice changes nothing.
	cancel *qt6.QPushButton
	// choice is the behavior of the answer: "" for "Annulla" and for a dialog
	// closed without an answer.
	choice settings.CloseBehavior
	// kept is the state of the option when the answer was chosen.
	kept     bool
	onAnswer func(settings.CloseBehavior, bool)
	// onClosed runs after the dialog closes, whatever the answer.
	onClosed func()
}

// newCloseDialog builds the question. Without a tray the answers leave
// "Nascondi" out, because the window could not come back.
func newCloseDialog(parent *qt6.QWidget, tray bool, onAnswer func(settings.CloseBehavior, bool)) *closeDialog {
	c := &closeDialog{onAnswer: onAnswer}

	c.dialog = qt6.NewQDialog(parent)
	c.dialog.SetWindowTitle(askCloseHeading)
	c.dialog.SetAccessibleName(askCloseHeading)
	c.dialog.SetModal(true)
	// The body wraps: the range keeps the sentence readable and still lets a
	// short answer stay short.
	c.dialog.SetMinimumWidth(dialogMinWidth)
	c.dialog.SetMaximumWidth(dialogMaxWidth)

	layout := qt6.NewQVBoxLayout(c.dialog.QWidget)
	layout.SetContentsMargins(design.SpaceL, design.SpaceL, design.SpaceL, design.SpaceL)
	layout.SetSpacing(design.SpaceM)

	title := detailLabel(askCloseHeading, "dialogTitle")
	layout.AddWidget(title.QWidget)

	body := detailLabel(closeQuestionBody(tray), "dialogMessage")
	layout.AddWidget(body.QWidget)

	c.remember = qt6.NewQCheckBox3(rememberCloseText)
	setObjectName(c.remember.QWidget, "closeRemember")
	c.remember.SetAccessibleName(rememberCloseText)
	c.remember.SetAccessibleDescription("Salva la risposta come impostazione, così la prossima chiusura non chiede.")
	keepTarget(c.remember.QWidget)
	layout.AddWidget(c.remember.QWidget)

	// The answers read in the order the platform gives them, and "Annulla" is
	// the default: Enter without a choice changes nothing.
	answers := qt6.NewQDialogButtonBox2()
	for _, answer := range availableCloseAnswers(tray) {
		answer := answer
		role := qt6.QDialogButtonBox__AcceptRole
		if answer.behavior == "" {
			role = qt6.QDialogButtonBox__RejectRole
		}
		button := answerButton(answers, answer.label, "closeAnswer", "", role)
		if answer.behavior == "" {
			c.cancel = button
			button.SetDefault(true)
			button.SetAutoDefault(true)
			button.OnClicked(func() { c.dialog.Reject() })
		} else {
			// Only the default answers Enter: a focused answer must not run by
			// a key the user did not mean as that answer.
			button.SetAutoDefault(false)
			button.OnClicked(func() {
				c.choice = answer.behavior
				c.kept = c.remember.IsChecked()
				c.dialog.Accept()
			})
		}
		c.answers = append(c.answers, button)
	}
	layout.AddWidget(answers.QWidget)

	c.dialog.OnFinished(func(int) {
		// The question is over: it goes away with the event loop, and the
		// answer runs after it, so the window it may hide is alone on screen.
		c.dialog.DeleteLater()
		if c.onClosed != nil {
			c.onClosed()
		}
		if c.choice != "" && c.onAnswer != nil {
			c.onAnswer(c.choice, c.kept)
		}
	})

	return c
}

// open shows the question and puts the focus on "Annulla", so a stray Enter does
// nothing.
func (c *closeDialog) open() {
	c.dialog.AdjustSize()
	c.dialog.Open()
	if c.cancel != nil {
		c.cancel.SetFocus()
		return
	}
	if len(c.answers) > 0 {
		c.answers[0].SetFocus()
	}
}

// closeRequested decides what the close of the window does: hide it in the tray,
// ask, or quit. A close that arrives while the application is quitting, and one
// that arrives with no window, start nothing.
func (a *application) closeRequested() {
	if a.quitting || a.shell == nil {
		return
	}
	a.performClose(a.prefs.CloseBehavior)
}

// performClose carries out a close behavior. The tray decides whether hiding is
// possible: without one the window could not come back, so the close quits.
func (a *application) performClose(behavior settings.CloseBehavior) {
	switch closeActionFor(behavior, a.hasTray()) {
	case closeQuit:
		a.quit()
	case closeAsk:
		a.askClose()
	default:
		a.shell.hideWindow()
	}
}

// hasTray reports whether the desktop offers an icon to bring the window back.
func (a *application) hasTray() bool {
	return a.tray != nil
}

// askClose opens the question of "Chiedi ogni volta". A question already on
// screen comes to the front instead of doubling.
func (a *application) askClose() {
	if a.shell == nil {
		return
	}
	if a.closeQuestion != nil {
		a.closeQuestion.dialog.Raise()
		a.closeQuestion.dialog.ActivateWindow()
		return
	}
	dialog := newCloseDialog(a.shell.window.QWidget, a.hasTray(), a.answerClose)
	dialog.onClosed = func() { a.closeQuestion = nil }
	a.closeQuestion = dialog
	dialog.open()
}

// answerClose carries out the answer of the question. A remembered answer is
// saved first, so the choice still holds after a restart; the close waits for
// the save, because the next close must already read the new value.
func (a *application) answerClose(choice settings.CloseBehavior, remember bool) {
	value, save := closeToSave(choice, remember)
	if !save {
		a.performClose(choice)
		return
	}
	store := a.store
	ctx := a.ctx
	go func() {
		err := saveCloseBehavior(ctx, store, value)
		mainthread.Start(func() {
			if err != nil {
				if a.shell != nil {
					a.shell.showNotice(UserError(err).Error())
				}
				return
			}
			a.prefs.CloseBehavior = value
			a.performClose(value)
		})
	}()
}
