package repotest

import (
	"strings"
	"testing"
)

// TestPairJudgeSites pins KAN-1119's call sites: every judged dispatch site
// names its `flow-guard pair-judge` call, the canonical rule in **Model and
// effort** names the call, its flags and its recording, and no planning
// contract keeps a roll, the planner's auto scorer or plan-agent-pair's
// no-effort form.
func TestPairJudgeSites(t *testing.T) {
	t.Parallel()
	for file, wants := range map[string][]string{
		"skills/flow/brainstorm-planner.md": {
			"flow-guard pair-judge <role> <digest-file> [-opus-only]",
			"-pair-pick '<source>: <reason>'",
			"Pair pick: <source>: <reason>",
			"flow-guard pair-judge panel-rerun",
			"flow-guard pair-judge panel-sanity-efficiency",
			"flow-guard pair-judge implementer",
		},
		"skills/flow/brainstorm.md": {
			"flow-guard pair-judge planner <digest>",
			"flow-guard pair-judge design-reviewer <prompt file>",
		},
		"skills/flow/implement.md":             {"flow-guard pair-judge gated-reviewer"},
		"skills/flow/gated-review-fix.md":      {"flow-guard pair-judge gated-re-reviewer"},
		"skills/flow/review-panel-late-fix.md": {"flow-guard pair-judge late-fix"},
		"skills/flow/document-fix.md":          {"flow-guard pair-judge implementer"},
		"skills/flow-self-review/SKILL.md": {
			"flow-guard pair-judge <role> <prompt file>",
			"flow-guard pair-judge self-review-review-1",
		},
	} {
		text := repoText(t, file)
		for _, want := range wants {
			if !strings.Contains(text, want) {
				t.Errorf("%s lacks %s", file, want)
			}
		}
	}
	for _, file := range []string{
		"skills/flow/brainstorm-planner.md",
		"skills/flow/brainstorm.md",
		"skills/flow-plan/SKILL.md",
		"skills/flow-self-review/SKILL.md",
	} {
		text := repoText(t, file)
		for _, gone := range []string{"effort_roll", "planner auto", "plan-agent-pair <seed>"} {
			if strings.Contains(text, gone) {
				t.Errorf("%s still holds %q", file, gone)
			}
		}
	}
}
