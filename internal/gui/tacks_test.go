package gui

import (
	"testing"

	"dockercontroller/internal/compose"
)

func TestContainerTack_ReadsTheStateOfOneContainer(t *testing.T) {
	for _, tc := range []struct {
		name      string
		container compose.Container
		want      TackState
	}{
		{"running", compose.Container{State: "running"}, TackRunning},
		{"running and healthy", compose.Container{State: "running", Health: "healthy"}, TackRunning},
		{"running but unhealthy", compose.Container{State: "running", Health: "unhealthy"}, TackFaulty},
		{"dead", compose.Container{State: "dead"}, TackFaulty},
		{"exited", compose.Container{State: "exited"}, TackStopped},
		{"created", compose.Container{State: "created"}, TackStopped},
		{"restarting", compose.Container{State: "restarting"}, TackStopped},
		{"paused", compose.Container{State: "paused"}, TackStopped},
		{"starting healthcheck", compose.Container{State: "running", Health: "starting"}, TackRunning},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ContainerTack(tc.container); got != tc.want {
				t.Errorf("ContainerTack(%+v) = %v, want %v", tc.container, got, tc.want)
			}
		})
	}
}

func TestTacks_DrawsTwelveThenCountsTheRest(t *testing.T) {
	services := []compose.Service{
		{Name: "web", Containers: []compose.Container{
			{State: "running"},
			{State: "exited"},
		}},
		{Name: "db", Containers: []compose.Container{
			{State: "running", Health: "unhealthy"},
		}},
	}

	shown, more := Tacks(services)

	want := []TackState{TackRunning, TackStopped, TackFaulty}
	if len(shown) != len(want) || more != 0 {
		t.Fatalf("Tacks = %v, +%d, want %v, +0", shown, more, want)
	}
	for i := range want {
		if shown[i] != want[i] {
			t.Errorf("tack %d = %v, want %v", i, shown[i], want[i])
		}
	}
}

func TestTacks_BeyondTwelveBecomesAPlusCount(t *testing.T) {
	var containers []compose.Container
	for i := 0; i < MaxTacks+5; i++ {
		containers = append(containers, compose.Container{State: "running"})
	}

	shown, more := Tacks([]compose.Service{{Name: "web", Containers: containers}})

	if len(shown) != MaxTacks {
		t.Errorf("%d tacks, want %d", len(shown), MaxTacks)
	}
	if more != 5 {
		t.Errorf("+%d, want +5", more)
	}
}

func TestTacks_NoContainerDrawsNothing(t *testing.T) {
	shown, more := Tacks(nil)

	if len(shown) != 0 || more != 0 {
		t.Errorf("Tacks(nil) = %v, +%d, want nothing", shown, more)
	}
}
