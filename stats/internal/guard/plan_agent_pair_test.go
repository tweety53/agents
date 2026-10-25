package guard

import (
	"fmt"
	"testing"
)

// TestPlanAgentPair pins scripts/plan-agent-pair.sh: fixed seeds to their
// computed pairs, a repeat run to the same line, the sonnet share over a
// fixed seed set (a property of sha256, pinned exactly, not a tolerance) and
// every usage refusal.
func TestPlanAgentPair(t *testing.T) {
	env := envWith(t, nil)
	t.Run("pinned seeds", func(t *testing.T) {
		for seed, want := range map[string]string{"alpha": "sonnet low", "beta": "opus low"} {
			if r := runGuard("plan-agent-pair", []string{seed, "1"}, env); r.rc != 0 || r.stdout != want+"\n" {
				t.Errorf("%s 1: rc=%d out=%q, want %q", seed, r.rc, r.out, want)
			}
		}
	})
	t.Run("reproducible", func(t *testing.T) {
		a := runGuard("plan-agent-pair", []string{"demo", "3"}, env)
		b := runGuard("plan-agent-pair", []string{"demo", "3"}, env)
		if a.rc != 0 || a.stdout != b.stdout {
			t.Errorf("repeat differs: %q vs %q", a.out, b.out)
		}
	})
	t.Run("sonnet share", func(t *testing.T) {
		n := 0
		for i := range 1000 {
			if runGuard("plan-agent-pair", []string{fmt.Sprintf("s%d", i), "1"}, env).stdout == "sonnet low\n" {
				n++
			}
		}
		if n != 381 {
			t.Errorf("sonnet share over s0..s999 = %d, want 381 (~40%%)", n)
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
