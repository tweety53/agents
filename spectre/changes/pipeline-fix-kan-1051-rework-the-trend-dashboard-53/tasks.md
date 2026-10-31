# pipeline-fix-kan-1051-rework-the-trend-dashboard-53

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** `check-unfinished-work` finds a dropped re-run's already-ticked appended tasks when the drop commit removed the plan.

**Architecture:** one fallback `git show <base>^:<plan>` in signal four's commit half.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; a diff that does not apply is reported, never hand-merged by guesswork.
- No new inline suppressions; never stop flowd or flow-postgres.

## Review Focus

- The fallback runs only when the base read fails, so a base that holds the plan is read unchanged (`read-plan-at-drop-parent`).

- [ ] 1. Read a dropped re-run's plan at the drop commit's parent
  - [ ] **Step 1: Reproduce.** Save the diff below to `/tmp/pf53-task1.diff`, apply the test hunk alone
    (`git apply --include='stats/internal/guard/*_test.go' /tmp/pf53-task1.diff`) and run
    `cd stats && go test ./internal/guard -run TestCheckUnfinishedWork`: FAIL on
    `an_appended_task_ticked_in_the_plan_at_the_parent_of_a_drop_commit_that_removed_it_needs_no_new_commit`.
  - [ ] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/guard/*_test.go' /tmp/pf53-task1.diff`):

````diff verified:applied to this worktree at df1564a9; go test ./internal/guard -run UnfinishedWork passes with it, the new case fails with unfinishedwork.go reverted, go vet and gofmt -l exit clean
diff --git a/stats/internal/guard/check_unfinished_work_test.go b/stats/internal/guard/check_unfinished_work_test.go
index 59fd7789..5ae5c302 100644
--- a/stats/internal/guard/check_unfinished_work_test.go
+++ b/stats/internal/guard/check_unfinished_work_test.go
@@ -406,6 +406,19 @@ func TestCheckUnfinishedWork(t *testing.T) {
 		gitRun(t, f.wt, "commit", "-q", "--allow-empty", "-m", "fix: no trailer")
 		return f.run(t, f.wt, "demo", "", f.wt+"="+base)
 	}, uwVerdict("an appended task already ticked in the plan at a dropped re-run's base needs no new commit", "CLEAR:"))
+	each(func(t *testing.T) uwResult {
+		f := newUW(t)
+		f.plan(t, "> **Tasks appended:** 1\n\n- [x] 1. done\n- [x] 7. appended by a fix run\n")
+		f.setFile(t, "dispatches.json", covered+"\n")
+		gitRun(t, "", "init", "-q", f.wt)
+		gitRun(t, f.wt, "add", "spectre/changes/demo/tasks.md")
+		gitRun(t, f.wt, "commit", "-q", "-m", "fix: the appended task, squashed")
+		gitRun(t, f.wt, "rm", "-q", "--cached", "spectre/changes/demo/tasks.md")
+		gitRun(t, f.wt, "commit", "-q", "-m", "chore(spectre): drop demo")
+		base := tcfGit(t, f.wt, "rev-parse", "HEAD")
+		gitRun(t, f.wt, "commit", "-q", "--allow-empty", "-m", "fix: no trailer")
+		return f.run(t, f.wt, "demo", "", f.wt+"="+base)
+	}, uwVerdict("an appended task ticked in the plan at the parent of a drop commit that removed it needs no new commit", "CLEAR:"))
 	each(func(t *testing.T) uwResult {
 		f := newUW(t)
 		f.plan(t, "> **Tasks appended:** 1\n\n- [x] 7. appended by a fix run\n")
diff --git a/stats/internal/guard/unfinishedwork.go b/stats/internal/guard/unfinishedwork.go
index 67e0dcb7..2fe6e14b 100644
--- a/stats/internal/guard/unfinishedwork.go
+++ b/stats/internal/guard/unfinishedwork.go
@@ -793,9 +793,15 @@ func checkUnfinishedWork(args []string, env Env, stdout, stderr io.Writer) int {
 					// A dropped re-run's range starts at the drop commit: run 1
 					// squashed the earlier task commits, trailers and all, and
 					// its gate already judged every task ticked in the plan
-					// as it stood there.
+					// as it stood there -- at its parent when the drop commit
+					// removed the change directory (kan-1051).
 					if rel, err := filepath.Rel(r[0], primary); err == nil && !strings.HasPrefix(rel, "..") {
-						if old, ok := capture(envGit(env)("-C", r[0], "show", r[1]+":./"+filepath.ToSlash(rel))); ok {
+						path := ":./" + filepath.ToSlash(rel)
+						old, ok := capture(envGit(env)("-C", r[0], "show", r[1]+path))
+						if !ok {
+							old, ok = capture(envGit(env)("-C", r[0], "show", r[1]+"^"+path))
+						}
+						if ok {
 							for _, id := range uwAppendedIn(old, true) {
 								committed[id] = true
 							}
````

  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/guard -run 'UnfinishedWork'` passes;
    `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean;
    `scripts/check-references.sh` exits 0.
  - [ ] **Step 4: Commit** `fix(stats): read a dropped re-run's plan at the drop commit's parent`.

**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`
**Tests:** `TestCheckUnfinishedWork`
**Regression:** `TestCheckUnfinishedWork` — reverting the commit makes the drop-removes-the-plan case OUTSTANDING.
**Baseline:** before=16 after=16 <!-- measured: grep -c '^func Test' stats/internal/guard/check_unfinished_work_test.go @ df1564a9 → 16 -->
**After:** none
**Commit:** fix(stats): read a dropped re-run's plan at the drop commit's parent
**Build:** green
