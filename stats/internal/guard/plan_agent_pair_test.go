package guard

import (
	"fmt"
	"testing"
)

// TestPlanAgentPair pins scripts/plan-agent-pair.sh: every seed to `opus low`
// and every usage refusal.
func TestPlanAgentPair(t *testing.T) {
	env := envWith(t, nil)
	t.Run("always opus low", func(t *testing.T) {
		for i := range 100 {
			if r := runGuard("plan-agent-pair", []string{fmt.Sprintf("s%d", i), "1"}, env); r.rc != 0 || r.stdout != "opus low\n" {
				t.Fatalf("s%d 1: rc=%d out=%q, want opus low", i, r.rc, r.out)
			}
		}
	})
	for _, args := range [][]string{{"demo", "0"}, {"demo", "x"}, {"demo"}} {
		t.Run(fmt.Sprintf("refuses %v", args), func(t *testing.T) {
			if r := runGuard("plan-agent-pair", args, env); r.rc != 2 || r.stdout != "" {
				t.Errorf("rc=%d out=%q, want 2 and no stdout", r.rc, r.out)
			}
		})
	}
}
