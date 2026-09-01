# kan-371-myflow-rebase-onto-the-base-branch-before-the

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

- [x] 1. Rebase onto the base branch before the landing question, on overlap

  In `skills/flow/integrate.md`'s **2. Ask how the branch should land**, where an overlap already
  triggers the existing base-moved prompt:

  1. Replace the two-option prompt with three, per design.md's `rebase-is-a-confirmed-choice`:

     > **The base branch has moved and touches paths this change also touched — how should
     > integration proceed?**
     > - **Stop — I'll rebase or reorder first** *(recommended)*
     > - **Rebase onto `<base>` now, then continue**
     > - **Continue — land anyway**

  2. On the new **Rebase** choice, run `git -C <worktree> rebase origin/$BASE`.

     - **Clean** (exit 0): the merge base carried forward for the rest of this run becomes
       `origin/$BASE`'s resolved tip at rebase time. Re-run `check-base-moved.sh` once more against
       that new merge base; a fresh `MOVED` overlap re-offers this same three-option prompt rather
       than looping silently. Otherwise, proceed to step 3, then the landing question.
     - **Conflict** (non-zero exit): never auto-abort. Leave the worktree mid-rebase, report the
       conflicting file(s) from `git status`, and hand off `git -C <worktree> rebase --continue`
       (after resolving) or `git -C <worktree> rebase --abort` as the operator's next manual step.
       State stays `IN_PROGRESS`; this run stops here, exactly as **Stop** already does.

  3. **Scoped re-verification**, per design.md's `scoped-reverify-not-full-suite`: for each path
     `check-base-moved.sh` reported under `overlaps:`, look for a discoverable guard test —
     `scripts/test-<basename-without-ext>.sh` beside `scripts/<name>.sh`, the same naming
     `scripts/run-guard-tests.sh` already discovers by glob — and run it if found. A path with none
     is stated in the handoff as having no verification to run, not silently skipped. A non-zero
     exit from any discovered test blocks this stage exactly like any other verify-stage failure:
     report it, leave the change `IN_PROGRESS`, stop before the landing question. **Never** re-run
     the project's whole `## lint`/`## test` list here.

  4. Qualify **## No verification gate**'s "Run no tests, no linters, and no spec-coverage check"
     with a one-sentence carve-out naming this one exception — the scoped re-verification above,
     triggered only by a rebase this stage itself performed, never a general re-opening of the
     no-verification-gate rule.

  5. Add the new option's outcome to the wall-clock-free control flow already documented for
     **Stop**/**Continue** in this section, and to the `flow.landing-question` mark's existing
     `outcome` values where the file already enumerates them (`completed`/`stopped`) — a rebase
     that stops on conflict closes the mark `stopped`, exactly like **Stop**; a rebase that lands
     cleanly and proceeds closes it `completed`, exactly like **Continue**.

  Run `scripts/check-contract-budget.sh` and reconcile this file's row in this commit, if the edit
  moves it across a budget boundary.

**Files:** `skills/flow/integrate.md`, `scripts/check-contract-budget.sh` (only if its row needs
reconciling)
**Tests:** **none** added — this task adds prose to a skill file with no executable guard of its
own; verified by `scripts/check-references.sh`, `scripts/check-contract-budget.sh` and
`scripts/check-vocabulary.sh`, none of which this task's own commit adds a new test to
**Regression:** reverting this commit removes the in-pipeline rebase option entirely, leaving the
operator back to a manual detour outside `/flow` on every base-branch overlap, exactly the KAN-304
friction this change exists to close
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `feat(flow): rebase onto the base before the landing question, on overlap`
**Build:** green
