package parse_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"dockercontroller/internal/compose/model"
	"dockercontroller/internal/compose/parse"
)

// The fixtures are the real output captured read-only from the projects on the
// developer machine. They stay in ../testdata, shared with the root package.
func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestLs(t *testing.T) {
	tests := []struct {
		name    string
		in      []byte
		want    []model.Discovered
		wantErr bool
	}{
		{
			name: "real compose ls output",
			in:   fixture(t, "ls.json"),
			want: []model.Discovered{
				{Name: "personalvpn", ComposeFiles: []string{"/home/cristiano/Personal/Dockers/PersonalVpn/docker-compose.yml"}},
				{Name: "stellantisvpn", ComposeFiles: []string{"/home/cristiano/Sunlink/Dockers/StellantisVpn/docker-compose.yml"}},
				{Name: "sunlinkdocker", ComposeFiles: []string{"/home/cristiano/Sunlink/Dockers/SunlinkDocker/docker-compose.yml"}},
				{Name: "sunlinkvpn", ComposeFiles: []string{"/home/cristiano/Personal/Dockers/SunlinkVpn/docker-compose.yml"}},
				{Name: "travelgrouplaravel", ComposeFiles: []string{"/home/cristiano/Personal/Lavoro/TravelGroupLaravel/docker-compose.yml"}},
			},
		},
		{name: "empty output", in: nil},
		{name: "malformed output", in: []byte("{"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse.Ls(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Ls() error = %v, want error %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ls() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPs(t *testing.T) {
	tests := []struct {
		name    string
		in      []byte
		wantLen int
		wantErr bool
	}{
		{name: "real sunlinkdocker output", in: fixture(t, "ps-sunlinkdocker.json"), wantLen: 14},
		{name: "real personalvpn output", in: fixture(t, "ps-personalvpn.json"), wantLen: 1},
		{name: "real stellantisvpn output", in: fixture(t, "ps-stellantisvpn.json"), wantLen: 1},
		{name: "empty output", in: nil},
		{name: "json array", in: []byte(`[{"Name":"web-1","Service":"web","State":"running"}]`), wantLen: 1},
		{name: "malformed output", in: []byte("{"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse.Ps(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Ps() error = %v, want error %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("Ps() returned %d Containers, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestPs_ReadsTheFieldsTheAppUses(t *testing.T) {
	rows, err := parse.Ps(fixture(t, "ps-sunlinkdocker.json"))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name                  string
		service, image, state string
		health                string
	}{
		{name: "sunlinkdocker-redis-1", service: "redis", image: "redis:7-alpine", state: "running", health: "healthy"},
		{name: "sunlinkdocker-cron-1", service: "cron", image: "sunlinkdocker-cron", state: "running"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, r := range rows {
				if r.Name != tt.name {
					continue
				}
				if r.Service != tt.service || r.Image != tt.image || r.State != tt.state || r.Health != tt.health {
					t.Errorf("row = %+v, want service %q image %q state %q health %q", r, tt.service, tt.image, tt.state, tt.health)
				}
				return
			}
			t.Fatalf("Container %q not in the real ps output", tt.name)
		})
	}
}

func TestGroupByService(t *testing.T) {
	tests := []struct {
		name           string
		fixture        string
		wantServices   int
		wantContainers int
	}{
		{name: "real sunlinkdocker output", fixture: "ps-sunlinkdocker.json", wantServices: 14, wantContainers: 14},
		{name: "real personalvpn output", fixture: "ps-personalvpn.json", wantServices: 1, wantContainers: 1},
		{name: "real stellantisvpn output", fixture: "ps-stellantisvpn.json", wantServices: 1, wantContainers: 1},
		{name: "no containers"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rows []parse.ContainerRow
			if tt.fixture != "" {
				var err error
				rows, err = parse.Ps(fixture(t, tt.fixture))
				if err != nil {
					t.Fatal(err)
				}
			}
			got := parse.GroupByService(rows)
			if got == nil {
				t.Fatal("GroupByService returned nil, want an empty list")
			}
			if len(got) != tt.wantServices {
				t.Fatalf("GroupByService returned %d Services, want %d", len(got), tt.wantServices)
			}
			containers := 0
			for i, s := range got {
				if i > 0 && got[i-1].Name > s.Name {
					t.Errorf("Services are not sorted: %q before %q", got[i-1].Name, s.Name)
				}
				containers += len(s.Containers)
			}
			if containers != tt.wantContainers {
				t.Errorf("GroupByService returned %d Containers, want %d", containers, tt.wantContainers)
			}
		})
	}
}

func TestGroupByService_RealContainers(t *testing.T) {
	rows, err := parse.Ps(fixture(t, "ps-sunlinkdocker.json"))
	if err != nil {
		t.Fatal(err)
	}
	services := parse.GroupByService(rows)
	byName := map[string]model.Service{}
	for _, s := range services {
		byName[s.Name] = s
	}
	tests := []struct {
		name string
		want model.Container
	}{
		{
			name: "redis",
			want: model.Container{
				Name: "sunlinkdocker-redis-1", Image: "redis:7-alpine", State: "running", Health: "healthy",
				Ports: []model.Port{{Published: 6379, Target: 6379, Protocol: "tcp"}},
			},
		},
		{
			name: "traefik",
			want: model.Container{
				Name: "sunlinkdocker-traefik-1", Image: "traefik:v3.2.3", State: "running",
				Ports: []model.Port{{Published: 8080, Target: 8080, Protocol: "tcp"}},
			},
		},
		{
			name: "timescale",
			want: model.Container{
				Name: "sunlinkdocker-timescale-1", Image: "sunlinkdocker-timescale", State: "running",
				Ports: []model.Port{{Published: 5433, Target: 5432, Protocol: "tcp"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, ok := byName[tt.name]
			if !ok || len(s.Containers) != 1 {
				t.Fatalf("Service %q = %+v, want one Container", tt.name, s)
			}
			if !reflect.DeepEqual(s.Containers[0], tt.want) {
				t.Errorf("Container = %+v, want %+v", s.Containers[0], tt.want)
			}
		})
	}
	cron := byName["cron"].Containers[0]
	if cron.Ports == nil || len(cron.Ports) != 0 {
		t.Errorf("cron ports = %#v, want empty non-nil", cron.Ports)
	}
}

func TestRunningCount(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		want    int
	}{
		{name: "every sunlinkdocker Container runs", fixture: "ps-sunlinkdocker.json", want: 14},
		{name: "the personalvpn Container runs", fixture: "ps-personalvpn.json", want: 1},
		{name: "the stellantisvpn Container is stopped", fixture: "ps-stellantisvpn.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := parse.Ps(fixture(t, tt.fixture))
			if err != nil {
				t.Fatal(err)
			}
			if got := parse.RunningCount(rows); got != tt.want {
				t.Errorf("RunningCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestLogLine(t *testing.T) {
	owners := map[string]string{
		"web": "web", "web-1": "web", "custom": "svc", "custom-1": "svc",
	}
	tests := []struct {
		name string
		line string
		want model.LogLine
	}{
		{
			name: "service replica and timestamp",
			line: "web-1 | 2026-10-21T12:00:00Z hello",
			want: model.LogLine{Service: "web", Timestamp: "2026-10-21T12:00:00Z", Text: "hello"},
		},
		{
			name: "Container with its own name",
			line: "custom | 2026-10-21T12:00:00Z hello",
			want: model.LogLine{Service: "svc", Timestamp: "2026-10-21T12:00:00Z", Text: "hello"},
		},
		{
			name: "no prefix",
			line: "no prefix at all",
			want: model.LogLine{Text: "no prefix at all"},
		},
		{
			name: "prefix with a space is not a prefix",
			line: "a b | 2026-10-21T12:00:00Z hello",
			want: model.LogLine{Text: "a b | 2026-10-21T12:00:00Z hello"},
		},
		{
			name: "terminal colours are stripped",
			line: "web-1 | 2026-10-21T12:00:00Z \x1b[32mok\x1b[0m",
			want: model.LogLine{Service: "web", Timestamp: "2026-10-21T12:00:00Z", Text: "ok"},
		},
		{
			name: "empty text after the timestamp",
			line: "web-1 | 2026-10-21T12:00:00Z ",
			want: model.LogLine{Service: "web", Timestamp: "2026-10-21T12:00:00Z"},
		},
		{
			name: "unknown replica keeps the Service name",
			line: "ghost-1 | 2026-10-21T12:00:00Z hello",
			want: model.LogLine{Service: "ghost", Timestamp: "2026-10-21T12:00:00Z", Text: "hello"},
		},
		{
			name: "line without a timestamp",
			line: "web-1 | plain",
			want: model.LogLine{Service: "web", Text: "plain"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parse.LogLine(tt.line, owners); got != tt.want {
				t.Errorf("LogLine(%q) = %+v, want %+v", tt.line, got, tt.want)
			}
		})
	}
}

func TestLogLine_RealComposeOutput(t *testing.T) {
	rows, err := parse.Ps(fixture(t, "ps-sunlinkdocker.json"))
	if err != nil {
		t.Fatal(err)
	}
	owners := map[string]string{}
	for _, r := range rows {
		owners[r.Name] = r.Service
		owners[r.Service] = r.Service
	}
	var got []model.LogLine
	for _, line := range strings.Split(string(fixture(t, "logs-sunlinkdocker.txt")), "\n") {
		// The Docker stream hands lines without the CR of a CRLF fixture, as
		// bufio.Scanner does.
		line = strings.TrimSuffix(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		got = append(got, parse.LogLine(line, owners))
	}
	if len(got) != 24 {
		t.Fatalf("got %d lines, want 24", len(got))
	}
	want := model.LogLine{
		Service:   "buffer",
		Timestamp: "2026-10-21T12:17:37.376631333Z",
		Text:      "Server listening on 0.0.0.0 port 22.",
	}
	if got[4] != want {
		t.Errorf("line 4 = %+v, want %+v", got[4], want)
	}
	if got[9].Service != "scheduler" || got[9].Text != "" {
		t.Errorf("empty line = %+v", got[9])
	}
	if want := "2026-10-21T12:17:40Z INF Starting provider *traefik.Provider"; got[10].Service != "traefik" || got[10].Text != want {
		t.Errorf("coloured line = %+v, want text %q", got[10], want)
	}
	if got[17].Service != "concentrator" || got[19].Service != "nodejs-api" {
		t.Errorf("services = %q, %q, want concentrator and nodejs-api", got[17].Service, got[19].Service)
	}
	for i, l := range got {
		if l.Service == "" || l.Timestamp == "" || l.Stderr {
			t.Errorf("line %d not attributed: %+v", i, l)
		}
	}
}

func TestProfiles(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want []string
	}{
		{name: "real profiles", in: fixture(t, "config-profiles.txt"), want: []string{"debug", "tools"}},
		{name: "real output without profiles", in: fixture(t, "config-profiles-none.txt"), want: []string{}},
		{name: "sorted, trimmed, blanks dropped", in: []byte("tools\n\n  debug  \n"), want: []string{"debug", "tools"}},
		{name: "empty output", in: nil, want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parse.Profiles(tt.in)
			if got == nil {
				t.Fatal("Profiles returned nil, want an empty list")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Profiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	tests := []struct {
		name    string
		in      []byte
		want    parse.Machine
		wantErr bool
	}{
		{
			name: "real docker info output",
			in:   fixture(t, "info.json"),
			want: parse.Machine{CPUs: 12, MemBytes: 33318752256},
		},
		{name: "no CPU count", in: []byte(`{"NCPU":0,"MemTotal":1}`), wantErr: true},
		{name: "no memory", in: []byte(`{"NCPU":1,"MemTotal":0}`), wantErr: true},
		{name: "malformed output", in: []byte("{"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse.Info(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Info() error = %v, want error %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("Info() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseStatsRow(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    parse.StatsRow
		wantErr bool
	}{
		{
			name: "binary memory, decimal network and disk",
			line: `{"Container":"id","Name":"n","CPUPerc":"74.64%","MemUsage":"1.5GiB / 31.03GiB","NetIO":"12.3kB / 0B","BlockIO":"127MB / 8.19kB"}`,
			want: parse.StatsRow{ID: "id", Name: "n", Measures: model.Measures{
				CPUPercent: 74.64, MemoryBytes: 1610612736,
				NetInBytes: 12300, NetOutBytes: 0,
				BlockReadBytes: 127000000, BlockWriteBytes: 8190,
			}},
		},
		{
			name: "binary memory units from KiB to TiB",
			line: `{"Container":"id","Name":"n","CPUPerc":"0.00%","MemUsage":"440KiB / 1GiB","NetIO":"0B / 0B","BlockIO":"0B / 0B"}`,
			want: parse.StatsRow{ID: "id", Name: "n", Measures: model.Measures{MemoryBytes: 450560}},
		},
		{
			name: "MiB with decimals",
			line: `{"Container":"id","Name":"n","CPUPerc":"0.01%","MemUsage":"20.68MiB / 31.03GiB","NetIO":"0B / 0B","BlockIO":"0B / 0B"}`,
			want: parse.StatsRow{ID: "id", Name: "n", Measures: model.Measures{CPUPercent: 0.01, MemoryBytes: 21684552}},
		},
		{
			name: "CPU above 100 percent",
			line: `{"Container":"id","Name":"n","CPUPerc":"372.50%","MemUsage":"0B / 0B","NetIO":"0B / 0B","BlockIO":"0B / 0B"}`,
			want: parse.StatsRow{ID: "id", Name: "n", Measures: model.Measures{CPUPercent: 372.5}},
		},
		{
			name: "placeholder counts as zero",
			line: `{"Container":"id","Name":"n","CPUPerc":"--","MemUsage":"-- / --","NetIO":"-- / --","BlockIO":"-- / --"}`,
			want: parse.StatsRow{ID: "id", Name: "n"},
		},
		{
			name: "empty fields count as zero",
			line: `{"Container":"id","Name":"n"}`,
			want: parse.StatsRow{ID: "id", Name: "n"},
		},
		{
			name: "the Container id falls back to ID then Name",
			line: `{"ID":"short","Name":"n","CPUPerc":"1%","MemUsage":"1B / 1B","NetIO":"1B / 1B","BlockIO":"1B / 1B"}`,
			want: parse.StatsRow{ID: "short", Name: "n", Measures: model.Measures{
				CPUPercent: 1, MemoryBytes: 1, NetInBytes: 1, NetOutBytes: 1, BlockReadBytes: 1, BlockWriteBytes: 1,
			}},
		},
		{
			name:    "unknown unit is an error",
			line:    `{"Container":"id","Name":"n","CPUPerc":"1%","MemUsage":"10furlongs / 1GiB","NetIO":"1B / 1B","BlockIO":"1B / 1B"}`,
			wantErr: true,
		},
		{name: "malformed JSON", line: "{", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse.ParseStatsRow(tt.line)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseStatsRow() error = %v, want error %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseStatsRow() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseStatsRow_RealNoStreamOutput(t *testing.T) {
	lines := strings.Split(strings.TrimSpace(string(fixture(t, "stats-no-stream.jsonl"))), "\n")
	if len(lines) != 16 {
		t.Fatalf("fixture has %d rows, want 16", len(lines))
	}
	got, err := parse.ParseStatsRow(lines[0])
	if err != nil {
		t.Fatal(err)
	}
	want := parse.StatsRow{
		ID:   "18e153c5f9115bb6a230428b4eda075948a392b535670323fe3cbb6d1f5dc74f",
		Name: "sunlinkdocker-nodejs-api-1",
		Measures: model.Measures{
			CPUPercent: 0.01, MemoryBytes: 21684552,
			NetInBytes: 75000, NetOutBytes: 126,
			BlockReadBytes: 127000000, BlockWriteBytes: 8190,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseStatsRow() = %+v, want %+v", got, want)
	}
}

func TestStatsFrames_RealStreamOutput(t *testing.T) {
	var frames [][]parse.StatsRow
	f := parse.NewStatsFrames(func(rows []parse.StatsRow) { frames = append(frames, rows) })
	for _, line := range strings.Split(string(fixture(t, "stats-stream.txt")), "\n") {
		f.Line(line)
	}
	f.End()
	if len(frames) != 3 {
		t.Fatalf("got %d frames, want 3", len(frames))
	}
	for i, frame := range frames {
		if len(frame) != 16 {
			t.Errorf("frame %d holds %d rows, want 16", i, len(frame))
		}
	}
}

func TestStatsFrames_NoStreamEndsTheFrame(t *testing.T) {
	var frames [][]parse.StatsRow
	f := parse.NewStatsFrames(func(rows []parse.StatsRow) { frames = append(frames, rows) })
	for _, line := range strings.Split(strings.TrimSpace(string(fixture(t, "stats-no-stream.jsonl"))), "\n") {
		f.Line(line)
	}
	f.End()
	if len(frames) != 1 {
		t.Fatalf("got %d frames, want one", len(frames))
	}
	if len(frames[0]) != 16 {
		t.Errorf("the frame holds %d rows, want 16", len(frames[0]))
	}
}

func TestStatsFrames_ARepeatedContainerStartsTheNextFrame(t *testing.T) {
	var frames [][]parse.StatsRow
	f := parse.NewStatsFrames(func(rows []parse.StatsRow) { frames = append(frames, rows) })
	lines := strings.Split(strings.TrimSpace(string(fixture(t, "stats-no-stream.jsonl"))), "\n")
	for i := 0; i < 2; i++ {
		for _, line := range lines {
			f.Line(line)
		}
	}
	f.End()
	if len(frames) != 2 {
		t.Fatalf("got %d frames, want two", len(frames))
	}
	for i, frame := range frames {
		if len(frame) != 16 {
			t.Errorf("frame %d holds %d rows, want 16", i, len(frame))
		}
	}
}

func TestStatsFrames_AnEmptyFrameIsAFrame(t *testing.T) {
	var frames [][]parse.StatsRow
	f := parse.NewStatsFrames(func(rows []parse.StatsRow) { frames = append(frames, rows) })
	for _, line := range strings.Split("\x1b[H\x1b[K\n\x1b[J\x1b[H\x1b[K\n\x1b[J", "\n") {
		f.Line(line)
	}
	f.End()
	if len(frames) != 2 || len(frames[0]) != 0 || len(frames[1]) != 0 {
		t.Errorf("got %d frames, want two empty frames", len(frames))
	}
}
