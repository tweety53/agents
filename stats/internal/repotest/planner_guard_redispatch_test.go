package repotest

// TestPlannerGuardRedispatch pins KAN-1110: a final planner's guard
// redispatch runs on the low planner pair and, after three failed rounds,
// takes the shared rule in skills/flow-contracts/pipeline.md.

import (
	"strings"
	"testing"
)

func TestPlannerGuardRedispatch(t *testing.T) {
	t.Parallel()
	run := strings.Join(strings.Fields(section(t, "skills/flow/brainstorm.md", "Run brainstorming and planning through the planner role")), " ")
	for _, want := range []string{"plan-agent-pair <name> <n> planner low", "**Three failed rounds** (`skills/flow-contracts/pipeline.md`)"} {
		if !strings.Contains(run, want) {
			t.Errorf("brainstorm.md's guard redispatch lacks %q", want)
		}
	}
	if strings.Contains(run, "until it returns its word") {
		t.Error("brainstorm.md's guard redispatch still loops until the planner returns its word")
	}
}
