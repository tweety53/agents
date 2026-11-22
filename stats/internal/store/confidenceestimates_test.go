package store_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// pageOf reads the status's whole list and keeps the rows of one project:
// the table is read across projects, so a test looks only at its own.
func pageOf(t *testing.T, st *store.Store, status, projectKey string) []records.ConfidenceEstimate {
	t.Helper()
	var out []records.ConfidenceEstimate
	for offset := 0; ; offset += 100 {
		p, err := st.ConfidenceEstimatesPage(context.Background(), status, 100, offset)
		if err != nil {
			t.Fatalf("ConfidenceEstimatesPage(%s): %v", status, err)
		}
		for _, r := range p.Rows {
			if r.Project == projectKey {
				out = append(out, r)
			}
		}
		if len(p.Rows) < 100 {
			return out
		}
	}
}

// TestConfidenceEstimatesRoundTrip records a plan row and an implement row
// and reads them back, newest first, every field intact and both to-review
// whatever the caller sent; the tag list counts each tag. Tags are unique to
// this test: the table is read across projects.
func TestConfidenceEstimatesRoundTrip(t *testing.T) {
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()
	shared, only := "rt-"+projectKey, "rt-only-"+projectKey

	plan := records.ConfidenceEstimate{Change: "kan-1", Stage: "plan", Score: 78, Disposition: "later",
		Gaps: []string{"F9 — closes with: plan — no task checks the flag", "F10 — closes with: operator — column order"}, Tags: []string{shared}}
	impl := records.ConfidenceEstimate{Change: "kan-1", Stage: "implement", Task: "2+3", Score: 90, ReviewStatus: "approved", Reason: "sent by the caller",
		Gaps: []string{"the fixture the test needs does not exist"}, Tags: []string{shared, only}}
	for _, e := range []records.ConfidenceEstimate{plan, impl} {
		got, err := st.RecordConfidenceEstimate(ctx, projectKey, e)
		if err != nil {
			t.Fatalf("RecordConfidenceEstimate(%+v): %v", e, err)
		}
		if got.ID == 0 || got.RecordedAt.IsZero() || got.Project != projectKey || got.ReviewStatus != "to-review" || got.Reason != "" {
			t.Errorf("recorded row lacks id, time or project, or is not to-review with no reason: %+v", got)
		}
	}

	rows := pageOf(t, st, "to-review", projectKey)
	if len(rows) != 2 {
		t.Fatalf("to-review: %d rows, want 2: %+v", len(rows), rows)
	}
	if g := rows[0]; g.Stage != "implement" || g.Task != "2+3" || g.Score != 90 || g.Disposition != "" || g.ReviewStatus != "to-review" || g.Reason != "" ||
		!slices.Equal(g.Gaps, impl.Gaps) || !slices.Equal(g.Tags, impl.Tags) {
		t.Errorf("newest row = %+v, want the implement row, to-review", g)
	}
	if g := rows[1]; g.Stage != "plan" || g.Task != "" || g.Disposition != "later" || !slices.Equal(g.Gaps, plan.Gaps) {
		t.Errorf("second row = %+v, want the plan row", g)
	}
	if p, err := st.ConfidenceEstimatesPage(ctx, "to-review", 1, 0); err != nil || len(p.Rows) != 1 || p.Total < 2 || p.ToReview != p.Total {
		t.Errorf("limit 1: %+v err %v, want one row, a total of 2 or more and the to-review count equal to it", p, err)
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

// TestReviewConfidenceEstimate: a review moves the row between the status
// lists and keeps its reason; moving it back clears the reason; an unknown
// status is invalid and another project's id is not found.
func TestReviewConfidenceEstimate(t *testing.T) {
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()
	row, err := st.RecordConfidenceEstimate(ctx, projectKey, records.ConfidenceEstimate{Change: "kan-1", Stage: "implement", Task: "1", Score: 80, Gaps: []string{"g"}, Tags: []string{"rv-" + projectKey}})
	if err != nil {
		t.Fatal(err)
	}
	before, err := st.ConfidenceEstimatesPage(ctx, "declined", 1, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := st.ReviewConfidenceEstimate(ctx, projectKey, row.ID, "declined", "not a real gap"); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if rows := pageOf(t, st, "to-review", projectKey); len(rows) != 0 {
		t.Errorf("to-review after a decline = %+v, want none", rows)
	}
	if rows := pageOf(t, st, "declined", projectKey); len(rows) != 1 || rows[0].ReviewStatus != "declined" || rows[0].Reason != "not a real gap" {
		t.Errorf("declined = %+v, want the row with its reason", rows)
	}
	// Other tests record in parallel, never decline: the declined count moved by this row at least.
	if after, err := st.ConfidenceEstimatesPage(ctx, "to-review", 1, 0); err != nil || after.Declined < before.Declined+1 {
		t.Errorf("declined count %d (err %v), want at least %d", after.Declined, err, before.Declined+1)
	}

	if err := st.ReviewConfidenceEstimate(ctx, projectKey, row.ID, "to-review", ""); err != nil {
		t.Fatalf("move back: %v", err)
	}
	if rows := pageOf(t, st, "to-review", projectKey); len(rows) != 1 || rows[0].Reason != "" {
		t.Errorf("to-review after moving back = %+v, want the row with no reason", rows)
	}

	if err := st.ReviewConfidenceEstimate(ctx, projectKey, row.ID, "fixed", ""); !errors.Is(err, store.ErrConfidenceEstimateInvalid) {
		t.Errorf("unknown status: err = %v, want ErrConfidenceEstimateInvalid", err)
	}
	if err := st.ReviewConfidenceEstimate(ctx, "no-such-project", row.ID, "approved", ""); !errors.Is(err, store.ErrConfidenceEstimateNotFound) {
		t.Errorf("another project's id: err = %v, want ErrConfidenceEstimateNotFound", err)
	}
	if _, err := st.ConfidenceEstimatesPage(ctx, "", 10, 0); !errors.Is(err, store.ErrConfidenceEstimateInvalid) {
		t.Errorf("page with no status: err = %v, want ErrConfidenceEstimateInvalid", err)
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
		"score 101":                func(e *records.ConfidenceEstimate) { e.Score = 101 },
		"100 with a gap":           func(e *records.ConfidenceEstimate) { e.Score = 100 },
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
	if rows := pageOf(t, st, "to-review", projectKey); len(rows) != 0 {
		t.Errorf("after refusals: %+v, want no row", rows)
	}
}

// TestConfidenceEstimatesForChange reads one change's estimates, oldest first,
// whatever their review status, and none of another change's or project's.
func TestConfidenceEstimatesForChange(t *testing.T) {
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()
	for _, e := range []records.ConfidenceEstimate{
		{Change: "kan-7", Stage: "plan", Score: 70, Disposition: "later", Gaps: []string{"a"}, Tags: []string{"fc-" + projectKey}},
		{Change: "kan-7", Stage: "implement", Task: "1", Score: 90, Gaps: []string{"b"}, Tags: []string{"fc-" + projectKey}},
		{Change: "kan-8", Stage: "plan", Score: 50, Disposition: "later", Gaps: []string{"c"}, Tags: []string{"fc-" + projectKey}},
	} {
		if _, err := st.RecordConfidenceEstimate(ctx, projectKey, e); err != nil {
			t.Fatal(err)
		}
	}
	got, err := st.ConfidenceEstimatesForChange(ctx, projectKey, "kan-7")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Stage != "plan" || got[0].Score != 70 || got[1].Task != "1" || got[1].Score != 90 {
		t.Errorf("kan-7 = %+v, want the plan row then the implement row", got)
	}
	if none, err := st.ConfidenceEstimatesForChange(ctx, projectKey, "kan-9"); err != nil || none == nil || len(none) != 0 {
		t.Errorf("kan-9 = %+v, %v, want an empty non-nil slice", none, err)
	}
}

// TestConfidenceHundredIsKeptOutsideTheReviewList: a 100% estimate with no gap and no tag is stored and
// read back for its change, and no review list or count holds it.
func TestConfidenceHundredIsKeptOutsideTheReviewList(t *testing.T) {
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()
	before, err := st.ConfidenceEstimatesPage(ctx, "to-review", 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.RecordConfidenceEstimate(ctx, projectKey, records.ConfidenceEstimate{Change: "kan-100", Stage: "implement", Task: "4", Score: 100}); err != nil {
		t.Fatalf("a 100%% estimate with no gap and no tag: %v", err)
	}
	got, err := st.ConfidenceEstimatesForChange(ctx, projectKey, "kan-100")
	if err != nil || len(got) != 1 || got[0].Score != 100 || len(got[0].Gaps) != 0 {
		t.Fatalf("for change: %+v, %v", got, err)
	}
	after, err := st.ConfidenceEstimatesPage(ctx, "to-review", 100, 0)
	if err != nil || after.ToReview != before.ToReview || after.Total != before.Total {
		t.Errorf("the review list changed: before %+v after %+v (%v)", before.ToReview, after.ToReview, err)
	}
}
