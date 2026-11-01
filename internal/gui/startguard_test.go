package gui

import (
	"slices"
	"testing"
)

func TestStartGuard_StartsOneActivityPerClick(t *testing.T) {
	const (
		click = iota
		result
		busySnapshot
		idleSnapshot
	)

	for _, tc := range []struct {
		name  string
		steps []int
		want  []bool
	}{
		{
			"one click starts one Activity",
			[]int{click},
			[]bool{true},
		},
		{
			"the second click of a double click starts nothing",
			[]int{click, click},
			[]bool{true, false},
		},
		{
			"the result of the launch opens the guard again",
			[]int{click, result, click},
			[]bool{true, true},
		},
		{
			"the snapshot of the Activity opens the guard again",
			[]int{click, busySnapshot, click},
			[]bool{true, true},
		},
		{
			"a snapshot without an Activity keeps the guard closed",
			[]int{click, idleSnapshot, click},
			[]bool{true, false},
		},
		{
			"the guard of a refused launch opens again",
			[]int{click, result, click, click},
			[]bool{true, true, false},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var guard startGuard
			var got []bool
			for _, step := range tc.steps {
				switch step {
				case click:
					got = append(got, guard.begin())
				case result:
					guard.release()
				case busySnapshot:
					guard.snapshot(true)
				case idleSnapshot:
					guard.snapshot(false)
				}
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("begin() = %v, want %v", got, tc.want)
			}
		})
	}
}
