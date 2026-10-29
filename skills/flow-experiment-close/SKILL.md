---
name: flow-experiment-close
description: Analyse an open allowed-pairs experiment from its stored before/after stats, close it with a verdict and summary, and on a positive verdict start the /flow change that makes its winning cells permanent. Standalone, not a pipeline stage. Use for /flow-experiment-close.
allowed-tools: Bash(flow:*)
license: MIT
compatibility: Requires the flow CLI and a reachable flowd.
---

# Close an allowed-pairs experiment

An experiment overrides whole cells of the **Allowed pairs** table
(`skills/flow-contracts/model-policy.md`) while it is open (KAN-1067). This skill judges one from
the figures flowd stored, closes it, and — on a positive verdict — starts the change that writes
its winning cells into the table, which stays the permanent source (decision `close-skill`).

## 1. Pick the experiment

The argument is the experiment's id, `/flow-experiment-close <id>`. With none, run
`flow experiment list -open`: one open experiment → that one; several → ask with
AskUserQuestion, one option per experiment (its id, name and cells), and silence closes nothing:
the skill stops and reports the open experiments; none → report that no
experiment is open and stop.

## 2. Read its figures

Run `flow experiment stats <id>`. It prints JSON: `runs` (`before`, `after` — changes counted in
each window), `excluded` (after-window changes a `main` cell left out because their harvested main
thread ran another pair or no run of theirs is harvested yet), and `metrics`, each with `before` and `after` means (`null` where no
counted change carries one) and a weekly `series`.

**With fewer than 5 changes in `after`, ask before closing** — options **Wait** (recommended: the
after figures are too few to judge) and **Close now**. On **Wait**, stop without closing.

## 3. Judge it

Lower is better for every metric. Compare each metric's `after` against its `before`, and state the
verdict with the metrics that decided it:

- `positive` — the decisive metrics improved and none worsened materially;
- `negative` — a decisive metric worsened;
- `inconclusive` — the figures move both ways, or too little to call.

**For a combined experiment** (more than one cell), say which cells the evidence supports, and
state that a combined result cannot attribute an effect to one cell.

## 4. Close it

Write the summary — at most 8 lines: the verdict, the runs before/after, each decisive metric
`before → after`, and the cells recommended — and run:

```bash
flow experiment close <id> -verdict <positive|negative|inconclusive> -summary "$(cat <<'SUMMARY'
<summary>
SUMMARY
)"
```

The quoted heredoc passes the summary verbatim, apostrophes included.

## 5. On a positive verdict, make the winners permanent

Ask which cells to make permanent — the options are the supported cells, **all supported cells**
recommended. Then invoke `/flow` with the change description:

> Make experiment <id>'s winning cells permanent in the Allowed pairs table: <role>=<pairs>; …

On `negative` or `inconclusive` the table is left as it is; report the closed experiment and stop.
