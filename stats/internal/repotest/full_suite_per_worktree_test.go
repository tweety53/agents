package repotest

import (
	"strings"
	"testing"
)

// TestFullSuiteRunsPerWorktree pins the parent's pre-panel test-list run to
// every worktree of the resolved set (kan-1084 self-review, F1: the frontend's
// `## test` list first ran at flow.verify, after three panel rounds had
// reviewed that tree).
func TestFullSuiteRunsPerWorktree(t *testing.T) {
	t.Parallel()
	s := strings.Join(strings.Fields(section(t, "skills/flow/implement.md", "4. Execute (SDD + TDD)")), " ")
	for _, want := range []string{
		"**The parent runs the test list — never an implementer.**",
		"once per worktree in this run's resolved set",
		"flow-guard run-verify-list <worktree> test",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("skills/flow/implement.md section 4 lacks %q", want)
		}
	}
	for _, gone := range []string{"once on the canonical worktree", "run-verify-list <canonical-worktree> test"} {
		if strings.Contains(s, gone) {
			t.Errorf("skills/flow/implement.md section 4 still has %q", gone)
		}
	}
}

// TestVisualVerifyStopsItsStackAndVerifyChecksFreshness pins that
// visual-verify.md stops the stack step 5 started and verify-and-handoff.md
// runs the dev-stack freshness check.
func TestVisualVerifyStopsItsStackAndVerifyChecksFreshness(t *testing.T) {
	t.Parallel()
	if want := "**Stop the stack step 5 started**"; !strings.Contains(repoText(t, "skills/flow/visual-verify.md"), want) {
		t.Errorf("skills/flow/visual-verify.md lacks %q", want)
	}
	if want := "flow-guard check-dev-stack-fresh"; !strings.Contains(repoText(t, "skills/flow/verify-and-handoff.md"), want) {
		t.Errorf("skills/flow/verify-and-handoff.md lacks %q", want)
	}
}
