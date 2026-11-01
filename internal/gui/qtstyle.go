package gui

import "fmt"

// The style sheet covers only the roles the desktop theme does not carry: the
// marker of the state of a Project, the headings of the Groups and the size of
// the text of a row. Every other colour comes from the palette of the theme,
// so a window on KDE reads as the rest of the system.
//
// The sheet is built from the tokens in designtokens.go: the type scale, the
// colours of the roles and the spacing rhythm. No colour is written here, so a
// change of the palette is one edit in one place.
//
// The type scale uses the ratio 1.125 with a body of 16 px: 18 px for a heading
// and 20 px for the title of the detail. Secondary text keeps the body size and
// the colour of the theme: a smaller size would break the minimum of 16 px, and
// a muted role does not guarantee the contrast a text needs.
func applicationStyleSheet(theme designTheme) string {
	return fmt.Sprintf(`
QLabel#sidebarTitle {
    font-size: %dpx;
    font-weight: 600;
    padding: 0 0 %dpx 0;
}
QLabel#listHeading {
    font-size: %dpx;
    font-weight: 600;
    padding: %dpx %dpx %dpx %dpx;
}
QLabel#listEmpty {
    padding: %dpx;
}
QLabel#projectName {
    font-size: %dpx;
}
QLabel#composeName {
    font-size: %dpx;
}
QLabel#projectCount {
    font-size: %dpx;
}
QLabel#tacksMore {
    font-size: %dpx;
}
QLabel#placeholderTitle {
    font-size: %dpx;
    font-weight: 600;
}
QLabel#statusActive {
    color: %s;
}
QLabel#statusPartial {
    color: %s;
}
`,
		headingFontSize, spaceS,
		headingFontSize, spaceM, spaceM, spaceXS, spaceM,
		spaceM,
		bodyFontSize,
		bodyFontSize,
		bodyFontSize,
		bodyFontSize,
		titleFontSize,
		theme.color(roleStatusActive),
		theme.color(roleStatusPartial),
	)
}
