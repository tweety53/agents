package harvest

import "testing"

// TestLaunchKeyNamesAPlannersAgentRow pins that a planner's Explore or Plan
// launch is keyed by the explore-<n> or design-<n> row its description
// names, even behind the planner's own plan-<n> prefix (kan-1061's
// explore-1..3 kept no agent id, so their cost stayed unattributed).
func TestLaunchKeyNamesAPlannersAgentRow(t *testing.T) {
	t.Parallel()
	for desc, want := range map[string]string{
		"plan-1 design round — explore-1 store and CLI":  "explore-1",
		"plan-1 design round — design-4 simplest design": "design-4",
		"explore-2 board SPA":                            "explore-2",
		"plan-2 design round":                            "plan-2",
		"Task 3/7":                                       "task-3-",
		"panel-0-primary over the diff":                  "panel-0-primary",
	} {
		if got := launchKey(desc, ""); got != want {
			t.Errorf("launchKey(%q) = %q, want %q", desc, got, want)
		}
	}
}

// TestLaunchKeyPrefersThePromptKeyLine pins decision f12-launch-key-from-prompt
// (kan-1099 self-review, F12): a launch whose prompt carries a column-0 Key:
// line is keyed by it, the row the hook began, not by a description naming a
// task the row lacks; with no Key: line, or only an indented one, the
// description rule stands.
func TestLaunchKeyPrefersThePromptKeyLine(t *testing.T) {
	t.Parallel()
	const desc = "Tasks 8+11+13/34 (fix)"
	for _, c := range []struct{ prompt, want string }{
		{"Role: implementer\nKey: task-11+13-implementer-fix-2\nfix", "task-11+13-implementer-fix-2"},
		{"Role: implementer\nfix", "task-8+11+13-"},
		{"Role: implementer\n  Key: task-11+13-implementer-fix-2\nfix", "task-8+11+13-"},
	} {
		if got := launchKey(desc, c.prompt); got != c.want {
			t.Errorf("launchKey(%q, %q) = %q, want %q", desc, c.prompt, got, c.want)
		}
	}
}
