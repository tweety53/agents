# The unfinished-work gate (run 1)

**Loaded by `skills/flow/integrate.md`'s step 1 only when** a worktree reported `OUTSTANDING` or
`VISUAL-VERIFY-MISSING`. The verdicts and the stage mark are `integrate-change.sh`'s, the stop-and-ask rows that step's;
**Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`) is canonical
for the guards.

## The prompt

- **`OUTSTANDING:`** → show the breakdown — and the guard's
  `prior false positives for this guard on this project` stderr line when it printed — relay the
  guard's hand-verification procedure per **Hand-verifying a guard verdict**
  (`skills/flow-contracts/pipeline.md`), and offer exactly three courses, shape per Operator
  prompts (`skills/flow-contracts/operator-prompts.md`):

  > **This change carries unfinished work — how should integration proceed?**
  > - **Stop — I'll finish it first** *(recommended)*
  > - **Continue — integrate anyway**
  > - **File or join a Jira follow-up, then continue**

  There is no fourth.

**The `OUTSTANDING:` prompt above is a filing ask** — its third course files exactly these items,
per **Follow-up issues** (`skills/flow-contracts/jira-followups.md`).
**Every filing ask explains before it asks.** Before the filing prompt fires, the message body
explains each item the run would file — what was observed, what breaks because of it, and what the
fix would be — never leaving that explanation to the prompt's option text. The prompt itself follows
the shape **Operator prompts** (`skills/flow-contracts/operator-prompts.md`) defines and records
only the decision. A filed issue is durable and already on the board; an explanation reaching the
operator afterward would describe something they never agreed to.

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

When the operator's answer says the verdict was verified structural — the plan held in another
worktree with every task ticked and no open finding — run, once per worktree that reported
`OUTSTANDING` and before the landing question:

```bash verified:the flag set is design.md §5's; the call shape mirrors the `flow record verdict` invocation in scripts/check-unfinished-work.sh
flow record verdict false-positive -change <name> -guard check-unfinished-work \
  -reason "<the operator's reason, verbatim>" -C <worktree>
```

A write that falls back to the journal is one warning line and the run continues.
