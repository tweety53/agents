# kan-285-myflow-prompt-a-rebase-before-the-review-panel

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

- [x] 1. Check base movement and offer a rebase at panel entry

  In `skills/flow/review-panel.md`, insert a new step between the `flow.review-panel` begin mark
  and **Run the citation pre-check before rebuilding the dispatch context bundle below** — the
  citation pre-check and the bundle rebuild both read the diff a rebase would change, so the check
  comes first. Per design.md's `panel-entry-only-integrate-kept`.

  - [x] **Step 1: State the check.** Once per worktree in this run's resolved set (**Resolving a
  change's worktrees**, `skills/flow-contracts/worktree-resolution.md`), on every panel run —
  creating, resumed, or fix:

  ```bash unverified:the same pair skills/flow/integrate.md step 2 runs; confirm resolve-base-branch.sh prints the bare base name on stdout before citing it this way
  BASE="$(resolve-base-branch.sh <worktree>)"
  check-base-moved.sh <worktree> "origin/$BASE" <working-notes-merge-base>
  ```

  `<working-notes-merge-base>` is the merge base `skills/flow/implement.md`'s isolate-workspace
  step recorded in this run's working notes — never the state file's `worktrees` map, which a
  creating run has not written yet. State that `check-base-moved.sh` performs no fetch and
  `resolve-base-branch.sh` is what fetches, so the order of the two lines is load-bearing.

  - [x] **Step 2: State the verdict handling**, citing integrate step 2 where the words are
  already there rather than restating them: every worktree's verdict is reported; `MOVED` with no
  overlap continues with no prompt; `REFUSE`, exit 2 with no verdict line, or an empty resolved
  set stops and asks; an overlap from any worktree asks once for the whole change, shape per
  Operator prompts (`skills/flow-contracts/operator-prompts.md`):

  ```markdown unverified:the wording of the third option differs from integrate's "land anyway" on purpose — confirm no vocabulary guard objects
  > **The base branch has moved and touches paths this change also touched — how should the
  > panel proceed?**
  > - **Stop — I'll rebase or reorder first** *(recommended)*
  > - **Rebase onto `<base>` now, then continue**
  > - **Continue — review as is**
  ```

  **Stop** closes `flow.review-panel` with `-outcome stopped` and leaves the change at its
  current state with nothing committed by this stage. **Continue** carries the reported movement
  into the handoff and proceeds to the citation pre-check.

  - [x] **Step 3: State the rebase outcome.** `git -C <worktree> rebase origin/$BASE` only in a
  worktree whose own verdict was `MOVED`, never one whose verdict was `CLEAR`.

  - **Clean** (exit 0): that worktree's working-notes merge base becomes `origin/$BASE`'s
    resolved tip at rebase time; every later `<merge-base>` in this file and the `worktrees` map
    `skills/flow/verify-and-handoff.md` writes read the working notes, so nothing else is
    plumbed. Re-run `check-base-moved.sh` once against the new value; a fresh overlap re-offers
    the prompt above rather than looping silently. Per design.md's
    `rebase-clears-held-slot-shas`: the rebase clears every slot's held last-reviewed sha, so a
    re-run after it reads the whole `final-review.diff` under the no-held-sha rule already stated
    in **Panel re-runs** — add one sentence there pointing back at this step. Per design.md's
    `no-scoped-reverification-at-panel-entry`: no re-verification runs here; say so in one
    sentence, naming that the panel reads the rebased tree and `flow.verify` follows. End the
    clean-rebase report with the fixed literal:

    > If this change's verification compares against a recorded baseline, recapture it now — a
    > proof taken against the pre-rebase base is void.

  - **Conflict** (non-zero exit): never auto-abort, and never resolve the conflict — by editing
    the conflicting files or otherwise. Leave the worktree mid-rebase exactly as `git rebase`
    left it, report the conflicting file(s) from `git status`, hand off
    `git -C <worktree> rebase --continue` (after the **operator** resolves) or
    `git -C <worktree> rebase --abort` as the next manual step, and close the mark `stopped`.

  - [x] **Step 4: Run the guards.** `scripts/check-references.sh`, `scripts/check-vocabulary.sh`
  and `scripts/check-contract-budget.sh`; reconcile `skills/flow/review-panel.md`'s row in the
  budget file in this same commit only if the edit pushes the file past 45213 bytes.
  <!-- measured: grep -n 'skills/flow/review-panel.md' scripts/check-contract-budget.sh @ branch main, 2026-09-03; the file is 38321 bytes today per wc -c -->

**Files:** `skills/flow/review-panel.md`, `scripts/check-contract-budget.sh` (only if its row
needs reconciling)
**Tests:** **none** added — this task adds prose to a skill file; `scripts/test-check-base-moved.sh`
already covers the guard it invokes, and the task's own commit adds no test
**Regression:** reverting this commit removes the panel-entry check entirely, so the panel again
reviews a diff against a possibly stale base and the movement is first seen at integrate step 2 or
the merge — the KAN-265 cost this change exists to close
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `feat(flow): check base movement and offer a rebase at panel entry`
**Build:** green

- [x] 2. Say a recorded baseline must be recaptured after an integrate-time rebase

  In `skills/flow/integrate.md`'s **2. Ask how the branch should land**, in the **Clean** branch of
  the **Rebase** outcome — the paragraph that begins "**Clean** (exit 0): the merge base carried
  forward for the rest of **this run**" — per design.md's `one-fixed-baseline-sentence`.

  - [x] **Step 1: Append the sentence.** After that branch's final instruction ("then proceed to
  the landing question"), add one sentence stating that the report ends with the same fixed
  literal task 1 carries:

  > If this change's verification compares against a recorded baseline, recapture it now — a
  > proof taken against the pre-rebase base is void.

  Byte-identical to task 1's literal — one sentence, two sites, no paraphrase.

  - [x] **Step 2: Run the guards.** `scripts/check-references.sh`, `scripts/check-vocabulary.sh`
  and `scripts/check-contract-budget.sh`; the edit is one sentence and the file sits 3304 bytes
  under its row, so no reconciliation is expected.
  <!-- measured: wc -c skills/flow/integrate.md (15298) against the 18602 row in scripts/check-contract-budget.sh @ branch main, 2026-09-03 -->

**Files:** `skills/flow/integrate.md`
**Tests:** **none** added — one sentence of prose in a skill file; no executable guard changes
**Regression:** reverting this commit drops the reminder at the integrate-time rebase, so a change
whose proof is relative to a recorded baseline can land on a stale one with no word said
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `feat(flow): say a recorded baseline must be recaptured after a rebase`
**Build:** green
