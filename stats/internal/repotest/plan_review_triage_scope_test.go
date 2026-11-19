package repotest

// TestPlanCheckTriageScope pins KAN-1110: a plan-check triage reads the
// report and the artifacts, and a planner file only where a finding's fix
// turns on a round's answer.

import (
	"strings"
	"testing"
)

func TestPlanCheckTriageScope(t *testing.T) {
	t.Parallel()
	text := repoText(t, "skills/flow/brainstorm-planner.md")
	i := strings.Index(text, "**A dispatch whose prompt names a `plan-check.md` report triages it.**")
	if i < 0 {
		t.Fatal("skills/flow/brainstorm-planner.md lacks the triage paragraph")
	}
	para := text[i:]
	j := strings.Index(para, "### Plan check gate")
	if j < 0 {
		t.Fatal("the triage paragraph is not followed by ### Plan check gate")
	}
	para = para[:j]
	if strings.Contains(para, "### Decide") {
		t.Error("a ### Decide heading remains")
	}
	for _, want := range []string{
		"plan-check.md",
		"It reads the report, `proposal.md`, `design.md` and `tasks.md`; it opens a file under `<abs-worktree>/.superpowers/sdd/planner/` only for a finding whose fix turns on a round's answer",
		"re-run `spectre validate`",
	} {
		if !strings.Contains(para, want) {
			t.Errorf("the triage paragraph lacks %q", want)
		}
	}
}
