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

// health.go backs the flow-health dashboard: three views asking whether
// flow's own machinery earns its keep -- which guards ever fire, which
// stages get re-entered, and how many review-panel rounds a change needs --
// plus the one write the first of those needs (RecordGuardRun). Every other
// view in aggregate.go asks what the pipeline costs; these ask where it
// loops, so a rule or guard that never catches anything can be found and
// removed, and a stage that often needs a second pass can be tightened.

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

// GuardActivityRow is one guard's record across a period: its runs as the
// flow-guard binary recorded them (guard_runs) beside the verdicts and
// operator-flagged false positives the guards that record one kept
// (guard_verdicts). A guard appears when either source names it.
//
// LastRunAt, LastFiredAt and MedianDurationMS are nil when the guard has no
// guard_runs row in the period (it is known only from a verdict), or, for
// LastFiredAt, when it never fired -- absence, never a zero time.
type GuardActivityRow struct {
	Guard            string
	Runs             int
	Fired            int
	CannotAnswer     int
	LastRunAt        *time.Time
	LastFiredAt      *time.Time
	MedianDurationMS *float64
	Verdicts         int
	FalsePositives   int
}

// GuardActivity reports one row per guard across period, optionally
// restricted to one project. guard_runs rows are scoped by their own
// recorded_at, and guard_verdicts rows by theirs.
func (s *Store) GuardActivity(ctx context.Context, period Period, project *string) ([]GuardActivityRow, error) {
	rows, err := s.pool.Query(ctx, `
		WITH runs AS (
			SELECT gr.guard,
				COUNT(*) AS runs,
				COUNT(*) FILTER (WHERE gr.outcome = 'fired') AS fired,
				COUNT(*) FILTER (WHERE gr.outcome = 'cannot-answer') AS cannot_answer,
				MAX(gr.recorded_at) AS last_run,
				MAX(gr.recorded_at) FILTER (WHERE gr.outcome = 'fired') AS last_fired,
				percentile_cont(0.5) WITHIN GROUP (ORDER BY gr.duration_ms) AS median_ms
			FROM guard_runs gr
			WHERE gr.recorded_at >= $1 AND gr.recorded_at < $2
			  AND ($3::text IS NULL OR gr.project_key = $3)
			GROUP BY gr.guard
		),
		verdicts AS (
			SELECT gv.guard,
				COUNT(*) AS verdicts,
				COUNT(*) FILTER (WHERE gv.false_positive) AS false_positives
			FROM guard_verdicts gv
			JOIN changes c ON c.id = gv.change_id
			WHERE gv.recorded_at >= $1 AND gv.recorded_at < $2
			  AND ($3::text IS NULL OR c.project_key = $3)
			GROUP BY gv.guard
		)
		SELECT COALESCE(r.guard, v.guard),
			COALESCE(r.runs, 0), COALESCE(r.fired, 0), COALESCE(r.cannot_answer, 0),
			r.last_run, r.last_fired, r.median_ms,
			COALESCE(v.verdicts, 0), COALESCE(v.false_positives, 0)
		FROM runs r
		FULL OUTER JOIN verdicts v ON v.guard = r.guard
		ORDER BY 1
	`, period.From, period.To, project)
	if err != nil {
		return nil, fmt.Errorf("store: guard activity: %w", err)
	}
	defer rows.Close()

	var out []GuardActivityRow
	for rows.Next() {
		var row GuardActivityRow
		if err := rows.Scan(&row.Guard, &row.Runs, &row.Fired, &row.CannotAnswer,
			&row.LastRunAt, &row.LastFiredAt, &row.MedianDurationMS,
			&row.Verdicts, &row.FalsePositives); err != nil {
			return nil, fmt.Errorf("store: guard activity: scan: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: guard activity: %w", err)
	}
	return out, nil
}

// StageRedoRow is one command/stage pair's re-entry record across a period.
// A stage run with attempt > 1 is a re-entry: the same stage begun again on
// the same change -- a fix run, a resumed session, a retried gate. Runs
// counts every attempt; Changes the distinct changes that ran the stage;
// ReenteredChanges those that ran it more than once.
//
// MedianSeconds and P90Seconds cover ended runs only, and are nil when no
// run in the group has ended.
type StageRedoRow struct {
	Command          string
	Stage            string
	Runs             int
	Changes          int
	Reentries        int
	ReenteredChanges int
	MedianSeconds    *float64
	P90Seconds       *float64
}

// StageRedo reports one row per command/stage pair whose runs started in
// period, optionally restricted to one project -- the same attribution
// StageLeaderboard uses. Plan-session runs (change_id NULL) join no change
// and are not counted.
func (s *Store) StageRedo(ctx context.Context, period Period, project *string) ([]StageRedoRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT sr.command, sr.stage,
			COUNT(*),
			COUNT(DISTINCT sr.change_id),
			COUNT(*) FILTER (WHERE sr.attempt > 1),
			COUNT(DISTINCT sr.change_id) FILTER (WHERE sr.attempt > 1),
			percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM sr.ended_at - sr.started_at))
				FILTER (WHERE sr.ended_at IS NOT NULL),
			percentile_cont(0.9) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM sr.ended_at - sr.started_at))
				FILTER (WHERE sr.ended_at IS NOT NULL)
		FROM stage_runs sr
		JOIN changes c ON c.id = sr.change_id
		WHERE sr.started_at >= $1 AND sr.started_at < $2
		  AND ($3::text IS NULL OR c.project_key = $3)
		GROUP BY sr.command, sr.stage
		ORDER BY sr.command, sr.stage
	`, period.From, period.To, project)
	if err != nil {
		return nil, fmt.Errorf("store: stage redo: %w", err)
	}
	defer rows.Close()

	var out []StageRedoRow
	for rows.Next() {
		var row StageRedoRow
		if err := rows.Scan(&row.Command, &row.Stage, &row.Runs, &row.Changes,
			&row.Reentries, &row.ReenteredChanges, &row.MedianSeconds, &row.P90Seconds); err != nil {
			return nil, fmt.Errorf("store: stage redo: scan: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: stage redo: %w", err)
	}
	return out, nil
}

// PanelRoundsRow is one change's review-panel record: how many rounds the
// panel took and what it found. A change is in scope when a flow.review-panel
// stage run of it started in the period.
//
// Rounds is the highest round any finding or pass-log line recorded, plus
// one (rounds are 0-based: 0 is the initial panel, 1..n the fix rounds) --
// nil when the change recorded neither, which is "not recorded", never "no
// rounds". Findings counts every finding the change recorded; the severity
// counts match case-insensitively, as Reviewers does.
type PanelRoundsRow struct {
	Project   string
	Change    string
	StartedAt time.Time
	Rounds    *int
	Findings  int
	Critical  int
	Important int
	Minor     int
}

// PanelRounds reports one row per change that entered the review panel in
// period, optionally restricted to one project, newest first.
func (s *Store) PanelRounds(ctx context.Context, period Period, project *string) ([]PanelRoundsRow, error) {
	rows, err := s.pool.Query(ctx, `
		WITH scoped AS (
			SELECT sr.change_id, MIN(sr.started_at) AS started_at
			FROM stage_runs sr
			JOIN changes c ON c.id = sr.change_id
			WHERE sr.stage = 'flow.review-panel'
			  AND sr.started_at >= $1 AND sr.started_at < $2
			  AND ($3::text IS NULL OR c.project_key = $3)
			GROUP BY sr.change_id
		),
		rounds AS (
			SELECT change_id, MAX(round) + 1 AS rounds
			FROM (
				SELECT f.change_id, f.round FROM findings f WHERE f.change_id IN (SELECT change_id FROM scoped)
				UNION ALL
				SELECT p.change_id, p.round FROM panel_passes p WHERE p.change_id IN (SELECT change_id FROM scoped)
			) r
			GROUP BY change_id
		),
		counts AS (
			SELECT f.change_id,
				COUNT(*) AS findings,
				COUNT(*) FILTER (WHERE f.severity ILIKE 'critical') AS critical,
				COUNT(*) FILTER (WHERE f.severity ILIKE 'important') AS important,
				COUNT(*) FILTER (WHERE f.severity ILIKE 'minor') AS minor
			FROM findings f
			WHERE f.change_id IN (SELECT change_id FROM scoped)
			GROUP BY f.change_id
		)
		SELECT c.project_key, c.name, s.started_at, r.rounds,
			COALESCE(n.findings, 0), COALESCE(n.critical, 0), COALESCE(n.important, 0), COALESCE(n.minor, 0)
		FROM scoped s
		JOIN changes c ON c.id = s.change_id
		LEFT JOIN rounds r ON r.change_id = s.change_id
		LEFT JOIN counts n ON n.change_id = s.change_id
		ORDER BY s.started_at DESC, c.name
	`, period.From, period.To, project)
	if err != nil {
		return nil, fmt.Errorf("store: panel rounds: %w", err)
	}
	defer rows.Close()

	var out []PanelRoundsRow
	for rows.Next() {
		var row PanelRoundsRow
		if err := rows.Scan(&row.Project, &row.Change, &row.StartedAt, &row.Rounds,
			&row.Findings, &row.Critical, &row.Important, &row.Minor); err != nil {
			return nil, fmt.Errorf("store: panel rounds: scan: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: panel rounds: %w", err)
	}
	return out, nil
}
