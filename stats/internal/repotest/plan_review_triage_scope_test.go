package repotest

// TestPlanReviewTriageScope pins KAN-1110: a plan-review triage reads the
// report and the artifacts, and a planner file only where a finding's fix
// turns on a round's answer.

import (
	"strings"
	"testing"
)

func TestPlanReviewTriageScope(t *testing.T) {
	t.Parallel()
	text := repoText(t, "skills/flow/brainstorm-planner.md")
	i := strings.Index(text, "**A dispatch whose prompt names a plan-review report triages it.**")
	if i < 0 {
		t.Fatal("skills/flow/brainstorm-planner.md lacks the triage paragraph")
	}
	para := text[i:]
	if j := strings.Index(para, "### Decide"); j >= 0 {
		para = para[:j]
	}
	for _, want := range []string{
		"It reads the report, `proposal.md` and `design.md`, and `tasks.md` when triaging `plan-review-2`; it opens a file under `<abs-worktree>/.superpowers/sdd/planner/` only for a finding whose fix turns on a round's answer",
		"Make no stage mark.",
		"re-run `spectre validate`",
	} {
		if !strings.Contains(para, want) {
			t.Errorf("the triage paragraph lacks %q", want)
		}
	}
}
