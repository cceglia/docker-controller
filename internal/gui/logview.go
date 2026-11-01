package gui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"dockercontroller/internal/compose"
)

// LogLineLimit is the most lines the Log page keeps; the oldest go first.
const LogLineLimit = 5000

// Tags the Log page applies to the parts of a line. They are not Service
// colours: stderr is the output of compose itself, not a Service.
const (
	// LogTimestampTag dims the clock of a line.
	LogTimestampTag = "log-timestamp"
	// LogStderrTag marks a message of compose itself: a different slant and
	// the error colour, never a Service colour.
	LogStderrTag = "log-stderr"
)

// serviceTagPrefix names the colour tag of a Service.
const serviceTagPrefix = "log-service-"

// ServiceHues is how many colours the Service palette has.
const ServiceHues = 16

// minServiceContrast is the WCAG contrast a Service colour keeps against the
// window background of Adwaita, light and dark. The Service name is bold, so
// 4.5 (the level AA of normal text) is the floor.
const minServiceContrast = 4.5

// servicePaletteLight and servicePaletteDark hold one colour per hue: dark
// enough to read on the light window, light enough to read on the dark one.
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

// LogSegment is a part of the text of one line with the tag the view applies
// to it. An empty Tag means no tag: the text of a Container has none.
type LogSegment struct {
	Tag  string
	Text string
}

// AppendLogLines appends incoming to lines and keeps only the last limit. The
// caller passes the slice it owns; the result may share its array.
func AppendLogLines(lines, incoming []compose.LogLine, limit int) []compose.LogLine {
	lines = append(lines, incoming...)
	if over := len(lines) - limit; over > 0 {
		n := copy(lines, lines[over:])
		lines = lines[:n]
	}
	return lines
}

// FilterLogLines returns the lines whose text contains term, ignoring case.
// It reads only the text: a Service name is not part of the search. An empty
// or blank term keeps every line.
func FilterLogLines(lines []compose.LogLine, term string) []compose.LogLine {
	term = strings.ToLower(strings.TrimSpace(term))
	if term == "" {
		return lines
	}
	out := make([]compose.LogLine, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line.Text), term) {
			out = append(out, line)
		}
	}
	return out
}

// LogTime is the clock of a line, "15:04:05" in local time. It is empty when
// the line carries no timestamp or when Docker wrote something else.
func LogTime(timestamp string) string {
	if timestamp == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return ""
	}
	return t.Local().Format("15:04:05")
}

// LogSegments is what the panel writes for one line, and the tag of each part:
// the clock when showTimestamps is true, the Service with the colour of its
// name, and the text with the stderr tag when compose wrote it.
func LogSegments(line compose.LogLine, showTimestamps bool) []LogSegment {
	segments := make([]LogSegment, 0, 3)
	if showTimestamps {
		if clock := LogTime(line.Timestamp); clock != "" {
			segments = append(segments, LogSegment{Tag: LogTimestampTag, Text: clock})
		}
	}
	if line.Service != "" {
		segments = append(segments, LogSegment{Tag: ServiceTag(line.Service), Text: line.Service})
	}
	tag := ""
	if line.Stderr {
		tag = LogStderrTag
	}
	return append(segments, LogSegment{Tag: tag, Text: line.Text})
}

// ServiceHue is the palette entry of a Service name: the same name always gets
// the same entry, so its colour never moves between runs.
func ServiceHue(service string) int {
	var h uint32
	for _, r := range service {
		h = h*31 + uint32(r)
	}
	return int(h % ServiceHues)
}

// ServiceTag is the name of the buffer tag of the colour of a Service.
func ServiceTag(service string) string {
	return serviceTag(ServiceHue(service))
}

// serviceTag is the name of the buffer tag of one palette entry.
func serviceTag(hue int) string {
	return serviceTagPrefix + strconv.Itoa(hue)
}

// ServiceColor is the colour of one palette entry in the theme, as "#rrggbb".
// An entry outside the palette has no colour.
func ServiceColor(hue int, dark bool) string {
	if hue < 0 || hue >= ServiceHues {
		return ""
	}
	if dark {
		return servicePaletteDark[hue]
	}
	return servicePaletteLight[hue]
}

// LogCountText is the count of the bar: how many lines are on screen, and how
// many the panel keeps when a search hides some of them.
func LogCountText(visible, total int) string {
	noun := "righe"
	if visible == 1 {
		noun = "riga"
	}
	if visible != total {
		return fmt.Sprintf("%d %s su %d (massimo %d)", visible, noun, total, LogLineLimit)
	}
	return fmt.Sprintf("%d %s (massimo %d)", visible, noun, LogLineLimit)
}
