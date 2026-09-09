package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

// ErrInvalidSuiteRun reports a suite run that cannot identify what it
// describes -- an empty suite or host, or a negative duration. The caller's
// validation is UX; this one is the boundary, ErrInvalidHazardShape's same
// split.
var ErrInvalidSuiteRun = errors.New("store: suite run needs a suite, a host and a non-negative duration")

// suiteRunColumns is the column list every read of a suite_runs row
// selects, in the order scanSuiteRunRow scans them.
const suiteRunColumns = `id, suite, host, duration_ms, exit_code, ran_at`

// suiteRunRowScanner is the one method scanSuiteRunRow needs, satisfied by
// both pgx.Row and pgx.Rows -- hazardRowScanner's reason, restated for the
// same shape of caller.
type suiteRunRowScanner interface {
	Scan(dest ...any) error
}

// InsertSuiteRun records one timed suite execution and returns the row as
// stored, timestamps included. RanAt left at the zero time takes the
// column's own now() default, the same disposition AddHazard's CreatedAt
// gets: a wrapper that omits it is stamped with when the run was actually
// recorded. The project_key foreign key is the boundary against filing
// runtimes under a project that does not exist.
func (s *Store) InsertSuiteRun(ctx context.Context, projectKey string, run records.SuiteRun) (records.SuiteRun, error) {
	if run.Suite == "" || run.Host == "" || run.DurationMs < 0 {
		return records.SuiteRun{}, fmt.Errorf("%w: got suite %q host %q duration %d", ErrInvalidSuiteRun, run.Suite, run.Host, run.DurationMs)
	}

	var ranAt *time.Time
	if !run.RanAt.IsZero() {
		ranAt = &run.RanAt
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO suite_runs (project_key, suite, host, duration_ms, exit_code, ran_at)
		VALUES ($1, $2, $3, $4, $5, COALESCE($6, now()))
		RETURNING `+suiteRunColumns+`
	`, projectKey, run.Suite, run.Host, run.DurationMs, run.ExitCode, ranAt)

	var out records.SuiteRun
	if err := row.Scan(&out.ID, &out.Suite, &out.Host, &out.DurationMs, &out.ExitCode, &out.RanAt); err != nil {
		return records.SuiteRun{}, fmt.Errorf("store: insert suite run for %s: %w", projectKey, err)
	}
	return out, nil
}

// ListSuiteRuns reads a project's suite runs, newest first. An empty suite
// is the caller's unfiltered listing and applies no suite filter at all;
// limit caps the row count and must be > 0 -- a caller asking for zero rows
// is a caller mistake, not an empty answer.
func (s *Store) ListSuiteRuns(ctx context.Context, projectKey, suite string, limit int) ([]records.SuiteRun, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("%w: limit must be > 0, got %d", ErrInvalidSuiteRun, limit)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT `+suiteRunColumns+`
		FROM suite_runs
		WHERE project_key = $1
		  AND ($2 = '' OR suite = $2)
		ORDER BY ran_at DESC, id DESC
		LIMIT $3
	`, projectKey, suite, limit)
	if err != nil {
		return nil, fmt.Errorf("store: list suite runs for %s: %w", projectKey, err)
	}
	defer rows.Close()

	var out []records.SuiteRun
	for rows.Next() {
		run, err := scanSuiteRunRow(rows)
		if err != nil {
			return nil, fmt.Errorf("store: list suite runs for %s: scan: %w", projectKey, err)
		}
		out = append(out, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list suite runs for %s: %w", projectKey, err)
	}
	return out, nil
}

func scanSuiteRunRow(row suiteRunRowScanner) (records.SuiteRun, error) {
	var run records.SuiteRun
	if err := row.Scan(&run.ID, &run.Suite, &run.Host, &run.DurationMs, &run.ExitCode, &run.RanAt); err != nil {
		return records.SuiteRun{}, err
	}
	return run, nil
}
