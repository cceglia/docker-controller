package compose

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
)

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

func (m *Measures) add(o Measures) {
	m.CPUPercent += o.CPUPercent
	m.MemoryBytes += o.MemoryBytes
	m.NetInBytes += o.NetInBytes
	m.NetOutBytes += o.NetOutBytes
	m.BlockReadBytes += o.BlockReadBytes
	m.BlockWriteBytes += o.BlockWriteBytes
}

// ServiceUsage is the Utilizzo risorse of one Service: the sum over its
// running Containers.
type ServiceUsage struct {
	Project  string   `json:"project"`
	Service  string   `json:"service"`
	Measures Measures `json:"measures"`
}

// UsageSample is the Utilizzo risorse at one instant. It is not part of the
// Snapshot: samples arrive about every 2 seconds and never wake Watch.
type UsageSample struct {
	// DockerReachable is false while the usage cannot be read; the sample then
	// holds no figure at all (never the last known ones) and Error says why.
	DockerReachable bool   `json:"dockerReachable"`
	Error           string `json:"error"`
	// Containers is how many running Containers the figures sum, whatever
	// Project they belong to.
	Containers int `json:"containers"`
	// Total is the global usage: the sum over every running Container.
	Total Measures `json:"total"`
	// CPUs and MemoryTotalBytes are the size of the machine (docker info).
	CPUs             int    `json:"cpus"`
	MemoryTotalBytes uint64 `json:"memoryTotalBytes"`
	// CPUMachinePercent is Total's CPU over all the CPUs of the machine
	// (100 is every core fully used); MemoryMachinePercent is the memory used
	// over MemoryTotalBytes.
	CPUMachinePercent    float64 `json:"cpuMachinePercent"`
	MemoryMachinePercent float64 `json:"memoryMachinePercent"`
	// Services holds the usage of the Services of the saved Projects that have
	// running Containers, sorted by Project then Service. A Service not listed
	// has no running Container: it has no measures, which is not zero.
	Services []ServiceUsage `json:"services"`
}

// Usage streams the Utilizzo risorse: one sample per frame of
// `docker stats --format json` (about every 2 seconds), delivered to onSample,
// which is never called concurrently. It is separate from Watch: samples never
// notify watchers.
//
// Usage blocks until ctx is cancelled, which is the normal way to end it and
// terminates the docker process; it then returns nil. While Docker is
// unreachable, or the stats stream drops, it delivers a sample with
// DockerReachable false and retries every RetryEvery, resuming by itself.
// Containers are attributed to Projects and Services by name, against the
// current snapshot.
func (m *Manager) Usage(ctx context.Context, onSample func(UsageSample)) error {
	for ctx.Err() == nil {
		err := m.streamUsage(ctx, onSample)
		if ctx.Err() != nil {
			return nil
		}
		onSample(UsageSample{Error: err.Error(), Services: []ServiceUsage{}})
		if !sleep(ctx, m.retryEvery) {
			return nil
		}
	}
	return nil
}

// streamUsage reads the machine size and then the stats stream; it returns why
// the stream ended (always an error: a stats stream has no good end).
func (m *Manager) streamUsage(ctx context.Context, onSample func(UsageSample)) error {
	out, err := m.docker.Run(ctx, infoArgs...)
	if err != nil {
		return fmt.Errorf("compose: docker info: %w", err)
	}
	size, err := parseInfo(out)
	if err != nil {
		return err
	}

	frames := statsFrames{emit: func(rows []statsRow) { onSample(m.sampleOf(rows, size)) }}
	var lastStderr string
	err = m.docker.Stream(ctx, statsArgs, func(line string, stderr bool) {
		if stderr {
			lastStderr = strings.TrimSpace(line)
			return
		}
		frames.line(line)
	})
	if ctx.Err() == nil {
		frames.end()
	}
	switch {
	case err != nil && lastStderr != "":
		return fmt.Errorf("compose: docker stats: %w: %s", err, lastStderr)
	case err != nil:
		return fmt.Errorf("compose: docker stats: %w", err)
	}
	return errors.New("compose: docker stats ended")
}

type serviceKey struct{ project, service string }

// sampleOf sums the rows of a frame, globally and per Service.
func (m *Manager) sampleOf(rows []statsRow, size machine) UsageSample {
	owners := map[string]serviceKey{} // Container name -> Service
	for _, p := range m.Snapshot().Projects {
		for _, s := range p.Services {
			for _, c := range s.Containers {
				owners[c.Name] = serviceKey{p.Name, s.Name}
			}
		}
	}

	s := UsageSample{DockerReachable: true, Containers: len(rows), CPUs: size.cpus, MemoryTotalBytes: size.memBytes}
	perService := map[serviceKey]*Measures{}
	for _, r := range rows {
		s.Total.add(r.measures)
		if k, ok := owners[r.name]; ok {
			if perService[k] == nil {
				perService[k] = &Measures{}
			}
			perService[k].add(r.measures)
		}
	}
	s.CPUMachinePercent = s.Total.CPUPercent / float64(size.cpus)
	s.MemoryMachinePercent = float64(s.Total.MemoryBytes) / float64(size.memBytes) * 100

	s.Services = make([]ServiceUsage, 0, len(perService))
	for k, v := range perService {
		s.Services = append(s.Services, ServiceUsage{Project: k.project, Service: k.service, Measures: *v})
	}
	sort.Slice(s.Services, func(i, j int) bool {
		a, b := s.Services[i], s.Services[j]
		return a.Project < b.Project || (a.Project == b.Project && a.Service < b.Service)
	})
	return s
}
