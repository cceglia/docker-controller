package logs

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/design"
)

// logLine is a short way to build a line of a Container.
func logLine(service, text string) compose.LogLine {
	return compose.LogLine{Service: service, Text: text}
}

func TestAppendLogLines(t *testing.T) {
	lines := func(names ...string) []compose.LogLine {
		out := make([]compose.LogLine, len(names))
		for i, name := range names {
			out[i] = logLine("web", name)
		}
		return out
	}
	for _, tc := range []struct {
		name     string
		lines    []compose.LogLine
		incoming []compose.LogLine
		limit    int
		want     []string
	}{
		{
			name:     "an empty history takes the first line",
			incoming: lines("a"),
			limit:    3,
			want:     []string{"a"},
		},
		{
			name:     "under the cap nothing is dropped",
			lines:    lines("a", "b"),
			incoming: lines("c"),
			limit:    3,
			want:     []string{"a", "b", "c"},
		},
		{
			name:     "at the cap nothing is dropped",
			lines:    lines("a", "b", "c"),
			incoming: lines("d"),
			limit:    4,
			want:     []string{"a", "b", "c", "d"},
		},
		{
			name:     "over the cap the oldest lines go",
			lines:    lines("a", "b", "c"),
			incoming: lines("d"),
			limit:    3,
			want:     []string{"b", "c", "d"},
		},
		{
			name:     "a burst bigger than the cap keeps its tail",
			lines:    lines("a"),
			incoming: lines("b", "c", "d", "e"),
			limit:    3,
			want:     []string{"c", "d", "e"},
		},
		{
			name:  "an empty burst changes nothing",
			lines: lines("a", "b"),
			limit: 3,
			want:  []string{"a", "b"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := AppendLogLines(tc.lines, tc.incoming, tc.limit)
			if len(got) != len(tc.want) {
				t.Fatalf("kept %d lines, want %d: %v", len(got), len(tc.want), texts(got))
			}
			for i, want := range tc.want {
				if got[i].Text != want {
					t.Errorf("line %d = %q, want %q", i, got[i].Text, want)
				}
			}
		})
	}
}

// texts returns the text of some lines, for the messages of a failed case.
func texts(lines []compose.LogLine) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = line.Text
	}
	return out
}

// TestAppendLogLines_KeepsTheCapOfThePanel uses the real limit of the page, so
// a change of LogLineLimit that leaves the panel's cap behind fails here.
func TestAppendLogLines_KeepsTheCapOfThePanel(t *testing.T) {
	numbered := func(from, to int) []compose.LogLine {
		out := make([]compose.LogLine, 0, to-from)
		for i := from; i < to; i++ {
			out = append(out, logLine("web", strconv.Itoa(i)))
		}
		return out
	}

	t.Run("one line over the cap drops the oldest", func(t *testing.T) {
		lines := numbered(0, LogLineLimit)
		got := AppendLogLines(lines, numbered(LogLineLimit, LogLineLimit+1), LogLineLimit)
		if len(got) != LogLineLimit {
			t.Fatalf("kept %d lines, want %d", len(got), LogLineLimit)
		}
		if got[0].Text != "1" {
			t.Errorf("oldest line = %q, want %q", got[0].Text, "1")
		}
		if last := got[len(got)-1].Text; last != strconv.Itoa(LogLineLimit) {
			t.Errorf("newest line = %q, want %q", last, strconv.Itoa(LogLineLimit))
		}
	})

	t.Run("a burst bigger than the cap keeps its tail", func(t *testing.T) {
		got := AppendLogLines(numbered(0, 100), numbered(100, 100+LogLineLimit+10), LogLineLimit)
		if len(got) != LogLineLimit {
			t.Fatalf("kept %d lines, want %d", len(got), LogLineLimit)
		}
		if got[0].Text != "110" {
			t.Errorf("oldest line = %q, want %q", got[0].Text, "110")
		}
		if last := got[len(got)-1].Text; last != strconv.Itoa(100+LogLineLimit+9) {
			t.Errorf("newest line = %q, want %q", last, strconv.Itoa(100+LogLineLimit+9))
		}
	})
}

func TestLogTargetText(t *testing.T) {
	shop := compose.Project{Name: "shop", DisplayName: "Negozio"}
	for _, tc := range []struct {
		name    string
		project compose.Project
		service string
		want    string
	}{
		{name: "no Project names no target"},
		{name: "a Project without a Display name reads its compose name", project: compose.Project{Name: "blog"}, want: "blog"},
		{name: "the Display name names the Project", project: shop, want: "Negozio"},
		{name: "one Service follows the Project", project: shop, service: "web", want: "Negozio · web"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := LogTargetText(tc.project, tc.service); got != tc.want {
				t.Errorf("LogTargetText() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFilterLogLines(t *testing.T) {
	lines := []compose.LogLine{
		logLine("web", "GET /index.html 200"),
		logLine("db", "database system is ready"),
		{Text: "no configuration file provided", Stderr: true},
	}
	for _, tc := range []struct {
		name string
		term string
		want []string
	}{
		{
			name: "an empty term keeps every line",
			want: []string{"GET /index.html 200", "database system is ready", "no configuration file provided"},
		},
		{
			name: "a blank term keeps every line",
			term: "   ",
			want: []string{"GET /index.html 200", "database system is ready", "no configuration file provided"},
		},
		{
			name: "the term is matched ignoring case",
			term: "DATABASE",
			want: []string{"database system is ready"},
		},
		{
			name: "the term is matched in the middle of the text",
			term: "index",
			want: []string{"GET /index.html 200"},
		},
		{
			name: "the term is trimmed",
			term: "  ready  ",
			want: []string{"database system is ready"},
		},
		{
			name: "a term that matches no line keeps none",
			term: "missing",
		},
		{
			name: "the Service name is not part of the search",
			term: "web",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := FilterLogLines(lines, tc.term)
			if len(got) != len(tc.want) {
				t.Fatalf("kept %d lines, want %d: %v", len(got), len(tc.want), texts(got))
			}
			for i, want := range tc.want {
				if got[i].Text != want {
					t.Errorf("line %d = %q, want %q", i, got[i].Text, want)
				}
			}
		})
	}
}

func TestLogTime(t *testing.T) {
	// 2024-03-01T12:34:56Z, as an independent source for the expected clock.
	const epoch = 1709296496
	local := time.Unix(epoch, 0).In(time.Local).Format("15:04:05")

	for _, tc := range []struct {
		name      string
		timestamp string
		want      string
	}{
		{name: "a UTC timestamp is shown in local time", timestamp: "2024-03-01T12:34:56Z", want: local},
		{name: "the nanoseconds of Docker are accepted", timestamp: "2024-03-01T12:34:56.123456789Z", want: local},
		{name: "an empty timestamp has no clock", timestamp: ""},
		{name: "a text that is not a timestamp has no clock", timestamp: "not a time"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := LogTime(tc.timestamp); got != tc.want {
				t.Errorf("LogTime(%q) = %q, want %q", tc.timestamp, got, tc.want)
			}
		})
	}
}

func TestLogSegments(t *testing.T) {
	line := compose.LogLine{
		Service:   "web",
		Timestamp: "2024-03-01T12:34:56Z",
		Text:      "hello",
	}
	clock := LogTime(line.Timestamp)

	for _, tc := range []struct {
		name           string
		line           compose.LogLine
		showTimestamps bool
		want           []LogSegment
	}{
		{
			name:           "timestamp, Service and text with the timestamps on",
			line:           line,
			showTimestamps: true,
			want: []LogSegment{
				{Tag: LogTimestampTag, Text: clock},
				{Tag: ServiceTag("web"), Text: "web"},
				{Tag: "", Text: "hello"},
			},
		},
		{
			name:           "the timestamps can be hidden without touching the line",
			line:           line,
			showTimestamps: false,
			want: []LogSegment{
				{Tag: ServiceTag("web"), Text: "web"},
				{Tag: "", Text: "hello"},
			},
		},
		{
			name:           "a line without timestamp shows no clock",
			line:           compose.LogLine{Service: "web", Text: "hello"},
			showTimestamps: true,
			want: []LogSegment{
				{Tag: ServiceTag("web"), Text: "web"},
				{Tag: "", Text: "hello"},
			},
		},
		{
			name:           "a line with an invalid timestamp shows no clock",
			line:           compose.LogLine{Service: "web", Timestamp: "yesterday", Text: "hello"},
			showTimestamps: true,
			want: []LogSegment{
				{Tag: ServiceTag("web"), Text: "web"},
				{Tag: "", Text: "hello"},
			},
		},
		{
			name:           "a message of compose has no Service and the stderr tag",
			line:           compose.LogLine{Text: "no configuration file provided", Stderr: true},
			showTimestamps: true,
			want: []LogSegment{
				{Tag: LogStderrTag, Text: "no configuration file provided"},
			},
		},
		{
			name:           "the output of a Container never carries the stderr tag",
			line:           compose.LogLine{Service: "db", Text: "ready"},
			showTimestamps: true,
			want: []LogSegment{
				{Tag: ServiceTag("db"), Text: "db"},
				{Tag: "", Text: "ready"},
			},
		},
		{
			name:           "an empty line keeps its empty text",
			line:           compose.LogLine{Service: "web"},
			showTimestamps: true,
			want: []LogSegment{
				{Tag: ServiceTag("web"), Text: "web"},
				{Tag: "", Text: ""},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := LogSegments(tc.line, tc.showTimestamps)
			if len(got) != len(tc.want) {
				t.Fatalf("%d segments, want %d: %+v", len(got), len(tc.want), got)
			}
			for i, want := range tc.want {
				if got[i] != want {
					t.Errorf("segment %d = %+v, want %+v", i, got[i], want)
				}
			}
		})
	}
}

func TestServiceHue(t *testing.T) {
	// The values of the rule the old view used.
	for _, tc := range []struct {
		service string
		want    int
	}{
		{service: "web", want: 4},
		{service: "db", want: 14},
		{service: "api", want: 10},
		{service: "redis", want: 11},
		{service: "db-1", want: 2},
	} {
		t.Run(tc.service, func(t *testing.T) {
			if got := ServiceHue(tc.service); got != tc.want {
				t.Errorf("ServiceHue(%q) = %d, want %d", tc.service, got, tc.want)
			}
		})
	}

	t.Run("the same name always gets the same hue", func(t *testing.T) {
		if ServiceHue("worker") != ServiceHue("worker") {
			t.Error("the hue of a Service changed between two calls")
		}
	})

	t.Run("every hue is in the palette", func(t *testing.T) {
		for i := 0; i < 500; i++ {
			service := "s" + strconv.Itoa(i)
			if hue := ServiceHue(service); hue < 0 || hue >= design.ServiceHues {
				t.Fatalf("ServiceHue(%q) = %d, want 0..%d", service, hue, design.ServiceHues-1)
			}
		}
	})
}

func TestServiceTag(t *testing.T) {
	for _, tc := range []struct {
		name    string
		service string
		want    string
	}{
		{name: "a name keeps the tag of its hue", service: "web", want: "log-service-4"},
		{name: "another name keeps its own tag", service: "db", want: "log-service-14"},
		{name: "an empty name still gets a tag", service: "", want: "log-service-0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ServiceTag(tc.service); got != tc.want {
				t.Errorf("ServiceTag(%q) = %q, want %q", tc.service, got, tc.want)
			}
		})
	}

	t.Run("every hue has its own tag, apart from the other tags", func(t *testing.T) {
		seen := make(map[string]int, design.ServiceHues)
		for hue := 0; hue < design.ServiceHues; hue++ {
			tag := serviceTag(hue)
			if want := serviceTagPrefix + strconv.Itoa(hue); tag != want {
				t.Errorf("serviceTag(%d) = %q, want %q", hue, tag, want)
			}
			if other, duplicate := seen[tag]; duplicate {
				t.Errorf("hues %d and %d share the tag %q", other, hue, tag)
			}
			seen[tag] = hue
			if tag == LogTimestampTag || tag == LogStderrTag {
				t.Errorf("hue %d takes the tag %q of another part of the line", hue, tag)
			}
		}
	})
}

func TestServiceHueOf(t *testing.T) {
	for _, name := range []string{"web", "db", "api", "worker", ""} {
		t.Run(name, func(t *testing.T) {
			hue, ok := ServiceHueOf(ServiceTag(name))
			if !ok {
				t.Fatalf("the tag %q belongs to no hue", ServiceTag(name))
			}
			if want := ServiceHue(name); hue != want {
				t.Errorf("the tag of %q reads as hue %d, want %d", name, hue, want)
			}
		})
	}

	for _, tag := range []string{"", LogTimestampTag, LogStderrTag, serviceTagPrefix, serviceTagPrefix + "x", serviceTagPrefix + "-1", serviceTagPrefix + strconv.Itoa(design.ServiceHues)} {
		t.Run("foreign/"+tag, func(t *testing.T) {
			if hue, ok := ServiceHueOf(tag); ok {
				t.Errorf("ServiceHueOf(%q) = %d, want no hue", tag, hue)
			}
		})
	}
}

func TestLogCountText(t *testing.T) {
	for _, tc := range []struct {
		name    string
		visible int
		total   int
		want    string
	}{
		{name: "no search shows the total", visible: 12, total: 12, want: "12 righe (massimo 5000)"},
		{name: "one line is singular", visible: 1, total: 1, want: "1 riga (massimo 5000)"},
		{name: "a search shows the visible lines out of the total", visible: 3, total: 5000, want: "3 righe su 5000 (massimo 5000)"},
		{name: "a search with no match shows zero out of the total", visible: 0, total: 5000, want: "0 righe su 5000 (massimo 5000)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := LogCountText(tc.visible, tc.total); got != tc.want {
				t.Errorf("LogCountText(%d, %d) = %q, want %q", tc.visible, tc.total, got, tc.want)
			}
		})
	}
}

// BenchmarkFilterLogLines measures the work a keystroke of the search costs on
// a full panel.
func BenchmarkFilterLogLines(b *testing.B) {
	lines := make([]compose.LogLine, LogLineLimit)
	for i := range lines {
		lines[i] = logLine("service-"+strconv.Itoa(i%20),
			fmt.Sprintf("2024-03-01 line %d of the stream, message body", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if got := FilterLogLines(lines, "message"); len(got) != LogLineLimit {
			b.Fatalf("%d lines, want %d", len(got), LogLineLimit)
		}
	}
}
