## Context

- Six findings of the kan-941 self-review pass planned per item 1 of **3. Fix every finding that is
  not `big`** (`skills/flow-self-review/SKILL.md`), on `self-review-kan-941-step-6-playwright-fidelity-specs-for-the-cardio`
  based at `f4844c40`. Each finding is carried by exactly one task; the pass report's F7 (an
  unlanded branch) is not part of this change. The brainstorm design is
  `.superpowers/sdd/2026-10-18-self-review-kan-941-step-6-playwright-fidelity-specs-for-the-cardio-design.md`
  in this worktree.
  - F1 (Task 1), blast radius ~15: `reshapeBranch` (`stats/internal/guard/reshapebranch.go`) runs
    `git reset --soft <merge-base>` with no look at what `<merge-base>..HEAD` holds. The pipeline
    never merges into a change branch — its sync rebases (`skills/flow/sync-onto-base.md`), and no
    script runs `git merge` into one — so a commit off the first-parent chain is always another
    branch's. Reproduced in Task 1's case 3b: a `develop` commit merged into the branch is reset
    into the index beside the task's work.
  - F2 (Task 2), blast radius 53: `uwAppendedIn` (`stats/internal/guard/unfinishedwork.go`) returns
    the last `**Tasks appended:** <n>` tasks; `skills/flow/verify-fix-loop.md` step 1 says the loop
    does not raise that count and titles its task `In-run fix <k> — <stage>: <the defects>`. The
    kan-941 plan holds no `Tasks appended` line, so its in-run tasks were never judged. Reproduced
    in Task 2's new case: an in-run fix task with no dispatch row is CLEAR at `f4844c40`.
    `ctUndispatched` (`stats/internal/guard/closetask.go`) shares `uwAppended`, so close-task's
    refusal of an undispatched appended task covers in-run tasks too.
  - F3 (Task 3), blast radius 3: `FlagVerdictFalsePositive` (`stats/internal/store/records.go`)
    flags the newest verdict of (project, change, guard) and refuses with `ErrDispatchNotFound`
    when there is none; the project is `-C`'s; `check-unfinished-work` records every worktree's
    verdict under the canonical project (STORE_ANCHOR). A satellite `-C` therefore always fails,
    and N calls at the canonical worktree would flag the same row N times.
  - F4 (Task 4), blast radius 5: `checkUntaskedCommits` runs `git log --no-merges <base>..HEAD`,
    which walks a merge's second parent. Reproduced in Task 4's new test.
  - F5 (Task 5), blast radius ~45: `check-plan-shape.py` has no check over where a task's files
    lie. Reproduced: a one-task plan whose `**Files:**` is `spectre/changes/x/visual-sweep.md`
    exits 0. Measured over every `tasks.md` under this repository's and gymie's `spectre/changes/`:
    13 tasks match, all in archived gymie plans (old live-verification tasks and kan-941's Task 4);
    no in-flight plan matches, and `scripts/check-plan-shape.sh` with no argument still exits 0
    with the check applied.
  - F6 (Task 6), blast radius ~6: `skills/flow-contracts/git-boundaries.md` § Commits outside every
    task requires the trailer and a pre-push `check-untasked-commits.sh` run, and nothing runs it.
    A change branch is pushed mid-run at two points: the parent's own `git push origin
    spectre/<name>` (**Branch backup**), which `hooks/protect-main-checkout.py` already parses, and
    `close-task.sh`'s push after each task close.
- Spiked at plan time, every edit reverted: the six diffs apply in task order at `f4844c40`
  (`git apply`); with all applied the whole `## lint` list exits 0 save `check-verbatim-moves.sh`,
  whose new and reworded sentences this change lists in its `verbatim-moves.txt`;
  `scripts/run-guard-tests.sh` passes (42 harnesses), and `go test ./internal/guard/` passes. Each
  task's new case fails with its fix reverted and its test kept.

## Decisions

### A merge of another branch stops the reshape

**ID:** reshape-foreign-merge-stop
**Status:** active
**Chosen:** `reshapeBranch`, after the base resolves and before the reset, lists
`rev-list <merge-base>..HEAD` minus `rev-list --first-parent <merge-base>..HEAD`; any commit left
prints `RESHAPE-FOREIGN-MERGE: <worktree> — <n> commit(s) …`, one `<sha12> <subject>` line each and
the remedy line, and exits 1 with HEAD untouched. A merge whose side the merge base already reaches
(a sync merge the merge base moved past) brings nothing and passes.
**Considered:** judging against `origin/<base>` as the finding words it — the script takes no base
branch, so a new argument through `integrate-change.sh`, and a merge of the base itself past the
recorded merge base would still fold base commits under the change's subject; refusing any merge
commit at all — a merge whose side the merge base already reaches folds nothing; refusing in
`integrate-change.sh prepare` before any worktree reshapes — a second copy of the check, while a
refusal mid-loop leaves an earlier worktree reset to its merge base with its work staged, which a
re-run reshapes again as a no-op.

### In-run fix tasks count as appended

**ID:** in-run-fix-tasks-appended
**Status:** active
**Chosen:** `uwAppendedIn` also returns every task whose line matches
`^- \[[ xX]\] [0-9]+\. In-run fix [0-9]+ — `, in plan order, unioned with the last-`<n>` set, so
signal four's dispatch-row, review-row and commit halves and close-task's `ctUndispatched` all judge
them.
**Considered:** raising `**Tasks appended:**` in the verify-fix loop — the count feeds document-fix's
appended-task budget, which counts appends made at the human gate only; a separate in-run list in
signal four — close-task's warning would then still miss them.

### The structural-verdict flag runs once, at the canonical worktree

**ID:** structural-verdict-canonical
**Status:** active
**Chosen:** `skills/flow/unfinished-work-gate.md` runs `flow record verdict false-positive … -C
<canonical-worktree>` once, stating why: every verdict lives under the canonical project and the
command flags its newest.
**Considered:** a `-worktree` flag on the command to flag each worktree's row — the store has no such
selector, and the gate's Continue is one decision over the whole change.

### check-untasked-commits reads the first-parent chain

**ID:** untasked-first-parent
**Status:** active
**Chosen:** `--first-parent` beside `--no-merges`.
**Considered:** subtracting the merge's second-parent range by hand — `--first-parent` is git's own
form of it.

### check-plan-shape refuses a planning-paths-only task

**ID:** plan-shape-planning-only-task
**Status:** active
**Chosen:** finding F14 in `_check_task`: a task whose every `**Files:**` path starts with
`spectre/changes/`, `openspec/changes/` or `docs/superpowers/` (`PLANNING_PREFIXES`), reported at the
task line, tagged `# F14` for the harness's mutation case.
**Considered:** resolving the project's spec leaf as `check-task-commit-planning-paths` does — the
plan guard reads a file with no project around it, and naming both leaves is a superset that no
implementation path matches; refusing any planning path in `**Files:**` — a task may legitimately
carry a record beside its real work.

### Flow-Key enforced at both mid-run push points

**ID:** flow-key-at-push
**Status:** active
**Chosen:** `hooks/protect-main-checkout.py`'s `untasked_push` runs `check-untasked-commits.sh
<repo> refs/remotes/origin/<dst>` for a typed push of HEAD's commit to a `spectre/` branch origin
already holds, and denies on exit 1; any other outcome, a missing checker or a 60s timeout lets the
push through (the hook's fail-open contract — run 1's gate still lists the commit). `close-task.sh`
runs the same guard per worktree on the same range before its push, its exit 1 joining the commit
guards' RE-COMMIT refusal. `UNTASKED_COMMITS_CHECK` overrides the checker for the hook's harness
only, as `LOOP_LANDING_CHECK` does.
**Considered:** a git `pre-push` hook — worktrees share the repository's config, so installing one
reaches the operator's own pushes in every project; the hook alone — close-task pushes whatever the
parent committed and did not push; `<merge-base>..HEAD` as the range — an already-pushed commit can
no longer be amended, and run 1's gate owns it; failing closed like the loop-landing check — a
backup push blocked by a checker that cannot build stops work the gate backstops anyway.

### First review effort and fixer grouping

**ID:** first-review-effort
**Status:** active
**Chosen:** first review on `flow-medium` (effort `medium`); one fixer dispatch (key
`self-review-kan-941-step-6-playwright-fidelity-specs-for-the-cardio-fix`) carrying Tasks 1–6 in
order — each task is a verified diff applied with `git apply`.
**Considered:** `flow-low` for the first review — Tasks 1 and 6 add refusals to the reshape and to
every change-branch push, where a false refusal stops runs, judgment a low-effort read may skim;
one fixer per task — six dispatches for six `git apply`s.

## Open questions

## Live check

No runtime to verify against: the reshape, the guards and the hook run per invocation and hold no
state; their harness cases exercise each against real git repositories.
