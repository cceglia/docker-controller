package gui

import (
	"github.com/mappu/miqt/qt6"

	"dockercontroller/internal/gui/design"
)

const (
	// dialogMinWidth and dialogMaxWidth bound a modal question. Without the
	// lower bound a word-wrapped message gives the dialog the width of one
	// word, and the text is cut.
	dialogMinWidth = 420
	dialogMaxWidth = 560
)

// answerButton adds an answer to a box of a dialog. The box orders the answers
// the way the platform expects, so the dialog reads as the rest of the desktop.
// The name of the style sheet, the accessible name and the tooltip stay the
// ones of the command it replaces.
func answerButton(box *qt6.QDialogButtonBox, text, name, tooltip string, role qt6.QDialogButtonBox__ButtonRole) *qt6.QPushButton {
	button := box.AddButton2(text, role)
	if name != "" {
		setObjectName(button.QWidget, name)
	}
	button.SetAccessibleName(text)
	if tooltip != "" {
		button.SetToolTip(tooltip)
	}
	keepTarget(button.QWidget)
	return button
}

// confirmDialog is the modal question of a destructive Action. It is built
// before it shows and it never blocks the caller: the affirmative answer runs
// onRun, and every answer, including Escape, runs onClosed.
//
// The affirmative button is the default of the dialog, so Enter confirms it,
// and the standard dialog behaviour rejects on Escape. When the question names
// a text to type, the affirmative button stays off until the field holds it.
type confirmDialog struct {
	dialog  *qt6.QDialog
	confirm *qt6.QPushButton
	cancel  *qt6.QPushButton
	// entry is the field of TypeToConfirm, nil when the question asks for a
	// click alone.
	entry *qt6.QLineEdit
	// onRun runs the Action the user confirmed.
	onRun func()
	// onClosed runs after the dialog closes, whatever the answer.
	onClosed func()
}

// newConfirmDialog builds the question and the two answers. The texts are the
// ones of the Confirmation, so the dialog asks what the old view asked.
func newConfirmDialog(parent *qt6.QWidget, question Confirmation, onRun func()) *confirmDialog {
	c := &confirmDialog{onRun: onRun}

	c.dialog = qt6.NewQDialog(parent)
	c.dialog.SetWindowTitle(question.Title)
	c.dialog.SetAccessibleName(question.Title)
	c.dialog.SetModal(true)
	// A word-wrapped message gives a dialog the width of one word, so the text
	// is cut. The range keeps the sentence readable and still lets a short
	// question stay short.
	c.dialog.SetMinimumWidth(dialogMinWidth)
	c.dialog.SetMaximumWidth(dialogMaxWidth)

	layout := qt6.NewQVBoxLayout(c.dialog.QWidget)
	layout.SetContentsMargins(design.SpaceL, design.SpaceL, design.SpaceL, design.SpaceL)
	layout.SetSpacing(design.SpaceM)

	title := qt6.NewQLabel3(question.Title)
	setObjectName(title.QWidget, "dialogTitle")
	title.SetWordWrap(true)
	layout.AddWidget(title.QWidget)

	message := qt6.NewQLabel3(question.Message)
	setObjectName(message.QWidget, "dialogMessage")
	message.SetWordWrap(true)
	layout.AddWidget(message.QWidget)

	if question.TypeToConfirm != "" {
		hint := qt6.NewQLabel3("Per confermare digita " + question.TypeToConfirm)
		setObjectName(hint.QWidget, "dialogHint")
		hint.SetWordWrap(true)
		layout.AddWidget(hint.QWidget)

		c.entry = qt6.NewQLineEdit2()
		c.entry.SetPlaceholderText(question.TypeToConfirm)
		c.entry.SetAccessibleName("Testo di conferma")
		c.entry.SetAccessibleDescription("Digita " + question.TypeToConfirm + " per confermare.")
		layout.AddWidget(c.entry.QWidget)
	}

	box := qt6.NewQDialogButtonBox2()
	c.cancel = answerButton(box, "Annulla", "", "", qt6.QDialogButtonBox__RejectRole)
	c.confirm = answerButton(box, question.ConfirmLabel, "", "", qt6.QDialogButtonBox__AcceptRole)
	c.confirm.SetDefault(true)
	c.confirm.SetAutoDefault(true)
	c.confirm.SetEnabled(question.TypeToConfirm == "")
	layout.AddWidget(box.QWidget)

	if c.entry != nil {
		c.entry.OnTextChanged(func(text string) {
			c.confirm.SetEnabled(text == question.TypeToConfirm)
		})
	}
	c.confirm.OnClicked(func() { c.dialog.Accept() })
	c.cancel.OnClicked(func() { c.dialog.Reject() })

	c.dialog.OnAccepted(func() {
		if c.onRun != nil {
			c.onRun()
		}
	})
	c.dialog.OnFinished(func(int) {
		// The question is over: it goes away with the event loop, and the
		// panel that asked takes the focus back.
		c.dialog.DeleteLater()
		if c.onClosed != nil {
			c.onClosed()
		}
	})

	return c
}

// open shows the modal question. The field of the text to type takes the focus
// when the question has one, otherwise the affirmative default button does.
func (c *confirmDialog) open() {
	// The message is measured only now, when the style has sized the labels.
	c.dialog.AdjustSize()
	c.dialog.Open()
	if c.entry != nil {
		c.entry.SetFocus()
		return
	}
	c.confirm.SetFocus()
}
