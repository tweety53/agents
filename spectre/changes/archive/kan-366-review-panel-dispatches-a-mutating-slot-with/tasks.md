# kan-366-review-panel-dispatches-a-mutating-slot-with

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

- [x] 1. Isolate Bugbot's dispatch in a throwaway worktree

**Build:** green
**Files:** `skills/flow/review-panel.md`, `scripts/check-contract-budget.sh`
**Tests:** none
**Regression:** none
**Baseline:** before=43 after=43
<!-- measured: scripts/run-guard-tests.sh @ branch spectre/kan-366-review-panel-dispatches-a-mutating-slot-with (44 harnesses, 43 passed, 1 pre-existing unrelated failure in test-check-task-build-green.sh's case 13, caused by this checkout carrying both spectre/changes/ and openspec/changes/ trees) -->
**Commit:** `flow(review-panel): isolate bugbot's mutation testing in a throwaway worktree`

  - [x] **Step 1: Add a "Bugbot's throwaway worktree" subsection**, placed directly after
    "### Bugbot's mutation-testing brief" in `skills/flow/review-panel.md`. State: Bugbot mutates
    code in place while every other slot only reads it, and dispatching it into the same worktree a
    reading slot concurrently reads is the KAN-366 collision. Bugbot's dispatch — pass 1, and every
    fix-round re-run, a substituted general-purpose Bugbot included — therefore runs against a
    throwaway worktree, never the shared `<worktree>` the other slots read.

```bash unverified:confirm `git apply` accepts a diff produced by `git diff --binary` unmodified when applied from a fresh `--detach` worktree checked out at the same HEAD the diff was taken against
git -C <worktree> worktree add --detach <worktree>-bugbot-<round> HEAD
git -C <worktree> diff --binary | git -C <worktree>-bugbot-<round> apply
git -C <worktree> status --porcelain | awk '/^\?\? /{print substr($0,4)}' | \
  while read -r f; do mkdir -p "<worktree>-bugbot-<round>/$(dirname "$f")"; \
  cp -a "<worktree>/$f" "<worktree>-bugbot-<round>/$f"; done
```

    Dispatch Bugbot with `Full Repository Path: <worktree>-bugbot-<round>` in place of `<worktree>`.
    Remove the copy unconditionally once that dispatch closes — completed, timed out (including
    after the wall-clock re-dispatch), or the run stopped:

```bash unverified:confirm `git worktree remove --force` succeeds against a worktree carrying only the transplanted diff and no further commits
git -C <worktree> worktree remove --force <worktree>-bugbot-<round>
```

    State plainly: findings and reproducers are unaffected — a finding's `file:line` is
    repo-relative, and every reproducer still runs against the real `<worktree>` at verification
    time, never against Bugbot's copy, exactly as today. Security is **not** isolated this way:
    nothing in this file requires it to mutate anything, so it keeps sharing `<worktree>` with the
    reading slots.

  - [x] **Step 2: Cross-reference the new subsection** from its two other call sites, without
    duplicating its procedure: the "How to spawn" cell of Bugbot's row in "The roster" table gains
    "(own throwaway worktree — see **Bugbot's throwaway worktree** below)", and **An unspawnable id
    is substituted, not skipped** gains one sentence: "A substituted Bugbot's dispatch runs against
    the same throwaway worktree treatment as the real slot (**Bugbot's throwaway worktree**), since
    it carries the same mutation-testing brief."

  - [x] **Step 3: Run the lint list**, fixing any hit:

```bash unverified:run after step 1-2 land; predicts a clean run since the edit adds prose and fenced blocks that follow existing conventions (tagged code blocks, no bare numeric claims, required dispatch paragraphs untouched)
scripts/check-markdown-integrity.py
scripts/check-dispatch-paragraphs.sh
scripts/check-references.sh
scripts/check-normative-inventory.sh
scripts/check-contract-budget.sh
```

    `check-contract-budget.sh` is expected to fail here — this task's own edit is the reason
    **Step 4** exists.

  - [x] **Step 4: Raise `skills/flow/review-panel.md`'s budget row** in
    `scripts/check-contract-budget.sh`'s `budgets()` to the file's new size (measured after step 1-2
    land) times 1.25, rounded up to the nearest integer, replacing the existing `27420` — the
    convention the file's own comment states ("the size its file had when the change that added its
    row landed, plus 25%"). Re-run `scripts/check-contract-budget.sh` and confirm it now exits 0.

- [x] 2. Register the throwaway copy in the temporary artifacts registry

**Build:** green
**Files:** `skills/flow-contracts/artifacts-registry.md`, `scripts/check-contract-budget.sh`
**Tests:** none
**Regression:** none
**Baseline:** before=43 after=43
<!-- measured: scripts/run-guard-tests.sh @ branch spectre/kan-366-review-panel-dispatches-a-mutating-slot-with (same pre-existing, unrelated failure noted on task 1) -->
**Commit:** `flow(artifacts-registry): register bugbot's throwaway worktree copy`

  - [x] **Step 1: Add a row** to the table in `skills/flow-contracts/artifacts-registry.md`:
    `| Bugbot's throwaway worktree copy | \`/flow\`'s review panel | sibling of the apply worktree, \`<worktree>-bugbot-<round>\` | the review panel itself, immediately after that Bugbot dispatch closes — never survives to run 2 |`.

  - [x] **Step 2: Run the lint list**, fixing any hit:

```bash unverified:run after step 1 lands; predicts a clean run for the same reason as task 1's step 3, save for the budget row this step's own Step 3 addresses
scripts/check-markdown-integrity.py
scripts/check-references.sh
scripts/check-contract-budget.sh
```

  - [x] **Step 3: Raise `skills/flow-contracts/artifacts-registry.md`'s budget row**, same method as
    task 1's step 4, replacing the existing `7620`. Re-run `scripts/check-contract-budget.sh` and
    confirm it now exits 0.
