package repotest

// TestBrainstormAsksOnlyClassifiedQuestions pins KAN-1110's question
// classifier: planning asks only architectural and impactful questions, and
// every other ambiguity takes its recommended option under `## Auto-taken`.

import (
	"strings"
	"testing"
)

func TestBrainstormAsksOnlyClassifiedQuestions(t *testing.T) {
	t.Parallel()
	checklist := strings.Join(strings.Fields(section(t, "skills/flow/brainstorm-planner.md", "B. Basic Workflow #1 — Brainstorming")), " ")
	for _, want := range []string{
		"[impactful: <criterion>]",
		"`no-recommendation`", "`irreversible`", "`scope`", "`plan-size`",
		"## Auto-taken",
		"Now or later is never asked",
		"Silence answers nothing",
		"Fails under:",
	} {
		if !strings.Contains(checklist, want) {
			t.Errorf("brainstorm-planner.md section B lacks %q", want)
		}
	}
	if !strings.Contains(checklist, "The Summary names no auto-taken decision") {
		t.Error("brainstorm-planner.md section B lets the relayed Summary name auto-taken decisions")
	}
	if strings.Contains(checklist, "overturns one by answering **Revise**") {
		t.Error("brainstorm-planner.md section B still has the operator overturn an unseen auto-taken pick")
	}
	if strings.Contains(checklist, "Every ambiguity is asked, however small") {
		t.Error("brainstorm-planner.md section B still asks every ambiguity")
	}
	if strings.Contains(repoText(t, "skills/flow/brainstorm.md"), "in a call of its own") {
		t.Error("skills/flow/brainstorm.md still asks each architectural question in a call of its own")
	}
}
