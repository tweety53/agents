# The unfinished-work gate (run 1)

**Loaded by `skills/flow/integrate.md`'s step 1 only when** a worktree reported `OUTSTANDING` or
`UNTASKED`. `VISUAL-VERIFY-MISSING` never loads it: that verdict is `STOP: visual-verify`, which no
course here passes. The verdicts and the stage mark are `integrate-change.sh`'s, the stop-and-ask rows that step's;
**Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`) is canonical
for the guards.

## Finish it — never a prompt

On `OUTSTANDING` or `UNTASKED` — from either guard — **nothing is asked: the parent finishes the
unfinished work itself and re-runs the guard**, at integrate as at **Write `IN_PROGRESS`**
(`skills/flow/verify-and-handoff.md`). Recorded at the operator's instruction (2026-10-22): an
unfinished-work verdict is work to do, never a stop and never a question.

- A missing dispatch, review, pass row or inline verify is run and recorded — never recorded after
  the fact for work that did not run.
- A verify failure is fixed at its source, held to `~/.claude/rules/fix-determinism-at-the-source.md`
  when it does not fail every run.
- An untasked commit is covered as **Commits outside every task**
  (`skills/flow-contracts/git-boundaries.md`) states.

**An item no change to the tree can finish** — a verify failure that does not recur in a measured
re-run, its run count and failure rate stated; a verdict the guard reaches over a report line that
names a passing spec — is recorded as a guard false positive per **A verdict the operator calls
structural** below, its `-reason` carrying that measurement, and the run continues to the landing
route. The outstanding items and their dispositions go into the message of the commit that carries
the planning artifacts and into the handoff's `**Decisions:**` lines. The three courses below are
never offered.

## The three courses

On `OUTSTANDING` — from either guard — the operator is offered **exactly three** courses:

| Course | What run 1 then does |
|--------|----------------------|
| **Stop — I'll finish it first** *(recommended)* | stop, leaving the change at `IN_PROGRESS` with nothing staged, committed or pushed |
| **Continue — integrate anyway** | proceed to the landing question, carrying the outstanding list into the planning commit's message and the handoff — and, where the operator called the verdict structural, records it as a guard false positive per **A verdict the operator calls structural** (`skills/flow/unfinished-work-gate.md`) |
| **File or join a Jira follow-up, then continue** | put the outstanding items on a follow-up issue — joining an open one where the operator confirms a candidate, otherwise filing a new one — then proceed |

The breakdown is relayed with the guard's own hand-verification procedure, so the operator can
verify before choosing, per **Hand-verifying a guard verdict** (`skills/flow-contracts/pipeline.md`).

**Stop is marked as the recommendation, and the reason is stated rather than left to be inferred.**

There is no fourth course, and in particular none that hands back to `/flow`'s implement phase inline. The filed
issue is labelled and linked per
**Labels on issues the pipeline creates** (`skills/flow-contracts/jira-integration-finish.md`).

**File or join a Jira follow-up** puts the outstanding items on a follow-up issue and continues.
See **Follow-up issues** (`skills/flow-contracts/jira-followups.md`) for the search, the
confirmation, and how it is labelled; a filing that fails is one skipped-with-reason line and the
run still continues.

**A filing that fails is one skipped-with-reason line, and the run still proceeds** — the same
degradation every other Jira write in this pipeline has, per
**Never blocking** (`skills/flow-contracts/jira-integration.md`). Creation can fail for the usual
reasons (auth, permission, an unknown project key or label) and there may be no tracker configured
at all. None of them changes the operator's answer, which was *continue*: the outstanding list still
reaches the planning commit's message and the handoff, which is where this change requires the
durable record to be. A failed filing is never silently upgraded to **Stop**, and never passes
unmentioned.

**Three more outcomes of that course behave the same way**, and all three belong to
**Follow-up issues** (`skills/flow-contracts/jira-followups.md`) rather than here: the search
that finds a candidate asks the operator to confirm the join before writing to it, a declined
confirmation files a new follow-up instead, and a search that *fails* files nothing and says so. Each
is one line and none of them stops the run or changes the answer already given. That file is
canonical for all of it — including the ordering of a join's three writes and what a partial one
reports.

**What the operator integrated over is recorded where a transcript is not**: the outstanding list
goes into the message of the commit that carries the planning artifacts, and into run 1's handoff.
The signals that produce that list are the script's own.

## A verdict the operator calls structural

When the verdict is structural — the plan held in another worktree with every task ticked and no
open finding — or an item **Finish it** above cannot finish, run once, before the landing question:

```bash verified:the flag set is design.md §5's; the call shape mirrors the `flow record verdict` invocation in scripts/check-unfinished-work.sh
flow record verdict false-positive -change <name> -guard check-unfinished-work \
  -reason "<the operator's reason, verbatim>" -C <canonical-worktree>
```

Once, and at the canonical worktree, whichever worktrees reported `OUTSTANDING`: the guard records
every worktree's verdict under the canonical project (THE STORE CALLS ANCHOR AT THE PLAN'S PROJECT,
`<agents repo>/scripts/check-unfinished-work.sh`), and the command flags that project's newest verdict for the
change — a satellite's `-C` names a project holding none, and the store refuses it (kan-941:
"dispatch not found" for both satellites).

A write that falls back to the journal is one warning line and the run continues.

**An `UNTASKED` listing is never flagged this way.** `check-untasked-commits.sh` records no
verdict, so the flag would land on `check-unfinished-work`'s newest row, a `CLEAR` (kan-1003:
verdict 646). A listed commit's author is the identity every session commits under, so the relay
never calls it the operator's own.
