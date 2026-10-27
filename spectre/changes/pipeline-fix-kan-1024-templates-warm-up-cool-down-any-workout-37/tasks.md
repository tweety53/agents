# pipeline-fix-kan-1024-templates-warm-up-cool-down-any-workout-37

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a `/flow` turn cannot end mid-phase: a `Stop` hook refuses it and names the next stage.

**Architecture:** `hooks/flow-stop-guard.py`, registered under Claude Code's `"Stop"` event, reads
this session's last `/flow` stage row and the change's record from the flow store and blocks the
stop unless the run is at a gate, an ask, an explicit operator stop, or made no progress since its
last block.

**Spec:** `design.md` of this change.

## Global constraints

- Pipeline-fix fixer: one commit per task with its `**Commit:**` subject and no `Task-Id:` or
  other trailer; never write `spectre/changes/`.
- Fail open on every error, like `hooks/flow-active-change.py`.
- No new inline suppressions; `scripts/check-python-suppressions.sh` exits 0.
- Read-only against the dev store; the test uses an `httptest` stub store and a stub `flow` on PATH.

## Review Focus

- Every allow case of `allow-at-gates-asks-and-explicit-stop` exits 0 with no output.
- `stop_hook_active` with an unchanged fingerprint allows (`stop-hook-active-needs-progress`).
- Store unreachable allows.

- [ ] 1. Refuse a mid-phase /flow turn end with a Stop hook
  - [ ] **Step 1: Write the failing test.** Add
    `stats/internal/repotest/flow_stop_guard_hook_test.go`, modelled on
    `flow_active_change_hook_test.go` (stub store via `FLOW_ADDR`, stub `flow` first on PATH
    printing ticked/total, fixture transcript JSONL, `HOME` in `t.TempDir()`). Cases:
    (a) open stage → stdout JSON `"decision":"block"` with the stage name in `reason`;
    (b) `STARTED`, 7/7 ticked, no open stage → block naming `flow.write-in-progress`;
    (c) `STARTED`, 0 total → allow; (d) `STARTED`, 0/7, no open stage → allow;
    (e) `IN_PROGRESS` → allow; (f) `FINISHED` → allow;
    (g) last assistant tool use `AskUserQuestion` → allow; (h) last user message "stop here" → allow;
    (i) `stop_hook_active` true, fingerprint unchanged since a block → allow;
    (j) `stop_hook_active` true, a stage ended since → block; (k) store 500 → allow;
    (l) no `/flow` row for the session → allow. Run
    `cd stats && go test ./internal/repotest -run TestFlowStopGuardHook`: fails (hook missing).
  - [ ] **Step 2: Implement** `hooks/flow-stop-guard.py` per `design.md` decisions
    `stop-hook-blocks-mid-phase`, `block-on-open-stage-or-all-ticked`,
    `allow-at-gates-asks-and-explicit-stop`, `stop-hook-active-needs-progress`,
    `import-flow-active-change-helpers`.
  - [ ] **Step 3: Register.** In `setup.sh` `install_hooks`, add a `warn_unregistered_hook` for
    `flow-stop-guard` with a `"Stop"` snippet; add the hook to `README.md`'s hook list; in
    `skills/flow-contracts/pipeline.md` append to the "A turn ends only at a gate…" paragraph one
    sentence: the `Stop` hook `hooks/flow-stop-guard.py` refuses such a turn end (kan-1024).
  - [ ] **Step 4: Verify.** The repotest passes; `cd stats && go test ./internal/setuptest -run TestLoopHookTimeoutsFailClosed`
    passes; with the block branch removed, case (a) fails.
  - [ ] **Step 5: Commit** the five files with the subject below.

**Build:** green
**Files:** `hooks/flow-stop-guard.py`, `stats/internal/repotest/flow_stop_guard_hook_test.go`, `setup.sh`, `README.md`, `skills/flow-contracts/pipeline.md`
**Tests:** `TestFlowStopGuardHook`
**Regression:** case (b) — reverting the commit lets a `STARTED` run with every task ticked end its turn: no block on stdout.
**Baseline:** before=0 after=12

**Commit:** fix(hooks): refuse a mid-phase /flow turn end with a Stop hook
**After:** none

**Decision:** stop-hook-blocks-mid-phase
**Decision:** block-on-open-stage-or-all-ticked
**Decision:** allow-at-gates-asks-and-explicit-stop
**Decision:** stop-hook-active-needs-progress
**Decision:** import-flow-active-change-helpers
