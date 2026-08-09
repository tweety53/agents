## Why

The `/myflow-*` pipeline has three human gates: reading the proposal artifact after
`/myflow-start`, reviewing the staged diff and running the apps after `/myflow-do`, and choosing
how to land the branch during `/myflow-finish` run 1. Between stages where nothing actually blocks
progress — after brainstorming converges, after `/myflow-do` hands off cleanly and the operator has
already reviewed it, after a merge+push route completes — the operator still has to notice the
state and type the next command by hand. For an operator who wants to stay hands-off except for the
parts only a human can do (answering brainstorming's design questions, reviewing the diff, testing
the app, choosing the integration route), that adds friction with no compensating benefit.

## What Changes

- Add `/myflow-fast` as a single command that drives the existing three-state pipeline
  (`STARTED` → `IN_PROGRESS` → `FINISHED`) end to end, chaining stages that have no human gate
  between them:
  - First invocation runs full interactive brainstorming (unchanged from `/myflow-start`) then,
    without a separate command, continues straight into SDD+TDD implementation and the review
    panel (unchanged from `/myflow-do`), ending at `IN_PROGRESS` with the same staged-diff +
    run-instructions handoff.
  - Re-invoking at `IN_PROGRESS` with an argument applies it as fix instructions, resuming the
    worktree exactly as re-running `/myflow-do` does today.
  - Re-invoking bare at `IN_PROGRESS` proceeds to the integrate question (open PR / merge+push /
    manual). Merge+push continues in the same invocation through archive/cleanup to `FINISHED`;
    open PR and manual both stop and hand off, since each needs an external action outside this
    tool's control before archiving can happen.
- Records `models: {implementation: sonnet, reviewPanel: sonnet, panelFix: sonnet}` and
  `reviewPanelRoster: light` as `/myflow-fast`'s recommended defaults on the run that creates the
  change (still overridable at the same question `/myflow-start` already asks).
- Skips publishing the proposal HTML artifact — the operator is present for the brainstorming
  dialogue that produces the design, so a separate published summary of the same conversation adds
  no information.

## Capabilities

### New Capabilities

- `myflow-fast-command`: defines `/myflow-fast`'s accepted states, its chaining behavior across
  stage boundaries with no human gate, the fix-vs-integrate disambiguation at `IN_PROGRESS`, and
  its recorded defaults.

### Modified Capabilities

- `myflow-command-surface`: the command surface grows from three pipeline commands plus one
  read-only command to four plus one, and the per-command accepted-states table gains a
  `/myflow-fast` row.

## Impact

- New skill `skills/myflow-fast/SKILL.md`, citing into `myflow-start`'s, `myflow-do`'s and
  `myflow-finish`'s existing stage content and the shared contracts rather than re-deriving them.
- New command files `commands-claude/myflow-fast.md` (+ Cursor/Codex variants), frontmatter
  `model: sonnet`.
- `CLAUDE.md`'s command table and skill index gain a `/myflow-fast` entry.
- `skills/myflow-contracts/pipeline.md`'s command surface section gains a `/myflow-fast` row.
- No change to the state file shape, the three states, or any other command's behavior.
