# pipeline-fix-kan-954-user-row-loaded-twice-per-request-13

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** the two "timeout counts as killed" cases of `TestMutateAndVerify` pass on a loaded
host: the bound under test fires once the mutated run is running, never on the wall clock, so
the baseline can never be the run that times out.

**Architecture:** `Env.MutateExpire`, a test-only expiry channel threaded from
`mutateAndVerify` through `mvTarget.run`, `mvRunHarness` and `mvRunCommand` into
`mvRunBounded`, replacing its timer when non-nil — the `Env.SurvivorsExpire` pattern.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:`
  trailer; never write `spectre/changes/`.
- The change is the unified diff Step 2 gives, applied from the repository root with
  `git apply`; it was applied to this worktree at 4963d4a6 and checked there. A diff that does
  not apply is reported, never hand-merged by guesswork.
- Never raise a timeout, a sleep or a wall-clock threshold to make a case pass.
- No new inline suppressions; `cd stats && gofmt -l ./internal/guard` prints nothing.

## Review Focus

- Production unchanged: with `Env.MutateExpire` nil, `mvRunBounded` waits on a
  `time.NewTimer(timeout)` exactly where it waited on `time.After(timeout)`; the timer is
  stopped on return.
- The marker is touched only on the mutated branch and lives outside the fixture repository
  (`<dir>/mutated`, the repository being `<dir>/repo`), so the restore and status checks are
  unaffected.
- `baseline pass timeout cannot answer` still runs on the real 3s timer.

- [x] 1. Fire the mutate-and-verify bound on readiness in its timeout tests
  - [x] **Step 1: Reproduce before touching anything; do not guess.** On the untouched tree,
    induce a slow baseline: in `stats/internal/guard/mutate_and_verify_test.go`, case
    `project command timeout counts as killed`, change the script's tail
    `sleep 30; fi\nexit 0\n` to `sleep 30; fi\nsleep 4\nexit 0\n`, then run
    `cd stats && go test ./internal/guard/ -run 'TestMutateAndVerify/project_command_timeout_counts_as_killed' -count=1`:
    it FAILS with `cannot answer — harness ./project-cmd.sh timed out after 3s on the baseline
    pass`. Revert the induction (`git checkout -- stats/internal/guard/mutate_and_verify_test.go`).
  - [x] **Step 2: Implement.** Save the diff below to `/tmp/pf13-task1.diff` and apply it
    (`git apply /tmp/pf13-task1.diff`):

````diff verified:applied with git apply to this worktree at 4963d4a6; go test ./internal/guard/ -run TestMutateAndVerify -count=1 green; 8 of 8 -race runs green under 40 yes processes (1 of 4 failed before); the induced 4s baseline passes; with expire: env.MutateExpire replaced by expire: nil both cases fail; gofmt -l and go vet clean
diff --git a/stats/internal/guard/guard.go b/stats/internal/guard/guard.go
index 689acac8..3e930e30 100644
--- a/stats/internal/guard/guard.go
+++ b/stats/internal/guard/guard.go
@@ -22,6 +22,10 @@ type Env struct {
 	// SurvivorsExpire, when non-nil, replaces the survivors bound's timer:
 	// the bound fires when it receives, so a test decides when (or never).
 	SurvivorsExpire <-chan time.Time
+	// MutateExpire, when non-nil, replaces mutate-and-verify's per-run
+	// timer: the running pass is killed as timed out when it receives, so a
+	// test decides which pass hits the bound and the baseline never races it.
+	MutateExpire <-chan time.Time
 	// ProcTable, when non-nil, replaces run-reproducer's process-table read
 	// (rrProcTable), so a test can make the read fail.
 	ProcTable func() ([]rrProc, error)
diff --git a/stats/internal/guard/mutate_and_verify_test.go b/stats/internal/guard/mutate_and_verify_test.go
index 3c1caffc..cacb5812 100644
--- a/stats/internal/guard/mutate_and_verify_test.go
+++ b/stats/internal/guard/mutate_and_verify_test.go
@@ -235,6 +235,16 @@ func (c *mvCase) run(vars map[string]string, args ...string) mvRes {
 	return mvRunIn(c.repo, vars, args...)
 }
 
+// runExpire is run with Env.MutateExpire set: the bound fires when expire
+// receives, never on the wall clock, so a slow baseline cannot hit it.
+func (c *mvCase) runExpire(expire <-chan time.Time, args ...string) mvRes {
+	env := crEnv(c.repo, nil)
+	env.MutateExpire = expire
+	var out, errb bytes.Buffer
+	rc := mutateAndVerify(args, env, &out, &errb)
+	return mvRes{rc, out.String(), errb.String()}
+}
+
 func mvRunIn(dir string, vars map[string]string, args ...string) mvRes {
 	var out, errb bytes.Buffer
 	rc := mutateAndVerify(args, crEnv(dir, vars), &out, &errb)
@@ -544,14 +554,15 @@ func TestMutateAndVerify(t *testing.T) {
 			mvParity(t, fx, map[string]string{"MUTATE_AND_VERIFY_MAX_NEW_FAILURES": "99999999999999999999"}, "")
 		}},
 		{"mutated pass timeout counts as killed", func(t *testing.T) {
-			c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
+			dir := t.TempDir()
+			c := fx.newFixture(t, dir, 4, "4/flip:2")
 			h := c.repo + "/sleepy.sh"
 			writeExec(t, h, "#!/usr/bin/env bash\necho 'ok: case-1'\n"+
-				"if grep -q '"+mvMutationMark+"' guard.sh; then sleep 30; fi\n")
+				"if grep -q '"+mvMutationMark+"' guard.sh; then touch '"+dir+"/mutated'; sleep 30; fi\n")
 			start := time.Now()
-			r := c.run(map[string]string{"MUTATE_AND_VERIFY_TIMEOUT_SECONDS": "3"}, c.patch, h)
+			r := c.runExpire(ccExpireWhen(dir+"/mutated"), c.patch, h)
 			mvCheck(t, "mutated timeout: caught, exit 0", r.rc == 0 && has(r.all(), "caught"), "rc=%d out=%s", r.rc, r.all())
-			mvCheck(t, "mutated timeout: counted as killed", has(r.all(), "timed out after 3s — counted as killed"), "%s", r.all())
+			mvCheck(t, "mutated timeout: counted as killed", has(r.all(), "timed out after 5m0s — counted as killed"), "%s", r.all())
 			mvCheck(t, "mutated timeout: bounded", time.Since(start) < 20*time.Second, "took %s", time.Since(start))
 			mvCheck(t, "mutated timeout: touched file clean after restore",
 				mvGit(t, c.repo, "status", "--porcelain", "--", "guard.sh") == "", "dirty")
@@ -592,11 +603,12 @@ func TestMutateAndVerify(t *testing.T) {
 			mvCheck(t, "PATH command: caught, exit 0", r.rc == 0 && has(r.all(), "caught"), "rc=%d out=%s", r.rc, r.all())
 		}},
 		{"project command timeout counts as killed", func(t *testing.T) {
-			c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
+			dir := t.TempDir()
+			c := fx.newFixture(t, dir, 4, "4/flip:2")
 			cmdPath := c.repo + "/project-cmd.sh"
 			writeExec(t, cmdPath, "#!/usr/bin/env bash\n"+
-				"if grep -q '"+mvMutationMark+"' guard.sh; then sleep 30; fi\nexit 0\n")
-			r := c.run(map[string]string{"MUTATE_AND_VERIFY_TIMEOUT_SECONDS": "3"}, c.patch, "--", "./project-cmd.sh")
+				"if grep -q '"+mvMutationMark+"' guard.sh; then touch '"+dir+"/mutated'; sleep 30; fi\nexit 0\n")
+			r := c.runExpire(ccExpireWhen(dir+"/mutated"), c.patch, "--", "./project-cmd.sh")
 			mvCheck(t, "command timeout: caught as killed", r.rc == 0 && has(r.all(), "counted as killed", "caught"), "rc=%d out=%s", r.rc, r.all())
 		}},
 		{"project command missing refused", func(t *testing.T) {
diff --git a/stats/internal/guard/mutateandverify.go b/stats/internal/guard/mutateandverify.go
index dbab78b6..807cd7dd 100644
--- a/stats/internal/guard/mutateandverify.go
+++ b/stats/internal/guard/mutateandverify.go
@@ -180,7 +180,7 @@ func mutateAndVerify(args []string, env Env, stdout, stderr io.Writer) int {
 	// ponytail: SIGINT restores at once, as SIGTERM does; bash waited for a
 	// running harness and carried on if that harness survived the SIGINT. A
 	// Ctrl-C reaches the harness too, so both end in the same restore.
-	r := &mvRun{root: root, touched: touched, snapshot: snapshot, stdout: stdout, stderr: stderr, timeout: mvTimeout(env)}
+	r := &mvRun{root: root, touched: touched, snapshot: snapshot, stdout: stdout, stderr: stderr, timeout: mvTimeout(env), expire: env.MutateExpire}
 	defer trapExitSignals(func() {
 		r.dying.Store(true)
 		r.exit(0)
@@ -201,11 +201,11 @@ type mvTarget struct {
 	argv    []string
 }
 
-func (t mvTarget) run(root string, timeout time.Duration) (ok int, fails []string, answered, timedOut bool) {
+func (t mvTarget) run(root string, timeout time.Duration, expire <-chan time.Time) (ok int, fails []string, answered, timedOut bool) {
 	if len(t.argv) == 0 {
-		return mvRunHarness(t.harness, root, timeout)
+		return mvRunHarness(t.harness, root, timeout, expire)
 	}
-	return mvRunCommand(t.argv, root, timeout)
+	return mvRunCommand(t.argv, root, timeout, expire)
 }
 
 // mvRun is the state the bash's EXIT trap read: whether the patch is applied,
@@ -215,6 +215,7 @@ type mvRun struct {
 	touched        []string
 	stdout, stderr io.Writer
 	timeout        time.Duration
+	expire         <-chan time.Time // Env.MutateExpire
 
 	mu      sync.Mutex // held across `git apply` and the restore, so neither interleaves
 	applied bool
@@ -236,7 +237,7 @@ func (r *mvRun) main(patchArg, patch string, targets []mvTarget, env Env, stdout
 	base := make([][]string, len(targets))
 	for i, t := range targets {
 		name := t.name
-		ok, fails, answered, timedOut := t.run(r.root, r.timeout)
+		ok, fails, answered, timedOut := t.run(r.root, r.timeout, r.expire)
 		if timedOut {
 			return cannotAnswer(fmt.Sprintf("harness %s timed out after %s on the baseline pass", name, r.timeout))
 		}
@@ -268,7 +269,7 @@ func (r *mvRun) main(patchArg, patch string, targets []mvTarget, env Env, stdout
 	mut := make([][]string, len(targets))
 	for i, t := range targets {
 		name := t.name
-		ok, fails, answered, timedOut := t.run(r.root, r.timeout)
+		ok, fails, answered, timedOut := t.run(r.root, r.timeout, r.expire)
 		if timedOut {
 			// A mutant that never finishes is killed by the bound, not
 			// left hanging: synthesise the new failure the verdict counts.
@@ -440,10 +441,11 @@ func mvTimeout(env Env) time.Duration {
 }
 
 // mvRunBounded runs cmd, killing its whole process group (SIGTERM, then
-// SIGKILL after 2s) once timeout elapses. The group kill is what reaches
+// SIGKILL after 2s) once timeout elapses -- or, when expire is non-nil, once
+// expire receives instead (Env.MutateExpire). The group kill is what reaches
 // a test executor the harness forked; WaitDelay keeps an escaped
 // descendant holding the capture pipe from hanging the wait.
-func mvRunBounded(cmd *exec.Cmd, timeout time.Duration) (timedOut bool) {
+func mvRunBounded(cmd *exec.Cmd, timeout time.Duration, expire <-chan time.Time) (timedOut bool) {
 	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
 	cmd.WaitDelay = 2 * time.Second
 	if err := cmd.Start(); err != nil {
@@ -451,10 +453,15 @@ func mvRunBounded(cmd *exec.Cmd, timeout time.Duration) (timedOut bool) {
 	}
 	done := make(chan struct{})
 	go func() { _ = cmd.Wait(); close(done) }()
+	if expire == nil {
+		bound := time.NewTimer(timeout)
+		defer bound.Stop()
+		expire = bound.C
+	}
 	select {
 	case <-done:
 		return false
-	case <-time.After(timeout):
+	case <-expire:
 	}
 	pid := cmd.Process.Pid
 	if syscall.Kill(-pid, syscall.SIGTERM) != nil {
@@ -472,7 +479,7 @@ func mvRunBounded(cmd *exec.Cmd, timeout time.Duration) (timedOut bool) {
 	return true
 }
 
-func mvRunHarness(h, dir string, timeout time.Duration) (ok int, fails []string, answered bool, timedOut bool) {
+func mvRunHarness(h, dir string, timeout time.Duration, expire <-chan time.Time) (ok int, fails []string, answered bool, timedOut bool) {
 	if timeout <= 0 {
 		timeout = mvDefaultTimeout
 	}
@@ -480,7 +487,7 @@ func mvRunHarness(h, dir string, timeout time.Duration) (ok int, fails []string,
 	cmd.Dir, cmd.Stdin = dir, os.Stdin
 	var out bytes.Buffer
 	cmd.Stdout, cmd.Stderr = &out, &out
-	timedOut = mvRunBounded(cmd, timeout)
+	timedOut = mvRunBounded(cmd, timeout, expire)
 	// $(...) dropped every NUL byte (bash 5 also warned "ignored null byte
 	// in input" on stderr, the one byte of its output not reproduced here).
 	for _, line := range strings.Split(strings.ReplaceAll(out.String(), "\x00", ""), "\n") {
@@ -499,7 +506,7 @@ func mvRunHarness(h, dir string, timeout time.Duration) (ok int, fails []string,
 // verdict: exit 0 synthesises one ok, a non-zero exit one failure named
 // for the exit code, so the shared verdict diff reads caught/surviving
 // exactly as it does for a harness's FAIL: lines.
-func mvRunCommand(argv []string, dir string, timeout time.Duration) (ok int, fails []string, answered bool, timedOut bool) {
+func mvRunCommand(argv []string, dir string, timeout time.Duration, expire <-chan time.Time) (ok int, fails []string, answered bool, timedOut bool) {
 	if timeout <= 0 {
 		timeout = mvDefaultTimeout
 	}
@@ -507,7 +514,7 @@ func mvRunCommand(argv []string, dir string, timeout time.Duration) (ok int, fai
 	cmd.Dir, cmd.Stdin = dir, os.Stdin
 	var out bytes.Buffer
 	cmd.Stdout, cmd.Stderr = &out, &out
-	timedOut = mvRunBounded(cmd, timeout)
+	timedOut = mvRunBounded(cmd, timeout, expire)
 	// The command's own output is captured and deliberately discarded:
 	// a Gradle log would flood the report, and its exit code,
 	// synthesised below, is the only evidence the verdict reads.
````

  - [x] **Step 3: Prove the seam.** Repeat Step 1's induction on the fixed tree (the tail now
    reads `sleep 30; fi\nexit 0\n` after `touch '"+dir+"/mutated';`): the case PASSES. Revert
    it. Then replace `expire: env.MutateExpire}` with `expire: nil}` in
    `stats/internal/guard/mutateandverify.go` and run
    `cd stats && go test ./internal/guard/ -run 'TestMutateAndVerify/(mutated_pass|project_command)_timeout_counts_as_killed' -count=1`:
    both cases FAIL after ~30s. Restore `expire: env.MutateExpire}`.
  - [x] **Step 4: Verify.** `cd stats && gofmt -l ./internal/guard && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestMutateAndVerify' -count=1`
    — green, `gofmt -l` prints nothing.
  - [x] **Step 5: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/guard.go`, `stats/internal/guard/mutateandverify.go`, `stats/internal/guard/mutate_and_verify_test.go`
**Tests:** none — repairs existing tests: TestMutateAndVerify subtests "mutated pass timeout counts as killed" and "project command timeout counts as killed"
**Regression:** reverting the commit puts both cases back on a live 3s bound shared with the baseline; Step 1's induced 4s baseline fails `project command timeout counts as killed` again with exit 4.
**Commit:** fix(guard): fire the mutate-and-verify bound on readiness in its timeout tests
**After:** none

**Decision:** inject-mutate-expire
**Decision:** one-bound-both-passes
**Decision:** drop-knob-in-expiry-cases
**Decision:** baseline-case-real-timer
