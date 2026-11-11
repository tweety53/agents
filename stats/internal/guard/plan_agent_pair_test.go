package guard

import (
	"fmt"
	"path/filepath"
	"testing"
)

// TestPlanAgentPair pins scripts/plan-agent-pair.sh: every seed and both roles
// to `opus low`, and every usage refusal.
func TestPlanAgentPair(t *testing.T) {
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
	t.Run("core prints the row's first fable pair", func(t *testing.T) {
		if r := runGuard("plan-agent-pair", []string{"demo", "1", "plan-designer", "core"}, env); r.rc != 0 || r.stdout != "fable xhigh\n" {
			t.Fatalf("rc=%d out=%q, want fable xhigh", r.rc, r.out)
		}
		for role, want := range map[string]string{"explorer": "opus low\n", "planner": "fable xhigh\n"} {
			if r := runGuard("plan-agent-pair", []string{"demo", "1", role, "core"}, env); r.rc != 0 || r.stdout != want {
				t.Fatalf("%s core: rc=%d out=%q, want %q", role, r.rc, r.out, want)
			}
		}
		if r := runGuard("plan-agent-pair", []string{"demo", "1", "planner"}, env); r.rc != 0 || r.stdout != "opus medium\n" {
			t.Fatalf("planner: rc=%d out=%q, want opus medium", r.rc, r.out)
		}
		if r := runGuard("plan-agent-pair", []string{"demo", "1", "plan-designer", "x"}, env); r.rc != 2 {
			t.Fatalf("bad fourth arg: rc=%d, want 2", r.rc)
		}
	})
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
printf '%s\n' '{"plan-designer":["opus/Plan@max"],"explorer":["sonnet/Explore@low"]}'
`

// TestPlanAgentPairReadsOverride: an open override of the named role's row
// decides its pair, xhigh included, and never the other role's -- the pair
// the agent-baseline hook admits for that Role line.
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
}
