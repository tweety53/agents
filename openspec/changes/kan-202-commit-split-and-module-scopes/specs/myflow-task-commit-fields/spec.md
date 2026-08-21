## MODIFIED Requirements

### Requirement: A runtime guard checks each field against the real commit

Immediately after `/myflow-do` commits a task, and before that task is dispatched for review, a
guard SHALL check the task's declared fields against the actual commit:

- `**Files:**` — `git diff --name-only <parent>..<task-sha>` SHALL be a subset of the declared
  files, or covered by the `**Allowed-collateral:**` glob when present.
- `**Tests:**` — each declared test name SHALL exist in the commit's diff.
- `**Commit:**` — the commit's actual subject line SHALL match the declared one.

The guard SHALL additionally check the **declared** `**Commit:**` field's own scope, independently of
whether the real commit matches it. It SHALL fail the task when that scope is the change name, the
change name's bare Jira key, or a dotted or numeric task id, per **A commit's scope names the module,
never the change or the task** (`myflow-commit-scope`). A `**Commit:**` field carrying no scope at all
SHALL pass — a scope is optional, and only a present-and-wrong scope fails.

The change name SHALL be derived from the `tasks.md` path the guard was given, which is already
`<worktree>/openspec/changes/<name>/tasks.md`; the check SHALL NOT require a new argument.

A guard failure on any of these SHALL send the task back to the same implementer rather than
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

#### Scenario: A declared change-name scope fails the guard

- **WHEN** a task's `Commit:` field declares a subject whose scope is the change name, or the
  change name's bare Jira key
- **THEN** the guard fails, naming the scope and the task, even though the real commit reproduces
  that subject exactly

#### Scenario: A declared task-id scope fails the guard

- **WHEN** a task's `Commit:` field declares a subject whose scope is a dotted or numeric task id
- **THEN** the guard fails and the task returns to the same implementer

#### Scenario: A declared subject with no scope passes

- **WHEN** a task's `Commit:` field declares `fix: <subject>`, carrying no scope
- **THEN** the scope check passes, because a scope is optional

#### Scenario: A clean task passes

- **WHEN** a task's commit's changed files are all declared, its declared tests all exist in the
  diff, its subject matches `Commit:`, and that subject's scope names neither the change nor a task
  id
- **THEN** the guard passes and the task is dispatched for review
