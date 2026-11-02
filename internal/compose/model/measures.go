package model

// Measures is the Utilizzo risorse of a set of Containers: the sum of their
// figures at one instant.
type Measures struct {
	// CPUPercent is docker's own figure summed: 100 is one fully used core, so
	// it can exceed 100.
	CPUPercent  float64 `json:"cpuPercent"`
	MemoryBytes uint64  `json:"memoryBytes"`
	// NetInBytes and NetOutBytes are the bytes received and sent since the
	// Containers started.
	NetInBytes  uint64 `json:"netInBytes"`
	NetOutBytes uint64 `json:"netOutBytes"`
	// BlockReadBytes and BlockWriteBytes are the disk I/O since the Containers
	// started.
	BlockReadBytes  uint64 `json:"blockReadBytes"`
	BlockWriteBytes uint64 `json:"blockWriteBytes"`
}

// Add adds the measures of another set of Containers.
func (m *Measures) Add(o Measures) {
	m.CPUPercent += o.CPUPercent
	m.MemoryBytes += o.MemoryBytes
	m.NetInBytes += o.NetInBytes
	m.NetOutBytes += o.NetOutBytes
	m.BlockReadBytes += o.BlockReadBytes
	m.BlockWriteBytes += o.BlockWriteBytes
}
