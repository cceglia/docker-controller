package gui

// appID names the application, its desktop entry and its window class. Qt
// derives the class of the window from the application name, so the two must
// stay equal to the StartupWMClass of the desktop entry.
const appID = "io.github.cceglia.DockerController"

const (
	// windowTitle is the title of the window and of the notifications.
	windowTitle = "Docker Controller"
	// windowWidth and windowHeight are the size the window opens at.
	windowWidth  = 1120
	windowHeight = 760
	// breakpointWidth is the width under which the window shows one page at a
	// time, in scale independent pixels.
	breakpointWidth = 900
	// emptyListText is the line of the list when no Project is saved.
	emptyListText = "Nessun Progetto."
)
