# flow pipeline — rationale

This file is the reasoning behind `skills/flow-contracts/pipeline.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.**

## States

## Stage exit — never the command's own judgment

The tuned threshold, the two prompts, and why their opposite recommendations are not to be
harmonised belong to the command itself — **Convergence** (`skills/flow/brainstorm.md`).

## Command surface

Behaviour a flag used to select is now either asked at invocation (the integration choice in
`/myflow-finish`), derived from the current state, or fixed at the single sensible default
(review-panel breadth, decided by its escalation triggers).

## State transitions

### Every command is re-entrant

### A fix never moves the state

## Wrong state for this command

## Progress visibility

**No third checkbox marker is added to `tasks.md`** to carry an in-progress state. A marker written
at dispatch and resolved at completion would survive a crashed run as a permanently in-progress
task, in a file two guards parse. The in-progress count comes from the harness's task list alone,
which no run persists.

**Stated against the mechanism, never against one harness's tool.** flow runs in Claude Code,
Cursor and Codex, and a rule written against one harness's API is unimplementable in the other two.
Where a harness offers no task-list mechanism, the command prints the equivalent block in its output
instead: a count line naming how many steps are done, in progress and open, followed by one line per
step marked done or not done. The rule is satisfied by whichever mechanism the harness provides, and
no harness has to gain a task tool to satisfy it.

## Stage marks

Two concurrent runs that
happened to carry the same token would resolve to more than one session, which the ambiguity rule
below already refuses — correctly, since a token shared by two sessions identifies neither.

Marking writes: where
the store has no record for the name a `stage begin` carries, the begin handler bootstraps a change
row so the mark has something to attach to, and that row outlives the run — it appears among the
open changes, carries a next command, and is never archived, because no change directory bears that
name. This is the sibling of **Requirement: A state gate reads the state before it marks** (`<agents repo>/openspec/specs/myflow-run-telemetry/spec.md`, frozen at the spectre cutover and kept as the record of where the rule came from):
that rule keeps a command from *reading* a state its own mark authored; this one keeps a command
from *creating* a change nobody named.

The reason is what makes the whole binding mechanism work: the transcript records `tool_use.input.command` — the text handed to the tool — **before** the shell ever
expands it. A substitution therefore lands in every calling session's transcript as the identical,
unexpanded string, and discriminates nothing between them.

## Handoff output

### The block each state renders

- **The next command is the last line** — bare, copy-pasteable, with no prose after it. An agent
  cannot drive a harness's autocomplete; nothing lets a running session prefill the operator's
  input box. The last-line convention plus a four-command surface is the whole mechanism.

- **`/myflow-do` never stages `<project>/spectre/changes/` or `<project>/docs/superpowers/` before
  finish.** The plan was read at `STARTED`; presenting it again as code to review hides the
  implementation diff it is mixed into. Leaving them unstaged, rather than filtering them out of one
  display command, is what makes them absent from *every* view of the staging area — a filtered
  display leaves them in the index, where `git status`, a graphical client and the IDE's
  staged-changes pane show them again. The list is fixed here rather than configured per project; the
  pipeline chooses these paths
  itself, so no project can differ. `/myflow-finish` run 1 stages them and commits them separately
  from the implementation, so nothing is lost.

### The tab commands, printed at the start of a run

They sit at the **start** of the run rather than in the block because labelling a tab is
useful before a long run rather than after it; the rules govern what a command prints when it
*ends*, and these lines are not part of a handoff. The colour is one fixed value for every command
and every change — `cyan`, chosen over `red`, `yellow` and `orange` because those already read as
error and warning states in this pipeline's output — and it signifies only that a pipeline command
owns the tab. `/flow-status` prints neither line: a read-only report does not
own the tab.

**Both of those facts are Claude Code's, and the rule is stated against the mechanism rather than
against them** — for the reason **Progress visibility** (`pipeline.md`) gives, which answers
the identical question for the task list. What every harness can do is
**print two lines of text**, which is why printing is the rule and the measurement is only the
reason invoking is not. So:

Both commands are real, and both routes to calling them are closed. That was established by measurement,
and it is recorded here so the next reader neither repeats the investigation nor treats the printing
as an oversight to correct:

- the harness's `SlashCommand` tool exposes only commands of `type: "prompt"`, while `/rename` and
  `/color` are `type: "local"` or `"local-jsx"` — so the tool route is closed; and
- no writable `/dev/tty` is available to a command — so writing the terminal escape sequence
  directly is closed too.

## Artifact brevity

Stated here rather than in each artifact-writing skill because this file is the one every
`/flow*` command loads before any other step. Four skill-local copies would drift, and whichever
skill lacked one would silently exempt its own artifacts.

## IntelliJ commands

`open` resolves the
app by name (bundle `com.jetbrains.intellij`), returns immediately instead of blocking the shell,
and reuses a running instance.

## Guard resolution

Resolution against the **running command's own** skill directory is what lets a contract loaded
by more than one command — `skills/flow-contracts/finish-contract-run1.md`, loaded by both
`/myflow-finish` and `/flow-status` — name a guard at all: the same basename resolves inside
whichever command is actually running, never a fixed one of them.

Carrying the prefix says which
repository is meant, and removes the need for any classifier to tell describing a guard from
running one.

## Guard presence check

A
guard that resolves a neighbour from its own `$SCRIPT_DIR` at runtime — for example
`check-panel-reproducers.sh` needing `<agents repo>/scripts/reproducer-metachars.sh`, `check-task-commit-fields.sh` needing
`check-task-commit-fields.py`, or `prepare-workspace.sh` needing `check-workspace-isolation.sh` —
fails at the moment it reaches for that neighbour if the neighbour alone is missing, so a missing
sibling is exactly as reportable as a missing guard, and the block names it the same way.

## Finish contract

### Run 1 — the branch is not merged

**The gate precedes the question, and that ordering is the point.** An operator asked how to land a
branch, and only then told it carries unfinished work, has already answered a question about a
branch they believed was complete — and the cheapest of the three courses to take by mistake is the
one that integrates.

**Implementation first is the order, not an accident of it.** The newest commit is the one a forge
shows first, and that should be the code; and the second commit's message is where the outstanding
list from the gate above is written down.

**That script has three outcomes, and they are not interchangeable.** This is where they are
defined; the two call sites point here rather than each describing them.

### Run 2 — the branch is merged

   **The removal goes after the worktree half, and the order is load-bearing.** Worktree cleanup's
   check 5 runs the project's `## stop` command, and it is the **only** place run 2 stops the stack —
   which run 2 nearly always has to do, because `/myflow-do`'s handoff prints the run instructions
   and the operator runs the applications against them, so a stack still up when run 2 starts is the
   common case rather than a rare one. Dropping a database the project's own stack still holds open is a
   removal that fails on the ordinary path, so the removal waits until that stack is down.
   Nothing pulls the other way: these resources live in the project's shared data services rather
   than in the worktree, so taking the worktree down neither removes them nor puts them out of reach,
   and `remove` runs from the main checkout — stated with the command table, in
   **Project configuration** (`skills/flow-contracts/project-configuration.md`) — so no worktree
   removal can destroy the directory it runs from. What has always constrained this step still does:
   the removal stays ahead of step 7, because a verification that runs before the thing it verifies
   can only ever fail.

   **Interleaving the removal into the worktree half was considered and rejected.** Slotting it
   between that half's checks and its destructive commands would put the stack down first just as
   well, but it would bind this step's order to another section's internal check numbering, and it
   would run a once-per-change removal inside a per-worktree loop — two couplings bought for a
   property the placement above already has.

   **The cost of this order is that a worktree half which stops early takes the removal with it.**
   Any failed check leaves every worktree alone —
   **Worktree cleanup** (`skills/flow-contracts/finish-contract-run2.md`) — and the removal behind it
   does not run, so a run blocked by something unrelated to the workspace, an
   uncommitted file in a worktree say, has its database and bucket named as leftovers at step 7 as
   well, with nothing wrong with either. That is the right cost to accept. It lands on a run that has
   already stopped and already needs the operator, it adds lines to a report rather than a failure,
   and run 2 is re-entrant, so the pass after the blocker is cleared does both halves. The check most
   likely to fail is check 5 — a stack that will not stop — and that is exactly the condition under
   which the removal would have failed anyway, which is also why it is not attempted regardless:
   running with the stack down is the guarantee this order buys, and removing after a failed check
   would spend it. The order this replaces put its cost on the ordinary path instead.

   **Both halves share one numbered step deliberately.** They are one act — undoing what this
   change's run created — with an order between them that has to hold, and giving the removal a
   number of its own would renumber steps 7, 8 and 9, which are cited *by number* from
   `skills/flow/archive.md` and from the capability specs frozen under `<agents repo>/openspec/specs/`.

### Worktree cleanup

## State file

## Project configuration

## Jira integration

## Change name resolution (all `/flow*` commands)

Going through the CLI rather
than a skill calling `curl` directly is deliberate, not a style preference: `state list` shares the
same `Client.ListStateBoard` method every other read uses, so it inherits the `Flow-Daemon`
header check, the timeout, and the unreachable/refused classification once, in the one package that
owns them, instead of every contract file that enumerates changes growing its own copy of that HTTP
handling — the outcome `design.md`'s `daemon-owns-db` decision names and rejects.

This is why the second source matters: `spectre list --json` only
sees change directories present in the *current* git checkout, so a change staged in a worktree —
`<project>/spectre/changes/<name>/` created there but never committed to the main checkout — is invisible to
it alone, even while it sits at a human gate with a fully staged diff.
