# kan-446-hand-verify-guard-false-positives

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Three serial commits: the habit's canonical statement (task 1), the four guard-header procedures
(task 2), and the call-site citations that join them (task 3, last — its citations resolve only
once task 1's heading exists). `design.md` is canonical for every decision.

**Baseline, measured before any edit:**

- `scripts/run-guard-tests.sh`: 58 harnesses, 58 passed, 0 failed.
  <!-- measured: scripts/run-guard-tests.sh @ branch spectre/kan-446-hand-verify-guard-false-positives, worktree at merge-base db09958 -->
- `scripts/check-vocabulary.sh`, `scripts/check-references.sh`, `scripts/check-contract-budget.sh`: each exit 0.
  <!-- measured: the three guards, run from the worktree root @ merge-base db09958 -->

---

- [x] 1. State the habit where every run reads it

Add a **Hand-verifying a guard verdict** section to `skills/flow-contracts/pipeline.md`, after
**Guard presence check** and before **Finish contract**. The section states, once: when a gate
guard's verdict fires (`OUTSTANDING`, `MOVED`, `LEFTOVER`, `REFUSE`) while the situation
contradicts the pipeline's own structural conventions, the run relays the guard's hand-verification
procedure — carried in the guard's own header — alongside the breakdown, and the operator verifies
before choosing a course. It names both failure directions (trusting the verdict blindly,
overriding it blindly), pins what hand verification may not do (rewrite the verdict line or the
exit code, bypass the courses the gate offers), keeps the recording path where it is today
(`flow record verdict false-positive` at the unfinished-work gate), and enumerates the four gate
guards: `check-finish-preflight.sh` (`REFUSE`), `check-unfinished-work.sh` (`OUTSTANDING`),
`check-base-moved.sh` (`MOVED`), `check-cleanup-complete.sh` (`LEFTOVER`). The addition stays
inside the file's contract budget (36,155 bytes against 25,167 measured today), so no `budgets()`
row moves.
<!-- measured: wc -c skills/flow-contracts/pipeline.md @ merge-base db09958 -->

**Files:** `skills/flow-contracts/pipeline.md`
**Tests:** none
**Regression:** reverting this commit leaves the habit unstated and task 3's citations pointing at
a section that does not exist — `check-references.sh` then fails at task 3's gate.
**Baseline:** before=58 after=58 guard harnesses in `scripts/run-guard-tests.sh`
<!-- measured: scripts/run-guard-tests.sh @ merge-base db09958 -->
<!-- predicted: scripts/run-guard-tests.sh after this task -->
**Commit:** `docs(flow-contracts): state the hand-verification habit for gate guard verdicts`
**Build:** green

Verify step:

  - [x] **Step 1: `scripts/check-vocabulary.sh` and `scripts/check-references.sh` exit 0**
  - [x] **Step 2: `scripts/check-contract-budget.sh` exits 0 with no `budgets()` edit**

- [x] 2. Carry each guard's hand-verification procedure in its own header

Header paragraphs — comments only, no executable line changes — one per guard, each opening with a
name a call site can cite:

- `scripts/check-unfinished-work.sh` — recompute signal 1 by resolving the plan through
  `scripts/lib/change-plan.sh` the way the guard does and counting column-0 `^- \[ \]` lines by
  hand; recompute signal 2 with `flow record findings -change <name> -C <worktree>` and expect
  `[]`. Names the KAN-423 shape — a cross-repo change whose plan resolves only in the canonical
  tree — as the known structural false positive.
- `scripts/check-base-moved.sh` — recount with `git rev-list <recorded-merge-base>..<base-ref>`
  and intersect the `overlaps:` paths with the change's own touched paths, index and working tree
  included. A stale recorded merge base is the known structural cause.
- `scripts/check-cleanup-complete.sh` — check each registry row the breakdown names against the
  filesystem; a worktree kept legitimately (the state file keeps an entry whose removal failed) is
  the known structural shape.
- `scripts/check-finish-preflight.sh` — re-derive the refused signal by hand: the branch states,
  the main checkout's branch, its tracked changes, stray worktrees.

**Files:** `scripts/check-unfinished-work.sh`, `scripts/check-base-moved.sh`,
`scripts/check-cleanup-complete.sh`, `scripts/check-finish-preflight.sh`
**Tests:** none
**Regression:** reverting this commit leaves the four guards with verdicts but no stated way to
verify them — task 3's citations then name procedures that are not there.
**Baseline:** before=58 after=58 guard harnesses in `scripts/run-guard-tests.sh`
<!-- measured: scripts/run-guard-tests.sh @ merge-base db09958 -->
<!-- predicted: scripts/run-guard-tests.sh after this task -->
**Commit:** `docs(scripts): carry hand-verification procedures in the gate guard headers`
**Build:** green

Verify step:

  - [x] **Step 1: `bash -n` on all four guards exits 0**
  - [x] **Step 2: `scripts/check-vocabulary.sh` exits 0 and the four guards' own harnesses pass
        (`test-check-unfinished-work.sh`, `test-check-base-moved.sh`,
        `test-check-cleanup-complete.sh`, `test-check-finish-preflight.sh`)**

- [x] 3. Cite the habit from the four gate call sites

One citing line per site — never a restatement of the habit or of a procedure:

- `skills/flow/integrate.md` — the preflight `REFUSE` bullet, the `OUTSTANDING` bullet in **1.
  Check for unfinished work**, and the `MOVED` prompt in **2** each relay the guard's
  hand-verification procedure per **Hand-verifying a guard verdict**
  (`skills/flow-contracts/pipeline.md`).
- `skills/flow-contracts/finish-contract-run1.md` — one sentence at the `OUTSTANDING` courses
  table and one at the base-moved section, citing the same section.
- `skills/flow/archive.md` step 7 — the `LEFTOVER:` stop relays the procedure; the wording fits
  inside the 279 bytes the file has left of its budget.
  <!-- measured: wc -c skills/flow/archive.md, budget 18748 @ merge-base db09958 -->
- `skills/flow-contracts/finish-contract-run2.md` step 7 — the `LEFTOVER:` row cites the section.

**Files:** `skills/flow/integrate.md`, `skills/flow-contracts/finish-contract-run1.md`,
`skills/flow/archive.md`, `skills/flow-contracts/finish-contract-run2.md`
**Tests:** none
**Regression:** reverting this commit leaves the verdict prompts silent about verification — the
habit and the procedures exist (tasks 1–2) but no gate relays them.
**Baseline:** before=58 after=58 guard harnesses in `scripts/run-guard-tests.sh`
<!-- measured: scripts/run-guard-tests.sh @ merge-base db09958 -->
<!-- predicted: scripts/run-guard-tests.sh after this task -->
**Commit:** `docs(skills): relay hand-verification guidance at the gate prompts`
**Build:** green

Verify step:

  - [x] **Step 1: `scripts/check-references.sh` exits 0 — every new citation resolves**
  - [x] **Step 2: `scripts/check-contract-budget.sh`, `scripts/check-dispatch-paragraphs.sh` and
        `scripts/check-markdown-integrity.py` exit 0**

FULL SUITE (this plan's last task, run inline once after the steps above and before the commit):
`scripts/run-guard-tests.sh`, then `cd stats && go test ./... -race -count=1`, then
`cd stats/web && npm test` — the resolved `## test` list, in order.
