package compose

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// discovered is a Project as reported by `docker compose ls -a`.
type discovered struct {
	Name         string
	ComposeFiles []string
}

func (d discovered) folder() string {
	if len(d.ComposeFiles) == 0 {
		return ""
	}
	return filepath.Dir(d.ComposeFiles[0])
}

// parseLs parses the output of `docker compose ls -a --format json`.
func parseLs(out []byte) ([]discovered, error) {
	var raw []struct {
		Name        string `json:"Name"`
		ConfigFiles string `json:"ConfigFiles"`
	}
	if len(bytes.TrimSpace(out)) == 0 {
		return nil, nil
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("compose: parse `compose ls`: %w", err)
	}
	res := make([]discovered, 0, len(raw))
	for _, r := range raw {
		d := discovered{Name: r.Name}
		for _, f := range strings.Split(r.ConfigFiles, ",") {
			if f = strings.TrimSpace(f); f != "" {
				d.ComposeFiles = append(d.ComposeFiles, f)
			}
		}
		res = append(res, d)
	}
	return res, nil
}

// parsePs parses the output of `docker compose -p <name> ps -a --format json`
// into the Project's Containers. Compose emits one JSON object per line; a JSON
// array is accepted too.
func parsePs(out []byte) ([]psContainer, error) {
	trimmed := bytes.TrimSpace(out)
	if len(trimmed) == 0 {
		return nil, nil
	}
	var cs []psContainer
	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &cs); err != nil {
			return nil, fmt.Errorf("compose: parse `compose ps`: %w", err)
		}
		return cs, nil
	}
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	for dec.More() {
		var c psContainer
		if err := dec.Decode(&c); err != nil {
			return nil, fmt.Errorf("compose: parse `compose ps`: %w", err)
		}
		cs = append(cs, c)
	}
	return cs, nil
}

// psContainer is the part of a `compose ps` entry the app uses.
type psContainer struct {
	Name       string `json:"Name"`
	Service    string `json:"Service"`
	Image      string `json:"Image"`
	State      string `json:"State"`
	Health     string `json:"Health"`
	Publishers []struct {
		TargetPort    int    `json:"TargetPort"`
		PublishedPort int    `json:"PublishedPort"`
		Protocol      string `json:"Protocol"`
	} `json:"Publishers"`
}

// groupByService groups Containers by Service, both sorted by name.
func groupByService(cs []psContainer) []Service {
	byName := map[string][]Container{}
	for _, c := range cs {
		byName[c.Service] = append(byName[c.Service], Container{
			Name: c.Name, Image: c.Image, State: c.State, Health: c.Health, Ports: c.ports(),
		})
	}
	services := make([]Service, 0, len(byName))
	for name, containers := range byName {
		sort.Slice(containers, func(i, j int) bool { return containers[i].Name < containers[j].Name })
		services = append(services, Service{Name: name, Containers: containers})
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })
	return services
}

// ports lists the published ports once each (a port bound on 0.0.0.0 and :: is
// reported twice by Docker), sorted by published port.
func (c psContainer) ports() []Port {
	ports := []Port{}
	seen := map[Port]bool{}
	for _, p := range c.Publishers {
		if p.PublishedPort == 0 {
			continue // exposed but not published
		}
		port := Port{Published: p.PublishedPort, Target: p.TargetPort, Protocol: p.Protocol}
		if !seen[port] {
			seen[port] = true
			ports = append(ports, port)
		}
	}
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].Published != ports[j].Published {
			return ports[i].Published < ports[j].Published
		}
		return ports[i].Protocol < ports[j].Protocol
	})
	return ports
}

// runningCount is how many of the Containers run.
func runningCount(cs []psContainer) int {
	n := 0
	for _, c := range cs {
		if c.State == "running" {
			n++
		}
	}
	return n
}

func statusOf(running, total int) Status {
	switch {
	case total == 0:
		return StatusRemoved
	case running == total:
		return StatusActive
	case running == 0:
		return StatusStopped
	default:
		return StatusPartial
	}
}
