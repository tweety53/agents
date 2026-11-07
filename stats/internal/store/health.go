package store

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/tweety53/agents/stats/internal/records"
)

// health.go backs the flow-health dashboard's guard-health view: which
// guards catch something, which never fire, and what each costs in run time
// (GuardHealth, one row per guard; GuardHealthProjects, the per-project run
// counts its Project chips show) -- plus the one write both read
// (RecordGuardRun).

// ErrProjectNotFound is returned by a project-scoped write naming a project
// the store holds no row for. RecordGuardRun is its caller: a guard run
// inside a test fixture's throwaway repository names such a project, and
// the flow-guard binary drops the refusal rather than recording it.
var ErrProjectNotFound = errors.New("store: project not found")

// ErrInvalidGuardRun is returned by RecordGuardRun for a run outside what
// 0032_guard_runs.sql accepts -- an empty guard or worktree, an outcome
// outside records.GuardRunOutcomes, or a negative duration -- checked here
// so the caller gets a 400 naming the field rather than a constraint
// violation.
var ErrInvalidGuardRun = errors.New("store: invalid guard run")

// RecordGuardRun records one flow-guard invocation against projectKey.
// Every call inserts a new row, as RecordVerdict does. An unknown project is
// ErrProjectNotFound: the INSERT ... SELECT inserts nothing, rather than
// tripping the foreign key, so the refusal is a sentinel and not a 500.
func (s *Store) RecordGuardRun(ctx context.Context, projectKey string, in records.GuardRun) (records.GuardRun, error) {
	if in.Guard == "" || in.Worktree == "" {
		return records.GuardRun{}, fmt.Errorf("%w: guard and worktree are both required", ErrInvalidGuardRun)
	}
	if !slices.Contains(records.GuardRunOutcomes, in.Outcome) {
		return records.GuardRun{}, fmt.Errorf("%w: outcome %q is not one of %v", ErrInvalidGuardRun, in.Outcome, records.GuardRunOutcomes)
	}
	if in.DurationMS < 0 {
		return records.GuardRun{}, fmt.Errorf("%w: durationMs must not be negative", ErrInvalidGuardRun)
	}
	recordedAt := in.RecordedAt
	if recordedAt.IsZero() {
		recordedAt = time.Now()
	}

	out := in
	err := s.pool.QueryRow(ctx, `
		INSERT INTO guard_runs (project_key, guard, worktree, exit_code, outcome, duration_ms, recorded_at)
		SELECT p.project_key, $2, $3, $4, $5, $6, $7
		FROM projects p
		WHERE p.project_key = $1
		RETURNING id, recorded_at
	`, projectKey, in.Guard, in.Worktree, in.ExitCode, in.Outcome, in.DurationMS, recordedAt).Scan(&out.ID, &out.RecordedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return records.GuardRun{}, fmt.Errorf("%w: %s", ErrProjectNotFound, projectKey)
		}
		return records.GuardRun{}, fmt.Errorf("store: record guard run for %s: %w", projectKey, err)
	}
	return out, nil
}

// GuardHealthRow is one guard's record in the guard-health view: how often
// it ran in the window, how often it fired or could not answer, its median
// and summed run time, and when it last fired -- nil when it never fired in
// the window. MedianMs is percentile_cont(0.5) rounded to whole milliseconds.
type GuardHealthRow struct {
	Guard        string
	Runs         int
	Fired        int
	CannotAnswer int
	MedianMs     int64
	TotalMs      int64
	LastFiredAt  *time.Time
}

// GuardHealth reports one row per guard that ran in period, optionally
// restricted to one project, ordered by guard. records.GuardHelpers are
// left out at read time: they are helpers, not guards.
func (s *Store) GuardHealth(ctx context.Context, period Period, project *string) ([]GuardHealthRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT guard,
			COUNT(*)::int,
			(COUNT(*) FILTER (WHERE outcome = 'fired'))::int,
			(COUNT(*) FILTER (WHERE outcome = 'cannot-answer'))::int,
			ROUND(percentile_cont(0.5) WITHIN GROUP (ORDER BY duration_ms))::bigint,
			SUM(duration_ms)::bigint,
			MAX(recorded_at) FILTER (WHERE outcome = 'fired')
		FROM guard_runs
		WHERE recorded_at >= $1 AND recorded_at < $2
		  AND ($3::text IS NULL OR project_key = $3)
		  AND guard <> ALL($4::text[])
		GROUP BY guard
		ORDER BY guard
	`, period.From, period.To, project, records.GuardHelpers)
	if err != nil {
		return nil, fmt.Errorf("store: guard health: %w", err)
	}
	defer rows.Close()

	var out []GuardHealthRow
	for rows.Next() {
		var row GuardHealthRow
		if err := rows.Scan(&row.Guard, &row.Runs, &row.Fired, &row.CannotAnswer,
			&row.MedianMs, &row.TotalMs, &row.LastFiredAt); err != nil {
			return nil, fmt.Errorf("store: guard health: scan: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: guard health: %w", err)
	}
	return out, nil
}

// ProjectGuardRuns is one project's guard-run count in a window -- a
// Project chip of the flow-health view.
type ProjectGuardRuns struct {
	ProjectKey string
	Runs       int
}

// GuardHealthProjects counts each project's guard runs in period, helpers
// excluded as GuardHealth excludes them, count descending then key.
func (s *Store) GuardHealthProjects(ctx context.Context, period Period) ([]ProjectGuardRuns, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT project_key, COUNT(*)::int
		FROM guard_runs
		WHERE recorded_at >= $1 AND recorded_at < $2
		  AND guard <> ALL($3::text[])
		GROUP BY project_key
		ORDER BY 2 DESC, 1
	`, period.From, period.To, records.GuardHelpers)
	if err != nil {
		return nil, fmt.Errorf("store: guard health projects: %w", err)
	}
	defer rows.Close()

	var out []ProjectGuardRuns
	for rows.Next() {
		var row ProjectGuardRuns
		if err := rows.Scan(&row.ProjectKey, &row.Runs); err != nil {
			return nil, fmt.Errorf("store: guard health projects: scan: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: guard health projects: %w", err)
	}
	return out, nil
}
