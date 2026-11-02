package project

import (
	"fmt"
	"path/filepath"
	"strings"

	"dockercontroller/internal/compose"
)

// ToneView is how a Container reads in the detail: the glyph of its marker and
// the Adwaita class that tints it. It is the dot the old view drew per row.
type ToneView struct {
	Glyph    string
	CSSClass string
}

// ToneOfContainer reads the tone of one Container: a fault wins, then a state
// that is not settled, then a running Container, and everything else is quiet.
// The glyphs are the ones of the list marker.
func ToneOfContainer(c compose.Container) ToneView {
	switch {
	case c.Health == "unhealthy" || c.State == "dead":
		return ToneView{Glyph: "●", CSSClass: "error"}
	case c.Health == "starting" || c.State == "restarting" || c.State == "paused":
		return ToneView{Glyph: "◐", CSSClass: "warning"}
	case c.State == "running":
		return ToneView{Glyph: "●", CSSClass: "success"}
	}
	return ToneView{Glyph: "○", CSSClass: "dim-label"}
}

// HealthView is how the health of a Container reads: the text Docker reports
// and the Adwaita class that tints it. An empty CSSClass means no problem, so
// the text stays quiet.
type HealthView struct {
	Text     string
	CSSClass string
}

// HealthOf marks only a health that is a problem. A Container with no
// healthcheck has no text, so the row shows nothing.
func HealthOf(health string) HealthView {
	switch health {
	case "unhealthy":
		return HealthView{Text: health, CSSClass: "error"}
	case "starting":
		return HealthView{Text: health, CSSClass: "warning"}
	}
	return HealthView{Text: health}
}

// PortText is one published port as the Container row shows it: the host port,
// the Container port and the protocol ("8080→80/tcp").
func PortText(p compose.Port) string {
	return fmt.Sprintf("%d→%d/%s", p.Published, p.Target, p.Protocol)
}

// ComposeFilesLine names the Compose files of the Project by their base name, in
// -f order. A Project without a file has no line: it returns "".
func ComposeFilesLine(files []string) string {
	if len(files) == 0 {
		return ""
	}
	names := make([]string, 0, len(files))
	for _, f := range files {
		names = append(names, filepath.Base(f))
	}
	return "File compose: " + strings.Join(names, ", ")
}

// ActiveProfilesLine names the Profiles the user chose for the Project. Without
// an active Profile it returns "", so the line stays hidden.
func ActiveProfilesLine(profiles []string) string {
	if len(profiles) == 0 {
		return ""
	}
	return "Profili attivi: " + strings.Join(profiles, ", ")
}

// ContainerCountLine is the count of the Containers of a Project. A Project
// without Container says so instead of counting "0/0".
func ContainerCountLine(p compose.Project) string {
	return containersLine(p.Running, p.Total)
}

// ServiceStateLine is the count of the Containers of a Service, as its block
// shows it. A Service without Container says so.
func ServiceStateLine(service compose.Service) string {
	running := 0
	for _, container := range service.Containers {
		if container.State == "running" {
			running++
		}
	}
	return containersLine(running, len(service.Containers))
}

// containersLine spells out the running Containers over the total. A total of
// zero is no Container, not "0/0".
func containersLine(running, total int) string {
	if total == 0 {
		return "Nessun Container"
	}
	return fmt.Sprintf("%d/%d in esecuzione", running, total)
}

// GroupLine names the Group of the Project. A Project without a Group says so,
// so the block of the data always carries the line.
func GroupLine(group string) string {
	if name := strings.TrimSpace(group); name != "" {
		return "Gruppo: " + name
	}
	return "Gruppo: nessuno"
}

// NoServicesText explains an empty list of Services. A missing Compose file is
// the cause when it applies; otherwise the Project is Rimosso and Up creates
// the Containers. It is not an error.
func NoServicesText(p compose.Project) string {
	if p.ComposeMissing {
		return "Nessun Servizio: i file compose del Progetto non esistono più."
	}
	return "Nessun Servizio: il Progetto è Rimosso. Avvialo con Up per creare i Container."
}
