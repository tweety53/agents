package repotest

import (
	"strings"
	"testing"
)

// TestPlannerBriefsLoadThePlanShapeRules pins decisions briefs-load-rules and
// cut-restating-prose: each brief that writes a tasks.md runs
// `flow-guard check-plan-shape -rules` first, and the sentences that only
// restated a printed rule are gone.
func TestPlannerBriefsLoadThePlanShapeRules(t *testing.T) {
	t.Parallel()
	const pointer = "flow-guard check-plan-shape -rules"
	for name, text := range map[string]string{
		"brainstorm-planner.md section D": section(t, "skills/flow/brainstorm-planner.md", "D. Basic Workflow #3 — Writing plans"),
		"flow-fast SKILL.md":              repoText(t, "skills/flow-fast/SKILL.md"),
		"verify-fix-loop.md":              repoText(t, "skills/flow/verify-fix-loop.md"),
	} {
		if !strings.Contains(text, pointer) {
			t.Errorf("%s does not run %s", name, pointer)
		}
	}
	for rel, cut := range map[string][]string{
		"skills/flow/brainstorm-planner.md": {
			"F22 refuses a plan whose design cites the prototype", "F27 refuses a plan whose",
			"F19 refuses a plan-mode design", "Go's test cache never rechecks", "F26 refuses any other npx",
			"(F12)", "(F17)",
		},
		"skills/flow-fast/SKILL.md":      {"an indented field is invisible to the parser"},
		"skills/flow/verify-fix-loop.md": {"F24 refuses an"},
	} {
		text := repoText(t, rel)
		for _, s := range cut {
			if strings.Contains(text, s) {
				t.Errorf("%s still restates a printed rule: %q", rel, s)
			}
		}
	}
}
