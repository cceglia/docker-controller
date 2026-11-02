package project

import (
	"reflect"
	"slices"

	"dockercontroller/internal/compose"
)

// DetailData is what the panel of the detail draws of a Project and of Docker.
// A snapshot whose data equals the one on screen draws the same panel, so the
// panel keeps its widgets, the value of the scroll bar and the focus.
type DetailData struct {
	Project         compose.Project
	DockerReachable bool
	Loading         bool
}

// SameDraw reports whether two sets of data draw the same panel. It compares
// the fields the panel draws, not the pointer of the snapshot: a change of the
// Profiles the compose files define reaches the list and the form of the
// settings, not the detail, so it alone does not rebuild it.
func (d DetailData) SameDraw(other DetailData) bool {
	if d.DockerReachable != other.DockerReachable || d.Loading != other.Loading {
		return false
	}

	a, b := d.Project, other.Project
	if a.Name != b.Name || a.DisplayName != b.DisplayName || a.Group != b.Group ||
		a.Folder != b.Folder ||
		a.Status != b.Status || a.Running != b.Running || a.Total != b.Total ||
		a.ComposeMissing != b.ComposeMissing {
		return false
	}
	if !slices.Equal(a.ComposeFiles, b.ComposeFiles) ||
		!slices.Equal(a.ActiveProfiles, b.ActiveProfiles) {
		return false
	}
	// The panel writes the reason of an insensitive Action and the Actions
	// themselves from the Activity: its identity is enough.
	if (a.Activity == nil) != (b.Activity == nil) {
		return false
	}
	// The Services carry the Containers, their state, their health and their
	// ports: a nested comparison cannot say more than this one.
	return reflect.DeepEqual(a.Services, b.Services)
}
