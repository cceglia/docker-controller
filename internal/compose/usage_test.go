package compose_test

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"dockercontroller/internal/compose"
	"dockercontroller/internal/compose/composefake"
	"dockercontroller/internal/db"
)

// The usage fixtures are real output captured read-only from the machine:
//
//	docker info --format json                (NCPU 12, MemTotal 33318752256)
//	docker stats --format json               (stats-stream.txt: 3 frames of the
//	                                          stream, raw, terminal escapes included)
//	docker stats --no-stream --format json   (stats-no-stream.jsonl)
//
// The stream holds the 16 running Containers of the machine: 14 of the saved
// Project sunlinkdocker, plus sunlink-vpn and personal-vpn. sunlink-vpn belongs
// to a Project that is not saved, so only the global figures count it.
var (
	infoArgs  = []string{"info", "--format", "json"}
	statsArgs = []string{"stats", "--format", "json"}
)

const (
	machineCPUs = 12
	machineMem  = 33318752256
)

// usageDocker is realDocker plus the usage commands: stream is what
// `docker stats` prints, err how it exits; with follow it stays open after
// its output, like the real command, until it is cancelled.
func usageDocker(t *testing.T, stream []byte, err error, follow bool) *composefake.Fake {
	t.Helper()
	f := realDocker(t)
	f.Respond(fixture(t, "info.json"), infoArgs...)
	f.RespondStream(stream, nil, err, statsArgs...)
	if follow {
		f.Follow(statsArgs...)
	}
	return f
}

// followingDocker streams the three real frames and stays open.
func followingDocker(t *testing.T) *composefake.Fake {
	t.Helper()
	return usageDocker(t, fixture(t, "stats-stream.txt"), nil, true)
}

// usageManager is a refreshed Manager that retries quickly.
func usageManager(t *testing.T, docker compose.Docker) *compose.Manager {
	t.Helper()
	conn, err := db.Open(tempDB(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	m, err := compose.New(compose.Config{Docker: docker, DB: conn, RetryEvery: 5 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	return m
}

// sampleCollector gathers the samples a Usage call delivers.
type sampleCollector struct {
	mu      sync.Mutex
	samples []compose.UsageSample
}

func (c *sampleCollector) on(s compose.UsageSample) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.samples = append(c.samples, s)
}

func (c *sampleCollector) all() []compose.UsageSample {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]compose.UsageSample(nil), c.samples...)
}

func (c *sampleCollector) wait(t *testing.T, n int) []compose.UsageSample {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if got := c.all(); len(got) >= n {
			return got
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("got %d samples, want %d: %+v", len(c.all()), n, c.all())
	return nil
}

// startUsage runs Usage in the background; the returned stop cancels it and
// waits for its return value.
func startUsage(t *testing.T, m *compose.Manager) (c *sampleCollector, stop func() error) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	c = &sampleCollector{}
	done := make(chan error, 1)
	go func() { done <- m.Usage(ctx, c.on) }()
	var once sync.Once
	var result error
	stop = func() error {
		once.Do(func() {
			cancel()
			select {
			case result = <-done:
			case <-time.After(2 * time.Second):
				t.Error("Usage did not return after the cancellation")
			}
		})
		return result
	}
	t.Cleanup(func() { _ = stop() })
	return c, stop
}

func near(got, want float64) bool { return math.Abs(got-want) < 0.005 }

func TestUsage_GlobalIsTheSumOverEveryRunningContainer_NormalisedOnTheMachine(t *testing.T) {
	m := usageManager(t, followingDocker(t))
	c, _ := startUsage(t, m)

	samples := c.wait(t, 3)
	// One sample per frame of the stream: 16 rows each, not one per row.
	if len(samples) != 3 {
		t.Fatalf("got %d samples for 3 frames: %+v", len(samples), samples)
	}
	s := samples[0]
	if !s.DockerReachable {
		t.Fatal("docker should be reachable")
	}
	// Expected values: sums worked out by hand from the rows of the fixture,
	// with decimal units for rete/disco and binary units for the memoria.
	if s.Containers != 16 {
		t.Errorf("containers = %d, want 16 (also those of Projects that are not saved)", s.Containers)
	}
	if !near(s.Total.CPUPercent, 20.75) {
		t.Errorf("CPU = %.2f, want 20.75 (docker's own scale: 100 is one core)", s.Total.CPUPercent)
	}
	if s.Total.MemoryBytes != 1140121599 {
		t.Errorf("memory = %d, want 1140121599", s.Total.MemoryBytes)
	}
	if s.Total.NetInBytes != 21567997000 || s.Total.NetOutBytes != 21383373504 {
		t.Errorf("net = %d/%d, want 21567997000/21383373504", s.Total.NetInBytes, s.Total.NetOutBytes)
	}
	if s.Total.BlockReadBytes != 2508930000 || s.Total.BlockWriteBytes != 392162419380 {
		t.Errorf("block = %d/%d, want 2508930000/392162419380", s.Total.BlockReadBytes, s.Total.BlockWriteBytes)
	}
	// 20.75 / 12 CPUs; 1140121599 / 33318752256.
	if s.CPUs != machineCPUs || !near(s.CPUMachinePercent, 1.7292) {
		t.Errorf("cpus %d, machine CPU = %.4f%%, want 12 and 1.7292%%", s.CPUs, s.CPUMachinePercent)
	}
	if s.MemoryTotalBytes != machineMem || !near(s.MemoryMachinePercent, 3.4220) {
		t.Errorf("machine memory %d, %.4f%%, want %d and 3.4220%%", s.MemoryTotalBytes, s.MemoryMachinePercent, int64(machineMem))
	}
	// The third frame carries newer values: a sample replaces the previous one.
	if !near(samples[2].Total.CPUPercent, 172.78) || !near(samples[2].CPUMachinePercent, 14.3983) {
		t.Errorf("third sample CPU = %.2f (%.4f%% of the machine), want 172.78 (14.3983%%)", samples[2].Total.CPUPercent, samples[2].CPUMachinePercent)
	}
}

func TestUsage_RunsTheStreamingFormOfStatsAndAsksTheMachineSizeFromInfo(t *testing.T) {
	f := followingDocker(t)
	m := usageManager(t, f)
	c, stop := startUsage(t, m)
	c.wait(t, 1)
	_ = stop()

	var info, stats int
	for _, call := range f.Calls() {
		switch strings.Join(call, " ") {
		case strings.Join(infoArgs, " "):
			info++
		case strings.Join(statsArgs, " "):
			stats++
		}
	}
	if info != 1 || stats != 1 {
		t.Errorf("ran info %d times and stats %d times, want once each: %v", info, stats, f.Calls())
	}
}

func TestUsage_AFrameWithoutTerminalEscapesIsASampleAtTheEndOfTheStream(t *testing.T) {
	// The --no-stream form has no escape codes: the frame ends with the stream.
	f := usageDocker(t, fixture(t, "stats-no-stream.jsonl"), nil, false)
	m := usageManager(t, f)
	c, _ := startUsage(t, m)

	samples := c.wait(t, 1)
	if !samples[0].DockerReachable || samples[0].Containers != 16 {
		t.Errorf("sample = %+v, want 16 Containers", samples[0])
	}
}

func TestUsage_AFrameEndsWhenAContainerRepeats(t *testing.T) {
	// Two frames back to back with nothing between them but the rows.
	rows := string(fixture(t, "stats-no-stream.jsonl"))
	f := usageDocker(t, []byte(rows+rows), nil, false)
	m := usageManager(t, f)
	c, _ := startUsage(t, m)

	samples := c.wait(t, 2)
	if samples[0].Containers != 16 || samples[1].Containers != 16 {
		t.Errorf("samples hold %d and %d Containers, want 16 and 16", samples[0].Containers, samples[1].Containers)
	}
}

func TestUsage_NoRunningContainerIsAnEmptySampleNotAnError(t *testing.T) {
	// What `docker stats` prints while nothing runs: only the screen clearing.
	f := usageDocker(t, []byte("\x1b[H\x1b[K\n\x1b[J\x1b[H\x1b[K\n\x1b[J"), nil, true)
	m := usageManager(t, f)
	c, _ := startUsage(t, m)

	samples := c.wait(t, 2)
	for i, s := range samples[:2] {
		if !s.DockerReachable || s.Containers != 0 || s.Total != (compose.Measures{}) || len(s.Services) != 0 {
			t.Errorf("sample %d = %+v, want reachable and empty", i, s)
		}
	}
}

// usageRow is a row of `docker stats --format json`.
func usageRow(name, cpu, mem, net, block string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	id := strings.Repeat(fmt.Sprintf("%016x", h.Sum64()), 4) // 64 hex digits, one per Container
	return fmt.Sprintf(`{"BlockIO":%q,"CPUPerc":%q,"Container":%q,"ID":%q,"MemPerc":"0.00%%","MemUsage":%q,"Name":%q,"NetIO":%q,"PIDs":"1"}`,
		block, cpu, id, id[:12], mem, name, net)
}

// oneFrame is a stream of a single frame, in the shape docker prints it.
func oneFrame(rows ...string) string {
	return "\x1b[H" + strings.Join(rows, "\x1b[K\n") + "\x1b[K\n\x1b[K\n\x1b[J"
}

// firstSample feeds the given stream to Usage and returns its first sample.
func firstSample(t *testing.T, stream string) compose.UsageSample {
	t.Helper()
	m := usageManager(t, usageDocker(t, []byte(stream), nil, true))
	c, _ := startUsage(t, m)
	return c.wait(t, 1)[0]
}

func TestUsage_ParsesTheUnitsOfDockerStats(t *testing.T) {
	tests := []struct {
		name               string
		cpu, mem, net, blk string
		want               compose.Measures
	}{
		{
			name: "binary memory, decimal rete and disco",
			cpu:  "74.64%", mem: "1.5GiB / 31.03GiB", net: "12.3kB / 0B", blk: "127MB / 8.19kB",
			want: compose.Measures{CPUPercent: 74.64, MemoryBytes: 1610612736, NetInBytes: 12300, NetOutBytes: 0, BlockReadBytes: 127000000, BlockWriteBytes: 8190},
		},
		{
			name: "binary memory units from KiB to TiB",
			cpu:  "0.00%", mem: "440KiB / 1GiB", net: "0B / 0B", blk: "0B / 0B",
			want: compose.Measures{MemoryBytes: 450560},
		},
		{
			name: "MiB with decimals",
			cpu:  "0.01%", mem: "20.68MiB / 31.03GiB", net: "0B / 0B", blk: "0B / 0B",
			want: compose.Measures{CPUPercent: 0.01, MemoryBytes: 21684552}, // 20.68 * 1048576, rounded
		},
		{
			name: "TiB",
			cpu:  "0.00%", mem: "1.25TiB / 2TiB", net: "0B / 0B", blk: "0B / 0B",
			want: compose.Measures{MemoryBytes: 1374389534720},
		},
		{
			name: "decimal GB, MB and plain bytes",
			cpu:  "1.15%", mem: "0B / 0B", net: "6.26GB / 126B", blk: "299MB / 6.61GB",
			want: compose.Measures{CPUPercent: 1.15, NetInBytes: 6260000000, NetOutBytes: 126, BlockReadBytes: 299000000, BlockWriteBytes: 6610000000},
		},
		{
			name: "decimal TB",
			cpu:  "0.00%", mem: "0B / 0B", net: "1.5TB / 0B", blk: "0B / 2TB",
			want: compose.Measures{NetInBytes: 1500000000000, BlockWriteBytes: 2000000000000},
		},
		{
			name: "CPU above 100% (several cores)",
			cpu:  "372.50%", mem: "0B / 0B", net: "0B / 0B", blk: "0B / 0B",
			want: compose.Measures{CPUPercent: 372.5},
		},
		{
			name: "placeholder -- counts as zero",
			cpu:  "--", mem: "-- / --", net: "-- / --", blk: "-- / --",
			want: compose.Measures{},
		},
		{
			name: "empty fields count as zero",
			cpu:  "", mem: "", net: "", blk: "",
			want: compose.Measures{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := firstSample(t, oneFrame(usageRow("sunlinkdocker-db-1", tt.cpu, tt.mem, tt.net, tt.blk)))
			if s.Containers != 1 {
				t.Fatalf("containers = %d, want 1: %+v", s.Containers, s)
			}
			got := s.Total
			if !near(got.CPUPercent, tt.want.CPUPercent) {
				t.Errorf("CPU = %v, want %v", got.CPUPercent, tt.want.CPUPercent)
			}
			got.CPUPercent, tt.want.CPUPercent = 0, 0
			if got != tt.want {
				t.Errorf("measures = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUsage_ARowWithAnUnknownUnitIsLeftOut(t *testing.T) {
	s := firstSample(t, oneFrame(
		usageRow("sunlinkdocker-db-1", "1.00%", "10MiB / 1GiB", "1kB / 1kB", "1kB / 1kB"),
		usageRow("sunlinkdocker-odoo-1", "1.00%", "10furlongs / 1GiB", "1kB / 1kB", "1kB / 1kB"),
	))
	if s.Containers != 1 || s.Total.MemoryBytes != 10485760 {
		t.Errorf("sample = %+v, want only the readable Container", s)
	}
}

// usageOf finds the usage of a Service of a Project.
func usageOf(s compose.UsageSample, project, name string) (compose.Measures, bool) {
	for _, u := range s.Services {
		if u.Project == project && u.Service == name {
			return u.Measures, true
		}
	}
	return compose.Measures{}, false
}

func TestUsage_AServiceIsTheSumOfItsReplicas_AttributedByContainerName(t *testing.T) {
	// Derived from the real ps output: a second replica of db, as compose names
	// it. The Containers are matched to the snapshot by name.
	var ps []string
	for _, line := range strings.Split(strings.TrimSpace(string(fixture(t, "ps-sunlinkdocker.json"))), "\n") {
		ps = append(ps, line)
		if strings.Contains(line, `"Name":"sunlinkdocker-db-1"`) {
			ps = append(ps, strings.ReplaceAll(line, "sunlinkdocker-db-1", "sunlinkdocker-db-2"))
		}
	}
	f := usageDocker(t, []byte(oneFrame(
		usageRow("sunlinkdocker-db-1", "0.18%", "118.7MiB / 31.03GiB", "226MB / 374MB", "298MB / 2.43GB"),
		usageRow("sunlinkdocker-db-2", "1.00%", "100MiB / 31.03GiB", "1MB / 2MB", "3MB / 4MB"),
		usageRow("sunlinkdocker-odoo-1", "6.98%", "207.5MiB / 2GiB", "6.26GB / 11.7GB", "489MB / 22.7MB"),
		usageRow("sunlink-vpn", "1.77%", "44.19MiB / 31.03GiB", "1.08GB / 1.01GB", "99.7MB / 15.8MB"),
	)), nil, true)
	f.Respond([]byte(strings.Join(ps, "\n")), psArgs("sunlinkdocker")...)
	m := usageManager(t, f)
	c, _ := startUsage(t, m)

	s := c.wait(t, 1)[0]
	db, ok := usageOf(s, "sunlinkdocker", "db")
	if !ok {
		t.Fatalf("no usage for sunlinkdocker/db: %+v", s.Services)
	}
	want := compose.Measures{
		CPUPercent: 1.18, MemoryBytes: 124465971 + 104857600, // 118.7MiB + 100MiB
		NetInBytes: 227000000, NetOutBytes: 376000000, BlockReadBytes: 301000000, BlockWriteBytes: 2434000000,
	}
	if !near(db.CPUPercent, want.CPUPercent) {
		t.Errorf("db CPU = %.2f, want 1.18", db.CPUPercent)
	}
	db.CPUPercent, want.CPUPercent = 0, 0
	if db != want {
		t.Errorf("db = %+v, want %+v", db, want)
	}
	if odoo, ok := usageOf(s, "sunlinkdocker", "odoo"); !ok || !near(odoo.CPUPercent, 6.98) {
		t.Errorf("odoo = %+v (%v), want its own Container only", odoo, ok)
	}
	// The Container of a Project that is not saved is in the global figures only.
	if len(s.Services) != 2 {
		t.Errorf("services = %+v, want db and odoo only", s.Services)
	}
	if s.Containers != 4 {
		t.Errorf("containers = %d, want 4", s.Containers)
	}
}

func TestUsage_AServiceWithoutRunningContainersHasNoMeasures(t *testing.T) {
	m := usageManager(t, followingDocker(t))
	c, _ := startUsage(t, m)

	s := c.wait(t, 1)[0]
	// stellantis-vpn is a saved Project whose Container is not in the stream.
	for _, u := range s.Services {
		if u.Project == "stellantisvpn" {
			t.Errorf("stopped Project has usage %+v, want none (not zeros)", u)
		}
	}
	if _, ok := usageOf(s, "personalvpn", "personal-vpn"); !ok {
		t.Errorf("running Service personalvpn has no usage: %+v", s.Services)
	}
	if _, ok := usageOf(s, "sunlinkdocker", "db"); !ok {
		t.Errorf("running Service sunlinkdocker/db has no usage: %+v", s.Services)
	}
}

func TestUsage_DockerUnreachableReportsAnErrorStateWithoutValues(t *testing.T) {
	f := followingDocker(t)
	f.RespondErr(nil, errors.New("Cannot connect to the Docker daemon"), infoArgs...)
	m := usageManager(t, f)
	c, _ := startUsage(t, m)

	s := c.wait(t, 1)[0]
	if s.DockerReachable {
		t.Error("docker must be reported unreachable")
	}
	if s.Containers != 0 || s.Total != (compose.Measures{}) || len(s.Services) != 0 || s.CPUs != 0 {
		t.Errorf("error sample carries values: %+v", s)
	}
}

func TestUsage_AStreamThatEndsIsAnErrorStateNotStaleValues_AndResumesByItself(t *testing.T) {
	// The stream ends by itself, with a failure.
	f := usageDocker(t, fixture(t, "stats-stream.txt"), errors.New("exit status 1"), false)
	m := usageManager(t, f)
	c, _ := startUsage(t, m)

	samples := c.wait(t, 7) // 3 frames, the failure, then the same 3 frames again
	for i, want := range []bool{true, true, true, false, true, true, true} {
		if samples[i].DockerReachable != want {
			t.Errorf("sample %d reachable = %v, want %v", i, samples[i].DockerReachable, want)
		}
	}
	if e := samples[3]; e.Containers != 0 || e.Total != (compose.Measures{}) || len(e.Services) != 0 {
		t.Errorf("sample after the failure keeps the old values: %+v", e)
	}
}

func TestUsage_NeverNotifiesWatchNorRefreshesTheSnapshot(t *testing.T) {
	f := followingDocker(t)
	m := usageManager(t, f)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	snaps := m.Watch(ctx)
	<-snaps // the current snapshot
	before := len(f.Calls())

	c, stop := startUsage(t, m)
	c.wait(t, 3)
	_ = stop()

	select {
	case s, ok := <-snaps:
		t.Errorf("Watch notified during the usage stream: %v %+v", ok, s)
	case <-time.After(50 * time.Millisecond):
	}
	for _, call := range f.Calls()[before:] {
		if k := strings.Join(call, " "); k != strings.Join(infoArgs, " ") && k != strings.Join(statsArgs, " ") {
			t.Errorf("usage ran %q, only info and stats are expected", k)
		}
	}
}

func TestUsage_EndsOnCancellationWithoutError_AndLeaksNoGoroutine(t *testing.T) {
	f := followingDocker(t)
	m := usageManager(t, f)
	before := runtime.NumGoroutine() // after the database and the Manager are up

	c, stop := startUsage(t, m)
	c.wait(t, 3)
	if err := stop(); err != nil {
		t.Errorf("a cancelled stream is not an error, got %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if after := runtime.NumGoroutine(); after > before {
		t.Errorf("goroutines: %d before, %d after the stream ended", before, after)
	}
}

func TestUsage_CancelledWhileDockerIsDownReturnsAtOnce(t *testing.T) {
	f := followingDocker(t)
	f.RespondErr(nil, errors.New("down"), infoArgs...)
	m := usageManager(t, f)
	c, stop := startUsage(t, m)
	c.wait(t, 1)
	if err := stop(); err != nil {
		t.Errorf("Usage = %v, want nil", err)
	}
}
