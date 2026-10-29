# kan-1067-flow-allowed-pairs-experiments-stored-in-flowd

## Why

- The **Allowed pairs** table (`skills/flow-contracts/model-policy.md`) is changed by hand on intuition; nothing measures whether a pair change made runs better or worse.
- KAN-1067 asks for time-boxed experiments: an override stored in flowd, trends from the start date, and an agent that closes the experiment with a summary and makes a winner permanent.
- Added during planning (2026-10-29): the **main (parent) agent's** model and effort is an experiment type; one experiment may combine several roles; the stats UI shows the current roles table; brainstorming and plan writing move to a dispatched `planner` role.

## What changes

- flowd stores experiments: open/closed, start/end date, per-role pair overrides (whole cells, one or many roles — a combined experiment), a summary and a verdict. One open experiment per role.
- `flow experiment open|list|overrides|stats|close` drives them.
- While an experiment is open, every reader of the table (check-dispatch-pair, check-loop-dispatch, plan-agent-pair, plan-class, pipeline-fix) reads its override on top of the markdown table; flowd unreachable → the table's own pairs.
- A `main` role cell names a parent model/effort: `/flow` prints one warning line when its session differs, `decision.json`'s `parent` gains `model`, and only runs whose transcript-harvested main-thread model/effort match count in the "after" figures.
- `flow tasks count` also records the plan's `Correction (…)` paragraph count; verify reds are derived from existing verify dispatch rows.
- The stats SPA gains an **Experiments** view: the current roles table with open overrides applied, the experiment list, override table, per-metric before/after stat panels, one weekly chart per metric with the experiment window and before/after means, and the closing summary.
- `/flow`'s brainstorming, artifacts and writing-plans run in a dispatched `planner` role, the parent relaying its questions, so planning has its own pair to experiment on.
- `/flow-experiment-close` analyses an experiment, closes it with a summary and verdict, and on a positive verdict starts a `/flow` change writing the winning pairs into `model-policy.md`.
