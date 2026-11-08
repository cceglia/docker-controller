package design

import "strconv"

// This file is the single source of the design tokens: the type scale, the
// colours of the semantic roles, the spacing rhythm and the minimum target of
// an interactive control. The style sheet in qtstyle.go is built from these
// values, and the token test reads the same values. A colour belongs to a role
// or it does not exist.

// contrastKind is the WCAG threshold a colour of a role must meet.
type contrastKind int

const (
	// contrastText is a colour of text: the ratio is 4.5:1.
	contrastText contrastKind = iota + 1
	// contrastComponent is a colour of a component or a border: the ratio is
	// 3:1.
	contrastComponent
)

// designRole is one semantic colour. The kind decides the check the token test
// applies, so a role without a kind fails the test.
type designRole struct {
	color string
	kind  contrastKind
}

// DesignTheme is one palette: the surface the roles are drawn on and the roles
// themselves. The desktop theme picks the palette at startup.
type DesignTheme struct {
	name    string
	surface string
	roles   map[string]designRole
}

// The role names. The style sheet selects on the first three, the tacks of a
// Project paint with the fourth, and the log page reads the last two and the
// colours of the Services.
const (
	roleStatusActive  = "statusActive"
	roleStatusPartial = "statusPartial"
	// RoleStatusFault is the role of a fault: the text of a failed state and
	// the output of compose itself.
	RoleStatusFault = "statusFault"
	// RoleTackFault is the role of the tack of a faulty Container.
	RoleTackFault = "tackFault"
	// RoleLogTimestamp is the role of the clock of a line of the log.
	RoleLogTimestamp = "logTimestamp"
	// serviceRolePrefix names the role of one entry of the Service palette.
	serviceRolePrefix = "service"
)

// ServiceHues is how many colours the Service palette has: one role per entry,
// so a Service name never moves colour between two runs.
const ServiceHues = 16

// minServiceContrast is the WCAG contrast a Service colour keeps against the
// surface of either theme. The name of a Service is bold, so 4.5 (the level AA
// of normal text) is the floor.
const minServiceContrast = 4.5

// servicePaletteLight and servicePaletteDark hold one colour per hue: dark
// enough to read on the light surface, light enough to read on the dark one.
// Each hue is 22.5 degrees of the colour wheel at a fixed saturation.
var (
	servicePaletteLight = [ServiceHues]string{
		"#721d1d", "#723d1d", "#725d1d", "#68721d",
		"#47721d", "#27721d", "#1d7232", "#1d7252",
		"#1d7272", "#1d5272", "#1d3272", "#271d72",
		"#471d72", "#681d72", "#721d5d", "#721d3d",
	}
	servicePaletteDark = [ServiceHues]string{
		"#e08585", "#e0a785", "#e0c985", "#d5e085",
		"#b3e085", "#90e085", "#85e09c", "#85e0be",
		"#85e0e0", "#85bee0", "#859ce0", "#9085e0",
		"#b385e0", "#d585e0", "#e085c9", "#e085a7",
	}
)

// serviceRole is the role of one entry of the Service palette.
func serviceRole(hue int) string {
	return serviceRolePrefix + strconv.Itoa(hue)
}

// withServiceRoles adds the entries of a Service palette to the roles of a
// theme. Every entry is a text colour: the token test reads the same values.
func withServiceRoles(roles map[string]designRole, palette [ServiceHues]string) map[string]designRole {
	for hue, colour := range palette {
		roles[serviceRole(hue)] = designRole{color: colour, kind: contrastText}
	}
	return roles
}

// StateRole is the role of the style sheet that carries a state of the old
// view: success reads as Attivo, warning as Parziale and error as the fault of
// a Container. A quiet state has no role and keeps the colour of the theme.
func StateRole(cssClass string) string {
	switch cssClass {
	case "success":
		return roleStatusActive
	case "warning":
		return roleStatusPartial
	case "error":
		return RoleStatusFault
	}
	return ""
}

// LightTheme is the palette of a light desktop. The surface is the base colour
// of the list, where the labels of state sit.
var LightTheme = DesignTheme{
	name:    "light",
	surface: "#ffffff",
	roles: withServiceRoles(map[string]designRole{
		// The green and the amber of the two states were 3.4:1 and 3.3:1 on
		// white: too faint for text. These values stay above 4.5:1.
		roleStatusActive:  {color: "#1a7f37", kind: contrastText},
		roleStatusPartial: {color: "#8a6508", kind: contrastText},
		// The fault reads as text too: the state of a Container carries a
		// label, so the red must pass the 4.5:1 of text, not the 3:1 of a tack.
		RoleStatusFault: {color: "#b42318", kind: contrastText},
		// The tack of a faulty Container is a component, not text.
		RoleTackFault: {color: "#e05561", kind: contrastComponent},
		// The clock of a log line is secondary text: it stays above the
		// 4.5:1 of a normal text, because a smaller role does not exist.
		RoleLogTimestamp: {color: "#5c5c5c", kind: contrastText},
	}, servicePaletteLight),
}

// DarkTheme is the palette of a dark desktop. The surface is the base colour of
// Breeze Dark. A colour that passes on white is usually too dark here, so the
// two palettes declare their own values.
var DarkTheme = DesignTheme{
	name:    "dark",
	surface: "#232629",
	roles: withServiceRoles(map[string]designRole{
		roleStatusActive:  {color: "#51b564", kind: contrastText},
		roleStatusPartial: {color: "#e0a800", kind: contrastText},
		RoleStatusFault:   {color: "#ff8882", kind: contrastText},
		RoleTackFault:     {color: "#e05561", kind: contrastComponent},
		RoleLogTimestamp:  {color: "#b8bcbf", kind: contrastText},
	}, servicePaletteDark),
}

// themes lists the palettes the UI can apply. The token test checks every one.
var themes = []DesignTheme{LightTheme, DarkTheme}

// ThemeForLightness picks the palette from the lightness of the window colour
// of the desktop palette. Qt lightness runs from 0 to 255, so 128 is the middle.
func ThemeForLightness(lightness int) DesignTheme {
	if lightness < 128 {
		return DarkTheme
	}
	return LightTheme
}

// Color returns the value of a role. A role missing from the theme returns the
// empty string, and the token test fails on it.
func (t DesignTheme) Color(role string) string {
	return t.roles[role].color
}

// IsDark reports whether the palette is the dark one. The selection of the list
// is the one colour the sheet takes from the palette of the desktop, and a dark
// desktop paints a selected row with the darker of its two highlights.
func (t DesignTheme) IsDark() bool {
	return t.name == DarkTheme.name
}

// ServiceColor is the colour of one palette entry in the theme, as "#rrggbb".
// The colours live with the other design tokens, so the token test checks
// them. An entry outside the palette has no colour.
func ServiceColor(hue int, theme DesignTheme) string {
	if hue < 0 || hue >= ServiceHues {
		return ""
	}
	return theme.Color(serviceRole(hue))
}

// The type scale: the ratio 1.125 with a body of 16 px. The heading and the
// title are the next two steps, rounded. Body text never goes below 16 px.
const (
	scaleRatio      = 1.125
	bodyFontSize    = 16
	headingFontSize = 18
	titleFontSize   = 20
)

// The spacing rhythm: 4, 8, 12, 16 and 24 logical pixels. The style sheet and
// the layouts take their paddings and margins from these steps.
const (
	SpaceXS = 4
	SpaceS  = 8
	SpaceM  = 12
	SpaceL  = 16
	SpaceXL = 24
)

// spacingSteps is the scale the token test checks a padding against.
var spacingSteps = []int{SpaceXS, SpaceS, SpaceM, SpaceL, SpaceXL}

// TargetMin is the smallest side of an interactive control, in logical pixels:
// the minimum target of WCAG 2.2 SC 2.5.8.
const TargetMin = 24
