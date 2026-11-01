package gui

import (
	"fmt"

	"github.com/mappu/miqt/qt6"

	"dockercontroller/internal/compose"
)

// rowKind says what a row of the list holds.
type rowKind int

const (
	// rowHeading is the title of a Group: it is neither selectable nor
	// activatable.
	rowHeading rowKind = iota
	// rowProject is a saved Project.
	rowProject
	// rowEmpty is the line of an empty list.
	rowEmpty
)

// rowHeight sizes a row of the list. Every row is one line: the Display name,
// or the compose name when it is missing, the count and the tacks.
const rowHeight = 46

// The object names of the row: the style sheet colours the labels of a selected
// row with the highlighted text, so the blue band keeps its contrast.
const (
	rowObjectName         = "projectRow"
	rowSelectedObjectName = "projectRowSelected"
)

// projectList is the column of the Projects: one section per Group, with the
// Projects without a Group at the end, the marker of the state and the tacks of
// the Containers. It reads the snapshot and calls onOpen when the user chooses
// a Project; it never changes one.
type projectList struct {
	widget *qt6.QListWidget
	// theme supplies the colours of the roles the style sheet does not reach,
	// like the tack of a faulty Container.
	theme designTheme
	// rows holds the Project of each row. A heading and the empty line carry
	// the zero Project, so the row of a Project is found without a map keyed on
	// the wrapper.
	rows  []compose.Project
	kinds []rowKind
	// strips holds the painted tacks of each row, nil for a heading and for the
	// line of an empty list.
	strips []*tackStrip

	onOpen func(compose.Project)
}

// newProjectList builds the empty list. The rows arrive with set.
func newProjectList(theme designTheme) *projectList {
	p := &projectList{widget: qt6.NewQListWidget2(), theme: theme}
	p.widget.SetSelectionMode(qt6.QAbstractItemView__SingleSelection)
	// Every row has the same height, so the list can size its viewport once.
	p.widget.SetUniformItemSizes(true)
	p.widget.SetSpacing(0)
	p.widget.SetAccessibleName("Progetti")
	// The click, not the change of the current row, opens the Project: a click
	// on the row already current must open it again after the narrow view went
	// back to the list.
	p.widget.OnItemClicked(p.choose)
	p.widget.OnItemActivated(p.choose)
	// The selected row paints its own labels and tacks, because a real widget
	// in a row does not take the colours of the item it sits in.
	p.widget.OnCurrentRowChanged(func(int) { p.markSelection() })
	return p
}

// choose opens the Project of the chosen row. A heading and the empty line are
// not selectable, so they never reach here.
func (p *projectList) choose(item *qt6.QListWidgetItem) {
	if item == nil {
		return
	}
	row := p.widget.Row(item)
	if row < 0 || row >= len(p.rows) || p.kinds[row] != rowProject {
		return
	}
	if p.onOpen != nil {
		p.onOpen(p.rows[row])
	}
}

// set replaces the rows with the Projects of the snapshot and keeps the open
// Project selected. An empty list with no load in flight says that no Project is
// saved; a list that is still loading stays blank.
func (p *projectList) set(projects []compose.Project, selected string, loading bool) {
	p.widget.Clear()
	p.rows = nil
	p.kinds = nil
	p.strips = nil

	shown := 0
	for _, section := range Sections(projects) {
		if len(section.Projects) == 0 {
			continue
		}
		if section.Title != "" {
			p.appendHeading(section.Title)
		}
		for _, project := range section.Projects {
			p.appendProject(project)
			shown++
		}
	}

	if shown == 0 {
		if loading {
			p.appendLineText(loadingListText)
		} else {
			p.appendLineText(emptyListText)
		}
	}

	if selected != "" {
		p.selectProject(selected)
	}
	p.markSelection()
}

// markSelection gives the selected row the colours of the highlight and the
// other rows the colours of the theme. A widget placed in a row does not take
// the palette of the item, so the row is painted on its own: the style sheet
// switches on the object name and the tacks are painted from the palette.
func (p *projectList) markSelection() {
	if p.widget == nil {
		return
	}
	palette := p.widget.Palette()
	if palette == nil {
		return
	}
	current := p.widget.CurrentRow()

	for i, kind := range p.kinds {
		if kind != rowProject {
			continue
		}
		item := p.widget.Item(i)
		if item == nil {
			continue
		}
		row := p.widget.ItemWidget(item)
		if row == nil {
			continue
		}
		name := rowObjectName
		if i == current {
			name = rowSelectedObjectName
		}
		if row.ObjectName() != name {
			setObjectName(row, name)
			row.Style().Unpolish(row)
			row.Style().Polish(row)
		}
		p.strips[i].setSelected(i == current)
	}
}

// selectProject selects the row of a Project, if it is still in the list.
func (p *projectList) selectProject(name string) {
	for i, project := range p.rows {
		if p.kinds[i] == rowProject && project.Name == name {
			p.widget.SetCurrentRow(i)
			return
		}
	}
	p.widget.SetCurrentRow(-1)
}

// appendHeading adds the title of a Group.
func (p *projectList) appendHeading(title string) {
	label := qt6.NewQLabel3(title)
	setObjectName(label.QWidget, "listHeading")
	label.SetAttribute2(qt6.WA_TransparentForMouseEvents, true)

	p.appendRow(rowHeading, compose.Project{}, label.QWidget, nil, rowHeight)
}

// appendLineText adds a line of text that is not a Project: the state of an
// empty list, or the wait for the first snapshot.
func (p *projectList) appendLineText(text string) {
	label := qt6.NewQLabel3(text)
	setObjectName(label.QWidget, "listEmpty")
	label.SetAttribute2(qt6.WA_TransparentForMouseEvents, true)

	p.appendRow(rowEmpty, compose.Project{}, label.QWidget, nil, rowHeight)
}

// appendProject adds the row of one Project.
func (p *projectList) appendProject(project compose.Project) {
	row, strip := newProjectRow(project, p.theme)
	p.appendRow(rowProject, project, row, strip, rowHeight)
}

// appendRow adds a row with its widget and its tooltip.
func (p *projectList) appendRow(kind rowKind, project compose.Project, widget *qt6.QWidget, strip *tackStrip, height int) {
	item := qt6.NewQListWidgetItem2("")
	item.SetSizeHint(qt6.NewQSize2(0, height))
	if kind != rowProject {
		item.SetFlags(qt6.NoItemFlags)
	}
	if kind == rowProject {
		status := StatusOf(project.Status)
		// The row is one line, so the compose name lives here when the Display
		// name hides it.
		name := ProjectName(project)
		if project.Name != name {
			name = fmt.Sprintf("%s (%s)", name, project.Name)
		}
		item.SetToolTip(fmt.Sprintf("%s, %s, %d Container in esecuzione su %d",
			name, status.Label, project.Running, project.Total))
	}

	p.widget.AddItemWithItem(item)
	p.widget.SetItemWidget(item, widget)
	p.rows = append(p.rows, project)
	p.kinds = append(p.kinds, kind)
	p.strips = append(p.strips, strip)
}

// newProjectRow builds the content of the row of one Project on one line: the
// marker of the state, the name, the count and the tacks. The compose name
// stays in the tooltip and in the detail, so the row never wraps. The strip of
// the tacks comes back with the row: it is painted, not styled, so the
// selection has to reach it in person.
func newProjectRow(project compose.Project, theme designTheme) (*qt6.QWidget, *tackStrip) {
	status := StatusOf(project.Status)

	row := qt6.NewQWidget2()
	setObjectName(row, rowObjectName)
	row.SetAttribute2(qt6.WA_TransparentForMouseEvents, true)

	layout := qt6.NewQHBoxLayout(row)
	layout.SetContentsMargins(12, 6, 12, 6)
	layout.SetSpacing(10)

	marker := qt6.NewQLabel3(status.Glyph)
	if name := statusObjectName(status); name != "" {
		setObjectName(marker.QWidget, name)
	}
	marker.SetFixedWidth(14)
	marker.SetToolTip(status.Label)
	marker.SetAccessibleName(status.Label)
	layout.AddWidget3(marker.QWidget, 0, qt6.AlignVCenter)

	name := qt6.NewQLabel3(ProjectName(project))
	setObjectName(name.QWidget, "projectName")
	layout.AddWidget3(name.QWidget, 1, qt6.AlignVCenter)

	count := qt6.NewQLabel3(ProjectCount(project))
	setObjectName(count.QWidget, "projectCount")
	layout.AddWidget3(count.QWidget, 0, qt6.AlignVCenter)

	tacks, more := Tacks(project.Services)
	strip := qt6.NewQHBoxLayout2()
	strip.SetSpacing(2)
	bar := newTackStrip(tacks, theme)
	strip.AddWidget(bar.QWidget)
	if more > 0 {
		moreLabel := qt6.NewQLabel3(fmt.Sprintf("+%d", more))
		setObjectName(moreLabel.QWidget, "tacksMore")
		strip.AddWidget(moreLabel.QWidget)
	}
	layout.AddLayout2(strip.QLayout, 0)

	return row, bar
}

// statusObjectName is the name the style sheet uses for the marker of a state.
// The classes of StatusView carry the meaning; only the two states the palette
// cannot express take a colour, and the others keep the colour of the theme.
func statusObjectName(status StatusView) string {
	switch status.CSSClass {
	case "success":
		return "statusActive"
	case "warning":
		return "statusPartial"
	}
	return ""
}

// setObjectName names a widget, so the style sheet can select it. The binding
// takes a QAnyStringView, which allocates on every call.
func setObjectName(widget *qt6.QWidget, name string) {
	widget.SetObjectName(*qt6.NewQAnyStringView3(name))
}
