# kan-353-flow-guard-against-a-stale-baseline-count

## Why

KAN-339: a plan's `Baseline:` chain was written from a stale build/test-results
directory — 1690 against a real 1768 — and every task's `Baseline:` had to be
rewritten mid-run. Nothing requires the baseline to come from a fresh run:
`check-task-commit-fields.py` parses `before=/after=` and verifies nothing
(KAN-442 removed the runtime check deliberately), and the `before=<N> after=<M>`
field form escapes plan-provenance attribution, which only fires on
digit + unit-word claims.

## What changes

- New shipped guard `check-baseline-fresh.sh` (+ stdlib-Python
  `check-baseline-fresh.py` reusing `lib/plan_grammar.py`): refuses a
  `Baseline:` count whose source test-results directory predates the worktree's
  newest commit.
- The source directory is project-declared: a new optional
  `## baseline results dirs` key in `.flow/project.md`, one worktree-relative
  path per line. No key or no `Baseline:` fields → skip, exit 0.
- Wiring: `skills/flow/brainstorm-planner.md` section D runs it beside
  `check-plan-shape.sh`; `skills/flow/SKILL.md`'s guard list names it;
  `skills/flow-contracts/project-configuration.md` documents the key;
  `scripts/check-contract-budget.sh` budgets raised for files that grow;
  `scripts/test-check-baseline-fresh.sh` harness (auto-discovered by the
  runner).
- This repository declares no results directory, so the guard stays skipped on
  its own plans.
