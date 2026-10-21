package compose

import (
	"testing"
	"time"
)

func row(id string, netIn, blockWrite uint64) statsRow {
	return statsRow{id: id, measures: Measures{NetInBytes: netIn, BlockWriteBytes: blockWrite}}
}

func TestRateTracker_IsTheCounterDeltaOverTheElapsedTime(t *testing.T) {
	var r rateTracker
	t0 := time.Unix(100, 0)
	if got := r.observe(t0, []statsRow{row("a", 1000, 0)}); got != (Rates{}) {
		t.Fatalf("first frame = %+v, want zero", got)
	}
	got := r.observe(t0.Add(2*time.Second), []statsRow{row("a", 5000, 400)})
	if got.NetInBytesPerSec != 2000 || got.BlockWriteBytesPerSec != 200 {
		t.Errorf("rate = %+v, want 2000 in and 200 written per second", got)
	}
}

func TestRateTracker_ANewOrRestartedContainerNeverGivesASpikeOrANegativeRate(t *testing.T) {
	var r rateTracker
	t0 := time.Unix(100, 0)
	r.observe(t0, []statsRow{row("a", 9000, 0)})
	got := r.observe(t0.Add(time.Second), []statsRow{row("a", 100, 0), row("b", 7000, 0)})
	if got.NetInBytesPerSec != 0 {
		t.Errorf("rate = %v, want 0 (counter reset, new container)", got.NetInBytesPerSec)
	}
}
