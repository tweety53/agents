# Design — guard against a stale Baseline: count (KAN-353)

Date: 2026-09-08
Change: kan-353-flow-guard-against-a-stale-baseline-count

## Problem

A planner wrote a plan's `Baseline:` chain from a stale build/test-results
directory — 1690 against a real 1768 (KAN-339) — and every task's `Baseline:`
had to be rewritten mid-run. Nothing requires the baseline to come from a fresh
run: `check-task-commit-fields.py` parses `before=/after=` and verifies nothing
(KAN-442 removed the runtime check deliberately), and the field form escapes
plan-provenance attribution, which only fires on digit + unit-word claims.

## Approach

A new shipped guard, `check-baseline-fresh.sh`, refuses a `Baseline:` count
whose source test-results directory predates the worktree's newest commit. The
source directory is project-declared — an optional `## baseline results dirs`
key in `.flow/project.md`, one worktree-relative path per line — because
nothing else records where the numbers came from, and the project is the only
authority on where its suites leave results.

The guard bites at plan-writing time only (section D of the planner): by
implementation time, mid-run test executions have refreshed any results
directory and masked the original staleness; KAN-442 keeps the runtime side
git/files-only.

## Mechanics

- `check-baseline-fresh.sh <changeRoot>`: thin Bash wrapper over stdlib-Python
  `check-baseline-fresh.py`, which reuses `lib/plan_grammar.py`'s tasks.md
  parser. Source lives in `scripts/`, symlinked into `skills/flow/scripts/`.
- Exits: 0 fresh / nothing to check; 1 stale hit; 2 cannot answer.
- No `Baseline:` fields → 0. No `## baseline results dirs` key → 0 (skip).
  Duplicate key → 2 (the `project-get.sh` ambiguity refusal).
- Key declared: every declared directory must exist under the worktree root
  and carry a recursive max mtime ≥ `git log -1 --format=%ct HEAD`; otherwise
  exit 1 naming field `file:line`, directory, both timestamps, and the fix
  (re-run the suite, re-derive the counts). Absent-but-declared counts as
  stale — absence cannot attest the numbers came from a run in this worktree.

## Wiring

`brainstorm-planner.md` section D runs it beside `check-plan-shape.sh`;
`flow/SKILL.md`'s guard list names it; `project-configuration.md` documents the
key; `check-contract-budget.sh` budgets raised where files grow;
`scripts/test-check-baseline-fresh.sh` harness auto-discovered by the runner.
This repository declares no results directory, so its own plans stay skipped.

## Ruled out

- Runtime re-verification of `Baseline:` — KAN-442 (reverted the shared
  worktree, two full suites per task, no project spoke its protocol).
- Source path in the field grammar — frozen family grows.
- Build-tool convention table in the guard — rotting, wrong for nonstandard
  layouts.
- Extending plan-provenance attribution to `before=/after=` — separate defect
  in a different frozen guard; no staleness benefit; own change.
