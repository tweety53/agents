# Progress visibility

Loaded by the load directive of **Progress visibility** (`skills/flow-contracts/pipeline.md`).

**`/flow` registers nothing with the harness's task-list mechanism.** Its live progress view on
Claude Code is the `flow-task-list` mod (`mods/flow-task-list/`): a band above the prompt with one
row per subagent, drawn from each dispatch's description, and a hint-line tail naming the tracker
key where the change name leads with one and the phase — read from the run's own `flow stage` marks
(**Stage marks**, below) — and the plan's done tasks over its total (`19/25`), read from the change's `tasks.md` and its
worktree's git log on every draw, followed by ` │ ` and a tally of those rows by emoji. A task counts
as done when it is ticked or a commit carrying its `Task-Id:` trailer is on the change branch since its
base — `branch.<current>.flowBase` when recorded (its `refs/remotes/origin/` ref first), otherwise
`refs/remotes/origin/HEAD`, as `<agents repo>/stats/internal/guard/taskrecords.go` resolves it — so a gated task counts before its deferred tick.
Only the change's worktree in the session's own repository is read: a task whose commit lands in a peer
repository's worktree counts once it is ticked. Work no subagent carries — brainstorming, an inline-executed task, a finish step —
shows on the main agent's own `main` row, always first on the band, as `main: <label>`, `in progress` while the
main turn runs and `waiting` between turns while a subagent still runs, and never numbered as a task.
Its label is the top-level unit the main agent is on, never a subtask or a Bash command: the latest
open unit of its own status lines (`rules/be-brief.mdc`) naming a task of the running change's plan,
else the stage key the run's latest `flow stage begin` named, else the latest open unit of its own
status lines, else none. A status-line unit is the main agent's own when no subagent row shares its key — its
name before any `(`, `:` or ` —`, a trailing ` review` dropped — or names one of its task numbers; a
`done` line closes its unit, and a `blocked` one lasts only until the next main turn starts. A status
line is never a row of its own. Nested under the `main` row, the band lists each running subagent row with its elapsed
time and a state word read from its description: `re-review` for a review past its first round
(`review-2`, `panel-2-…`), one re-run after a fix (`visual-verify-fix-1`, `task-3+4-reviewer-fix-1`) or one naming a
re-review, `review` for any other review, `fix` for a fix, `in progress` for any other agent. A row with a `Task <x>/<n>`
prefix is a `review` or `re-review` only when its words are exactly that — `Tasks 3+4+7/<n> (review)` — and `in
progress` otherwise, whatever its title names. A subagent that failed is drawn `blocked` among the running rows, in
spawn order, until the next main turn starts. Below those it lists each task of the running change's plan that is neither ticked nor landed and named by
neither a running row's nor the `main` row's `Task <x>/<n>` prefix as a `pending` row, read from the change's
`tasks.md` in its worktree on every draw — never stored. Only when no task is pending, the last two done subagent
rows dispatched under the change the run's latest `flow stage begin` named — with no change marked, the last two
done since the current main turn began — are drawn above the running rows, in the order they finished, in the room
the running rows leave; a running row is never left out for a done one. The band draws
at most five rows besides the `main` row.

**Every subagent dispatch's description is the board row's label**, in the shape
**Dispatch sites — the parent's closed list** (`skills/flow/implement.md`) states — no emoji and
no state of its own: the mod adds both, and numbers a task's row from its `Task <x>/<n>` prefix. A row
is numbered only when every task its prefix names is a task of the running change's plan and `<n>` is
that plan's size, as `Task <x>/<n>` — a group of tasks 3, 4 and 7 as `Task 3+4+7/<n>` while it runs; every
other row — an in-run pipeline fix, a panel or fix-round row — is shown unnumbered, and no row is numbered by its
dispatch order. Once done, a row with a `Task <x>/<n>` prefix shows its highest task alone over its own `<n>` —
that group as `Task 7/<n>` — whether or not a plan is running.
Claude Code's own agent panel, not the board, shows a running agent's latest tool-call
description, so every dispatch prompt also tells the agent to open each tool call's description
with its dispatch description's unit and its round (`Task 30/31 review-1 — run the drawer tests`,
`Visual verify review-2 — resolve the specs`).

**The progress view is a view, never a record.** No command, guard or contract reads it back as
evidence of what was done. `tasks.md` remains the single source of truth for a
plan's completion state, and `<agents repo>/scripts/flow-guard check-unfinished-work` reads that file.

**No third checkbox marker is added to `tasks.md`** to carry an in-progress state; the in-progress
count comes from no persisted record at all — the board's rows are session state, never written
to disk.

**On a harness without the mod, `/flow` prints the equivalent block instead** — a count
line naming how many steps are done, in progress and open, followed by one line per step marked
done or not done. One step per whichever cited stage is running at the time, at that stage's own
granularity — brainstorming checklist items and artifacts on the creating/resuming branch, tasks on
the implementation branch, a finish run's steps on the integrate/cleanup branch.
