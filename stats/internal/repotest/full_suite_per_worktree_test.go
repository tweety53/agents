package repotest

import (
	"strings"
	"testing"
)

// TestFullSuiteRunsPerWorktree pins the parent's pre-panel full-suite run to
// every worktree of the resolved set (kan-1084 self-review, F1: the frontend's
// `## test` list first ran at flow.verify, after three panel rounds had
// reviewed that tree).
func TestFullSuiteRunsPerWorktree(t *testing.T) {
	t.Parallel()
	s := strings.Join(strings.Fields(section(t, "skills/flow/implement.md", "4. Execute (SDD + TDD)")), " ")
	for _, want := range []string{
		"**The parent runs the full suite — never an implementer.**",
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
