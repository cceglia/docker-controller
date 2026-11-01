package gui

import (
	"strings"

	"dockercontroller/internal/compose"
)

// AddProjectForm is the pending form of a new Project: the chosen File compose
// in -f order, the main one first, and the compose name. It holds only what the
// user typed, so the Manager stays the judge of what is saved.
type AddProjectForm struct {
	Files []string
	Name  string
	// edited is true once the user typed a name, so the name of a later main
	// file no longer replaces it.
	edited bool
}

// AddFile appends a File compose after the others and reports whether it added
// it. A blank path and a file already chosen are ignored, so the picker may
// hand the same file twice.
func (f *AddProjectForm) AddFile(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	for _, chosen := range f.Files {
		if chosen == path {
			return false
		}
	}
	f.Files = append(f.Files, path)
	return true
}

// RemoveFile drops the File compose at index. An index outside the list changes
// nothing, so a stale row never drops another file.
func (f *AddProjectForm) RemoveFile(index int) {
	if index < 0 || index >= len(f.Files) {
		return
	}
	f.Files = append(f.Files[:index], f.Files[index+1:]...)
}

// MainFile is the File compose compose derives the name from: the first one, ""
// when the user chose none.
func (f *AddProjectForm) MainFile() string {
	if len(f.Files) == 0 {
		return ""
	}
	return f.Files[0]
}

// Pick applies the files the picker returned and proposes the name of the main
// one. It reports whether the form gained a File compose, so the caller opens
// the dialog only after a real choice: a picker the user closed adds nothing.
// propose names a File compose, so a test needs no file on disk. A failure of
// propose keeps the added files and the name as it was.
func (f *AddProjectForm) Pick(paths []string, propose func(string) (string, error)) (bool, error) {
	before := f.MainFile()
	added := false
	for _, path := range paths {
		if f.AddFile(path) {
			added = true
		}
	}
	if !added {
		return false, nil
	}
	if f.MainFile() == before {
		return true, nil
	}
	name, err := propose(f.MainFile())
	if err != nil {
		return true, err
	}
	f.Propose(name)
	return true, nil
}

// Propose shows the name compose proposes. A name the user typed wins, so a
// later main file never replaces it.
func (f *AddProjectForm) Propose(name string) {
	if f.edited {
		return
	}
	f.Name = name
}

// Rename stores the name the user typed. From now on the main file no longer
// proposes a name, and a cleared name stays cleared.
func (f *AddProjectForm) Rename(name string) {
	f.Name = name
	f.edited = true
}

// NameEdited reports whether the user typed the name.
func (f *AddProjectForm) NameEdited() bool {
	return f.edited
}

// CleanName is the name the form sends: the trimmed text of the field.
func (f *AddProjectForm) CleanName() string {
	return strings.TrimSpace(f.Name)
}

// CanAdd reports whether the form may be sent: a File compose and a name are
// both needed. A name already saved is refused when the user sends the form,
// not here, so the field is never insensitive without a reason.
func (f *AddProjectForm) CanAdd() bool {
	return len(f.Files) > 0 && f.CleanName() != ""
}

// NameTaken reports whether a saved Project already has this compose name. The
// comparison is exact, because a name of another case is invalid and not a
// duplicate.
func NameTaken(name string, projects []compose.Project) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	for _, project := range projects {
		if project.Name == name {
			return true
		}
	}
	return false
}
