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
// flowd's guard-runs ingestion drops the refusal rather than recording it.
var ErrProjectNotFound = errors.New("store: project not found")

// ErrInvalidGuardRun is returned by RecordGuardRun for a run outside what
// 0032_guard_runs.sql accepts -- an empty guard or worktree, an outcome
// outside records.GuardRunOutcomes, a negative duration, or an empty
// reason id (0064_guard_run_reasons.sql) -- checked here
// so the caller gets a 400 naming the field rather than a constraint
// violation.
var ErrInvalidGuardRun = errors.New("store: invalid guard run")

// RecordGuardRun records one flow-guard invocation against projectKey.
// Every call inserts a new row, as RecordVerdict does. A key the store does
// not hold falls back to the project of the newest change whose worktrees map
// names the run's worktree, a directory under it: a
// guard run from a peer repository's worktree keys to that repository, which
// is no project, and every capture and verifier run of gymie-playwright was
// dropped. Any other unknown project is
// ErrProjectNotFound: the INSERT ... SELECT inserts nothing, rather than
// tripping the foreign key, so the refusal is a sentinel and not a 500.
func (s *Store) RecordGuardRun(ctx context.Context, projectKey string, in records.GuardRun) (records.GuardRun, error) {
	return recordGuardRun(ctx, s.pool, projectKey, in)
}

// RecordGuardRuns records runs[i] against projectKeys[i], as RecordGuardRun
// does each, in one transaction: a run RecordGuardRun refuses with
// ErrProjectNotFound or ErrInvalidGuardRun is skipped, and any other failure
// records none of them, so the caller can retry the whole batch without a
// duplicate row. flowd's guard-runs ingestion (cmd/flowd/hookruns.go) is its
// caller.
func (s *Store) RecordGuardRuns(ctx context.Context, projectKeys []string, runs []records.GuardRun) error {
	if len(runs) == 0 {
		return nil
	}
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		for i, in := range runs {
			_, err := recordGuardRun(ctx, tx, projectKeys[i], in)
			if err != nil && !errors.Is(err, ErrProjectNotFound) && !errors.Is(err, ErrInvalidGuardRun) {
				return err
			}
		}
		return nil
	})
}

// rowQuerier is satisfied by both *pgxpool.Pool and pgx.Tx.
type rowQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func recordGuardRun(ctx context.Context, q rowQuerier, projectKey string, in records.GuardRun) (records.GuardRun, error) {
	if in.Guard == "" || in.Worktree == "" {
		return records.GuardRun{}, fmt.Errorf("%w: guard and worktree are both required", ErrInvalidGuardRun)
	}
	if !slices.Contains(records.GuardRunOutcomes, in.Outcome) {
		return records.GuardRun{}, fmt.Errorf("%w: outcome %q is not one of %v", ErrInvalidGuardRun, in.Outcome, records.GuardRunOutcomes)
	}
	if in.DurationMS < 0 {
		return records.GuardRun{}, fmt.Errorf("%w: durationMs must not be negative", ErrInvalidGuardRun)
	}
	if in.WaitMS < 0 || in.WaitMS > in.DurationMS {
		return records.GuardRun{}, fmt.Errorf("%w: waitMs must lie between 0 and durationMs", ErrInvalidGuardRun)
	}
	if in.BootstrapMS != nil && *in.BootstrapMS < 0 {
		return records.GuardRun{}, fmt.Errorf("%w: bootstrapMs must not be negative", ErrInvalidGuardRun)
	}
	if slices.Contains(in.Reasons, "") {
		return records.GuardRun{}, fmt.Errorf("%w: reasons must not hold an empty id", ErrInvalidGuardRun)
	}
	// An absent or empty list is stored NULL, never '{}' (decision reason-nullable).
	var reasons []string
	if len(in.Reasons) > 0 {
		reasons = in.Reasons
	}
	recordedAt := in.RecordedAt
	if recordedAt.IsZero() {
		recordedAt = time.Now()
	}

	var hl struct {
		role, model, effort *string
		cost                *float64
		ms                  *int
	}
	if h := in.Headless; h != nil {
		hl.role, hl.model, hl.effort, hl.cost, hl.ms = &h.Role, &h.Model, &h.Effort, h.CostUSD, &h.DurationMS
	}
	out := in
	err := q.QueryRow(ctx, `
		INSERT INTO guard_runs (project_key, guard, worktree, exit_code, outcome, duration_ms, recorded_at, reasons,
			headless_role, headless_model, headless_effort, headless_cost_usd, headless_ms, wait_ms, bootstrap_ms)
		SELECT p.project_key, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		FROM projects p
		WHERE p.project_key = COALESCE(
			(SELECT project_key FROM projects WHERE project_key = $1),
			(SELECT c.project_key FROM changes c, jsonb_object_keys(c.worktrees) k
				WHERE $3 = k OR starts_with($3, k || '/')
				ORDER BY c.id DESC LIMIT 1))
		RETURNING id, recorded_at, reasons
	`, projectKey, in.Guard, in.Worktree, in.ExitCode, in.Outcome, in.DurationMS, recordedAt, reasons,
		hl.role, hl.model, hl.effort, hl.cost, hl.ms, in.WaitMS, in.BootstrapMS).Scan(&out.ID, &out.RecordedAt, &out.Reasons)
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
	// Helper marks a records.GuardHelpers command: it refuses nothing, so its fire columns mean nothing.
	Helper bool
	// Parent names the guard that most often ran this one inside its own run (same worktree, the
	// run's interval inside the caller's), "" when it mostly ran alone. Derived, never stored.
	Parent string
}

// GuardHealth reports one row per guard that ran in period, optionally
// restricted to one project, ordered by guard. records.GuardHelpers are
// marked Helper; records.RetiredGuards are left out.
func (s *Store) GuardHealth(ctx context.Context, period Period, project *string) ([]GuardHealthRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT guard,
			COUNT(*)::int,
			(COUNT(*) FILTER (WHERE outcome = 'fired'))::int,
			(COUNT(*) FILTER (WHERE outcome = 'cannot-answer'))::int,
			ROUND(percentile_cont(0.5) WITHIN GROUP (ORDER BY duration_ms))::bigint,
			SUM(duration_ms)::bigint,
			MAX(recorded_at) FILTER (WHERE outcome = 'fired'),
			guard = ANY($5::text[])
		FROM guard_runs
		WHERE recorded_at >= $1 AND recorded_at < $2
		  AND ($3::text IS NULL OR project_key = $3)
		  AND guard <> ALL($4::text[])
		GROUP BY guard
		ORDER BY guard
	`, period.From, period.To, project, records.RetiredGuards, records.GuardHelpers)
	if err != nil {
		return nil, fmt.Errorf("store: guard health: %w", err)
	}
	defer rows.Close()

	var out []GuardHealthRow
	for rows.Next() {
		var row GuardHealthRow
		if err := rows.Scan(&row.Guard, &row.Runs, &row.Fired, &row.CannotAnswer,
			&row.MedianMs, &row.TotalMs, &row.LastFiredAt, &row.Helper); err != nil {
			return nil, fmt.Errorf("store: guard health: scan: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: guard health: %w", err)
	}
	rows.Close()

	runs, err := s.pool.Query(ctx, `
		SELECT guard, worktree, duration_ms, recorded_at FROM guard_runs
		WHERE recorded_at >= $1 AND recorded_at < $2 AND ($3::text IS NULL OR project_key = $3)
		ORDER BY worktree, recorded_at
	`, period.From, period.To, project)
	if err != nil {
		return nil, fmt.Errorf("store: guard health: parents: %w", err)
	}
	defer runs.Close()
	var all []guardSpan
	for runs.Next() {
		var g guardSpan
		if err := runs.Scan(&g.guard, &g.worktree, &g.ms, &g.end); err != nil {
			return nil, fmt.Errorf("store: guard health: parents: scan: %w", err)
		}
		g.start = g.end.Add(-time.Duration(g.ms) * time.Millisecond)
		all = append(all, g)
	}
	if err := runs.Err(); err != nil {
		return nil, fmt.Errorf("store: guard health: parents: %w", err)
	}
	byGuard := nestedParents(all)
	for i := range out {
		out[i].Parent = byGuard[out[i].Guard]
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
	`, period.From, period.To, hiddenGuards())
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

// GuardWindow is one change's run window ChangeGuardRuns reads.
type GuardWindow struct {
	ProjectKey, Change string
	From, To           time.Time
}

// ChangeGuards is the guard runs one GuardWindow holds.
type ChangeGuards struct {
	Claimed   []records.GuardRun
	Unclaimed int
}

// ChangeGuardRuns reads the guard runs each window holds, in two queries
// whatever the number of windows (decision parent-time-guard-runs).
// Claimed are every project's runs whose worktree carries the path segment
// -worktrees/<change> followed by "/" or the end -- any repository's
// worktree directory of the change -- and whose
// [recorded_at - duration_ms, recorded_at) overlaps [From, To), oldest
// first. Unclaimed counts ProjectKey's runs in the same window whose
// worktree carries no -worktrees/ segment: main-checkout runs no change can
// claim. The result is in ws' order.
func (s *Store) ChangeGuardRuns(ctx context.Context, ws []GuardWindow) ([]ChangeGuards, error) {
	const inWindow = `g.recorded_at > w.f AND g.recorded_at - g.duration_ms * interval '1 millisecond' < w.t`
	segs, keys := make([]string, len(ws)), make([]string, len(ws))
	froms, tos := make([]time.Time, len(ws)), make([]time.Time, len(ws))
	for i, w := range ws {
		segs[i], keys[i], froms[i], tos[i] = "-worktrees/"+w.Change, w.ProjectKey, w.From, w.To
	}
	out := make([]ChangeGuards, len(ws))
	rows, err := s.pool.Query(ctx, `
		SELECT w.i, g.id, g.guard, g.worktree, g.exit_code, g.outcome, g.duration_ms, g.recorded_at, g.reasons
		FROM unnest($1::text[], $2::timestamptz[], $3::timestamptz[]) WITH ORDINALITY AS w(seg, f, t, i)
		JOIN guard_runs g ON `+inWindow+`
		  AND (strpos(g.worktree, w.seg || '/') > 0 OR right(g.worktree, length(w.seg)) = w.seg)
		ORDER BY w.i, g.recorded_at, g.id
	`, segs, froms, tos)
	if err != nil {
		return nil, fmt.Errorf("store: change guard runs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			i int64
			r records.GuardRun
		)
		if err := rows.Scan(&i, &r.ID, &r.Guard, &r.Worktree, &r.ExitCode, &r.Outcome, &r.DurationMS, &r.RecordedAt, &r.Reasons); err != nil {
			return nil, fmt.Errorf("store: change guard runs: scan: %w", err)
		}
		out[i-1].Claimed = append(out[i-1].Claimed, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: change guard runs: %w", err)
	}
	counts, err := s.pool.Query(ctx, `
		SELECT w.i, COUNT(g.id)::int
		FROM unnest($1::text[], $2::timestamptz[], $3::timestamptz[]) WITH ORDINALITY AS w(pk, f, t, i)
		JOIN guard_runs g ON `+inWindow+` AND g.project_key = w.pk AND strpos(g.worktree, '-worktrees/') = 0
		GROUP BY w.i
	`, keys, froms, tos)
	if err != nil {
		return nil, fmt.Errorf("store: unclaimed guard runs: %w", err)
	}
	defer counts.Close()
	for counts.Next() {
		var i int64
		var n int
		if err := counts.Scan(&i, &n); err != nil {
			return nil, fmt.Errorf("store: unclaimed guard runs: scan: %w", err)
		}
		out[i-1].Unclaimed = n
	}
	if err := counts.Err(); err != nil {
		return nil, fmt.Errorf("store: unclaimed guard runs: %w", err)
	}
	return out, nil
}

// hiddenGuards is every name the guard-health views leave out: the helpers and the retired guards.
func hiddenGuards() []string {
	return append(slices.Clone(records.GuardHelpers), records.RetiredGuards...)
}

// guardSpan is one guard run as an interval: it ended at end and began ms earlier.
type guardSpan struct {
	guard, worktree string
	ms              int
	start, end      time.Time
}

// maxGuardNesting bounds how long after a run ends its caller can end: no guard runs longer.
const maxGuardNesting = 2 * time.Hour

// nestedParents maps each guard to the caller whose run most often held its runs (the narrowest
// holder per run), for guards whose runs were mostly held by one. spans are ordered by worktree,
// then end.
func nestedParents(spans []guardSpan) map[string]string {
	votes := map[string]map[string]int{}
	total := map[string]int{}
	held := map[string]int{}
	for i, c := range spans {
		total[c.guard]++
		best := -1
		for j := i - 1; j >= 0 && spans[j].worktree == c.worktree && spans[j].end.Equal(c.end); j-- {
			best = holder(spans, best, j, c)
		}
		for j := i + 1; j < len(spans) && spans[j].worktree == c.worktree && spans[j].end.Sub(c.end) <= maxGuardNesting; j++ {
			best = holder(spans, best, j, c)
		}
		if best < 0 {
			continue
		}
		held[c.guard]++
		if votes[c.guard] == nil {
			votes[c.guard] = map[string]int{}
		}
		votes[c.guard][spans[best].guard]++
	}
	out := map[string]string{}
	for g, v := range votes {
		if held[g]*2 <= total[g] {
			continue
		}
		top, n := "", 0
		for p, k := range v {
			if k > n || (k == n && p < top) {
				top, n = p, k
			}
		}
		out[g] = top
	}
	return out
}

// holder returns j when span j holds c (another guard, longer, c's interval inside it) and is
// narrower than the current best, else best.
func holder(spans []guardSpan, best, j int, c guardSpan) int {
	p := spans[j]
	if p.guard == c.guard || p.ms <= c.ms || p.start.After(c.start) || p.end.Before(c.end) {
		return best
	}
	if best < 0 || p.ms < spans[best].ms {
		return j
	}
	return best
}
