package repotest

// TestPlannerPairFollowsScope pins KAN-1110's planner pair: the effort follows
// the round's scope, and no planning contract keeps a CORE special case — a
// pair outside the base rows comes only through an operator grant.

import (
	"strings"
	"testing"
)

func TestPlannerPairFollowsScope(t *testing.T) {
	t.Parallel()
	brainstorm := repoText(t, "skills/flow/brainstorm.md")
	planner := repoText(t, "skills/flow/brainstorm-planner.md")
	policy := repoText(t, "skills/flow-contracts/model-policy.md")
	for name, text := range map[string]string{
		"skills/flow/brainstorm.md":             brainstorm,
		"skills/flow/brainstorm-planner.md":     planner,
		"skills/flow-contracts/model-policy.md": policy,
	} {
		for _, gone := range []string{"CORE change", "planner core", "Agents (CORE)"} {
			if strings.Contains(text, gone) {
				t.Errorf("%s still holds %q", name, gone)
			}
		}
	}
	if want := "plan-agent-pair <name> <n> planner <effort>"; !strings.Contains(brainstorm, want) {
		t.Errorf("skills/flow/brainstorm.md lacks %q", want)
	}
	if want := "**Scope:** small|regular|large"; !strings.Contains(planner, want) {
		t.Errorf("skills/flow/brainstorm-planner.md lacks %q", want)
	}
	if want := "Granted: <role> <model> <effort>"; !strings.Contains(planner, want) {
		t.Errorf("skills/flow/brainstorm-planner.md lacks %q", want)
	}
}
