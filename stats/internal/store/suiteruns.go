package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/tweety53/agents/stats/internal/records"
)

// ErrInvalidSuiteRun reports a suite run that cannot identify what it
// describes -- an empty suite or host, a negative duration, or a test row
// with no package, no test name or a negative cost -- and a read asking
// for no rows. The caller's validation is UX; this one is the boundary,
// ErrInvalidHazardShape's same split.
var ErrInvalidSuiteRun = errors.New("store: suite run needs a suite, a host, a non-negative duration and well-formed test rows")

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
//
// The run's Tests land in suite_run_tests in the same transaction as the
// suite row, every one of them whatever its cost (KAN-1124, decision
// store-every-test), so a malformed test row refuses the whole run before
// any SQL and a failed copy leaves no half-recorded run. The returned row
// carries no Tests; LatestSuiteRunTests reads them back.
func (s *Store) InsertSuiteRun(ctx context.Context, projectKey string, run records.SuiteRun) (records.SuiteRun, error) {
	if run.Suite == "" || run.Host == "" || run.DurationMs < 0 {
		return records.SuiteRun{}, fmt.Errorf("%w: got suite %q host %q duration %d", ErrInvalidSuiteRun, run.Suite, run.Host, run.DurationMs)
	}
	for _, test := range run.Tests {
		if test.Package == "" || test.Test == "" || test.CostMs < 0 {
			return records.SuiteRun{}, fmt.Errorf("%w: got test %q in package %q cost %d", ErrInvalidSuiteRun, test.Test, test.Package, test.CostMs)
		}
	}

	var ranAt *time.Time
	if !run.RanAt.IsZero() {
		ranAt = &run.RanAt
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return records.SuiteRun{}, fmt.Errorf("store: begin insert suite run for %s: %w", projectKey, err)
	}
	// Rollback after a successful Commit is a documented no-op in pgx.
	defer func() { _ = tx.Rollback(ctx) }()

	row := tx.QueryRow(ctx, `
		INSERT INTO suite_runs (project_key, suite, host, duration_ms, exit_code, ran_at)
		VALUES ($1, $2, $3, $4, $5, COALESCE($6, now()))
		RETURNING `+suiteRunColumns+`
	`, projectKey, run.Suite, run.Host, run.DurationMs, run.ExitCode, ranAt)

	var out records.SuiteRun
	if err := row.Scan(&out.ID, &out.Suite, &out.Host, &out.DurationMs, &out.ExitCode, &out.RanAt); err != nil {
		return records.SuiteRun{}, fmt.Errorf("store: insert suite run for %s: %w", projectKey, err)
	}

	if len(run.Tests) > 0 {
		rows := make([][]any, len(run.Tests))
		for i, test := range run.Tests {
			rows[i] = []any{out.ID, test.Package, test.Test, test.CostMs}
		}
		if _, err := tx.CopyFrom(ctx, pgx.Identifier{"suite_run_tests"},
			[]string{"suite_run_id", "package", "test", "cost_ms"}, pgx.CopyFromRows(rows)); err != nil {
			return records.SuiteRun{}, fmt.Errorf("store: insert suite run %d tests for %s: %w", out.ID, projectKey, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return records.SuiteRun{}, fmt.Errorf("store: commit suite run for %s: %w", projectKey, err)
	}
	return out, nil
}

// LatestSuiteRunTests reads the newest run of one suite that carries test
// rows, with its tests heaviest first and capped at limit (KAN-1124,
// decision suite-tests-read). A newer run with no test rows -- a line whose
// output carried no `go test -json` event -- is passed over, so the read
// never answers an empty list while an older run has costs. No such run is
// nil and no error; limit must be > 0, ListSuiteRuns' rule.
func (s *Store) LatestSuiteRunTests(ctx context.Context, projectKey, suite string, limit int) (*records.SuiteRun, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("%w: limit must be > 0, got %d", ErrInvalidSuiteRun, limit)
	}

	run, err := scanSuiteRunRow(s.pool.QueryRow(ctx, `
		SELECT `+suiteRunColumns+`
		FROM suite_runs
		WHERE project_key = $1
		  AND suite = $2
		  AND EXISTS (SELECT 1 FROM suite_run_tests WHERE suite_run_id = suite_runs.id)
		ORDER BY ran_at DESC, id DESC
		LIMIT 1
	`, projectKey, suite))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: latest %s run with tests for %s: %w", suite, projectKey, err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT package, test, cost_ms
		FROM suite_run_tests
		WHERE suite_run_id = $1
		ORDER BY cost_ms DESC, package, test
		LIMIT $2
	`, run.ID, limit)
	if err != nil {
		return nil, fmt.Errorf("store: suite run %d tests for %s: %w", run.ID, projectKey, err)
	}
	defer rows.Close()
	for rows.Next() {
		var test records.SuiteRunTest
		if err := rows.Scan(&test.Package, &test.Test, &test.CostMs); err != nil {
			return nil, fmt.Errorf("store: suite run %d tests for %s: scan: %w", run.ID, projectKey, err)
		}
		run.Tests = append(run.Tests, test)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: suite run %d tests for %s: %w", run.ID, projectKey, err)
	}
	return &run, nil
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
