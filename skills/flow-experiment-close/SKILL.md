---
name: flow-experiment-close
description: Analyse an open allowed-pairs experiment from its stored before/after stats, close it with a verdict and summary, and on a positive verdict make its cells permanent in flowd's role pairs. Standalone, not a pipeline stage. Use for /flow-experiment-close.
allowed-tools: Bash(flow:*)
license: MIT
compatibility: Requires the flow CLI and a reachable flowd.
---

# Close an allowed-pairs experiment

An experiment overrides roles' base pairs (**Allowed pairs**,
`skills/flow-contracts/model-policy.md`) while it is open. This skill judges one from
the figures flowd stored, closes it, and — on a positive verdict — makes its cells those roles'
base pairs in flowd, the pairs' one source (decision `promote-experiment`).

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
`before → after`, and the cells recommended. On a positive verdict go to step 5 with it; otherwise
run:

```bash
flow experiment close <id> -verdict <positive|negative|inconclusive> -summary "$(cat <<'SUMMARY'
<summary>
SUMMARY
)"
```

The quoted heredoc passes the summary verbatim, apostrophes included.

## 5. On a positive verdict, make the winners permanent

Ask whether to make the cells permanent — options **Make permanent** (recommended: the evidence
supports them) and **Close without promoting**. On **Make permanent**, run:

```bash
flow experiment promote <id> -summary "$(cat <<'SUMMARY'
<summary>
SUMMARY
)"
```

It writes every cell but `main` into that role's base pairs, keeping its opt-in and marked pairs,
and closes the experiment `positive` with the summary — the Experiments view's **Make permanent**
button does the same. Where only some cells are supported, promote and then edit the rest back by
hand in the Experiments view. On **Close without promoting**, run step 4's
`flow experiment close <id> -verdict positive …`.

On `negative` or `inconclusive` the pairs are left as they are; report the closed experiment and stop.
