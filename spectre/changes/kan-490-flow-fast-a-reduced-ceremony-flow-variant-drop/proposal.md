# kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop

## Why

`/flow`'s phase files total roughly 10k lines and every run loads them plus a subagent-dispatched
planner and conductor, a settings-store review roster, and the full guard set — the right shape for
a change big enough to need SDD, multiple reviewers, and self-review. Small, low-risk changes pay
the same ceremony. There is no cheaper path through the same three-state pipeline today.

## What changes

- New `skills/flow-fast/` skill (`SKILL.md` router + slim `brainstorm.md`, `implement.md`,
  `review.md`, `finish.md`) plus `commands/flow-fast.md` and `commands-claude/flow-fast.md`,
  invoked as `/flow-fast`. Shares `skills/flow-contracts/*`, the reviewer prompts under
  `skills/flow/` and `skills/flow/scripts/` — restates nothing from `skills/flow/*.md`. Same state
  record and `flow.*` stage keys as `/flow`, so one change can move between the two commands.
- Brainstorm runs inline (no planner subagent), auto-picks the recommended option every round, and
  skips the design-approval gate — the operator's review happens once, on the staged `IN_PROGRESS`
  diff.
- Implementation runs inline (no conductor), TDD per task, one commit per task, targeted tests and
  lint on touched files only; a full suite run is opt-in, never automatic.
- The review panel is always exactly `primary` (on `DEFAULT_MODEL`) + `simple-reviewer` (on
  haiku), bundled as one dispatch. Critical and Major findings are fixed inline; every Minor is
  deferred, never fixed, with no per-finding judgment call. Both slots re-run on the delta until
  no Critical or Major remains open.
- Only seven guards run: `check-unfinished-work.sh`, `check-base-moved.sh`,
  `check-finish-preflight.sh`, `check-cleanup-complete.sh`, `check-workspace-isolation.sh`,
  `check-worktree-processes.sh`, `check-panel-findings-closed.sh`. Every other
  `skills/flow/scripts/` guard is dropped outright for `/flow-fast`, not hand-run.
- Finish follows `finish-contract-run1.md` and `finish-contract-run2.md` exactly, except run 2
  skips the self-review subagent (step 9) and the verify-cleanup pass (step 7).
- `/flow` drops its `/rename <change-name>` / `/color cyan` printed handoff lines and the
  paragraph explaining them, in `skills/flow/SKILL.md`, `skills/flow-contracts/pipeline.md` and
  `skills/flow-contracts/pipeline-rationale.md`. `/flow-fast` never prints them.
- `plan-class.sh`'s dynamic thresholds move so more changes classify `small`/`regular` and
  compact: small `tasks≤10 files≤25` (was `≤5`/`≤12`), big `tasks≥30 files≥80
  (repos>1∧tasks≥15) (migration∧tasks≥15)` (was `≥15`/`≥40`/`≥8`/`≥8`), compact-roll cutoffs
  `<90` small / `<60` regular+big (was `<70`/`<30`). Experimental and bundle rolls unchanged. This
  affects every `/flow` and `/flow-fast` run's dynamic decision, not only `/flow-fast`'s own.
- `CLAUDE.md`'s skill index gains a `flow-fast` row.
