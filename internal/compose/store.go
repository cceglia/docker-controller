package compose

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// store persists the saved Projects in the app database.
type store struct{ db *sql.DB }

// upsertDiscovered saves (or refreshes folder and files of) a discovered
// Project, leaving the user's display name, group and profiles untouched.
func (s store) upsertDiscovered(ctx context.Context, d discovered) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("compose: save %s: %w", d.Name, err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO projects (name, folder) VALUES (?, ?)
		 ON CONFLICT(name) DO UPDATE SET folder = excluded.folder`,
		d.Name, d.folder()); err != nil {
		return fmt.Errorf("compose: save %s: %w", d.Name, err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM project_files WHERE project_name = ?`, d.Name); err != nil {
		return fmt.Errorf("compose: save %s files: %w", d.Name, err)
	}
	for i, f := range d.ComposeFiles {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO project_files (project_name, position, path) VALUES (?, ?, ?)`,
			d.Name, i, f); err != nil {
			return fmt.Errorf("compose: save %s files: %w", d.Name, err)
		}
	}
	return tx.Commit()
}

// add saves a Project added by hand; a name already saved is refused with
// ErrDuplicateProject and nothing is changed.
func (s store) add(ctx context.Context, d discovered) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("compose: add %s: %w", d.Name, err)
	}
	defer func() { _ = tx.Rollback() }()

	var taken bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM projects WHERE name = ?)`, d.Name).Scan(&taken); err != nil {
		return fmt.Errorf("compose: add %s: %w", d.Name, err)
	}
	if taken {
		return fmt.Errorf("%w: %q", ErrDuplicateProject, d.Name)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO projects (name, folder) VALUES (?, ?)`, d.Name, d.folder()); err != nil {
		return fmt.Errorf("compose: add %s: %w", d.Name, err)
	}
	for i, f := range d.ComposeFiles {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO project_files (project_name, position, path) VALUES (?, ?, ?)`,
			d.Name, i, f); err != nil {
			return fmt.Errorf("compose: add %s files: %w", d.Name, err)
		}
	}
	return tx.Commit()
}

// updateSettings applies the non-nil fields of st to a saved Project; an empty
// value is stored as NULL. It reports whether the Project exists.
func (s store) updateSettings(ctx context.Context, name string, st Settings) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("compose: update settings of %s: %w", name, err)
	}
	defer func() { _ = tx.Rollback() }()

	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM projects WHERE name = ?)`, name).Scan(&exists); err != nil {
		return false, fmt.Errorf("compose: update settings of %s: %w", name, err)
	}
	if !exists {
		return false, nil
	}
	if st.DisplayName != nil {
		if _, err := tx.ExecContext(ctx, `UPDATE projects SET display_name = NULLIF(?, '') WHERE name = ?`, *st.DisplayName, name); err != nil {
			return false, fmt.Errorf("compose: update display name of %s: %w", name, err)
		}
	}
	if st.Group != nil {
		if _, err := tx.ExecContext(ctx, `UPDATE projects SET group_name = NULLIF(?, '') WHERE name = ?`, *st.Group, name); err != nil {
			return false, fmt.Errorf("compose: update group of %s: %w", name, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("compose: update settings of %s: %w", name, err)
	}
	return true, nil
}

// list returns every saved Project, sorted by name, without status.
func (s store) list(ctx context.Context) ([]Project, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT name, folder, COALESCE(display_name, ''), COALESCE(group_name, '') FROM projects`)
	if err != nil {
		return nil, fmt.Errorf("compose: list projects: %w", err)
	}
	defer rows.Close()

	byName := map[string]*Project{}
	var out []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.Name, &p.Folder, &p.DisplayName, &p.Group); err != nil {
			return nil, fmt.Errorf("compose: list projects: %w", err)
		}
		p.ComposeFiles = []string{}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("compose: list projects: %w", err)
	}
	for i := range out {
		byName[out[i].Name] = &out[i]
	}

	frows, err := s.db.QueryContext(ctx,
		`SELECT project_name, path FROM project_files ORDER BY project_name, position`)
	if err != nil {
		return nil, fmt.Errorf("compose: list project files: %w", err)
	}
	defer frows.Close()
	for frows.Next() {
		var name, path string
		if err := frows.Scan(&name, &path); err != nil {
			return nil, fmt.Errorf("compose: list project files: %w", err)
		}
		if p := byName[name]; p != nil {
			p.ComposeFiles = append(p.ComposeFiles, path)
		}
	}
	if err := frows.Err(); err != nil {
		return nil, fmt.Errorf("compose: list project files: %w", err)
	}

	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

// remove deletes a saved Project; its compose files and active Profiles go
// with it (ON DELETE CASCADE).
func (s store) remove(ctx context.Context, name string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM projects WHERE name = ?`, name); err != nil {
		return fmt.Errorf("compose: remove %s: %w", name, err)
	}
	return nil
}
