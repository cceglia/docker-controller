package gui

import (
	"math"
	"regexp"
	"strconv"
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

func TestStyleSheetColoursComeFromTheRoles(t *testing.T) {
	for _, theme := range themes {
		t.Run(theme.name, func(t *testing.T) {
			sheet := applicationStyleSheet(theme)
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
			sizes := fontSizePattern.FindAllStringSubmatch(applicationStyleSheet(chooser), -1)
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
			paddings := paddingPattern.FindAllStringSubmatch(applicationStyleSheet(chooser), -1)
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
		want      designTheme
	}{
		{name: "light surface", lightness: 255, want: lightTheme},
		{name: "dark surface", lightness: 0, want: darkTheme},
		{name: "middle reads as light", lightness: 128, want: lightTheme},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := themeForLightness(tt.lightness); got.name != tt.want.name {
				t.Errorf("themeForLightness(%d) = %q, want %q", tt.lightness, got.name, tt.want.name)
			}
		})
	}
}
