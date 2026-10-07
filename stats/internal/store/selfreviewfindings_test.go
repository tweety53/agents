package store_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// TestSelfReviewFindingsRoundTrip records one row of every disposition and
// reads them back in insertion order with every field intact -- the nil
// blast radius of a declined row staying nil rather than reading back 0.
func TestSelfReviewFindingsRoundTrip(t *testing.T) {
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()

	three := 3
	in := []records.SelfReviewFinding{
		{Change: "kan-1", Angle: "myflow-fix", Note: "guard misses a case", Disposition: "fixed", Ref: "abc1234", BlastRadius: &three},
		{Change: "kan-1", Angle: "myflow-cost", Note: "needs a redesign", Disposition: "filed", Ref: "KAN-9"},
		{Change: "kan-1", Angle: "myflow-improvement", Note: "not worth it", Disposition: "declined"},
	}
	for _, f := range in {
		if _, err := st.RecordSelfReviewFinding(ctx, projectKey, f); err != nil {
			t.Fatalf("RecordSelfReviewFinding(%+v): %v", f, err)
		}
	}
	if _, err := st.RecordSelfReviewFinding(ctx, projectKey, records.SelfReviewFinding{
		Change: "kan-2", Angle: "myflow-fix", Note: "other change", Disposition: "declined",
	}); err != nil {
		t.Fatalf("RecordSelfReviewFinding for another change: %v", err)
	}

	got, err := st.ListSelfReviewFindings(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("ListSelfReviewFindings: %v", err)
	}
	if len(got) != len(in) {
		t.Fatalf("got %d rows, want %d: %+v", len(got), len(in), got)
	}
	for i, want := range in {
		g := got[i]
		if g.ID == 0 || g.RecordedAt.IsZero() {
			t.Errorf("row %d: ID/RecordedAt not set: %+v", i, g)
		}
		if g.Change != want.Change || g.Angle != want.Angle || g.Note != want.Note ||
			g.Disposition != want.Disposition || g.Ref != want.Ref {
			t.Errorf("row %d = %+v, want %+v", i, g, want)
		}
		switch {
		case want.BlastRadius == nil && g.BlastRadius != nil:
			t.Errorf("row %d: BlastRadius = %d, want nil", i, *g.BlastRadius)
		case want.BlastRadius != nil && (g.BlastRadius == nil || *g.BlastRadius != *want.BlastRadius):
			t.Errorf("row %d: BlastRadius = %v, want %d", i, g.BlastRadius, *want.BlastRadius)
		}
	}
}

// TestSelfReviewFindingRefusesBadDisposition asserts the store's own
// validation: a disposition outside the closed set, a ref that does not fit
// its disposition, an empty change, angle or note, or a negative blast radius is ErrSelfReviewFindingInvalid
// and writes nothing.
func TestSelfReviewFindingRefusesBadDisposition(t *testing.T) {
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()

	const a, n = "myflow-fix", "a finding"
	negative := -1
	cases := map[string]records.SelfReviewFinding{
		"unknown disposition": {Change: "kan-1", Angle: a, Note: n, Disposition: "bogus"},
		"fixed with a key":    {Change: "kan-1", Angle: a, Note: n, Disposition: "fixed", Ref: "KAN-9"},
		"fixed with no ref":   {Change: "kan-1", Angle: a, Note: n, Disposition: "fixed"},
		"filed with a sha":    {Change: "kan-1", Angle: a, Note: n, Disposition: "filed", Ref: "abc1234"},
		"declined with a ref": {Change: "kan-1", Angle: a, Note: n, Disposition: "declined", Ref: "KAN-9"},
		"empty angle":         {Change: "kan-1", Note: n, Disposition: "declined"},
		"empty note":          {Change: "kan-1", Angle: a, Disposition: "declined"},
		"empty change":        {Angle: a, Note: n, Disposition: "declined"},
		"negative blast":      {Change: "kan-1", Angle: a, Note: n, Disposition: "declined", BlastRadius: &negative},
	}
	for name, f := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := st.RecordSelfReviewFinding(ctx, projectKey, f); !errors.Is(err, store.ErrSelfReviewFindingInvalid) {
				t.Fatalf("err = %v, want ErrSelfReviewFindingInvalid", err)
			}
		})
	}

	got, err := st.ListSelfReviewFindings(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("ListSelfReviewFindings: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("refused writes stored %d rows: %+v", len(got), got)
	}
}

// TestSelfReviewFindingsInPeriod records findings of two changes in two
// projects on two days and asserts the period and project filters select
// exactly the matching rows, newest first, each carrying its project key,
// ref and blast radius.
func TestSelfReviewFindingsInPeriod(t *testing.T) {
	st, pool, projectA := newHazardStore(t)
	ctx := context.Background()

	projectB := projectA + "-b"
	if _, err := pool.Exec(ctx,
		`INSERT INTO projects (project_key, main_checkout_path) VALUES ($1, $2)`,
		projectB, t.TempDir(),
	); err != nil {
		t.Fatalf("seed project %s: %v", projectB, err)
	}

	day1 := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)
	two := 2
	seed := []struct {
		project string
		at      time.Time
		f       records.SelfReviewFinding
	}{
		{projectA, day1, records.SelfReviewFinding{Change: "kan-1", Angle: "flow-fix", Note: "a1", Disposition: "fixed", Ref: "abc1234", BlastRadius: &two}},
		{projectA, day2, records.SelfReviewFinding{Change: "kan-1", Angle: "flow-cost", Note: "a2", Disposition: "filed", Ref: "KAN-9"}},
		{projectB, day2.Add(time.Hour), records.SelfReviewFinding{Change: "kan-2", Angle: "flow-speed", Note: "b1", Disposition: "declined"}},
	}
	for _, s := range seed {
		got, err := st.RecordSelfReviewFinding(ctx, s.project, s.f)
		if err != nil {
			t.Fatalf("RecordSelfReviewFinding(%+v): %v", s.f, err)
		}
		if _, err := pool.Exec(ctx, `UPDATE self_review_findings SET recorded_at = $1 WHERE id = $2`, s.at, got.ID); err != nil {
			t.Fatalf("pin recorded_at: %v", err)
		}
	}

	notes := func(rows []store.SelfReviewFindingRow) []string {
		out := []string{}
		for _, r := range rows {
			out = append(out, r.ProjectKey+"/"+r.Note)
		}
		return out
	}

	// Both days, both projects: newest first. Other tests share the
	// database, so only rows of this test's two projects are compared.
	all, err := st.SelfReviewFindingsInPeriod(ctx, store.Period{From: day1, To: day2.Add(24 * time.Hour)}, nil)
	if err != nil {
		t.Fatalf("SelfReviewFindingsInPeriod: %v", err)
	}
	var mine []store.SelfReviewFindingRow
	for _, r := range all {
		if r.ProjectKey == projectA || r.ProjectKey == projectB {
			mine = append(mine, r)
		}
	}
	if got, want := notes(mine), []string{projectB + "/b1", projectA + "/a2", projectA + "/a1"}; !slices.Equal(got, want) {
		t.Fatalf("all = %v, want %v", got, want)
	}
	if r := mine[2]; r.Ref != "abc1234" || r.BlastRadius == nil || *r.BlastRadius != 2 || r.Change != "kan-1" || r.Angle != "flow-fix" || r.Disposition != "fixed" {
		t.Errorf("fixed row lost a field: %+v", r)
	}

	// Day 2 only, project A only.
	got, err := st.SelfReviewFindingsInPeriod(ctx, store.Period{From: day2, To: day2.Add(24 * time.Hour)}, &projectA)
	if err != nil {
		t.Fatalf("SelfReviewFindingsInPeriod(project): %v", err)
	}
	if got, want := notes(got), []string{projectA + "/a2"}; !slices.Equal(got, want) {
		t.Fatalf("day 2, project A = %v, want %v", got, want)
	}
}
