package compose

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
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

// containerCounts returns how many Containers a Project has and how many run,
// from the output of `docker compose -p <name> ps -a --format json`. Compose
// emits one JSON object per line; a JSON array is accepted too.
func containerCounts(out []byte) (running, total int, err error) {
	trimmed := bytes.TrimSpace(out)
	if len(trimmed) == 0 {
		return 0, 0, nil
	}
	type container struct {
		State string `json:"State"`
	}
	var cs []container
	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &cs); err != nil {
			return 0, 0, fmt.Errorf("compose: parse `compose ps`: %w", err)
		}
	} else {
		dec := json.NewDecoder(bytes.NewReader(trimmed))
		for dec.More() {
			var c container
			if err := dec.Decode(&c); err != nil {
				return 0, 0, fmt.Errorf("compose: parse `compose ps`: %w", err)
			}
			cs = append(cs, c)
		}
	}
	for _, c := range cs {
		if c.State == "running" {
			running++
		}
	}
	return running, len(cs), nil
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
