## Context

The failure is plan-time: a planner read `Baseline:` counts from a results
directory that predated the worktree's newest commit (KAN-339). By
implementation time, mid-run test executions have already refreshed any results
directory and masked the original staleness, and KAN-442 deliberately keeps the
runtime side git/files-only — so the guard bites at plan-writing time (section
D), the only place the freshness of the numbers' source is still observable.

Nothing in the repository, the contracts or the field grammar records which
directory a `Baseline:` count came from, and this repository's own suites write
no results directory at all — so the source directory is declared by the
project that has one, and a project without one is skipped, per the frozen
tree's "skip, rather than fail, when unsupported" precedent.

Shipped-guard layout this change follows: source in `scripts/`, symlinked into
`skills/flow/scripts/`, non-trivial logic in stdlib Python behind a thin Bash
wrapper (the `check-plan-provenance` / `check-plan-shape` split), tasks.md
parsing reused from `lib/plan_grammar.py` rather than reimplemented.

## How it works

- Invocation: `check-baseline-fresh.sh <changeRoot>` — one argument, the change
  root; the guard derives the worktree root by walking up to the directory
  holding `.flow/project.md`.
- Exit contract: 0 fresh or nothing to check; 1 stale hit; 2 cannot answer.
- No `**Baseline:** before=<N> after=<M>` field in the change's `tasks.md`
  → 0, nothing to check.
- No `## baseline results dirs` key in `.flow/project.md` → 0, skip. Duplicate
  key declarations → 2, the same ambiguity refusal `project-get.sh` makes.
- Key declared (one worktree-relative path per line): every declared directory
  must exist under the worktree root and carry a recursive max mtime ≥ the
  worktree's newest commit time (`git log -1 --format=%ct HEAD`). Any declared
  directory absent or older → 1, naming the field's `file:line`, the
  directory, both epoch timestamps, and the fix: re-run the suite, re-derive
  the counts. All offending directories are listed in one report, not one per
  task field.
- Absent-but-declared counts as stale: a directory that does not exist cannot
  attest the numbers came from a run in this worktree — that absence is the
  KAN-339 shape.

## Wiring

- `skills/flow/brainstorm-planner.md` section D: run
  `check-baseline-fresh.sh <changeRoot>` unconditionally beside
  `check-plan-shape.sh`; fix any hit before continuing.
- `skills/flow/SKILL.md`: the guard-presence list gains
  `check-baseline-fresh.sh`.
- `skills/flow-contracts/project-configuration.md`: the optional key
  documented.
- `scripts/check-contract-budget.sh`: budgets raised for every file that
  grows.
- `scripts/test-check-baseline-fresh.sh`: harness, auto-discovered by
  `scripts/run-guard-tests.sh`. This repository declares no results directory,
  so its own plans stay skipped.

## Decisions

### Baseline source directory is project-declared

**ID:** baseline-source-project-declared
**Status:** active
**Chosen:** `.flow/project.md` `## baseline results dirs` — the project is the only authority on where its suites leave results; no field-grammar change; undeclared projects skip.
**Considered:** source path in the `Baseline:` field — the frozen field family grows a component every plan must write and every guard must parse; build-tool convention table inside the guard — a rotting table in shipped code, wrong for nonstandard layouts, speculative for projects with no convention.

### The guard bites at plan time only

**ID:** plan-time-enforcement-only
**Status:** active
**Chosen:** run in section D beside `check-plan-shape.sh` — the only point where a results directory's staleness against the worktree's newest commit is still observable.
**Considered:** implementation-time/runtime re-verification of `Baseline:` — ruled out by KAN-442 (it reverted the shared worktree, cost two full suites per task, and no project's `## test` command spoke its `COUNT:`/`RESULT:` protocol); mid-run runs refresh the directory and mask staleness.

### A separate shipped guard, not a fold-in

**ID:** separate-shipped-guard
**Status:** active
**Chosen:** new `check-baseline-fresh.sh` beside `check-plan-shape.sh` — different invocation site (change root + project config + git), conditional skip semantics, and `check-task-commit-fields.py` deliberately verifies nothing.
**Considered:** folding into `check-plan-shape.py` — mixes grammar validation with environment-dependent freshness, breaking its unconditional bare-tree contract; folding into `check-task-commit-fields` — runs per-commit at implementation time, where staleness is already masked.

### Provenance gap left alone

**ID:** provenance-gap-not-fixed
**Status:** active
**Chosen:** do not extend `check-plan-provenance.py`'s numeric rule to `before=/after=` numbers in this change — a separate defect in a different frozen guard, with no staleness benefit; reshaping every future plan's provenance tagging is scope the issue does not ask for.
**Considered:** fixing it here — blast radius on every future plan's provenance obligations; deserves its own change.

## Open questions

*(none — the one question the repository could not answer, which directory a `Baseline:` count's freshness is judged against, was answered by the operator: project-declared.)*
