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
// state, and when it was last updated.
type LiveStateRow struct {
	ProjectKey string
	Name       string
	State      State
	UpdatedAt  time.Time
	UpdatedBy  string
}

// LiveStateBoard lists every change updated within period, optionally
// restricted to one project. A period containing no changes returns an
// empty slice, never an error.
func (s *Store) LiveStateBoard(ctx context.Context, period Period, project *string) ([]LiveStateRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT project_key, name, state, updated_at, updated_by
		FROM changes
		WHERE updated_at >= $1 AND updated_at < $2
		  AND ($3::text IS NULL OR project_key = $3)
		ORDER BY updated_at DESC, project_key, name
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
		if err := rows.Scan(&row.ProjectKey, &row.Name, &state, &row.UpdatedAt, &row.UpdatedBy); err != nil {
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
}

// StageLeaderboard ranks stages by cost -- mean, median and p90 -- across
// period, optionally restricted to one project.
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
			percentile_cont(0.9) WITHIN GROUP (ORDER BY cost)
		FROM costed
		GROUP BY command, stage
		ORDER BY command, stage
	`, period.From, period.To, project, model)
	if err != nil {
		return nil, fmt.Errorf("store: stage leaderboard: %w", err)
	}
	defer rows.Close()

	var out []StageLeaderboardRow
	for rows.Next() {
		var row StageLeaderboardRow
		if err := rows.Scan(&row.Command, &row.Stage, &row.RunCount, &row.MeanCostUSD, &row.MedianCostUSD, &row.P90CostUSD); err != nil {
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
// cheaper or more expensive as myflow itself changes.
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

// CacheEfficiencyRow is one command/stage pair's cache-read to
// cache-creation ratio across a period -- the largest single lever on
// cost. Ratio is nil when either total is unavailable (no run in the group
// recorded that token type) or cache creation totalled zero, rather than
// reporting a division by zero or an inferred ratio of zero.
type CacheEfficiencyRow struct {
	Command string
	Stage   string

	CacheReadTotal     *int64
	CacheCreationTotal *int64
	Ratio              *float64
}

// CacheEfficiency reports cache-read against cache-creation totals per
// command and stage across period, optionally restricted to one project.
//
// model, when non-nil, restricts to stage runs whose metrics recorded that
// model (per CostPerChange's own doc comment) and sums that model's own
// token bucket, never the run's total.
func (s *Store) CacheEfficiency(ctx context.Context, period Period, project, model *string) ([]CacheEfficiencyRow, error) {
	rows, err := s.pool.Query(ctx, `
		WITH scoped AS (
			SELECT
				sr.command, sr.stage,
				CASE WHEN $4::text IS NULL THEN sr.metrics->'tokens'
				     ELSE sr.metrics->'models'->$4::text->'tokens' END AS tok
			FROM stage_runs sr
			JOIN changes c ON c.id = sr.change_id
			WHERE sr.started_at >= $1 AND sr.started_at < $2
			  AND ($3::text IS NULL OR c.project_key = $3)
			  AND ($4::text IS NULL OR sr.metrics->'models' ? $4)
		)
		SELECT
			command, stage,
			SUM(
				CASE WHEN tok IS NOT NULL THEN
					COALESCE((tok->'main'->>'cache_read')::numeric, 0)
					+ COALESCE((tok->'sidechain'->>'cache_read')::numeric, 0)
				END
			),
			SUM(
				CASE WHEN tok IS NOT NULL THEN
					COALESCE((tok->'main'->>'cache_creation')::numeric, 0)
					+ COALESCE((tok->'sidechain'->>'cache_creation')::numeric, 0)
				END
			)
		FROM scoped
		GROUP BY command, stage
		ORDER BY command, stage
	`, period.From, period.To, project, model)
	if err != nil {
		return nil, fmt.Errorf("store: cache efficiency: %w", err)
	}
	defer rows.Close()

	var out []CacheEfficiencyRow
	for rows.Next() {
		var row CacheEfficiencyRow
		if err := rows.Scan(&row.Command, &row.Stage, &row.CacheReadTotal, &row.CacheCreationTotal); err != nil {
			return nil, fmt.Errorf("store: cache efficiency: scan: %w", err)
		}
		if row.CacheReadTotal != nil && row.CacheCreationTotal != nil && *row.CacheCreationTotal != 0 {
			ratio := float64(*row.CacheReadTotal) / float64(*row.CacheCreationTotal)
			row.Ratio = &ratio
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: cache efficiency: %w", err)
	}
	return out, nil
}

// PanelEconomicsRow is one review panel roster preset's findings-per-token
// yield across a period, over stage runs belonging to a change that
// recorded that roster.
type PanelEconomicsRow struct {
	ReviewPanelRoster string

	FindingsTotal   int64
	TokensTotal     *int64
	FindingsPerMTok *float64
}

// PanelEconomics reports findings produced per token spent, by review
// panel roster preset, across period and optionally one project -- whether
// a heavier roster earns its extra cost.
//
// model, when non-nil, restricts to stage runs whose metrics recorded that
// model (per CostPerChange's own doc comment) and sums that model's own
// token bucket, never the run's total. Findings are not recorded per
// model -- they are a whole-panel-run figure -- so a model filter narrows
// which runs contribute findings without changing how a contributing run's
// own findings are counted.
func (s *Store) PanelEconomics(ctx context.Context, period Period, project, model *string) ([]PanelEconomicsRow, error) {
	rows, err := s.pool.Query(ctx, `
		WITH scoped AS (
			SELECT
				sr.id,
				c.review_panel_roster AS roster,
				CASE WHEN $4::text IS NULL THEN sr.metrics->'tokens'
				     ELSE sr.metrics->'models'->$4::text->'tokens' END AS tok,
				COALESCE(sr.metrics->'findings_by_severity', '{}'::jsonb) AS findings
			FROM stage_runs sr
			JOIN changes c ON c.id = sr.change_id
			WHERE sr.started_at >= $1 AND sr.started_at < $2
			  AND ($3::text IS NULL OR c.project_key = $3)
			  AND c.review_panel_roster IS NOT NULL
			  AND ($4::text IS NULL OR sr.metrics->'models' ? $4)
		),
		runs AS (
			SELECT
				id, roster, findings,
				CASE WHEN tok IS NOT NULL THEN
					COALESCE((tok->'main'->>'input')::numeric, 0)
					+ COALESCE((tok->'sidechain'->>'input')::numeric, 0)
					+ COALESCE((tok->'main'->>'output')::numeric, 0)
					+ COALESCE((tok->'sidechain'->>'output')::numeric, 0)
					+ COALESCE((tok->'main'->>'cache_creation')::numeric, 0)
					+ COALESCE((tok->'sidechain'->>'cache_creation')::numeric, 0)
					+ COALESCE((tok->'main'->>'cache_read')::numeric, 0)
					+ COALESCE((tok->'sidechain'->>'cache_read')::numeric, 0)
				END AS total_tokens
			FROM scoped
		),
		findings AS (
			SELECT r.id, COALESCE(SUM((fs.value)::text::numeric), 0) AS findings
			FROM runs r
			LEFT JOIN LATERAL jsonb_each(r.findings) fs ON true
			GROUP BY r.id
		)
		SELECT r.roster, SUM(f.findings)::bigint, SUM(r.total_tokens)::bigint
		FROM runs r
		JOIN findings f ON f.id = r.id
		GROUP BY r.roster
		ORDER BY r.roster
	`, period.From, period.To, project, model)
	if err != nil {
		return nil, fmt.Errorf("store: panel economics: %w", err)
	}
	defer rows.Close()

	var out []PanelEconomicsRow
	for rows.Next() {
		var row PanelEconomicsRow
		var tokensTotal *int64
		if err := rows.Scan(&row.ReviewPanelRoster, &row.FindingsTotal, &tokensTotal); err != nil {
			return nil, fmt.Errorf("store: panel economics: scan: %w", err)
		}
		row.TokensTotal = tokensTotal
		if tokensTotal != nil && *tokensTotal != 0 {
			perMTok := float64(row.FindingsTotal) / (float64(*tokensTotal) / tokensPerMillion)
			row.FindingsPerMTok = &perMTok
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: panel economics: %w", err)
	}
	return out, nil
}

// ModelComparisonRow is one model's cost and rework rate for one
// command/stage pair across a period.
type ModelComparisonRow struct {
	Model   string
	Command string
	Stage   string

	RunCount       int
	MeanCostUSD    *float64
	ReworkAttempts int
}

// ModelComparison compares cost and subsequent rework for the same stage
// across models, restricted to stage runs whose metrics recorded at
// least one model, across period and optionally one project.
//
// This reads the metrics bag's "models" object (task 22), not the
// retired scalar "model" key: a stage run's model is last-write-wins per
// harvest batch under the old scheme, which is wrong for the common
// case rather than an edge case -- a review panel's own parent runs one
// model and its reviewer slots another, so a stage run recording a
// single model under-reports the very mix this view exists to compare.
// jsonb_each expands that object so one stage run carrying two models
// contributes two rows here, one per model, each grouped and averaged
// over that model's own cost_usd (written per model by Store.Price,
// pricing.go) -- never the whole run's cost charged twice.
//
// model, when non-nil, restricts the expanded rows to that one model --
// this view's own rows are already scoped to one model's own bucket by
// construction (m.value), so filtering here narrows *which* model's rows
// come back rather than changing what any one row reports, unlike every
// other aggregation in this file which must swap its source between the
// run's total and a model's bucket.
func (s *Store) ModelComparison(ctx context.Context, period Period, project, model *string) ([]ModelComparisonRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			m.key AS model, sr.command, sr.stage,
			COUNT(*),
			AVG((m.value->>'cost_usd')::numeric),
			COUNT(*) FILTER (WHERE sr.attempt > 1)
		FROM stage_runs sr
		JOIN changes c ON c.id = sr.change_id
		CROSS JOIN LATERAL jsonb_each(sr.metrics->'models') AS m(key, value)
		WHERE sr.started_at >= $1 AND sr.started_at < $2
		  AND ($3::text IS NULL OR c.project_key = $3)
		  AND sr.metrics ? 'models'
		  AND ($4::text IS NULL OR m.key = $4)
		GROUP BY m.key, sr.command, sr.stage
		ORDER BY model, sr.command, sr.stage
	`, period.From, period.To, project, model)
	if err != nil {
		return nil, fmt.Errorf("store: model comparison: %w", err)
	}
	defer rows.Close()

	var out []ModelComparisonRow
	for rows.Next() {
		var row ModelComparisonRow
		if err := rows.Scan(&row.Model, &row.Command, &row.Stage, &row.RunCount, &row.MeanCostUSD, &row.ReworkAttempts); err != nil {
			return nil, fmt.Errorf("store: model comparison: scan: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: model comparison: %w", err)
	}
	return out, nil
}

// ReworkRateRow is one command/stage pair's rework and abandonment rate
// across a period. TotalAttempts counts every recorded attempt, including
// attempt 1; ReworkAttempts counts attempt 2 and beyond -- a fix re-run
// recorded as its own attempt, per stage_runs' unique key, rather than
// inferred from timing.
type ReworkRateRow struct {
	Command string
	Stage   string

	TotalAttempts  int
	ReworkAttempts int
	AbandonedCount int
}

// ReworkRate reports how often a command re-runs as a fix, and how often a
// stage is abandoned, across period and optionally one project.
//
// model, when non-nil, restricts to stage runs whose metrics recorded that
// model (per CostPerChange's own doc comment). This view reports no cost
// or token figure, so there is no per-model bucket to swap the source of --
// the filter here is a pure WHERE restriction, unlike every view above
// that also reports a model-scoped number.
func (s *Store) ReworkRate(ctx context.Context, period Period, project, model *string) ([]ReworkRateRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			sr.command, sr.stage,
			COUNT(*),
			COUNT(*) FILTER (WHERE sr.attempt > 1),
			COUNT(*) FILTER (WHERE sr.outcome = 'abandoned')
		FROM stage_runs sr
		JOIN changes c ON c.id = sr.change_id
		WHERE sr.started_at >= $1 AND sr.started_at < $2
		  AND ($3::text IS NULL OR c.project_key = $3)
		  AND ($4::text IS NULL OR sr.metrics->'models' ? $4)
		GROUP BY sr.command, sr.stage
		ORDER BY sr.command, sr.stage
	`, period.From, period.To, project, model)
	if err != nil {
		return nil, fmt.Errorf("store: rework rate: %w", err)
	}
	defer rows.Close()

	var out []ReworkRateRow
	for rows.Next() {
		var row ReworkRateRow
		if err := rows.Scan(&row.Command, &row.Stage, &row.TotalAttempts, &row.ReworkAttempts, &row.AbandonedCount); err != nil {
			return nil, fmt.Errorf("store: rework rate: scan: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: rework rate: %w", err)
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
