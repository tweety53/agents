# myflow pipeline — rationale

This file is the reasoning behind `skills/myflow-contracts/pipeline.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

## States

## Pipeline flow

The diagram and the stage table are the only copy in this repository. Every command summary
elsewhere — the `/myflow-*` tables in `README.md`, `skills/README.md`, `CLAUDE.md` and `AGENTS.md` —
says what each command is *for* and cites the level-1 table for the stages, rather than carrying a
second ordered list of them; no skill carries one either. All four used to, and the copies had
already drifted: not one of their `/myflow-do` rows named the state gate or the fix-documentation
stage, so an agent working from an entry-point file alone would have skipped both. That is the whole
argument for keeping the stages in one place, and it is why a summary elsewhere may name a command's
purpose but never its order. Placing them here is what lets
`/myflow-info` show them: that command reads this file at invocation time and is forbidden from
answering from memory, so a diagram held only in `README.md` is one it can never present.

### Level 1 — the stages of each command

### Level 2 — the stages that hide substructure

Each expansion states the **structure** — the shape that changes only when the pipeline changes —
and cites the file that owns the tuned values. A threshold copied here is a copy that goes wrong
silently, which is the same reason the README carries no diagram; accepting it one level down would
make the rule contradict itself.

#### Brainstorm — `/myflow-start`

#### Writing-plans — `/myflow-start`

#### SDD + TDD per task — `/myflow-do`

#### The review panel — `/myflow-do`

#### The preflight verdict — `/myflow-finish`

#### The unfinished-work gate — `/myflow-finish` run 1

#### The landing routes — `/myflow-finish` run 1

#### Cleanup — `/myflow-finish` run 2

#### Self-review — `/myflow-finish` run 2

## Command surface

Behaviour a flag used to select is now either asked at invocation (the integration choice in
`/myflow-finish`), derived from the current state, or fixed at the single sensible default
(review-panel breadth, decided by its escalation triggers).

## State transitions

### Every command is re-entrant

### A fix never moves the state

## Wrong state for this command

## Git boundaries

**Why `/myflow-do` commits once a PR exists and not before.** A PR is a remote surface, so a
staged-only fix would be invisible on it. Before that the operator reviews a staged diff in the
IDE, and committing would take that away.

## Progress visibility

## Handoff output

### The block each state renders

**Why the `Jira` line is run-only.** It reports the transition *this run made*, and nothing on disk
records one: the state file carries the bare `jiraIssue` key and no transition history at all. Nor
can the value be re-derived by asking the tracker — `/myflow-status` is forbidden from calling Jira,
its own guardrail being that the report is read-only and never transitions or queries an issue — and
even a permitted read would not recover it, because a current status cannot separate `→ In Progress`
from *already In Progress (no transition)* without the status as it stood before the run, which
nothing records. Two of the line's three alternatives are therefore unreproducible, and the third,
*none linked*, is not worth a line that would be wrong for every other change. The key itself is not
lost with it: `/myflow-status` surfaces `jiraIssue` in its table's Jira column and as the first entry
of its detail view, so what the omission drops is the transition, which is the run-only part.

**`IN_PROGRESS` has two renderings, and one template could not have served both.** Run 1 ends at
`IN_PROGRESS` but hands off a branch waiting on a merge rather than a diff waiting on review: a
worktree path, a test-guide path and a staged-diff command are all wrong for it, and it prints none
of them. Forcing both into one template would leave the rule at the top of this section
unsatisfiable rather than merely unsatisfied — no single block is correct for both commands.

**Why `Route` and `Outstanding` are run-only.** The landing answer is never remembered between runs,
per **Run 1 — the branch is not merged** (`skills/myflow-contracts/finish-contract.md`), so no
field records which route was taken: a recorded `prUrl` implies the pull-request route, and
nothing separates the other two. The
outstanding list is the unfinished-work gate's verdict at the moment run 1 asked; its durable copy
is the planning commit's message, which is where a later reader looks, and the state file does not
carry it.

**Why `Panel` is run-only.** It names the roster *that run selected* — which optional slots fired
and which did not — and no field carries it. The only on-disk trace is the panel record
`/myflow-do` writes under `.superpowers/sdd/`, which is gitignored, sits in a worktree run 2
removes, and may legitimately be absent for a change that ran no panel; a value that is sometimes
there and sometimes not is not a source `/myflow-status` can regenerate from, and reporting it
*missing* on every change whose worktree is gone would name a fault where there is none. The
durable copy is the preserved record under `docs/superpowers/reviews/`, which run 1 writes into the
repository — an operator who needs the roster after the fact reads that, not a regenerated block.

**A proven *not merged* is that same pre-check read forward, which is why `prUrl` does not split
it.** Reaching that row means the pre-check resolved the recorded merge base and found `HEAD` past
it — the branch carries commits of its own — and `/myflow-do` puts a commit on a branch only when a
`prUrl` is already recorded. So every route this pipeline has that leaves a commit there has been
through run 1: *handle it manually* commits, pushes and leaves `prUrl` `null`; *merge and push*
lands on the merged row; and a `/myflow-do` fix commits only while a pull request is already
open. Splitting the row on `prUrl` was what rendered a manually landed branch as *Implementation
staged — review and test* — for work that is committed, pushed and already past the human gate —
with the `Git` line's third variant telling the truth one line under a heading that did not. What
the row cannot tell apart is a commit made by hand outside the pipeline, which now renders as
integrated; both renderings end in `/myflow-finish <name>`, so that costs the fields shown and never
the command named.

**The pre-check paragraph and the recorded-merge-base one are the only statement of that
ordering for a renderer.**
`/myflow-status` performs the check and cites this section for why; it deliberately carries no copy
of the argument, because two copies of one piece of reasoning are two things to keep in step and the
next editor would have no way to tell which was authoritative. Change it here and the consumer
follows.

**Using the weaker signal where the stronger one is in hand is what made one invocation contradict
itself.** A change stopped at a run-2 cleanup leftover is merged and stays at `IN_PROGRESS`, so the
table reported *branch merged → it will archive* from the ancestor test while the block, keyed on
`prUrl` alone, printed *waiting on the merge* — two answers from one command, one of them false.
The two splits still do not compete: the table splits on merge status to say which `/myflow-finish`
run comes next, this splits on it to say which wait the operator is in, and both end in
`/myflow-finish <name>`.

**The `prUrl` test is one-way, and the gap is named rather than papered over — it now applies only
to the inconclusive rows.** `prUrl` is `null` until a pull request is opened, and only the
pull-request route ever writes it — see
**State file** (`skills/myflow-contracts/state-file.md`). *Merge and push* and *handle it manually*
both complete run 1 and leave it `null`, and self-heal may clear one that was real. So a non-null
`prUrl` proves run 1 happened; a `null` one proves nothing, and where merge status cannot be
determined — no remote, no network, an unresolvable base ref — the report shows the `/myflow-do`
rendering for a branch that may already be integrated.

**What a wrong choice costs is bounded, which is why the imperfect test is accepted rather than
replaced.** Both renderings end in the same last line, `/myflow-finish <name>`, so the test can
never send the operator to the wrong command — only show them the wrong fields. And what it shows is
regenerated from the state as it now stands, so a worktree still present is still named and a
removed one reads *missing*.

**No field is added to close it.**
**Finish contract** (`skills/myflow-contracts/finish-contract.md`) already refuses one:
the branch's merge status is the only source of truth for whether the branch has been integrated,
and a field could disagree with it. That is the same reason merge status governs the table —
the rule was already stated here, and the defect was reading `prUrl` in front of it rather than
behind it. The preflight verdict cannot stand in either — a pushed but unmerged branch returns
`RUN1` both before run 1 and after it, so it does not answer this question.

### The tab commands, printed at the start of a run

**Both of those facts are Claude Code's, and the rule is stated against the mechanism rather than
against them** — for the reason **Progress visibility** (`pipeline.md`) gives, which answers
the identical question for the task list and is not restated here. What every harness can do is
**print two lines of text**, which is why printing is the rule and the measurement is only the
reason invoking is not. So:

## IntelliJ commands

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
   which run 2 nearly always has to do, because `/myflow-do` hands off a manual test guide and the
   operator runs the applications against it, so a stack still up when run 2 starts is the common
   case rather than a rare one. Dropping a database the project's own stack still holds open is a
   removal that fails on the ordinary path, so the removal waits until that stack is down.
   Nothing pulls the other way: these resources live in the project's shared data services rather
   than in the worktree, so taking the worktree down neither removes them nor puts them out of reach,
   and `remove` runs from the main checkout — stated with the command table, in
   **Project configuration** (`skills/myflow-contracts/project-configuration.md`) — so no worktree
   removal can destroy the directory it runs from. What has always constrained this step still does:
   the removal stays ahead of step 6, because a verification that runs before the thing it verifies
   can only ever fail.

   **Interleaving the removal into the worktree half was considered and rejected.** Slotting it
   between that half's checks and its destructive commands would put the stack down first just as
   well, but it would bind this step's order to another section's internal check numbering, and it
   would run a once-per-change removal inside a per-worktree loop — two couplings bought for a
   property the placement above already has.

   **The cost of this order is that a worktree half which stops early takes the removal with it.**
   Any failed check leaves every worktree alone —
   **Worktree cleanup** (`skills/myflow-contracts/finish-contract.md`) — and the removal behind it
   does not run, so a run blocked by something unrelated to the workspace, an
   uncommitted file in a worktree say, has its database and bucket named as leftovers at step 6 as
   well, with nothing wrong with either. That is the right cost to accept. It lands on a run that has
   already stopped and already needs the operator, it adds lines to a report rather than a failure,
   and run 2 is re-entrant, so the pass after the blocker is cleared does both halves. The check most
   likely to fail is check 5 — a stack that will not stop — and that is exactly the condition under
   which the removal would have failed anyway, which is also why it is not attempted regardless:
   running with the stack down is the guarantee this order buys, and removing after a failed check
   would spend it. The order this replaces put its cost on the ordinary path instead.

   **Both halves share one numbered step deliberately.** They are one act — undoing what this
   change's run created — with an order between them that has to hold, and giving the removal a
   number of its own would renumber steps 6, 7 and 8, which are cited *by number* from
   `openspec/specs/` and from `skills/myflow-finish/SKILL.md`.

### Worktree cleanup

## Temporary artifacts registry

**What that leaves behind is bounded, and where it stops being acceptable is named rather than
glossed.** What stays in the index is sessions and cache entries, which are disposable by
construction — the accepted cost under
**The cache index** (`skills/myflow-contracts/workspace-isolation.md`) is precisely that nothing
which must survive a restart may be kept there, so leaving them costs a login and never data. The
cost that is *not* free is slot exhaustion: a cache offers sixteen indices, one of which is the
empty-id default, so a probe that reads a non-empty index as taken finds fewer free slots as
finished changes accumulate, and a workspace that can claim none falls back to sharing — the failure
this contract exists to remove. The remedy is the operator flushing the cache, and it is safe for
exactly the reason the leftovers are: nothing durable is in there. A project may ship its own
command to list or flush its stale indices; that is the project's tooling, and this row does not
claim it — the `Removed by` cell stays `nothing in this pipeline` either way.

**Which rows run 2 verifies is read off this table, not listed again.** Every row whose lifetime
ends at run 2 is checked back by `scripts/check-cleanup-complete.sh`, whose header explains which
rows that leaves it reading and why; step 6 of
**Run 2 — the branch is merged** (`skills/myflow-contracts/finish-contract.md`) is where its
verdict is acted on.

**That derivation is declared, not left implicit.** The guard carries one marker line per row of
this table saying whether it checks that row or deliberately does not, with the reason; its harness
reads this table and those markers and fails when the two disagree in either direction. So a row
added here goes nowhere until someone records a decision about it — which is what stops a future
artifact from being confirmed clean by a guard that never looked for it.

## State file

## State self-heal

## Project configuration

## Jira integration

## Model policy

That citation is a **checked** one, not a courtesy: the guard associates a bold token with the path
beside it and matches it against the target's headings, and an OpenSpec `### Requirement: …` heading
is a heading like any other, so naming the requirement in full is what makes
`scripts/check-references.sh` fire when it moves. A bare backticked path with no bold token beside
it is **not** checked and rots silently — which is what this bullet's predecessor did.

The two rules point in opposite directions on purpose. A reviewer's job is to be many independent
readings of a finished diff, so its cost must not scale with the operator's session model. An
implementer's job is to get the diff right the first time, where capability compounds.

## Change name resolution (all `/myflow-*` commands)
