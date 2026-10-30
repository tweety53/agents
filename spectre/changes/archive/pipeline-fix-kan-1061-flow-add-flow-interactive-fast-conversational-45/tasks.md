# pipeline-fix-kan-1061-flow-add-flow-interactive-fast-conversational-45

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** `scripts/plan-agent-pair.sh` prints the pair of the dispatch's own role row — `explorer`
or `plan-designer` — so the pair it prints is the one `hooks/enforce-agent-baseline.py` admits.

**Architecture:** `planAgentPair` (`stats/internal/guard/planagentpair.go`) takes a required third
argument `<role>` and reads `policyTable`'s row of that name, overrides overlaid; the shim's usage
header and the four contracts citing the usage name the role.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; it was applied to this worktree at e6c5c19e and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- No new inline suppressions; `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits
  clean.
- Read-only against the dev store: never stop flowd or flow-postgres, never drop the `flow` DB.
  The test never calls the real `flow`; it stubs `flow experiment overrides` through `fakeFlowEnv`.

## Review Focus

- The row read is the `<role>` argument's, never a fixed one (`role-argument`):
  `TestPlanAgentPairReadsOverride` gives the two roles different overrides.
- A missing or unknown role exits 2 with no stdout — no silent default to `plan-designer`.
- Every live caller passes the role: `git grep -n 'plan-agent-pair.sh <'` outside
  `spectre/changes/archive/` and `docs/self-review/` shows a third argument on each line.

- [x] 1. Print the planning pair of the dispatch's own role row
  - [x] **Step 1: Reproduce before touching anything; do not guess.** From this worktree run
    `scripts/plan-agent-pair.sh x 1` with experiment 1 open in the live store: it prints
    `opus xhigh`, the plan-designer pair, which an Explore dispatch cannot use. Save the diff
    below to `/tmp/pf45-task1.diff`, apply its test hunk alone
    (`git apply --include='stats/internal/guard/plan_agent_pair_test.go' /tmp/pf45-task1.diff`) and
    run `cd stats && go test ./internal/guard -run TestPlanAgentPair`: `TestPlanAgentPairReadsOverride`
    fails with `plan-designer: rc=2 out="usage: plan-agent-pair.sh <seed> <n>"`.
  - [x] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/guard/plan_agent_pair_test.go' /tmp/pf45-task1.diff`):

````diff verified:applied with git apply to this worktree at e6c5c19e; go test ./internal/guard passes with it and TestPlanAgentPairReadsOverride fails with planagentpair.go reverted; check-references, check-dispatch-paragraphs and check-guard-symlinks exit 0; plan-agent-pair.sh x 1 explorer prints sonnet low and x 1 plan-designer opus xhigh under experiment 1
diff --git a/scripts/plan-agent-pair.sh b/scripts/plan-agent-pair.sh
index 2ddac9bf..44e0251a 100755
--- a/scripts/plan-agent-pair.sh
+++ b/scripts/plan-agent-pair.sh
@@ -2,14 +2,16 @@
 # plan-agent-pair.sh — the model and effort of a planning session's Explore or
 # Plan dispatch.
 #
-# Usage: plan-agent-pair.sh <seed> <n>
+# Usage: plan-agent-pair.sh <seed> <n> <explorer|plan-designer>
 #
-# Prints one line, `<model> <effort>` — the first pair of the plan-designer
-# row of **Allowed pairs** (skills/flow-contracts/model-policy.md). <seed> is the change name — in /flow-plan
-# before capture, the session token; <n> counts the session's planning
-# dispatches from 1. Exit 0 on the printed line, exit 2 on a usage error
-# (<n> not a positive integer, or an argument count other than 2) or an
-# unreadable table.
+# Prints one line, `<model> <effort>` — the first pair of the <role> row,
+# `explorer` for an Explore dispatch and `plan-designer` for a Plan one, of
+# **Allowed pairs** (skills/flow-contracts/model-policy.md), an open
+# experiment's override of that row applied. <seed> is the change name — in
+# /flow-plan before capture, the session token; <n> counts the session's
+# planning dispatches from 1. Exit 0 on the printed line, exit 2 on a usage
+# error (<n> not a positive integer, a role other than the two, or an
+# argument count other than 3) or an unreadable table.
 #
 # The logic is the Go port in stats/internal/guard/planagentpair.go; flow-guard
 # is built from this checkout by scripts/lib/flow-guard.sh, never taken from
diff --git a/skills/flow-plan/SKILL.md b/skills/flow-plan/SKILL.md
index ba36591c..dc8c77d0 100644
--- a/skills/flow-plan/SKILL.md
+++ b/skills/flow-plan/SKILL.md
@@ -31,7 +31,7 @@ get there is scripted. You're a thinking partner helping the user explore.
 `/flow-plan` runs entirely in the current session, on its own model, with any argument shape —
 a Jira key, an existing change's name, a bare topic, or nothing at all. The session may dispatch
 the fresh Explore and Plan agents **Explore and design agents** (`skills/flow/brainstorm-planner.md`)
-allows — its pair from `plan-agent-pair.sh <session-token> <n>`, seeded with this session's
+allows — its pair from `plan-agent-pair.sh <session-token> <n> <role>`, seeded with this session's
 token — never a `fork`; the session itself
 does the thinking: it reads the tree, runs the investigate-then-ask rounds and the convergence check
 below, and creates the change or offers the `design.md` capture directly. Every "you" elsewhere
diff --git a/skills/flow-self-review/SKILL.md b/skills/flow-self-review/SKILL.md
index 5ca5392a..f9a28180 100644
--- a/skills/flow-self-review/SKILL.md
+++ b/skills/flow-self-review/SKILL.md
@@ -341,9 +341,9 @@ any other project's run, `self-review-<name>` based at `origin/<agents-base>`.
    > **PLANNING DELEGATION:** You may dispatch fresh `Explore` and `Plan` subagents through the
    > `Agent` tool while you plan, per **B. Basic Workflow #1 — Brainstorming** (`skills/flow/brainstorm-planner.md`) — never a
    > `fork`, and never an agent that edits files or commits. Each dispatch's model and effort come
-   > from `plan-agent-pair.sh <your Key> <n>`, `<n>` counting your planning dispatches from 1, and
-   > every dispatch prompt carries the line `Role: explorer` (an Explore agent) or
-   > `Role: plan-designer` (a Plan agent) and the two sentences of **Propagate this**
+   > from `plan-agent-pair.sh <your Key> <n> <role>`, `<n>` counting your planning dispatches from
+   > 1 and `<role>` that dispatch's own Role, and every dispatch prompt carries the line
+   > `Role: explorer` (an Explore agent) or `Role: plan-designer` (a Plan agent) and the two sentences of **Propagate this**
    > (`~/.claude/rules/agent-baseline.md`) verbatim. The agents investigate and propose; you
    > decide, and every choice is auto-resolved to its recommended option. Fixing, implementing and
    > reviewing are never delegated.
diff --git a/skills/flow/brainstorm-planner.md b/skills/flow/brainstorm-planner.md
index 7f7368e3..c2c51bf3 100644
--- a/skills/flow/brainstorm-planner.md
+++ b/skills/flow/brainstorm-planner.md
@@ -81,9 +81,10 @@ the change branch** (`skills/flow-contracts/finish-contract-run1.md`).
 
 ### Explore and design agents
 
-**Every planning dispatch** takes its pair from `plan-agent-pair.sh <seed> <n>` — `<seed>` the
+**Every planning dispatch** takes its pair from `plan-agent-pair.sh <seed> <n> <role>` — `<seed>` the
 change name (in `/flow-plan` before capture, the session token; for the self-review and in-run fix
-planners, their own `Key:`), `<n>` counting the session's planning dispatches from 1 — passing the
+planners, their own `Key:`), `<n>` counting the session's planning dispatches from 1, `<role>` the
+dispatch's own `Role:` line, `explorer` or `plan-designer`, whose row the hook admits it on — passing the
 printed model as the Agent tool's `model` and the effort as its `effort`. Its prompt carries the
 line `Role: explorer` for an Explore agent or `Role: plan-designer` for a Plan agent (**Allowed
 pairs**, `skills/flow-contracts/model-policy.md`), the
diff --git a/skills/flow/brainstorm.md b/skills/flow/brainstorm.md
index 8cb8ad7e..dd976e97 100644
--- a/skills/flow/brainstorm.md
+++ b/skills/flow/brainstorm.md
@@ -124,7 +124,7 @@ flow record dispatch end -change <name> -key plan-<n> -session-token mf-<literal
 1. **Design round.** The planner runs section **B** of `skills/flow/brainstorm-planner.md` without <!-- refs-guard:allow -->
    plan mode, reading the ticket and every file under `<abs-worktree>/.superpowers/sdd/planner/`,
    dispatching its own Explore and Plan agents (it records their rows, `-role explorer|planner`,
-   `-key explore-<n>|design-<n>`, each pair from `plan-agent-pair.sh <name> <n>`, `<n>` continuing
+   `-key explore-<n>|design-<n>`, each pair from `plan-agent-pair.sh <name> <n> explorer|plan-designer`, `<n>` continuing
    the run's count across rounds — the parent names the next free `<n>` in each round's prompt, since
    a reused key restarts the earlier round's row), publishing or
    updating the prototype a visible change needs, and writes `round-<r>.md`: the design draft, the
diff --git a/stats/internal/guard/plan_agent_pair_test.go b/stats/internal/guard/plan_agent_pair_test.go
index b5c526ca..384699e7 100644
--- a/stats/internal/guard/plan_agent_pair_test.go
+++ b/stats/internal/guard/plan_agent_pair_test.go
@@ -8,8 +8,8 @@ import (
 	"testing"
 )
 
-// TestPlanAgentPair pins scripts/plan-agent-pair.sh: every seed to `opus low`
-// and every usage refusal.
+// TestPlanAgentPair pins scripts/plan-agent-pair.sh: every seed and both roles
+// to `opus low`, and every usage refusal.
 func TestPlanAgentPair(t *testing.T) {
 	root, err := filepath.Abs("../../..")
 	if err != nil {
@@ -18,12 +18,14 @@ func TestPlanAgentPair(t *testing.T) {
 	env := envWith(t, []string{"FLOW_GUARD_REPO_ROOT=" + root})
 	t.Run("always opus low", func(t *testing.T) {
 		for i := range 100 {
-			if r := runGuard("plan-agent-pair", []string{fmt.Sprintf("s%d", i), "1"}, env); r.rc != 0 || r.stdout != "opus low\n" {
-				t.Fatalf("s%d 1: rc=%d out=%q, want opus low", i, r.rc, r.out)
+			for _, role := range []string{"explorer", "plan-designer"} {
+				if r := runGuard("plan-agent-pair", []string{fmt.Sprintf("s%d", i), "1", role}, env); r.rc != 0 || r.stdout != "opus low\n" {
+					t.Fatalf("s%d 1 %s: rc=%d out=%q, want opus low", i, role, r.rc, r.out)
+				}
 			}
 		}
 	})
-	for _, args := range [][]string{{"demo", "0"}, {"demo", "x"}, {"demo"}} {
+	for _, args := range [][]string{{"demo", "0", "explorer"}, {"demo", "x", "explorer"}, {"demo", "1"}, {"demo", "1", "planner"}, {"demo"}} {
 		t.Run(fmt.Sprintf("refuses %v", args), func(t *testing.T) {
 			if r := runGuard("plan-agent-pair", args, env); r.rc != 2 || r.stdout != "" {
 				t.Errorf("rc=%d out=%q, want 2 and no stdout", r.rc, r.out)
@@ -45,23 +47,32 @@ func TestPlanAgentPairFollowsPolicy(t *testing.T) {
 		t.Fatal("the real table carries no plan-designer row to rewrite")
 	}
 	writeFile(t, repo+"/"+mpFile, table)
-	if r := runGuard("plan-agent-pair", []string{"demo", "1"}, envWith(t, []string{"FLOW_GUARD_REPO_ROOT=" + repo})); r.rc != 0 || r.stdout != "opus medium\n" {
+	if r := runGuard("plan-agent-pair", []string{"demo", "1", "plan-designer"}, envWith(t, []string{"FLOW_GUARD_REPO_ROOT=" + repo})); r.rc != 0 || r.stdout != "opus medium\n" {
 		t.Fatalf("rc=%d out=%q, want opus medium", r.rc, r.out)
 	}
-	if r := runGuard("plan-agent-pair", []string{"demo", "1"}, envWith(t, []string{"FLOW_GUARD_REPO_ROOT=" + t.TempDir()})); r.rc != 2 {
+	if r := runGuard("plan-agent-pair", []string{"demo", "1", "plan-designer"}, envWith(t, []string{"FLOW_GUARD_REPO_ROOT=" + t.TempDir()})); r.rc != 2 {
 		t.Fatalf("rc=%d out=%q, want 2 on an unreadable table", r.rc, r.out)
 	}
 }
 
-// TestPlanAgentPairReadsOverride: an open plan-designer override decides the
-// planning pair, xhigh included.
+// experimentOneFlow is open experiment 1, plan-designer-xhigh-explorer-low:
+// each planning role overridden to its own pair.
+const experimentOneFlow = `[ "$1 $2" = "experiment overrides" ] || exit 9
+printf '%s\n' '{"plan-designer":["opus/Plan@xhigh"],"explorer":["sonnet/Explore@low"]}'
+`
+
+// TestPlanAgentPairReadsOverride: an open override of the named role's row
+// decides its pair, xhigh included, and never the other role's -- the pair
+// the agent-baseline hook admits for that Role line.
 func TestPlanAgentPairReadsOverride(t *testing.T) {
 	root, err := filepath.Abs("../../..")
 	if err != nil {
 		t.Fatal(err)
 	}
-	env := fakeFlowEnv(t, overridesFlow, "FLOW_GUARD_REPO_ROOT="+root)
-	if r := runGuard("plan-agent-pair", []string{"demo", "1"}, env); r.rc != 0 || r.stdout != "opus xhigh\n" {
-		t.Fatalf("rc=%d out=%q, want opus xhigh", r.rc, r.out)
+	env := fakeFlowEnv(t, experimentOneFlow, "FLOW_GUARD_REPO_ROOT="+root)
+	for role, want := range map[string]string{"plan-designer": "opus xhigh\n", "explorer": "sonnet low\n"} {
+		if r := runGuard("plan-agent-pair", []string{"demo", "1", role}, env); r.rc != 0 || r.stdout != want {
+			t.Fatalf("%s: rc=%d out=%q, want %q", role, r.rc, r.out, want)
+		}
 	}
 }
diff --git a/stats/internal/guard/planagentpair.go b/stats/internal/guard/planagentpair.go
index f88f065b..ee2edf7b 100644
--- a/stats/internal/guard/planagentpair.go
+++ b/stats/internal/guard/planagentpair.go
@@ -9,9 +9,10 @@ import (
 
 // planAgentPair is scripts/plan-agent-pair.sh: the model and effort of a
 // planning session's <n>th Explore or Plan dispatch -- the first pair of the
-// plan-designer row of the allowed-pairs table (skills/flow-contracts/
-// model-policy.md), read from FLOW_GUARD_REPO_ROOT, which the shim sets, with
-// an open experiment's plan-designer override applied (policyTable).
+// <role> row, explorer or plan-designer, of the allowed-pairs table
+// (skills/flow-contracts/model-policy.md), read from FLOW_GUARD_REPO_ROOT,
+// which the shim sets, with an open experiment's override of that row applied
+// (policyTable), as the agent-baseline hook admits it.
 // <seed> is the change name, or the session token in /flow-plan before
 // capture; <n> counts the session's planning dispatches from 1. Exit 0 on the
 // printed line, exit 2 on a usage error or an unreadable table.
@@ -22,8 +23,13 @@ func init() {
 var papPair = regexp.MustCompile(`^([a-z0-9.-]+)/[A-Za-z]+@([a-z]+)$`)
 
 func planAgentPair(args []string, env Env, stdout, stderr io.Writer) int {
-	if len(args) != 2 {
-		fmt.Fprintln(stderr, "usage: plan-agent-pair.sh <seed> <n>")
+	if len(args) != 3 {
+		fmt.Fprintln(stderr, "usage: plan-agent-pair.sh <seed> <n> <explorer|plan-designer>")
+		return 2
+	}
+	role := args[2]
+	if role != "explorer" && role != "plan-designer" {
+		fmt.Fprintf(stderr, "plan-agent-pair.sh: <role> must be explorer or plan-designer, got %q\n", role)
 		return 2
 	}
 	if n, err := strconv.Atoi(args[1]); err != nil || n < 1 {
@@ -36,11 +42,11 @@ func planAgentPair(args []string, env Env, stdout, stderr io.Writer) int {
 		return 2
 	}
 	var m []string
-	if rs := rows["plan-designer"]; len(rs) > 0 {
+	if rs := rows[role]; len(rs) > 0 {
 		m = papPair.FindStringSubmatch(rs[0].Pairs[0])
 	}
 	if m == nil {
-		fmt.Fprintf(stderr, "plan-agent-pair.sh: %s names no `<model>/Plan@<effort>` plan-designer pair\n", mpFile)
+		fmt.Fprintf(stderr, "plan-agent-pair.sh: %s names no `<model>/<agent>@<effort>` %s pair\n", mpFile, role)
 		return 2
 	}
 	fmt.Fprintln(stdout, m[1]+" "+m[2])
````

  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean and
    `go test ./internal/guard -run TestPlanAgentPair` passes; `scripts/check-references.sh`, `scripts/check-dispatch-paragraphs.sh`
    and `scripts/check-guard-symlinks.sh` exit 0; `scripts/plan-agent-pair.sh x 1 explorer` prints
    the explorer row's pair (`sonnet low` while experiment 1 is open), `x 1 plan-designer` the
    plan-designer row's, and `x 1` exits 2 with the new usage line.
  - [x] **Step 4: Commit** the seven files with the subject below.

**Build:** green
**Files:** `scripts/plan-agent-pair.sh`, `skills/flow-plan/SKILL.md`, `skills/flow-self-review/SKILL.md`, `skills/flow/brainstorm-planner.md`, `skills/flow/brainstorm.md`, `stats/internal/guard/plan_agent_pair_test.go`, `stats/internal/guard/planagentpair.go`
**Tests:** `TestPlanAgentPair`, `TestPlanAgentPairFollowsPolicy`, `TestPlanAgentPairReadsOverride`
**Regression:** TestPlanAgentPairReadsOverride — reverting the commit prints the plan-designer pair for every role: `explorer` gets `opus xhigh` where experiment 1 gives `sonnet low`, the pair the hook refused.
**Baseline:** before=3 after=3
<!-- measured: grep -cE '^func Test' stats/internal/guard/plan_agent_pair_test.go @ e6c5c19e → 3; after with this task's diff applied → 3 -->
**Commit:** fix(scripts): print the planning pair of the dispatch's own role
**After:** none

**Decision:** role-argument
