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
	text := repoText(t, "skills/flow/plan-triage.md")
	i := strings.Index(text, "**A dispatch whose prompt names a `plan-check.md` report triages it.**")
	if i < 0 {
		t.Fatal("skills/flow/plan-triage.md lacks the triage paragraph")
	}
	para := text[i:]
	if !strings.Contains(repoText(t, "skills/flow/brainstorm.md"), "`skills/flow/plan-triage.md`") {
		t.Error("skills/flow/brainstorm.md's plan check does not name skills/flow/plan-triage.md")
	}
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
