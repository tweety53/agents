# pipeline-fix-kan-954-user-row-loaded-twice-per-request-8

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a self-review fix loop on a change's own branch in any project records its dispatches in
that project's flowd project, so the loop runs; and a loop tool's refusal is fixed in-run, never a
stop.

**Architecture:** `lpLoop.main` carries the checkout each loop's flow calls run with; an
own-branch self-review takes it from flow-guard's working directory, which the hook sets from the
payload's `cwd`. `pipeline-fix.sh verify` reads the worktree's own `origin/HEAD`. One sentence in
`skills/flow-self-review/SKILL.md`.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:`
  trailer; never write `spectre/changes/`.
- Each task's change is the unified diff its step gives, applied from the repository root with
  `git apply`; all three were applied to this worktree at b803ec00 and checked there. A diff that
  does not apply is reported, never hand-merged by guesswork.
- No new inline suppressions; `cd stats && gofmt -l .` prints nothing.

## Review Focus

- A pipeline-fix loop and a `self-review-<name>` loop branch still file under the agents checkout
  whatever the caller's working directory (`TestCheckLoopDispatchOwnBranchInAnotherProject`'s
  second half, and every existing case, whose `env.Dir` is no git checkout).
- `pfEnd`'s own-branch escalation clear and `pfEscalate`'s marker write use the same checkout
  `checkLoopDispatch` reads them from (`lp.main`), or an escalation is never seen.
- An own-branch self-review for an agents change run from the agents repository resolves to the
  agents main checkout exactly as before.

- [x] 1. File an own-branch self-review loop in the caller's project
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save the diff below to
    `/tmp/pf8-task1.diff`, apply its test hunks alone
    (`git apply --include='stats/internal/guard/*_test.go' /tmp/pf8-task1.diff`) and run
    `cd stats && go test ./internal/guard/ -run 'OwnBranchInAnotherProject' -count=1`: both tests
    FAIL — `end left the caller's project` shows `record dispatch end -C <…>/agents`, and the
    dispatch test reports the begin not filed in the caller's project.
  - [x] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='stats/internal/guard/*_test.go' /tmp/pf8-task1.diff`):

````diff verified:applied with git apply to this worktree at b803ec00; both new tests pass with it and fail with loop.go and pipelinefix.go reverted; go test ./internal/guard/ green
diff --git a/scripts/pipeline-fix.sh b/scripts/pipeline-fix.sh
index 2b98da09..12d27f39 100755
--- a/scripts/pipeline-fix.sh
+++ b/scripts/pipeline-fix.sh
@@ -5,7 +5,9 @@
 # retyped or skipped. The loop is its own change in this repository's flowd
 # project, named after its branch — pipeline-fix-<name>-<k> or
 # self-review-<name> — and check-loop-dispatch.sh files every loop dispatch
-# under it (a self-review on the change's own branch files under <name>).
+# under it (a self-review on the change's own branch files under <name>, in
+# the flowd project of the caller's working directory: end and escalate of
+# such a key run from the run's own project).
 #
 # Usage:
 #   pipeline-fix.sh start <name> [<k>]
@@ -30,7 +32,7 @@
 #       escalation for <key> and an earlier round of its kind completed.
 #   pipeline-fix.sh verify <worktree>
 #       runs the ## lint lines and tests the branch's touched paths need,
-#       against origin/<default>: every guard line, the `cd stats` lines for
+#       against <worktree>'s own origin/HEAD: every guard line, the `cd stats` lines for
 #       touched Go, the `cd stats/web` lines for a touched SPA file; go test
 #       of each touched stats package, scripts/test-<n>.sh or
 #       scripts/test-<n>-hook.sh of a touched script or hook, the setuptest
diff --git a/stats/internal/guard/check_loop_dispatch_test.go b/stats/internal/guard/check_loop_dispatch_test.go
index 021e0429..95482e83 100644
--- a/stats/internal/guard/check_loop_dispatch_test.go
+++ b/stats/internal/guard/check_loop_dispatch_test.go
@@ -179,3 +179,36 @@ func TestCheckLoopDispatchUnreadableTable(t *testing.T) {
 		t.Fatalf("rc %d, want 2: %s", rc, out)
 	}
 }
+
+// A self-review on a change's own branch in another project -- no
+// self-review-<name> branch in <agents repo> -- files its rows in that
+// project's flowd project: flow runs with -C the caller's main checkout, and
+// review-1's first-review-effort is read from that project's spectre/<name>.
+// A self-review-<name> loop branch in <agents repo> still files there.
+func TestCheckLoopDispatchOwnBranchInAnotherProject(t *testing.T) {
+	t.Parallel()
+	fx := lpNew(t, lpRowsJSON("self-review-kan-1-demo-", "self-review-kan-1-demo", "plan completed"))
+	proj := filepath.Dir(fx.repo) + "/gymie"
+	writeFile(t, proj+"/README.md", "gymie\n")
+	gitRun(t, "", "init", "-q", "-b", "develop", proj)
+	gitRun(t, proj, "add", "-A")
+	gitRun(t, proj, "commit", "-q", "-m", "init")
+	gitRun(t, proj, "checkout", "-q", "-b", "spectre/kan-1-demo")
+	writeFile(t, proj+"/spectre/changes/kan-1-demo/design.md", lpDesignLow)
+	gitRun(t, proj, "add", "-A")
+	gitRun(t, proj, "commit", "-q", "-m", "plan")
+	gitRun(t, proj, "checkout", "-q", "develop")
+	fx.env.Dir = proj
+	if rc, out := fx.run(t, "self-review-kan-1-demo-review-1", "flow-low", "opus"); rc != 0 {
+		t.Fatalf("rc %d: %s", rc, out)
+	}
+	if calls := fx.calls(t); !strings.Contains(calls, "record dispatch begin -C "+proj+" -change kan-1-demo -role reviewer") ||
+		!strings.Contains(calls, "record dispatches -change kan-1-demo -C "+proj) {
+		t.Fatalf("the own-branch loop was not filed in the caller's project:\n%s", calls)
+	}
+	gitRun(t, fx.repo, "branch", "self-review-kan-1-demo")
+	if rc, out := fx.run(t, "self-review-kan-1-demo-fix", "flow-low", "opus"); rc != 0 ||
+		!strings.Contains(fx.calls(t), "record dispatch begin -C "+fx.repo+" -change self-review-kan-1-demo") {
+		t.Fatalf("a self-review loop branch left <agents repo>: rc %d: %s\n%s", rc, out, fx.calls(t))
+	}
+}
diff --git a/stats/internal/guard/loop.go b/stats/internal/guard/loop.go
index f49e61b2..570978ae 100644
--- a/stats/internal/guard/loop.go
+++ b/stats/internal/guard/loop.go
@@ -19,9 +19,10 @@ import (
 // steps. A loop is its own change in <agents repo>'s flowd project, named
 // after its branch -- pipeline-fix-<name>-<k> or self-review-<name> --
 // except a self-review on the change's own branch, which has no branch of its
-// own and files its rows under <name> itself. Either way the rows carry the
-// loop's branch name as their session token, so one loop's rows never mix
-// with the run's own.
+// own and files its rows under <name> itself -- in the flowd project of the
+// caller's working directory, which for a change in another project is that
+// project, not <agents repo>. Either way the rows carry the loop's branch name
+// as their session token, so one loop's rows never mix with the run's own.
 func init() {
 	Registry["check-loop-dispatch"] = checkLoopDispatch
 }
@@ -105,9 +106,11 @@ func lpTable(repo string) ([]lpRow, error) {
 }
 
 // lpLoop names one loop's records: the change they are filed under, the
-// session token (the loop's branch name) and the key prefix of its rows.
+// session token (the loop's branch name), the key prefix of its rows, and
+// main, the main checkout every flow call and git config read for the loop
+// runs with -C.
 type lpLoop struct {
-	change, token, prefix string
+	change, token, prefix, main string
 }
 
 // lpMain is <repo>'s main checkout, the -C every flow call runs with.
@@ -153,7 +156,7 @@ func lpFromKey(env Env, git, repo, key string, rows []lpRow) (lpLoop, string, lp
 		if len(hits) != 1 {
 			return lpLoop{}, "", lpRow{}, fmt.Sprintf("NOT-STARTED — %s: %d branches pipeline-fix-<name>-%s; start the loop with scripts/pipeline-fix.sh start <name> %s", key, len(hits), k, k), false
 		}
-		return lpLoop{change: hits[0], token: hits[0], prefix: "pipeline-fix-" + k + "-"}, suffix, row, "", true
+		return lpLoop{change: hits[0], token: hits[0], prefix: "pipeline-fix-" + k + "-", main: repo}, suffix, row, "", true
 	}
 	if !strings.HasPrefix(key, "self-review-") {
 		return lpLoop{}, "", lpRow{}, "", false
@@ -169,9 +172,13 @@ func lpFromKey(env Env, git, repo, key string, rows []lpRow) (lpLoop, string, lp
 		if !ok || !plainChangeName(name) {
 			continue
 		}
-		loop := lpLoop{token: "self-review-" + name, prefix: "self-review-" + name + "-", change: name}
+		loop := lpLoop{token: "self-review-" + name, prefix: "self-review-" + name + "-", change: name, main: repo}
 		if lpBranchExists(env, git, repo, loop.token) {
 			loop.change = loop.token
+		} else if m, ok := lpMain(env, git, env.Dir); env.Dir != "" && ok {
+			// On the change's own branch the change lives in the caller's
+			// project -- gymie's for a gymie run -- never necessarily <agents repo>.
+			loop.main = m
 		}
 		return loop, suffix, row, "", true
 	}
@@ -318,7 +325,7 @@ func checkLoopDispatch(args []string, env Env, stdout, stderr io.Writer) int {
 	family := lpFamily(suffix)
 	var mine []lpDispatch
 	if family != "pass" {
-		if mine, ok = lpRows(env, main, lp); !ok {
+		if mine, ok = lpRows(env, lp.main, lp); !ok {
 			fmt.Fprintf(stderr, "%s: flow record dispatches -change %s failed -- cannot answer (is flowd up?)\n", self, lp.change)
 			return 2
 		}
@@ -328,11 +335,11 @@ func checkLoopDispatch(args []string, env Env, stdout, stderr io.Writer) int {
 			return refuse(fmt.Sprintf("PAIR — %s runs on subagent_type %s, model %s (the pairs table, %s); this dispatch carries subagent_type %q, model %q",
 				key, strings.Join(row.types, " or "), row.model, lpSkill, typ, model))
 		}
-		if !lpEscalated(env, git, main, lp, key) || !lpEarlierRound(lp, suffix, mine) {
+		if !lpEscalated(env, git, lp.main, lp, key) || !lpEarlierRound(lp, suffix, mine) {
 			return refuse(fmt.Sprintf("ESCALATION — %s on %s needs a completed earlier round and its no-progress escalation recorded first: scripts/pipeline-fix.sh escalate %s", key, typ, key))
 		}
 	} else if suffix == "review-1" {
-		if want := lpFirstReviewType(env, git, main, lp); typ != want {
+		if want := lpFirstReviewType(env, git, lp.main, lp); typ != want {
 			return refuse(fmt.Sprintf("PAIR — %s runs on %s, as the loop's design.md first-review-effort decision records (flow-medium when none); this dispatch carries %q", key, want, typ))
 		}
 	}
@@ -360,7 +367,7 @@ func checkLoopDispatch(args []string, env Env, stdout, stderr io.Writer) int {
 	if !slices.Contains([]string{"low", "medium", "high"}, effort) {
 		effort = "default"
 	}
-	out, _, rc := pcFlow(env, true, "record", "dispatch", "begin", "-C", main, "-change", lp.change,
+	out, _, rc := pcFlow(env, true, "record", "dispatch", "begin", "-C", lp.main, "-change", lp.change,
 		"-role", role, "-model", model, "-effort", effort, "-key", key, "-session-token", lp.token)
 	if rc != 0 {
 		fmt.Fprintf(stderr, "%s: flow record dispatch begin failed (exit %d) -- cannot answer:\n%s", self, rc, out)
diff --git a/stats/internal/guard/pipeline_fix_test.go b/stats/internal/guard/pipeline_fix_test.go
index 2fa9edf0..e68bcc0c 100644
--- a/stats/internal/guard/pipeline_fix_test.go
+++ b/stats/internal/guard/pipeline_fix_test.go
@@ -310,3 +310,39 @@ func TestPipelineFixEndClearsOwnBranchEscalation(t *testing.T) {
 		t.Fatalf("the own-branch escalation marker outlived its final clean review: %s", out)
 	}
 }
+
+// end and escalate of an own-branch self-review in another project run in
+// that project -- the caller's working directory -- and verify diffs a
+// worktree against its own origin/HEAD, not <agents repo>'s.
+func TestPipelineFixOwnBranchInAnotherProject(t *testing.T) {
+	pfIdentity(t)
+	const token = "self-review-kan-1-demo"
+	fx := pfNew(t, lpRowsJSON(token+"-", token, "plan completed", "fix completed"))
+	tmp := filepath.Dir(fx.repo)
+	porigin, proj := tmp+"/gymie.git", tmp+"/gymie"
+	gitRun(t, "", "init", "-q", "--bare", "-b", "develop", porigin)
+	gitRun(t, "", "clone", "-q", porigin, proj)
+	writeFile(t, proj+"/README.md", "gymie\n")
+	gitRun(t, proj, "add", "-A")
+	gitRun(t, proj, "commit", "-q", "-m", "init")
+	gitRun(t, proj, "push", "-q", "origin", "HEAD:develop")
+	gitRun(t, proj, "remote", "set-head", "origin", "develop")
+	gitRun(t, proj, "checkout", "-q", "-b", "spectre/kan-1-demo")
+	writeFile(t, proj+"/.flow/project.md", "# gymie\n\n## lint\n\n```bash\ntrue\n```\n")
+	gitRun(t, proj, "add", "-A")
+	gitRun(t, proj, "commit", "-q", "-m", "fix: tooling")
+	fx.env.Dir = proj
+	if rc, out := fx.pf(t, "end", token+"-fix-1", "completed"); rc != 0 ||
+		!strings.Contains(fx.calls(t), "record dispatch end -C "+proj+" -change kan-1-demo -key "+token+"-fix-1") {
+		t.Fatalf("end left the caller's project: rc %d: %s\n%s", rc, out, fx.calls(t))
+	}
+	if rc, out := fx.pf(t, "escalate", token+"-fix-2"); rc != 0 {
+		t.Fatalf("escalate: rc %d: %s", rc, out)
+	}
+	if _, out := kwRun(fx.env, "", fixtureGit, "-C", proj, "config", "--get-all", "flow.escalation"); !strings.Contains(out, token+":"+token+"-fix-2") {
+		t.Fatalf("the escalation was not recorded in the caller's project: %q", out)
+	}
+	if rc, out := fx.pf(t, "verify", proj); rc != 0 || !strings.Contains(out, "green: true") {
+		t.Fatalf("verify against the worktree's own origin/HEAD: rc %d: %s", rc, out)
+	}
+}
diff --git a/stats/internal/guard/pipelinefix.go b/stats/internal/guard/pipelinefix.go
index 166483ea..7c10dcb2 100644
--- a/stats/internal/guard/pipelinefix.go
+++ b/stats/internal/guard/pipelinefix.go
@@ -95,9 +95,9 @@ func pfRecord(state, branch string, worktrees map[string]string) map[string]any
 		"artifactUrl": nil, "jiraIssue": nil, "prUrl": nil, "updatedBy": "/flow"}
 }
 
-// pfBase is origin/HEAD's short name, e.g. origin/main.
-func pfBase(env Env, pe pfEnv) (string, bool) {
-	rc, out := kwRun(env, "", pe.git, "-C", pe.main, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
+// pfBase is dir's origin/HEAD short name, e.g. origin/main.
+func pfBase(env Env, git, dir string) (string, bool) {
+	rc, out := kwRun(env, "", git, "-C", dir, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
 	return strings.TrimSpace(out), rc == 0 && strings.HasPrefix(out, "origin/")
 }
 
@@ -116,7 +116,7 @@ func pfStart(env Env, pe pfEnv, rest []string, stdout, stderr io.Writer) int {
 		fmt.Fprint(stderr, out)
 		return 1
 	}
-	base, ok := pfBase(env, pe)
+	base, ok := pfBase(env, pe.git, pe.main)
 	if !ok {
 		fmt.Fprintf(stderr, "%s: refs/remotes/origin/HEAD is unset (git remote set-head origin --auto)\n", self)
 		return 2
@@ -256,7 +256,7 @@ func pfEnd(env Env, pe pfEnv, key, outcome string, stdout, stderr io.Writer) int
 		fmt.Fprintf(stderr, "pipeline-fix end: %s takes outcome %s\n", key, strings.Join(want, " or "))
 		return 2
 	}
-	out, _, frc := pcFlow(env, true, "record", "dispatch", "end", "-C", pe.main, "-change", lp.change,
+	out, _, frc := pcFlow(env, true, "record", "dispatch", "end", "-C", lp.main, "-change", lp.change,
 		"-key", key, "-session-token", lp.token, "-outcome", outcome)
 	fmt.Fprint(stdout, string(out))
 	if frc != 0 {
@@ -265,7 +265,7 @@ func pfEnd(env Env, pe pfEnv, key, outcome string, stdout, stderr io.Writer) int
 	// An own-branch self-review has no loop worktree for pfRemove to clear
 	// its escalation markers with; its final clean review clears them.
 	if outcome == "clean" && lp.change != lp.token {
-		kwRun(env, "", pe.git, "-C", pe.main, "config", "--unset-all", "flow.escalation", "^"+regexp.QuoteMeta(lp.token)+":")
+		kwRun(env, "", pe.git, "-C", lp.main, "config", "--unset-all", "flow.escalation", "^"+regexp.QuoteMeta(lp.token)+":")
 	}
 	return 0
 }
@@ -279,7 +279,7 @@ func pfEscalate(env Env, pe pfEnv, key string, stdout, stderr io.Writer) int {
 		fmt.Fprintf(stderr, "pipeline-fix escalate: the pairs table admits no escalation for %s\n", key)
 		return 1
 	}
-	rows, ok := lpRows(env, pe.main, lp)
+	rows, ok := lpRows(env, lp.main, lp)
 	if !ok {
 		fmt.Fprintf(stderr, "pipeline-fix escalate: flow record dispatches -change %s failed -- cannot answer\n", lp.change)
 		return 2
@@ -289,7 +289,7 @@ func pfEscalate(env Env, pe pfEnv, key string, stdout, stderr io.Writer) int {
 		fmt.Fprintf(stderr, "pipeline-fix escalate: %s has no completed earlier round to escalate from\n", key)
 		return 1
 	}
-	if rc, out := kwRun(env, "", pe.git, "-C", pe.main, "config", "--add", "flow.escalation", lp.token+":"+key); rc != 0 {
+	if rc, out := kwRun(env, "", pe.git, "-C", lp.main, "config", "--add", "flow.escalation", lp.token+":"+key); rc != 0 {
 		fmt.Fprint(stderr, out)
 		return 1
 	}
@@ -374,7 +374,9 @@ func pfChecks(env Env, pe pfEnv, wt string, touched []string) ([]string, int) {
 }
 
 func pfVerify(env Env, pe pfEnv, wt string, stdout, stderr io.Writer) int {
-	base, ok := pfBase(env, pe)
+	// wt's own origin/HEAD: an own-branch self-review verifies a worktree of
+	// the change's project, whose default branch need not be <agents repo>'s.
+	base, ok := pfBase(env, pe.git, wt)
 	if !ok {
 		fmt.Fprint(stderr, "pipeline-fix verify: refs/remotes/origin/HEAD is unset\n")
 		return 2
@@ -452,7 +454,7 @@ func pfLand(env Env, pe pfEnv, wt string, stdout, stderr io.Writer) int {
 		fmt.Fprintf(stderr, "%s: the default landing route is %q, not merge and push — land it by that route\n", self, first)
 		return 1
 	}
-	base, ok := pfBase(env, pe)
+	base, ok := pfBase(env, pe.git, pe.main)
 	if !ok {
 		fmt.Fprintf(stderr, "%s: refs/remotes/origin/HEAD is unset\n", self)
 		return 2
````

  - [x] **Step 3: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'OwnBranchInAnotherProject|CheckLoopDispatch|PipelineFix' -count=1`
    — gofmt prints nothing, the rest exit 0.
  - [x] **Step 4: Commit** the five files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/loop.go`, `stats/internal/guard/pipelinefix.go`, `stats/internal/guard/check_loop_dispatch_test.go`, `stats/internal/guard/pipeline_fix_test.go`, `scripts/pipeline-fix.sh`
**Tests:** `TestCheckLoopDispatchOwnBranchInAnotherProject`, `TestPipelineFixOwnBranchInAnotherProject`
**Regression:** reverting the commit files an own-branch self-review's `dispatch begin`, `dispatch end` and escalation marker under the agents checkout and diffs verify against its `origin/HEAD` — both tests fail.
**Baseline:** before=10 after=12
<!-- measured: cat stats/internal/guard/check_loop_dispatch_test.go stats/internal/guard/pipeline_fix_test.go | grep -cE '^func Test' (before at b803ec00; after with this task's diff applied) -->
**Commit:** fix(guard): file an own-branch self-review loop in the caller's project
**After:** none

**Decision:** caller-dir-resolves-own-branch
**Decision:** verify-base-from-worktree

- [x] 2. Run the loop-dispatch checker in the payload's working directory
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Save the diff below to
    `/tmp/pf8-task2.diff`, apply its harness hunks alone
    (`git apply --include='scripts/test-enforce-loop-dispatch-hook.sh' /tmp/pf8-task2.diff`) and
    run `bash scripts/test-enforce-loop-dispatch-hook.sh`: it FAILS with
    `FAIL: 12 checker ran in <the worktree>`.
  - [x] **Step 2: Implement.** Apply the rest
    (`git apply --exclude='scripts/test-enforce-loop-dispatch-hook.sh' /tmp/pf8-task2.diff`):

````diff verified:applied with git apply to this worktree at b803ec00; case 12 passes with it and fails with the hook hunk reverted
diff --git a/hooks/enforce-loop-dispatch.py b/hooks/enforce-loop-dispatch.py
index e9baec79..ae8c4024 100644
--- a/hooks/enforce-loop-dispatch.py
+++ b/hooks/enforce-loop-dispatch.py
@@ -9,6 +9,8 @@ review. This hook hands every such call to scripts/check-loop-dispatch.sh in the
 checkout it is installed from, which refuses a pair off the canonical table, a fix with no
 completed plan, a dispatch while an earlier one is still open -- and records the admitted
 dispatch in flowd, so check-loop-landing.sh can refuse a landing those records do not cover.
+The checker runs in the payload's `cwd`, so a self-review on a change's own branch in another
+project files its rows in that project's flowd project, not the agents checkout's.
 
 The PreToolUse contract cannot rewrite tool input, so a refusal is a deny whose reason is the
 checker's own lines.
@@ -94,8 +96,10 @@ def main():
     if BASELINE not in prompt:
         return deny(f"Loop dispatch {key} blocked: the prompt carries no {BASELINE} pointer.")
     args = [CHECK, REPO, key, str(tool_input.get("subagent_type") or ""), str(tool_input.get("model") or "")]
+    cwd = payload.get("cwd")
+    cwd = cwd if isinstance(cwd, str) and os.path.isdir(cwd) else None
     try:
-        r = subprocess.run(args, capture_output=True, text=True, timeout=CHECK_TIMEOUT)
+        r = subprocess.run(args, capture_output=True, text=True, timeout=CHECK_TIMEOUT, cwd=cwd)
     except (OSError, subprocess.SubprocessError) as e:
         return deny(f"Loop dispatch {key} blocked: check-loop-dispatch.sh could not run ({e}).")
     if r.returncode == 0:
diff --git a/scripts/check-loop-dispatch.sh b/scripts/check-loop-dispatch.sh
index b87c8006..94b0d9b1 100755
--- a/scripts/check-loop-dispatch.sh
+++ b/scripts/check-loop-dispatch.sh
@@ -20,7 +20,10 @@
 #   PLAN         a fix or review with no completed plan row.
 # Admitted (exit 0, LOOP-DISPATCH-OK): the row is opened with `flow record
 # dispatch begin` under the loop's change, its session token the loop's
-# branch name — except a `pass`, which is pair-checked and not recorded.
+# branch name — except a `pass`, which is pair-checked and not recorded. A
+# self-review on the change's own branch (no self-review-<name> branch in
+# <agents-repo>) reads and records in the flowd project of the working
+# directory it runs in — the hook runs it in the payload's cwd.
 # LOOP-DISPATCH-NA, exit 0: the key is no loop key.
 # Exit 2: cannot answer — the table unreadable, flowd unreadable, the begin
 # refused (a loop change no checkout carries: pipeline-fix.sh start), usage.
diff --git a/scripts/test-enforce-loop-dispatch-hook.sh b/scripts/test-enforce-loop-dispatch-hook.sh
index 7efe59c2..17d422d4 100755
--- a/scripts/test-enforce-loop-dispatch-hook.sh
+++ b/scripts/test-enforce-loop-dispatch-hook.sh
@@ -16,6 +16,7 @@
 #   9. a loop key with no agent-baseline pointer       -> deny, checker never run
 #  10. a Key line with trailing text                   -> deny, checker never run
 #  11. a pipeline-fix key with a non-numeric <k>       -> deny, checker never run
+#  12. a payload carrying a cwd                       -> the checker runs in it
 set -euo pipefail
 
 SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
@@ -33,6 +34,7 @@ RC_FILE="$SANDBOX/rc"
 cat >"$SANDBOX/check" <<STUB
 #!/usr/bin/env bash
 printf '%s\n' "\$@" >"$LOG"
+pwd -P >"$SANDBOX/cwd"
 echo "LOOP-DISPATCH-REFUSED: PAIR — stub"
 exit "\$(cat "$RC_FILE")"
 STUB
@@ -104,6 +106,12 @@ hook "$SANDBOX/check" "$(payload Agent $'Before anything else, read `~/.claude/r
 expect "11 non-numeric <k>" deny "no valid loop key"
 [ ! -e "$LOG" ] && pass "11 checker not run" || fail "11 checker ran"
 
+mkdir "$SANDBOX/project"; echo 0 >"$RC_FILE"
+hook "$SANDBOX/check" "$(python3 -c 'import json,sys; print(json.dumps({"tool_name": "Agent", "cwd": sys.argv[2], "tool_input": {"prompt": sys.argv[1], "subagent_type": "flow-low", "model": "opus"}}))' "$LOOP_PROMPT" "$SANDBOX/project")"
+expect "12 admitted with a cwd" allow
+[ "$(cat "$SANDBOX/cwd")" = "$(cd "$SANDBOX/project" && pwd -P)" ] \
+  && pass "12 checker ran in the payload's cwd" || fail "12 checker ran in $(cat "$SANDBOX/cwd")"
+
 if [ "$FAILURES" -gt 0 ]; then
   printf '%d case(s) failed\n' "$FAILURES" >&2
   exit 1
````

  - [x] **Step 3: Verify.** `bash scripts/test-enforce-loop-dispatch-hook.sh`,
    `python3 -m py_compile hooks/enforce-loop-dispatch.py` and
    `scripts/check-python-suppressions.sh` — all exit 0.
  - [x] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `hooks/enforce-loop-dispatch.py`, `scripts/test-enforce-loop-dispatch-hook.sh`, `scripts/check-loop-dispatch.sh`
**Tests:** `checker ran in the payload's cwd`
**Regression:** reverting the commit runs the checker in the hook's own working directory — case 12 prints `FAIL: 12 checker ran in …`.
**Baseline:** before=11 after=12
<!-- measured: grep -cE '^expect ' scripts/test-enforce-loop-dispatch-hook.sh (before at b803ec00; after with this task's diff applied) -->
**Commit:** fix(hooks): run the loop-dispatch checker in the payload's working directory
**After:** none

**Decision:** hook-runs-in-payload-cwd

- [x] 3. State that a loop tool's refusal is fixed in-run, never a stop
  - [x] **Step 1: Implement.** Apply the diff below with `git apply`:

````diff verified:applied with git apply to this worktree at b803ec00; check-verbatim-moves.sh reports 0 violations with it
diff --git a/skills/flow-self-review/SKILL.md b/skills/flow-self-review/SKILL.md
index c965dbeb..8e2ce23b 100644
--- a/skills/flow-self-review/SKILL.md
+++ b/skills/flow-self-review/SKILL.md
@@ -162,7 +162,10 @@ closed `aborted` and may be dispatched again under the same key. A review is clo
 `Important: <n>` lines decide the outcome: `clean` when both are 0, `fix` otherwise. A dispatch on its row's escalation column is admitted only after
 `<agents repo>/scripts/pipeline-fix.sh escalate <key>`, and every loop branch's landing push only
 once `<agents repo>/scripts/check-loop-landing.sh` finds its plan, its fix and a last review
-closed `clean`.
+closed `clean`. A refusal or failure of any of these tools on a step this file prescribes is a
+pipeline defect, fixed in-run per **Pipeline defects found mid-run**
+(`skills/flow-contracts/pipeline.md`) and the loop resumed once the fix lands — never a stop or a
+question to the operator.
 
 `<agents-base>` is the one **Pipeline defects found mid-run** binds. `<fix-branch>` and
 `<fix-base>` are where the fixes go (`design.md`, `fixes-ride-the-change`): inside a run whose
````

  - [x] **Step 2: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py` and `scripts/check-dispatch-paragraphs.sh` — all exit 0.
  - [x] **Step 3: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow-self-review/SKILL.md`
**Tests:** **none** — prose; `check-verbatim-moves.sh` judges the new sentence.
**Regression:** none — no behaviour.
**Commit:** docs(flow-self-review): fix a loop tool's refusal in-run, never stop to ask
**After:** none

**Decision:** no-stop-sentence-in-step-3
