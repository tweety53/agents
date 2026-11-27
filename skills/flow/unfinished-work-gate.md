# The unfinished-work gate (run 1)

**Loaded by `skills/flow/integrate.md`'s step 1 only when** a worktree reported `OUTSTANDING`. The verdicts and the stage mark are `flow-guard integrate-change`'s, the stop-and-ask rows that step's;
**Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`) is canonical
for the guards.

## Finish it — never a prompt

On `OUTSTANDING` **nothing is asked: the parent finishes the
unfinished work itself and re-runs the guard**, at integrate as at **Write `IN_PROGRESS`**
(`skills/flow/verify-and-handoff.md`). An
unfinished-work verdict is work to do, never a stop and never a question. Three re-runs of one
guard that each print `OUTSTANDING` take **Three failed
rounds** (`skills/flow-contracts/pipeline.md`), applied inline by the parent from every earlier
verdict.

- A missing inline verify is run and recorded — never recorded after
  the fact for work that did not run.
- A verify failure is fixed at its source, held to `~/.claude/rules/fix-determinism-at-the-source.md`
  when it does not fail every run.

**An item no change to the tree can finish** is recorded as a guard false positive per **A verdict the operator calls
structural** below, and the run continues to the
landing route. The outstanding items and their dispositions go into the handoff's `**Decisions:**`
lines.

## A verdict the operator calls structural

When the verdict is structural — the plan held in another worktree with every task ticked and no
open finding — or an item **Finish it** above cannot finish, run once, before the landing question:

```bash
flow record verdict false-positive -change <name> -guard check-unfinished-work \
  -reason "<the operator's reason, verbatim>" -C <canonical-worktree>
```

Once, and at the canonical worktree, whichever worktrees reported `OUTSTANDING`: the guard records
every worktree's verdict under the canonical project (THE STORE CALLS ANCHOR AT THE PLAN'S PROJECT,
`<agents repo>/scripts/flow-guard check-unfinished-work`), and the command flags that project's newest verdict for the
change — a satellite's `-C` names a project holding none, and the store refuses it.

A write that falls back to the journal is one warning line and the run continues.

