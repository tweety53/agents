package store_test

import (
	"context"
	"errors"
	"fmt"
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()

	one := 1
	writes := []records.SelfReviewFinding{
		{Change: "kan-1", Angle: "flow-fix", Note: "the landing's own row", Disposition: "fixed", Ref: "14a900ef"},
		{Change: "kan-1", Angle: "flow-fix", Note: "the landing's own row", Disposition: "fixed", Ref: "14a900ef", BlastRadius: &one},
		{Change: "kan-1", Angle: "flow-fix", Note: "the landing's own row", Disposition: "fixed", Ref: "14a900ef", BlastRadius: &one},
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
	t.Parallel()
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

// TestRecordSelfReviewFindingSeverity pins the optional severity: recorded
// and listed back lowercased, empty when omitted, and refused outside
// critical, important and minor.
func TestRecordSelfReviewFindingSeverity(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-sr-sev-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	in := records.SelfReviewFinding{Change: "kan-1", Angle: "review", Note: "n1", Disposition: "noted", Severity: "Important"}
	if _, err := st.RecordSelfReviewFinding(ctx, projectKey, in); err != nil {
		t.Fatalf("RecordSelfReviewFinding: %v", err)
	}
	in.Note, in.Severity = "n2", ""
	if _, err := st.RecordSelfReviewFinding(ctx, projectKey, in); err != nil {
		t.Fatalf("RecordSelfReviewFinding(no severity): %v", err)
	}
	in.Note, in.Severity = "n3", "major"
	if _, err := st.RecordSelfReviewFinding(ctx, projectKey, in); !errors.Is(err, store.ErrSelfReviewFindingInvalid) {
		t.Errorf("severity major: err = %v, want ErrSelfReviewFindingInvalid", err)
	}
	got, err := st.ListSelfReviewFindings(ctx, projectKey, "kan-1")
	if err != nil || len(got) != 2 || got[0].Severity != "important" || got[1].Severity != "" {
		t.Fatalf("ListSelfReviewFindings = %+v, %v; want severities important, empty", got, err)
	}
}

// TestSelfReviewFindingsSharingARefStayDistinct pins that findings filed
// together under one issue -- the same ref, different notes -- each keep
// their row: a shared ref never hides a note.
func TestSelfReviewFindingsSharingARefStayDistinct(t *testing.T) {
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()

	for _, note := range []string{"first cost finding", "second cost finding"} {
		if _, err := st.RecordSelfReviewFinding(ctx, projectKey, records.SelfReviewFinding{
			Change: "kan-1", Angle: "flow-cost", Note: note, Disposition: "filed", Ref: "KAN-1136",
		}); err != nil {
			t.Fatalf("RecordSelfReviewFinding(%q): %v", note, err)
		}
	}
	got, err := st.ListSelfReviewFindings(ctx, projectKey, "kan-1")
	if err != nil {
		t.Fatalf("ListSelfReviewFindings: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d rows, want 2: %+v", len(got), got)
	}
}

// TestSelfReviewSuggestionsAreReviewed pins the suggested disposition: it needs a detail, starts
// to-review, is reviewed by id, and a finding of any other disposition cannot be reviewed.
func TestSelfReviewSuggestionsAreReviewed(t *testing.T) {
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()

	if _, err := st.RecordSelfReviewFinding(ctx, projectKey, records.SelfReviewFinding{Change: "kan-1", Angle: "flow-improvement", Note: "x", Disposition: "suggested"}); !errors.Is(err, store.ErrSelfReviewFindingInvalid) {
		t.Fatalf("a suggestion without a detail: err = %v, want ErrSelfReviewFindingInvalid", err)
	}
	if _, err := st.RecordSelfReviewFinding(ctx, projectKey, records.SelfReviewFinding{Change: "kan-1", Angle: "flow-improvement", Note: "tiny", Disposition: "suggested", Severity: "minor", Detail: "d"}); !errors.Is(err, store.ErrSelfReviewFindingInvalid) {
		t.Fatalf("a minor suggestion: err = %v, want ErrSelfReviewFindingInvalid", err)
	}
	got, err := st.RecordSelfReviewFinding(ctx, projectKey, records.SelfReviewFinding{Change: "kan-1", Angle: "flow-improvement", Note: "cache the guard", Disposition: "suggested", Severity: "important", Detail: "## Why\nslow\n## How\ncache\n## DoD\nfaster"})
	if err != nil {
		t.Fatalf("record suggestion: %v", err)
	}
	if got.ReviewStatus != "to-review" || got.Detail == "" {
		t.Fatalf("recorded = %+v, want to-review with its detail", got)
	}
	if err := st.ReviewSelfReviewFinding(ctx, projectKey, got.ID, "approved", "cheap and visible"); err != nil {
		t.Fatalf("approve: %v", err)
	}
	list, _ := st.ListSelfReviewFindings(ctx, projectKey, "kan-1")
	if len(list) != 1 || list[0].ReviewStatus != "approved" || list[0].Detail != got.Detail {
		t.Fatalf("listed = %+v, want the approved suggestion with its detail", list)
	}
	pg, err := st.SelfReviewFindingsPage(ctx, store.Period{From: time.Now().Add(-time.Hour), To: time.Now().Add(time.Hour)}, nil, "to-review", "approved", 10, 0)
	if err != nil || len(pg.Rows) != 1 || pg.Rows[0].Reason != "cheap and visible" {
		t.Fatalf("approved page = %+v, err %v, want the row with reason %q", pg.Rows, err, "cheap and visible")
	}
	if err := st.ReviewSelfReviewFinding(ctx, projectKey, got.ID, "to-review", ""); err != nil {
		t.Fatalf("move back: %v", err)
	}
	if pg, _ = st.SelfReviewFindingsPage(ctx, store.Period{From: time.Now().Add(-time.Hour), To: time.Now().Add(time.Hour)}, nil, "to-review", "to-review", 10, 0); len(pg.Rows) != 1 || pg.Rows[0].Reason != "" {
		t.Fatalf("moved back page = %+v, want the row with its reason cleared", pg.Rows)
	}
	fixed, err := st.RecordSelfReviewFinding(ctx, projectKey, records.SelfReviewFinding{Change: "kan-1", Angle: "flow-fix", Note: "f", Disposition: "fixed", Ref: "abc1234", Severity: "minor", Detail: "d"})
	if err != nil {
		t.Fatalf("record fixed: %v", err)
	}
	if err := st.ReviewSelfReviewFinding(ctx, projectKey, fixed.ID, "approved", ""); !errors.Is(err, store.ErrSelfReviewFindingNotSuggested) {
		t.Fatalf("review of a fixed finding: err = %v, want ErrSelfReviewFindingNotSuggested", err)
	}
}

// TestSelfReviewFindingsPagePagesAndCounts: a page holds at most limit rows of the chosen list,
// newest first, with the list's total and the suggestion counts by status.
func TestSelfReviewFindingsPagePagesAndCounts(t *testing.T) {
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if _, err := st.RecordSelfReviewFinding(ctx, projectKey, records.SelfReviewFinding{Change: "kan-1", Angle: "flow-fix", Note: fmt.Sprintf("fix %d", i), Disposition: "fixed", Ref: "abc1234", Severity: "important"}); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 3; i++ {
		f, err := st.RecordSelfReviewFinding(ctx, projectKey, records.SelfReviewFinding{Change: "kan-1", Angle: "flow-improvement", Note: fmt.Sprintf("idea %d", i), Disposition: "suggested", Severity: "important", Detail: "d"})
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			_ = st.ReviewSelfReviewFinding(ctx, projectKey, f.ID, "approved", "")
		}
	}
	period := store.Period{From: time.Now().Add(-time.Hour), To: time.Now().Add(time.Hour)}
	page, err := st.SelfReviewFindingsPage(ctx, period, nil, "fixed", "", 2, 2)
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if len(page.Rows) != 2 || page.Total != 5 || page.Rows[0].Note != "fix 2" {
		t.Errorf("fixed page 2 = %d rows, total %d, first %q; want 2 rows, total 5, first \"fix 2\"", len(page.Rows), page.Total, page.Rows[0].Note)
	}
	if page.ToReview != 2 || page.Approved != 1 || page.Declined != 0 {
		t.Errorf("counts = %d/%d/%d, want 2/1/0", page.ToReview, page.Approved, page.Declined)
	}
	page, err = st.SelfReviewFindingsPage(ctx, period, nil, "to-review", "to-review", 10, 0)
	if err != nil || len(page.Rows) != 2 || page.Total != 2 {
		t.Errorf("to-review page = %+v, err %v; want 2 rows", page, err)
	}
	if _, err := st.SelfReviewFindingsPage(ctx, period, nil, "bogus", "", 10, 0); !errors.Is(err, store.ErrSelfReviewFindingInvalid) {
		t.Errorf("bogus show: err = %v, want ErrSelfReviewFindingInvalid", err)
	}
}

// TestGetSelfReviewFinding: a finding reads back by id with its detail, from any project's key, and an id no row
// has is ErrSelfReviewFindingNotFound.
func TestGetSelfReviewFinding(t *testing.T) {
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()
	got, err := st.RecordSelfReviewFinding(ctx, projectKey, records.SelfReviewFinding{Change: "kan-1", Angle: "flow-improvement", Note: "cache the guard", Disposition: "suggested", Severity: "important", Detail: "## Proposal\nx"})
	if err != nil {
		t.Fatal(err)
	}
	read, err := st.GetSelfReviewFinding(ctx, projectKey, got.ID)
	if err != nil || read.ID != got.ID || read.Detail != "## Proposal\nx" || read.Change != "kan-1" {
		t.Fatalf("read = %+v, err %v, want the recorded finding with its detail", read, err)
	}
	if _, err := st.GetSelfReviewFinding(ctx, projectKey, got.ID+1000000); !errors.Is(err, store.ErrSelfReviewFindingNotFound) {
		t.Fatalf("unknown id: err = %v, want ErrSelfReviewFindingNotFound", err)
	}
	// The id is global: a read from another project's directory finds it (the dashboard's Copy for agent).
	if other, err := st.GetSelfReviewFinding(ctx, projectKey+"-other", got.ID); err != nil || other.ID != got.ID {
		t.Fatalf("another project's read = %+v, err %v, want the finding", other, err)
	}
}

// TestSelfReviewFindingThemeAndEffortColumns: the theme and effort columns are read from the note's
// theme prefix and the detail's Effort line, and are empty when the text carries neither.
func TestSelfReviewFindingThemeAndEffortColumns(t *testing.T) {
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()
	for _, f := range []records.SelfReviewFinding{
		{Change: "kan-1", Angle: "flow-improvement", Note: "[theme: guard-false-pass] cache it", Disposition: "suggested", Severity: "important", Detail: "## Effort and risk\nEffort: M\nTouches: x"},
		{Change: "kan-1", Angle: "flow-improvement", Note: "no prefix", Disposition: "suggested", Severity: "important", Detail: "## Why\nlong"},
	} {
		if _, err := st.RecordSelfReviewFinding(ctx, projectKey, f); err != nil {
			t.Fatal(err)
		}
	}
	page, err := st.SelfReviewFindingsPage(ctx, store.Period{From: time.Now().Add(-time.Hour), To: time.Now().Add(time.Hour)}, nil, "to-review", "to-review", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string][2]string{}
	for _, r := range page.Rows {
		got[r.Note] = [2]string{r.Theme, r.Effort}
	}
	if got["[theme: guard-false-pass] cache it"] != [2]string{"guard-false-pass", "M"} || got["no prefix"] != [2]string{"", ""} {
		t.Fatalf("theme/effort = %v", got)
	}
}

// TestSimilarSelfReviewFindings: the findings sharing the query's words come back best first with a
// score, the theme prefix is not searched, an unrelated one is left out, and a bad query or n is invalid.
func TestSimilarSelfReviewFindings(t *testing.T) {
	t.Parallel()
	st, _, projectKey := newHazardStore(t)
	ctx := context.Background()
	for _, f := range []records.SelfReviewFinding{
		{Change: "kan-1", Angle: "a", Note: "[theme: guard-false-pass] guard passes on a stale cache", Disposition: "noted"},
		{Change: "kan-2", Angle: "a", Note: "[theme: other] button colour", Disposition: "noted", Detail: "dashboard styling"},
	} {
		if _, err := st.RecordSelfReviewFinding(ctx, projectKey, f); err != nil {
			t.Fatal(err)
		}
	}
	got, err := st.SimilarSelfReviewFindings(ctx, projectKey, "stale cache guard", 5)
	if err != nil || len(got) != 1 || got[0].Change != "kan-1" || got[0].Score <= 0 || got[0].Theme != "guard-false-pass" {
		t.Fatalf("got %+v, err %v; want only kan-1 with a score and its theme", got, err)
	}
	if got, err := st.SimilarSelfReviewFindings(ctx, projectKey, "theme", 5); err != nil || len(got) != 0 {
		t.Fatalf("theme prefix searched: %+v, err %v", got, err)
	}
	for _, c := range []struct {
		q string
		n int
	}{{"", 5}, {"x", 0}, {"x", 51}} {
		if _, err := st.SimilarSelfReviewFindings(ctx, projectKey, c.q, c.n); !errors.Is(err, store.ErrSelfReviewFindingInvalid) {
			t.Fatalf("%q n=%d: err = %v, want ErrSelfReviewFindingInvalid", c.q, c.n, err)
		}
	}
}
