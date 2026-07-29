## ADDED Requirements

### Requirement: Finish is a two-run command

`/myflow-finish` SHALL branch on whether the change's branch has already reached the base branch.

- **Run 1 — the branch is not merged.** It integrates: commits, pushes, and lands the branch by
  whichever route the operator chose. It then **stops**, leaving the change at `IN_PROGRESS`.
- **Run 2 — the branch is merged.** It syncs delta specs, archives the change, commits and pushes
  the archive, removes the worktrees, and writes `FINISHED`.

The branch SHALL be the only thing that decides which run happens; no field records "integration
started". Re-invoking after a merge that happened by any route — a human merging the PR on the
forge, a colleague, or run 1's own merge — proceeds to run 2.

#### Scenario: An unmerged change integrates and stops

- **WHEN** `/myflow-finish` runs and the branch has not reached the base branch
- **THEN** it integrates by the chosen route, leaves the state at `IN_PROGRESS`, and its handoff
  names `/myflow-finish` again as the next command

#### Scenario: A merged change archives

- **WHEN** `/myflow-finish` runs and the branch has reached the base branch
- **THEN** it syncs, archives, commits, pushes, cleans up worktrees, and writes `FINISHED`

#### Scenario: A PR merged by a human is picked up

- **WHEN** run 1 opened a PR, the human merged it on the forge, and `/myflow-finish` is run again
- **THEN** it detects the merge and performs run 2 without asking anything

### Requirement: Run 1 asks how to land the branch, before doing anything

On its first run `/myflow-finish` SHALL ask, **before performing any git action**, how the branch
should be landed:

- **Open a pull request** *(default, recommended)* — commit, push, open a PR, record `prUrl`, stop
- **Merge and push** — commit, push, merge into the base branch, push that
- **Handle it manually** — commit and push the branch only, then stop and say what is left to do

Having asked once, the run SHALL proceed to completion without a further interruption. The answer
SHALL never be remembered between runs and never inferred from anything else.

#### Scenario: The default opens a PR and stops

- **WHEN** the operator accepts the default
- **THEN** the branch is committed, pushed and a PR opened; `prUrl` is recorded; nothing is merged

#### Scenario: Merging is chosen

- **WHEN** the operator chooses to merge and push
- **THEN** the branch is merged into the base branch and pushed, and the next `/myflow-finish`
  run will archive

#### Scenario: Manual handling commits but does not integrate

- **WHEN** the operator chooses to handle it manually
- **THEN** the branch is committed and pushed, no PR is opened and nothing is merged, and the
  handoff states what the operator must do before running `/myflow-finish` again

#### Scenario: The base branch is resolved, never assumed

- **WHEN** finish needs the base branch
- **THEN** it resolves the repository's default branch rather than assuming `main` or `develop`

### Requirement: No verification gate runs before integration

`/myflow-finish` SHALL NOT run tests, linters, or a spec-coverage check before opening a PR or
merging.

Correctness is established during `/myflow-do` — test-driven development per task, per-task
review, and the final review panel — and by the human gate at `IN_PROGRESS`. A second gate
immediately before integration re-runs work already done and delays the one irreversible step.

#### Scenario: Integration does not re-run the project's tests

- **WHEN** `/myflow-finish` integrates a change
- **THEN** it runs no test command, no lint command, and no coverage comparison against the delta
  specs

### Requirement: Run 2 verifies the merge before changing anything

Before syncing or archiving, `/myflow-finish` SHALL confirm that the branch actually reached the
base branch, using a PR CLI when one is usable for the host and `git merge-base --is-ancestor`
otherwise. The `git merge-base` path SHALL remain reachable on its own, because it is the only
merge evidence available on a non-GitHub forge.

If the merge is not verified, run 2 SHALL NOT happen — the command falls to run 1 behaviour or
reports what is outstanding, and archives nothing.

#### Scenario: An unmerged change is never archived

- **WHEN** the branch has not reached the base branch
- **THEN** the change directory remains outside the archive and no state is written to `FINISHED`

#### Scenario: Merge evidence without a PR CLI

- **WHEN** no usable PR CLI exists for the remote host
- **THEN** `git merge-base --is-ancestor` decides, and a positive result is sufficient to proceed

### Requirement: Run 2 commits and pushes the archive

After the delta-spec sync and the archive move, `/myflow-finish` SHALL commit the resulting
changes on the base branch in the main checkout and push them.

There is no merge to do in the normal case: the change branch was already merged, which the
verification step proved. When `/myflow-finish` is invoked with a non-base branch checked out, it
SHALL commit there, merge into the base branch, and push that.

A finished change SHALL NOT leave the archive move uncommitted in the working tree.

#### Scenario: The archive move reaches the remote

- **WHEN** `/myflow-finish` completes run 2 successfully
- **THEN** `git status` in the main checkout shows no pending archive changes, and the archive
  commit is present on the pushed base branch

#### Scenario: Nested fix sub-changes are archived together

- **WHEN** a change has nested `<name>-fix-N` sub-changes that are not yet archived
- **THEN** they are archived in the same operation and included in the same commit

### Requirement: Run 2 removes the change's worktrees and branches

After the archive is committed and pushed, `/myflow-finish` SHALL remove every worktree belonging
to the change. The set SHALL be the keys of the state file's `worktrees` map; when that is absent,
it SHALL be found by scanning `git worktree list` in each affected repository for branch
`openspec/<name>`.

A worktree path SHALL NEVER be guessed or assumed from a conventional layout, because layout
differs per repository.

For each worktree, all of the following SHALL be checked before anything is removed:

- no uncommitted tracked changes
- no untracked files that git does not already ignore
- no commits that exist only in this worktree (already being merged into the base branch
  satisfies this; otherwise an upstream must prove they are pushed)
- the project's local stack is stopped

plus a **disclosure**, which does not gate removal: the ignored files `--force` will destroy are
listed, and the operator confirms. `--force` destroys every ignored file, and "ignored" is not
"disposable" — no check prevents that, so it is surfaced instead.

Only when every gating check passes, and the operator has confirmed any disclosed ignored files,
SHALL it run `git worktree remove --force`, then `git branch -d`, then `git worktree prune`.

`git branch -d` SHALL never be `-D`: it must be free to refuse an unmerged branch.

#### Scenario: A clean worktree is removed

- **WHEN** every gating check passes for a worktree
- **THEN** the worktree is removed, its branch is deleted with `git branch -d`, and
  `git worktree prune` runs

#### Scenario: A failed check leaves everything alone

- **WHEN** any one of the gating checks fails for any worktree
- **THEN** no worktree is removed and no branch is deleted, and the reason is reported

#### Scenario: An already-removed worktree is not an error

- **WHEN** a worktree recorded in the state file no longer exists
- **THEN** that is treated as success and the run continues

#### Scenario: Ignored files are disclosed, not silently destroyed

- **WHEN** a worktree contains ignored files alongside a clean tracked tree
- **THEN** every gating check passes, the ignored files are listed to the operator, and removal
  proceeds only on explicit confirmation — because `--force` destroys them and no check prevents it

#### Scenario: A merged branch with no upstream can still be cleaned up

- **WHEN** the branch was squash-merged and its remote tracking ref was pruned, so `@{upstream}`
  no longer resolves
- **THEN** the commits-exist-elsewhere check passes on the strength of the branch already being an
  ancestor of the base branch, and cleanup is not locked out

### Requirement: The stack-stopped check reads a project-supplied command

`.myflow/project.md` SHALL support an optional `## stop` section naming the command that stops the
project's local stack. `/myflow-finish` SHALL run it before the stack-stopped check.

When the key or the file is absent, the stack-stopped check SHALL be skipped rather than failed,
and cleanup SHALL proceed on the strength of the other two checks.

#### Scenario: A project without a stack cleans up normally

- **WHEN** a project has no `## stop` key
- **THEN** cleanup proceeds using the uncommitted-changes and unpushed-commits checks alone

#### Scenario: A declared stop command runs first

- **WHEN** a project declares `## stop`
- **THEN** that command runs before the worktree is removed, so nothing holds a port or file
  handle open
