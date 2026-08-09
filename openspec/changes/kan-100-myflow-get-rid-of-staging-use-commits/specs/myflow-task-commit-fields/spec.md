## ADDED Requirements

### Requirement: Every task declares mechanically-checkable fields

`/myflow-start`'s writing-plans stage SHALL tag every task in `tasks.md` with `**Files:**`,
`**Tests:**`, `**Regression:**`, `**Baseline:**`, and `**Commit:**`, alongside the existing
`**Build:**` tag. A task tagged `Build: red` SHALL additionally carry `**Squash-with:** Task <N>`,
naming the green task its commit folds into.

- `**Files:**` declares the paths this task touches. An optional `**Allowed-collateral:**` glob
  names paths a legitimate sweep may also touch without being a declared file.
- `**Tests:**` declares the names of tests this task adds.
- `**Regression:**` declares, per test, what fails if the task is reverted.
- `**Baseline:**` declares `before=<N> after=<N>` expected test counts.
- `**Commit:**` declares the commit subject line the task's implementer must use.

#### Scenario: A task carries all required fields

- **WHEN** a task is written into `tasks.md` during writing-plans
- **THEN** it carries `Build:`, `Files:`, `Tests:`, `Regression:`, `Baseline:`, and `Commit:` tags

#### Scenario: A red task also carries Squash-with

- **WHEN** a task is tagged `Build: red`
- **THEN** it also carries `**Squash-with:** Task <N>` naming its green partner

### Requirement: A runtime guard checks each field against the real commit

Immediately after `/myflow-do` commits a task, and before that task is dispatched for review, a
guard SHALL check the task's declared fields against the actual commit:

- `**Files:**` — `git diff --name-only <parent>..<task-sha>` SHALL be a subset of the declared
  files, or covered by the `**Allowed-collateral:**` glob when present.
- `**Tests:**` — each declared test name SHALL exist in the commit's diff.
- `**Commit:**` — the commit's actual subject line SHALL match the declared one.

A guard failure on any of these three SHALL send the task back to the same implementer rather than
counting as a review finding.

#### Scenario: An undeclared file fails the guard

- **WHEN** a task's commit touches a path not in its declared `Files:` list and not covered by
  `Allowed-collateral:`
- **THEN** the guard fails and the task returns to the same implementer, without being dispatched
  for review

#### Scenario: A declared test that was never written fails the guard

- **WHEN** a task declares a test name under `Tests:` that does not appear in the commit's diff
- **THEN** the guard fails and the task returns to the same implementer

#### Scenario: A commit subject mismatch fails the guard

- **WHEN** a task's actual commit subject does not match its declared `Commit:` line
- **THEN** the guard fails and the task returns to the same implementer

#### Scenario: A clean task passes

- **WHEN** a task's commit's changed files are all declared, its declared tests all exist in the
  diff, and its subject matches `Commit:`
- **THEN** the guard passes and the task is dispatched for review

### Requirement: Regression and Baseline checks skip, rather than fail, when unsupported

The guard SHALL verify `**Regression:**` by temporarily reverting the task's commit, running the
named test, confirming it fails, and un-reverting; and SHALL verify `**Baseline:**` by running the
project's `## test` command at the task's parent commit and at the task's own commit, comparing the
counts to the declared `before=<N> after=<N>`.

When the project's `## test` command cannot target a named test, or does not report a parseable
count, the corresponding check SHALL report **skipped, not verified** rather than failing the run.

#### Scenario: A regression check passes

- **WHEN** reverting a task's commit causes its declared test to fail, and un-reverting restores it
  to passing
- **THEN** the `Regression:` check passes

#### Scenario: A baseline check passes

- **WHEN** the project's `## test` command reports counts at the parent and task commits matching
  the declared `before=<N> after=<N>`
- **THEN** the `Baseline:` check passes

#### Scenario: An unsupported project skips rather than fails

- **WHEN** the project's `## test` command cannot target a named test, or cannot report a parseable
  count
- **THEN** the corresponding `Regression:` or `Baseline:` check reports skipped-not-verified, and the
  run continues
