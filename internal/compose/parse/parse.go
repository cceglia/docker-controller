package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"dockercontroller/internal/compose/model"
)

// Ls parses the output of `docker compose ls -a --format json`.
func Ls(out []byte) ([]model.Discovered, error) {
	var raw []struct {
		Name        string `json:"Name"`
		ConfigFiles string `json:"ConfigFiles"`
	}
	if len(bytes.TrimSpace(out)) == 0 {
		return nil, nil
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("compose: parse `compose ls`: %w", err)
	}
	res := make([]model.Discovered, 0, len(raw))
	for _, r := range raw {
		d := model.Discovered{Name: r.Name}
		for _, f := range strings.Split(r.ConfigFiles, ",") {
			if f = strings.TrimSpace(f); f != "" {
				d.ComposeFiles = append(d.ComposeFiles, f)
			}
		}
		res = append(res, d)
	}
	return res, nil
}

// Ps parses the output of `docker compose -p <name> ps -a --format json`
// into the Project's Containers. Compose emits one JSON object per line; a JSON
// array is accepted too.
func Ps(out []byte) ([]ContainerRow, error) {
	trimmed := bytes.TrimSpace(out)
	if len(trimmed) == 0 {
		return nil, nil
	}
	var cs []ContainerRow
	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &cs); err != nil {
			return nil, fmt.Errorf("compose: parse `compose ps`: %w", err)
		}
		return cs, nil
	}
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	for dec.More() {
		var c ContainerRow
		if err := dec.Decode(&c); err != nil {
			return nil, fmt.Errorf("compose: parse `compose ps`: %w", err)
		}
		cs = append(cs, c)
	}
	return cs, nil
}

// ContainerRow is the part of a `compose ps` entry the app uses.
type ContainerRow struct {
	Name       string `json:"Name"`
	Service    string `json:"Service"`
	Image      string `json:"Image"`
	State      string `json:"State"`
	Health     string `json:"Health"`
	Publishers []struct {
		TargetPort    int    `json:"TargetPort"`
		PublishedPort int    `json:"PublishedPort"`
		Protocol      string `json:"Protocol"`
	} `json:"Publishers"`
}

// GroupByService groups Containers by Service, both sorted by name.
func GroupByService(cs []ContainerRow) []model.Service {
	byName := map[string][]model.Container{}
	for _, c := range cs {
		byName[c.Service] = append(byName[c.Service], model.Container{
			Name: c.Name, Image: c.Image, State: c.State, Health: c.Health, Ports: c.ports(),
		})
	}
	services := make([]model.Service, 0, len(byName))
	for name, containers := range byName {
		sort.Slice(containers, func(i, j int) bool { return containers[i].Name < containers[j].Name })
		services = append(services, model.Service{Name: name, Containers: containers})
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })
	return services
}

// ports lists the published ports once each (a port bound on 0.0.0.0 and :: is
// reported twice by Docker), sorted by published port.
func (c ContainerRow) ports() []model.Port {
	ports := []model.Port{}
	seen := map[model.Port]bool{}
	for _, p := range c.Publishers {
		if p.PublishedPort == 0 {
			continue // exposed but not published
		}
		port := model.Port{Published: p.PublishedPort, Target: p.TargetPort, Protocol: p.Protocol}
		if !seen[port] {
			seen[port] = true
			ports = append(ports, port)
		}
	}
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].Published != ports[j].Published {
			return ports[i].Published < ports[j].Published
		}
		return ports[i].Protocol < ports[j].Protocol
	})
	return ports
}

// RunningCount is how many of the Containers run.
func RunningCount(cs []ContainerRow) int {
	n := 0
	for _, c := range cs {
		if c.State == "running" {
			n++
		}
	}
	return n
}

// ansi matches the terminal escape sequences Containers write for colours.
var ansi = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

// LogLine splits "<prefix> | <timestamp> <text>".
func LogLine(line string, owners map[string]string) model.LogLine {
	prefix, rest, found := strings.Cut(line, " | ")
	prefix = strings.TrimRight(prefix, " ")
	if !found || prefix == "" || strings.ContainsAny(prefix, " \t") {
		return model.LogLine{Text: ansi.ReplaceAllString(line, "")}
	}
	out := model.LogLine{Service: serviceOf(prefix, owners), Text: rest}
	ts, text, hasText := strings.Cut(rest, " ")
	if _, err := time.Parse(time.RFC3339Nano, ts); err == nil {
		out.Timestamp = ts
		out.Text = ""
		if hasText {
			out.Text = text
		}
	}
	out.Text = ansi.ReplaceAllString(out.Text, "")
	return out
}

// serviceOf resolves a compose line prefix to a Service name.
func serviceOf(prefix string, owners map[string]string) string {
	if s, ok := owners[prefix]; ok {
		return s
	}
	if i := strings.LastIndexByte(prefix, '-'); i > 0 && isNumber(prefix[i+1:]) {
		if s, ok := owners[prefix[:i]]; ok {
			return s
		}
		// A Service of a Project Docker no longer shows: the replica number
		// is not part of its name.
		return prefix[:i]
	}
	return prefix
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Profiles parses the output of `compose config --profiles`: one name per
// line. The result is sorted and never nil.
func Profiles(out []byte) []string {
	names := []string{}
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			names = append(names, line)
		}
	}
	sort.Strings(names)
	return names
}

// Machine is the size of the computer Docker runs on, from `docker info`.
type Machine struct {
	CPUs     int
	MemBytes uint64
}

// Info parses the output of `docker info --format json`.
func Info(out []byte) (Machine, error) {
	var info struct {
		NCPU     int    `json:"NCPU"`
		MemTotal uint64 `json:"MemTotal"`
	}
	if err := json.Unmarshal(out, &info); err != nil {
		return Machine{}, fmt.Errorf("compose: read docker info: %w", err)
	}
	if info.NCPU <= 0 || info.MemTotal == 0 {
		return Machine{}, fmt.Errorf("compose: docker info reports no CPU count or memory (%d, %d)", info.NCPU, info.MemTotal)
	}
	return Machine{CPUs: info.NCPU, MemBytes: info.MemTotal}, nil
}

// StatsRow is one Container of one frame of `docker stats`.
type StatsRow struct {
	ID       string
	Name     string
	Measures model.Measures
}

// byteUnits maps the unit docker prints to its size. Docker prints the memory
// with binary units (KiB, MiB, GiB: powers of 1024) and the network and disk
// counters with decimal ones (kB, MB, GB: powers of 1000).
var byteUnits = map[string]float64{
	"b":  1,
	"kb": 1e3, "mb": 1e6, "gb": 1e9, "tb": 1e12, "pb": 1e15,
	"kib": 1 << 10, "mib": 1 << 20, "gib": 1 << 30, "tib": 1 << 40, "pib": 1 << 50,
}

// parseBytes reads a size like "1.5GiB", "12.3kB" or "0B". Docker prints "--"
// (or nothing) while it has no figure yet, which counts as zero.
func parseBytes(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "--" {
		return 0, nil
	}
	i := strings.IndexFunc(s, func(r rune) bool { return (r < '0' || r > '9') && r != '.' })
	if i <= 0 {
		return 0, fmt.Errorf("compose: size %q has no number or no unit", s)
	}
	n, err := strconv.ParseFloat(s[:i], 64)
	if err != nil {
		return 0, fmt.Errorf("compose: size %q: %w", s, err)
	}
	mult, ok := byteUnits[strings.ToLower(strings.TrimSpace(s[i:]))]
	if !ok {
		return 0, fmt.Errorf("compose: size %q has an unknown unit", s)
	}
	return uint64(math.Round(n * mult)), nil
}

// parsePair reads "<a> / <b>" (NetIO, BlockIO; MemUsage is "<used> / <limit>").
func parsePair(s string) (uint64, uint64, error) {
	a, b, _ := strings.Cut(s, " / ")
	x, err := parseBytes(a)
	if err != nil {
		return 0, 0, err
	}
	y, err := parseBytes(b)
	return x, y, err
}

// parseCPU reads "74.64%", the percentage of one core ("--" is zero).
func parseCPU(s string) (float64, error) {
	s = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "%"))
	if s == "" || s == "--" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

// ParseStatsRow reads a JSON line of `docker stats --format json`. A row that
// cannot be read is refused as a whole, so a sum never holds half a Container.
func ParseStatsRow(line string) (StatsRow, error) {
	var r struct {
		Container string `json:"Container"`
		ID        string `json:"ID"`
		Name      string `json:"Name"`
		CPUPerc   string `json:"CPUPerc"`
		MemUsage  string `json:"MemUsage"`
		NetIO     string `json:"NetIO"`
		BlockIO   string `json:"BlockIO"`
	}
	if err := json.Unmarshal([]byte(line), &r); err != nil {
		return StatsRow{}, fmt.Errorf("compose: read docker stats row: %w", err)
	}
	row := StatsRow{ID: r.Container, Name: r.Name}
	if row.ID == "" {
		row.ID = r.ID
	}
	if row.ID == "" {
		row.ID = r.Name
	}
	var err error
	if row.Measures.CPUPercent, err = parseCPU(r.CPUPerc); err != nil {
		return StatsRow{}, fmt.Errorf("compose: stats row %s: CPU: %w", r.Name, err)
	}
	if row.Measures.MemoryBytes, _, err = parsePair(r.MemUsage); err != nil {
		return StatsRow{}, fmt.Errorf("compose: stats row %s: memory: %w", r.Name, err)
	}
	if row.Measures.NetInBytes, row.Measures.NetOutBytes, err = parsePair(r.NetIO); err != nil {
		return StatsRow{}, fmt.Errorf("compose: stats row %s: network: %w", r.Name, err)
	}
	if row.Measures.BlockReadBytes, row.Measures.BlockWriteBytes, err = parsePair(r.BlockIO); err != nil {
		return StatsRow{}, fmt.Errorf("compose: stats row %s: block I/O: %w", r.Name, err)
	}
	return row, nil
}

// StatsFrames groups the lines of `docker stats` into frames, one per tick of
// the daemon (about every 2 seconds), each holding a row per running
// Container.
//
// Docker redraws the terminal for every frame, even when stdout is a pipe:
// a frame starts with "ESC[H" (cursor home, after "ESC[J" for the frames
// after the first), every row ends with "ESC[K", and a line holding only
// "ESC[K" closes the frame. So a frame is complete as soon as that closing
// line arrives, without waiting for the next tick or any timer. Two fallbacks
// cover output without redraw codes (like `--no-stream`): a Container seen
// twice means a new frame has started, and the end of the stream ends the
// frame in progress. An empty frame (no Container running) is a frame too.
type StatsFrames struct {
	emit   func([]StatsRow)
	rows   []StatsRow
	seen   map[string]bool
	active bool // a frame has started and has not been emitted yet
}

const (
	escHome = "\x1b[H"
)

// NewStatsFrames creates a StatsFrames that delivers every complete frame to
// emit.
func NewStatsFrames(emit func([]StatsRow)) *StatsFrames {
	return &StatsFrames{emit: emit}
}

// Line takes a line of stdout.
func (f *StatsFrames) Line(raw string) {
	clean := strings.TrimSpace(ansi.ReplaceAllString(raw, ""))
	if strings.Contains(raw, escHome) {
		f.flush() // a new frame starts: whatever was open is complete
		f.active = true
	}
	if clean == "" {
		f.flush() // the closing line of the frame
		return
	}
	row, err := ParseStatsRow(clean)
	if err != nil {
		return // docker's own text, or a row that cannot be trusted
	}
	if f.seen[row.ID] {
		f.flush() // the same Container again: this is the next frame
	}
	f.active = true
	if f.seen == nil {
		f.seen = map[string]bool{}
	}
	f.seen[row.ID] = true
	f.rows = append(f.rows, row)
}

// End takes the end of the stream: a frame left open is complete.
func (f *StatsFrames) End() {
	if len(f.rows) > 0 {
		f.flush()
	}
}

func (f *StatsFrames) flush() {
	if !f.active {
		return
	}
	rows := f.rows
	f.rows, f.seen, f.active = nil, nil, false
	f.emit(rows)
}
