package usage

import (
	"testing"

	"dockercontroller/internal/compose"
)

// The expected texts come from the old view for the same figures, so the two
// agree on the digits, the units and the Italian comma.

func TestTextsOf_TheFourMeasuresOfARow(t *testing.T) {
	rate := compose.Rates{
		NetInBytesPerSec:      1234.5,
		NetOutBytesPerSec:     340,
		BlockReadBytesPerSec:  1234.5,
		BlockWriteBytesPerSec: 340,
	}
	for _, tc := range []struct {
		name    string
		figures Figures
		want    Texts
	}{
		{
			name: "the machine, whose memory has a size to sit in",
			figures: Figures{
				Measures:    compose.Measures{CPUPercent: 49.36, MemoryBytes: 1 << 30},
				Rate:        rate,
				CPUPercent:  12.34,
				MemoryTotal: 16 << 30,
			},
			want: Texts{
				CPU:    Text{Value: "12,3%", Note: "0,49 core"},
				Memory: Text{Value: "1,00 GiB", Note: "di 16,0 GiB"},
				Net:    Text{Value: "↓ 1,23 kB/s", Note: "↑ 340 B/s"},
				Disk:   Text{Value: "R 1,23 kB/s", Note: "W 340 B/s"},
			},
		},
		{
			name: "a Service, whose memory has no machine to sit in",
			figures: Figures{
				Measures:   compose.Measures{CPUPercent: 250, MemoryBytes: 512 << 20},
				Rate:       compose.Rates{},
				CPUPercent: 250,
			},
			want: Texts{
				CPU:    Text{Value: "250,0%", Note: "2,50 core"},
				Memory: Text{Value: "512 MiB"},
				Net:    Text{Value: "↓ 0 B/s", Note: "↑ 0 B/s"},
				Disk:   Text{Value: "R 0 B/s", Note: "W 0 B/s"},
			},
		},
		{
			name:    "an idle row without measures",
			figures: Figures{},
			want: Texts{
				CPU:    Text{Value: "0,0%", Note: "0,00 core"},
				Memory: Text{Value: "0 B"},
				Net:    Text{Value: "↓ 0 B/s", Note: "↑ 0 B/s"},
				Disk:   Text{Value: "R 0 B/s", Note: "W 0 B/s"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := TextsOf(tc.figures); got != tc.want {
				t.Errorf("TextsOf = %+v, want %+v", got, tc.want)
			}
		})
	}
}
