package repotest

// TestPlannerReadsLatestRoundOnly pins KAN-1110's rolling round state: a
// planner round reads the latest round, its answers and findings.md, and the
// agents a planner dispatch launches are soft-capped.

import (
	"strings"
	"testing"
)

func TestPlannerReadsLatestRoundOnly(t *testing.T) {
	t.Parallel()
	run := strings.Join(strings.Fields(section(t, "skills/flow/brainstorm.md", "Run brainstorming and planning through the planner role")), " ")
	for _, want := range []string{"findings.md", "never an earlier round or answers file"} {
		if !strings.Contains(run, want) {
			t.Errorf("brainstorm.md's planner section lacks %q", want)
		}
	}
	if strings.Contains(run, "every file under") {
		t.Error("brainstorm.md's planner section still reads every file under the planner directory")
	}
	planner := repoText(t, "skills/flow/brainstorm-planner.md")
	for _, want := range []string{"**Findings outlive the round.**", "### Settled", "Over cap:"} {
		if !strings.Contains(planner, want) {
			t.Errorf("skills/flow/brainstorm-planner.md lacks %q", want)
		}
	}
}
