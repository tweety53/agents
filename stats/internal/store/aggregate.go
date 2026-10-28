package store

import (
	"context"
	"fmt"
	"time"
)

// Period bounds an aggregation by a stage run's start instant:
// [From, To). A stage run that starts before the period and ends inside it
// is attributed to whichever period contains its start -- never the one
// that contains its end -- and every aggregation method below applies that
// same convention, so a run is never double-counted across two adjacent
// periods and never silently dropped from both.
type Period struct {
	From time.Time
	To   time.Time
}

// LiveStateRow is one row of the live state board: one change, its current
// state, and when it was last updated. PlannedTasks and CurrentTasks are
// the change's first and latest change_task_counts observations (KAN-415)
// -- nil together when the change was never observed, which is the board's
// absence-is-never-zero case: no observation reads as no figure, never as
// a count of zero.
//
// RunCount, CostUsd, WallClockMs and Errors total the change's runs over
// its whole life, never bounded by the board's period (design.md's
// board-run-totals): a run is the stage rows sharing a session token, a
// token-less stage row is a run of its own; a run's wall clock spans its
// first start to its last end, and is 0 while any of its stages is open;
// cost sums each row's numeric cost_usd; errors sum tool and API errors
// from both the main and sidechain signals.
type LiveStateRow struct {
	ProjectKey string
	Name       string
	State      State
	UpdatedAt  time.Time
	UpdatedBy  string

	PlannedTasks *int
	CurrentTasks *int

	RunCount    int
	CostUsd     float64
	WallClockMs int64
	Errors      int
}

// signalErrorsExpr is one stage run's tool and API errors, main and
// sidechain, each counted only where the signal is a JSON number. Four
// plain terms rather than an unnest per row: the board pays this per stage
// run of every change it lists.
const signalErrorsExpr = `(CASE WHEN jsonb_typeof(sr.metrics #> '{signals,main,tool_errors}') = 'number' THEN (sr.metrics #>> '{signals,main,tool_errors}')::int ELSE 0 END
	+ CASE WHEN jsonb_typeof(sr.metrics #> '{signals,main,api_errors}') = 'number' THEN (sr.metrics #>> '{signals,main,api_errors}')::int ELSE 0 END
	+ CASE WHEN jsonb_typeof(sr.metrics #> '{signals,sidechain,tool_errors}') = 'number' THEN (sr.metrics #>> '{signals,sidechain,tool_errors}')::int ELSE 0 END
	+ CASE WHEN jsonb_typeof(sr.metrics #> '{signals,sidechain,api_errors}') = 'number' THEN (sr.metrics #>> '{signals,sidechain,api_errors}')::int ELSE 0 END)`

// LiveStateBoard lists every change updated within period, optionally
// restricted to one project. A period containing no changes returns an
// empty slice, never an error.
func (s *Store) LiveStateBoard(ctx context.Context, period Period, project *string) ([]LiveStateRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.project_key, c.name, c.state, c.updated_at, c.updated_by,
		       first_obs.total_tasks, latest_obs.total_tasks,
		       totals.runs, totals.cost, totals.wall_ms, totals.errors
		FROM changes c
		LEFT JOIN LATERAL (
			SELECT t.total_tasks
			FROM change_task_counts t
			WHERE t.change_id = c.id
			ORDER BY t.observed_at, t.id
			LIMIT 1
		) first_obs ON true
		LEFT JOIN LATERAL (
			SELECT t.total_tasks
			FROM change_task_counts t
			WHERE t.change_id = c.id
			ORDER BY t.observed_at DESC, t.id DESC
			LIMIT 1
		) latest_obs ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*)::int AS runs,
			       COALESCE(SUM(r.cost), 0)::float8 AS cost,
			       COALESCE(SUM(r.wall_ms), 0)::bigint AS wall_ms,
			       COALESCE(SUM(r.errors), 0)::int AS errors
			FROM (
				SELECT SUM(`+stageRunCostExpr+`) AS cost,
				       CASE WHEN bool_and(sr.ended_at IS NOT NULL)
				            THEN floor(EXTRACT(EPOCH FROM MAX(sr.ended_at) - MIN(sr.started_at)) * 1000)
				            ELSE 0 END AS wall_ms,
				       SUM(`+signalErrorsExpr+`) AS errors
				FROM stage_runs sr
				WHERE sr.change_id = c.id
				GROUP BY COALESCE(sr.session_token, 'stage-' || sr.id)
			) r
		) totals ON true
		WHERE c.updated_at >= $1 AND c.updated_at < $2
		  AND ($3::text IS NULL OR c.project_key = $3)
		ORDER BY c.updated_at DESC, c.project_key, c.name
	`, period.From, period.To, project)
	if err != nil {
		return nil, fmt.Errorf("store: live state board: %w", err)
	}
	defer rows.Close()

	var out []LiveStateRow
	for rows.Next() {
		var (
			row   LiveStateRow
			state string
		)
		if err := rows.Scan(&row.ProjectKey, &row.Name, &state, &row.UpdatedAt, &row.UpdatedBy,
			&row.PlannedTasks, &row.CurrentTasks,
			&row.RunCount, &row.CostUsd, &row.WallClockMs, &row.Errors); err != nil {
			return nil, fmt.Errorf("store: live state board: scan: %w", err)
		}
		row.State = State(state)
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: live state board: %w", err)
	}
	return out, nil
}

// CostPerChangeRow is one change's cost, broken down by command and stage,
// for one period. RunCount is every stage run in the group; MeasuredRuns
// is how many of them carried a "tokens" key at all -- the rest are
// excluded from every token figure below rather than averaged in as zero,
// per the metrics bag's absence-is-not-a-value rule. MainTokens and
// SidechainTokens total the metrics bag's tokens.main and tokens.sidechain
// buckets' own fields (internal/harvest.Bucket's shape) separately, so a
// dispatching command's own cost stays distinguishable from what it
// dispatched. A bucket that was never written (no "main" or "sidechain"
// key at all) contributes nothing to its side's total rather than a
// recorded zero -- distinct from a bucket that was written with every
// field at zero, which is a real, measured absence-of-usage.
type CostPerChangeRow struct {
	ProjectKey string
	ChangeName string
	Command    string
	Stage      string

	RunCount     int
	MeasuredRuns int

	TotalTokensInput *int64
	MeanTokensInput  *float64
	TotalCostUSD     *float64
	TotalDurationMs  *int64

	MainTokens      *int64
	SidechainTokens *int64
}

// CostPerChange returns end-to-end cost broken down by change, command and
// stage, for every stage run starting within period and, when project is
// non-nil, belonging to that project. A multi-repository change is one
// unit here: rows group by the change's identity, never by its individual
// repositories.
//
// model, when non-nil, restricts to stage runs whose metrics bag's "models"
// object recorded that model (sr.metrics->'models' ? *model) -- never the
// retired scalar "model" key, which nothing has written since task 22 --
// and every token and cost figure the row reports is that model's own
// bucket (metrics.models.<model>), not the whole run's: a two-model run
// costing $61.10 across Opus and Sonnet, filtered to Sonnet, reports
// Sonnet's own $19.90, never the run's $61.10 (which would attribute the
// Opus parent to Sonnet) and never $0 (which would read as "Sonnet was
// never used"). A run whose metrics carry no "models" key at all matches
// no model filter, per this task's own absence-is-not-a-match rule
// (task 21, step 2) -- store.CountRunsWithoutModel answers how many such
// runs a filtered caller silently excluded.
func (s *Store) CostPerChange(ctx context.Context, period Period, project, model *string) ([]CostPerChangeRow, error) {
	rows, err := s.pool.Query(ctx, `
		WITH scoped AS (
			SELECT
				c.project_key, c.name, sr.command, sr.stage, sr.started_at, sr.ended_at,
				CASE WHEN $4::text IS NULL THEN sr.metrics->'tokens'
				     ELSE sr.metrics->'models'->$4::text->'tokens' END AS tok,
				CASE WHEN $4::text IS NULL THEN (sr.metrics->>'cost_usd')::numeric
				     ELSE (sr.metrics->'models'->$4::text->>'cost_usd')::numeric END AS cost
			FROM stage_runs sr
			JOIN changes c ON c.id = sr.change_id
			WHERE sr.started_at >= $1 AND sr.started_at < $2
			  AND ($3::text IS NULL OR c.project_key = $3)
			  AND ($4::text IS NULL OR sr.metrics->'models' ? $4)
		)
		SELECT
			project_key, name, command, stage,
			COUNT(*) AS run_count,
			COUNT(*) FILTER (WHERE tok IS NOT NULL) AS measured_runs,
			SUM(
				CASE WHEN tok IS NOT NULL THEN
					COALESCE((tok->'main'->>'input')::numeric, 0)
					+ COALESCE((tok->'sidechain'->>'input')::numeric, 0)
				END
			) AS total_tokens_input,
			AVG(
				CASE WHEN tok IS NOT NULL THEN
					COALESCE((tok->'main'->>'input')::numeric, 0)
					+ COALESCE((tok->'sidechain'->>'input')::numeric, 0)
				END
			) AS mean_tokens_input,
			SUM(cost) AS total_cost_usd,
			SUM(
				CASE WHEN ended_at IS NOT NULL
				THEN EXTRACT(EPOCH FROM (ended_at - started_at)) * 1000
				END
			)::bigint AS total_duration_ms,
			SUM(
				CASE WHEN tok->'main' IS NOT NULL THEN
					COALESCE((tok->'main'->>'input')::numeric, 0)
					+ COALESCE((tok->'main'->>'output')::numeric, 0)
					+ COALESCE((tok->'main'->>'cache_creation')::numeric, 0)
					+ COALESCE((tok->'main'->>'cache_read')::numeric, 0)
					+ COALESCE((tok->'main'->>'thinking')::numeric, 0)
				END
			) AS main_tokens,
			SUM(
				CASE WHEN tok->'sidechain' IS NOT NULL THEN
					COALESCE((tok->'sidechain'->>'input')::numeric, 0)
					+ COALESCE((tok->'sidechain'->>'output')::numeric, 0)
					+ COALESCE((tok->'sidechain'->>'cache_creation')::numeric, 0)
					+ COALESCE((tok->'sidechain'->>'cache_read')::numeric, 0)
					+ COALESCE((tok->'sidechain'->>'thinking')::numeric, 0)
				END
			) AS sidechain_tokens
		FROM scoped
		GROUP BY project_key, name, command, stage
		ORDER BY project_key, name, command, stage
	`, period.From, period.To, project, model)
	if err != nil {
		return nil, fmt.Errorf("store: cost per change: %w", err)
	}
	defer rows.Close()

	var out []CostPerChangeRow
	for rows.Next() {
		var row CostPerChangeRow
		if err := rows.Scan(
			&row.ProjectKey, &row.ChangeName, &row.Command, &row.Stage,
			&row.RunCount, &row.MeasuredRuns,
			&row.TotalTokensInput, &row.MeanTokensInput, &row.TotalCostUSD, &row.TotalDurationMs,
			&row.MainTokens, &row.SidechainTokens,
		); err != nil {
			return nil, fmt.Errorf("store: cost per change: scan: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: cost per change: %w", err)
	}
	return out, nil
}

// StageLeaderboardRow ranks one command/stage pair by cost across a
// period. Only stage runs a Price call has already costed contribute --
// an uncosted run is excluded from every figure here, never priced as
// zero.
type StageLeaderboardRow struct {
	Command  string
	Stage    string
	RunCount int

	MeanCostUSD   float64
	MedianCostUSD float64
	P90CostUSD    float64
	TotalCostUSD  float64
}

// StageLeaderboard ranks stages by total cost across period, optionally
// restricted to one project, and keeps only the ten costliest (design.md's
// leaderboard-top10-total); each row also carries mean, median and p90.
//
// model, when non-nil, restricts to stage runs whose metrics recorded that
// model (per CostPerChange's own doc comment) and ranks by that model's own
// cost_usd bucket, never the run's total.
func (s *Store) StageLeaderboard(ctx context.Context, period Period, project, model *string) ([]StageLeaderboardRow, error) {
	rows, err := s.pool.Query(ctx, `
		WITH costed AS (
			SELECT sr.command, sr.stage,
				CASE WHEN $4::text IS NULL THEN (sr.metrics->>'cost_usd')::numeric
				     ELSE (sr.metrics->'models'->$4::text->>'cost_usd')::numeric END AS cost
			FROM stage_runs sr
			JOIN changes c ON c.id = sr.change_id
			WHERE sr.started_at >= $1 AND sr.started_at < $2
			  AND ($3::text IS NULL OR c.project_key = $3)
			  AND ($4::text IS NULL OR sr.metrics->'models' ? $4)
			  AND (CASE WHEN $4::text IS NULL THEN sr.metrics ? 'cost_usd'
			            ELSE sr.metrics->'models'->$4::text ? 'cost_usd' END)
		)
		SELECT
			command, stage, COUNT(*),
			AVG(cost),
			percentile_cont(0.5) WITHIN GROUP (ORDER BY cost),
			percentile_cont(0.9) WITHIN GROUP (ORDER BY cost),
			SUM(cost)
		FROM costed
		GROUP BY command, stage
		ORDER BY SUM(cost) DESC, command, stage
		LIMIT 10
	`, period.From, period.To, project, model)
	if err != nil {
		return nil, fmt.Errorf("store: stage leaderboard: %w", err)
	}
	defer rows.Close()

	var out []StageLeaderboardRow
	for rows.Next() {
		var row StageLeaderboardRow
		if err := rows.Scan(&row.Command, &row.Stage, &row.RunCount, &row.MeanCostUSD, &row.MedianCostUSD, &row.P90CostUSD, &row.TotalCostUSD); err != nil {
			return nil, fmt.Errorf("store: stage leaderboard: scan: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: stage leaderboard: %w", err)
	}
	return out, nil
}

// TrendPoint is one day's totals within a trend-over-time query.
type TrendPoint struct {
	Day          time.Time
	RunCount     int
	TotalCostUSD *float64
}

// TrendOverTime buckets cost by day across period, optionally restricted
// to one project, so a reader can see whether the pipeline is getting
// cheaper or more expensive as flow itself changes.
//
// model, when non-nil, restricts to stage runs whose metrics recorded that
// model (per CostPerChange's own doc comment) and sums that model's own
// cost_usd bucket per day, never the run's total.
func (s *Store) TrendOverTime(ctx context.Context, period Period, project, model *string) ([]TrendPoint, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			date_trunc('day', sr.started_at) AS day,
			COUNT(*),
			SUM(
				CASE WHEN $4::text IS NULL THEN (sr.metrics->>'cost_usd')::numeric
				     ELSE (sr.metrics->'models'->$4::text->>'cost_usd')::numeric END
			)
		FROM stage_runs sr
		JOIN changes c ON c.id = sr.change_id
		WHERE sr.started_at >= $1 AND sr.started_at < $2
		  AND ($3::text IS NULL OR c.project_key = $3)
		  AND ($4::text IS NULL OR sr.metrics->'models' ? $4)
		GROUP BY day
		ORDER BY day
	`, period.From, period.To, project, model)
	if err != nil {
		return nil, fmt.Errorf("store: trend over time: %w", err)
	}
	defer rows.Close()

	var out []TrendPoint
	for rows.Next() {
		var p TrendPoint
		if err := rows.Scan(&p.Day, &p.RunCount, &p.TotalCostUSD); err != nil {
			return nil, fmt.Errorf("store: trend over time: scan: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: trend over time: %w", err)
	}
	return out, nil
}

// weeklyCostStages are the stage keys each WeeklyCosts category counts,
// drawn from the stages vocabulary; TestWeeklyCostStagesAreDeclared fails on
// a key it does not declare, so a rename there cannot silently move cost
// into overall-only.
var weeklyCostStages = struct {
	selfReview, visualVerify, implementation, review string
	planning                                         []string
}{
	selfReview:     "flow.self-review",
	visualVerify:   "flow.visual-verify",
	implementation: "flow.sdd-tdd",
	review:         "flow.review-panel",
	planning:       []string{"plan.session", "flow.brainstorm", "flow.design-approval", "flow.create-artifacts", "flow.writing-plans", "flow.decide"},
}

// WeeklyCostPoint is one Wednesday-to-Wednesday week's cost totals: overall,
// the five pipeline categories, and Other -- every run in none of them, so
// the categories and Other add up to overall (decision other-category). A
// category with no costed run that week is nil, never 0.
type WeeklyCostPoint struct {
	Week            time.Time
	OverallUSD      *float64
	PlanningUSD     *float64
	ImplementUSD    *float64
	ReviewUSD       *float64
	SelfReviewUSD   *float64
	VisualVerifyUSD *float64
	OtherUSD        *float64
}

// WeeklyCosts buckets cost by week starting Wednesday 00:00 UTC, split by
// category from the run's stage. A self-review or pipeline-fix loop change
// is classified first, so its stages never count as planning or
// implementation; flow.visual-verify is classified after it (decision
// visual-verify-category); a stage outside every category counts in Other.
// project and model filter as TrendOverTime's do.
func (s *Store) WeeklyCosts(ctx context.Context, period Period, project, model *string) ([]WeeklyCostPoint, error) {
	rows, err := s.pool.Query(ctx, `
		WITH r AS (
			SELECT
				date_trunc('week', sr.started_at - interval '2 days', 'UTC') + interval '2 days' AS week,
				CASE WHEN $4::text IS NULL THEN (sr.metrics->>'cost_usd')::numeric
				     ELSE (sr.metrics->'models'->$4::text->>'cost_usd')::numeric END AS cost,
				CASE
					WHEN c.name LIKE 'self-review-%' OR c.name LIKE 'pipeline-fix-%' OR sr.stage = $5 THEN 'self-review'
					WHEN sr.stage = ANY($8::text[]) THEN 'planning'
					WHEN sr.stage = $6 THEN 'implementation'
					WHEN sr.stage = $7 THEN 'review'
					WHEN sr.stage = $9 THEN 'visual-verify'
				END AS category
			FROM stage_runs sr
			JOIN changes c ON c.id = sr.change_id
			WHERE sr.started_at >= $1 AND sr.started_at < $2
			  AND ($3::text IS NULL OR c.project_key = $3)
			  AND ($4::text IS NULL OR sr.metrics->'models' ? $4)
		)
		SELECT week,
			SUM(cost),
			SUM(cost) FILTER (WHERE category = 'planning'),
			SUM(cost) FILTER (WHERE category = 'implementation'),
			SUM(cost) FILTER (WHERE category = 'review'),
			SUM(cost) FILTER (WHERE category = 'self-review'),
			SUM(cost) FILTER (WHERE category = 'visual-verify'),
			SUM(cost) FILTER (WHERE category IS NULL)
		FROM r
		GROUP BY week
		ORDER BY week
	`, period.From, period.To, project, model,
		weeklyCostStages.selfReview, weeklyCostStages.implementation, weeklyCostStages.review, weeklyCostStages.planning,
		weeklyCostStages.visualVerify)
	if err != nil {
		return nil, fmt.Errorf("store: weekly costs: %w", err)
	}
	defer rows.Close()

	var out []WeeklyCostPoint
	for rows.Next() {
		var p WeeklyCostPoint
		if err := rows.Scan(&p.Week, &p.OverallUSD, &p.PlanningUSD, &p.ImplementUSD, &p.ReviewUSD, &p.SelfReviewUSD, &p.VisualVerifyUSD, &p.OtherUSD); err != nil {
			return nil, fmt.Errorf("store: weekly costs: scan: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: weekly costs: %w", err)
	}
	return out, nil
}

// WeeklyModelEffortPoint is one week's cost for one model and effort: a
// dispatch effort, "default" for a dispatch with none served, or
// "session" for the parent session's own share of that model's cost.
type WeeklyModelEffortPoint struct {
	Week    time.Time
	Model   string
	Effort  string
	CostUSD float64
}

// WeeklyModelEffortCosts buckets cost by week (WeeklyCosts' Wednesday
// weeks) and model·effort across every project. A priced dispatch counts
// under its most-turns served model (dispatch-served-model) and most-turns
// served effort, else "default" (dispatch-served-effort); each stage run's
// priced model bucket, less the dispatch costs served by that model and
// floored at 0, counts as that model's "session" (session-line). Ties in
// most-turns break by key name. Rows summing to 0 are dropped; rows are
// ordered by week, model, effort.
func (s *Store) WeeklyModelEffortCosts(ctx context.Context, period Period) ([]WeeklyModelEffortPoint, error) {
	rows, err := s.pool.Query(ctx, `
		WITH sr AS (
			SELECT id, started_at,
				CASE WHEN jsonb_typeof(metrics->'dispatches') = 'object' THEN metrics->'dispatches' END AS dispatches,
				CASE WHEN jsonb_typeof(metrics->'models') = 'object' THEN metrics->'models' END AS models
			FROM stage_runs
			WHERE started_at >= $1 AND started_at < $2
		), d AS (
			SELECT sr.id, sr.started_at, (b.value->>'cost_usd')::numeric AS cost,
				COALESCE(
					(SELECT k FROM jsonb_each(CASE WHEN jsonb_typeof(b.value->'signals'->'served_models') = 'object'
						THEN b.value->'signals'->'served_models' END) AS e(k, v) ORDER BY v DESC, k LIMIT 1),
					b.value->>'model') AS model,
				COALESCE(
					(SELECT k FROM jsonb_each(CASE WHEN jsonb_typeof(b.value->'signals'->'served_efforts') = 'object'
						THEN b.value->'signals'->'served_efforts' END) AS e(k, v) ORDER BY v DESC, k LIMIT 1),
					'default') AS effort
			FROM sr, jsonb_each(sr.dispatches) AS b
			WHERE b.value ? 'cost_usd'
		), s AS (
			SELECT sr.started_at, m.key AS model, 'session' AS effort,
				GREATEST(0, (m.value->>'cost_usd')::numeric
					- COALESCE((SELECT SUM(d.cost) FROM d WHERE d.id = sr.id AND d.model = m.key), 0)) AS cost
			FROM sr, jsonb_each(sr.models) AS m
			WHERE m.value ? 'cost_usd'
		), all_rows AS (
			SELECT started_at, model, effort, cost FROM d
			UNION ALL
			SELECT started_at, model, effort, cost FROM s
		)
		SELECT date_trunc('week', started_at - interval '2 days', 'UTC') + interval '2 days' AS week,
			model, effort, SUM(cost)
		FROM all_rows
		GROUP BY week, model, effort
		HAVING SUM(cost) <> 0
		ORDER BY week, model, effort
	`, period.From, period.To)
	if err != nil {
		return nil, fmt.Errorf("store: weekly model-effort costs: %w", err)
	}
	defer rows.Close()

	var out []WeeklyModelEffortPoint
	for rows.Next() {
		var p WeeklyModelEffortPoint
		if err := rows.Scan(&p.Week, &p.Model, &p.Effort, &p.CostUSD); err != nil {
			return nil, fmt.Errorf("store: weekly model-effort costs: scan: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: weekly model-effort costs: %w", err)
	}
	return out, nil
}

// CountRunsWithoutModel returns how many stage runs starting within period
// (and, when project is non-nil, belonging to that project) recorded no
// model at all -- their metrics bag carries no "models" key, or carries it
// as an empty object or JSON null. These are the runs a model filter
// excludes on the ground of absence, not on the ground of belonging to a
// different model (task 21, step 2): a harness that wrote no readable
// transcript, for instance. It is unaffected by which model a caller is
// about to filter for, since a run with no recorded model could not have
// matched any model filter -- callers invoke this only when a model
// filter is set, so an unfiltered request never pays for it.
func (s *Store) CountRunsWithoutModel(ctx context.Context, period Period, project *string) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM stage_runs sr
		JOIN changes c ON c.id = sr.change_id
		WHERE sr.started_at >= $1 AND sr.started_at < $2
		  AND ($3::text IS NULL OR c.project_key = $3)
		  AND (
			NOT (sr.metrics ? 'models')
			OR sr.metrics->'models' = 'null'::jsonb
			OR sr.metrics->'models' = '{}'::jsonb
		  )
	`, period.From, period.To, project).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("store: count runs without model: %w", err)
	}
	return count, nil
}

// ListModels returns the distinct models recorded by any stage run starting
// within period (and, when project is non-nil, belonging to that project),
// sorted -- the metrics bag's "models" object keys, obtained from the
// server rather than assumed or hard-coded, so a model used for the first
// time appears the moment it has been recorded, with no change to this
// build. A stage run recording no model contributes no key.
func (s *Store) ListModels(ctx context.Context, period Period, project *string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT key
		FROM stage_runs sr
		JOIN changes c ON c.id = sr.change_id
		CROSS JOIN LATERAL jsonb_object_keys(sr.metrics->'models') AS key
		WHERE sr.started_at >= $1 AND sr.started_at < $2
		  AND ($3::text IS NULL OR c.project_key = $3)
		ORDER BY key
	`, period.From, period.To, project)
	if err != nil {
		return nil, fmt.Errorf("store: list models: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var model string
		if err := rows.Scan(&model); err != nil {
			return nil, fmt.Errorf("store: list models: scan: %w", err)
		}
		out = append(out, model)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list models: %w", err)
	}
	return out, nil
}

// ReviewerRow is one review-panel slot's record across a period: how often
// it was dispatched, across how many changes, and what it found by
// severity. Experimental is true for a slot whose id carries the "exp-" prefix (design.md's
// exp-slot-prefix decision -- no schema flag, the id itself is the fact);
// Description is that slot's newest recorded decision roster entry naming
// it, empty when no decision ever named it (a default-panel run, or a slot
// whose roster entry carries no description at all).
type ReviewerRow struct {
	Slot         string
	Experimental bool
	Description  string

	Dispatches int
	Changes    int
	Critical   int
	Important  int
	Minor      int
}

// Reviewers reports one row per role -- design.md's "Stats views ›
// reviewers" -- for every dispatch whose slot is set and whose own
// started_at falls within period (and, when project/model is non-nil,
// matches it). It scopes on the dispatch's start, never through
// stage_run_id: that column is nullable and nothing that records a
// dispatch sets it, so a join on it returns no rows at all.
//
// A bundle is one dispatches row whose slot carries several roles
// "+"-joined (design.md's compound-slot-one-row-per-bundle); scoped_dispatches
// unnests that column so the bundle counts once for each role it carried,
// never once under the compound string. findings.dispatch_id (the row the
// -dispatch-seq option resolves to, 0010_run_records.sql:85) is what
// scoped_findings joins on, not (change_id, slot) -- a bundle's slot column
// never equals any one role's name, so a raw-slot join would drop every
// bundled finding. A finding recorded with no dispatch seq (dispatch_id
// NULL) joins nothing and is not counted.
//
// dispatch_agg reduces scoped_dispatches to one row per slot before
// findings ever enter the query, so a slot dispatched more than once in the
// same change cannot fan a finding out into being counted twice --
// scoped_findings then joins findings to that same scope and DISTINCTs on
// the finding's own identity for the same reason. Severity and status are
// matched case-insensitively: this store
// treats them as free text written by whichever caller recorded the
// finding (records_test.go's own fixtures use both "important" and "Important"),
// and a reviewer's catch-rate should not depend on which casing a given
// dispatch happened to write.
//
// Description is read from the newest decisions row whose
// decision->'panel'->'roster' carries an entry naming this slot -- one
// DISTINCT ON pass over every roster entry (slot_desc), never a scan of
// decisions per slot -- independent of period and
// project -- a roster entry describes what a slot IS, not when it ran, so
// scoping it to the query's own period would make the same slot's
// description flicker between periods for no reason tied to the roster
// itself.
//
// roster_slots is also the view's membership test: a slot is a row only
// when some recorded roster names it. Dispatches of other roles (panel-fix,
// verifier) and one-off reviewers sometimes carry an invented -slot, and
// without this join each surfaced as a zero-findings row. The always-on
// extra dispatches' slots are recorded under extra_dispatch.slots and
// extra_dispatch_mutation.slots, never in the roster, so they count as
// members too; those records carry no description, so theirs comes from
// extraDispatchDescriptions.
func (s *Store) Reviewers(ctx context.Context, period Period, project, model *string) ([]ReviewerRow, error) {
	rows, err := s.pool.Query(ctx, `
		WITH scoped_dispatches AS (
			SELECT d.id, d.change_id, unnest(string_to_array(d.slot, '+')) AS slot
			FROM dispatches d
			JOIN changes c ON c.id = d.change_id
			WHERE d.started_at >= $1 AND d.started_at < $2
			  AND ($3::text IS NULL OR c.project_key = $3)
			  AND ($4::text IS NULL OR d.model = $4)
			  AND d.slot IS NOT NULL
		),
		dispatch_agg AS (
			SELECT slot, COUNT(*) AS dispatches, COUNT(DISTINCT change_id) AS changes
			FROM scoped_dispatches
			GROUP BY slot
		),
		scoped_findings AS (
			SELECT DISTINCT f.change_id, f.slot, f.ref, f.severity, f.status
			FROM findings f
			JOIN scoped_dispatches sd ON sd.id = f.dispatch_id AND sd.slot = f.slot
		),
		finding_agg AS (
			SELECT slot,
				COUNT(*) FILTER (WHERE severity ILIKE 'critical') AS critical,
				COUNT(*) FILTER (WHERE severity ILIKE 'important') AS important,
				COUNT(*) FILTER (WHERE severity ILIKE 'minor') AS minor
			FROM scoped_findings
			GROUP BY slot
		),
		slot_desc AS (
			SELECT DISTINCT ON (r->>'slot') r->>'slot' AS slot, r->>'description' AS description
			FROM decisions d, jsonb_path_query(d.decision, '$.panel.roster[*]') r
			WHERE r->>'slot' IS NOT NULL
			ORDER BY r->>'slot', d.recorded_at DESC
		),
		roster_slots AS (
			SELECT slot FROM slot_desc
			UNION
			SELECT jsonb_path_query(decision, '$.extra_dispatch.slots[*]') #>> '{}'
			FROM decisions
			UNION
			SELECT jsonb_path_query(decision, '$.extra_dispatch_mutation.slots[*]') #>> '{}'
			FROM decisions
		)
		SELECT
			da.slot,
			da.slot LIKE 'exp-%' AS experimental,
			COALESCE(sd.description, '') AS description,
			da.dispatches,
			da.changes,
			COALESCE(fa.critical, 0),
			COALESCE(fa.important, 0),
			COALESCE(fa.minor, 0)
		FROM dispatch_agg da
		JOIN roster_slots rs ON rs.slot = da.slot
		LEFT JOIN slot_desc sd ON sd.slot = da.slot
		LEFT JOIN finding_agg fa ON fa.slot = da.slot
		ORDER BY da.slot
	`, period.From, period.To, project, model)
	if err != nil {
		return nil, fmt.Errorf("store: reviewers: %w", err)
	}
	defer rows.Close()

	var out []ReviewerRow
	for rows.Next() {
		var row ReviewerRow
		if err := rows.Scan(
			&row.Slot, &row.Experimental, &row.Description,
			&row.Dispatches, &row.Changes, &row.Critical, &row.Important, &row.Minor,
		); err != nil {
			return nil, fmt.Errorf("store: reviewers: scan: %w", err)
		}
		if row.Description == "" {
			row.Description = extraDispatchDescriptions[row.Slot]
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: reviewers: %w", err)
	}
	return out, nil
}

// extraDispatchDescriptions describes the always-on extra dispatches'
// slots (extra_dispatch and extra_dispatch_mutation), in the words of
// skills/flow/review-panel.md's table for them.
var extraDispatchDescriptions = map[string]string{
	"test-sanity":     "Test-sanity — would each test the diff adds or changes fail if its behaviour broke",
	"call-efficiency": "Call-efficiency — second callers, a read repeated in one request, a query per list item",
	"failure-modes":   "Failure-modes — error return, timeout, partial write, concurrent re-entry",
	"mutation":        "Mutation — sabotage-proofing",
}

// DecisionRow is one recorded planner decision (design.md's "The decision
// chain") joined to its own run's totals: the wall-clock, tokens, cost,
// findings and fallback/timed-out dispatch counts of every stage run and
// dispatch sharing that decision's (change_id, session_token) -- the same
// correlator RecordDecision itself is keyed on. The model filter does not
// apply here (a decision spans models, design.md's "Stats views ›
// decisions"), so unlike every other row type in this file there is no
// model field.
//
// Overridden is true exactly when the decision's own "override" field
// carries a reason (design.md's mechanical-class-one-step-override
// decision) -- a JSON null or an absent key both read as false. ImplementerModel
// and ImplementerEffort come from the decision's "implementer" field, which
// is polymorphic: an object ({"model":...,"effort":...}) when execution
// was "sdd", or a bare string ("skipped — inline", or "default" on rows
// recorded before the project toggles were removed) otherwise -- ImplementerModel carries the model in
// the first case and the string itself in the second, so the column always
// has something to show; ImplementerEffort is empty in the second case,
// since a string carries no effort to read. Fixer reads the decision's
// "fixer" field the same way, rendered as one "model/effort" string when it
// is an object and the string itself otherwise; a decision recorded before
// the fixer pair existed has no key and reads as "". RerunDispatch renders
// "panel.rerun_dispatch" as "model/effort" when panel is an object carrying
// one, else "". RosterSize, Compact,
// ExperimentalSlot and Rerun read the same way from "panel", which is
// either an object or the bare string "default" -- the zero values (0,
// false, "", "") stand in for "default" rather than an error, since a
// default-panel run recorded no per-slot detail to report. Grouping is
// "panel.grouping" when panel is an object, else the literal "default".
// Dispatches renders "panel.dispatches" -- one to two {slots, model,
// effort} objects -- as each entry's slot ids '+'-joined and the entries
// themselves ' · '-joined, in roster order; empty when panel carries no
// dispatches to show. ImplementerGroups renders the decision's own
// top-level "groups" the same way over {bundles, model, effort} objects;
// empty when "groups" is JSON null (inline execution) or absent.
type DecisionRow struct {
	Project    string
	Change     string
	RecordedAt time.Time

	Class      string
	Overridden bool
	Execution  string

	ImplementerModel  string
	ImplementerEffort string
	Fixer             string
	RerunDispatch     string

	RosterSize       int
	Compact          bool
	ExperimentalSlot string
	Rerun            string

	Grouping          string
	Dispatches        string
	ImplementerGroups string

	WallClockSeconds float64

	InputTokens     int64
	OutputTokens    int64
	CacheReadTokens int64
	CostUsd         float64

	Critical  int
	Important int
	Minor     int
	FixRounds int

	Fallbacks int
	TimedOut  int
}

// Decisions reports one row per decisions row recorded within period
// (matched against the decision's own recorded_at, the same "the row's own
// instant governs" convention LiveStateBoard applies to updated_at) and,
// when project is non-nil, belonging to that project -- design.md's "Stats
// views › decisions". Every run total is joined by (change_id,
// session_token), the same pair RecordDecision, RecordDispatch and
// BeginStage's own SessionToken all key on; a decision recorded with no
// matching stage run, dispatch or finding still returns a row, with every
// joined figure at its zero value -- a decision is a real event on its
// own, independent of whether anything was ever attributed to its run.
//
// Token figures sum dispatches.metrics' "main" and "sidechain" buckets;
// cost sums the run's own stage_runs.metrics "cost_usd", where Price
// writes it, the same source CostPerChange reads -- a dispatch's own
// metrics bag never carries cost_usd. Outcome counts (fallback,
// timed-out) use the literal outcome strings skills/flow/*.md's dispatch
// paragraphs record. FixRounds is the highest findings.round any of the
// run's own dispatches raised, 0 when none did.
func (s *Store) Decisions(ctx context.Context, period Period, project *string) ([]DecisionRow, error) {
	rows, err := s.pool.Query(ctx, `
		WITH scoped_decisions AS (
			SELECT d.id, d.change_id, d.session_token, d.recorded_at, d.decision,
				c.project_key, c.name AS change_name
			FROM decisions d
			JOIN changes c ON c.id = d.change_id
			WHERE d.recorded_at >= $1 AND d.recorded_at < $2
			  AND ($3::text IS NULL OR c.project_key = $3)
		),
		panel_meta AS (
			SELECT id, jsonb_typeof(decision->'panel') = 'object' AS panel_is_object
			FROM scoped_decisions
		),
		group_sources AS (
			-- One row per (decision, array) pair worth flattening -- "panel.dispatches"
			-- and top-level "groups" are both an array of id-carrying objects, so both feed
			-- the same flattening logic below rather than each carrying its own copy.
			SELECT sd.id AS decision_id, src.label, src.arr
			FROM scoped_decisions sd
			JOIN panel_meta pm ON pm.id = sd.id
			CROSS JOIN LATERAL (VALUES
				('dispatches', CASE WHEN pm.panel_is_object THEN sd.decision->'panel'->'dispatches' ELSE NULL END),
				('groups', CASE WHEN jsonb_typeof(sd.decision->'groups') = 'array' THEN sd.decision->'groups' ELSE NULL END)
			) AS src(label, arr)
			WHERE src.arr IS NOT NULL
		),
		flattened_groups AS (
			-- Flatten one array of groups into a "·"-joined string of "+"-joined
			-- groups, in array order -- the one shape both "dispatches" and
			-- "groups" render into. A group is an object carrying its ids under
			-- "slots" (dispatches) or "bundles" (groups) beside its own model and
			-- effort.
			SELECT gs.decision_id, gs.label,
				(SELECT string_agg(joined, ' · ' ORDER BY grp.ord)
					FROM jsonb_array_elements(gs.arr) WITH ORDINALITY AS grp(val, ord)
					CROSS JOIN LATERAL (
						SELECT string_agg(e.elem, '+' ORDER BY e.eord) AS joined
						FROM jsonb_array_elements_text(COALESCE(grp.val->'slots', grp.val->'bundles'))
							WITH ORDINALITY AS e(elem, eord)
					) roles
				) AS flattened
			FROM group_sources gs
		),
		run_dispatches AS (
			SELECT disp.id, disp.change_id, disp.session_token, disp.metrics, disp.outcome
			FROM dispatches disp
			JOIN scoped_decisions sd ON sd.change_id = disp.change_id AND sd.session_token = disp.session_token
		),
		dispatch_agg AS (
			SELECT change_id, session_token,
				SUM(
					COALESCE((metrics->'tokens'->'main'->>'input')::numeric, 0)
					+ COALESCE((metrics->'tokens'->'sidechain'->>'input')::numeric, 0)
				) AS input_tokens,
				SUM(
					COALESCE((metrics->'tokens'->'main'->>'output')::numeric, 0)
					+ COALESCE((metrics->'tokens'->'sidechain'->>'output')::numeric, 0)
				) AS output_tokens,
				SUM(
					COALESCE((metrics->'tokens'->'main'->>'cache_read')::numeric, 0)
					+ COALESCE((metrics->'tokens'->'sidechain'->>'cache_read')::numeric, 0)
				) AS cache_read_tokens,
				COUNT(*) FILTER (WHERE outcome = 'fallback') AS fallbacks,
				COUNT(*) FILTER (WHERE outcome = 'timed-out') AS timed_out
			FROM run_dispatches
			GROUP BY change_id, session_token
		),
		finding_agg AS (
			SELECT rd.change_id, rd.session_token,
				COUNT(*) FILTER (WHERE f.severity ILIKE 'critical') AS critical,
				COUNT(*) FILTER (WHERE f.severity ILIKE 'important') AS important,
				COUNT(*) FILTER (WHERE f.severity ILIKE 'minor') AS minor,
				MAX(f.round) AS fix_rounds
			FROM run_dispatches rd
			JOIN findings f ON f.dispatch_id = rd.id
			GROUP BY rd.change_id, rd.session_token
		),
		runtime_agg AS (
			SELECT sd.change_id, sd.session_token,
				EXTRACT(EPOCH FROM (MAX(sr.ended_at) - MIN(sr.started_at))) AS wall_clock_seconds,
				SUM(COALESCE((sr.metrics->>'cost_usd')::numeric, 0)) AS cost_usd
			FROM scoped_decisions sd
			JOIN stage_runs sr ON sr.change_id = sd.change_id AND sr.session_token = sd.session_token
			GROUP BY sd.change_id, sd.session_token
		)
		SELECT
			sd.project_key, sd.change_name, sd.recorded_at,
			COALESCE(sd.decision->>'class', ''),
			(sd.decision->>'override') IS NOT NULL,
			sd.decision->>'execution',
			CASE WHEN jsonb_typeof(sd.decision->'implementer') = 'string'
			     THEN sd.decision->>'implementer'
			     ELSE sd.decision->'implementer'->>'model' END,
			CASE WHEN jsonb_typeof(sd.decision->'implementer') = 'object'
			     THEN COALESCE(sd.decision->'implementer'->>'effort', '')
			     ELSE '' END,
			CASE WHEN jsonb_typeof(sd.decision->'fixer') = 'string'
			     THEN sd.decision->>'fixer'
			     WHEN jsonb_typeof(sd.decision->'fixer') = 'object'
			     THEN COALESCE(sd.decision->'fixer'->>'model', '') || '/' || COALESCE(sd.decision->'fixer'->>'effort', '')
			     ELSE '' END,
			CASE WHEN pm.panel_is_object AND jsonb_typeof(sd.decision->'panel'->'rerun_dispatch') = 'object'
			     THEN COALESCE(sd.decision->'panel'->'rerun_dispatch'->>'model', '') || '/' || COALESCE(sd.decision->'panel'->'rerun_dispatch'->>'effort', '')
			     ELSE '' END,
			CASE WHEN pm.panel_is_object
			     THEN COALESCE(jsonb_array_length(sd.decision->'panel'->'roster'), 0)
			     ELSE 0 END,
			CASE WHEN pm.panel_is_object
			     THEN COALESCE((sd.decision->'panel'->>'compact')::boolean, false)
			     ELSE false END,
			CASE WHEN pm.panel_is_object
			     THEN COALESCE((
					SELECT r->>'slot' FROM jsonb_array_elements(sd.decision->'panel'->'roster') r
					WHERE r->>'slot' LIKE 'exp-%'
					LIMIT 1
				), '')
			     ELSE '' END,
			CASE WHEN pm.panel_is_object
			     THEN COALESCE(sd.decision->'panel'->>'rerun', '')
			     ELSE '' END,
			CASE WHEN pm.panel_is_object
			     THEN COALESCE(sd.decision->'panel'->>'grouping', 'default')
			     ELSE 'default' END,
			COALESCE((SELECT fg.flattened FROM flattened_groups fg WHERE fg.decision_id = sd.id AND fg.label = 'dispatches'), ''),
			COALESCE((SELECT fg.flattened FROM flattened_groups fg WHERE fg.decision_id = sd.id AND fg.label = 'groups'), ''),
			COALESCE(ra.wall_clock_seconds, 0),
			COALESCE(da.input_tokens, 0)::bigint,
			COALESCE(da.output_tokens, 0)::bigint,
			COALESCE(da.cache_read_tokens, 0)::bigint,
			COALESCE(ra.cost_usd, 0),
			COALESCE(fa.critical, 0),
			COALESCE(fa.important, 0),
			COALESCE(fa.minor, 0),
			COALESCE(fa.fix_rounds, 0),
			COALESCE(da.fallbacks, 0),
			COALESCE(da.timed_out, 0)
		FROM scoped_decisions sd
		JOIN panel_meta pm ON pm.id = sd.id
		LEFT JOIN dispatch_agg da ON da.change_id = sd.change_id AND da.session_token = sd.session_token
		LEFT JOIN finding_agg fa ON fa.change_id = sd.change_id AND fa.session_token = sd.session_token
		LEFT JOIN runtime_agg ra ON ra.change_id = sd.change_id AND ra.session_token = sd.session_token
		ORDER BY sd.recorded_at DESC
	`, period.From, period.To, project)
	if err != nil {
		return nil, fmt.Errorf("store: decisions: %w", err)
	}
	defer rows.Close()

	var out []DecisionRow
	for rows.Next() {
		var row DecisionRow
		if err := rows.Scan(
			&row.Project, &row.Change, &row.RecordedAt,
			&row.Class, &row.Overridden, &row.Execution,
			&row.ImplementerModel, &row.ImplementerEffort,
			&row.Fixer, &row.RerunDispatch,
			&row.RosterSize, &row.Compact, &row.ExperimentalSlot, &row.Rerun,
			&row.Grouping, &row.Dispatches, &row.ImplementerGroups,
			&row.WallClockSeconds,
			&row.InputTokens, &row.OutputTokens, &row.CacheReadTokens, &row.CostUsd,
			&row.Critical, &row.Important, &row.Minor, &row.FixRounds,
			&row.Fallbacks, &row.TimedOut,
		); err != nil {
			return nil, fmt.Errorf("store: decisions: scan: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: decisions: %w", err)
	}
	return out, nil
}

// AllRecordedRunsUnmeasured reports the third arm of the absence
// distinction (design.md, "the third arm of the absence distinction"):
// whether stage runs exist for period (and, when project is non-nil,
// belong to that project) and *none* of them carry a measurement. It is
// deliberately not "is the period empty" -- a period with zero runs is a
// different, already-handled case (this store's own empty-slice returns
// from every aggregation method, and (*statsHandler).recorded in
// internal/api) -- this method only ever contributes a true when it has
// first confirmed at least one run exists; a period with no runs at all
// returns false here, same as a period where every run was measured.
//
// "Measured" mirrors CostPerChangeRow.MeasuredRuns' own convention: a run
// counts as measured when its metrics bag carries a "tokens" key at all,
// regardless of that key's value -- the same absence-is-not-a-value rule
// this file applies everywhere else. A run priced at a real, measured
// zero still has "tokens", so it is never mistaken here for one that was
// never attributed.
func (s *Store) AllRecordedRunsUnmeasured(ctx context.Context, period Period, project *string) (bool, error) {
	var total, measured int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE sr.metrics ? 'tokens')
		FROM stage_runs sr
		JOIN changes c ON c.id = sr.change_id
		WHERE sr.started_at >= $1 AND sr.started_at < $2
		  AND ($3::text IS NULL OR c.project_key = $3)
	`, period.From, period.To, project).Scan(&total, &measured)
	if err != nil {
		return false, fmt.Errorf("store: all recorded runs unmeasured: %w", err)
	}
	return total > 0 && measured == 0, nil
}
