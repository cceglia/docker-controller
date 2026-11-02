package model_test

import (
	"testing"

	"dockercontroller/internal/compose/model"
)

func TestMeasures_Add(t *testing.T) {
	tests := []struct {
		name string
		a, b model.Measures
		want model.Measures
	}{
		{
			name: "zero plus zero",
			want: model.Measures{},
		},
		{
			name: "every figure",
			a: model.Measures{
				CPUPercent: 1.5, MemoryBytes: 2,
				NetInBytes: 3, NetOutBytes: 4,
				BlockReadBytes: 5, BlockWriteBytes: 6,
			},
			b: model.Measures{
				CPUPercent: 0.5, MemoryBytes: 10,
				NetInBytes: 20, NetOutBytes: 30,
				BlockReadBytes: 40, BlockWriteBytes: 50,
			},
			want: model.Measures{
				CPUPercent: 2, MemoryBytes: 12,
				NetInBytes: 23, NetOutBytes: 34,
				BlockReadBytes: 45, BlockWriteBytes: 56,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a
			got.Add(tt.b)
			if got != tt.want {
				t.Errorf("Add() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
