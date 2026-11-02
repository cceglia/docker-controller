package compose

import (
	"testing"
	"time"

	"dockercontroller/internal/compose/parse"
)

func row(id string, netIn, blockWrite uint64) parse.StatsRow {
	return parse.StatsRow{ID: id, Measures: Measures{NetInBytes: netIn, BlockWriteBytes: blockWrite}}
}

func TestRateTracker_IsTheCounterDeltaOverTheElapsedTimeOfEachContainer(t *testing.T) {
	var r rateTracker
	t0 := time.Unix(100, 0)
	if got := r.observe(t0, []parse.StatsRow{row("a", 1000, 0), row("c", 500, 0)}); len(got) != 0 {
		t.Fatalf("first frame = %+v, want no rate (there is no previous frame)", got)
	}
	got := r.observe(t0.Add(2*time.Second), []parse.StatsRow{row("a", 5000, 400), row("c", 1500, 0)})
	if a := got["a"]; a.NetInBytesPerSec != 2000 || a.BlockWriteBytesPerSec != 200 {
		t.Errorf("rate of a = %+v, want 2000 in and 200 written per second", a)
	}
	if c := got["c"]; c.NetInBytesPerSec != 500 {
		t.Errorf("rate of c = %+v, want 500 in per second", c)
	}
}

func TestRateTracker_ANewContainerHasNoRateAndARestartedOneHasZero(t *testing.T) {
	var r rateTracker
	t0 := time.Unix(100, 0)
	r.observe(t0, []parse.StatsRow{row("a", 9000, 0)})
	got := r.observe(t0.Add(time.Second), []parse.StatsRow{row("a", 100, 0), row("b", 7000, 0)})
	if a, ok := got["a"]; !ok || a != (Rates{}) {
		t.Errorf("restarted Container a = %+v (%v), want a zero rate", a, ok)
	}
	if b, ok := got["b"]; ok {
		t.Errorf("new Container b = %+v, want no rate at all (absent, not zero)", b)
	}
}
