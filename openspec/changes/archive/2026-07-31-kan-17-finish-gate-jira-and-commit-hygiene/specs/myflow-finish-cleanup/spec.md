## ADDED Requirements

### Requirement: Run 1 refuses to integrate silently over unfinished work

Before it asks how the branch should land, and before any git action, `/myflow-finish` SHALL
determine whether the change carries unfinished work, by running a script rather than by asserting
it in prose. The script SHALL be run once per worktree recorded in the state file's `worktrees` map.

Four signals SHALL count as unfinished:

- unticked checklist boxes in `docs/manual-test/<name>.md`
- unchecked items in `openspec/changes/<name>/tasks.md` and in any nested `<name>-fix-N` sub-change
- findings whose recorded status is open in `.superpowers/sdd/final-review-panel.md`
- a non-empty `## Known incomplete` section in the manual test guide

The script SHALL print exactly one verdict line — `CLEAR` when no signal fires, or `OUTSTANDING`
with a per-signal breakdown — and SHALL exit non-zero **with no verdict line** when it cannot read
the worktree. A run that receives no verdict line SHALL stop and ask the operator; it SHALL NOT read
missing output as either verdict, and SHALL check the exit code as well as the line.

A file it cannot find, and a manual test guide with no `## Known incomplete` section, SHALL count as
outstanding rather than as clear. Treating silence as clearance is the failure this requirement
exists to prevent.

On `OUTSTANDING` the run SHALL show the breakdown and offer the operator exactly three courses:
continue and integrate anyway; stop so the work can be finished; or file a tracker issue carrying
the outstanding items and then continue. Filing SHALL follow the labelling rules for issues the
pipeline creates.

Choosing to stop SHALL leave the change at `IN_PROGRESS` with no git action performed.

#### Scenario: An unticked guide stops the run before it asks how to land

- **WHEN** run 1 begins for a change whose manual test guide has unticked boxes
- **THEN** the breakdown is shown and the three courses are offered before the landing question is
  asked, and before any git command runs

#### Scenario: A clean change is not interrupted

- **WHEN** every signal is clear
- **THEN** run 1 proceeds directly to the landing question with no extra prompt

#### Scenario: An unreadable worktree stops the run

- **WHEN** the script exits non-zero without printing a verdict line
- **THEN** the run stops and asks the operator, and integrates nothing

#### Scenario: A missing Known incomplete section counts as outstanding

- **WHEN** the manual test guide predates this requirement and carries no `## Known incomplete`
  section
- **THEN** that signal counts as outstanding and the operator is prompted once

#### Scenario: Filing a task records the outstanding work and continues

- **WHEN** the operator chooses to file a tracker issue
- **THEN** an issue carrying the outstanding items is created, linked to the change's issue, and
  the run continues to the landing question

### Requirement: Integrating over known-unfinished work is recorded durably

When the operator chooses to integrate anyway, the outstanding list SHALL appear both in run 1's
handoff and in the message of the commit that carries the change's planning artifacts, so the
repository's history answers whether anything was known to be unfinished when the change merged.

A record that exists only in the session transcript SHALL NOT satisfy this requirement.

#### Scenario: The history records what was outstanding

- **WHEN** the operator continues past the gate with outstanding items
- **THEN** the planning commit's message lists those items
- **AND** the handoff lists them too

### Requirement: Every artifact the pipeline creates is enumerated in one registry

The pipeline's temporary and preserved artifacts SHALL be listed in a single registry in the
canonical pipeline contract, each row naming what creates the artifact, where it lives, and what
removes it. The registry SHALL cover the per-task and review diffs, the review panel record, the
session ledger, the proposal artifact source, each worktree, the local branch, the remote branch,
the change directory, and the state file.

Cleanup rules SHALL NOT be stated in more than one place. A file describing a removal SHALL point at
the registry rather than restate it.

#### Scenario: An artifact with no owner is a defect in the registry

- **WHEN** the pipeline creates an artifact that no registry row accounts for
- **THEN** the registry is incomplete and SHALL be corrected rather than the artifact left
  unaccounted for

### Requirement: Run 2 verifies that cleanup actually completed

After removing what it removes, `/myflow-finish` SHALL verify, by running a script, that every
registry row whose lifetime ends at run 2 is in fact gone, and SHALL report anything still present
rather than assuming removal succeeded.

The verification SHALL be a separate script from the run-1 gate, because it answers a different
question at a different point in the run.

**A leftover, and a verification that could not be performed, SHALL both block the `FINISHED`
write.** The run SHALL name what remains and leave the change at `IN_PROGRESS`. `FINISHED` is
terminal — no command acts on a change at that state — so a change written `FINISHED` over a known
leftover leaves that leftover recorded only in the session transcript, which this capability's own
requirement on recording unfinished work SHALL NOT accept. The state the change already carries is
the durable record; no new field is added to hold the same fact.

Run 2 SHALL therefore be re-entrant: every removal step is performed only if its artifact is still
present, an artifact already gone SHALL count as success rather than as an error, and an
already-archived change directory SHALL mean the sync-and-archive step is skipped rather than
repeated.

The set of rows the verification checks SHALL be derived from the registry rather than restated
independently of it, and the derivation SHALL be checkable in both directions: a registry row the
verification accounts for nowhere, and an accounted-for row the registry no longer carries, SHALL
each be a failure.

#### Scenario: A leftover is reported rather than assumed away

- **WHEN** an artifact whose lifetime ends at run 2 is still present after cleanup
- **THEN** it is named in the report and the operator is told what remains

#### Scenario: A complete cleanup is confirmed

- **WHEN** every registry row whose lifetime ends at run 2 is gone
- **THEN** the run reports the cleanup as verified

#### Scenario: A leftover blocks the terminal state

- **WHEN** the verification reports that something remains, or cannot produce a verdict at all
- **THEN** `FINISHED` is not written, the change stays at `IN_PROGRESS`, and the handoff names what
  remains and points back at `/myflow-finish`

#### Scenario: Re-running after the leftover is cleared finishes the change

- **WHEN** the operator clears what remained and runs `/myflow-finish` again
- **THEN** the steps whose artifacts are already gone are treated as done, the verification runs
  again, and `FINISHED` is written once it reports a complete cleanup

#### Scenario: A registry row nothing verifies is a failure

- **WHEN** the registry gains a row whose lifetime ends at run 2 and the verification accounts for
  it nowhere
- **THEN** that is a failure to be corrected, rather than a row silently left unverified

## MODIFIED Requirements

### Requirement: Run 1 asks how to land the branch, before doing anything

On its first run `/myflow-finish` SHALL ask, **before performing any git action**, how the branch
should be landed:

- **Open a pull request** *(default, recommended)* — commit, push, open a PR, record `prUrl`, stop
- **Merge and push** — commit, push, merge into the base branch, push that
- **Handle it manually** — commit and push the branch only, then stop and say what is left to do

The unfinished-work gate SHALL run **before** this question. The operator SHALL NOT be asked how to
land a branch and only then told that the branch carries unfinished work.

Having asked once, the run SHALL proceed to completion without a further interruption. The answer
SHALL never be remembered between runs and never inferred from anything else.

Both routes that commit SHALL produce **two** commits: the implementation first, then the change's
planning artifacts and session records. Neither route SHALL mix the two.

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

#### Scenario: The gate precedes the question

- **WHEN** run 1 starts for a change carrying unfinished work
- **THEN** the unfinished-work breakdown is shown and answered before the landing question appears

#### Scenario: Implementation and planning artifacts land in separate commits

- **WHEN** any committing route runs
- **THEN** one commit carries the implementation and a second carries the planning artifacts and
  preserved session records

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

The change's **remote** branch SHALL be deleted as well, without a further prompt. Run 2 has already
proved the branch is an ancestor of the base branch, so its commits are in the base branch and
nothing can be lost. A remote branch that is already absent — deleted by the forge on merge — SHALL
be treated as success. The remote deletion SHALL be reported either way.

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

#### Scenario: The remote branch is deleted too

- **WHEN** run 2 completes its local cleanup for a change whose branch was pushed
- **THEN** the remote branch is deleted and the deletion is reported

#### Scenario: An already-deleted remote branch is not an error

- **WHEN** the forge deleted the head branch on merge
- **THEN** the absent remote branch is treated as success and reported as already gone
