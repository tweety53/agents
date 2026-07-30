# myflow-finish-cleanup Specification

## Purpose
TBD - created by archiving change kan-8-myflow-updates. Update Purpose after archive.
## Requirements

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
base branch. The confirmation SHALL rest on three signals, evaluated in this order, and SHALL be
performed by an executable check rather than by prose an agent follows:

1. **`HEAD` against the merge base recorded in the state file's `worktrees` map.** When they are
   equal the branch has no commits of its own, and run 2 SHALL NOT happen whatever any ancestor test
   says.
2. **The ancestor test** — a PR CLI when one is usable for the host, and `git merge-base
   --is-ancestor` otherwise. The `git merge-base` path SHALL remain reachable on its own, because it
   is the only merge evidence available on a non-GitHub forge.
3. **The worktree's cleanliness.** A branch that genuinely reached the base branch has nothing left
   to commit; uncommitted tracked changes or untracked-unignored files at this point SHALL refuse
   run 2.

Signal 1 SHALL be evaluated before signal 2. A branch with no commits of its own is an ancestor of
every branch, so the ancestor test alone reports *merged* on precisely the input where archiving
destroys the work. Signal 3 SHALL be evaluated after signal 2, because a dirty unmerged worktree is
the ordinary state at `IN_PROGRESS` and SHALL yield run 1 rather than a refusal.

Counting commits ahead of the base branch SHALL NOT be used as a signal. A branch with no commits
and a branch whose commits have been merged both report zero, so that test cannot separate the
dangerous state from the correct terminal one.

When no merge base is recorded for a worktree, the outcome SHALL be a refusal that reports what is
known and asks the operator, never an inferred verdict. On a change spanning several repositories,
run 2 SHALL proceed only when every recorded worktree returns the same merged verdict.

If the merge is not verified, run 2 SHALL NOT happen — the command falls to run 1 behaviour or
reports what is outstanding, and archives nothing.

#### Scenario: A branch with staged but uncommitted work is never archived

- **WHEN** `/myflow-finish` runs on a branch whose `HEAD` equals the merge base recorded in the state
  file, with work staged and no commits made
- **THEN** the verification returns the integrating verdict, not the archiving one
- **AND** nothing is synced, archived, pushed, or removed

#### Scenario: An unmerged change is never archived

- **WHEN** the branch has not reached the base branch
- **THEN** the change directory remains outside the archive and no state is written to `FINISHED`

#### Scenario: Merge evidence without a PR CLI

- **WHEN** no usable PR CLI exists for the remote host
- **THEN** `git merge-base --is-ancestor` decides, and a positive result together with the other two
  signals is sufficient to proceed

#### Scenario: A merged branch with a dirty worktree refuses

- **WHEN** the branch is an ancestor of the base branch but its worktree still holds uncommitted
  tracked changes or untracked-unignored files
- **THEN** run 2 refuses, reports `HEAD`, the base branch and how many entries are uncommitted, and
  archives nothing until the operator explicitly confirms

#### Scenario: No recorded merge base produces a refusal, not a guess

- **WHEN** the state file records no merge base for a worktree
- **THEN** the verification refuses and asks the operator rather than deciding from the ancestor test
  alone

#### Scenario: A shortened recorded sha still compares correctly

- **WHEN** the merge base recorded in the state file is abbreviated and `HEAD` is a full object name
- **THEN** the two are compared as resolved commits, so an abbreviation is not mistaken for a
  difference

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

### Requirement: Session records are preserved in the repository

The SDD ledger and the review panel record SHALL be copied out of the change's gitignored working
directory and into the repository, so that they survive the worktree's removal.

The copy SHALL happen during `/myflow-finish` run 1, before it stages the work, so the records land
in the same commit as the implementation they describe and reach the base branch with it. It SHALL
also happen on `/myflow-do`'s commit path — the case where a pull request already exists and a fix is
committed and pushed — so a fix round raised after integration refreshes the records rather than
leaving them stale.

The destination path for a change SHALL be fixed at the first copy and reused on every later copy,
so repeated runs overwrite in place rather than accumulating one dated file per round.

A source that does not exist SHALL be reported and skipped. It SHALL NOT fail the run: a change may
legitimately have no panel record, and a preservation step able to block an integration would be a
worse failure than the gap it closes.

Every path the copy reads from and writes to SHALL be required to resolve inside the root it belongs
to — the worktree for the records taken from it and for every destination, the state directory for the
artifact source. A path resolving outside its root SHALL be refused: reported as a failure of that
one copy, distinctly from the skip that reports an absent source, while the remaining records are
still copied. The refusal exists because the copy runs automatically and its result is committed and
pushed, so a planted symlink at any of those paths would otherwise read or write an arbitrary file
under the repository's name.

A change name SHALL be required to be a single plain path component before any directory is touched.
A name carrying a path separator or a glob metacharacter SHALL be rejected outright, because the
search for an already-preserved file is anchored on the name and a metacharacter in it would let one
change adopt and overwrite another change's preserved record.

Records other than the ledger and the panel record SHALL NOT be preserved. Per-task diffs duplicate
commits already present in git history.

Preservation SHALL NOT change what the ledger may contain. A dispatch whose model the dispatcher
could not observe SHALL still record `unknown (agent-defined)`; durability SHALL NOT be a reason to
fill such an entry with a plausible value.

#### Scenario: The ledger survives the change

- **WHEN** a change is integrated and later archived, and its worktree is removed
- **THEN** the SDD ledger is present in the repository at its preserved path
- **AND** a reader can still determine which model implemented each task

#### Scenario: A missing source is skipped, not fatal

- **WHEN** a change has no review panel record at the expected path
- **THEN** the absence is reported in the run's output
- **AND** the integration proceeds and the remaining records are still preserved

#### Scenario: A fix round refreshes rather than duplicates

- **WHEN** a fix is committed to a branch that already has a pull request, after records were already
  preserved
- **THEN** the existing preserved files are overwritten in place
- **AND** no second dated copy is created for the same change

#### Scenario: A destination outside the worktree is refused, not followed

- **WHEN** one of the destination directories under `docs/superpowers/` resolves outside the worktree
- **THEN** nothing is written through it, the refusal is reported as a failure rather than a skip, and
  the remaining records are still preserved

#### Scenario: A source outside its own root is refused, not read

- **WHEN** one of the three sources resolves outside the root it belongs to — the worktree for the
  records taken from it, the state directory for the artifact source
- **THEN** its content is not copied into the repository, the refusal is reported as a failure rather
  than a skip, and the remaining records are still preserved

#### Scenario: A change name that is not one plain component is rejected

- **WHEN** the copy is invoked with a change name containing a path separator or a glob metacharacter
- **THEN** it is rejected before any directory is created or any file is written
- **AND** no other change's preserved record is read or overwritten

#### Scenario: An unobservable model stays unobserved

- **WHEN** a preserved ledger contains a slot dispatched by agent type, whose model the dispatcher
  never read
- **THEN** that entry records `unknown (agent-defined)`
- **AND** preservation does not substitute a guess

### Requirement: Run 2 removes the proposal artifact source

`/myflow-start` writes the published proposal's HTML source beside the state file so that a revision
round can republish to the same URL. `/myflow-finish` SHALL account for that file rather than leaving
its fate to whether an operator noticed it.

The source SHALL be preserved into the repository alongside the other session records at run 1, and
the copy beside the state file SHALL be removed during run 2, in the same disclosed cleanup sequence
as the worktree removal.

Removal SHALL NOT happen without the preservation copy existing, because the recorded `artifactUrl`
remains in the terminal state file indefinitely and deleting the only source that could republish it
would leave a URL advertised and unrepublishable.

#### Scenario: A finished change leaves no artifact source behind

- **WHEN** run 2 completes successfully
- **THEN** no `<name>-proposal-artifact.html` remains beside the state file
- **AND** the same content is present in the repository

#### Scenario: The recorded URL stays republishable

- **WHEN** an operator wants to republish a finished change's proposal to its recorded `artifactUrl`
- **THEN** the artifact source is recoverable from the repository

#### Scenario: Removal is skipped when nothing was preserved

- **WHEN** run 2 reaches the artifact-source removal and no preserved copy exists in the repository
- **THEN** the state-directory file is left in place and the situation is reported
