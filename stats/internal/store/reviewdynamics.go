package store

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// ReviewDynamics is the Review dynamics view's whole answer for one period:
// the trend series, the slot × model · effort heatmap and the summary for
// the period and for the previous period of the same length
// (previous-range), the prompt changes inside the period as markers, and
// the newest prompt change whatever the period.
type ReviewDynamics struct {
	Bucket             string
	Series             []DynamicsPoint
	Heat, PreviousHeat []HeatCell
	Summary            DynamicsSummary
	PreviousSummary    DynamicsSummary
	Markers            []PromptChange
	LastPromptChangeAt *time.Time
}

// DynamicsPoint is one value of one series in one bucket. Basis is what
// the value is normalised by: changes for per-change metrics, dispatches
// for per-dispatch ones.
type DynamicsPoint struct {
	Bucket time.Time
	Metric string
	Series string
	Value  float64
	Basis  int
}

// HeatCell is one slot × model · effort cell of the heatmap: the reviewer
// dispatches carrying that slot and the findings joined to them by
// dispatch and slot, by canonical severity.
type HeatCell struct {
	Slot       string
	Model      string
	Effort     string
	Dispatches int
	Critical   int
	Important  int
	Minor      int
}

// DynamicsSummary is the verdict tiles' values over a whole period.
// ImplFindingsPerTask and VVFirstPassShare are nil when no change in the
// period had an implementer or a visual-verify dispatch.
type DynamicsSummary struct {
	FindingsPerChange          float64
	CriticalImportantPerChange float64
	MeanFixRounds              float64
	ImplFindingsPerTask        *float64
	VVFirstPassShare           *float64
	Changes                    int
}

// implBadOutcomes are the implementer dispatch outcomes counted as blocked
// (impl-vv-metrics).
var implBadOutcomes = []string{"blocked", "aborted", "failed", "stopped", "cancelled"}

// ReviewDynamics trends review-panel, implementer and visual-verify
// outcomes over period, bucketed by UTC day when the period is 14 days or
// shorter and by ISO week otherwise (bucket-from-range). project, when
// set, restricts every series to that project.
//
// Per-change metrics put a change in the bucket of its cohort dispatch's
// started_at and count every finding of the change (cohort-by-change);
// severities fold through canonical_severity:
//
//   - findings-per-change (critical, important, minor): findings ÷ changes,
//     cohort = first reviewer dispatch.
//   - fix-rounds (mean, max): a change's highest findings.round, 0 with no
//     finding.
//   - changes (changes): changes reviewed.
//   - impl-findings-per-task (<model · effort>): the cohort's findings ÷
//     its implementer dispatches, pooled; cohort = first implementer
//     dispatch, series = the change's most frequent implementer
//     model · effort, ties by name (impl-vv-metrics).
//   - vv-rounds (mean, max): the highest n of a visual-verify-fix-<n> key,
//     0 when none; cohort = first verifier dispatch keyed visual-verify%.
//   - vv-first-pass (first-pass): share of those changes with 0 rounds.
//
// Per-dispatch metrics bucket each reviewer dispatch by its own started_at
// and count the findings joined to it by dispatch and slot, unnesting a
// bundle's slots as Reviewers does:
//
//   - per-dispatch-slot (<slot>): findings ÷ dispatches carrying the slot.
//   - per-dispatch-model (<model>), per-dispatch-effort (<effort>):
//     findings ÷ dispatch rows, never unnested slots.
//   - impl-blocked-share (blocked): implementer dispatches started in the
//     bucket with an implBadOutcomes outcome ÷ all of them.
//
// The summary applies the same formulas to the whole period.
func (s *Store) ReviewDynamics(ctx context.Context, period Period, project *string) (ReviewDynamics, error) {
	unit := "week"
	if period.To.Sub(period.From) <= 14*24*time.Hour {
		unit = "day"
	}
	out := ReviewDynamics{Bucket: unit}

	cur, err := s.dynamicsFor(ctx, period, project, unit)
	if err != nil {
		return ReviewDynamics{}, err
	}
	prev, err := s.dynamicsFor(ctx, Period{From: period.From.Add(-period.To.Sub(period.From)), To: period.From}, project, unit)
	if err != nil {
		return ReviewDynamics{}, err
	}
	out.Series, out.Heat, out.Summary = cur.series, cur.heat, cur.summary
	out.PreviousHeat, out.PreviousSummary = prev.heat, prev.summary

	if out.Markers, err = s.ListPromptChanges(ctx, period.From, period.To); err != nil {
		return ReviewDynamics{}, err
	}
	if out.LastPromptChangeAt, err = s.LastPromptChangeAt(ctx); err != nil {
		return ReviewDynamics{}, err
	}
	return out, nil
}

type dynamicsPeriod struct {
	series  []DynamicsPoint
	heat    []HeatCell
	summary DynamicsSummary
}

func (s *Store) dynamicsFor(ctx context.Context, period Period, project *string, unit string) (dynamicsPeriod, error) {
	var out dynamicsPeriod
	add := func(b time.Time, metric, series string, value float64, basis int) {
		out.series = append(out.series, DynamicsPoint{Bucket: b, Metric: metric, Series: series, Value: value, Basis: basis})
	}
	ratio := func(n, d int) float64 {
		if d == 0 {
			return 0
		}
		return float64(n) / float64(d)
	}
	args := []any{period.From, period.To, project, unit}

	// Review panel, per change.
	var changes, findings, critImp, rounds int
	err := s.scanRows(ctx, "review per change", `
		WITH cohort AS (
			SELECT d.change_id, min(d.started_at) AS first_at
			FROM dispatches d JOIN changes c ON c.id = d.change_id
			WHERE d.role = 'reviewer' AND d.slot IS NOT NULL
			  AND ($3::text IS NULL OR c.project_key = $3)
			GROUP BY d.change_id
			HAVING min(d.started_at) >= $1 AND min(d.started_at) < $2
		),
		per_change AS (
			SELECT date_trunc($4::text, co.first_at AT TIME ZONE 'UTC') AS bucket,
				count(f.id) FILTER (WHERE canonical_severity(f.severity) = 'critical') AS critical,
				count(f.id) FILTER (WHERE canonical_severity(f.severity) = 'important') AS important,
				count(f.id) FILTER (WHERE canonical_severity(f.severity) = 'minor') AS minor,
				COALESCE(max(f.round), 0) AS rounds
			FROM cohort co LEFT JOIN findings f ON f.change_id = co.change_id
			GROUP BY co.change_id, co.first_at
		)
		SELECT bucket, count(*), sum(critical), sum(important), sum(minor), sum(rounds), max(rounds)
		FROM per_change GROUP BY bucket ORDER BY bucket
	`, args, func(scan func(...any) error) error {
		var b time.Time
		var n, cr, im, mi, sr, mr int
		if err := scan(&b, &n, &cr, &im, &mi, &sr, &mr); err != nil {
			return err
		}
		add(b, "findings-per-change", "critical", ratio(cr, n), n)
		add(b, "findings-per-change", "important", ratio(im, n), n)
		add(b, "findings-per-change", "minor", ratio(mi, n), n)
		add(b, "fix-rounds", "mean", ratio(sr, n), n)
		add(b, "fix-rounds", "max", float64(mr), n)
		add(b, "changes", "changes", float64(n), n)
		changes, findings, critImp, rounds = changes+n, findings+cr+im+mi, critImp+cr+im, rounds+sr
		return nil
	})
	if err != nil {
		return out, err
	}
	out.summary.Changes = changes
	out.summary.FindingsPerChange = ratio(findings, changes)
	out.summary.CriticalImportantPerChange = ratio(critImp, changes)
	out.summary.MeanFixRounds = ratio(rounds, changes)

	// Review panel, per dispatch: one row per (dispatch, slot).
	type group struct {
		dispatches map[int64]bool
		findings   int
	}
	type groupKey struct {
		bucket       time.Time
		metric, name string
	}
	groups := map[groupKey]*group{}
	heat := map[[3]string]*HeatCell{}
	bump := func(k groupKey, id int64, n int) {
		g := groups[k]
		if g == nil {
			g = &group{dispatches: map[int64]bool{}}
			groups[k] = g
		}
		g.dispatches[id] = true
		g.findings += n
	}
	err = s.scanRows(ctx, "review per dispatch", `
		WITH d AS (
			SELECT d.id, d.started_at, d.model, d.effort, unnest(string_to_array(d.slot, '+')) AS slot
			FROM dispatches d JOIN changes c ON c.id = d.change_id
			WHERE d.role = 'reviewer' AND d.slot IS NOT NULL
			  AND d.started_at >= $1 AND d.started_at < $2
			  AND ($3::text IS NULL OR c.project_key = $3)
		)
		SELECT d.id, date_trunc($4::text, d.started_at AT TIME ZONE 'UTC'), d.slot, d.model, d.effort,
			count(f.id) FILTER (WHERE canonical_severity(f.severity) = 'critical'),
			count(f.id) FILTER (WHERE canonical_severity(f.severity) = 'important'),
			count(f.id) FILTER (WHERE canonical_severity(f.severity) = 'minor')
		FROM d LEFT JOIN findings f ON f.dispatch_id = d.id AND f.slot = d.slot
		GROUP BY d.id, d.started_at, d.slot, d.model, d.effort
	`, args, func(scan func(...any) error) error {
		var id int64
		var b time.Time
		var slot, model, effort string
		var cr, im, mi int
		if err := scan(&id, &b, &slot, &model, &effort, &cr, &im, &mi); err != nil {
			return err
		}
		n := cr + im + mi
		bump(groupKey{b, "per-dispatch-slot", slot}, id, n)
		bump(groupKey{b, "per-dispatch-model", model}, id, n)
		bump(groupKey{b, "per-dispatch-effort", effort}, id, n)
		k := [3]string{slot, model, effort}
		c := heat[k]
		if c == nil {
			c = &HeatCell{Slot: slot, Model: model, Effort: effort}
			heat[k] = c
		}
		c.Dispatches++
		c.Critical, c.Important, c.Minor = c.Critical+cr, c.Important+im, c.Minor+mi
		return nil
	})
	if err != nil {
		return out, err
	}
	for k, g := range groups {
		add(k.bucket, k.metric, k.name, ratio(g.findings, len(g.dispatches)), len(g.dispatches))
	}
	out.heat = []HeatCell{}
	for _, c := range heat {
		out.heat = append(out.heat, *c)
	}
	sort.Slice(out.heat, func(i, j int) bool {
		a, b := out.heat[i], out.heat[j]
		if a.Slot != b.Slot {
			return a.Slot < b.Slot
		}
		if a.Model != b.Model {
			return a.Model < b.Model
		}
		return a.Effort < b.Effort
	})

	// Implementer, per change.
	var implFindings, implDispatches int
	err = s.scanRows(ctx, "implementer per change", `
		WITH cohort AS (
			SELECT d.change_id, min(d.started_at) AS first_at, count(*) AS dispatches,
				mode() WITHIN GROUP (ORDER BY d.model || ' · ' || d.effort) AS series
			FROM dispatches d JOIN changes c ON c.id = d.change_id
			WHERE d.role = 'implementer'
			  AND ($3::text IS NULL OR c.project_key = $3)
			GROUP BY d.change_id
			HAVING min(d.started_at) >= $1 AND min(d.started_at) < $2
		)
		SELECT date_trunc($4::text, co.first_at AT TIME ZONE 'UTC') AS bucket, co.series, count(*),
			sum((SELECT count(*) FROM findings f WHERE f.change_id = co.change_id))::int,
			sum(co.dispatches)::int
		FROM cohort co GROUP BY bucket, co.series ORDER BY bucket, co.series
	`, args, func(scan func(...any) error) error {
		var b time.Time
		var series string
		var n, f, d int
		if err := scan(&b, &series, &n, &f, &d); err != nil {
			return err
		}
		add(b, "impl-findings-per-task", series, ratio(f, d), d)
		implFindings, implDispatches = implFindings+f, implDispatches+d
		return nil
	})
	if err != nil {
		return out, err
	}
	if implDispatches > 0 {
		v := ratio(implFindings, implDispatches)
		out.summary.ImplFindingsPerTask = &v
	}

	err = s.scanRows(ctx, "implementer blocked share", `
		SELECT date_trunc($4::text, d.started_at AT TIME ZONE 'UTC') AS bucket,
			count(*) FILTER (WHERE d.outcome = ANY($5)), count(*)
		FROM dispatches d JOIN changes c ON c.id = d.change_id
		WHERE d.role = 'implementer' AND d.started_at >= $1 AND d.started_at < $2
		  AND ($3::text IS NULL OR c.project_key = $3)
		GROUP BY bucket ORDER BY bucket
	`, append(args, implBadOutcomes), func(scan func(...any) error) error {
		var b time.Time
		var bad, n int
		if err := scan(&b, &bad, &n); err != nil {
			return err
		}
		add(b, "impl-blocked-share", "blocked", ratio(bad, n), n)
		return nil
	})
	if err != nil {
		return out, err
	}

	// Visual verify, per change.
	var vvChanges, vvFirst int
	err = s.scanRows(ctx, "visual verify per change", `
		WITH cohort AS (
			SELECT d.change_id, min(d.started_at) AS first_at,
				COALESCE(max(substring(d.dispatch_key FROM '^visual-verify-fix-([0-9]+)')::int), 0) AS rounds
			FROM dispatches d JOIN changes c ON c.id = d.change_id
			WHERE d.role = 'verifier' AND d.dispatch_key LIKE 'visual-verify%'
			  AND ($3::text IS NULL OR c.project_key = $3)
			GROUP BY d.change_id
			HAVING min(d.started_at) >= $1 AND min(d.started_at) < $2
		)
		SELECT date_trunc($4::text, first_at AT TIME ZONE 'UTC') AS bucket, count(*),
			sum(rounds)::int, max(rounds), count(*) FILTER (WHERE rounds = 0)
		FROM cohort GROUP BY bucket ORDER BY bucket
	`, args, func(scan func(...any) error) error {
		var b time.Time
		var n, sr, mr, first int
		if err := scan(&b, &n, &sr, &mr, &first); err != nil {
			return err
		}
		add(b, "vv-rounds", "mean", ratio(sr, n), n)
		add(b, "vv-rounds", "max", float64(mr), n)
		add(b, "vv-first-pass", "first-pass", ratio(first, n), n)
		vvChanges, vvFirst = vvChanges+n, vvFirst+first
		return nil
	})
	if err != nil {
		return out, err
	}
	if vvChanges > 0 {
		v := ratio(vvFirst, vvChanges)
		out.summary.VVFirstPassShare = &v
	}

	sort.SliceStable(out.series, func(i, j int) bool {
		a, b := out.series[i], out.series[j]
		if a.Metric != b.Metric {
			return a.Metric < b.Metric
		}
		if !a.Bucket.Equal(b.Bucket) {
			return a.Bucket.Before(b.Bucket)
		}
		return a.Series < b.Series
	})
	return out, nil
}

// scanRows runs query and hands each row's Scan to each, wrapping every
// error with what.
func (s *Store) scanRows(ctx context.Context, what, query string, args []any, each func(scan func(...any) error) error) error {
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("store: review dynamics: %s: %w", what, err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := each(rows.Scan); err != nil {
			return fmt.Errorf("store: review dynamics: %s: scan: %w", what, err)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("store: review dynamics: %s: %w", what, err)
	}
	return nil
}
