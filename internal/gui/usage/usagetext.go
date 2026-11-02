package usage

import (
	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/view"
)

// Text is the figure of one measure and the line under it. An empty note
// leaves the measure on one line.
type Text struct {
	Value string
	Note  string
}

// Texts are the four measures of a row of the Utilizzo risorse, ready to
// draw: the CPU, the memory, the network and the disk.
type Texts struct {
	CPU    Text
	Memory Text
	Net    Text
	Disk   Text
}

// Figures is the input of a row of the Utilizzo risorse: the measures of the
// machine or of the Containers of a Service, the rates that come with them and
// the size of the machine. CPUPercent is the figure of the bar, which is the
// share of the machine or the percentage of the Service itself.
type Figures struct {
	Measures    compose.Measures
	Rate        compose.Rates
	CPUPercent  float64
	MemoryTotal uint64 // the memory of the machine, 0 when it is unknown
}

// TextsOf turns the figures of one row into the texts to draw. The notes say
// what a figure alone cannot: the cores of the CPU, the memory of the machine
// the used memory sits in, and the other direction of a rate. A machine of
// unknown size leaves the memory without a note.
func TextsOf(f Figures) Texts {
	texts := Texts{
		CPU: Text{
			Value: view.FormatPercent(f.CPUPercent),
			Note:  view.FormatCores(f.Measures.CPUPercent),
		},
		Memory: Text{
			Value: view.FormatBytes(float64(f.Measures.MemoryBytes), view.ByteBaseBinary),
		},
		Net: Text{
			Value: "↓ " + view.FormatRate(f.Rate.NetInBytesPerSec),
			Note:  "↑ " + view.FormatRate(f.Rate.NetOutBytesPerSec),
		},
		Disk: Text{
			Value: "R " + view.FormatRate(f.Rate.BlockReadBytesPerSec),
			Note:  "W " + view.FormatRate(f.Rate.BlockWriteBytesPerSec),
		},
	}
	if f.MemoryTotal > 0 {
		texts.Memory.Note = "di " + view.FormatBytes(float64(f.MemoryTotal), view.ByteBaseBinary)
	}
	return texts
}
