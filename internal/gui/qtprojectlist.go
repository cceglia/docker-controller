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

// rowHeight and rowHeightTwoLines size a row of the list: a Project with a
// Display name that differs from the compose name needs two lines.
const (
	rowHeight         = 46
	rowHeightTwoLines = 62
)

// projectList is the column of the Projects: one section per Group, with the
// Projects without a Group at the end, the marker of the state and the tacks of
// the Containers. It reads the snapshot and calls onOpen when the user chooses
// a Project; it never changes one.
type projectList struct {
	widget *qt6.QListWidget
	// rows holds the Project of each row. A heading and the empty line carry
	// the zero Project, so the row of a Project is found without a map keyed on
	// the wrapper.
	rows  []compose.Project
	kinds []rowKind

	onOpen func(compose.Project)
}

// newProjectList builds the empty list. The rows arrive with set.
func newProjectList() *projectList {
	p := &projectList{widget: qt6.NewQListWidget2()}
	p.widget.SetSelectionMode(qt6.QAbstractItemView__SingleSelection)
	// A Project with a Display name that differs from the compose name takes
	// two lines, so the rows do not share one size and the list must not assume
	// they do.
	p.widget.SetUniformItemSizes(false)
	p.widget.SetSpacing(0)
	p.widget.SetAccessibleName("Progetti")
	// The click, not the change of the current row, opens the Project: a click
	// on the row already current must open it again after the narrow view went
	// back to the list.
	p.widget.OnItemClicked(p.choose)
	p.widget.OnItemActivated(p.choose)
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

	if shown == 0 && !loading {
		p.appendEmpty()
	}

	if selected != "" {
		p.selectProject(selected)
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

	p.appendRow(rowHeading, compose.Project{}, label.QWidget, rowHeight)
}

// appendEmpty adds the line of an empty list.
func (p *projectList) appendEmpty() {
	label := qt6.NewQLabel3(emptyListText)
	setObjectName(label.QWidget, "listEmpty")
	label.SetAttribute2(qt6.WA_TransparentForMouseEvents, true)

	p.appendRow(rowEmpty, compose.Project{}, label.QWidget, rowHeight)
}

// appendProject adds the row of one Project.
func (p *projectList) appendProject(project compose.Project) {
	height := rowHeight
	if project.DisplayName != "" && project.DisplayName != project.Name {
		height = rowHeightTwoLines
	}
	p.appendRow(rowProject, project, newProjectRow(project), height)
}

// appendRow adds a row with its widget and its tooltip.
func (p *projectList) appendRow(kind rowKind, project compose.Project, widget *qt6.QWidget, height int) {
	item := qt6.NewQListWidgetItem2("")
	item.SetSizeHint(qt6.NewQSize2(0, height))
	if kind != rowProject {
		item.SetFlags(qt6.NoItemFlags)
	}
	if kind == rowProject {
		status := StatusOf(project.Status)
		item.SetToolTip(fmt.Sprintf("%s, %s, %d Container in esecuzione su %d",
			ProjectName(project), status.Label, project.Running, project.Total))
	}

	p.widget.AddItemWithItem(item)
	p.widget.SetItemWidget(item, widget)
	p.rows = append(p.rows, project)
	p.kinds = append(p.kinds, kind)
}

// newProjectRow builds the content of the row of one Project: the marker of the
// state, the name, the compose name when it differs, the count and the tacks.
func newProjectRow(project compose.Project) *qt6.QWidget {
	status := StatusOf(project.Status)

	row := qt6.NewQWidget2()
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

	text := qt6.NewQVBoxLayout2()
	text.SetSpacing(0)

	name := qt6.NewQLabel3(ProjectName(project))
	setObjectName(name.QWidget, "projectName")
	text.AddWidget(name.QWidget)

	if project.DisplayName != "" && project.DisplayName != project.Name {
		composeName := qt6.NewQLabel3(project.Name)
		setObjectName(composeName.QWidget, "composeName")
		text.AddWidget(composeName.QWidget)
	}
	layout.AddLayout2(text.QLayout, 1)

	count := qt6.NewQLabel3(ProjectCount(project))
	setObjectName(count.QWidget, "projectCount")
	layout.AddWidget3(count.QWidget, 0, qt6.AlignVCenter)

	tacks, more := Tacks(project.Services)
	strip := qt6.NewQHBoxLayout2()
	strip.SetSpacing(2)
	strip.AddWidget(newTackStrip(tacks))
	if more > 0 {
		moreLabel := qt6.NewQLabel3(fmt.Sprintf("+%d", more))
		setObjectName(moreLabel.QWidget, "tacksMore")
		strip.AddWidget(moreLabel.QWidget)
	}
	layout.AddLayout2(strip.QLayout, 0)

	return row
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
