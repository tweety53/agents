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
		{Change: "kan-1", Angle: "myflow-improvement", Note: "the panel caught it early", Disposition: "noted"},
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
		"noted with a ref":    {Change: "kan-1", Angle: a, Note: n, Disposition: "noted", Ref: "abc1234"},
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

// TestSelfReviewFindingRecordedOnce asserts a repeated write of one finding
// -- the same sha recorded again with another note, the same refless note
// recorded again -- stores no second row and answers with the recorded one,
// its missing blast radius filled from the repeat, while a rewritten sha and
// a different refless note are new findings.
func TestSelfReviewFindingRecordedOnce(t *testing.T) {
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()

	one := 1
	writes := []records.SelfReviewFinding{
		{Change: "kan-1", Angle: "flow-fix", Note: "the landing's own row", Disposition: "fixed", Ref: "14a900ef"},
		{Change: "kan-1", Angle: "flow-fix", Note: "recorded again by hand", Disposition: "fixed", Ref: "14a900ef", BlastRadius: &one},
		{Change: "kan-1", Angle: "flow-fix", Note: "recorded again by hand", Disposition: "fixed", Ref: "14a900ef", BlastRadius: &one},
		{Change: "kan-1", Angle: "flow-fix", Note: "the landing's own row", Disposition: "fixed", Ref: "b0d0cdd5"},
		{Change: "kan-1", Angle: "flow-cost", Note: "an observation", Disposition: "noted"},
		{Change: "kan-1", Angle: "flow-cost", Note: "an observation", Disposition: "noted"},
		{Change: "kan-1", Angle: "flow-cost", Note: "another observation", Disposition: "noted"},
	}
	var ids []int64
	for _, f := range writes {
		got, err := st.RecordSelfReviewFinding(ctx, projectKey, f)
		if err != nil {
			t.Fatalf("RecordSelfReviewFinding(%+v): %v", f, err)
		}
		ids = append(ids, got.ID)
	}
	if ids[1] != ids[0] || ids[2] != ids[0] || ids[5] != ids[4] {
		t.Errorf("a repeat answered with a new row: ids %v", ids)
	}

	got, err := st.ListSelfReviewFindings(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("ListSelfReviewFindings: %v", err)
	}
	var keys []string
	for _, f := range got {
		keys = append(keys, f.Ref+"/"+f.Note)
	}
	want := []string{"14a900ef/the landing's own row", "b0d0cdd5/the landing's own row", "/an observation", "/another observation"}
	if !slices.Equal(keys, want) {
		t.Fatalf("rows = %v, want %v", keys, want)
	}
	if got[0].BlastRadius == nil || *got[0].BlastRadius != 1 {
		t.Errorf("first row's blast radius = %v, want 1 from the repeat", got[0].BlastRadius)
	}
}

// TestSelfReviewFindingsMigrationCollapsesRepeats asserts
// 0038_self_review_findings_once.sql collapses rows written twice before it
// onto their first row, which takes the first non-null blast radius among
// them, and leaves distinct findings alone.
func TestSelfReviewFindingsMigrationCollapsesRepeats(t *testing.T) {
	st, pool, projectKey := newHazardStore(t)
	ctx := context.Background()

	if _, err := pool.Exec(ctx, `DROP INDEX self_review_findings_once`); err != nil {
		t.Fatalf("drop the 0038 index: %v", err)
	}
	for _, r := range []struct {
		note, disposition, ref string
		blast                  *int
	}{
		{"the landing's own row", "fixed", "a4a32afc", nil},
		{"recorded again by hand", "fixed", "a4a32afc", ptr(16)},
		{"recorded a third time", "fixed", "a4a32afc", ptr(15)},
		{"another fix", "fixed", "766e24c4", ptr(2)},
		{"an observation", "noted", "", nil},
		{"an observation", "noted", "", nil},
	} {
		if _, err := pool.Exec(ctx, `
			INSERT INTO self_review_findings (project_key, change, angle, note, disposition, ref, blast_radius)
			VALUES ($1, 'kan-1', 'flow-fix', $2, $3, NULLIF($4, ''), $5)
		`, projectKey, r.note, r.disposition, r.ref, r.blast); err != nil {
			t.Fatalf("seed %q: %v", r.note, err)
		}
	}
	if _, err := pool.Exec(ctx, `DELETE FROM schema_migrations WHERE filename = '0038_self_review_findings_once.sql'`); err != nil {
		t.Fatalf("unrecord migration: %v", err)
	}
	if err := st.RunMigrations(ctx); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	got, err := st.ListSelfReviewFindings(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("ListSelfReviewFindings: %v", err)
	}
	var keys []string
	for _, f := range got {
		keys = append(keys, f.Ref+"/"+f.Note)
	}
	want := []string{"a4a32afc/the landing's own row", "766e24c4/another fix", "/an observation"}
	if !slices.Equal(keys, want) {
		t.Fatalf("rows = %v, want %v", keys, want)
	}
	if got[0].BlastRadius == nil || *got[0].BlastRadius != 16 {
		t.Errorf("kept row's blast radius = %v, want 16, the first non-null among its repeats", got[0].BlastRadius)
	}
}
