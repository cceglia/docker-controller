package store

import (
	"context"
	"database/sql"
	"fmt"

	"dockercontroller/internal/compose/model"
)

// queryer is what a transaction and the database have in common: the helpers of
// the Groups read through both, so a lookup works inside a transaction of the
// caller and on its own.
type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// groupRow is one saved Group as the database holds it. The name comparison
// lives in model, which knows the Unicode folding; the store never repeats it.
type groupRow struct {
	id       int64
	name     string
	position int
}

// ListGroups returns the names of the saved Groups in the user's order, empty
// Groups included; never nil.
func (s *Store) ListGroups(ctx context.Context) ([]string, error) {
	rows, err := readGroups(ctx, s.db)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.name)
	}
	return out, nil
}

// CreateGroup saves a new Group after the ones already saved. A name that
// already exists, even with different capitalisation, is refused with
// model.DuplicateGroupError before anything is written, and a blank or reserved
// name with model.ErrInvalidGroupName.
func (s *Store) CreateGroup(ctx context.Context, name string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("compose: create Group %q: %w", name, err)
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := readGroups(ctx, tx)
	if err != nil {
		return err
	}
	check := model.ValidateGroupName(name, groupNames(rows), model.GroupNameCreate, "")
	switch check.Status {
	case model.GroupNameInvalid:
		return fmt.Errorf("%w: %q", model.ErrInvalidGroupName, name)
	case model.GroupNameDuplicate:
		return &model.DuplicateGroupError{Existing: check.Existing}
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO groups (name, position) VALUES (?, ?)`, check.Name, nextPosition(rows)); err != nil {
		return fmt.Errorf("compose: create Group %q: %w", name, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("compose: create Group %q: %w", name, err)
	}
	return nil
}

// RenameGroup renames a saved Group and keeps its place. A name that another
// Group already has, even with different capitalisation, unites the two: the
// Projects of the renamed Group move to the Group that already had the name,
// which keeps its place and its spelling. A blank or reserved name is refused
// with model.ErrInvalidGroupName.
func (s *Store) RenameGroup(ctx context.Context, from, to string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("compose: rename Group %q: %w", from, err)
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := readGroups(ctx, tx)
	if err != nil {
		return err
	}
	source, ok := groupNamed(rows, from)
	if !ok {
		return fmt.Errorf("%w: %q", model.ErrUnknownGroup, from)
	}
	check := model.ValidateGroupName(to, groupNames(rows), model.GroupNameRename, from)
	switch check.Status {
	case model.GroupNameInvalid:
		return fmt.Errorf("%w: %q", model.ErrInvalidGroupName, to)
	case model.GroupNameDuplicate:
		target, ok := groupNamed(rows, check.Existing)
		if !ok {
			return fmt.Errorf("%w: %q", model.ErrUnknownGroup, check.Existing)
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE projects SET group_id = ? WHERE group_id = ?`, target.id, source.id); err != nil {
			return fmt.Errorf("compose: unite Group %q into %q: %w", from, target.name, err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM groups WHERE id = ?`, source.id); err != nil {
			return fmt.Errorf("compose: unite Group %q into %q: %w", from, target.name, err)
		}
		if err := renumberFrom(ctx, tx, source.position); err != nil {
			return err
		}
	default:
		if _, err := tx.ExecContext(ctx,
			`UPDATE groups SET name = ? WHERE id = ?`, check.Name, source.id); err != nil {
			return fmt.Errorf("compose: rename Group %q: %w", from, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("compose: rename Group %q: %w", from, err)
	}
	return nil
}

// DeleteGroup removes a Group and leaves its Projects without one. Their rows
// stay as they are otherwise, and Docker is not touched.
func (s *Store) DeleteGroup(ctx context.Context, name string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("compose: delete Group %q: %w", name, err)
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := readGroups(ctx, tx)
	if err != nil {
		return err
	}
	group, ok := groupNamed(rows, name)
	if !ok {
		return fmt.Errorf("%w: %q", model.ErrUnknownGroup, name)
	}
	// The foreign key clears group_id on its own (ON DELETE SET NULL); the
	// explicit update keeps one fact in one place for a database opened without
	// the pragma.
	if _, err := tx.ExecContext(ctx, `UPDATE projects SET group_id = NULL WHERE group_id = ?`, group.id); err != nil {
		return fmt.Errorf("compose: delete Group %q: %w", name, err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM groups WHERE id = ?`, group.id); err != nil {
		return fmt.Errorf("compose: delete Group %q: %w", name, err)
	}
	if err := renumberFrom(ctx, tx, group.position); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("compose: delete Group %q: %w", name, err)
	}
	return nil
}

// MoveGroup moves a saved Group by delta places, swapping it with its neighbour.
// A move past either end changes nothing: the command is a no-op there, not an
// error.
func (s *Store) MoveGroup(ctx context.Context, name string, delta int) error {
	if delta == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("compose: move Group %q: %w", name, err)
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := readGroups(ctx, tx)
	if err != nil {
		return err
	}
	group, ok := groupNamed(rows, name)
	if !ok {
		return fmt.Errorf("%w: %q", model.ErrUnknownGroup, name)
	}
	neighbour, ok := groupAt(rows, group.position+delta)
	if !ok {
		return nil
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE groups SET position = ? WHERE id = ?`, neighbour.position, group.id); err != nil {
		return fmt.Errorf("compose: move Group %q: %w", name, err)
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE groups SET position = ? WHERE id = ?`, group.position, neighbour.id); err != nil {
		return fmt.Errorf("compose: move Group %q: %w", name, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("compose: move Group %q: %w", name, err)
	}
	return nil
}

// readGroups reads every Group in the user's order.
func readGroups(ctx context.Context, q queryer) ([]groupRow, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT id, name, position FROM groups ORDER BY position`)
	if err != nil {
		return nil, fmt.Errorf("compose: list Groups: %w", err)
	}
	defer rows.Close()

	var out []groupRow
	for rows.Next() {
		var g groupRow
		if err := rows.Scan(&g.id, &g.name, &g.position); err != nil {
			return nil, fmt.Errorf("compose: list Groups: %w", err)
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("compose: list Groups: %w", err)
	}
	return out, nil
}

// groupNamed finds the Group with exactly this name.
func groupNamed(rows []groupRow, name string) (groupRow, bool) {
	for _, row := range rows {
		if row.name == name {
			return row, true
		}
	}
	return groupRow{}, false
}

// groupNames lists the saved names, which the validation of the model reads.
func groupNames(rows []groupRow) []string {
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.name)
	}
	return out
}

// groupAt finds the Group at one position.
func groupAt(rows []groupRow, position int) (groupRow, bool) {
	for _, row := range rows {
		if row.position == position {
			return row, true
		}
	}
	return groupRow{}, false
}

// nextPosition is where a new Group goes: after every saved one.
func nextPosition(rows []groupRow) int {
	next := 0
	for _, row := range rows {
		if row.position >= next {
			next = row.position + 1
		}
	}
	return next
}

// renumberFrom closes the gap a removed Group leaves, so the positions stay
// 0, 1, 2, … and "Sposta su" always has a neighbour to swap with.
func renumberFrom(ctx context.Context, tx *sql.Tx, position int) error {
	if _, err := tx.ExecContext(ctx,
		`UPDATE groups SET position = position - 1 WHERE position > ?`, position); err != nil {
		return fmt.Errorf("compose: renumber Groups: %w", err)
	}
	return nil
}

// ensureGroup resolves a Group name to its id for a Project being saved, and
// creates the Group when the name is new. A name that differs only in
// capitalisation from a saved Group resolves to that Group, so one Group never
// becomes two and the Project keeps the saved spelling. A blank or reserved
// name is refused with model.ErrInvalidGroupName.
func ensureGroup(ctx context.Context, tx *sql.Tx, name string) (int64, error) {
	rows, err := readGroups(ctx, tx)
	if err != nil {
		return 0, err
	}
	check := model.ValidateGroupName(name, groupNames(rows), model.GroupNameAssign, "")
	switch check.Status {
	case model.GroupNameInvalid:
		return 0, fmt.Errorf("%w: %q", model.ErrInvalidGroupName, name)
	case model.GroupNameDuplicate:
		row, ok := groupNamed(rows, check.Existing)
		if !ok {
			return 0, fmt.Errorf("%w: %q", model.ErrUnknownGroup, check.Existing)
		}
		return row.id, nil
	}
	res, err := tx.ExecContext(ctx,
		`INSERT INTO groups (name, position) VALUES (?, ?)`, check.Name, nextPosition(rows))
	if err != nil {
		return 0, fmt.Errorf("compose: create Group %q: %w", name, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("compose: create Group %q: %w", name, err)
	}
	return id, nil
}
