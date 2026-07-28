# KAN-8 — myflow updates: three states, three commands

**Jira:** KAN-8 "Myflow updates"
**Date:** 2026-07-28
**Repo:** `/Users/tweety53/Projects/agents`

> **Revision note.** This design was first written for a **five-state** model
> (`STARTED → IN_PROGRESS → TEST → REVIEW → FINISHED`, six commands). The operator collapsed it
> further mid-implementation: `/myflow-test` and `/myflow-review` were folded into `/myflow-do` and
> `/myflow-finish`, and `/myflow-fast` was dropped. This document describes **what shipped**. The
> decisions the collapse superseded are retained at the end, marked as such, because the reasoning
> that produced them is still the reasoning that has to be argued against to reverse it.

## Why

myflow's twelve stages encoded two different facts in one field — how far the work got, and who is
waiting on it. That conflation is what forced seven commands whose only job was to say "a human
looked at this", an `originStage` field to remember where a fix was raised, and a monotonic-gates
contract to stop those recorded confirmations from being demoted.

The cost was paid three times over: eighteen commands duplicated across two harness trees, nineteen
skills whose names did not match the commands that loaded them, and a vocabulary guard that had to
be taught every retired spelling because a rename touched so many layers. The operator paid too —
they had to know which of eighteen commands was legal at the current stage.

Collapse the model so the state says how far the work got and the human gate is a property of that
state. Three states, three pipeline commands, one human gate.

Two further observations drove the final shape. **Reviewing a staged diff and running the apps
against a checklist are the same sitting at the keyboard**, so they are one gate, not two stages.
And **integration is not a stage** — it is the first half of finishing, which is why
`/myflow-finish` absorbed it.

## What shipped

### The state machine

```text
/myflow-start  → STARTED      you: read the proposal artifact
/myflow-do     → IN_PROGRESS  you: review the staged diff and run the apps
/myflow-finish → FINISHED     terminal (it integrates on its first run)
```

Each command ends in the state named after it. **Every command is re-entrant**: re-run
`/myflow-start` to revise the proposal (republishing the artifact to the same URL), `/myflow-do` to
fix something. **A fix never moves the state** — `/myflow-do` advances only `STARTED` →
`IN_PROGRESS`, which is what removes `originStage` and the whole fix re-entry table.

`/myflow-do` produces **both** the staged diff **and** `docs/manual-test/<name>.md`, refreshing
both together on a fix run so they cannot drift apart. Skipping the manual test needs no mechanism:
the guide is there to use or ignore, and nothing observes which.

### `/myflow-finish` runs twice

**The branch's merge status alone decides which run happens.** No field records "integration
started" — a field could disagree with git, and a PR a human merged on the forge is then
indistinguishable from one the tool merged, which is correct.

- **Run 1 — not merged.** Asks up front how the branch should land (open a PR *(default)*, merge
  and push, or handle it manually), commits the staged work, takes that route, stops at
  `IN_PROGRESS`. Its handoff names `/myflow-finish` again.
- **Run 2 — merged.** Verifies the merge, syncs delta specs, archives, **commits and pushes the
  archive**, removes the worktrees, writes `FINISHED`.

**No tests, linters, or coverage check run before integration.** Correctness rests on `/myflow-do`
(TDD per task, per-task review, the final panel, and the project's own lint/test commands) and the
human gate.

### Worktree cleanup

Four checks must all pass before anything is removed: no uncommitted tracked changes; **no
untracked files git does not already ignore**; no unpushed commits (with a missing upstream treated
as *unknown*, which fails the check rather than passing it); and the project's local stack stopped
via an optional `## stop` key. Then `git worktree remove --force`, `git branch -d` (never `-D`),
`git worktree prune`. Any failed check leaves every worktree alone.

The untracked-file check is what makes `--force` safe: plain `git worktree remove` refuses on any
untracked file, and `--force` overrides exactly that refusal — so without it, a never-staged `.env`
is destroyed as silently as a `build/` directory.

### State file

`state`, `branch`, `worktrees` (keyed by absolute path), `artifactUrl`, `jiraIssue`, `prUrl`,
`updatedAt`, `updatedBy`. Gone: the whole gates object, the tested flag, the fix-origin field, the
reviewed-tree hash, the fast-path flag, and the separate worktree/merge-base pair.

### Everything else

No command takes a flag. One skill per command, named after it — nineteen skills to seven,
eighteen commands to five. Every review-panel reviewer runs on Sonnet, deleting the provider-family
economy-tier mapping. Handoffs carry only what the operator must act on, with absolute paths and
the next command as a bare last line.

## Decisions

### Collapsing to three states

**ID:** three-state-collapse · **Status:** active
**Chosen:** `STARTED`, `IN_PROGRESS`, `FINISHED`, deleting the separate test and review commands.
**Considered:** the five-state model this document originally specified — rejected because two of
its states existed only to separate work the operator does in one sitting, and a third existed only
to run verification `/myflow-do` had already done.

### Reviewing and testing are one gate

**ID:** one-human-gate · **Status:** active
**Chosen:** `/myflow-do` emits the diff and the guide together.
**Considered:** keeping a separate test command, on the grounds that a guide written before review
may describe behaviour review changes — rejected because a fix re-runs `/myflow-do`, refreshing
both surfaces together.

### Finish absorbs integration

**ID:** finish-absorbs-integration · **Status:** active
**Chosen:** two runs, branching on the branch's merge status.
**Considered:** a separate integration command (reintroduces the state this removes); a recorded
"integration started" field (a second source of truth that can disagree with git).

### No verification gate before integration

**ID:** no-preflight-verification · **Status:** active
**Chosen:** nothing runs before a PR or merge.
**Considered:** keeping the delta-spec coverage check — rejected on the operator's instruction, and
because it ran at the point of least leverage. **Accepted cost, stated plainly:** a change can
merge with a delta-spec scenario no test covers, and nothing will say so.

### One model for the whole review panel

**ID:** panel-model-uniform · **Status:** active
**Chosen:** every slot on Sonnet.
**Considered:** the parent-model/economy-tier split — rejected because it made the panel's cost
depend on which model the operator was running, and needed a provider table to maintain.

### Where a fix leaves the state

**ID:** fix-re-entry · **Status:** active
**Chosen:** the state is untouched except `STARTED` → `IN_PROGRESS`.
**Considered:** always dropping back (a one-line fix would cost a full re-walk); a staleness marker
(reintroduces the bookkeeping field this decision deletes).

### How the next command is surfaced

**ID:** next-command-hint · **Status:** active
**Chosen:** a bare final line in every handoff.
**Considered:** a `Stop` hook — cannot prefill the input box, is Claude-Code-only, and runs on
every stop in every project. Driving autocomplete directly was investigated and is not possible.

### Whether any flag survives

**ID:** flag-removal · **Status:** active
**Chosen:** none. **Considered:** keeping the full-panel override — rejected because "no flags" is
a contract you can hold in your head. **Cost:** panel breadth now rests entirely on escalation
triggers.

### Fate of the `/opsx:*` commands

**ID:** opsx-removal · **Status:** active
**Chosen:** delete the five that duplicate pipeline steps; keep `/opsx:explore`.

### Skill layout

**ID:** skill-layout · **Status:** active
**Chosen:** one skill per command, named after it.
**Considered:** minimal churn (the tree would contradict the docs); a single dispatching skill
(every command would load all stages' instructions).

### Superseded by the three-state collapse

- **`review-role`** — *superseded by `three-state-collapse`.* Kept a separate review command as
  the integration step, with commit + push + PR, on the grounds that moving integration into finish
  would make it one large irreversible step. Reversed when that command was deleted; the concern
  was answered instead by splitting finish into two runs with the merge between them.
- **`test-skip-mechanism`** — *superseded by `one-human-gate`.* Made skipping the manual test a
  matter of not running a separate command. Moot once there is no separate command.
- **`merge-choice-timing`** — *superseded by `finish-absorbs-integration`.* Put the merge question
  at the start of the review command; it now opens `/myflow-finish` run 1 instead.

## Risks / Trade-offs

**A fix does not re-open a gate that already ran** → Accepted (see `fix-re-entry`). The operator
decides whether to re-review or re-test; myflow does not decide for them.

**Nothing checks delta-spec scenario coverage any more** → Accepted and disclosed in `README.md`,
`CLAUDE.md` and `AGENTS.md`, not only here.

**A rename this wide leaves stale references** → `check-references.sh` catches a moved section;
`check-vocabulary.sh` catches retired literals, extended here to include the retired *field* names
after a wholly stale contracts index passed a clean run without them. Neither proves completeness —
a manual sweep is still what makes a rename done.

**`git worktree remove --force` can destroy work** → Four preflight checks, all must pass, and the
untracked-file check is the one that makes `--force` honest.

**An upgrade leaves stale symlinks** → `setup.sh` now prunes destination symlinks whose source no
longer exists, with an assertion in `test-setup.sh` covering both the prune and the guarantee that
a real file the user placed there is never removed.

**Existing state files use the old shape** → No migration is written; self-heal rewrites from
artifacts. Every state file on this machine was already finished when the change was made.

## Verification

```bash
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/test-check-references.sh
scripts/test-setup.sh
SANDBOX="$(mktemp -d)"; HOME="$SANDBOX" ./setup.sh global
```
