package store

// experimentstats.go — the before/after figures of an allowed-pairs
// experiment (KAN-1067, decisions role-metric-families and
// main-check-and-tag).

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"
)

// ErrExperimentNotFound is returned by ExperimentStats for an unknown id.
var ErrExperimentNotFound = errors.New("store: experiment not found")

// ExperimentWeek is one week of a metric's series: the mean over the
// changes whose first stage run started in the ISO week opening at Week.
type ExperimentWeek struct {
	Week  time.Time `json:"week"`
	Value float64   `json:"value"`
}

// ExperimentMetric is one metric's mean before and after the experiment's
// start -- nil where no counted change carries a figure -- and its weekly
// series over both windows.
type ExperimentMetric struct {
	Name   string           `json:"name"`
	Before *float64         `json:"before"`
	After  *float64         `json:"after"`
	Series []ExperimentWeek `json:"series"`
}

// ExperimentRuns counts the changes in each window.
type ExperimentRuns struct {
	Before int `json:"before"`
	After  int `json:"after"`
}

// ExperimentStatsResult is ExperimentStats' answer. The after window is
// [Experiment.StartedAt, End), End being closed_at or now; the before
// window [BeforeFrom, StartedAt) has the same length. Excluded counts the
// after-window changes a main cell left out because their harvested main
// thread did not run its pair.
type ExperimentStatsResult struct {
	Experiment Experiment         `json:"experiment"`
	BeforeFrom time.Time          `json:"beforeFrom"`
	End        time.Time          `json:"end"`
	Runs       ExperimentRuns     `json:"runs"`
	Excluded   int                `json:"excluded"`
	Metrics    []ExperimentMetric `json:"metrics"`
}

// Metric families per role (decision role-metric-families). Every metric
// is lower-is-better.
var (
	plannerMetrics  = []string{"reviewRounds", "round0CritImp", "planCorrections", "tokens", "wallClockMin"}
	fixerMetrics    = []string{"reviewRounds", "findingsPerTask", "verifyReds", "tokens", "wallClockMin"}
	reviewerMetrics = []string{"findingsPerDispatch", "lateRoundFindings", "reviewRounds", "tokens", "wallClockMin"}
	visualMetrics   = []string{"vvMismatches", "vvReruns", "tokens", "wallClockMin"}
	mainMetrics     = append(append([]string{}, plannerMetrics...), "verifyReds")
	otherMetrics    = []string{"tokens", "wallClockMin"}
)

func roleMetrics(role string) []string {
	switch role {
	case "main":
		return mainMetrics
	case "planner", "plan-designer", "explorer", "self-review-plan", "pipeline-fix-plan":
		return plannerMetrics
	case "implementer", "panel-fix", "self-review-fix", "pipeline-fix-fix", "late-fix":
		return fixerMetrics
	case "panel-reviewer", "panel-extra", "panel-rerun", "gated-reviewer", "gated-re-reviewer",
		"self-review-review-1", "self-review-review-r", "pipeline-fix-review-1", "pipeline-fix-review-r":
		return reviewerMetrics
	case "visual-verifier":
		return visualMetrics
	}
	return otherMetrics
}

// roleDispatchSQL selects, over dispatch row d, the rows an Allowed pairs
// role ran as. Dispatch rows record the ledger role and the key, not the
// table role, so each is derived from the dispatch-site table in
// skills/flow/implement.md and the key shapes the skills record.
// ponytail: derived from key shapes; record the table role on dispatch rows
// if a shape ever stops telling two roles apart. late-fix and
// self-review-pass have no key shape of their own and select nothing.
var roleDispatchSQL = map[string]string{
	"planner":               `d.role = 'planner' AND d.dispatch_key ~ '^plan-[0-9]+$'`,
	"plan-designer":         `d.role = 'planner' AND d.dispatch_key ~ '^design-[0-9]+'`,
	"explorer":              `d.role = 'explorer'`,
	"tooling-analyst":       `d.role = 'planner' AND d.dispatch_key LIKE 'tooling-analysis-%'`,
	"implementer":           `d.role = 'implementer' AND d.dispatch_key LIKE 'task-%'`,
	"panel-fix":             `d.role = 'panel-fix'`,
	"panel-reviewer":        `d.role = 'reviewer' AND d.dispatch_key ~ '^panel-0-' AND d.dispatch_key !~ 'test-sanity\+call-efficiency'`,
	"panel-extra":           `d.role = 'reviewer' AND d.dispatch_key ~ '^panel-[0-9]+-test-sanity\+call-efficiency'`,
	"panel-rerun":           `d.role = 'reviewer' AND d.dispatch_key ~ '^panel-[1-9][0-9]*-' AND d.dispatch_key !~ 'test-sanity\+call-efficiency'`,
	"panel-merger":          `d.role = 'merger'`,
	"gated-reviewer":        `d.role = 'reviewer' AND d.dispatch_key ~ '^task-[0-9+]+-reviewer(-retry)?$'`,
	"gated-re-reviewer":     `d.role = 'reviewer' AND d.dispatch_key ~ '^task-[0-9+]+-reviewer-fix-'`,
	"visual-verifier":       `d.role = 'verifier' AND d.dispatch_key LIKE 'visual-verify%'`,
	"self-review-plan":      `d.dispatch_key ~ '^self-review-.*-plan(-[0-9]+)?$'`,
	"self-review-fix":       `d.role = 'implementer' AND d.dispatch_key ~ '^self-review-.*-fix(-[gv]?[0-9]+)?$'`,
	"self-review-review-1":  `d.dispatch_key ~ '^self-review-.*-review-1$'`,
	"self-review-review-r":  `d.dispatch_key ~ '^self-review-.*-review-([2-9]|[1-9][0-9]+)$'`,
	"pipeline-fix-plan":     `d.dispatch_key ~ '^pipeline-fix-[0-9]+-plan'`,
	"pipeline-fix-fix":      `d.role = 'implementer' AND d.dispatch_key ~ '^pipeline-fix-[0-9]+-fix'`,
	"pipeline-fix-review-1": `d.dispatch_key ~ '^pipeline-fix-[0-9]+-review-1$'`,
	"pipeline-fix-review-r": `d.dispatch_key ~ '^pipeline-fix-[0-9]+-review-([2-9]|[1-9][0-9]+)$'`,
}

// tokensSQL sums a {main, sidechain} token bag's input, output and cache
// creation -- cache reads, re-reads of context already paid for, left out.
func tokensSQL(bag string) string {
	var parts []string
	for _, side := range []string{"main", "sidechain"} {
		for _, k := range []string{"input", "output", "cache_creation"} {
			parts = append(parts, fmt.Sprintf("COALESCE((%s->'%s'->>'%s')::numeric, 0)", bag, side, k))
		}
	}
	return strings.Join(parts, " + ")
}

// mainTokensSQL is a stage run sr's main-thread tokens.
const mainTokensSQL = `COALESCE((sr.metrics->'tokens'->'main'->>'input')::numeric, 0)
	+ COALESCE((sr.metrics->'tokens'->'main'->>'output')::numeric, 0)
	+ COALESCE((sr.metrics->'tokens'->'main'->>'cache_creation')::numeric, 0)`

// ExperimentStats computes experiment id's metrics. A change belongs to
// the window its first stage run started in -- once, whatever its later
// runs -- and per-change metrics are means over the window's changes:
//
//   - reviewRounds: the change's highest findings.round, 0 with none.
//   - round0CritImp: its round-0 Critical and Important findings.
//   - planCorrections: its latest recorded plan-correction count; a change
//     with none recorded (or only pre-0055 NULL rows) is skipped.
//   - verifyReds: its dispatches keyed verify% ended blocked by
//     test-failure.
//   - lateRoundFindings: its findings of round 2 or later.
//   - vvMismatches: findings raised by its visual-verify dispatches.
//   - vvReruns: its visual-verify-fix-* dispatches.
//   - tokens, wallClockMin: summed over the experiment roles' dispatches
//     (roleDispatchSQL) and, with a main cell, the change's stage runs'
//     main thread; mean over the changes that have any.
//
// findingsPerTask (findings ÷ implementer task dispatches) and
// findingsPerDispatch (findings joined to the experiment's reviewer-role
// dispatches ÷ those dispatches) are pooled ratios. A main cell
// <model>@<effort> counts an after-window change only when it has a
// harvested stage run and every one ran its main thread on that model
// family (alias or full id, through model_family) at that effort.
func (s *Store) ExperimentStats(ctx context.Context, id int64) (ExperimentStatsResult, error) {
	one, err := s.listExperiments(ctx, false, &id)
	if err != nil {
		return ExperimentStatsResult{}, err
	}
	if len(one) == 0 {
		return ExperimentStatsResult{}, fmt.Errorf("%w: %d", ErrExperimentNotFound, id)
	}
	e := &one[0]
	end := time.Now().UTC()
	if e.ClosedAt != nil {
		end = *e.ClosedAt
	}
	out := ExperimentStatsResult{Experiment: *e, End: end, BeforeFrom: e.StartedAt.Add(-end.Sub(e.StartedAt))}

	var names []string
	seen := map[string]bool{}
	var roleSQL, reviewerSQL []string
	for _, role := range sortedRoles(e.Cells) {
		for _, m := range roleMetrics(role) {
			if !seen[m] {
				seen[m] = true
				names = append(names, m)
			}
		}
		if sql, ok := roleDispatchSQL[role]; ok {
			roleSQL = append(roleSQL, "("+sql+")")
			if slices.Contains(roleMetrics(role), "findingsPerDispatch") {
				reviewerSQL = append(reviewerSQL, "("+sql+")")
			}
		}
	}
	or := func(preds []string) string {
		if len(preds) == 0 {
			return "FALSE"
		}
		return strings.Join(preds, " OR ")
	}
	var mainModel, mainEffort *string
	if cell, ok := e.Cells["main"]; ok && len(cell) == 1 {
		m, eff, _ := strings.Cut(cell[0], "@")
		mainModel, mainEffort = &m, &eff
	}

	type sums struct{ num, den float64 }
	type window struct {
		vals map[string]*sums
	}
	newWindow := func() *window { return &window{vals: map[string]*sums{}} }
	before, after := newWindow(), newWindow()
	weeks := map[time.Time]*window{}
	add := func(w *window, name string, num, den float64) {
		v := w.vals[name]
		if v == nil {
			v = &sums{}
			w.vals[name] = v
		}
		v.num, v.den = v.num+num, v.den+den
	}

	err = s.scanRows(ctx, "experiment stats", `
		WITH cohort AS (
			SELECT change_id, min(started_at) AS first_at FROM stage_runs
			WHERE change_id IS NOT NULL
			GROUP BY change_id
			HAVING min(started_at) >= $1 AND min(started_at) < $2
		)
		SELECT co.first_at, date_trunc('week', co.first_at AT TIME ZONE 'UTC'),
			(SELECT COALESCE(max(f.round), 0) FROM findings f WHERE f.change_id = co.change_id),
			(SELECT count(*) FROM findings f WHERE f.change_id = co.change_id AND f.round = 0
				AND canonical_severity(f.severity) IN ('critical', 'important')),
			(SELECT tc.corrections FROM change_task_counts tc WHERE tc.change_id = co.change_id
				AND tc.corrections IS NOT NULL ORDER BY tc.observed_at DESC, tc.id DESC LIMIT 1),
			(SELECT count(*) FROM dispatches d WHERE d.change_id = co.change_id AND d.dispatch_key LIKE 'verify%'
				AND d.outcome = 'blocked' AND d.cause = 'test-failure'),
			(SELECT count(*) FROM findings f WHERE f.change_id = co.change_id AND f.round >= 2),
			(SELECT count(*) FROM findings f JOIN dispatches d ON d.id = f.dispatch_id
				WHERE f.change_id = co.change_id AND d.dispatch_key LIKE 'visual-verify%'),
			(SELECT count(*) FROM dispatches d WHERE d.change_id = co.change_id AND d.dispatch_key LIKE 'visual-verify-fix-%'),
			(SELECT count(*) FROM findings f WHERE f.change_id = co.change_id),
			(SELECT count(*) FROM dispatches d WHERE d.change_id = co.change_id AND d.role = 'implementer' AND d.dispatch_key LIKE 'task-%'),
			(SELECT count(*) FROM dispatches d WHERE d.change_id = co.change_id AND (`+or(reviewerSQL)+`)),
			(SELECT count(*) FROM findings f JOIN dispatches d ON d.id = f.dispatch_id
				WHERE f.change_id = co.change_id AND (`+or(reviewerSQL)+`)),
			rd.n, COALESCE(rd.tokens, 0)::float8, COALESCE(rd.secs, 0)::float8,
			sr.n, COALESCE(sr.tokens, 0)::float8, COALESCE(sr.secs, 0)::float8, COALESCE(sr.matches, false)
		FROM cohort co,
		LATERAL (
			SELECT count(*) AS n, sum(`+tokensSQL("d.metrics->'tokens'")+`) AS tokens,
				sum(EXTRACT(EPOCH FROM d.ended_at - d.started_at)) AS secs
			FROM dispatches d WHERE d.change_id = co.change_id AND (`+or(roleSQL)+`)
		) rd,
		LATERAL (
			SELECT count(*) AS n, sum(`+mainTokensSQL+`) AS tokens,
				sum(EXTRACT(EPOCH FROM sr.ended_at - sr.started_at)) AS secs,
				bool_or(sr.metrics ? 'models') AND bool_and(NOT (sr.metrics ? 'models') OR (
					COALESCE(sr.metrics->>'effort' = $4, false)
					AND jsonb_typeof(sr.metrics->'models') = 'object'
					AND EXISTS (SELECT 1 FROM jsonb_each(sr.metrics->'models') m(k, v)
						WHERE COALESCE((v->'tokens'->'main'->>'output')::numeric, 0) > 0 AND model_family(k) = $3)
					AND NOT EXISTS (SELECT 1 FROM jsonb_each(sr.metrics->'models') m(k, v)
						WHERE COALESCE((v->'tokens'->'main'->>'output')::numeric, 0) > 0 AND model_family(k) <> $3))) AS matches
			FROM stage_runs sr WHERE sr.change_id = co.change_id
		) sr
	`, []any{out.BeforeFrom, end, mainModel, mainEffort}, func(scan func(...any) error) error {
		var firstAt, week time.Time
		var corrections *int
		var rounds, critImp, verifyReds, late, vvMis, vvRe, findings, implN, revN, revF, roleN, stageN int
		var roleTok, roleSecs, stageTok, stageSecs float64
		var matches bool
		if err := scan(&firstAt, &week, &rounds, &critImp, &corrections, &verifyReds, &late, &vvMis, &vvRe,
			&findings, &implN, &revN, &revF, &roleN, &roleTok, &roleSecs, &stageN, &stageTok, &stageSecs, &matches); err != nil {
			return err
		}
		w := before
		if !firstAt.Before(e.StartedAt) {
			if mainModel != nil && !matches {
				out.Excluded++
				return nil
			}
			w = after
			out.Runs.After++
		} else {
			out.Runs.Before++
		}
		tokens, secs, has := roleTok, roleSecs, roleN > 0
		if mainModel != nil {
			tokens, secs, has = tokens+stageTok, secs+stageSecs, has || stageN > 0
		}
		perChange := map[string][2]float64{
			"reviewRounds": {float64(rounds), 1}, "round0CritImp": {float64(critImp), 1},
			"verifyReds":        {float64(verifyReds), 1},
			"lateRoundFindings": {float64(late), 1}, "vvMismatches": {float64(vvMis), 1}, "vvReruns": {float64(vvRe), 1},
			"findingsPerTask": {float64(findings), float64(implN)}, "findingsPerDispatch": {float64(revF), float64(revN)},
		}
		if corrections != nil {
			perChange["planCorrections"] = [2]float64{float64(*corrections), 1}
		}
		if has {
			perChange["tokens"] = [2]float64{tokens, 1}
			perChange["wallClockMin"] = [2]float64{secs / 60, 1}
		}
		week = week.UTC()
		if weeks[week] == nil {
			weeks[week] = newWindow()
		}
		for _, name := range names {
			if v, ok := perChange[name]; ok {
				add(w, name, v[0], v[1])
				add(weeks[week], name, v[0], v[1])
			}
		}
		return nil
	})
	if err != nil {
		return ExperimentStatsResult{}, err
	}

	mean := func(w *window, name string) *float64 {
		v := w.vals[name]
		if v == nil || v.den == 0 {
			return nil
		}
		m := v.num / v.den
		return &m
	}
	order := make([]time.Time, 0, len(weeks))
	for w := range weeks {
		order = append(order, w)
	}
	sort.Slice(order, func(i, j int) bool { return order[i].Before(order[j]) })
	for _, name := range names {
		m := ExperimentMetric{Name: name, Before: mean(before, name), After: mean(after, name), Series: []ExperimentWeek{}}
		for _, w := range order {
			if v := mean(weeks[w], name); v != nil {
				m.Series = append(m.Series, ExperimentWeek{Week: w, Value: *v})
			}
		}
		out.Metrics = append(out.Metrics, m)
	}
	return out, nil
}
