# Process lessons

**This file is the canonical definition.** Skills reference it by name; none of them restate
the contract.

A process lesson is a procedure or practice a run paid for once — a bisection method, a debug
dance, an environment quirk's workaround — that a later run in any project would otherwise have
to re-derive. It is recorded twice, on purpose: the run's own narrative records the run; the
durable form lives in the lessons home and is what a future run reads. A lesson left only in a
narrative is a lesson the next run re-derives.

## The lessons home

A project's durable briefs live in `<project>/docs/briefs/`, one Markdown file per brief named
`<slug>.md`, opening with a `# <title>` heading. A brief is repo content: it is committed,
reviewed, and lands with the change that promotes it — it survives every cleanup the run's
disposable artifacts do not.

## Resolve, never search

When a ticket, a plan, a review finding, or a conversation names a practice or a brief — "the
flake-bisection brief", "the usual bisect dance" — resolve it through the lessons home, never by
searching repositories or narrative files by hand:

```bash
flow lesson resolve -topic <words>
```

The daemon scans every registered project's `<project>/docs/briefs/` and every archived change's
`narrative.md`, ranks briefs before narrative mentions, and prints the answer as one Markdown
document: each matching brief in full, then each mention's matching lines beside the path that
holds them. The read carries the findings read's contract: an unreachable store is reported on
stderr and exits non-zero, never a partial answer on stdout.

A resolve that finds only narrative mentions — no brief — is the system working, not a gap: the
mention names where the knowledge lives today, and the run that consumes it is the one
positioned to promote it.

## Promote at the close

When a run's own experience yields a durable process lesson, the run writes or updates its
project's `<project>/docs/briefs/<slug>.md` as part of the change, committing it with the
change's own work. The narrative still records the run; the brief is the form the next run
resolves.
