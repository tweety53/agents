package store

import "testing"

func TestAutoStageFor(t *testing.T) {
	for key, want := range map[string]string{
		"plan-1":              "flow.brainstorm",
		"plan-4":              "flow.brainstorm",
		"task-3-implementer":  "flow.sdd-tdd",
		"panel-1-bugs":        "flow.review-panel",
		"visual-verify-1":     "flow.visual-verify",
		"self-review-plan":    "flow.self-review",
		"pipeline-fix-1-plan": "flow.pipeline-fix",
		"something-unknown":   "flow.sdd-tdd",
	} {
		if got := autoStageFor(key); got != want {
			t.Errorf("autoStageFor(%q) = %q, want %q", key, got, want)
		}
	}
}
