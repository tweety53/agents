# pipeline-fix-kan-954-user-row-loaded-twice-per-request-9

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a self-review fix loop on a change's own branch admits `review-1` on the effort its plan
recorded, read from `spectre/changes/self-review-<name>/design.md` on `spectre/<name>` or, on
`/flow-fast`, `<name>`.

**Architecture:** one function, `lpFirstReviewType` in `stats/internal/guard/loop.go`, reads the
plan directory named after the loop's token and tries the own-branch refs in order.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:`
  trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; it was applied to this worktree at 711c69c9 and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- No new inline suppressions; `cd stats && gofmt -l .` prints nothing.

## Review Focus

- A pipeline-fix loop and a `self-review-<name>` loop branch read exactly the path they read
  before (`lp.token == lp.change` for both) — every existing `TestCheckLoopDispatch` case.
- The ref loop stops at the first existing branch in the order loop branch, `spectre/<name>`,
  `<name>`, and falls back to `flow-medium` when none carries the plan.

- [x] 1. Read an own-branch self-review's first-review effort from the loop's plan
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save the diff below to
    `/tmp/pf9-task1.diff`, apply its test hunks alone
    (`git apply --include='stats/internal/guard/*_test.go' /tmp/pf9-task1.diff`) and run
    `cd stats && go test ./internal/guard/ -run 'OwnBranchInAnotherProject|OwnFastBranchReadsLoopPlan' -count=1`:
    both tests FAIL with `LOOP-DISPATCH-REFUSED: PAIR — self-review-kan-1-demo-review-1 runs on
    flow-medium, as the loop's design.md first-review-effort decision records (flow-medium when
    none); this dispatch carries "flow-low"`.
  - [x] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/guard/*_test.go' /tmp/pf9-task1.diff`):

````diff verified:applied with git apply to this worktree at 711c69c9; both tests fail with loop.go untouched and pass with it; go test ./internal/guard/ -count=1 green; gofmt -l and go vet clean
diff --git a/stats/internal/guard/check_loop_dispatch_test.go b/stats/internal/guard/check_loop_dispatch_test.go
index 95482e83..1d43695c 100644
--- a/stats/internal/guard/check_loop_dispatch_test.go
+++ b/stats/internal/guard/check_loop_dispatch_test.go
@@ -183,7 +183,8 @@ func TestCheckLoopDispatchUnreadableTable(t *testing.T) {
 // A self-review on a change's own branch in another project -- no
 // self-review-<name> branch in <agents repo> -- files its rows in that
 // project's flowd project: flow runs with -C the caller's main checkout, and
-// review-1's first-review-effort is read from that project's spectre/<name>.
+// review-1's first-review-effort is read from the loop's own plan,
+// spectre/changes/self-review-<name>/, on that project's spectre/<name>.
 // A self-review-<name> loop branch in <agents repo> still files there.
 func TestCheckLoopDispatchOwnBranchInAnotherProject(t *testing.T) {
 	t.Parallel()
@@ -194,7 +195,7 @@ func TestCheckLoopDispatchOwnBranchInAnotherProject(t *testing.T) {
 	gitRun(t, proj, "add", "-A")
 	gitRun(t, proj, "commit", "-q", "-m", "init")
 	gitRun(t, proj, "checkout", "-q", "-b", "spectre/kan-1-demo")
-	writeFile(t, proj+"/spectre/changes/kan-1-demo/design.md", lpDesignLow)
+	writeFile(t, proj+"/spectre/changes/self-review-kan-1-demo/design.md", lpDesignLow)
 	gitRun(t, proj, "add", "-A")
 	gitRun(t, proj, "commit", "-q", "-m", "plan")
 	gitRun(t, proj, "checkout", "-q", "develop")
@@ -212,3 +213,18 @@ func TestCheckLoopDispatchOwnBranchInAnotherProject(t *testing.T) {
 		t.Fatalf("a self-review loop branch left <agents repo>: rc %d: %s\n%s", rc, out, fx.calls(t))
 	}
 }
+
+// A self-review on a /flow-fast change's own branch, <name> with no spectre/
+// prefix, reads review-1's first-review-effort from the loop's plan there.
+func TestCheckLoopDispatchOwnFastBranchReadsLoopPlan(t *testing.T) {
+	t.Parallel()
+	fx := lpNew(t, lpRowsJSON("self-review-kan-1-demo-", "self-review-kan-1-demo", "plan completed", "fix completed"))
+	gitRun(t, fx.repo, "checkout", "-q", "-b", "kan-1-demo")
+	writeFile(t, fx.repo+"/spectre/changes/self-review-kan-1-demo/design.md", lpDesignLow)
+	gitRun(t, fx.repo, "add", "-A")
+	gitRun(t, fx.repo, "commit", "-q", "-m", "plan")
+	gitRun(t, fx.repo, "checkout", "-q", "main")
+	if rc, out := fx.run(t, "self-review-kan-1-demo-review-1", "flow-low", "opus"); rc != 0 {
+		t.Fatalf("rc %d: %s", rc, out)
+	}
+}
diff --git a/stats/internal/guard/loop.go b/stats/internal/guard/loop.go
index 570978ae..b53d6eed 100644
--- a/stats/internal/guard/loop.go
+++ b/stats/internal/guard/loop.go
@@ -268,14 +268,19 @@ var (
 
 // lpFirstReviewType is the subagent_type the planner recorded for review-1 in
 // the loop's design.md (its `first-review-effort` decision's **Chosen:**
-// line), read from the loop's branch -- the change's own spectre/<name> for a
-// self-review on it; flow-medium when it recorded none.
+// line). The loop's plan is always spectre/changes/<loop branch name>/ --
+// self-review-<name> on the change's own branch too -- read from the loop's
+// branch, or for a self-review on the change's own branch from spectre/<name>
+// (/flow) or <name> (/flow-fast); flow-medium when it recorded none.
 func lpFirstReviewType(env Env, git, repo string, lp lpLoop) string {
 	ref := lp.token
-	if !lpBranchExists(env, git, repo, ref) {
-		ref = "spectre/" + lp.change
+	for _, b := range []string{"spectre/" + lp.change, lp.change} {
+		if lpBranchExists(env, git, repo, ref) {
+			break
+		}
+		ref = b
 	}
-	rc, body := kwRun(env, "", git, "-C", repo, "show", "refs/heads/"+ref+":spectre/changes/"+lp.change+"/design.md")
+	rc, body := kwRun(env, "", git, "-C", repo, "show", "refs/heads/"+ref+":spectre/changes/"+lp.token+"/design.md")
 	if rc != 0 {
 		return "flow-medium"
 	}
````

  - [x] **Step 3: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard/ && go test ./internal/guard/ -run 'CheckLoopDispatch|OwnFastBranchReadsLoopPlan|PipelineFix' -count=1`
    — gofmt prints nothing, the rest exit 0.
  - [x] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/loop.go`, `stats/internal/guard/check_loop_dispatch_test.go`
**Tests:** `TestCheckLoopDispatchOwnFastBranchReadsLoopPlan`
**Regression:** reverting the commit reads `spectre/changes/<name>/design.md` for an own-branch loop and never tries `<name>` — `TestCheckLoopDispatchOwnFastBranchReadsLoopPlan` and the corrected `TestCheckLoopDispatchOwnBranchInAnotherProject` fail with the `PAIR — … runs on flow-medium` refusal.
**Baseline:** before=4 after=5
<!-- measured: grep -cE '^func Test' stats/internal/guard/check_loop_dispatch_test.go (before at 711c69c9; after with this task's diff applied) -->
**Commit:** fix(guard): read an own-branch self-review's first-review effort from the loop's plan
**After:** none

**Decision:** plan-dir-is-the-token
**Decision:** fast-own-branch-fallback
