# pipeline-fix-kan-1060-flow-cost-visual-verify-now-costs-more-than-47

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** no hook test depends on wall-clock time, so `TestEnforceLoopDispatchHookRecordsADenial`
stops failing under the full `-race` suite.

**Architecture:** `hooks/enforce-loop-dispatch.py` gains kan-1033's `bound()` seam; `runHook` in
`stats/internal/repotest/helpers_test.go` injects `FLOW_HOOK_UNBOUNDED=1` for every hook.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`. A diff that does not apply is reported, never hand-merged by guesswork.
- No retry, sleep or wider timeout (`fix-determinism-at-the-source`). No new inline suppressions.
- Never stop flowd or flow-postgres, never drop the `flow` DB.

## Review Focus

- Production behaviour unchanged: every bound still applies unless `FLOW_HOOK_UNBOUNDED=1`
  (`lift-bound-via-existing-seam`).
- The forced-cause proof fails without the seam and passes with it (`cause-is-the-bound`).

- [ ] 1. Lift the loop hook's wall-clock bounds in every hook test
  - [ ] **Step 1: Force the cause before touching anything.** In
    `stats/internal/repotest/helpers_test.go` `incidentStub`, change the script body's start from
    `#!/bin/sh\nprintf` to `#!/bin/sh\nsleep 11\nprintf` (temporary, never committed). From `stats/`
    run `make -s web-dist && go test -count=1 -run TestEnforceLoopDispatchHookRecordsADenial ./internal/repotest/`:
    FAIL with `checker rc 1: incident calls = [], want 1`. Revert the `sleep 11`.
  - [ ] **Step 2: Implement.** Save the diff below to `/tmp/pf1060.diff` and run
    `git apply /tmp/pf1060.diff` from the repository root.

````diff verified:applied with git apply at b8ee4473; with incidentStub sleeping 11 s the isolated run passes, and fails with incident calls = [] when only the FLOW_HOOK_UNBOUNDED=1 injection is removed; go test -race -count=1 ./internal/repotest/ ok
diff --git a/hooks/enforce-loop-dispatch.py b/hooks/enforce-loop-dispatch.py
index c6adad30..3ebccb02 100644
--- a/hooks/enforce-loop-dispatch.py
+++ b/hooks/enforce-loop-dispatch.py
@@ -61,6 +61,13 @@ REPO = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
 # The checker the hook runs; overridable for the hook's own harness only.
 CHECK = os.environ.get("LOOP_DISPATCH_CHECK") or os.path.join(REPO, "scripts", "check-loop-dispatch.sh")
 
+def bound(seconds):
+    """A subprocess call's wall-clock bound. FLOW_HOOK_UNBOUNDED=1 lifts it: the hook's tests
+    inject that so the full -race suite starving a stub past the bound never fails a case
+    (kan-1060) -- the bound is production's fail-open, not under test."""
+    return None if os.environ.get("FLOW_HOOK_UNBOUNDED") == "1" else seconds
+
+
 # The payload's cwd, set by main for deny's incident row.
 PAYLOAD_CWD = None
 
@@ -92,7 +99,7 @@ def record_incident(guard, reason, cwd):
     if cwd:
         args += ["-C", cwd]
     try:
-        subprocess.run(args, capture_output=True, text=True, timeout=10, check=False)
+        subprocess.run(args, capture_output=True, text=True, timeout=bound(10), check=False)
     except (OSError, ValueError, subprocess.SubprocessError):
         pass
 
@@ -102,7 +109,7 @@ def has_ref(repo, *refs):
     try:
         r = subprocess.run(
             ["git", "-C", repo, "for-each-ref", "--count=1", "--format=x", *refs],
-            capture_output=True, text=True, timeout=10,
+            capture_output=True, text=True, timeout=bound(10),
         )
     except (OSError, subprocess.SubprocessError):
         return False
@@ -161,7 +168,7 @@ def main():
         return deny(f"Loop dispatch {key} blocked: the prompt carries no {BASELINE} pointer.")
     args = [CHECK, REPO, key, str(tool_input.get("subagent_type") or ""), str(tool_input.get("model") or "")]
     try:
-        r = subprocess.run(args, input=prompt, capture_output=True, text=True, timeout=CHECK_TIMEOUT, cwd=cwd)
+        r = subprocess.run(args, input=prompt, capture_output=True, text=True, timeout=bound(CHECK_TIMEOUT), cwd=cwd)
     except (OSError, subprocess.SubprocessError) as e:
         return deny(f"Loop dispatch {key} blocked: check-loop-dispatch.sh could not run ({e}).")
     if r.returncode == 0:
diff --git a/stats/internal/repotest/helpers_test.go b/stats/internal/repotest/helpers_test.go
index 76d13e2a..702ff5d3 100644
--- a/stats/internal/repotest/helpers_test.go
+++ b/stats/internal/repotest/helpers_test.go
@@ -87,8 +87,9 @@ func runHook(t *testing.T, hook string, payload []byte, env ...string) (stdout,
 	cmd := exec.Command("python3", hook)
 	cmd.Stdin = bytes.NewReader(payload)
 	// FLOW_HOOK_INCIDENT_CMD=true keeps a deny's incident write off the dev store; a caller's
-	// env, appended after it, may point it at a stub.
-	cmd.Env = append(append(append(append(os.Environ(), "HOME="+t.TempDir()), goCacheEnv()...), "FLOW_GUARD_CACHE_DIR="+cache, "FLOW_HOOK_INCIDENT_CMD=true"), env...)
+	// env, appended after it, may point it at a stub. FLOW_HOOK_UNBOUNDED=1 lifts the hooks'
+	// wall-clock bounds: under the full -race suite a starved stub timed out silently (kan-1060).
+	cmd.Env = append(append(append(append(os.Environ(), "HOME="+t.TempDir()), goCacheEnv()...), "FLOW_GUARD_CACHE_DIR="+cache, "FLOW_HOOK_INCIDENT_CMD=true", "FLOW_HOOK_UNBOUNDED=1"), env...)
 	var out, errb bytes.Buffer
 	cmd.Stdout, cmd.Stderr = &out, &errb
 	err = cmd.Run()
````

  - [ ] **Step 3: Prove the seam.** Re-add Step 1's `sleep 11` and rerun Step 1's command: ok.
    Remove only `, "FLOW_HOOK_UNBOUNDED=1"` from `runHook` and rerun: FAIL with
    `incident calls = []`. Restore it and revert the `sleep 11`. Record both runs in the commit
    body: the seam replaces the hooks' 10 s wall-clock bound with no bound under test.
  - [ ] **Step 4: Verify.** From `stats/`: `make -s web-dist && go vet ./... && gofmt -l .` clean
    and the **Tests:** selector ok; from the repository root every `scripts/check-*.sh`
    guard named in `.flow/project.md`'s `## lint` exits 0.
  - [ ] **Step 5: Commit** with the subject above.

**Build:** green
**Files:** `hooks/enforce-loop-dispatch.py`, `stats/internal/repotest/helpers_test.go`
**Tests:** none — repairs existing tests: `TestEnforceLoopDispatchHookRecordsADenial`, `TestEnforceAgentBaselineHookRecordsADenial`
**Regression:** `TestEnforceLoopDispatchHookRecordsADenial` — with Step 3's forced `sleep 11` stub, reverting the commit fails it with `incident calls = [], want 1`.
**Baseline:** before=0 after=0

**Commit:** fix(hooks): lift the loop hook's wall-clock bounds under test
**After:** none

**Decision:** cause-is-the-bound
**Decision:** lift-bound-via-existing-seam
**Decision:** seam-in-runhook-not-per-call
