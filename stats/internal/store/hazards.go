package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/tweety53/agents/stats/internal/records"
)

// ErrHazardDuplicate reports an AddHazard for a (project_key, name) the
// project already holds -- the unique constraint refusing a re-add of a
// hazard the operator already recorded, rather than a silent duplicate row.
var ErrHazardDuplicate = errors.New("store: hazard already recorded")

// ErrHazardNotFound reports a RetireHazard naming no hazard the project
// holds -- ErrFindingNotFound's same shape of answer for the same kind of
// miss, and the sentinel the handler maps to 404.
var ErrHazardNotFound = errors.New("store: hazard not found")

// ErrInvalidHazardShape reports a ListHazards shape outside the closed
// vocabulary the hazards table's own CHECK enforces. The caller's
// validation is UX; this one is the boundary.
var ErrInvalidHazardShape = errors.New("store: hazard shape must be one of all, cross-repo, single-repo")

// hazardShapes is the closed applies vocabulary, shared by the store's
// boundary validation and mirrored by the migration's CHECK constraint.
var hazardShapes = map[string]bool{"all": true, "cross-repo": true, "single-repo": true}

// hazardColumns is the column list every read of a hazards row selects, in
// the order scanHazardRow scans them.
const hazardColumns = `id, name, body, applies, active, created_at`

// hazardRowScanner is the one method scanHazardRow needs, satisfied by both
// pgx.Row and pgx.Rows -- dispatchRowScanner's reason, restated for the
// same shape of caller.
type hazardRowScanner interface {
	Scan(dest ...any) error
}

// AddHazard inserts one hazard row, active: retiring is RetireHazard's job,
// never an insert's. CreatedAt left at the zero time takes the column's own
// now() default, the same disposition RecordIncident's OccurredAt gets: a
// hand-written hazard that omits it is stamped with when it was actually
// recorded. A duplicate (project_key, name) is ErrHazardDuplicate, never a
// bare unique-violation.
func (s *Store) AddHazard(ctx context.Context, projectKey string, h records.Hazard) (records.Hazard, error) {
	var createdAt *time.Time
	if !h.CreatedAt.IsZero() {
		createdAt = &h.CreatedAt
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO hazards (project_key, name, body, applies, active, created_at)
		VALUES ($1, $2, $3, $4, true, COALESCE($5, now()))
		RETURNING `+hazardColumns+`
	`, projectKey, h.Name, h.Body, h.Applies, createdAt)

	var out records.Hazard
	if err := row.Scan(&out.ID, &out.Name, &out.Body, &out.Applies, &out.Active, &out.CreatedAt); err != nil {
		if isUniqueViolation(err, "hazards_project_key_name_key") {
			return records.Hazard{}, fmt.Errorf("%w: %s/%s", ErrHazardDuplicate, projectKey, h.Name)
		}
		return records.Hazard{}, fmt.Errorf("store: add hazard for %s: %w", projectKey, err)
	}
	return out, nil
}

// ListHazards reads a project's hazards, newest first. shape filters to
// applies IN ('all', shape) and must be one of the closed-vocabulary values
// when non-empty; the empty shape is the operator's unfiltered listing and
// applies no applies filter at all. includeInactive lifts the active filter
// for the CLI's -all view.
func (s *Store) ListHazards(ctx context.Context, projectKey, shape string, includeInactive bool) ([]records.Hazard, error) {
	if shape != "" && !hazardShapes[shape] {
		return nil, ErrInvalidHazardShape
	}

	rows, err := s.pool.Query(ctx, `
		SELECT `+hazardColumns+`
		FROM hazards
		WHERE project_key = $1
		  AND ($2 = '' OR applies IN ('all', $2))
		  AND (active OR $3)
		ORDER BY created_at DESC, id DESC
	`, projectKey, shape, includeInactive)
	if err != nil {
		return nil, fmt.Errorf("store: list hazards for %s: %w", projectKey, err)
	}
	defer rows.Close()

	var out []records.Hazard
	for rows.Next() {
		h, err := scanHazardRow(rows)
		if err != nil {
			return nil, fmt.Errorf("store: list hazards for %s: scan: %w", projectKey, err)
		}
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list hazards for %s: %w", projectKey, err)
	}
	return out, nil
}

// RetireHazard sets active=false on the named row and returns it; an
// unknown name is ErrHazardNotFound for the handler to map to 404.
func (s *Store) RetireHazard(ctx context.Context, projectKey, name string) (records.Hazard, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE hazards SET active = false
		WHERE project_key = $1 AND name = $2
		RETURNING `+hazardColumns+`
	`, projectKey, name)

	var out records.Hazard
	if err := row.Scan(&out.ID, &out.Name, &out.Body, &out.Applies, &out.Active, &out.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return records.Hazard{}, fmt.Errorf("%w: %s/%s", ErrHazardNotFound, projectKey, name)
		}
		return records.Hazard{}, fmt.Errorf("store: retire hazard %s/%s: %w", projectKey, name, err)
	}
	return out, nil
}

func scanHazardRow(row hazardRowScanner) (records.Hazard, error) {
	var h records.Hazard
	if err := row.Scan(&h.ID, &h.Name, &h.Body, &h.Applies, &h.Active, &h.CreatedAt); err != nil {
		return records.Hazard{}, err
	}
	return h, nil
}
