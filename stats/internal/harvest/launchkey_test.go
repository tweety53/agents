package harvest

import "testing"

// TestLaunchKeyNamesAPlannersAgentRow pins that a planner's Explore or Plan
// launch is keyed by the explore-<n> or design-<n> row its description
// names, even behind the planner's own plan-<n> prefix (kan-1061's
// explore-1..3 kept no agent id, so their cost stayed unattributed).
func TestLaunchKeyNamesAPlannersAgentRow(t *testing.T) {
	for desc, want := range map[string]string{
		"plan-1 design round — explore-1 store and CLI": "explore-1",
		"plan-1 design round — design-4 simplest design": "design-4",
		"explore-2 board SPA":                            "explore-2",
		"plan-2 design round":                            "plan-2",
		"Task 3/7":                                       "task-3-",
		"panel-0-primary over the diff":                  "panel-0-primary",
	} {
		if got := launchKey(desc); got != want {
			t.Errorf("launchKey(%q) = %q, want %q", desc, got, want)
		}
	}
}
