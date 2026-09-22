# Handoff blocks — rationale

This file is the reasoning behind `skills/flow-contracts/handoff-blocks.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.**

## Handoff blocks

### The block each state renders

### Why regeneration beats storage

`/flow-status <name>` regenerates the block from the state file and the
artifacts as they now stand; it never reads back a stored copy, because a stored copy reproduces the
original exactly and then goes wrong silently the moment anything it names moves — a worktree
removed, an artifact republished, a PR opened. Regeneration is the mechanism, not an implementation
detail of it.

### Why the open-questions count is on-disk, not run-only

It is derived from an
artifact on disk — the entries under `## Open questions` in the change's design whose status is
still `open` — exactly as the decisions count beside it in the same `Recorded` line is, so
`/flow-status <name>` regenerates it rather than omitting it. The `Recorded` line sits next to the
`Jira` line and is the opposite case to it: what makes `Jira` run-only is that nothing on disk holds
it, and that test is about where the value lives, not about how close it sits to a line that failed
it. A count that has changed since `/flow`'s creating run printed it — a revision round answered a question
and moved the entry to `answered by <decision-id>` — is this field working: it reports what is open
now, not what was open then. Both the decisions count and the open-questions count read `none` when
zero, by the missing-rather-than-dropped rule; the fold moved that wording out of the inline
placeholder and into this paragraph, which is a layout change and not a content one.

### Why the pre-check must run before the ancestor test

`git merge-base --is-ancestor HEAD <base>` returns **true** for a branch carrying no commits
of its own — every commit it has, the base branch already had — and that is the *ordinary*
`IN_PROGRESS` shape, because `/flow`'s implement phase stages without committing. Without the pre-check every
change that has never been through bare `/flow` reads as merged and is shown *merged and waiting
on run 2*. The test is `HEAD` against the merge base **already recorded for that worktree** in the
state file's `worktrees` map: equal means the branch has no commits of its own and is therefore not
merged, whatever the ancestor test then says. `<agents repo>/scripts/check-finish-preflight.sh` documents this
trap and guards it in exactly that order — its comment (b), on why the recorded-merge-base check
must run before the ancestor test.

### Why "recorded but unresolvable" is the dangerous condition

History
being rewritten, the clone being shallow, or the object having been pruned are the ordinary ways
`git rev-parse --verify` fails to turn a stored sha into a commit — and unlike **absent**, this
condition carries a value, which is exactly what makes it easy to mishandle: compared as a *string*
it is merely "not equal to `HEAD`", which reads as "the branch has commits of its own" and falls
straight through to the bare ancestor test, reporting *merged* for a branch that has never been
through bare `/flow`. That is the same refusal-to-infer `check-finish-preflight.sh` makes twice
over — when it is handed `-` for the recorded merge base, and when `rev-parse --verify` on a
recorded one fails — which is why **The block each state renders**
(`skills/flow-contracts/handoff-blocks.md`) states resolve-then-compare as a rule rather than
leaving it implied by the two conditions alone.

### Why the `Jira` line is run-only

It reports the transition *this run made*, and nothing on disk
records one: the state file carries the bare `jiraIssue` key and no transition history at all. Nor
can the value be re-derived by asking the tracker — `/flow-status` is forbidden from calling Jira,
its own guardrail being that the report is read-only and never transitions or queries an issue — and
even a permitted read would not recover it, because a current status cannot separate `→ In Progress`
from *already In Progress (no transition)* without the status as it stood before the run, which
nothing records. Two of the line's three alternatives are therefore unreproducible, and the third,
*none linked*, is not worth a line that would be wrong for every other change. The key itself is not
lost with it: `/flow-status` surfaces `jiraIssue` in its table's Jira column and as the first entry
of its detail view, so what the omission drops is the transition, which is the run-only part.

### Why `IN_PROGRESS` needs two renderings

Run 1 ends at
`IN_PROGRESS` but hands off a branch waiting on a merge rather than a diff waiting on review: a
worktree path, run instructions and a staged-diff command are all wrong for it, and it prints none
of them. Forcing both into one template would leave the rule at the top of this section
unsatisfiable rather than merely unsatisfied — no single block is correct for both commands.

### Why `Route` and `Outstanding` are run-only

The landing answer is never remembered between runs,
per **Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`), so no
field records which route was taken: a recorded `prUrl` implies the pull-request route, and
nothing separates the other two. The
outstanding list is the unfinished-work gate's verdict at the moment run 1 asked; its durable copy
is the planning commit's message, which is where a later reader looks, and the state file does not
carry it.

### Why `Panel` is run-only

It names the roster *that run selected* — which optional slots fired
and which did not — and no field carries it. The only on-disk trace is the panel record
`/flow`'s implement phase writes under `<abs-worktree>/.superpowers/sdd/`, which is gitignored, sits in a worktree run 2
removes, and may legitimately be absent for a change that ran no panel; a value that is sometimes
there and sometimes not is not a source `/flow-status` can regenerate from, and reporting it
*missing* on every change whose worktree is gone would name a fault where there is none. The
durable copy is the preserved record under `<project>/docs/superpowers/reviews/`, which run 1 writes into the
repository — an operator who needs the roster after the fact reads that, not a regenerated block.

### Why `prUrl` never splits the *not merged* row

**A proven *not merged* is that same pre-check read forward, which is why `prUrl` does not split
it.** Reaching that row means the pre-check resolved the recorded merge base and found `HEAD` past
it — the branch carries commits of its own — and `/flow`'s implement phase puts a commit on a branch only when a
`prUrl` is already recorded. So every route this pipeline has that leaves a commit there has been
through run 1: *handle it manually* commits, pushes and leaves `prUrl` `null`; *merge and push*
lands on the merged row; and a `/flow`'s implement phase fix commits only while a pull request is already
open. Splitting the row on `prUrl` was what rendered a manually landed branch as *Implementation
staged — review and test* — for work that is committed, pushed and already past the human gate —
with the `Git` line's third variant telling the truth one line under a heading that did not. What
the row cannot tell apart is a commit made by hand outside the pipeline, which now renders as
integrated; both renderings end in `/flow <name>`, so that costs the fields shown and never
the command named.

### Why `/flow-status` cites this file instead of restating the check

**The pre-check paragraph and the recorded-merge-base one are the only statement of that
ordering for a renderer.**
`/flow-status` performs the check and cites this section for why; it deliberately carries no copy
of the argument, because two copies of one piece of reasoning are two things to keep in step and the
next editor would have no way to tell which was authoritative. Change it here and the consumer
follows.

### Why the table and the block never read merge status differently

**Using the weaker signal where the stronger one is in hand is what made one invocation contradict
itself.** A change stopped at a run-2 cleanup leftover is merged and stays at `IN_PROGRESS`, so the
table reported *branch merged → it will archive* from the ancestor test while the block, keyed on
`prUrl` alone, printed *waiting on the merge* — two answers from one command, one of them false.
The two splits still do not compete: the table splits on merge status to say which bare `/flow`
run comes next, this splits on it to say which wait the operator is in, and both end in
`/flow <name>`.

### Why the `prUrl` test is one-way

**The `prUrl` test is one-way, and the gap is named rather than papered over — it now applies only
to the inconclusive rows.** `prUrl` is `null` until a pull request is opened, and only the
pull-request route ever writes it — see
**State file** (`skills/flow-contracts/state-file.md`). *Merge and push* and *handle it manually*
both complete run 1 and leave it `null`. So a non-null
`prUrl` proves run 1 happened; a `null` one proves nothing, and where merge status cannot be
determined — no remote, no network, an unresolvable base ref — the report shows the `/flow`'s implement phase
rendering for a branch that may already be integrated.

### Why the imperfect test is accepted rather than replaced

**What a wrong choice costs is bounded, which is why the imperfect test is accepted rather than
replaced.** Both renderings end in the same last line, `/flow <name>`, so the test can
never send the operator to the wrong command — only show them the wrong fields. And what it shows is
regenerated from the state as it now stands, so a worktree still present is still named and a
removed one reads *missing*.

### Why no field is added to close the gap

**No field is added to close it.**
**Finish contract** (`skills/flow-contracts/finish-contract-run1.md`) already refuses one:
the branch's merge status is the only source of truth for whether the branch has been integrated,
and a field could disagree with it. That is the same reason merge status governs the table —
the rule was already stated here, and the defect was reading `prUrl` in front of it rather than
behind it. The preflight verdict cannot stand in either — a pushed but unmerged branch returns
`RUN1` both before run 1 and after it, so it does not answer this question.

### Editor-facing passages moved from the contract

Moved verbatim from **The template is the definition, and it carries what the commands print.**,
where it closed the sentence on each template carrying every field its command emits: , and a
field added to a command is added here in the same change.

Moved verbatim, the paragraph that followed it:

**"Here and nowhere else" is a duty on the producing skills, not a claim about them.** Each of the
three phase files — `skills/flow/verify-and-handoff.md`, `skills/flow/integrate.md` and
`skills/flow/archive.md` — carries the block it prints, and each **cites this section as the
definition** at that block. A block sitting in a skill with no citation is a second, independently
authored definition however faithfully it happens to match today, and it is exactly how the two
copies drift: nothing tells the next editor of the skill that this file exists. The citation is what
turns three copies into one definition and three renderings of it.

Moved verbatim from the `FINISHED` paragraph, describing the terminal block `skills/flow/archive.md`
prints: That block also carries `**Self-review:** <path> (rating: <n>/5) | deferred —
docs/self-review/<name>-context.md | skipped | skipped — project default`, immediately after
`**Cleanup:** verified`, naming step 9's outcome, and `**Guards:** all present | N missing — those
checks were performed by hand`, immediately after `Self-review`, naming what that run's own
start-of-run guard presence check found — both values only run 2 ever has, exactly like the fields
beside them.

Moved verbatim from the same paragraph, after the interrupted-run report's description: That
interrupted-run report carries neither the `Self-review` nor the `Guards` field: it is printed
only when run 2 stops **before** step 8, so step 9 never runs there and there is nothing for
`Self-review` to name, and its own text prints no `Guards` line either — adding either field
regardless would misstate a run that never reached self-review and a report that does not carry it.
