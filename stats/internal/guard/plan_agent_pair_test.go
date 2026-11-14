package guard

import (
	"fmt"
	"os"
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

// TestPlanAgentPairAuto: `planner auto <round-file>` prints the planner row's
// pair at the effort the round file's own signals call for — Scope sets the
// start, three architectural questions or ten distinct cited repo paths
// each raise it one step, never past high, and no file or no Scope line is
// medium; auto for another role, or without a round file, is a usage error.
func TestPlanAgentPairAuto(t *testing.T) {
	env := policyFlowEnv(t, `{"planner":{"base":["opus/flow-planner-medium","opus/flow-planner-low","opus/flow-planner-high","opus/flow-planner-xhigh"],"optin":[],"marked":[]},"explorer":{"base":["opus/Explore@low"],"optin":[],"marked":[]}}`)
	dir := t.TempDir()
	arch := func(n int) string {
		s := ""
		for i := 1; i <= n; i++ {
			s += fmt.Sprintf("### Q%d. [architectural] which way?\n\n", i)
		}
		return s
	}
	paths := func(n int) string {
		s := ""
		for i := range n {
			// each path cited twice, once with a line suffix: distinct count is n
			s += fmt.Sprintf("see `skills/a%d/f.md` and `skills/a%d/f.md:12-14`\n", i, i)
		}
		return s
	}
	linePaths := func(n int) string {
		s := ""
		for i := range n {
			s += fmt.Sprintf("see `skills/b%d/f.md:%d`\n", i, i+1)
		}
		return s
	}
	for name, tc := range map[string]struct{ body, want string }{
		"small, no signals is low":                        {"**Scope:** small — one file\n", "low"},
		"regular is medium":                               {"**Scope:** regular — two areas\n", "medium"},
		"large is high":                                   {"**Scope:** large — spans repos\n", "high"},
		"no scope line is medium":                         {"# Round 2\n", "medium"},
		"two architectural questions stay":                {"**Scope:** small — x\n" + arch(2), "low"},
		"three architectural questions raise":             {"**Scope:** small — x\n" + arch(3), "medium"},
		"nine distinct paths stay":                        {"**Scope:** small — x\n" + paths(9), "low"},
		"ten distinct paths raise":                        {"**Scope:** small — x\n" + paths(10), "medium"},
		"both signals raise two steps":                    {"**Scope:** small — x\n" + arch(3) + paths(10), "high"},
		"never above high":                                {"**Scope:** large — x\n" + arch(4) + paths(12), "high"},
		"ten paths cited only with a line suffix raise":   {"**Scope:** small — x\n" + linePaths(10), "medium"},
		"a mid-line architectural heading is no question": {"**Scope:** small — x\n" + arch(2) + "quoted `### Q3. [architectural] which way?` text\n", "low"},
		"a mid-line Scope mention is not read":            {"see the `**Scope:** large` line\n**Scope:** small — x\n", "low"},
		"a slashless or extensionless token is no path":   {"**Scope:** small — x\n" + "`a.md` `b/c` `d.go`\n" + paths(9), "low"},
	} {
		t.Run(name, func(t *testing.T) {
			f := filepath.Join(t.TempDir(), "round-2.md")
			if err := os.WriteFile(f, []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			if r := runGuard("plan-agent-pair", []string{"demo", "2", "planner", "auto", f}, env); r.rc != 0 || r.stdout != "opus "+tc.want+"\n" {
				t.Fatalf("rc=%d out=%q, want opus %s", r.rc, r.out, tc.want)
			}
		})
	}
	t.Run("a missing round file is medium", func(t *testing.T) {
		if r := runGuard("plan-agent-pair", []string{"demo", "2", "planner", "auto", filepath.Join(dir, "absent.md")}, env); r.rc != 0 || r.stdout != "opus medium\n" {
			t.Fatalf("rc=%d out=%q, want opus medium", r.rc, r.out)
		}
	})
	for _, args := range [][]string{{"demo", "2", "explorer", "auto", "r.md"}, {"demo", "2", "planner", "auto"}, {"demo", "2", "planner", "high", "r.md"}} {
		t.Run(fmt.Sprintf("refuses %v", args), func(t *testing.T) {
			if r := runGuard("plan-agent-pair", args, env); r.rc != 2 || r.stdout != "" {
				t.Errorf("rc=%d out=%q, want 2 and no stdout", r.rc, r.out)
			}
		})
	}
}
