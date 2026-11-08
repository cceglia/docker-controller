package design

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The tests read the same tokens the style sheet is built from. A role with no
// contrast kind, a colour in the sheet that no role declares and a font size
// outside the scale all fail here.

var (
	hexColorPattern   = regexp.MustCompile(`#[0-9a-f]{6}`)
	fontSizePattern   = regexp.MustCompile(`font-size:\s*(\d+)px`)
	paddingPattern    = regexp.MustCompile(`padding:\s*([^;]+);`)
	paddingFieldRegex = regexp.MustCompile(`\d+`)
)

func TestThemeRolesMeetContrast(t *testing.T) {
	for _, theme := range themes {
		t.Run(theme.name, func(t *testing.T) {
			if theme.surface == "" {
				t.Fatal("the theme declares no surface: no contrast can be measured")
			}
			if len(theme.roles) == 0 {
				t.Fatal("the theme declares no role")
			}
			for name, role := range theme.roles {
				name, role := name, role
				t.Run(name, func(t *testing.T) {
					if !isHexColor(role.color) {
						t.Fatalf("role %q: colour %q is not #rrggbb", name, role.color)
					}
					threshold := 0.0
					switch role.kind {
					case contrastText:
						threshold = 4.5
					case contrastComponent:
						threshold = 3.0
					default:
						t.Fatalf("role %q: no contrast kind, so no check", name)
					}
					if got := contrastRatio(role.color, theme.surface); got < threshold {
						t.Errorf("role %q: %s on %s is %.2f:1, want at least %.1f:1",
							name, role.color, theme.surface, got, threshold)
					}
				})
			}
		})
	}
}

// TestServiceColoursComeFromTheRoles reads the colours the log page paints a
// Service name with: one role per palette entry, in both themes, with the
// contrast of a text.
func TestServiceColoursComeFromTheRoles(t *testing.T) {
	for _, theme := range themes {
		theme := theme
		t.Run(theme.name, func(t *testing.T) {
			for hue := 0; hue < ServiceHues; hue++ {
				role, ok := theme.roles[serviceRole(hue)]
				if !ok {
					t.Fatalf("hue %d has no role in the %s theme", hue, theme.name)
				}
				if !isHexColor(role.color) {
					t.Fatalf("hue %d: colour %q is not #rrggbb", hue, role.color)
				}
				if role.kind != contrastText {
					t.Errorf("hue %d: kind %d, want a text colour", hue, role.kind)
				}
				if got := contrastRatio(role.color, theme.surface); got < minServiceContrast {
					t.Errorf("hue %d: %s on %s is %.2f:1, want at least %.1f:1",
						hue, role.color, theme.surface, got, minServiceContrast)
				}
				if got := ServiceColor(hue, theme); got != role.color {
					t.Errorf("ServiceColor(%d, %s) = %q, want the role colour %q", hue, theme.name, got, role.color)
				}
			}
		})
	}

	t.Run("an entry outside the palette has no colour", func(t *testing.T) {
		for _, hue := range []int{-1, ServiceHues} {
			if got := ServiceColor(hue, LightTheme); got != "" {
				t.Errorf("ServiceColor(%d) = %q, want no colour", hue, got)
			}
		}
	})

	t.Run("the two themes differ on every entry", func(t *testing.T) {
		for hue := 0; hue < ServiceHues; hue++ {
			if ServiceColor(hue, LightTheme) == ServiceColor(hue, DarkTheme) {
				t.Errorf("hue %d is the same colour on both themes", hue)
			}
		}
	})

	t.Run("the entries of a theme differ from each other", func(t *testing.T) {
		for _, theme := range themes {
			seen := make(map[string]int, ServiceHues)
			for hue := 0; hue < ServiceHues; hue++ {
				colour := ServiceColor(hue, theme)
				if other, duplicate := seen[colour]; duplicate {
					t.Errorf("%s: hues %d and %d share %s", theme.name, other, hue, colour)
				}
				seen[colour] = hue
			}
		}
	})
}

func TestStyleSheetColoursComeFromTheRoles(t *testing.T) {
	for _, theme := range themes {
		t.Run(theme.name, func(t *testing.T) {
			sheet := ApplicationStyleSheet(theme, "")
			found := hexColorPattern.FindAllString(sheet, -1)
			if len(found) == 0 {
				t.Fatal("the style sheet declares no colour: the sheet is not built from the tokens")
			}
			known := make(map[string]bool, len(theme.roles))
			for _, role := range theme.roles {
				known[role.color] = true
			}
			for _, hex := range found {
				if !known[hex] {
					t.Errorf("colour %s in the style sheet belongs to no role", hex)
				}
			}
		})
	}
}

// TestStyleSheetTakesTheSelectionOfTheDesktop guards the one colour of the
// sheet that comes from the palette: a desktop that hands a selection colour
// gets it on the focused row and on the inactive one, so the band does not
// change when the window loses focus.
func TestStyleSheetTakesTheSelectionOfTheDesktop(t *testing.T) {
	const selection = "#1b4155"
	for _, theme := range themes {
		theme := theme
		t.Run(theme.name, func(t *testing.T) {
			sheet := ApplicationStyleSheet(theme, selection)
			for _, rule := range []string{
				"QListWidget::item:selected {\n    background: " + selection + ";",
				"QListWidget::item:selected:!active {\n    background: " + selection + ";",
			} {
				if !strings.Contains(sheet, rule) {
					t.Errorf("the sheet has no rule %q", rule)
				}
			}
		})
	}

	t.Run("an empty selection keeps the highlight of the palette", func(t *testing.T) {
		sheet := ApplicationStyleSheet(DarkTheme, "")
		for _, rule := range []string{
			"QListWidget::item:selected {\n    background: palette(highlight);",
			"QListWidget::item:selected:!active {\n    background: palette(highlight);",
		} {
			if !strings.Contains(sheet, rule) {
				t.Errorf("the sheet has no rule %q", rule)
			}
		}
	})
}

func TestTypeScale(t *testing.T) {
	if bodyFontSize < 16 {
		t.Fatalf("body font size = %d px, want at least 16 px", bodyFontSize)
	}
	steps := []int{
		bodyFontSize,
		int(math.Round(float64(bodyFontSize) * scaleRatio)),
		int(math.Round(float64(bodyFontSize) * math.Pow(scaleRatio, 2))),
	}
	if headingFontSize != steps[1] {
		t.Errorf("heading font size = %d px, want %d px from the ratio", headingFontSize, steps[1])
	}
	if titleFontSize != steps[2] {
		t.Errorf("title font size = %d px, want %d px from the ratio", titleFontSize, steps[2])
	}

	inScale := make(map[int]bool, len(steps))
	for _, step := range steps {
		inScale[step] = true
	}
	for _, theme := range themes {
		chooser := theme
		t.Run("sheet/"+chooser.name, func(t *testing.T) {
			sizes := fontSizePattern.FindAllStringSubmatch(ApplicationStyleSheet(chooser, ""), -1)
			if len(sizes) == 0 {
				t.Fatal("the style sheet declares no font size")
			}
			for _, match := range sizes {
				px, err := strconv.Atoi(match[1])
				if err != nil {
					t.Fatalf("font size %q is not a number", match[1])
				}
				if px < bodyFontSize {
					t.Errorf("font size %d px is under the body minimum of %d px", px, bodyFontSize)
				}
				if !inScale[px] {
					t.Errorf("font size %d px belongs to no step of the scale", px)
				}
			}
		})
	}
}

func TestStyleSheetPaddingComesFromTheSpacingScale(t *testing.T) {
	inScale := make(map[int]bool, len(spacingSteps))
	for _, step := range spacingSteps {
		inScale[step] = true
	}
	for _, theme := range themes {
		chooser := theme
		t.Run(chooser.name, func(t *testing.T) {
			paddings := paddingPattern.FindAllStringSubmatch(ApplicationStyleSheet(chooser, ""), -1)
			if len(paddings) == 0 {
				t.Fatal("the style sheet declares no padding")
			}
			for _, match := range paddings {
				for _, field := range paddingFieldRegex.FindAllString(match[1], -1) {
					px, err := strconv.Atoi(field)
					if err != nil {
						t.Fatalf("padding %q is not a number", field)
					}
					if px != 0 && !inScale[px] {
						t.Errorf("padding %d px belongs to no step of the spacing scale", px)
					}
				}
			}
		})
	}
}

func TestThemeForLightness(t *testing.T) {
	tests := []struct {
		name      string
		lightness int
		want      DesignTheme
	}{
		{name: "light surface", lightness: 255, want: LightTheme},
		{name: "dark surface", lightness: 0, want: DarkTheme},
		{name: "middle reads as light", lightness: 128, want: LightTheme},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ThemeForLightness(tt.lightness); got.name != tt.want.name {
				t.Errorf("ThemeForLightness(%d) = %q, want %q", tt.lightness, got.name, tt.want.name)
			}
		})
	}
}

// TestPrimaryActionShowsItsStates guards the one control the sheet redraws: a
// background of its own takes the native focus and hover away, so the sheet
// must draw them back.
func TestPrimaryActionShowsItsStates(t *testing.T) {
	for _, theme := range themes {
		sheet := ApplicationStyleSheet(theme, "")
		for _, state := range []string{
			"QPushButton#primaryAction:hover",
			"QPushButton#primaryAction:focus",
			"QPushButton#primaryAction:pressed",
		} {
			if !strings.Contains(sheet, state) {
				t.Errorf("%s: the sheet has no rule %s", theme.name, state)
			}
		}
	}
}

// isHexColor reports whether color is a "#rrggbb" colour.
func isHexColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	for _, c := range color[1:] {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}

// contrastRatio is the WCAG contrast between two "#rrggbb" colours.
func contrastRatio(a, b string) float64 {
	la, lb := relativeLuminance(a), relativeLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// relativeLuminance is the WCAG relative luminance of a "#rrggbb" colour.
func relativeLuminance(color string) float64 {
	channel := func(i int) float64 {
		v, err := strconv.ParseUint(color[1+i*2:3+i*2], 16, 8)
		if err != nil {
			panic(err)
		}
		c := float64(v) / 255
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(0) + 0.7152*channel(1) + 0.0722*channel(2)
}
