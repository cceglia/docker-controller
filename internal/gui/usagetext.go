package gui

import (
	"dockercontroller/internal/compose"
	"dockercontroller/internal/gui/view"
)

// usageText is the figure of one measure and the line under it. An empty note
// leaves the measure on one line.
type usageText struct {
	value string
	note  string
}

// usageTexts are the four measures of a row of the Utilizzo risorse, ready to
// draw: the CPU, the memory, the network and the disk.
type usageTexts struct {
	cpu    usageText
	memory usageText
	net    usageText
	disk   usageText
}

// usageFigures is the input of a row of the Utilizzo risorse: the measures of
// the machine or of the Containers of a Service, the rates that come with them
// and the size of the machine. cpuPercent is the figure of the bar, which is the
// share of the machine or the percentage of the Service itself.
type usageFigures struct {
	measures    compose.Measures
	rate        compose.Rates
	cpuPercent  float64
	memoryTotal uint64 // the memory of the machine, 0 when it is unknown
}

// usageTextsOf turns the figures of one row into the texts to draw. The notes
// say what a figure alone cannot: the cores of the CPU, the memory of the
// machine the used memory sits in, and the other direction of a rate. A machine
// of unknown size leaves the memory without a note.
func usageTextsOf(f usageFigures) usageTexts {
	texts := usageTexts{
		cpu: usageText{
			value: view.FormatPercent(f.cpuPercent),
			note:  view.FormatCores(f.measures.CPUPercent),
		},
		memory: usageText{
			value: view.FormatBytes(float64(f.measures.MemoryBytes), view.ByteBaseBinary),
		},
		net: usageText{
			value: "↓ " + view.FormatRate(f.rate.NetInBytesPerSec),
			note:  "↑ " + view.FormatRate(f.rate.NetOutBytesPerSec),
		},
		disk: usageText{
			value: "R " + view.FormatRate(f.rate.BlockReadBytesPerSec),
			note:  "W " + view.FormatRate(f.rate.BlockWriteBytesPerSec),
		},
	}
	if f.memoryTotal > 0 {
		texts.memory.note = "di " + view.FormatBytes(float64(f.memoryTotal), view.ByteBaseBinary)
	}
	return texts
}
