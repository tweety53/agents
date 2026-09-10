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
// it was dispatched, across how many changes, what it found by severity,
// and how much of what it found was deferred or withdrawn rather than
// fixed. Experimental is true for a slot whose id carries the "exp-"
// prefix (design.md's exp-slot-prefix decision -- no schema flag, the id
// itself is the fact); Description is that slot's newest recorded
// decision roster entry naming it, empty when no decision ever named it
// (a default-panel run, or a slot whose roster entry carries no
// description at all).
type ReviewerRow struct {
	Slot         string
	Experimental bool
	Description  string

	Dispatches int
	Changes    int
	Critical   int
	Important  int
	Minor      int

	FindingsPerDispatch float64
	DeferredShare       float64
	WithdrawnShare      float64
}

// Reviewers reports one row per role -- design.md's "Stats views ›
// reviewers" -- for every dispatch whose slot is set and whose stage run
// started within period (and, when project/model is non-nil, matches it).
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
// decision->'panel'->'roster' carries an entry naming this slot
// (jsonb_path_query_first, per design.md), independent of period and
// project -- a roster entry describes what a slot IS, not when it ran, so
// scoping it to the query's own period would make the same slot's
// description flicker between periods for no reason tied to the roster
// itself.
func (s *Store) Reviewers(ctx context.Context, period Period, project, model *string) ([]ReviewerRow, error) {
	rows, err := s.pool.Query(ctx, `
		WITH scoped_dispatches AS (
			SELECT d.id, d.change_id, unnest(string_to_array(d.slot, '+')) AS slot
			FROM dispatches d
			JOIN stage_runs sr ON sr.id = d.stage_run_id
			JOIN changes c ON c.id = d.change_id
			WHERE sr.started_at >= $1 AND sr.started_at < $2
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
				COUNT(*) FILTER (WHERE severity ILIKE 'minor') AS minor,
				COUNT(*) AS total,
				COUNT(*) FILTER (WHERE status ILIKE 'deferred%') AS deferred,
				COUNT(*) FILTER (WHERE status ILIKE 'withdrawn%') AS withdrawn
			FROM scoped_findings
			GROUP BY slot
		)
		SELECT
			da.slot,
			da.slot LIKE 'exp-%' AS experimental,
			COALESCE((
				SELECT jsonb_path_query_first(
					d2.decision, '$.panel.roster[*] ? (@.slot == $slot)',
					jsonb_build_object('slot', da.slot), true
				) ->> 'description'
				FROM decisions d2
				WHERE jsonb_path_exists(
					d2.decision, '$.panel.roster[*] ? (@.slot == $slot)',
					jsonb_build_object('slot', da.slot), true
				)
				ORDER BY d2.recorded_at DESC
				LIMIT 1
			), '') AS description,
			da.dispatches,
			da.changes,
			COALESCE(fa.critical, 0),
			COALESCE(fa.important, 0),
			COALESCE(fa.minor, 0),
			CASE WHEN da.dispatches > 0 THEN COALESCE(fa.total, 0)::float8 / da.dispatches ELSE 0 END,
			CASE WHEN COALESCE(fa.total, 0) > 0 THEN fa.deferred::float8 / fa.total ELSE 0 END,
			CASE WHEN COALESCE(fa.total, 0) > 0 THEN fa.withdrawn::float8 / fa.total ELSE 0 END
		FROM dispatch_agg da
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
			&row.FindingsPerDispatch, &row.DeferredShare, &row.WithdrawnShare,
		); err != nil {
			return nil, fmt.Errorf("store: reviewers: scan: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: reviewers: %w", err)
	}
	return out, nil
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
// is polymorphic: an object ({"model":...,"effort":...}) when the toggle
// was dynamic and execution was "sdd", or a bare string ("skipped —
// inline" / "default") otherwise -- ImplementerModel carries the model in
// the first case and the string itself in the second, so the column always
// has something to show; ImplementerEffort is empty in the second case,
// since a string carries no effort to read. RosterSize, Compact,
// ExperimentalSlot and Rerun read the same way from "panel", which is
// either an object or the bare string "default" -- the zero values (0,
// false, "", "") stand in for "default" rather than an error, since a
// default-panel run recorded no per-slot detail to report. Grouping is
// "panel.grouping" when panel is an object, else the literal "default".
// Dispatches renders "panel.dispatches" -- one to two arrays of slot ids --
// as each array's ids '+'-joined and the arrays themselves ' · '-joined, in
// roster order; empty when panel carries no dispatches to show.
// ImplementerGroups renders the decision's own top-level "groups" the same
// way over bundle ids; empty when "groups" is JSON null (inline execution)
// or absent.
type DecisionRow struct {
	Project    string
	Change     string
	RecordedAt time.Time

	Class      string
	Overridden bool
	Execution  string

	ImplementerModel  string
	ImplementerEffort string

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
// Token and cost figures sum dispatches.metrics' "main" and "sidechain"
// buckets the same way CostPerChange does; outcome counts (fallback,
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
			-- and top-level "groups" are both an array of string-arrays, so both feed
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
			-- Flatten one array of string-arrays into a "·"-joined string of
			-- "+"-joined groups, in array order -- the one shape both "dispatches"
			-- and "groups" render into.
			SELECT gs.decision_id, gs.label,
				(SELECT string_agg(joined, ' · ' ORDER BY grp.ord)
					FROM jsonb_array_elements(gs.arr) WITH ORDINALITY AS grp(val, ord)
					CROSS JOIN LATERAL (
						SELECT string_agg(e.elem, '+' ORDER BY e.eord) AS joined
						FROM jsonb_array_elements_text(grp.val) WITH ORDINALITY AS e(elem, eord)
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
				SUM(COALESCE((metrics->>'cost_usd')::numeric, 0)) AS cost_usd,
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
				EXTRACT(EPOCH FROM (MAX(sr.ended_at) - MIN(sr.started_at))) AS wall_clock_seconds
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
			COALESCE(da.cost_usd, 0),
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
