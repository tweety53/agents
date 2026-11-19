package store

import "testing"

func TestAutoStageForSevenMarks(t *testing.T) {
	t.Parallel()
	for key, want := range map[string]string{
		"plan-1":              "flow.plan",
		"panel-1-bugs":        "flow.review",
		"fix-3-review":        "flow.review",
		"verify-1":            "flow.verify",
		"visual-verify-1":     "flow.verify",
		"self-review-plan":    "flow.self-review",
		"pipeline-fix-1-plan": "flow.pipeline-fix",
		"task-3-implementer":  "flow.implement",
		"something-unknown":   "flow.implement",
	} {
		if got := autoStageFor(key); got != want {
			t.Errorf("autoStageFor(%q) = %q, want %q", key, got, want)
		}
	}
}
