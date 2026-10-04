package store_test

import (
	"context"
	"errors"
	"testing"

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
