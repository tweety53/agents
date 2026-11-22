package store_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// TestConfidenceEstimatesRoundTrip records a plan row and an implement row
// and reads them back, newest first, every field intact; the tag filter
// keeps only the rows carrying the tag, and the tag list counts each tag.
// Tags are unique to this test: the table is read across projects.
func TestConfidenceEstimatesRoundTrip(t *testing.T) {
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()
	shared, only := "rt-"+projectKey, "rt-only-"+projectKey

	plan := records.ConfidenceEstimate{Change: "kan-1", Stage: "plan", Score: 78, Disposition: "later",
		Gaps: []string{"F9 — closes with: plan — no task checks the flag", "F10 — closes with: operator — column order"}, Tags: []string{shared}}
	impl := records.ConfidenceEstimate{Change: "kan-1", Stage: "implement", Task: "2+3", Score: 90,
		Gaps: []string{"the fixture the test needs does not exist"}, Tags: []string{shared, only}}
	for _, e := range []records.ConfidenceEstimate{plan, impl} {
		got, err := st.RecordConfidenceEstimate(ctx, projectKey, e)
		if err != nil {
			t.Fatalf("RecordConfidenceEstimate(%+v): %v", e, err)
		}
		if got.ID == 0 || got.RecordedAt.IsZero() || got.Project != projectKey {
			t.Errorf("recorded row lacks id, time or project: %+v", got)
		}
	}

	rows, total, err := st.ConfidenceEstimatesPage(ctx, shared, 10, 0)
	if err != nil {
		t.Fatalf("ConfidenceEstimatesPage: %v", err)
	}
	if total != 2 || len(rows) != 2 {
		t.Fatalf("tag %s: total %d, %d rows, want 2 and 2", shared, total, len(rows))
	}
	if g := rows[0]; g.Stage != "implement" || g.Task != "2+3" || g.Score != 90 || g.Disposition != "" ||
		!slices.Equal(g.Gaps, impl.Gaps) || !slices.Equal(g.Tags, impl.Tags) || g.Project != projectKey {
		t.Errorf("newest row = %+v, want the implement row", g)
	}
	if g := rows[1]; g.Stage != "plan" || g.Task != "" || g.Disposition != "later" || !slices.Equal(g.Gaps, plan.Gaps) {
		t.Errorf("second row = %+v, want the plan row", g)
	}
	if rows, total, err := st.ConfidenceEstimatesPage(ctx, only, 10, 0); err != nil || total != 1 || len(rows) != 1 || rows[0].Stage != "implement" {
		t.Errorf("tag %s: rows %+v total %d err %v, want the implement row alone", only, rows, total, err)
	}
	if rows, total, err := st.ConfidenceEstimatesPage(ctx, shared, 1, 1); err != nil || total != 2 || len(rows) != 1 || rows[0].Stage != "plan" {
		t.Errorf("page 2 of 1: rows %+v total %d err %v, want the plan row and total 2", rows, total, err)
	}
	if _, total, err := st.ConfidenceEstimatesPage(ctx, "", 1, 0); err != nil || total < 2 {
		t.Errorf("no tag: total %d err %v, want every row", total, err)
	}

	tags, err := st.ConfidenceTags(ctx)
	if err != nil {
		t.Fatalf("ConfidenceTags: %v", err)
	}
	counts := map[string]int{}
	for _, tg := range tags {
		counts[tg.Tag] = tg.Count
	}
	if counts[shared] != 2 || counts[only] != 1 {
		t.Errorf("tag counts = %v, want %s 2 and %s 1", counts, shared, only)
	}
}

// TestConfidenceEstimateRefusals: every shape the boundary refuses is
// ErrConfidenceEstimateInvalid and writes nothing.
func TestConfidenceEstimateRefusals(t *testing.T) {
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	ok := records.ConfidenceEstimate{Change: "kan-1", Stage: "implement", Task: "1", Score: 80, Gaps: []string{"g"}, Tags: []string{"refusal-" + projectKey}}
	for name, mutate := range map[string]func(*records.ConfidenceEstimate){
		"no change":                func(e *records.ConfidenceEstimate) { e.Change = "" },
		"score 100":                func(e *records.ConfidenceEstimate) { e.Score = 100 },
		"score 0":                  func(e *records.ConfidenceEstimate) { e.Score = 0 },
		"no gap":                   func(e *records.ConfidenceEstimate) { e.Gaps = nil },
		"blank gap":                func(e *records.ConfidenceEstimate) { e.Gaps = []string{" "} },
		"no tag":                   func(e *records.ConfidenceEstimate) { e.Tags = nil },
		"tag not kebab":            func(e *records.ConfidenceEstimate) { e.Tags = []string{"Missing Fixture"} },
		"unknown stage":            func(e *records.ConfidenceEstimate) { e.Stage = "review" },
		"implement without task":   func(e *records.ConfidenceEstimate) { e.Task = "" },
		"implement with a verdict": func(e *records.ConfidenceEstimate) { e.Disposition = "later" },
		"plan without disposition": func(e *records.ConfidenceEstimate) { e.Stage, e.Task = "plan", "" },
		"plan with a task":         func(e *records.ConfidenceEstimate) { e.Stage, e.Disposition = "plan", "raised" },
		"plan with bad verdict":    func(e *records.ConfidenceEstimate) { e.Stage, e.Task, e.Disposition = "plan", "", "fixed" },
	} {
		e := ok
		mutate(&e)
		if _, err := st.RecordConfidenceEstimate(context.Background(), projectKey, e); !errors.Is(err, store.ErrConfidenceEstimateInvalid) {
			t.Errorf("%s: err = %v, want ErrConfidenceEstimateInvalid", name, err)
		}
	}
	if _, total, err := st.ConfidenceEstimatesPage(context.Background(), ok.Tags[0], 10, 0); err != nil || total != 0 {
		t.Errorf("after refusals: total %d err %v, want 0", total, err)
	}
}
