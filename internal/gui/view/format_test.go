package view

import (
	"math"
	"testing"
)

// The expected strings are the output of the old TypeScript view for the same
// input, so the figures of the GTK view and the old ones agree.

func TestFormatBytes_BinaryUnits(t *testing.T) {
	for _, tc := range []struct {
		name  string
		bytes float64
		want  string
	}{
		{"zero bytes", 0, "0 B"},
		{"one byte", 1, "1 B"},
		{"bytes only", 512, "512 B"},
		{"just under a kibibyte", 1023, "1023 B"},
		{"one kibibyte", 1024, "1,00 KiB"},
		{"one and a half kibibytes", 1536, "1,50 KiB"},
		{"ten kibibytes", 10240, "10,0 KiB"},
		{"one mebibyte", 1048576, "1,00 MiB"},
		{"one and a half mebibytes", 1572864, "1,50 MiB"},
		{"past a hundred mebibytes", 124465971, "119 MiB"},
		{"a hundred kibibytes drop their decimals", 102400, "100 KiB"},
		{"one GiB and some", 33318208799, "31,0 GiB"},
		{"four digits keep no separator", 1048166, "1024 KiB"},
		{"the last byte of a kibibyte carries over", 1048575, "1024 KiB"},
		{"a large size", 1e15, "909 TiB"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatBytes(tc.bytes, ByteBaseBinary); got != tc.want {
				t.Errorf("FormatBytes(%v, binary) = %q, want %q", tc.bytes, got, tc.want)
			}
		})
	}
}

func TestFormatBytes_DecimalUnits(t *testing.T) {
	for _, tc := range []struct {
		name  string
		bytes float64
		want  string
	}{
		{"zero bytes", 0, "0 B"},
		{"bytes only", 999, "999 B"},
		{"one kilobyte", 1000, "1,00 kB"},
		{"one and a half kilobytes", 1500, "1,50 kB"},
		{"a rate with two decimals", 1234.5, "1,23 kB"},
		{"1024 bytes stay a thousand and twenty-four", 1024, "1,02 kB"},
		{"one megabyte", 1e6, "1,00 MB"},
		{"just under a megabyte", 999999, "1000 kB"},
		{"gigabytes", 6260000000, "6,26 GB"},
		{"one terabyte", 1e12, "1,00 TB"},
		{"petabytes", 2.5e15, "2,50 PB"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatBytes(tc.bytes, ByteBaseDecimal); got != tc.want {
				t.Errorf("FormatBytes(%v, decimal) = %q, want %q", tc.bytes, got, tc.want)
			}
		})
	}
}

func TestFormatPercent_OneDecimalWithTheItalianComma(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value float64
		want  string
	}{
		{"zero", 0, "0,0%"},
		{"a fraction of a percent", 0.04, "0,0%"},
		{"one decimal", 1.234, "1,2%"},
		{"rounds down", 12.34, "12,3%"},
		{"rounds up", 12.35, "12,4%"},
		{"rounds up to a hundred", 99.95, "100,0%"},
		{"a full core", 100, "100,0%"},
		{"above a hundred", 123.45, "123,5%"},
		{"half a percent", 0.5, "0,5%"},
		{"a hundredth rounds half away from zero", 0.05, "0,1%"},
		{"a negative percentage keeps its sign", -12.34, "-12,3%"},
		{"five digits take the separator", 12345.6, "12.345,6%"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatPercent(tc.value); got != tc.want {
				t.Errorf("FormatPercent(%v) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestFormatCores_TwoDecimalsOfACore(t *testing.T) {
	for _, tc := range []struct {
		name       string
		cpuPercent float64
		want       string
	}{
		{"no CPU", 0, "0,00 core"},
		{"half a percent of a core", 0.5, "0,01 core"},
		{"one percent of a core", 1, "0,01 core"},
		{"a tenth of a core", 12.34, "0,12 core"},
		{"one core and a bit", 123.4, "1,23 core"},
		{"twelve cores", 1234.5, "12,35 core"},
		{"half a core", 50, "0,50 core"},
		{"a fifth of a core", 20, "0,20 core"},
		{"a hundred cores", 10000, "100,00 core"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatCores(tc.cpuPercent); got != tc.want {
				t.Errorf("FormatCores(%v) = %q, want %q", tc.cpuPercent, got, tc.want)
			}
		})
	}
}

func TestFormatRate_UsesDecimalUnitsPerSecond(t *testing.T) {
	for _, tc := range []struct {
		name        string
		bytesPerSec float64
		want        string
	}{
		{"an idle rate", 0, "0 B/s"},
		{"just under a kilobyte per second", 999, "999 B/s"},
		{"a kilobyte per second", 1000, "1,00 kB/s"},
		{"a kilobyte and a half per second", 1234.5, "1,23 kB/s"},
		{"a megabyte and a half per second", 1.5e6, "1,50 MB/s"},
		{"gigabytes per second", 2.5e9, "2,50 GB/s"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatRate(tc.bytesPerSec); got != tc.want {
				t.Errorf("FormatRate(%v) = %q, want %q", tc.bytesPerSec, got, tc.want)
			}
		})
	}
}

func TestLevelValue_ClampsTheMeter(t *testing.T) {
	for _, tc := range []struct {
		name    string
		percent float64
		want    float64
	}{
		{"zero", 0, 0},
		{"half", 50, 0.5},
		{"all", 100, 1},
		{"above all", 150, 1},
		{"below zero", -5, 0},
		{"not a number", math.NaN(), 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := LevelValue(tc.percent); got != tc.want {
				t.Errorf("LevelValue(%v) = %v, want %v", tc.percent, got, tc.want)
			}
		})
	}
}
