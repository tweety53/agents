# Resume at `STARTED`

Loaded by `skills/flow/SKILL.md` on a run finding `STARTED`, and by **A. Resolve the change and write `STARTED`** (`skills/flow/brainstorm.md`) when the
name resolves to a change already recorded at `STARTED`.

**Load `skills/flow/brainstorm.md`** only when the resume point is not `skills/flow/implement.md`:
a missing worktree runs its steps 1–4, and every other resume point below runs under its **Run
brainstorming and planning through the planner role**. A resume at **B** re-enters the planner loop
at the round after the last `round-<r>.md` under `<abs-worktree>/.superpowers/sdd/planner/` — its
relay first when that round has no `answers-<r>.md`, the resumed round reading `findings.md`, that
latest round file and its answers. Every planner dispatch a resume makes runs on `planner-short`
exactly when `<abs-worktree>/.superpowers/sdd/planner/triage.md` holds `TRIAGE: short`, else on
`planner` (**Judge easy or hard**, `skills/flow/brainstorm.md`); a resume that finds no
`triage.md` and no round file runs the judge first.

### Resuming at `STARTED`

A run finding `"state": "STARTED"` already recorded is resuming a creating run that stopped before
reaching `IN_PROGRESS` — an interrupted session, a context limit, an earlier stop. Skip **A. Resolve the change and write `STARTED`** (`skills/flow/brainstorm.md`)
(the name and the `STARTED` write both already exist) and find where the run left off from what is
on disk, never from a judgement of how finished a file looks. Take the first point whose condition
holds; `<sdd>` is `<abs-worktree>/.superpowers/sdd`:

1. **No worktree** — `git worktree list` names no `<project>-worktrees/<name>`: the run stopped
   between the `STARTED` write and kickoff. Run steps 1–4 of **A. Resolve the change and write
   `STARTED`** (`skills/flow/brainstorm.md`), then take the next point. `<changeRoot>` is
   always `<project>-worktrees/<name>/spectre/changes/<name>/`, never a main-checkout path.
2. **Design not approved** — no `<sdd>/planner/answers-<r>.md` records the operator's explicit
   approve of the convergence confirm: resume at **B**, the planner loop's next round.
3. **Plan not done** — `flow record decisions -change <name>` names no decision: dispatch a final
   planner (step 3 of **Run brainstorming and planning through the planner role**,
   `skills/flow/brainstorm.md`), its prompt naming every artifact already under `<changeRoot>` —
   it keeps what is written and completes what is missing — then run **Plan check** and the gate
   (**Plan check gate**, `skills/flow/brainstorm-planner.md`).
4. **Otherwise** planning is done: go to `skills/flow/implement.md`.

State the resumption point plainly before continuing:
"resuming `<name>` at `<point>`."

### Resuming a recorded decision

A run resumed at `STARTED` rests on **Plan check gate** (`skills/flow/brainstorm-planner.md`): it reads
`flow record decisions -change <name>` and follows the newest row rather than re-deciding.
