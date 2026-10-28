package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestCanonicalSeverityFolds pins canonical_severity
// (0049_review_dynamics.sql): every spelling a reviewer wrote folds onto
// critical, important or minor, case-insensitively (canonical-severity).
func TestCanonicalSeverityFolds(t *testing.T) {
	_, pool := newRecordStore(t)
	ctx := context.Background()

	cases := map[string]string{
		"Critical": "critical", "blocking": "critical", "Severe": "critical",
		"Important": "important", "MAJOR": "important", "high": "important", "Medium": "important", "moderate": "important",
		"Minor": "minor", "low": "minor", "Informational": "minor", "closed": "minor",
	}
	for in, want := range cases {
		var got string
		if err := pool.QueryRow(ctx, "SELECT canonical_severity($1)", in).Scan(&got); err != nil {
			t.Fatalf("canonical_severity(%q): %v", in, err)
		}
		if got != want {
			t.Errorf("canonical_severity(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestUpsertPromptChangeTuningOverEdit pins prompt-changes-table: a tuning
// row upserts over an edit row of the same sha and is never demoted back.
func TestUpsertPromptChangeTuningOverEdit(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	slot := "primary"
	sha := "abc:skills/flow/primary-reviewer-prompt.md"
	at := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	if last, err := st.LastPromptChangeAt(ctx); err != nil || last != nil {
		t.Errorf("LastPromptChangeAt on an empty table = %v, %v; want nil, nil", last, err)
	}

	edit := store.PromptChange{SHA: sha, At: at, Kind: "edit", Area: "review", Slot: &slot,
		Target: "skills/flow/primary-reviewer-prompt.md", Summary: "tighten wording"}
	if err := st.UpsertPromptChange(ctx, edit); err != nil {
		t.Fatalf("upsert edit: %v", err)
	}
	caseID := "kan-1-x-1"
	tuning := edit
	tuningSlot := "test-sanity"
	tuning.Kind, tuning.Summary, tuning.CaseID, tuning.Area = "tuning", "missed nil check", &caseID, "impl"
	tuning.At, tuning.Slot, tuning.Target = at.Add(30*time.Minute), &tuningSlot, "skills/flow/test-sanity-reviewer-prompt.md"
	if err := st.UpsertPromptChange(ctx, tuning); err != nil {
		t.Fatalf("upsert tuning: %v", err)
	}
	if err := st.UpsertPromptChange(ctx, edit); err != nil {
		t.Fatalf("re-upsert edit: %v", err)
	}
	later := store.PromptChange{SHA: "def:skills/flow/visual-verify-verifier.md", At: at.Add(48 * time.Hour),
		Kind: "edit", Area: "vv", Target: "skills/flow/visual-verify-verifier.md", Summary: "later"}
	if err := st.UpsertPromptChange(ctx, later); err != nil {
		t.Fatalf("upsert later: %v", err)
	}

	got, err := st.ListPromptChanges(ctx, at.Add(-time.Hour), at.Add(72*time.Hour))
	if err != nil {
		t.Fatalf("ListPromptChanges: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListPromptChanges = %d rows, want 2: %+v", len(got), got)
	}
	if got[0].SHA != sha || got[0].Kind != "tuning" || got[0].CaseID == nil || *got[0].CaseID != caseID || got[0].Summary != "missed nil check" ||
		!got[0].At.Equal(tuning.At) || got[0].Slot == nil || *got[0].Slot != tuningSlot || got[0].Target != tuning.Target || got[0].Area != "impl" {
		t.Errorf("first row = %+v, want the tuning kept over the edit", got[0])
	}
	if got[1].SHA != later.SHA || got[1].Slot != nil {
		t.Errorf("second row = %+v, want %s ordered by at, no slot", got[1], later.SHA)
	}
	if inRange, err := st.ListPromptChanges(ctx, at.Add(time.Hour), at.Add(48*time.Hour)); err != nil || len(inRange) != 0 {
		t.Errorf("ListPromptChanges over [at+1h, at+48h) = %+v, %v; want none (half-open)", inRange, err)
	}

	last, err := st.LastPromptChangeAt(ctx)
	if err != nil || last == nil || !last.Equal(later.At) {
		t.Errorf("LastPromptChangeAt = %v, %v; want %v", last, err, later.At)
	}
	bad := later
	bad.SHA, bad.Area = "zzz:x", "planner"
	if err := st.UpsertPromptChange(ctx, bad); err == nil {
		t.Error("UpsertPromptChange accepted area planner")
	}
}
