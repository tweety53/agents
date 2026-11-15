package repotest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSendBacksRunResumeOrFresh pins that both immediate send-backs — the
// close-task re-commit and the -cite send-back — resume the agent only on
// `flow-guard resume-or-fresh`'s `resume`, that a resumed row's begin passes
// the id the guard printed (KAN-1117; self-review-kan-1110 F17 for the
// -agent-id half), and that the pick-conflict fixer and the loop contracts
// stay one-shot (decisions pick-conflict-stays-fresh,
// loop-contracts-stay-one-shot).
func TestSendBacksRunResumeOrFresh(t *testing.T) {
	t.Parallel()
	for path, wants := range map[string][]string{
		"skills/flow/implement.md": {
			"flow-guard resume-or-fresh <worktree> <change> <the closed row's key>",
			"its `begin` passing `-agent-id <the id the guard printed>`",
			"behind `flow-guard resume-or-fresh` on the group's closed `task-<ids>-implementer` row",
		},
		"skills/flow/sdd-dispatch.md": {
			"behind `flow-guard resume-or-fresh` on the member's own closed row",
		},
		"skills/flow/review-panel-fix-round.md": {
			"flow-guard resume-or-fresh <worktree> <change> <the failed re-run's key>",
			"resumes the failed re-run's agent with `SendMessage`",
			"On any other exit it is a fresh dispatch of the same slot prompt",
		},
		"skills/flow/review-panel.md": {
			"save a send-back `flow-guard resume-or-fresh` answered `resume` for",
		},
	} {
		text := repoText(t, path)
		for _, want := range wants {
			if !strings.Contains(text, want) {
				t.Errorf("%s lacks %q", path, want)
			}
		}
	}
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), "skills/flow/implement.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "| a wave member's pick conflict |") && strings.Contains(line, "resume-or-fresh") {
			t.Errorf("implement.md's pick-conflict row runs resume-or-fresh: %s", line)
		}
	}
	if strings.Contains(repoText(t, "skills/flow-contracts/pipeline.md"), "resume-or-fresh") {
		t.Error("pipeline.md's loop contracts name resume-or-fresh; they stay one-shot")
	}
}
