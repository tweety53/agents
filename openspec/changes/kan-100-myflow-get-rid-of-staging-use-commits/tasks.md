> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

## 1. Retire the checkpoint/review-package mechanism

### 1.1 Delete `skills/myflow-do/scripts/checkpoint` and `skills/myflow-do/scripts/uncommitted-review-package`

- [x] Delete `skills/myflow-do/scripts/checkpoint`, `skills/myflow-do/scripts/uncommitted-review-package`,
  and `scripts/test-uncommitted-review-package.sh`.
- [x] Confirm no other file still shells out to them:
  `grep -rn "scripts/checkpoint\|uncommitted-review-package" skills/ scripts/ .myflow/ | grep -v openspec/changes/archive`
  returns nothing once tasks 2–6 below are also done — this task only removes the scripts and their
  test harness, so a hit here now is expected until later tasks land.
- [x] Verify: `test ! -e skills/myflow-do/scripts/checkpoint && test ! -e skills/myflow-do/scripts/uncommitted-review-package && test ! -e scripts/test-uncommitted-review-package.sh`

**Build:** green

### 1.2 Remove the retired test entry from `.myflow/project.md`'s `## test` list

- [x] Drop the `scripts/test-uncommitted-review-package.sh` line from `.myflow/project.md`'s
  `## test` section.
- [x] Verify: `grep -c "test-uncommitted-review-package" .myflow/project.md` prints `0`

This task is `Build: red` on its own because task 3.3 adds the new guard's test line to the same
list in the same edit pass — done together to avoid two separate edits to one file.

**Build:** red
**Squash-with:** Task 3.3

## 2. Commit-per-task mechanism in `/myflow-do`

### 2.1 Replace the `NO-COMMITS` dispatch clause with `COMMIT-PER-TASK` in `skills/myflow-do/SKILL.md` section 4

- [x] Replace the `NO-COMMITS` block (the implementer dispatch clause requiring all four
  sub-blocks) with a `COMMIT-PER-TASK` block per `myflow-task-commits`'s requirement **Each task
  commits after finishing, before review**: implementer commits `task(<n>): <subject>` with a
  `Task-Id: <n>` trailer immediately after RED-GREEN-REFACTOR completes, before dispatch to review.
- [x] Remove the `TASK_BASE=$(skills/myflow-do/scripts/checkpoint)` / `uncommitted-review-package`
  sentence from that block.
- [x] Verify: `grep -n "COMMIT-PER-TASK\|NO-COMMITS" skills/myflow-do/SKILL.md` shows
  `COMMIT-PER-TASK` present and `NO-COMMITS` absent from the dispatch clause (a mention inside a
  changelog-style note is acceptable; the operative block is not).

**Build:** green

### 2.2 Change per-task and fix-round review to read a commit-range diff

- [x] In `skills/myflow-do/SKILL.md` section 4 and section 5 ("Panel re-runs"), replace every
  reference to `uncommitted-review-package`/`checkpoint`-based review packages with
  `git diff <task-base>..<task-sha>` (or the fixup-folded equivalent) as the diff the reviewer
  reads, per `myflow-task-commits`'s requirement **Review reads a real commit diff, not a
  checkpoint snapshot**.
- [x] Verify: `grep -n "checkpoint\|uncommitted-review-package" skills/myflow-do/SKILL.md` returns
  nothing.

**Build:** green

### 2.3 Add the fixup-and-autosquash procedure for fix rounds

- [x] In `skills/myflow-do/SKILL.md` section 5 ("Panel re-runs"), state that a fix-round change
  commits as `git commit --fixup=<task-sha>` and is folded in with `git rebase --autosquash`
  immediately, before the next review pass reads the diff — per `myflow-task-commits`'s requirement
  **A fix-round change commits as a fixup and is autosquashed immediately**.
- [x] Verify: `grep -n -- "--fixup\|--autosquash" skills/myflow-do/SKILL.md` shows both present in
  section 5.

**Build:** green

### 2.4 Add the red-task-folds-into-green-partner procedure

- [x] In `skills/myflow-do/SKILL.md` section 4, state that a `Build: red` task's commit folds into
  its `Squash-with:` partner's commit via the same fixup mechanism, per `myflow-task-commits`'s
  requirement **A red task's commit folds into its green partner**.
- [x] Verify: `grep -n "Squash-with" skills/myflow-do/SKILL.md` shows it referenced in section 4.

This is `Build: red` because it reads the `Squash-with:` field task 4.1 introduces in
`build-green.md` — done together so the field name is defined and consumed in one pass.

**Build:** red
**Squash-with:** Task 4.1

### 2.5 Rewrite the Git boundaries table in `skills/myflow-contracts/pipeline.md`

- [x] Update the `/myflow-do` rows: from `STARTED` it now creates the branch/worktree and
  **commits each task + fixups** (no longer "git add excluding planning paths, no commits"); from
  `IN_PROGRESS` with no `prUrl` it resumes and **commits fixups**; the `prUrl` row is unchanged
  (still commits twice and pushes at the end).
- [x] Update the guarded-commit bash snippet's surrounding prose if it now describes a mid-run
  per-task commit rather than only the two staging-time commits — the snippet itself
  (implementation commit, then planning commit) still describes `/myflow-finish`'s two commits and
  does not change.
- [x] Verify: `grep -n "no commits\|No commits\|excluding the planning paths — no commits" skills/myflow-contracts/pipeline.md`
  no longer describes `/myflow-do`'s `STARTED`/`IN_PROGRESS`-no-`prUrl` rows as commit-free.

**Build:** green

### 2.6 Update `skills/myflow-do/SKILL.md`'s Guardrails and handoff template

- [x] Remove the guardrail "Never commit, push, merge, or open a PR — except the `prUrl` exception
  above" and replace it with one distinguishing task/fixup commits (now expected) from
  push/merge/PR (still forbidden on every run).
- [x] Update the handoff's `Staged:` line — it currently reads `N/N tasks · staged and uncommitted
  | committed as two commits and pushed to the PR branch`; add the ordinary (non-`prUrl`) case,
  which is now `N/N tasks committed on branch` rather than `staged and uncommitted`.
- [x] Update the "Review the diff" git commands in the same block to read the branch's own commit
  range instead of `--cached`.
- [x] Verify: `grep -n "staged and uncommitted" skills/myflow-do/SKILL.md` returns nothing; the
  ordinary case's wording names commits, not staging.

**Build:** green

## 3. Mechanically-checked per-task fields

### 3.1 Add the `Files:`/`Tests:`/`Regression:`/`Baseline:`/`Commit:` tag family to `/myflow-start`'s writing-plans stage

- [x] In `skills/myflow-start/SKILL.md` section **D. Basic Workflow #3 — Writing plans**, add the
  requirement that every task written during writing-plans carries `**Files:**`, `**Tests:**`,
  `**Regression:**`, `**Baseline:**`, and `**Commit:**` tags (plus `**Squash-with:**` for
  `Build: red` tasks, added in task 4.1), per `myflow-task-commit-fields`'s requirement **Every
  task declares mechanically-checkable fields**.
- [x] Verify: `grep -n "Files:\|Tests:\|Regression:\|Baseline:\|Commit:" skills/myflow-start/SKILL.md`
  shows all five named in section D.

**Build:** green

### 3.2 Write `scripts/check-task-commit-fields.py`

- [x] Model it on `scripts/check-task-build-green.py` (parses a single `tasks.md`, one class per
  violation). This script additionally takes a worktree path, a task id, and the task's commit sha
  (and its parent), and checks the three mechanically-verifiable-without-running-tests fields —
  `Files:`, `Tests:`, `Commit:` — per `myflow-task-commit-fields`'s requirement **A runtime guard
  checks each field against the real commit**.
- [x] `Regression:` and `Baseline:` are implemented as separate functions in the same module (task
  3.4) since they need the project's `## test` command, not just git — stub them so the module
  imports cleanly, without wiring them into the CLI entry point yet.
- [x] Verify: `python3 -c "import ast; ast.parse(open('scripts/check-task-commit-fields.py').read())"`
  exits 0.

This task is `Build: red` because it has no test harness to run against until task 3.3 exists — the
two are dispatched as one unit.

**Files:** `scripts/check-task-commit-fields.py`
**Tests:** Case 1: files subset of declared passes; Case 2: undeclared file fails; Case 3: declared
test name found in diff passes; Case 4: missing declared test fails; Case 5: commit subject matches
declared `Commit:` passes; Case 6: commit subject mismatch fails; Case 7: extra path covered by
`Allowed-collateral:` glob passes
**Regression:** Case 2 (undeclared file fails): reverting this task's guard logic makes an
undeclared-file commit pass silently, which this case catches by asserting a non-zero exit.
**Baseline:** not applicable — this task has no runnable harness of its own; task 3.3's harness
covers both, per the merge pairing above.
**Commit:** `feat(kan-100-myflow-get-rid-of-staging-use-commits): add the task-commit-fields guard`
**Build:** red
**Squash-with:** Task 3.3

### 3.3 Write `scripts/check-task-commit-fields.sh` wrapper and `scripts/test-check-task-commit-fields.sh`

- [x] Write the wrapper following `scripts/check-task-build-green.sh`'s shape (resolves which
  `tasks.md`/worktree to check, execs the Python script).
- [x] Write the test harness following `scripts/test-check-task-build-green.sh`'s shape: a
  throwaway git repo per case, asserting exit codes and messages for each scenario in
  `myflow-task-commit-fields`'s spec (undeclared file, missing test, subject mismatch, clean pass —
  `Regression:`/`Baseline:` skip-not-verified cases are added in task 3.4).
- [x] Add `scripts/test-check-task-commit-fields.sh` to `.myflow/project.md`'s `## test` list, and
  remove `scripts/test-uncommitted-review-package.sh` from the same list — task 1.2 depends on this
  edit landing together with it.
- [x] Verify: `bash scripts/test-check-task-commit-fields.sh` exits 0.

**Build:** green
**Files:** `scripts/check-task-commit-fields.sh`, `scripts/test-check-task-commit-fields.sh`, `.myflow/project.md`
**Tests:** the 7 cases listed in task 3.2 (Case 1-7), run for the first time against the wrapper
this task adds; plus three cases verifying the `Tests:` field's prose-parsing grammar that this
task's guard implementation depends on and that task 3.2 does not name individually: Case 8: a
`Tests:` field written as free prose naming `Case N:` labels passes when both labels appear in the
diff; Case 9: the same prose shape fails, naming the specific missing case, when one label is
absent from the diff; Case 10: a `Tests:` field with neither `Case N:` labels nor backtick-quoted
identifiers declares no checkable tests at all and never false-fails
**Regression:** Case 2 (undeclared file fails): if the wrapper mis-resolves which `tasks.md`/worktree
to check, this case fails against the wrong fixture instead of the intended one.
**Baseline:** before=0 after=10 cases in `scripts/test-check-task-commit-fields.sh` (new harness; no
prior tests for this script)
<!-- measured: bash scripts/test-check-task-commit-fields.sh @ branch openspec/kan-100-myflow-get-rid-of-staging-use-commits -->
**Commit:** `test(kan-100-myflow-get-rid-of-staging-use-commits): add the task-commit-fields harness`

### 3.4 Implement `Regression:` and `Baseline:` checks with skip-not-verified fallback

- [x] Implement `Regression:` per `myflow-task-commit-fields`'s requirement **Regression and
  Baseline checks skip, rather than fail, when unsupported**: revert the task commit, run the named
  test via the project's `## test` command, confirm it fails, un-revert.
- [x] Implement `Baseline:`: run `## test` at the parent commit and at the task commit, compare
  counts to the declared `before=<N> after=<N>`.
- [x] Implement the skip-not-verified fallback for both: report skipped, not failed, when `## test`
  can't target a named test or doesn't report a parseable count.
- [x] Wire both into the script's CLI entry point (stubbed in task 3.2).
- [x] Verify: `bash scripts/test-check-task-commit-fields.sh` (re-run; now covers all cases) exits 0.

**Build:** green
**Files:** `scripts/check-task-commit-fields.py`, `scripts/test-check-task-commit-fields.sh`
**Tests:** Case 11: regression check passes (revert makes named test fail, un-revert restores it);
Case 12: regression check skips when the project's `## test` command can't target a named test;
Case 13: baseline check passes (parent/task counts match declared `before=<N> after=<N>`); Case 14:
baseline check skips when the count is unparseable
**Regression:** Case 12 (regression skip): reverting the skip-fallback logic makes an unsupported
project's run fail instead of skip, which this case catches.
**Baseline:** before=10 after=14 cases in `scripts/test-check-task-commit-fields.sh`
<!-- measured: bash scripts/test-check-task-commit-fields.sh @ branch openspec/kan-100-myflow-get-rid-of-staging-use-commits (before, from task 3.3, 10 cases) -->
<!-- predicted: bash scripts/test-check-task-commit-fields.sh after this task (14 cases) -->
**Commit:** `feat(kan-100-myflow-get-rid-of-staging-use-commits): add Regression/Baseline checks with skip fallback`

### 3.5 Wire the guard into `/myflow-do` section 4, right after each task's commit

- [x] In `skills/myflow-do/SKILL.md` section 4, add: immediately after the implementer's
  `task(<n>):` commit and before dispatch to review, run
  `scripts/check-task-commit-fields.sh <worktree> <task-id>` (or the equivalent when the script is
  absent — apply `myflow-task-commit-fields`'s rules by hand, per this repo's own established
  pattern for an absent guard script).
- [x] State that a guard failure sends the task back to the same implementer, per
  `myflow-task-commit-fields`'s requirement **A runtime guard checks each field against the real
  commit** — it is not a review finding and does not consume a review round.
- [x] Verify: `grep -n "check-task-commit-fields" skills/myflow-do/SKILL.md` shows it invoked in
  section 4, before the review-dispatch step.

**Build:** green

## 4. `Build: red` migrates to `Squash-with:`

### 4.1 Update `skills/myflow-contracts/build-green.md`'s tag syntax and guard description

- [x] Change `**Build:** red — merges with Task <N>` to bare `**Build:** red`, with the partner now
  named by `**Squash-with:** Task <N>` (defined in `myflow-task-commit-fields`).
- [x] Update the guard's scope section: it now reads the partner from `Squash-with:` rather than
  parsing it out of the `Build:` tag text.
- [x] Verify: `grep -n "merges with Task" skills/myflow-contracts/build-green.md` returns nothing;
  `grep -n "Squash-with" skills/myflow-contracts/build-green.md` shows it present.

**Build:** green

### 4.2 Update `scripts/check-task-build-green.py` to parse `Squash-with:` instead of the inline red-tag suffix

- [x] Change the parser to read the red-task partner from a task's `**Squash-with:**` field instead
  of from the `Build:` tag's own inline suffix.
- [x] Fail with a named violation when a `Build: red` task carries no `Squash-with:` field.
- [x] Verify: `python3 -c "import ast; ast.parse(open('scripts/check-task-build-green.py').read())"`
  exits 0.

`Build: red` on its own because it and task 4.3 (updating the test fixtures) touch the same
partner-parsing logic and must land together for the test suite to pass at either commit alone.

**Build:** red
**Squash-with:** Task 4.3
**Files:** `scripts/check-task-build-green.py`
**Tests:** Case 15: red task reads its partner from `Squash-with:`; Case 16: `Build: red` with no
`Squash-with:` field fails; Case 17: `Squash-with:` partner must itself be tagged `green`
**Regression:** Case 16 (missing `Squash-with:` fails): reverting this parser change makes a
`Build: red` task with no `Squash-with:` field pass silently (falling back to the old inline-suffix
parse, which now finds nothing there either but previously treated the absence as a different
violation shape); this case catches the exit-code and message change.
**Baseline:** before=14 after=17 cases in `scripts/test-check-task-build-green.sh`
<!-- measured: bash scripts/test-check-task-build-green.sh @ branch main (14 cases, current script) -->
<!-- predicted: bash scripts/test-check-task-build-green.sh after task 4.3 -->
**Commit:** `refactor(kan-100-myflow-get-rid-of-staging-use-commits): read the red-task partner from Squash-with`

### 4.3 Update `scripts/test-check-task-build-green.sh` fixtures to the new syntax

- [x] Replace every fixture's `**Build:** red — merges with Task <N>` with bare `**Build:** red`
  plus a separate `**Squash-with:** Task <N>` line, matching task 4.1's new syntax.
- [x] Verify: `bash scripts/test-check-task-build-green.sh` exits 0.

**Build:** green

## 5. Finish reshapes the branch before its two commits

### 5.1 Add the squash-to-merge-base step to `skills/myflow-contracts/finish-contract.md`'s run 1 procedure

- [x] Before the existing two-commit sequence in run 1, add:
  `git -C <abs-worktree> reset --soft <recorded-merge-base>`, collapsing every per-task and fixup
  commit `/myflow-do` made back into the working tree.
- [x] Confirm everything after that step is unchanged, per `myflow-finish-cleanup`'s modified
  requirement **Run 1 asks how to land the branch, before doing anything**.
- [x] Verify: `grep -n -- "--soft" skills/myflow-contracts/finish-contract.md` shows the reset step
  present, ordered before the guarded-commits bash chain.

**Build:** green

### 5.2 Check `skills/myflow-finish/SKILL.md` for a restated copy of run 1's step sequence and update it

- [x] Confirm `skills/myflow-finish/SKILL.md`'s own "Run 1 — integrate" section cites
  `finish-contract.md` rather than restating its steps, per this repo's own citation convention.
- [x] If it does restate the commit sequence, add the reshape step there too so the two files do
  not disagree.
- [x] Verify: `grep -n "reset --soft\|Run 1" skills/myflow-finish/SKILL.md` — either the step is
  cited via the contract file (no restatement to update) or the restated sequence now includes it.

**Build:** green

## 6. Contract housekeeping

### 6.1 Update `check-contract-budget.sh`'s budget table for every file grown in this change

- [x] Run `scripts/check-contract-budget.sh` after all edits in sections 1–5 land.
- [x] For each file it flags — `skills/myflow-contracts/pipeline.md`,
  `skills/myflow-contracts/build-green.md`, `skills/myflow-contracts/finish-contract.md`,
  `skills/myflow-do/SKILL.md`, `skills/myflow-start/SKILL.md` are the candidates — raise that
  file's budget row to its new size plus 25%, per this repo's own rule that a budget hit means a
  deliberate table edit, never a narrowed guard.
- [x] Verify: `bash scripts/check-contract-budget.sh` exits 0.

**Build:** green

### 6.2 Add the new guard to `.myflow/project.md`'s `## lint` list

- [x] Superseded during implementation: `scripts/check-task-commit-fields.sh` was NOT added to the
  `## lint` list. Unlike `scripts/check-task-build-green.sh`, it needs `<worktree> <task-id>
  <commit-sha>` arguments and cannot run against a bare tree, so it cannot be a bare `## lint`
  entry — it matches this repository's own documented exception for parametrized guards
  (`check-finish-preflight.sh`, `preserve-session-records.sh`, `check-unfinished-work.sh` and
  `check-cleanup-complete.sh` in `.myflow/project.md`'s "deliberately not lint steps" note). This
  is a corrected design decision made mid-run, not an oversight; `.myflow/project.md` was correctly
  left unchanged.
- [x] The guard is instead covered under `## test` via `scripts/test-check-task-commit-fields.sh`,
  added by task 3.3.
- [x] Verify: `grep -n "scripts/test-check-task-commit-fields.sh" .myflow/project.md` shows it
  present in the `## test` section; `grep -n "deliberately not lint steps" .myflow/project.md`
  shows the general documented exception for parametrized guards is still present and unchanged —
  `check-task-commit-fields.sh` is not named there since the note predates this change and this
  task does not amend it.

**Build:** green

### 6.3 Run `scripts/check-references.sh` and `scripts/check-vocabulary.sh` and fix any hit

- [x] Run both guards and fix any hit — a named-section reference that no longer exists, or
  vocabulary drift the edits above introduced.
- [x] Verify: `bash scripts/check-references.sh && bash scripts/check-vocabulary.sh` both exit 0.

**Build:** green

## 7. Full verification pass

### 7.1 Run the complete `## test` and `## lint` command lists from `.myflow/project.md`

- [x] Run every command listed under `## test` and `## lint` in `.myflow/project.md`, confirming
  every guard and test harness — old and new — passes together, including the retired script's
  absence and the new guard's presence.
- [x] Verify: every command listed under `## test` and `## lint` in `.myflow/project.md` exits 0.

**Build:** green
