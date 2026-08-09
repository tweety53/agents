# Handoff blocks — rationale

This file is the reasoning behind `skills/myflow-contracts/handoff-blocks.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

## Handoff blocks

### The block each state renders

Both commands resolve those lines the same way: **6. Resolve the run instructions**
(`skills/myflow-do/SKILL.md`) is canonical for the procedure, and `/myflow-status`'s detail-view
step cites it rather than restating it.

### Why regeneration beats storage

The block a state hands off is defined **here and nowhere else**, as one template per rendering. Two
commands render it: the command that ends in that state, and `/myflow-status <name>`. **No command
stores the emitted text.** `/myflow-status <name>` regenerates the block from the state file and the
artifacts as they now stand; it never reads back a stored copy, because a stored copy reproduces the
original exactly and then goes wrong silently the moment anything it names moves — a worktree
removed, an artifact republished, a PR opened. Regeneration is the mechanism, not an implementation
detail of it.

### Why the open-questions count is on-disk, not run-only

**Why the open-questions value is not run-only, and carries no marker.** It is derived from an
artifact on disk — the entries under `## Open questions` in the change's design whose status is
still `open` — exactly as the decisions count beside it in the same `Recorded` line is, so
`/myflow-status <name>` regenerates it rather than omitting it. The `Recorded` line sits next to the
`Jira` line and is the opposite case to it: what makes `Jira` run-only is that nothing on disk holds
it, and that test is about where the value lives, not about how close it sits to a line that failed
it. A count that has changed since `/myflow-start` printed it — a revision round answered a question
and moved the entry to `answered by <decision-id>` — is this field working: it reports what is open
now, not what was open then. Both the decisions count and the open-questions count read `none` when
zero, by the missing-rather-than-dropped rule; the fold moved that wording out of the inline
placeholder and into this paragraph, which is a layout change and not a content one. The entry
shape, the immutable ID and the never-delete rule the count reads through are stated once under
**Open questions** (`skills/myflow-start/SKILL.md`) and are not repeated here.

### Why the pre-check must run before the ancestor test

**The first row is a pre-check, not a special case, and it must be answered before the ancestor
test.** `git merge-base --is-ancestor HEAD <base>` returns **true** for a branch carrying no commits
of its own — every commit it has, the base branch already had — and that is the *ordinary*
`IN_PROGRESS` shape, because `/myflow-do` stages without committing. Without the pre-check every
change that has never been through `/myflow-finish` reads as merged and is shown *merged and waiting
on run 2*. The test is `HEAD` against the merge base **already recorded for that worktree** in the
state file's `worktrees` map: equal means the branch has no commits of its own and is therefore not
merged, whatever the ancestor test then says. `scripts/check-finish-preflight.sh` documents this
trap and guards it in exactly that order — its comment (b), on why the recorded-merge-base check
must run before the ancestor test — and the reasoning is not re-derived here.

### Why "recorded but unresolvable" is the dangerous condition

**Why "recorded but unresolvable" is the dangerous condition, not merely a third case.** History
being rewritten, the clone being shallow, or the object having been pruned are the ordinary ways
`git rev-parse --verify` fails to turn a stored sha into a commit — and unlike **absent**, this
condition carries a value, which is exactly what makes it easy to mishandle: compared as a *string*
it is merely "not equal to `HEAD`", which reads as "the branch has commits of its own" and falls
straight through to the bare ancestor test, reporting *merged* for a branch that has never been
through `/myflow-finish`. That is the same refusal-to-infer `check-finish-preflight.sh` makes twice
over — when it is handed `-` for the recorded merge base, and when `rev-parse --verify` on a
recorded one fails — which is why **The block each state renders**
(`skills/myflow-contracts/handoff-blocks.md`) states resolve-then-compare as a rule rather than
leaving it implied by the two conditions alone.

### Why the `Jira` line is run-only

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

### Why `IN_PROGRESS` needs two renderings

**`IN_PROGRESS` has two renderings, and one template could not have served both.** Run 1 ends at
`IN_PROGRESS` but hands off a branch waiting on a merge rather than a diff waiting on review: a
worktree path, run instructions and a staged-diff command are all wrong for it, and it prints none
of them. Forcing both into one template would leave the rule at the top of this section
unsatisfiable rather than merely unsatisfied — no single block is correct for both commands.

### Why `Route` and `Outstanding` are run-only

**Why `Route` and `Outstanding` are run-only.** The landing answer is never remembered between runs,
per **Run 1 — the branch is not merged** (`skills/myflow-contracts/finish-contract.md`), so no
field records which route was taken: a recorded `prUrl` implies the pull-request route, and
nothing separates the other two. The
outstanding list is the unfinished-work gate's verdict at the moment run 1 asked; its durable copy
is the planning commit's message, which is where a later reader looks, and the state file does not
carry it.

### Why `Panel` is run-only

**Why `Panel` is run-only.** It names the roster *that run selected* — which optional slots fired
and which did not — and no field carries it. The only on-disk trace is the panel record
`/myflow-do` writes under `.superpowers/sdd/`, which is gitignored, sits in a worktree run 2
removes, and may legitimately be absent for a change that ran no panel; a value that is sometimes
there and sometimes not is not a source `/myflow-status` can regenerate from, and reporting it
*missing* on every change whose worktree is gone would name a fault where there is none. The
durable copy is the preserved record under `docs/superpowers/reviews/`, which run 1 writes into the
repository — an operator who needs the roster after the fact reads that, not a regenerated block.

### Why `prUrl` never splits the *not merged* row

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

### Why `/myflow-status` cites this file instead of restating the check

**The pre-check paragraph and the recorded-merge-base one are the only statement of that
ordering for a renderer.**
`/myflow-status` performs the check and cites this section for why; it deliberately carries no copy
of the argument, because two copies of one piece of reasoning are two things to keep in step and the
next editor would have no way to tell which was authoritative. Change it here and the consumer
follows.

### Why the table and the block never read merge status differently

**Using the weaker signal where the stronger one is in hand is what made one invocation contradict
itself.** A change stopped at a run-2 cleanup leftover is merged and stays at `IN_PROGRESS`, so the
table reported *branch merged → it will archive* from the ancestor test while the block, keyed on
`prUrl` alone, printed *waiting on the merge* — two answers from one command, one of them false.
The two splits still do not compete: the table splits on merge status to say which `/myflow-finish`
run comes next, this splits on it to say which wait the operator is in, and both end in
`/myflow-finish <name>`.

### Why the `prUrl` test is one-way

**The `prUrl` test is one-way, and the gap is named rather than papered over — it now applies only
to the inconclusive rows.** `prUrl` is `null` until a pull request is opened, and only the
pull-request route ever writes it — see
**State file** (`skills/myflow-contracts/state-file.md`). *Merge and push* and *handle it manually*
both complete run 1 and leave it `null`. So a non-null
`prUrl` proves run 1 happened; a `null` one proves nothing, and where merge status cannot be
determined — no remote, no network, an unresolvable base ref — the report shows the `/myflow-do`
rendering for a branch that may already be integrated.

### Why the imperfect test is accepted rather than replaced

**What a wrong choice costs is bounded, which is why the imperfect test is accepted rather than
replaced.** Both renderings end in the same last line, `/myflow-finish <name>`, so the test can
never send the operator to the wrong command — only show them the wrong fields. And what it shows is
regenerated from the state as it now stands, so a worktree still present is still named and a
removed one reads *missing*.

### Why no field is added to close the gap

**No field is added to close it.**
**Finish contract** (`skills/myflow-contracts/finish-contract.md`) already refuses one:
the branch's merge status is the only source of truth for whether the branch has been integrated,
and a field could disagree with it. That is the same reason merge status governs the table —
the rule was already stated here, and the defect was reading `prUrl` in front of it rather than
behind it. The preflight verdict cannot stand in either — a pushed but unmerged branch returns
`RUN1` both before run 1 and after it, so it does not answer this question.
