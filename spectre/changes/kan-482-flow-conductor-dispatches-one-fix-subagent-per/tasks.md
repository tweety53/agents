# kan-482-flow-conductor-dispatches-one-fix-subagent-per

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** the review panel's fix round is held to one combined panel-fix dispatch per round by prose in both skill files and a guard at the panel's close.
**Architecture:** a read-only `flow record dispatches` verb over the store's existing dispatch rows feeds a new bash guard that counts panel-fix dispatches per round key and checks key shape; the skill prose states the contract the guard enforces.
**Spec:** this change's own `design.md` (the repo carries no `spectre/specs/` entries; the skill markdown is the contract surface).

Baseline for every `Baseline:` below is the stats Go suite, whole-repo run:
`cd stats && go test ./... -race -count=1`, counted as `go test ./... -race -count=1 -v | grep -c '^=== RUN'` (parents + subtests).
Before task 1 the suite is green with 1073 test cases.
<!-- measured: cd stats && go test ./... -race -count=1 -v | grep -c '^=== RUN' @ branch spectre/kan-482-flow-conductor-dispatches-one-fix-subagent-per, 2026-09-10 -->

- [x] 1. Add the `flow record dispatches` read command
**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRunRecordDispatchesPrintsDispatchesAsJSONArray`, `TestRunRecordDispatchesEmptyChangePrintsEmptyArray`, `TestRunRecordDispatchesUnreachableStoreFails`, `TestRunRecordDispatchesNullDispatchesPrintsEmptyArray`
**Regression:** reverting leaves the store's dispatch rows write-only — `check-panel-fix-single-dispatch.sh` (task 2) has nothing to query, so an over-dispatching conductor is caught by no gate again and KAN-482's drift is invisible
**Baseline:** before=1073 after=1077
<!-- measured: full suite after the pass-1 fix round counts 1077 (+4: the three pass-1 parent tests plus TestRunRecordDispatchesNullDispatchesPrintsEmptyArray from finding F4) @ branch spectre/kan-482-flow-conductor-dispatches-one-fix-subagent-per, 2026-09-10 -->

**Interfaces:**
- Consumes: `Store.RunRecord` / `client.GetRunRecord` — the load `runRecordFindings` already uses (`stats/cmd/flow/record.go:1188`), whose `Run.Dispatches` is seq-ordered per `TestRunRecordOrdersDispatchesAndFindings`; the `records.Dispatch` JSON shape (`stats/internal/records/types.go:88` — `id`, `agentId`, `seq`, `key`, `role`, `model`, `effort`, `commitSha`, `outcome`, `sessionToken`, `startedAt`, `endedAt`)
- Produces: CLI `flow record dispatches -change <name>` printing a JSON array of `records.Dispatch` objects, `[]` when none, a non-zero exit with `flow: dispatches: <err>` on stderr when the store cannot be reached

Amended at implementation: the planned `ListDispatches` store method is dropped — `RunRecord` already loads every dispatch row ordered by seq (`TestRunRecordOrdersDispatchesAndFindings`), so the verb reuses it and no store code changes.

  - [ ] **Step 1: RED** — in `record_test.go`, mirroring the `TestRecordFindings*` cases: `TestRunRecordDispatchesPrintsDispatchesAsJSONArray` records two dispatches then asserts stdout parses as a JSON array of two objects in seq order carrying `key` and `role`; `TestRunRecordDispatchesEmptyChangePrintsEmptyArray` asserts a change with no dispatches prints exactly `[]` and exits 0; `TestRunRecordDispatchesUnreachableStoreFails` points the verb at a dead store and asserts non-zero exit, empty stdout, and a `flow: dispatches:` line on stderr. Run `cd stats && go test -race -count=1 -run 'TestRunRecordDispatches' ./cmd/flow/` and report the failure (verb undefined).
  - [ ] **Step 2: GREEN** — in `record.go`: the usage block gains the `flow record dispatches [-addr url] [-timeout dur] [-C dir] -change name` line and one prose line ("dispatches prints one change's dispatches as a JSON array on stdout — seq order, every role; `[]` for a change with none; a store it cannot reach is a non-zero exit, never a silent empty array"); `case "dispatches": return runRecordDispatches(...)` beside `findings`; the verb mirrors `runRecordFindings` — same flags, same `ErrNotFound`-as-empty-array handling — marshalling the `Dispatches` slice ALONE (never the whole `Run`, per the findings verb's own rule).
  - [ ] **Step 3: Verify** — `cd stats && gofmt -w .` then `gofmt -l .` (empty), `cd stats && go vet ./...`, `cd stats && go test -race -count=1 -run 'TestRunRecordDispatches' ./cmd/flow/`, then the full suite per the header and record the count.

**Commit:** `feat(flow): add flow record dispatches read command`

- [x] 2. Add the `check-panel-fix-single-dispatch.sh` guard
**Build:** green
**Files:** `scripts/check-panel-fix-single-dispatch.sh`, `scripts/test-check-panel-fix-single-dispatch.sh`, `skills/flow/scripts/check-panel-fix-single-dispatch.sh`, `skills/flow/SKILL.md`
**Allowed-collateral:** `scripts/lib/*`
**Tests:** `scripts/test-check-panel-fix-single-dispatch.sh`
**Regression:** reverting turns the panel's close gate back to findings-only — a conductor that dispatched four fix subagents in one round (KAN-482's exact shape, or the same drift under invented keys like `panel-fix-f1`) closes the panel clean and reaches the operator's review gate unflagged
**Baseline:** before=1073 after=1073
<!-- predicted: no Go test is added by this task; the count is unchanged after it and after task 3; confirmed at flow.verify -->

**Behavior contract** (the harness's cases are this contract, one per case):
- usage: `check-panel-fix-single-dispatch.sh <worktree> <change-name> <session-token>`
- queries `flow record dispatches -change <name> -C <worktree>` (the verb from task 1), jq-filters rows to `sessionToken == <token>` and `role == "panel-fix"`, and checks, per row, that `key` matches `^panel-fix-[0-9]+(-retry)?$`; grouping key is the base (a trailing `-retry` stripped); per base round: at most one non-retry row, at most two rows total, and two rows only when exactly one is the `-retry` variant
- exit 0: every panel-fix row of this session token passes; rows of other roles never count, and a foreign session token's rows never count even when counting them would violate — case 6's foreign-token row carries two originals, so a scoping-dropping mutant flips the case (pass-1 finding F3)
- exit 1: each violation named on stderr — an over-counted round names the round and its row count; an out-of-shape key names the key verbatim
- exit 2: cannot answer — missing arguments, a non-directory worktree, a change name outside the containment `case` (duplicated verbatim from `check-panel-findings-closed.sh`, whose canonicalisation-before-`-C` hazard comment that guard carries and this one duplicates with it), the `flow` call exiting non-zero (store unreachable — never read as "no dispatches"), or jq missing or failing

  - [ ] **Step 1: RED** — write `scripts/test-check-panel-fix-single-dispatch.sh` following the existing harnesses' shape (discover how `scripts/test-check-panel-findings-closed.sh` stubs `flow` on PATH with canned output; do the same with canned dispatch arrays). Cases: (a) one canonical row per of two rounds → 0; (b) original plus `-retry` in one round → 0; (c) two non-retry rows in one round → 1 naming `panel-fix-1`; (d) invented key `panel-fix-f1` → 1 naming it; (e) a `-retry` row with no original → 1; (f) `implementer`- and `reviewer`-role rows and a foreign session token's `panel-fix` row present → 0; (g) missing session-token argument → 2; non-directory worktree → 2; stub `flow` exiting 1 → 2; change name `../evil` → 2. Run the harness and report the failures (guard does not exist yet).
  - [ ] **Step 2: GREEN** — write `scripts/check-panel-fix-single-dispatch.sh` to the contract above: `set -euo pipefail`, `LC_ALL=C`, the duplicated containment `case`, canonicalise the worktree before it ever reaches `flow -C`, `jq -c` per-row checks, violations accumulated and printed before exit 1.
  - [ ] **Step 3:** run the harness — every case passes; `bash -n` clean on both scripts.
  - [ ] **Step 4: register** — `ln -s ../../../scripts/check-panel-fix-single-dispatch.sh skills/flow/scripts/` and add the bare basename to `skills/flow/SKILL.md`'s guard list, alphabetical, after `check-panel-findings-closed.sh`.
  - [ ] **Step 5: Verify** — `bash scripts/test-check-panel-fix-single-dispatch.sh`, `bash scripts/check-guard-symlinks.sh`, `scripts/check-vocabulary.sh`.
**Commit:** `feat(scripts): add check-panel-fix-single-dispatch guard`

- [x] 3. State the one-panel-fix-dispatch contract and wire the guard into the panel close
**Build:** green
**Files:** `skills/flow/review-panel.md`, `skills/flow/implement.md`
**Tests:** none — prose only; the verification is the repo's own guard set below
**Regression:** reverting leaves line 905's single sentence as the whole contract — the exact text a conductor re-derived KAN-482's four-dispatch drift past — and a panel close that checks findings but never dispatch counts
**Baseline:** before=1073 after=1073
<!-- predicted: unchanged; no Go surface this task touches; confirmed at flow.verify -->

  - [ ] **Step 1: contract prose** — in `skills/flow/review-panel.md`'s fix step, expand "Give the surviving findings to **one** fix subagent as the combined list." into the full contract: exactly **one** panel-fix dispatch per fix round, carrying the combined list of every surviving open finding — never one dispatch per reviewer, slot, or finding (the KAN-482 drift named as the anti-pattern); the dispatch is awaited in the foreground before the round's reproducer re-runs begin and no fix subagent is left in flight at the turn's end; every panel-fix `-key` is exactly `panel-fix-<round>`, the handshake retry's `panel-fix-<round>-retry` the only second key a round may carry; before recording the `dispatch begin`, confirm this round has no panel-fix begin already recorded — a second begin under a fresh key is the violation itself.
  - [ ] **Step 2: wiring prose** — at `skills/flow/review-panel.md`'s panel close, beside the `check-panel-findings-closed.sh` call: add `check-panel-fix-single-dispatch.sh <worktree> <name> <session-token>` immediately before `flow stage end -command '/flow' -stage flow.review-panel`; exit 0 proceeds; exit 1 is a handback `## Question` — **Continue — the violation stays recorded in this run's output** *(default, recommended)* / **Stop the run** — because findings may already be verified closed and the round's work is real; exit 2 reports and stops the stage close.
  - [ ] **Step 3: conductor prose** — in `skills/flow/implement.md`'s **Dispatch the conductor** relay-contract list, add the constraint the conductor's prompt must state: the review panel's fix step is one panel-fix dispatch per round on the combined list, never one per reviewer, slot or finding, awaited in the foreground like every other child.
  - [ ] **Step 4: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`, `scripts/check-dispatch-paragraphs.sh`, `scripts/check-contract-budget.sh`, `scripts/check-stage-mark-calls.sh`, `scripts/check-markdown-integrity.py`; for `scripts/check-normative-inventory.sh`, capture its output before the first edit and diff after the last, resolving any difference by restoring the sentence.
**Commit:** `docs(flow): state the one-panel-fix-dispatch contract and wire its guard`
