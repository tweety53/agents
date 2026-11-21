package records

import "testing"

func TestStoredRole(t *testing.T) {
	for _, c := range []struct{ role, key, policy, want string }{
		{"verifier", "verify", "", "runner"},
		{"verifier", "verify-2-gymie", "", "runner"},
		{"verifier", "runner-1", "", "runner"},
		{"verifier", "visual-verify-2", "", "visual-verifier"},
		{"merger", "panel-1-merger", "", "panel-merger"},
		{"reviewer", "panel-1-primary", "", "panel-reviewer"},
		{"reviewer", "panel-1-primary", "panel-rerun", "panel-rerun"},
		{"reviewer", "task-1-reviewer", "", "gated-reviewer"},
		{"reviewer", "self-review-x-pass", "", "reviewer"},
		{"planner", "design-1", "", "plan-designer"},
		{"runner", "anything", "", "runner"},
		{"implementer", "operator-3", "", "operator-asked"},
		{"implementer", "task-2-implementer", "", "implementer"},
		{"reviewer", "self-review-x-review-1", "self-review-review-1", "self-review-review-1"},
	} {
		if got := StoredRole(c.role, c.key, c.policy); got != c.want {
			t.Errorf("StoredRole(%q, %q, %q) = %q, want %q", c.role, c.key, c.policy, got, c.want)
		}
	}
}

func TestClassFoldsRolesTableRolesToTheCoarseClass(t *testing.T) {
	for role, want := range map[string]string{
		"runner": "verifier", "visual-verifier": "verifier", "panel-reviewer": "reviewer",
		"gated-re-reviewer": "reviewer", "panel-merger": "merger", "minor-fix": "panel-fix",
		"plan-designer": "planner", "operator-asked": "implementer",
		"explorer": "explorer", "reviewer": "reviewer", "decider": "decider",
	} {
		if got := Class(role); got != want {
			t.Errorf("Class(%q) = %q, want %q", role, got, want)
		}
	}
}
