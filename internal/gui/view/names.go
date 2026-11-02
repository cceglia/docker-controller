package view

import (
	"strings"

	"dockercontroller/internal/compose"
)

// MissingComposeText is the notice of a Project whose Compose files are gone.
const MissingComposeText = "File compose mancante: i file del Progetto non esistono più."

// ProjectName is the Display name of the Project, or its compose name when the
// Display name is unset.
func ProjectName(p compose.Project) string {
	return DisplayNameOf(p.DisplayName, p.Name)
}

// DisplayNameOf is the name the detail and the list show: the trimmed Display
// name, or the compose name when the Display name is blank.
func DisplayNameOf(displayName, composeName string) string {
	if name := strings.TrimSpace(displayName); name != "" {
		return name
	}
	return composeName
}
