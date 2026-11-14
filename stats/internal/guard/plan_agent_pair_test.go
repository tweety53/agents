package guard

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// TestPlanAgentPair pins plan-agent-pair: every seed and both Explore and
// Plan roles to `opus low`, an effort's pick, and every usage refusal.
func TestPlanAgentPair(t *testing.T) {
	smoke.Heavy(t)
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	env := envWith(t, []string{"FLOW_GUARD_REPO_ROOT=" + root})
	t.Run("every seed gets the row's first pair", func(t *testing.T) {
		for i := range 100 {
			for role, want := range map[string]string{"explorer": "opus low\n", "plan-designer": "opus low\n"} {
				if r := runGuard("plan-agent-pair", []string{fmt.Sprintf("s%d", i), "1", role}, env); r.rc != 0 || r.stdout != want {
					t.Fatalf("s%d 1 %s: rc=%d out=%q, want %q", i, role, r.rc, r.out, want)
				}
			}
		}
	})
	t.Run("an effort picks the row's first pair at it", func(t *testing.T) {
		if r := runGuard("plan-agent-pair", []string{"demo", "1", "planner", "high"}, env); r.rc != 0 || r.stdout != "opus high\n" {
			t.Fatalf("planner high: rc=%d out=%q, want opus high", r.rc, r.out)
		}
		if r := runGuard("plan-agent-pair", []string{"demo", "1", "planner"}, env); r.rc != 0 || r.stdout != "opus medium\n" {
			t.Fatalf("planner: rc=%d out=%q, want opus medium", r.rc, r.out)
		}
		// explorer's row holds no xhigh pair: its first pair answers.
		if r := runGuard("plan-agent-pair", []string{"demo", "1", "explorer", "xhigh"}, env); r.rc != 0 || r.stdout != "opus low\n" {
			t.Fatalf("explorer xhigh: rc=%d out=%q, want opus low", r.rc, r.out)
		}
	})
	for _, args := range [][]string{{"demo", "1", "planner", "core"}, {"demo", "1", "plan-designer", "x"}, {"demo", "1", "planner", "max"}} {
		t.Run(fmt.Sprintf("refuses fourth argument %v", args[3]), func(t *testing.T) {
			if r := runGuard("plan-agent-pair", args, env); r.rc != 2 || r.stdout != "" {
				t.Errorf("rc=%d out=%q, want 2 and no stdout", r.rc, r.out)
			}
		})
	}
	for _, args := range [][]string{{"demo", "0", "explorer"}, {"demo", "x", "explorer"}, {"demo", "1"}, {"demo"}} {
		t.Run(fmt.Sprintf("refuses %v", args), func(t *testing.T) {
			if r := runGuard("plan-agent-pair", args, env); r.rc != 2 || r.stdout != "" {
				t.Errorf("rc=%d out=%q, want 2 and no stdout", r.rc, r.out)
			}
		})
	}
}

// TestPlanAgentPairFollowsPolicy: the pair is the plan-designer base
// pairs' first, read through `flow policy get`, and no answer exits 2.
// smoke: core — a planning dispatch's pair, read from the policy
func TestPlanAgentPairFollowsPolicy(t *testing.T) {
	env := policyFlowEnv(t, `{"plan-designer":{"base":["opus/Plan@medium","opus/Plan@low"],"optin":[],"marked":[]}}`)
	if r := runGuard("plan-agent-pair", []string{"demo", "1", "plan-designer"}, env); r.rc != 0 || r.stdout != "opus medium\n" {
		t.Fatalf("rc=%d out=%q, want opus medium", r.rc, r.out)
	}
	if r := runGuard("plan-agent-pair", []string{"demo", "1", "plan-designer"}, fakeFlowEnvBare(t, "exit 1\n")); r.rc != 2 {
		t.Fatalf("rc=%d out=%q, want 2 with no pairs", r.rc, r.out)
	}
}

// experimentOneFlow is open experiment 1, plan-designer-xhigh-explorer-low:
// each planning role overridden to its own pair.
const experimentOneFlow = `[ "$1 $2" = "experiment overrides" ] || exit 9
printf '%s\n' '{"plan-designer":["opus/Plan@max","sonnet/Plan@high"],"explorer":["sonnet/Explore@low"]}'
`

// TestPlanAgentPairReadsOverride: an open override of the named role's row
// decides its pair, xhigh included, and never the other role's -- the pair
// the agent-baseline hook admits for that Role line; an effort is looked up
// in the overriding pairs, never the replaced base row.
func TestPlanAgentPairReadsOverride(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	env := fakeFlowEnv(t, experimentOneFlow, "FLOW_GUARD_REPO_ROOT="+root)
	for role, want := range map[string]string{"plan-designer": "opus max\n", "explorer": "sonnet low\n"} {
		if r := runGuard("plan-agent-pair", []string{"demo", "1", role}, env); r.rc != 0 || r.stdout != want {
			t.Fatalf("%s: rc=%d out=%q, want %q", role, r.rc, r.out, want)
		}
	}
	// The base row's opus/Plan@high would print `opus high`.
	if r := runGuard("plan-agent-pair", []string{"demo", "1", "plan-designer", "high"}, env); r.rc != 0 || r.stdout != "sonnet high\n" {
		t.Fatalf("plan-designer high: rc=%d out=%q, want sonnet high", r.rc, r.out)
	}
}
