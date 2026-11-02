package model_test

import (
	"path/filepath"
	"testing"

	"dockercontroller/internal/compose/model"
)

func TestStatusOf(t *testing.T) {
	tests := []struct {
		name    string
		running int
		total   int
		want    model.Status
	}{
		{name: "no containers is removed", running: 0, total: 0, want: model.StatusRemoved},
		{name: "all running is active", running: 3, total: 3, want: model.StatusActive},
		{name: "none running is stopped", running: 0, total: 3, want: model.StatusStopped},
		{name: "some running is partial", running: 1, total: 3, want: model.StatusPartial},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := model.StatusOf(tt.running, tt.total); got != tt.want {
				t.Errorf("StatusOf(%d, %d) = %q, want %q", tt.running, tt.total, got, tt.want)
			}
		})
	}
}

func TestFolder(t *testing.T) {
	tests := []struct {
		name         string
		composeFiles []string
		want         string
	}{
		{name: "nil files", composeFiles: nil, want: ""},
		{name: "empty list", composeFiles: []string{}, want: ""},
		{name: "one file", composeFiles: []string{filepath.Join("/srv", "app", "compose.yml")}, want: filepath.Join("/srv", "app")},
		{name: "many files uses the first", composeFiles: []string{filepath.Join("/srv", "app", "compose.yml"), filepath.Join("/srv", "app", "override.yml")}, want: filepath.Join("/srv", "app")},
		{name: "file without directory", composeFiles: []string{"compose.yml"}, want: "."},
		{name: "relative path", composeFiles: []string{filepath.Join("a", "b", "compose.yml")}, want: filepath.Join("a", "b")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := model.Folder(tt.composeFiles); got != tt.want {
				t.Errorf("Folder(%v) = %q, want %q", tt.composeFiles, got, tt.want)
			}
		})
	}
}
