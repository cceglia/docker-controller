package gui

import "dockercontroller/internal/compose"

const (
	// waitingUsageText is the line before the first sample arrives.
	waitingUsageText = "In attesa della prima misura…"
	// unreachableUsageText is the warning that takes the place of the figures.
	unreachableUsageText = "Docker non raggiungibile"
	// usageTitle names the figures: they are the machine's, not the Project's.
	usageTitle = "Utilizzo risorse"
)

// UsageState is what a Service block shows for the Utilizzo risorse.
type UsageState int

const (
	// UsageFigures are the measures of the Service.
	UsageFigures UsageState = iota
	// UsageWaiting is the text before the first sample, while the Service runs.
	UsageWaiting
	// UsageNone is the dash: the Service has no measures, which is not zero.
	UsageNone
)

// UsageStateOf tells what the block of a Service shows from the sample and the
// entry of the Service in it. A Service without a running Container, or a
// sample that carries no figures, shows the dash and never zero.
func UsageStateOf(sample *compose.UsageSample, hasEntry, running bool) UsageState {
	switch {
	case sample == nil:
		if running {
			return UsageWaiting
		}
		return UsageNone
	case !sample.DockerReachable:
		return UsageNone
	case !hasEntry:
		return UsageNone
	}
	return UsageFigures
}

// ServiceRuns reports whether a Service has a Container in execution. Only
// then can it have measures: a Service with no running Container shows the dash.
func ServiceRuns(service compose.Service) bool {
	for _, container := range service.Containers {
		if container.State == "running" {
			return true
		}
	}
	return false
}

// StripState is what the global bar shows.
type StripState int

const (
	// StripWaiting is the text before the first sample.
	StripWaiting StripState = iota
	// StripUnreachable is the warning of a Docker that does not answer.
	StripUnreachable
	// StripFigures are the figures of the machine.
	StripFigures
)

// StripStateOf tells what the global bar shows from the sample. A nil sample is
// a stream that has not delivered yet or that the view stopped.
func StripStateOf(sample *compose.UsageSample) StripState {
	switch {
	case sample == nil:
		return StripWaiting
	case !sample.DockerReachable:
		return StripUnreachable
	}
	return StripFigures
}
