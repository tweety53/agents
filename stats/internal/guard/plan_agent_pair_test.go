package guard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	t.Run("always opus low", func(t *testing.T) {
		for i := range 100 {
			for _, role := range []string{"explorer", "plan-designer"} {
				if r := runGuard("plan-agent-pair", []string{fmt.Sprintf("s%d", i), "1", role}, env); r.rc != 0 || r.stdout != "opus low\n" {
					t.Fatalf("s%d 1 %s: rc=%d out=%q, want opus low", i, role, r.rc, r.out)
				}
			}
		}
	})
	for _, args := range [][]string{{"demo", "0", "explorer"}, {"demo", "x", "explorer"}, {"demo", "1"}, {"demo", "1", "planner"}, {"demo"}} {
		t.Run(fmt.Sprintf("refuses %v", args), func(t *testing.T) {
			if r := runGuard("plan-agent-pair", args, env); r.rc != 2 || r.stdout != "" {
				t.Errorf("rc=%d out=%q, want 2 and no stdout", r.rc, r.out)
			}
		})
	}
}

// TestPlanAgentPairFollowsPolicy: the pair is the plan-designer row's first,
// read from the allowed-pairs table, and an unreadable table exits 2.
func TestPlanAgentPairFollowsPolicy(t *testing.T) {
	b, err := os.ReadFile("../../../" + mpFile)
	if err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir()
	table := strings.Replace(string(b), "| `plan-designer` | `opus/Plan@low`, `opus/Plan@medium` |", "| `plan-designer` | `opus/Plan@medium`, `opus/Plan@low` |", 1)
	if table == string(b) {
		t.Fatal("the real table carries no plan-designer row to rewrite")
	}
	writeFile(t, repo+"/"+mpFile, table)
	if r := runGuard("plan-agent-pair", []string{"demo", "1", "plan-designer"}, envWith(t, []string{"FLOW_GUARD_REPO_ROOT=" + repo})); r.rc != 0 || r.stdout != "opus medium\n" {
		t.Fatalf("rc=%d out=%q, want opus medium", r.rc, r.out)
	}
	if r := runGuard("plan-agent-pair", []string{"demo", "1", "plan-designer"}, envWith(t, []string{"FLOW_GUARD_REPO_ROOT=" + t.TempDir()})); r.rc != 2 {
		t.Fatalf("rc=%d out=%q, want 2 on an unreadable table", r.rc, r.out)
	}
}

// experimentOneFlow is open experiment 1, plan-designer-xhigh-explorer-low:
// each planning role overridden to its own pair.
const experimentOneFlow = `[ "$1 $2" = "experiment overrides" ] || exit 9
printf '%s\n' '{"plan-designer":["opus/Plan@xhigh"],"explorer":["sonnet/Explore@low"]}'
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
	for role, want := range map[string]string{"plan-designer": "opus xhigh\n", "explorer": "sonnet low\n"} {
		if r := runGuard("plan-agent-pair", []string{"demo", "1", role}, env); r.rc != 0 || r.stdout != want {
			t.Fatalf("%s: rc=%d out=%q, want %q", role, r.rc, r.out, want)
		}
	}
}
