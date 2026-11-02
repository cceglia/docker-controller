package compose

import (
	"context"
	"fmt"
	"os"
	"strings"

	"dockercontroller/internal/compose/parse"
)

// profileCache is what `config --profiles` said for one Project, valid while
// the compose files it was read from are unchanged.
type profileCache struct {
	signature string
	names     []string
}

// availableProfiles returns the Profiles the compose files of p define. They
// are read with `compose config --profiles` only when the compose files are
// new or changed (path, size, modification time), never per Docker event or
// per 30 second refresh. A failure counts as "no Profiles" until the files
// change. It must be called with refreshMu
// held.
func (m *Manager) availableProfiles(ctx context.Context, p Project) []string {
	prev, known := m.profiles[p.Name]
	if p.ComposeMissing || len(p.ComposeFiles) == 0 {
		return orEmpty(prev.names)
	}
	sig := filesSignature(p)
	if known && prev.signature == sig {
		return prev.names
	}
	out, err := m.docker.Run(ctx, append(projectArgs(p), "config", "--profiles")...)
	if ctx.Err() != nil {
		return orEmpty(prev.names) // interrupted: say nothing new, try again later
	}
	names := parse.Profiles(out)
	if err != nil {
		// compose cannot read the files (invalid YAML, say): that does not
		// change until they do, so remember it instead of failing again at
		// every refresh.
		names = []string{}
	}
	c := profileCache{signature: sig, names: names}
	m.profiles[p.Name] = c
	return c.names
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// filesSignature identifies the state of the compose files of p.
func filesSignature(p Project) string {
	var b strings.Builder
	b.WriteString(p.Folder)
	for _, f := range p.ComposeFiles {
		if info, err := os.Stat(f); err == nil {
			fmt.Fprintf(&b, "|%s:%d:%d", f, info.Size(), info.ModTime().UnixNano())
		} else {
			fmt.Fprintf(&b, "|%s:?", f)
		}
	}
	return b.String()
}
