package gui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"dockercontroller/internal/settings"
)

const (
	// preferencesTitle is the title of the settings window.
	preferencesTitle = "Impostazioni"
	// preferencesWidth and preferencesHeight are the size it opens at.
	preferencesWidth  = 620
	preferencesHeight = 560
	// saveSettingsText is the action that stores the settings.
	saveSettingsText = "Salva"
	// settingsSavedText confirms a save of the settings.
	settingsSavedText = "Impostazioni salvate."
	// closeBehaviorGroupTitle and closeBehaviorGroupHint describe the group of
	// the close behavior.
	closeBehaviorGroupTitle = "Alla chiusura della finestra"
	closeBehaviorGroupHint  = "Cosa fa l'app quando chiudi la finestra. Restando nel tray la riapri " +
		"con un click sull'icona; la cronologia delle Attività va persa."
	// terminalGroupTitle and terminalGroupHint describe the group of the
	// terminal command.
	terminalGroupTitle = "Comando del terminale"
	terminalGroupHint  = "{dir} viene sostituito con la cartella del Progetto. " +
		"Vuoto per il comando predefinito."
	// terminalRowTitle names the field of the command.
	terminalRowTitle = "Comando"
)

// closeBehaviorChoice is one option of the close behavior: the label it shows
// and the value it saves.
type closeBehaviorChoice struct {
	label    string
	behavior settings.CloseBehavior
}

// closeBehaviorChoices are the three behaviors, in the order the window shows
// them.
var closeBehaviorChoices = []closeBehaviorChoice{
	{label: "Chiudi la finestra e resta nel tray", behavior: settings.CloseHide},
	{label: "Esci dall'app", behavior: settings.CloseQuit},
	{label: "Chiedi ogni volta", behavior: settings.CloseAsk},
}

// closeBehaviorIndex is the place of a behavior in the choices. A value the
// settings do not know shows the default.
func closeBehaviorIndex(b settings.CloseBehavior) int {
	for i, choice := range closeBehaviorChoices {
		if choice.behavior == b {
			return i
		}
	}
	return 0
}

// closeBehaviorAt is the behavior of a place in the choices. A place out of the
// list is the default.
func closeBehaviorAt(i int) settings.CloseBehavior {
	if i < 0 || i >= len(closeBehaviorChoices) {
		return settings.Defaults().CloseBehavior
	}
	return closeBehaviorChoices[i].behavior
}

// preferencesSaveOutcome turns the result of a save into the message of the
// window and says whether the settings are saved. The refusal of an invalid
// terminal command is the Italian message of UserError.
func preferencesSaveOutcome(err error) (message string, saved bool) {
	if err != nil {
		return UserError(err).Error(), false
	}
	return settingsSavedText, true
}

// prefsWindow is the settings window: the close behavior and the terminal
// command. It reads the saved settings when it opens and writes them back with
// its save action.
type prefsWindow struct {
	app *application
	// window is the AdwPreferencesWindow: one page and two groups.
	window *adw.PreferencesWindow
	// checks are the three choices of the close behavior, and behavior is the
	// one on screen.
	checks   []*gtk.CheckButton
	behavior settings.CloseBehavior
	// terminal is the field of the command.
	terminal *adw.EntryRow
}

// openPreferences shows the settings window, building it on the first use. It is
// a window of its own, and it goes away with the window that opened it.
func (a *application) openPreferences() {
	if a.prefs != nil {
		a.prefs.window.Present()
		return
	}
	p := newPreferencesWindow(a)
	a.prefs = p
	if a.window != nil {
		p.window.SetTransientFor(&a.window.window.ApplicationWindow.Window)
		p.window.SetDestroyWithParent(true)
	}
	// The window remembers itself while it is on screen, so the next click on
	// "Impostazioni" presents it instead of building a second one.
	p.window.ConnectCloseRequest(func() bool {
		if a.prefs == p {
			a.prefs = nil
		}
		return false
	})
	p.load()
	p.window.Present()
}

// newPreferencesWindow builds the settings window with the default values. The
// saved ones arrive with load.
func newPreferencesWindow(a *application) *prefsWindow {
	p := &prefsWindow{app: a, behavior: settings.Defaults().CloseBehavior}

	p.window = adw.NewPreferencesWindow()
	p.window.SetTitle(preferencesTitle)
	p.window.SetDefaultSize(preferencesWidth, preferencesHeight)

	page := adw.NewPreferencesPage()
	page.Add(p.closeBehaviorGroup())
	page.Add(p.terminalGroup())
	page.Add(p.saveGroup())
	p.window.Add(page)

	return p
}

// saveGroup is the action of the window: an AdwButtonRow, because an
// AdwPreferencesWindow has no place for an extra action in its header and
// refuses a titlebar of its own.
func (p *prefsWindow) saveGroup() *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()

	save := adw.NewButtonRow()
	save.SetTitle(saveSettingsText)
	save.AddCSSClass("suggested-action")
	save.ConnectActivated(p.save)
	group.Add(save)

	return group
}

// closeBehaviorGroup is the group of the three choices. The checks act as radio
// buttons, and a click on the row chooses the same option as the check.
func (p *prefsWindow) closeBehaviorGroup() *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle(closeBehaviorGroupTitle)
	group.SetDescription(closeBehaviorGroupHint)

	p.checks = make([]*gtk.CheckButton, 0, len(closeBehaviorChoices))
	for _, choice := range closeBehaviorChoices {
		check := gtk.NewCheckButton()
		if len(p.checks) > 0 {
			check.SetGroup(p.checks[0])
		}
		check.SetActive(choice.behavior == p.behavior)
		check.SetVAlign(gtk.AlignCenter)

		behavior := choice.behavior
		check.ConnectToggled(func() {
			if check.Active() {
				p.behavior = behavior
			}
		})

		row := adw.NewActionRow()
		row.SetTitle(choice.label)
		row.AddSuffix(check)
		row.SetActivatableWidget(check)
		group.Add(row)

		p.checks = append(p.checks, check)
	}

	return group
}

// terminalGroup is the group of the terminal command.
func (p *prefsWindow) terminalGroup() *adw.PreferencesGroup {
	group := adw.NewPreferencesGroup()
	group.SetTitle(terminalGroupTitle)
	group.SetDescription(terminalGroupHint)

	p.terminal = adw.NewEntryRow()
	p.terminal.SetTitle(terminalRowTitle)
	group.Add(p.terminal)

	return group
}

// load puts the saved settings in the widgets. The read runs off the GTK loop.
func (p *prefsWindow) load() {
	if p.app.store == nil {
		return
	}
	ctx := p.app.ctx
	go func() {
		cfg, err := p.app.store.Get(ctx)
		glib.IdleAdd(func() {
			if err != nil {
				p.toast(UserError(err).Error())
				return
			}
			p.show(cfg)
		})
	}()
}

// show puts cfg in the widgets. The check of the saved behavior is the checked
// one, so the window never shows a choice that is not the saved one.
func (p *prefsWindow) show(cfg settings.Settings) {
	p.behavior = cfg.CloseBehavior
	index := closeBehaviorIndex(cfg.CloseBehavior)
	for i, check := range p.checks {
		if i == index {
			check.SetActive(true)
		}
	}
	p.terminal.SetText(cfg.TerminalCommand)
}

// save stores the settings of the window. The write runs off the GTK loop and
// the outcome comes back as a toast: the refusal of an invalid terminal command
// keeps the window open.
func (p *prefsWindow) save() {
	if p.app.store == nil {
		return
	}
	value := settings.Settings{
		CloseBehavior:   p.behavior,
		TerminalCommand: p.terminal.Text(),
	}
	ctx := p.app.ctx
	store := p.app.store
	go func() {
		err := store.Save(ctx, value)
		message, saved := preferencesSaveOutcome(err)
		glib.IdleAdd(func() {
			p.toast(message)
			if saved {
				// Show what the database holds: a blank command is stored as
				// the default one.
				p.load()
			}
		})
	}()
}

// toast shows a short message over the window.
func (p *prefsWindow) toast(message string) {
	toast := adw.NewToast(message)
	toast.SetTimeout(toastTimeout)
	p.window.AddToast(toast)
}
