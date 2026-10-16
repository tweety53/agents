package store

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/tweety53/agents/stats/internal/records"
)

// health.go backs the flow-health dashboard: one set of period totals
// (FlowSummary) asking whether flow's own machinery earns its keep -- how
// many changes start and finish, what they cost and how long they take, how
// many review-panel rounds and findings they need, and how often a guard
// fires -- plus the one write the last of those needs (RecordGuardRun).

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

// FlowSummaryRow is the flow-health dashboard's eight period totals
// (design.md's flow-health-tiles). A change starts at its first stage run
// and finishes at its FINISHED write -- a withdrawal is not a finish.
// MeanCostPerFinished is the whole-life cost of the changes finished in the
// period; MedianLeadTimeMs runs from a finished change's first stage start
// to its finish. Both are nil when nothing finished, and MeanPanelRounds
// when no change in the period recorded a round -- absence, never a zero.
type FlowSummaryRow struct {
	ChangesStarted      int
	ChangesFinished     int
	TotalCostUsd        float64
	MeanCostPerFinished *float64
	MedianLeadTimeMs    *int64
	MeanPanelRounds     *float64
	Critical            int
	Important           int
	Minor               int
	GuardFires          int
}

// FlowSummary reports the period's totals, optionally restricted to one
// project. Stage-run cost is attributed by start instant, as every
// aggregation in aggregate.go is; the panel figures are PanelRounds' rows
// folded together.
func (s *Store) FlowSummary(ctx context.Context, period Period, project *string) (FlowSummaryRow, error) {
	var out FlowSummaryRow
	var lead *float64
	err := s.pool.QueryRow(ctx, `
		WITH costed AS (
			SELECT sr.change_id, sr.started_at,
				CASE WHEN jsonb_typeof(sr.metrics->'cost_usd') = 'number'
				     THEN (sr.metrics->>'cost_usd')::float8 END AS cost
			FROM stage_runs sr
			JOIN changes c ON c.id = sr.change_id
			WHERE ($3::text IS NULL OR c.project_key = $3)
		),
		firsts AS (
			SELECT change_id, MIN(started_at) AS first_start, SUM(cost) AS life_cost
			FROM costed
			GROUP BY change_id
		),
		finished AS (
			SELECT c.updated_at, f.first_start, COALESCE(f.life_cost, 0) AS life_cost
			FROM changes c
			LEFT JOIN firsts f ON f.change_id = c.id
			WHERE c.state = 'FINISHED' AND NOT c.withdrawn
			  AND c.updated_at >= $1 AND c.updated_at < $2
			  AND ($3::text IS NULL OR c.project_key = $3)
		)
		SELECT
			(SELECT COUNT(*)::int FROM firsts WHERE first_start >= $1 AND first_start < $2),
			(SELECT COUNT(*)::int FROM finished),
			(SELECT COALESCE(SUM(cost), 0)::float8 FROM costed WHERE started_at >= $1 AND started_at < $2),
			(SELECT AVG(life_cost)::float8 FROM finished),
			(SELECT percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM updated_at - first_start) * 1000)
				FROM finished WHERE first_start IS NOT NULL),
			(SELECT COUNT(*)::int FROM guard_runs gr
				WHERE gr.outcome = 'fired' AND gr.recorded_at >= $1 AND gr.recorded_at < $2
				  AND ($3::text IS NULL OR gr.project_key = $3))
	`, period.From, period.To, project).Scan(&out.ChangesStarted, &out.ChangesFinished, &out.TotalCostUsd,
		&out.MeanCostPerFinished, &lead, &out.GuardFires)
	if err != nil {
		return FlowSummaryRow{}, fmt.Errorf("store: flow summary: %w", err)
	}
	if lead != nil {
		ms := int64(math.Round(*lead))
		out.MedianLeadTimeMs = &ms
	}

	panels, err := s.PanelRounds(ctx, period, project)
	if err != nil {
		return FlowSummaryRow{}, err
	}
	var rounds, roundChanges int
	for _, p := range panels {
		out.Critical += p.Critical
		out.Important += p.Important
		out.Minor += p.Minor
		if p.Rounds != nil {
			rounds += *p.Rounds
			roundChanges++
		}
	}
	if roundChanges > 0 {
		mean := float64(rounds) / float64(roundChanges)
		out.MeanPanelRounds = &mean
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
