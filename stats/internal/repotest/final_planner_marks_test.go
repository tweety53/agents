package repotest

// TestParentBeginsFinalPlannerStages pins KAN-1110: the parent begins
// flow.create-artifacts and flow.writing-plans before it records the final
// planner that ends each, so the stage window covers the dispatch.

import (
	"strings"
	"testing"
)

func TestParentBeginsFinalPlannerStages(t *testing.T) {
	t.Parallel()
	run := section(t, "skills/flow/brainstorm.md", "Run brainstorming and planning through the planner role")
	dispatch := strings.Index(run, "- the first runs section **C**")
	if dispatch < 0 {
		t.Fatal("brainstorm.md's planner section lacks step 3's first dispatch")
	}
	for _, stage := range []string{"flow.create-artifacts", "flow.writing-plans"} {
		begin := strings.Index(run, "flow stage begin -command '/flow' -stage "+stage+" ")
		if begin < 0 || begin > dispatch {
			t.Errorf("brainstorm.md's planner section has no `flow stage begin` of %s before step 3's dispatch text", stage)
		}
	}
}
