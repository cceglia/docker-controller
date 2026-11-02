package usage

import (
	"math"
	"testing"

	"dockercontroller/internal/compose"
)

func TestServiceRuns_FindsAnExecutingContainer(t *testing.T) {
	for _, tc := range []struct {
		name       string
		containers []compose.Container
		want       bool
	}{
		{"no container", nil, false},
		{"all stopped", []compose.Container{{State: "exited"}, {State: "created"}}, false},
		{"one running", []compose.Container{{State: "exited"}, {State: "running"}}, true},
		{"a restarting container does not run", []compose.Container{{State: "restarting"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := compose.Service{Name: "web", Containers: tc.containers}
			if got := ServiceRuns(service); got != tc.want {
				t.Errorf("ServiceRuns = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestUsageStateOf_NeverShowsZeroWithoutMeasures(t *testing.T) {
	reachable := &compose.UsageSample{DockerReachable: true}
	unreachable := &compose.UsageSample{}

	for _, tc := range []struct {
		name     string
		sample   *compose.UsageSample
		hasEntry bool
		running  bool
		want     UsageState
	}{
		{"first sample not arrived and the Service runs", nil, false, true, UsageWaiting},
		{"first sample not arrived and nothing runs", nil, false, false, UsageNone},
		{"stopped stream and the Service runs", nil, false, true, UsageWaiting},
		{"a sample lists the Service", reachable, true, true, UsageFigures},
		{"a sample leaves the Service out", reachable, false, true, UsageNone},
		{"a stopped Service in a full sample", reachable, false, false, UsageNone},
		{"Docker does not answer", unreachable, false, true, UsageNone},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := UsageStateOf(tc.sample, tc.hasEntry, tc.running); got != tc.want {
				t.Errorf("UsageStateOf = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMeterValue_StepsOfTenthsOfAPercent(t *testing.T) {
	for _, tc := range []struct {
		name    string
		percent float64
		want    int
	}{
		{"zero", 0, 0},
		{"half", 50, 500},
		{"all", 100, 1000},
		{"a decimal of a percent", 12.34, 123},
		{"above all stays inside the bar", 150, 1000},
		{"below zero stays inside the bar", -5, 0},
		{"not a number", math.NaN(), 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := MeterValue(tc.percent); got != tc.want {
				t.Errorf("MeterValue(%v) = %d, want %d", tc.percent, got, tc.want)
			}
		})
	}
}

func TestStripStateOf_ReplacesTheFiguresWithAMessage(t *testing.T) {
	for _, tc := range []struct {
		name   string
		sample *compose.UsageSample
		want   StripState
	}{
		{"no sample yet", nil, StripWaiting},
		{"Docker does not answer", &compose.UsageSample{Error: "no docker"}, StripUnreachable},
		{"a good sample", &compose.UsageSample{DockerReachable: true}, StripFigures},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := StripStateOf(tc.sample); got != tc.want {
				t.Errorf("StripStateOf = %v, want %v", got, tc.want)
			}
		})
	}
}
