package db

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"dockercontroller/internal/compose/model"
)

// MigrateGroupNames builds the saved Groups from the names written on the
// Projects before the 0003 drops the column. It reads projects.group_name,
// trims every name, unites the variants the model considers the same Group and
// keeps the spelling of the first Project in ascending compose name order,
// ignoring case. A blank name and a reserved label create no Group, and their
// Projects stay without one. The first run assigns the alphabetical order; a
// second run changes nothing. A database without the column needs nothing.
//
// It runs in one transaction and commits at the end.
func MigrateGroupNames(ctx context.Context, conn *sql.DB) error {
	has, err := columnExists(ctx, conn, "projects", "group_name")
	if err != nil {
		return err
	}
	if !has {
		return nil
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db: migrate Group names: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	projects, err := legacyGroups(ctx, tx)
	if err != nil {
		return err
	}
	sort.Slice(projects, func(i, j int) bool {
		left, right := strings.ToLower(projects[i].project), strings.ToLower(projects[j].project)
		if left != right {
			return left < right
		}
		return projects[i].project < projects[j].project
	})

	// The representative of each Group is the spelling of the first Project
	// that carries it. links holds, for every Project, the representative of
	// its Group.
	var names []string
	links := make(map[string]string, len(projects))
	for _, p := range projects {
		check := model.ValidateGroupName(p.group, nil, model.GroupNameCreate, "")
		if check.Status != model.GroupNameValid {
			continue
		}
		representative := ""
		for _, name := range names {
			if model.SameGroupName(name, check.Name) {
				representative = name
				break
			}
		}
		if representative == "" {
			representative = check.Name
			names = append(names, representative)
		}
		links[p.project] = representative
	}
	sort.Slice(names, func(i, j int) bool { return model.LessGroupName(names[i], names[j]) })

	saved, err := savedGroups(ctx, tx)
	if err != nil {
		return err
	}
	type resolvedGroup struct {
		name string
		id   int64
	}
	resolved := make([]resolvedGroup, 0, len(saved)+len(names))
	for _, g := range saved {
		resolved = append(resolved, resolvedGroup{name: g.name, id: g.id})
	}
	next := 0
	for _, g := range saved {
		if g.position >= next {
			next = g.position + 1
		}
	}
	find := func(name string) (int64, bool) {
		for _, g := range resolved {
			if model.SameGroupName(g.name, name) {
				return g.id, true
			}
		}
		return 0, false
	}
	for _, name := range names {
		if _, ok := find(name); ok {
			continue
		}
		res, err := tx.ExecContext(ctx, `INSERT INTO groups (name, position) VALUES (?, ?)`, name, next)
		if err != nil {
			return fmt.Errorf("db: migrate Group names: save %q: %w", name, err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("db: migrate Group names: save %q: %w", name, err)
		}
		resolved = append(resolved, resolvedGroup{name: name, id: id})
		next++
	}
	for project, name := range links {
		id, ok := find(name)
		if !ok {
			return fmt.Errorf("db: migrate Group names: Group %q of %s is missing", name, project)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE projects SET group_id = ? WHERE name = ?`, id, project); err != nil {
			return fmt.Errorf("db: migrate Group names: link %s: %w", project, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("db: migrate Group names: %w", err)
	}
	return nil
}

// legacyGroups is one Project with the Group name written on it before the
// migration.
type legacyGroup struct {
	project string
	group   string
}

// legacyGroups reads the Projects and their Group names.
func legacyGroups(ctx context.Context, tx *sql.Tx) ([]legacyGroup, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT name, COALESCE(group_name, '') FROM projects`)
	if err != nil {
		return nil, fmt.Errorf("db: migrate Group names: read Projects: %w", err)
	}
	defer rows.Close()
	var out []legacyGroup
	for rows.Next() {
		var p legacyGroup
		if err := rows.Scan(&p.project, &p.group); err != nil {
			return nil, fmt.Errorf("db: migrate Group names: read Projects: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db: migrate Group names: read Projects: %w", err)
	}
	return out, nil
}

// savedGroup is a row of groups already saved by an earlier run of the step.
type savedGroup struct {
	id       int64
	name     string
	position int
}

// savedGroups reads the Groups already saved, so a second run keeps them.
func savedGroups(ctx context.Context, tx *sql.Tx) ([]savedGroup, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id, name, position FROM groups ORDER BY position`)
	if err != nil {
		return nil, fmt.Errorf("db: migrate Group names: read Groups: %w", err)
	}
	defer rows.Close()
	var out []savedGroup
	for rows.Next() {
		var g savedGroup
		if err := rows.Scan(&g.id, &g.name, &g.position); err != nil {
			return nil, fmt.Errorf("db: migrate Group names: read Groups: %w", err)
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db: migrate Group names: read Groups: %w", err)
	}
	return out, nil
}
