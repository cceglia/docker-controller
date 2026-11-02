package design

import "fmt"

// The style sheet covers only the roles the desktop theme does not carry: the
// markers of the state of a Project and of a Container, the headings of the
// Groups and of the detail, and the size of the text of a row. Every other
// colour comes from the palette of the theme, so a window on KDE reads as the
// rest of the system.
//
// The sheet is built from the tokens in designtokens.go: the type scale, the
// colours of the roles and the spacing rhythm. No colour is written here, so a
// change of the palette is one edit in one place.
//
// The type scale uses the ratio 1.125 with a body of 16 px: 18 px for a heading
// and 20 px for the title of the detail. Secondary text keeps the body size and
// the colour of the theme: a smaller size would break the minimum of 16 px, and
// a muted role does not guarantee the contrast a text needs.
func ApplicationStyleSheet(theme DesignTheme) string {
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
QLabel#detailTitle {
    font-size: %dpx;
    font-weight: 600;
}
QLabel#detailHeading {
    font-size: %dpx;
    font-weight: 600;
}
QLabel#serviceName {
    font-weight: 600;
}
QLabel#fileComposeName {
    font-weight: 600;
}
QLabel#dialogTitle {
    font-size: %dpx;
    font-weight: 600;
}
QFrame#toast {
    background: palette(base);
    border: 1px solid palette(mid);
    border-radius: %dpx;
}
QLabel#toastText {
    padding: %dpx %dpx;
}
QPlainTextEdit#activityOutput {
    padding: %dpx %dpx;
}
QPlainTextEdit#logView {
    padding: %dpx %dpx;
}
QLabel#logTarget {
    font-weight: 600;
}
QLabel#logCount {
    padding: 0 %dpx;
}
QLabel#logStatus {
    padding: 0 %dpx %dpx %dpx;
}
QLabel#logStatusFault {
    color: %s;
}
QCheckBox#logClocks {
    padding: %dpx 0;
}
QLabel#statusActive {
    color: %s;
}
QLabel#statusPartial {
    color: %s;
}
QLabel#statusFault {
    color: %s;
}
QLabel#fieldError {
    color: %s;
}
QLabel#fieldLabel {
    font-weight: 600;
}
QFrame#editPanel {
    background: palette(base);
    border: 1px solid palette(mid);
    border-radius: %dpx;
}
QPushButton#primaryAction {
    background: palette(highlight);
    color: palette(highlighted-text);
    border: 2px solid palette(highlight);
    /* The border belongs to the sheet, so the padding keeps the height of
       the buttons the desktop draws. */
    padding: 4px 12px;
    font-weight: 600;
}
QPushButton#primaryAction:hover {
    border-color: palette(highlighted-text);
}
QPushButton#primaryAction:focus {
    border: 2px solid palette(highlighted-text);
}
QPushButton#primaryAction:pressed {
    border-color: palette(mid);
}
QPushButton#primaryAction:disabled {
    background: palette(button);
    color: palette(mid);
    border-color: palette(mid);
    font-weight: 600;
}
QListWidget::item:selected {
    background: palette(highlight);
    color: palette(highlighted-text);
}
QListWidget::item:selected:!active {
    background: palette(highlight);
}
QWidget#projectRowSelected QLabel {
    color: palette(highlighted-text);
}
`,
		headingFontSize, SpaceS,
		headingFontSize, SpaceM, SpaceM, SpaceXS, SpaceM,
		SpaceM,
		bodyFontSize,
		bodyFontSize,
		bodyFontSize,
		titleFontSize,
		titleFontSize,
		headingFontSize,
		headingFontSize,
		SpaceXS,
		SpaceS, SpaceM,
		SpaceS, SpaceS,
		SpaceS, SpaceS,
		SpaceS,
		SpaceS, SpaceS, SpaceS,
		theme.Color(RoleStatusFault),
		SpaceXS,
		theme.Color(roleStatusActive),
		theme.Color(roleStatusPartial),
		theme.Color(RoleStatusFault),
		theme.Color(RoleStatusFault),
		SpaceS,
	)
}
