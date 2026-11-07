package guard

import (
	"testing"

	"github.com/tweety53/agents/stats/internal/records"
)

// TestGuardHelpersAreRegistryKeys pins records.GuardHelpers to Registry:
// every helper the guard-health view leaves out must be a real flow-guard
// command, so a renamed or removed command cannot linger in the set.
func TestGuardHelpersAreRegistryKeys(t *testing.T) {
	if len(records.GuardHelpers) == 0 {
		t.Fatal("records.GuardHelpers is empty")
	}
	for _, name := range records.GuardHelpers {
		if _, ok := Registry[name]; !ok {
			t.Errorf("records.GuardHelpers entry %q is not a guard.Registry key", name)
		}
	}
}
