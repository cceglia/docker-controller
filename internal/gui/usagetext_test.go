package gui

import (
	"testing"

	"dockercontroller/internal/compose"
)

// The expected texts come from the old view for the same figures, so the two
// agree on the digits, the units and the Italian comma.

func TestUsageTextsOf_TheFourMeasuresOfARow(t *testing.T) {
	rate := compose.Rates{
		NetInBytesPerSec:      1234.5,
		NetOutBytesPerSec:     340,
		BlockReadBytesPerSec:  1234.5,
		BlockWriteBytesPerSec: 340,
	}
	for _, tc := range []struct {
		name    string
		figures usageFigures
		want    usageTexts
	}{
		{
			name: "the machine, whose memory has a size to sit in",
			figures: usageFigures{
				measures:    compose.Measures{CPUPercent: 49.36, MemoryBytes: 1 << 30},
				rate:        rate,
				cpuPercent:  12.34,
				memoryTotal: 16 << 30,
			},
			want: usageTexts{
				cpu:    usageText{value: "12,3%", note: "0,49 core"},
				memory: usageText{value: "1,00 GiB", note: "di 16,0 GiB"},
				net:    usageText{value: "↓ 1,23 kB/s", note: "↑ 340 B/s"},
				disk:   usageText{value: "R 1,23 kB/s", note: "W 340 B/s"},
			},
		},
		{
			name: "a Service, whose memory has no machine to sit in",
			figures: usageFigures{
				measures:   compose.Measures{CPUPercent: 250, MemoryBytes: 512 << 20},
				rate:       compose.Rates{},
				cpuPercent: 250,
			},
			want: usageTexts{
				cpu:    usageText{value: "250,0%", note: "2,50 core"},
				memory: usageText{value: "512 MiB"},
				net:    usageText{value: "↓ 0 B/s", note: "↑ 0 B/s"},
				disk:   usageText{value: "R 0 B/s", note: "W 0 B/s"},
			},
		},
		{
			name:    "an idle row without measures",
			figures: usageFigures{},
			want: usageTexts{
				cpu:    usageText{value: "0,0%", note: "0,00 core"},
				memory: usageText{value: "0 B"},
				net:    usageText{value: "↓ 0 B/s", note: "↑ 0 B/s"},
				disk:   usageText{value: "R 0 B/s", note: "W 0 B/s"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := usageTextsOf(tc.figures); got != tc.want {
				t.Errorf("usageTextsOf = %+v, want %+v", got, tc.want)
			}
		})
	}
}
