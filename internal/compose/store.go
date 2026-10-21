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
