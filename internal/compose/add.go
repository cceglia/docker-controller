package compose

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// ErrDuplicateProject is returned by AddProject when a Project with that
	// name is already saved.
	ErrDuplicateProject = errors.New("compose: a Project with this name already exists")
	// ErrInvalidProject is returned for a Project that compose could not use: a
	// name it would reject, no compose file, or a file that cannot be read.
	ErrInvalidProject = errors.New("compose: invalid Project")
	// The specific reasons of an ErrInvalidProject; each error wraps both.
	ErrInvalidName         = errors.New("name is not a valid compose name")
	ErrNoComposeFile       = errors.New("no compose file")
	ErrComposeFileUnusable = errors.New("compose file is not an absolute path to a readable file")
)

// validName is what compose accepts as a project name.
var validName = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// unsafeNameChars are the characters compose drops when it normalises a name.
var unsafeNameChars = regexp.MustCompile(`[^a-z0-9_-]`)

// normalizeName is compose's normalisation of a project name: lower case, only
// letters, digits, "_" and "-", and no leading "_" or "-".
func normalizeName(s string) string {
	s = unsafeNameChars.ReplaceAllString(strings.ToLower(s), "")
	return strings.TrimLeft(s, "_-")
}

// ProposeName returns the compose project name for a main compose file, as
// compose itself would choose it: the top-level `name:` of the file, otherwise
// the name of its folder, normalised. It only reads the file.
func ProposeName(mainFile string) (string, error) {
	f, err := os.Open(mainFile)
	if err != nil {
		return "", fmt.Errorf("%w: %w: %w", ErrInvalidProject, ErrComposeFileUnusable, err)
	}
	defer f.Close()

	name := ""
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		// Only a key at column 0 is the project name; a `name:` nested under a
		// service is not.
		rest, ok := strings.CutPrefix(sc.Text(), "name:")
		if !ok {
			continue
		}
		if i := strings.Index(rest, " #"); i >= 0 {
			rest = rest[:i]
		}
		name = strings.Trim(strings.TrimSpace(rest), `"'`)
		break
	}
	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("%w: read %s: %w", ErrInvalidProject, mainFile, err)
	}

	if name == "" {
		abs, err := filepath.Abs(mainFile)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrInvalidProject, err)
		}
		name = filepath.Base(filepath.Dir(abs))
	}
	if name = normalizeName(name); name == "" {
		return "", fmt.Errorf("%w: no usable name for %s", ErrInvalidProject, mainFile)
	}
	return name, nil
}

// AddProject saves a Project added by hand: name is its compose name and
// files its compose files in -f order (the main file first). Its folder is the
// directory of the main file. It never touches Docker beyond re-reading its
// state, so a Project never started appears as Rimosso. It is refused with
// ErrInvalidProject (bad name, no files, relative path, a file that is not a
// readable regular file) or ErrDuplicateProject.
func (m *Manager) AddProject(ctx context.Context, name string, files []string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("%w: %w: %q (lower case letters, digits, - and _)", ErrInvalidProject, ErrInvalidName, name)
	}
	if len(files) == 0 {
		return fmt.Errorf("%w: %w", ErrInvalidProject, ErrNoComposeFile)
	}
	for _, f := range files {
		if !filepath.IsAbs(f) {
			return fmt.Errorf("%w: %w: %q is not an absolute path", ErrInvalidProject, ErrComposeFileUnusable, f)
		}
		if info, err := os.Stat(f); err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("%w: %w: %q is not a readable file", ErrInvalidProject, ErrComposeFileUnusable, f)
		}
	}

	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()

	if err := m.store.add(ctx, discovered{Name: name, ComposeFiles: files}); err != nil {
		return err
	}
	next, err := m.build(ctx)
	if err != nil {
		return err
	}
	m.publish(next)
	return nil
}
