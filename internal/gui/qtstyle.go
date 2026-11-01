package gui

// The style sheet covers only the roles the desktop theme does not carry: the
// marker of the state of a Project, the headings of the Groups and the size of
// the text of a row. Every other colour comes from the palette of the theme,
// so a window on KDE reads as the rest of the system.
//
// The type scale uses the ratio 1.125 with a body of 16 px: 18 px for a heading
// and 20 px for the title of the detail. Secondary text keeps the body size and
// the colour of the theme: a smaller size would break the minimum of 16 px, and
// a muted role does not guarantee the contrast a text needs.
//
// The two colours of state are declared here, next to their role. The marker is
// never the only cue: its shape differs and its accessible name says the state
// in words.
const applicationStyleSheet = `
QLabel#sidebarTitle {
    font-size: 18px;
    font-weight: 600;
    padding: 0 0 8px 0;
}
QLabel#listHeading {
    font-size: 18px;
    font-weight: 600;
    padding: 12px 12px 4px 12px;
}
QLabel#listEmpty {
    padding: 12px;
}
QLabel#projectName {
    font-size: 16px;
}
QLabel#composeName {
    font-size: 16px;
}
QLabel#projectCount {
    font-size: 16px;
}
QLabel#tacksMore {
    font-size: 16px;
}
QLabel#placeholderTitle {
    font-size: 20px;
    font-weight: 600;
}
QLabel#statusActive {
    color: #2f9e44;
}
QLabel#statusPartial {
    color: #b8860b;
}
`
