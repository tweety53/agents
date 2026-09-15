package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/tweety53/agents/stats/internal/records"
)

// ErrInvalidSpecRun reports a spec run that cannot identify what it
// describes -- an empty spec file. The caller's validation is UX; this one
// is the boundary, ErrInvalidSuiteRun's same split.
var ErrInvalidSpecRun = errors.New("store: spec run needs a spec file")

// changesSinceSubquery is the unrun-for-N figure the inventory stat is
// built from: the project's changes updated after the row's last_ran_at.
// COALESCE to -infinity makes a never-recorded spec (NULL last_ran_at from
// the LEFT JOIN) count every change there is -- a spec with no recorded run
// has been unrun for the whole history. Every query below exposes the
// inventory row as the alias `s`.
const changesSinceSubquery = `(SELECT count(*) FROM changes c
		WHERE c.project_key = $1
		  AND c.updated_at > COALESCE(s.last_ran_at, '-infinity'::timestamptz))`

// specLastRunRowScanner is the one method scanSpecLastRunRow needs,
// satisfied by both pgx.Row and pgx.Rows -- suiteRunRowScanner's reason,
// restated for the same shape of caller.
type specLastRunRowScanner interface {
	Scan(dest ...any) error
}

// RecordSpecRun upserts one spec file's last-run timestamp and returns the
// inventory row as stored. RanAt left at the zero time takes the column's
// own now() default, InsertSuiteRun's same disposition: a caller that does
// not name a time is stamped with when the run was recorded. The upsert
// keeps the LATEST timestamp -- GREATEST, not EXCLUDED -- because the
// table holds an inventory, not a history: a retried capture or a
// backfilled timestamp arriving out of order must never drag the last-run
// figure backwards. The project_key foreign key is the boundary against
// filing runs under a project that does not exist.
func (s *Store) RecordSpecRun(ctx context.Context, projectKey, spec string, ranAt time.Time) (records.SpecLastRun, error) {
	if spec == "" {
		return records.SpecLastRun{}, fmt.Errorf("%w: got spec %q", ErrInvalidSpecRun, spec)
	}

	var ranAtArg *time.Time
	if !ranAt.IsZero() {
		ranAtArg = &ranAt
	}

	row := s.pool.QueryRow(ctx, `
		WITH upserted AS (
			INSERT INTO spec_lastruns (project_key, spec, last_ran_at)
			VALUES ($1, $2, COALESCE($3, now()))
			ON CONFLICT (project_key, spec) DO UPDATE
			SET last_ran_at = GREATEST(spec_lastruns.last_ran_at, EXCLUDED.last_ran_at)
			RETURNING spec, last_ran_at
		)
		SELECT s.spec, s.last_ran_at, `+changesSinceSubquery+`
		FROM upserted s
	`, projectKey, spec, ranAtArg)

	out, err := scanSpecLastRunRow(row)
	if err != nil {
		return records.SpecLastRun{}, fmt.Errorf("store: record spec run for %s: %w", projectKey, err)
	}
	return out, nil
}

// ListSpecLastRuns reads a project's per-spec inventory, ordered by spec.
// An empty specs slice lists every recorded spec alone -- a spec that never
// ran cannot appear, because the store has no way to know it exists. Named
// specs answer one row each, recorded or not: a name with no row is the
// never-run case, nil LastRanAt and the project's whole change count for
// ChangesSince, which is how the seven-suites-never-ran accident the
// inventory exists to prevent becomes a visible row.
func (s *Store) ListSpecLastRuns(ctx context.Context, projectKey string, specs []string) ([]records.SpecLastRun, error) {
	if len(specs) == 0 {
		rows, err := s.pool.Query(ctx, `
			SELECT s.spec, s.last_ran_at, `+changesSinceSubquery+`
			FROM spec_lastruns s
			WHERE s.project_key = $1
			ORDER BY s.spec
		`, projectKey)
		if err != nil {
			return nil, fmt.Errorf("store: list spec last runs for %s: %w", projectKey, err)
		}
		return collectSpecLastRunRows(rows, projectKey)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT wanted.spec, s.last_ran_at, `+changesSinceSubquery+`
		FROM unnest($2::text[]) AS wanted(spec)
		LEFT JOIN spec_lastruns s
		  ON s.project_key = $1 AND s.spec = wanted.spec
		ORDER BY wanted.spec
	`, projectKey, specs)
	if err != nil {
		return nil, fmt.Errorf("store: list spec last runs for %s: %w", projectKey, err)
	}
	return collectSpecLastRunRows(rows, projectKey)
}

func collectSpecLastRunRows(rows pgx.Rows, projectKey string) ([]records.SpecLastRun, error) {
	defer rows.Close()

	var out []records.SpecLastRun
	for rows.Next() {
		run, err := scanSpecLastRunRow(rows)
		if err != nil {
			return nil, fmt.Errorf("store: list spec last runs for %s: scan: %w", projectKey, err)
		}
		out = append(out, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list spec last runs for %s: %w", projectKey, err)
	}
	return out, nil
}

func scanSpecLastRunRow(row specLastRunRowScanner) (records.SpecLastRun, error) {
	var out records.SpecLastRun
	if err := row.Scan(&out.Spec, &out.LastRanAt, &out.ChangesSince); err != nil {
		return records.SpecLastRun{}, err
	}
	return out, nil
}
