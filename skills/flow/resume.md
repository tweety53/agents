# Resume at `STARTED`

Loaded by `skills/flow/SKILL.md` on a run finding `STARTED`, and by **A. Resolve the change and write `STARTED`** (`skills/flow/brainstorm.md`) when the
name resolves to a change already recorded at `STARTED`.

**Load `skills/flow/brainstorm.md`** only when the resume point is not `skills/flow/implement.md`:
a missing worktree runs its steps 1–5, and a resume at **B** or **D** runs under its **Run
brainstorming and planning directly**.

**Load `skills/flow/withdrawal.md`** only when `total == 0` below.

### Resuming at `STARTED`

A run finding `"state": "STARTED"` already recorded is resuming a creating run that stopped before
reaching `IN_PROGRESS` — an interrupted session, a context limit, an earlier stop. Skip **A. Resolve the change and write `STARTED`** (`skills/flow/brainstorm.md`)
(the name and the `STARTED` write both already exist) and determine where the run actually left off
by reading, not by assuming:

- **Does the worktree exist** — `git worktree list` naming `<project>/.worktrees/<name>`. It is
  created inside `flow.kickoff` (steps 1–5 of **A. Resolve the change and write `STARTED`**, `skills/flow/brainstorm.md`), so a missing one means the run stopped between
  the `STARTED` write and that step: run steps 1–5 now, then continue below. `<changeRoot>` is
  always `<project>/.worktrees/<name>/spectre/changes/<name>/`, never a main-checkout path.
- `spectre list --json`'s entry for this change's `done`/`total`, run in the worktree —
  `total == 0` means no plan exists yet: the run first opens with the withdrawal route's resume
  ask (**The withdrawal route**, `skills/flow/withdrawal.md`) — resume brainstorming, the default, or withdraw — and a
  resume answer continues at **B** in
  `skills/flow/brainstorm-planner.md`. <!-- refs-guard:allow -->
- The change root's own `tasks.md` — a scaffold with no enriched steps means writing-plans has not
  run: resume at **D** in `skills/flow/brainstorm-planner.md`. A plan meeting writing-plans <!-- refs-guard:allow -->
  quality (exact paths, verification commands, no placeholders) means planning is fully done: skip
  straight to `skills/flow/implement.md`.

State the resumption point plainly before continuing:
"resuming `<name>` at `<point>`."

### Resuming a recorded decision

A run resumed at `STARTED` rests on **Decide** (`skills/flow/brainstorm-planner.md`): it reads
`flow record decisions -change <name>` and follows the newest row rather than re-rolling, or
re-runs the Decide step alone when none exists yet.
