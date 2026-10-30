# pipeline-fix-kan-1060-flow-cost-visual-verify-now-costs-more-than-48

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** no hook test reads the live dev store, so `TestEnforceAgentBaselineHookRefusesOffPolicyDispatch`
passes whatever `FLOW_RECORDS_ADDR` the verify inherits.

**Architecture:** `runHook` in `stats/internal/repotest/helpers_test.go` pins both store addresses to a
dead port ahead of the caller's env.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- The task's change is the unified diff its step gives, applied from the repository root with
  `git apply`. A diff that does not apply is reported, never hand-merged by guesswork.
- No retry, sleep or wider timeout (`fix-determinism-at-the-source`). No new inline suppressions.
- Never stop flowd or flow-postgres, never drop the `flow` DB, never close or edit a live experiment.

## Review Focus

- Store-stub callers still override `FLOW_ADDR` (caller env appended last) (`pin-store-addrs-in-runhook`).
- The forced proof fails without the pin and passes with it.

- [ ] 1. Keep hook tests off the live flowd
  - [ ] **Step 1: Force the cause before touching anything.** From `stats/` run
    `make -s web-dist && FLOW_RECORDS_ADDR=http://127.0.0.1:4173 go test -count=1 -run TestEnforceAgentBaselineHookRefusesOffPolicyDispatch ./internal/repotest/`:
    FAIL with `DISPATCH-PAIR-REFUSED: explorer runs on sonnet/Explore@low` while the live experiment
    is open. Without the variable: ok.
  - [ ] **Step 2: Implement.** Save the diff below to `/tmp/pf1060-48.diff` and run
    `git apply /tmp/pf1060-48.diff` from the repository root.

````diff verified:applied with git apply at 3deff9fa; FLOW_RECORDS_ADDR=FLOW_ADDR=http://127.0.0.1:4173 go test -count=1 ./internal/repotest/ ok
diff --git a/stats/internal/repotest/helpers_test.go b/stats/internal/repotest/helpers_test.go
index 702ff5d3..43e80841 100644
--- a/stats/internal/repotest/helpers_test.go
+++ b/stats/internal/repotest/helpers_test.go
@@ -89,7 +89,9 @@ func runHook(t *testing.T, hook string, payload []byte, env ...string) (stdout,
 	// FLOW_HOOK_INCIDENT_CMD=true keeps a deny's incident write off the dev store; a caller's
 	// env, appended after it, may point it at a stub. FLOW_HOOK_UNBOUNDED=1 lifts the hooks'
 	// wall-clock bounds: under the full -race suite a starved stub timed out silently (kan-1060).
-	cmd.Env = append(append(append(append(os.Environ(), "HOME="+t.TempDir()), goCacheEnv()...), "FLOW_GUARD_CACHE_DIR="+cache, "FLOW_HOOK_INCIDENT_CMD=true", "FLOW_HOOK_UNBOUNDED=1"), env...)
+	// FLOW_ADDR and FLOW_RECORDS_ADDR point at a dead port so no hook reads the live dev store: an
+	// inherited FLOW_RECORDS_ADDR let an open experiment's overrides flip a policy case (kan-1060).
+	cmd.Env = append(append(append(append(os.Environ(), "HOME="+t.TempDir()), goCacheEnv()...), "FLOW_GUARD_CACHE_DIR="+cache, "FLOW_HOOK_INCIDENT_CMD=true", "FLOW_HOOK_UNBOUNDED=1", "FLOW_ADDR=http://127.0.0.1:1", "FLOW_RECORDS_ADDR=http://127.0.0.1:1"), env...)
 	var out, errb bytes.Buffer
 	cmd.Stdout, cmd.Stderr = &out, &errb
 	err = cmd.Run()
````

  - [ ] **Step 3: Prove the pin.** Rerun Step 1's forced command: ok. Remove only
    `"FLOW_RECORDS_ADDR=http://127.0.0.1:1"` from `runHook` and rerun: FAIL. Restore it. Record both
    runs in the commit body: the pin replaces the inherited store address with a dead port.
  - [ ] **Step 4: Verify.** From `stats/`: `make -s web-dist && go vet ./... && gofmt -l .` clean
    and `FLOW_RECORDS_ADDR=http://127.0.0.1:4173 go test -count=1 ./internal/repotest/` ok; from the
    repository root every `scripts/check-*.sh` guard named in `.flow/project.md`'s `## lint` exits 0.
  - [ ] **Step 5: Commit** with the subject below.

**Build:** green
**Files:** `stats/internal/repotest/helpers_test.go`
**Tests:** none — repairs existing tests: `TestEnforceAgentBaselineHookRefusesOffPolicyDispatch`
**Regression:** `TestEnforceAgentBaselineHookRefusesOffPolicyDispatch` — run with `FLOW_RECORDS_ADDR=http://127.0.0.1:4173` while an experiment overrides `explorer`, reverting the commit fails it with `DISPATCH-PAIR-REFUSED`.
**Baseline:** before=0 after=0

**Commit:** fix(repotest): keep hook tests off the live flowd
**After:** none

**Decision:** pin-store-addrs-in-runhook
