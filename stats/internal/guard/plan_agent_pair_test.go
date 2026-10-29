package guard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPlanAgentPair pins scripts/plan-agent-pair.sh: every seed to `opus low`
// and every usage refusal.
func TestPlanAgentPair(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	env := envWith(t, []string{"FLOW_GUARD_REPO_ROOT=" + root})
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
	if r := runGuard("plan-agent-pair", []string{"demo", "1"}, envWith(t, []string{"FLOW_GUARD_REPO_ROOT=" + repo})); r.rc != 0 || r.stdout != "opus medium\n" {
		t.Fatalf("rc=%d out=%q, want opus medium", r.rc, r.out)
	}
	if r := runGuard("plan-agent-pair", []string{"demo", "1"}, envWith(t, []string{"FLOW_GUARD_REPO_ROOT=" + t.TempDir()})); r.rc != 2 {
		t.Fatalf("rc=%d out=%q, want 2 on an unreadable table", r.rc, r.out)
	}
}
