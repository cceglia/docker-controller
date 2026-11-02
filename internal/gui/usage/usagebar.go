package usage

import (
	"math"

	"dockercontroller/internal/gui/view"
)

// MeterSteps is the number of steps of a bar of a percentage: the tenth of a
// percent is the smallest change a row shows.
const MeterSteps = 1000

// MeterValue turns a percentage into the steps of a bar: a value that is not a
// number and a value outside 0-100 stay inside the bar.
func MeterValue(percent float64) int {
	return int(math.Round(view.LevelValue(percent) * MeterSteps))
}
