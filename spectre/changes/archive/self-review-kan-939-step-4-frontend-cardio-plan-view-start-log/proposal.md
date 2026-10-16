# self-review-kan-939-step-4-frontend-cardio-plan-view-start-log

## Why

The kan-939 step-4 self-review pass found eight non-`big` defects in the pipeline and in
gymie-frontend's project configuration:

- F-1: integrate's `--plan-message` replaces the planning commit's whole message, so a planning
  message that lists outstanding work and has no Conventional Commits subject stops `commit` at
  commit-split after the reshape has already run.
- F-2: the push-time `check-untasked-commits.sh` check runs only on the full-suite-fix route, so a
  panel-fix commit can be pushed without its `Flow-Key:` trailer.
- F-3: signal four of `check-unfinished-work` accepts a ticked appended task whose dispatch row was
  back-filled even when no commit carrying `Task-Id: <n>` exists on the branch.
- F-4: `check-visual-verify-dispatched` accepts any completed verifier row, even when UI-path
  commits landed after the HEAD it verified.
- F-5: back-filled `flow record dispatch` rows take the recording moment as their start and end;
  only close-task's message asks for real `-started-at`/`-ended-at`.
- F-6: a panel round dispatched with no pass row goes unflagged, because the gate has no read verb
  for pass rows.
- F-7: gymie-frontend's full suite never runs `:shared:desktopTest`; its `.flow/project.md` has no
  `## test`, and `./gradlew test` does not depend on the desktop UI tests.
- F-8: `flow record dispatch begin` stores `-model` verbatim, so a parent's inline row records
  `claude-opus-5-5` beside dispatches recorded as `opus`.

Planning found one more, on `origin/main`: `skills/flow/verify-and-handoff.md` cites
`scripts/check-visual-verify-dispatched.sh` without a root, which `check-installed-citations.sh`
refuses.

## What changes

- `integrate-change.sh commit` takes `--outstanding <text>` and writes `Outstanding: <text>` into
  the planning commit's body under its own subject; `--plan-message` is gone.
- **Commits outside every task** (`skills/flow-contracts/git-boundaries.md`) owns the push-time
  untasked-commit check; the full-suite fix and the panel fix cite it.
- `check-unfinished-work.sh` takes `<worktree>=<merge-base>` ranges and reports a ticked appended
  task with no `Task-Id: <n>` commit in any of them; it also reports a panel round dispatched with
  no pass row, read through the new `flow record passes`.
- A verifier row counts only for the HEAD it recorded as `-diff-base`, and only while no UI path the
  change touches has changed since.
- `flow record dispatch backfill` records a finished dispatch in one call with its real instants
  required, and the pipeline's back-fills use it.
- gymie-frontend's `## test` runs `./gradlew test :shared:desktopTest`.
- `flow record dispatch begin` records `claude-opus-*` and `claude-sonnet-*` as `opus` and `sonnet`.
- The verify-and-handoff citation names its root.
