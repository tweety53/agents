package store_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestSkillLoads pins the skill-load record and its window read: one row per path loaded in
// [from, to), counted, ordered by path; a load outside the window is not counted, and an empty
// path is refused.
func TestSkillLoads(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	for _, l := range []struct {
		path string
		at   time.Time
	}{
		{"skills/flow/fix.md", now.Add(-time.Hour)},
		{"skills/flow/fix.md", now.Add(-2 * time.Hour)},
		{"rules/agent-baseline.md", now.Add(-time.Hour)},
		{"skills/flow/old.md", now.Add(-40 * 24 * time.Hour)},
	} {
		if err := st.RecordSkillLoad(ctx, l.path, "s-1", l.at); err != nil {
			t.Fatalf("RecordSkillLoad(%s): %v", l.path, err)
		}
	}
	if err := st.RecordSkillLoad(ctx, "", "s-1", now); err == nil {
		t.Errorf("an empty path was recorded")
	}
	got, err := st.SkillLoadCounts(ctx, now.Add(-30*24*time.Hour), now)
	if err != nil {
		t.Fatal(err)
	}
	want := []store.SkillLoadCount{{Path: "rules/agent-baseline.md", Loads: 1}, {Path: "skills/flow/fix.md", Loads: 2}}
	if !slices.Equal(got, want) {
		t.Errorf("counts = %+v, want %+v", got, want)
	}
}
