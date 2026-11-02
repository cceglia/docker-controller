package model

import "path/filepath"

// StatusOf returns the Status of a Project from its Container counts.
func StatusOf(running, total int) Status {
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

// Folder returns the directory of the first compose file, or "" when there is
// no file.
func Folder(composeFiles []string) string {
	if len(composeFiles) == 0 {
		return ""
	}
	return filepath.Dir(composeFiles[0])
}
