# myflow-uncommitted-review-package Specification

## Purpose
TBD - created by archiving change kan-53-myflow-no-uncommitted-diff-helper-for-subagent. Update Purpose after archive.
## Requirements

### Requirement: A checkpoint isolates uncommitted work without touching any ref

`/myflow-do`'s dispatch loop SHALL obtain an isolating BASE for each task and each fix round by
running `skills/myflow-do/scripts/checkpoint`, which SHALL print `git stash create`'s output when
the worktree carries uncommitted changes, or the current `HEAD` sha when the worktree is clean.
`checkpoint` SHALL NOT modify `HEAD`, the index, or the stash list under either outcome.

#### Scenario: Dirty tree yields a stash-create snapshot

- **WHEN** `checkpoint` runs against a worktree with uncommitted changes
- **THEN** it prints a valid commit-ish distinct from `HEAD`
- **AND** `git status`, `git stash list`, and `HEAD` are unchanged by the call

#### Scenario: Clean tree falls back to HEAD

- **WHEN** `checkpoint` runs against a worktree with no uncommitted changes
- **THEN** it prints the current `HEAD` sha, because `git stash create` itself prints nothing when
  there is nothing to snapshot

### Requirement: A per-task or per-fix-round review package diffs BASE against the live working tree

`skills/myflow-do/scripts/uncommitted-review-package PLAN_FILE BASE [OUTFILE]` SHALL write a review
package containing a files-changed summary (`git diff --stat BASE`) and a diff with extended context
(`git diff -U10 BASE`), diffing `BASE` directly against the current working tree rather than against
`HEAD`. The package SHALL NOT include a commit-log section, since NO-COMMITS means no commits exist
to list. The package SHALL be written into the same plan-scoped `.superpowers/sdd/<plan>/` workspace
that `subagent-driven-development`'s `review-package` script resolves via its `sdd-workspace` helper,
so both scripts' output files sit side by side.

#### Scenario: A later task's changes are isolated from an earlier task's

- **WHEN** task 1 makes uncommitted changes, a checkpoint `BASE` is recorded, and task 2 then makes
  further uncommitted changes
- **THEN** `uncommitted-review-package PLAN_FILE BASE` produces a diff containing only task 2's
  changes, not task 1's

#### Scenario: The output file has no Commits section

- **WHEN** `uncommitted-review-package` writes its package
- **THEN** the file contains a `## Files changed` section and a `## Diff` section
- **AND** the file contains no `## Commits` section

#### Scenario: A re-review after a fix round gets a distinct file

- **WHEN** `uncommitted-review-package` is invoked again with a different `BASE` and an explicit
  `OUTFILE` naming the fix round
- **THEN** the new package is written to that distinct path, leaving the prior package file intact

### Requirement: Invalid arguments fail the same way review-package's do

`uncommitted-review-package` SHALL exit non-zero with a message identifying the problem when
`PLAN_FILE` does not exist, or when `BASE` does not resolve via `git rev-parse --verify --quiet` —
mirroring the two argument checks `subagent-driven-development`'s `review-package` already performs
on its own arguments.

#### Scenario: Missing plan file

- **WHEN** `uncommitted-review-package` is given a `PLAN_FILE` path that does not exist
- **THEN** it exits non-zero and reports "no such plan file: `<path>`"
- **AND** it writes no output file

#### Scenario: Unresolvable BASE

- **WHEN** `uncommitted-review-package` is given a `BASE` that `git rev-parse --verify --quiet`
  cannot resolve
- **THEN** it exits non-zero and reports "bad BASE: `<base>`"
- **AND** it writes no output file
