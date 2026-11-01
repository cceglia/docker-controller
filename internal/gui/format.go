package gui

import (
	"strconv"
	"strings"
)

// ByteBase chooses the units of a size.
type ByteBase int

const (
	// ByteBaseBinary uses KiB, MiB and GiB: powers of 1024, for memory.
	ByteBaseBinary ByteBase = iota
	// ByteBaseDecimal uses kB, MB and GB: powers of 1000, for network and disk.
	ByteBaseDecimal
)

var binaryUnits = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}
var decimalUnits = []string{"B", "kB", "MB", "GB", "TB", "PB"}

// FormatBytes returns a size with its unit, like the old view did: memory with
// binary units (MiB, GiB) and network and disk with decimal ones (MB, GB), as
// `docker stats` does. The digits follow the size: none from a hundred units
// up, one from ten, two below.
func FormatBytes(bytes float64, base ByteBase) string {
	step, units := 1024.0, binaryUnits
	if base == ByteBaseDecimal {
		step, units = 1000.0, decimalUnits
	}
	value, i := bytes, 0
	for value >= step && i < len(units)-1 {
		value /= step
		i++
	}
	digits := 0
	switch {
	case i == 0 || value >= 100:
		digits = 0
	case value >= 10:
		digits = 1
	default:
		digits = 2
	}
	return formatNumber(value, digits) + " " + units[i]
}

// FormatPercent returns a percentage with one decimal and the Italian comma.
func FormatPercent(value float64) string {
	return formatNumber(value, 1) + "%"
}

// FormatRate returns a rate in bytes per second in decimal units, like the
// global strip shows the network and the disk.
func FormatRate(bytesPerSec float64) string {
	return FormatBytes(bytesPerSec, ByteBaseDecimal) + "/s"
}

// FormatCores returns a CPU figure in cores: docker reports 100 for one fully
// used core.
func FormatCores(cpuPercent float64) string {
	return formatNumber(cpuPercent/100, 2) + " core"
}

// levelValue clamps a percentage to the 0-1 interval of a GtkLevelBar and
// treats a value that is not a number as an empty bar.
func levelValue(percent float64) float64 {
	if percent != percent { // NaN
		return 0
	}
	value := percent / 100
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

// formatNumber renders a number the Italian way: a comma before the decimals
// and a dot every three digits, but only from five digits up, as the locale
// does. digits is the exact count of decimals and the rounding is half away
// from zero, like the old view.
func formatNumber(value float64, digits int) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	integer, fraction, _ := strings.Cut(strconv.FormatFloat(value, 'f', -1, 64), ".")
	fraction, carry := roundDecimal(fraction, digits)
	if carry {
		integer = addOne(integer)
	}
	integer = groupThousands(integer)
	if digits == 0 {
		return sign + integer
	}
	return sign + integer + "," + fraction
}

// roundDecimal rounds a run of decimal digits to n places, half away from
// zero, and reports the carry into the integer part. A shorter run is padded
// with zeros.
func roundDecimal(fraction string, n int) (string, bool) {
	kept := []byte(fraction)
	carry := false
	if len(kept) > n {
		up := kept[n] >= '5'
		kept = kept[:n]
		if up {
			carry = increment(kept) == nil
		}
	}
	for len(kept) < n {
		kept = append(kept, '0')
	}
	return string(kept), carry
}

// increment adds one to a run of digits. It returns nil when every digit was a
// nine and the carry leaves the run.
func increment(digits []byte) []byte {
	for i := len(digits) - 1; i >= 0; i-- {
		if digits[i] != '9' {
			digits[i]++
			return digits
		}
		digits[i] = '0'
	}
	return nil
}

// addOne adds one to a run of digits, growing it when it was all nines.
func addOne(digits string) string {
	out := []byte(digits)
	if increment(out) != nil {
		return string(out)
	}
	return "1" + string(out)
}

// groupThousands puts a dot every three digits, but only from five digits up,
// as the Italian locale does: 1234 stays 1234 and 12345 reads 12.345.
func groupThousands(digits string) string {
	if len(digits) < 5 {
		return digits
	}
	var out strings.Builder
	first := len(digits) % 3
	if first > 0 {
		out.WriteString(digits[:first])
	}
	for i := first; i < len(digits); i += 3 {
		if out.Len() > 0 {
			out.WriteByte('.')
		}
		out.WriteString(digits[i : i+3])
	}
	return out.String()
}
