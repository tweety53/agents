# pipeline-fix-kan-1024-templates-warm-up-cool-down-any-workout-39

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** `scripts/close-task.sh` exports `FLOW_GUARD_SCRIPT_DIR`, so close-task's in-process zip
rebuild can run.

**Architecture:** four lines in `scripts/close-task.sh`; one test in
`stats/internal/guard/close_task_test.go` that runs the real shim.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; it was applied to this worktree at a9f37e89 and checked there. A diff that does not
  apply is reported, never hand-merged by guesswork.
- No new inline suppressions. Never stop flowd or flow-postgres.

## Review Focus

- The test's environment carries no `FLOW_GUARD_SCRIPT_DIR`, so only the shim can supply it
  (`shim-exports-script-dir`).
- The rebuild's `<worktree>` stays `absCanonical` (`rebuild-worktree-stays-canonical`).

- [ ] 1. Export FLOW_GUARD_SCRIPT_DIR from close-task.sh
  - [ ] **Step 1: Reproduce.** Save the diff below to `/tmp/pf39.diff`, apply its test hunk alone
    (`git apply --include='stats/*' /tmp/pf39.diff`), run
    `cd stats && make -s web-dist && go test -count=1 ./internal/guard/ -run TestCloseTaskShimExportsTheRebuildsScriptDir`:
    it fails with exit 2 and `FLOW_GUARD_SCRIPT_DIR is unset`.
  - [ ] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/*' /tmp/pf39.diff`):

````diff verified:applied with git apply to this worktree at a9f37e89; go test -count=1 ./internal/guard/ -run TestCloseTask passes with it and TestCloseTaskShimExportsTheRebuildsScriptDir fails (exit 2, FLOW_GUARD_SCRIPT_DIR is unset) with scripts/close-task.sh untouched; go vet and gofmt -l clean; scripts/check-references.sh exits 0
diff --git a/scripts/close-task.sh b/scripts/close-task.sh
index 0fed8fa5..10f555de 100755
--- a/scripts/close-task.sh
+++ b/scripts/close-task.sh
@@ -81,4 +81,8 @@ SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
   echo "close-task: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
   exit 2
 }
+# The zip rebuild runs in-process and finds resolve-visual-screenshots.sh here.
+: "$SCRIPT_DIR/resolve-visual-screenshots.sh"
+FLOW_GUARD_SCRIPT_DIR="$SCRIPT_DIR"
+export FLOW_GUARD_SCRIPT_DIR
 flow_guard_exec close-task 2 "close-task:" "$@"
diff --git a/stats/internal/guard/close_task_test.go b/stats/internal/guard/close_task_test.go
index b0547264..e94b1c11 100644
--- a/stats/internal/guard/close_task_test.go
+++ b/stats/internal/guard/close_task_test.go
@@ -648,3 +648,37 @@ func TestCloseTaskSuiteZipAgainstTheTaskCommit(t *testing.T) {
 		t.Fatalf("deleted zip: exit %d, want 0\n%s", r.rc, r.out)
 	}
 }
+
+// scripts/close-task.sh is close-task's only caller, and it runs the zip
+// rebuild in-process: the shim exports FLOW_GUARD_SCRIPT_DIR, which the
+// rebuild needs to find resolve-visual-screenshots.sh (kan-1024 Task 14:
+// unset, close-task stopped on exit 2 for every committed zip).
+func TestCloseTaskShimExportsTheRebuildsScriptDir(t *testing.T) {
+	t.Parallel()
+	if _, err := exec.LookPath("zip"); err != nil {
+		t.Skip("zip is not on PATH")
+	}
+	repo, err := filepath.Abs("../../..")
+	if err != nil {
+		t.Fatal(err)
+	}
+	fx := ctFixture(t, true, "0", "0")
+	writeFile(t, fx.wt+"/.flow/project.md", "# p\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `screenshots` | `shots` |\n")
+	writeFile(t, fx.wt+"/shots/full-app-suite.spec.ts-snapshots/a.png", "a")
+	gitRun(t, fx.wt, "add", "-f", ".flow", "shots")
+	gitRun(t, fx.wt, "commit", "-q", "-m", "shots")
+	fx.base = tcfGit(t, fx.wt, "rev-parse", "HEAD")
+	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", strings.Replace(ctPlan, "**Files:** `a.txt`", "**Files:** `a.txt`, `shots/full-app-suite.zip`", 1))
+	writeFile(t, fx.wt+"/a.txt", "a\n")
+	writeFile(t, fx.wt+"/shots/full-app-suite.zip", "hand-built")
+	gitRun(t, fx.wt, "add", "-f", "a.txt", "shots/full-app-suite.zip")
+	gitRun(t, fx.wt, "commit", "-q", "-m", "one", "-m", "Task-Id: 1")
+	s1 := tcfGit(t, fx.wt, "rev-parse", "HEAD")
+	cmd := exec.Command(filepath.Join(repo, "scripts", "close-task.sh"), fx.wt, "demo", fx.base, "1:"+s1)
+	cmd.Dir = fx.wt
+	cmd.Env = []string{"PATH=" + fx.env.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "FLOW_GUARD_TELEMETRY=off", "FLOW_GUARD_CACHE_DIR=" + guardCache(t)}
+	out, err := cmd.CombinedOutput()
+	if code := exitCode(err); code != 1 || !strings.Contains(string(out), "is not rebuild-full-app-suite-zip.sh's output") {
+		t.Fatalf("close-task.sh on a hand-built zip: exit %d, want 1 (the rebuild ran and differed)\n%s", code, out)
+	}
+}
````

  - [ ] **Step 3: Verify.** `cd stats && make -s web-dist && go test -count=1 ./internal/guard/ -run TestCloseTask`
    passes; `go vet ./...` and `gofmt -l .` are clean; `scripts/check-references.sh` exits 0.
  - [ ] **Step 4: Commit** the two files with the subject below.

**Build:** green
**Files:** `scripts/close-task.sh`, `stats/internal/guard/close_task_test.go`
**Tests:** `TestCloseTaskShimExportsTheRebuildsScriptDir`
**Regression:** TestCloseTaskShimExportsTheRebuildsScriptDir — reverting the shim export makes close-task.sh exit 2 on a committed zip: the test fails.
**Baseline:** before=10 after=11
<!-- measured: grep -c '^func Test' stats/internal/guard/close_task_test.go @ a9f37e89 → 10; the task adds one -->
**Commit:** fix(scripts): export FLOW_GUARD_SCRIPT_DIR from close-task.sh for the zip rebuild
**After:** none

**Decision:** shim-exports-script-dir
