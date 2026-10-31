# pipeline-fix-kan-1051-rework-the-trend-dashboard-51

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** `close-task.sh` closes an implementer group whose later commit rewrites the suite zip an earlier one wrote, and still refuses a hand-built zip.

**Architecture:** `ctCheckSuiteZips` gains the close's other task commits and a `superseded` predicate.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; a diff that does not apply is reported, never hand-merged by guesswork.
- No new inline suppressions; never stop flowd or flow-postgres.

## Review Focus

- The close's last zip-changing commit is never skipped, and a commit is never skipped for itself or for a commit in another worktree (`judge-last-zip-of-close`).

- [ ] 1. Judge only the last suite zip a close's commits write
  - [ ] **Step 1: Reproduce.** Save the diff below to `/tmp/pf51-task1.diff`, apply the test hunk alone
    (`git apply --include='stats/internal/guard/*_test.go' /tmp/pf51-task1.diff`) and run
    `cd stats && go test ./internal/guard -run TestCloseTaskJudgesTheLastSuiteZipOfAClose`: FAIL,
    `superseded script-built zip: exit 2, want 0` and `close-task: task 1: … is not the commit's`.
  - [ ] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/guard/*_test.go' /tmp/pf51-task1.diff`):

````diff verified:applied with git apply to this worktree at 50e0a1f7; go test ./internal/guard -run TestCloseTask passes with it, the new test fails with closetask.go reverted, and run-verify-list.sh lint exits clean but for check-plan-shape.sh on the then-empty plan
diff --git a/scripts/close-task.sh b/scripts/close-task.sh
index ea0d0b78..f37af76f 100755
--- a/scripts/close-task.sh
+++ b/scripts/close-task.sh
@@ -35,7 +35,9 @@
 #      A commit — any map pair's — whose diff-tree lists a path named
 #      full-app-suite.zip has that zip rebuilt in its worktree by
 #      rebuild-full-app-suite-zip.sh, the zip's directory as <screenshots
-#      root>, and the committed bytes written back: a difference prints
+#      root>, and the committed bytes written back — save a zip a later
+#      commit of the same close, in the same worktree, changes again, which
+#      is that commit's to judge (kan-1051): a difference prints
 #      `close-task: task <id>: <path> is not rebuild-full-app-suite-zip.sh's
 #      output — rebuild it with that script and re-commit` (exit 1); a
 #      screenshots root git status reports dirty, or a rebuild exit other
diff --git a/stats/internal/guard/close_task_test.go b/stats/internal/guard/close_task_test.go
index 74bd1e9c..95dfdebf 100644
--- a/stats/internal/guard/close_task_test.go
+++ b/stats/internal/guard/close_task_test.go
@@ -649,6 +649,74 @@ func TestCloseTaskSuiteZipAgainstTheTaskCommit(t *testing.T) {
 	}
 }
 
+// A close whose later commit re-records the screenshots and rebuilds the zip
+// judges that later commit's zip alone: the earlier one's screenshots root is
+// no longer the worktree's (kan-1051: tasks 6 and 7, exit 2 before). The
+// close's last zip is still judged, so a hand-built one is refused
+// (kan-1016's self-review).
+func TestCloseTaskJudgesTheLastSuiteZipOfAClose(t *testing.T) {
+	t.Parallel()
+	if _, err := exec.LookPath("zip"); err != nil {
+		t.Skip("zip is not on PATH")
+	}
+	repo, err := filepath.Abs("../../..")
+	if err != nil {
+		t.Fatal(err)
+	}
+	fx := ctFixture(t, true, "0", "0")
+	base := fx.env.Getenv
+	fx.env.Getenv = func(k string) string {
+		switch k {
+		case "FLOW_GUARD_SCRIPT_DIR":
+			return filepath.Join(repo, "scripts")
+		case "FLOW_GUARD_CACHE_DIR":
+			return guardCache(t)
+		}
+		return base(k)
+	}
+	writeFile(t, fx.wt+"/.flow/project.md", "# p\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `screenshots` | `shots` |\n")
+	writeFile(t, fx.wt+"/shots/full-app-suite.spec.ts-snapshots/a.png", "a")
+	gitRun(t, fx.wt, "add", "-f", ".flow", "shots")
+	gitRun(t, fx.wt, "commit", "-q", "-m", "shots")
+	fx.base = tcfGit(t, fx.wt, "rev-parse", "HEAD")
+	plan := strings.Replace(ctPlan, "**Files:** `a.txt`", "**Files:** `a.txt`, `shots/full-app-suite.zip`", 1)
+	plan = strings.Replace(plan, "**Files:** `b.txt`", "**Files:** `b.txt`, `shots/full-app-suite.zip`, `shots/full-app-suite.spec.ts-snapshots/a.png`\n**After:** Task 1", 1)
+	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", plan)
+	rebuild := func() {
+		t.Helper()
+		if rc := Registry["rebuild-full-app-suite-zip"]([]string{fx.wt, fx.wt + "/shots"}, fx.env, io.Discard, os.Stderr); rc != 0 {
+			t.Fatalf("rebuild exit %d", rc)
+		}
+	}
+	// Task 1 writes a script-built zip; task 2 re-records a.png and rebuilds it.
+	rebuild()
+	writeFile(t, fx.wt+"/a.txt", "a\n")
+	gitRun(t, fx.wt, "add", "-f", "a.txt", "shots/full-app-suite.zip")
+	gitRun(t, fx.wt, "commit", "-q", "-m", "one", "-m", "Task-Id: 1")
+	s1 := tcfGit(t, fx.wt, "rev-parse", "HEAD")
+	writeFile(t, fx.wt+"/shots/full-app-suite.spec.ts-snapshots/a.png", "b")
+	rebuild()
+	writeFile(t, fx.wt+"/b.txt", "b\n")
+	gitRun(t, fx.wt, "add", "-f", "b.txt", "shots")
+	gitRun(t, fx.wt, "commit", "-q", "-m", "two", "-m", "Task-Id: 2")
+	s2 := tcfGit(t, fx.wt, "rev-parse", "HEAD")
+	if r := fx.run(fx.wt, "demo", fx.base, "1:"+s1, "2:"+s2); r.rc != 0 {
+		t.Fatalf("superseded script-built zip: exit %d, want 0\n%s", r.rc, r.out)
+	}
+
+	// The close's last zip is still judged: a hand-built one is refused.
+	gitRun(t, fx.wt, "reset", "-q", "--hard", s1)
+	writeFile(t, fx.wt+"/shots/full-app-suite.spec.ts-snapshots/a.png", "b")
+	writeFile(t, fx.wt+"/shots/full-app-suite.zip", "hand-built")
+	writeFile(t, fx.wt+"/b.txt", "b\n")
+	gitRun(t, fx.wt, "add", "-f", "b.txt", "shots")
+	gitRun(t, fx.wt, "commit", "-q", "-m", "two", "-m", "Task-Id: 2")
+	s2 = tcfGit(t, fx.wt, "rev-parse", "HEAD")
+	if r := fx.run(fx.wt, "demo", fx.base, "1:"+s1, "2:"+s2); r.rc != 1 || !strings.Contains(r.out, "close-task: task 2: shots/full-app-suite.zip is not rebuild-full-app-suite-zip.sh's output") {
+		t.Fatalf("hand-built last zip: exit %d, want 1\n%s", r.rc, r.out)
+	}
+}
+
 // scripts/close-task.sh is close-task's only caller, and it runs the zip
 // rebuild in-process: the shim exports FLOW_GUARD_SCRIPT_DIR, which the
 // rebuild needs to find resolve-visual-screenshots.sh (kan-1024 Task 14:
diff --git a/stats/internal/guard/closetask.go b/stats/internal/guard/closetask.go
index 47f42cd7..4b3b6b00 100644
--- a/stats/internal/guard/closetask.go
+++ b/stats/internal/guard/closetask.go
@@ -129,7 +129,13 @@ func closeTask(args []string, env Env, stdout, stderr io.Writer) (exit int) {
 	for _, t := range tasks {
 		worst = max(worst, Registry["check-task-commit-fields"]([]string{canonical, t.id, t.commit, "", canonical, name}, env, stdout, stderr))
 		worst = max(worst, ctCheckTrailer(env, stdout, stderr, canonical, name, t.id, t.commit))
-		worst = max(worst, ctCheckSuiteZips(env, stderr, canonical, t.id, t.commit))
+		var others []string
+		for _, o := range tasks {
+			if o.id != t.id {
+				others = append(others, o.commit)
+			}
+		}
+		worst = max(worst, ctCheckSuiteZips(env, stderr, canonical, t.id, t.commit, others))
 	}
 	worst = max(worst, Registry["check-task-commit-planning-paths"]([]string{canonical, mergeBase}, env, stdout, stderr))
 	// The push below sends every commit origin lacks: one outside every task
@@ -420,13 +426,20 @@ func ctUndispatched(env Env, canonical, name string, closed []string) []string {
 // hand-built with archive/zip was committed). The rebuild's <worktree> is the
 // canonical worktree, whose .flow/project.md declares where screenshots live
 // — the zip may sit in another repository, the regression checkout. A commit
-// deleting the zip is not judged. 1 a difference; 2 a screenshots root that
-// differs from the commit's or a rebuild that did not answer 0.
-func ctCheckSuiteZips(env Env, stderr io.Writer, canonical, id, commit string) int {
-	pairs := []string{canonical + "=" + commit}
-	if strings.Contains(commit, "=") {
-		pairs = strings.Split(commit, ",")
+// deleting the zip is not judged, and neither is one a later commit of the
+// same close changes again — others holds the close's other task commits —
+// since the worktree holds only the last one's screenshots root (kan-1051:
+// tasks 6 and 7 both wrote the zip and closed together). 1 a difference; 2 a
+// screenshots root that differs from the commit's or a rebuild that did not
+// answer 0.
+func ctCheckSuiteZips(env Env, stderr io.Writer, canonical, id, commit string, others []string) int {
+	pairsOf := func(c string) []string {
+		if strings.Contains(c, "=") {
+			return strings.Split(c, ",")
+		}
+		return []string{canonical + "=" + c}
 	}
+	pairs := pairsOf(commit)
 	absOf := func(p string) string {
 		if filepath.IsAbs(p) {
 			return p
@@ -435,6 +448,21 @@ func ctCheckSuiteZips(env Env, stderr io.Writer, canonical, id, commit string) i
 	}
 	absCanonical := absOf(canonical)
 	git := envGit(env)
+	// superseded: a later commit of the close, in the same worktree, changes path too.
+	superseded := func(wt, sha, path string) bool {
+		for _, o := range others {
+			for _, op := range pairsOf(o) {
+				owt, osha, _ := strings.Cut(strings.TrimSpace(op), "=")
+				if osha == sha || absOf(owt) != absOf(wt) || git("-C", wt, "merge-base", "--is-ancestor", sha, osha).Run() != nil {
+					continue
+				}
+				if out, _ := capture(git("-C", wt, "diff-tree", "--no-renames", "--no-commit-id", "--name-only", "-r", osha, "--", path)); out != "" {
+					return true
+				}
+			}
+		}
+		return false
+	}
 	worst := 0
 	for _, p := range pairs {
 		wt, sha, _ := strings.Cut(strings.TrimSpace(p), "=")
@@ -443,7 +471,7 @@ func ctCheckSuiteZips(env Env, stderr io.Writer, canonical, id, commit string) i
 			continue // the fields guard already answers for an unreadable commit
 		}
 		for _, path := range strings.Split(out, "\n") {
-			if filepath.Base(path) != "full-app-suite.zip" {
+			if filepath.Base(path) != "full-app-suite.zip" || superseded(wt, sha, path) {
 				continue
 			}
 			abs := filepath.Join(absOf(wt), path)
````

  - [ ] **Step 3: Verify.** `cd stats && go test ./internal/guard -run 'TestCloseTask'` passes;
    `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean;
    `scripts/check-references.sh` exits 0.
  - [ ] **Step 4: Commit** `fix(stats): judge only the last suite zip a close's commits write`.

**Files:** `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go`, `scripts/close-task.sh`
**Tests:** `TestCloseTaskJudgesTheLastSuiteZipOfAClose`
**Regression:** `TestCloseTaskJudgesTheLastSuiteZipOfAClose` — reverting the commit makes the two-commit close exit 2 on task 1's zip.
**Baseline:** before=14 after=15 <!-- measured: grep -c '^func Test' stats/internal/guard/close_task_test.go @ 50e0a1f7 → 14 -->
**After:** none
**Commit:** fix(stats): judge only the last suite zip a close's commits write
**Build:** green
