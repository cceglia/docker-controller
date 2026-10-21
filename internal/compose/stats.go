package compose

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// statsArgs streams the usage of every running Container, whatever Project it
// belongs to.
var statsArgs = []string{"stats", "--format", "json"}

// infoArgs asks the daemon for the size of the machine.
var infoArgs = []string{"info", "--format", "json"}

// machine is the size of the computer Docker runs on, from `docker info`.
type machine struct {
	cpus     int
	memBytes uint64
}

func parseInfo(out []byte) (machine, error) {
	var info struct {
		NCPU     int    `json:"NCPU"`
		MemTotal uint64 `json:"MemTotal"`
	}
	if err := json.Unmarshal(out, &info); err != nil {
		return machine{}, fmt.Errorf("compose: read docker info: %w", err)
	}
	if info.NCPU <= 0 || info.MemTotal == 0 {
		return machine{}, fmt.Errorf("compose: docker info reports no CPU count or memory (%d, %d)", info.NCPU, info.MemTotal)
	}
	return machine{cpus: info.NCPU, memBytes: info.MemTotal}, nil
}

// statsRow is one Container of one frame of `docker stats`.
type statsRow struct {
	id, name string
	measures Measures
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

// parseStatsRow reads a JSON line of `docker stats --format json`. A row that
// cannot be read is refused as a whole, so a sum never holds half a Container.
func parseStatsRow(line string) (statsRow, error) {
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
		return statsRow{}, fmt.Errorf("compose: read docker stats row: %w", err)
	}
	row := statsRow{id: r.Container, name: r.Name}
	if row.id == "" {
		row.id = r.ID
	}
	if row.id == "" {
		row.id = r.Name
	}
	var err error
	if row.measures.CPUPercent, err = parseCPU(r.CPUPerc); err != nil {
		return statsRow{}, fmt.Errorf("compose: stats row %s: CPU: %w", r.Name, err)
	}
	if row.measures.MemoryBytes, _, err = parsePair(r.MemUsage); err != nil {
		return statsRow{}, fmt.Errorf("compose: stats row %s: memory: %w", r.Name, err)
	}
	if row.measures.NetInBytes, row.measures.NetOutBytes, err = parsePair(r.NetIO); err != nil {
		return statsRow{}, fmt.Errorf("compose: stats row %s: network: %w", r.Name, err)
	}
	if row.measures.BlockReadBytes, row.measures.BlockWriteBytes, err = parsePair(r.BlockIO); err != nil {
		return statsRow{}, fmt.Errorf("compose: stats row %s: block I/O: %w", r.Name, err)
	}
	return row, nil
}

// statsFrames groups the lines of `docker stats` into frames, one per tick of
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
type statsFrames struct {
	emit   func([]statsRow)
	rows   []statsRow
	seen   map[string]bool
	active bool // a frame has started and has not been emitted yet
}

const (
	escHome = "\x1b[H"
)

// line takes a line of stdout.
func (f *statsFrames) line(raw string) {
	clean := strings.TrimSpace(ansi.ReplaceAllString(raw, ""))
	if strings.Contains(raw, escHome) {
		f.flush() // a new frame starts: whatever was open is complete
		f.active = true
	}
	if clean == "" {
		f.flush() // the closing line of the frame
		return
	}
	row, err := parseStatsRow(clean)
	if err != nil {
		return // docker's own text, or a row that cannot be trusted
	}
	if f.seen[row.id] {
		f.flush() // the same Container again: this is the next frame
	}
	f.active = true
	if f.seen == nil {
		f.seen = map[string]bool{}
	}
	f.seen[row.id] = true
	f.rows = append(f.rows, row)
}

// end takes the end of the stream: a frame left open is complete.
func (f *statsFrames) end() {
	if len(f.rows) > 0 {
		f.flush()
	}
}

func (f *statsFrames) flush() {
	if !f.active {
		return
	}
	rows := f.rows
	f.rows, f.seen, f.active = nil, nil, false
	f.emit(rows)
}
