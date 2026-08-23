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

The unfinished-work gate SHALL run **before** this question. The operator SHALL NOT be asked how to
land a branch and only then told that the branch carries unfinished work.

Having asked once, the run SHALL proceed to completion without a further interruption. The answer
SHALL never be remembered between runs and never inferred from anything else.

Before committing anything, every route SHALL first reshape the branch: `git reset --soft` to the
recorded merge base, collapsing every per-task and fixup commit `/myflow-do` made back into the
working tree. Both routes that commit SHALL then produce **two** commits from that reshaped state:
the implementation first, then the change's planning artifacts and session records. Neither route
SHALL mix the two, and neither SHALL leave any of `/myflow-do`'s per-task commits on the final
branch.

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

#### Scenario: Per-task commits are reshaped away before committing

- **WHEN** `/myflow-do` left the branch carrying multiple per-task and fixup commits
- **THEN** run 1's `git reset --soft` to the recorded merge base collapses them into the working
  tree before the implementation commit is made
- **AND** the final branch carries no trace of the individual task commits — only the two commits
  this requirement produces

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

After the delta-spec sync and the archive move, `/myflow-finish` SHALL commit the resulting changes
on the archive branch `chore/archive-<name>` that the positioning step put the main checkout on, and
SHALL NOT push at this point — the push is the final step of run 2, per **Requirement: Run 2 pushes
the archive branch and opens its pull request**.

There is no merge to do: the change branch was already merged, which the verification step proved.
Run 2 SHALL NOT merge anything into the base branch, and SHALL NOT commit the archive on the base
branch itself.

Every commit run 2 makes after the positioning step — the archive commit and the self-review report
alike — SHALL be made on `chore/archive-<name>`, and SHALL assert that branch rather than assume it:
naming the directory with `git -C <main-checkout>` fixes the directory and not the branch.

A finished change SHALL NOT leave the archive move uncommitted in the working tree.

#### Scenario: The archive move is committed on the archive branch

- **WHEN** `/myflow-finish` completes run 2 successfully
- **THEN** `git status` in the main checkout shows no pending archive changes, and the archive
  commit is present on `chore/archive-<name>` and on no other local branch

#### Scenario: Nested fix sub-changes are archived together

- **WHEN** a change has nested `<name>-fix-N` sub-changes that are not yet archived
- **THEN** they are archived in the same operation and included in the same commit

#### Scenario: A non-base branch is never merged into the base branch

- **WHEN** run 2 is invoked with the main checkout on an unrelated branch carrying unmerged work
- **THEN** that branch is never committed to and never merged into the base branch — the positioning
  step either switches away from it or refuses

### Requirement: Run 2 removes the change's worktrees and branches

After the archive is committed and pushed, `/myflow-finish` SHALL remove every worktree belonging
to the change. The set SHALL be the keys of the state file's `worktrees` map; when that is **absent
or empty**, it SHALL be found by scanning `git worktree list` in each affected repository for branch
`openspec/<name>`.

**An empty map SHALL be treated exactly as an absent one, and never as "there are no worktrees".**
`/myflow-finish`'s preflight verdict and its unfinished-work gate are each defined as *once per
worktree in the resolved set*, per **Resolving a change's worktrees**
(`skills/myflow-contracts/worktree-resolution.md`) — never a raw read of the map, because a map carrying zero
keys would otherwise make both pass having examined nothing — `RUN2` from
every worktree and `CLEAR` from every worktree are each vacuously true of the empty set — and run 2
would then archive a change that may still hold an unmerged worktree. State self-heal previously
rebuilt those keys and is removed by this change, so nothing repopulates the map any more.

**A resolved set that is still empty SHALL stop the run rather than pass it.** Where the map is
absent or empty *and* the scan finds no worktree on the change's branch in any affected repository,
that is a state the pipeline cannot explain and SHALL be reported to the operator, never treated as
a completed removal.

Each removal SHALL be *remove-or-move if present*, so a step whose artifact is already gone is a
success rather than an error, and run 2 stays re-entrant.

#### Scenario: The recorded map drives the removal

- **WHEN** run 2 removes worktrees for a change whose state file records two worktree paths
- **THEN** both are removed, and no repository is scanned to find them

#### Scenario: An empty map falls through to the scan

- **WHEN** run 2 runs for a change whose state file records `worktrees: {}` while a worktree on its
  branch still exists
- **THEN** the worktree is found by scanning `git worktree list` for `openspec/<name>`
- **AND** it is not treated as a change with no worktrees

#### Scenario: Neither gate passes on an empty set

- **WHEN** the preflight verdict or the unfinished-work gate runs for a change whose `worktrees` map
  is empty
- **THEN** the set is resolved by the scan before either is answered
- **AND** neither reports success on the strength of having examined no worktree

#### Scenario: An unexplainable empty result stops the run

- **WHEN** the map is empty and the scan finds no worktree on the change's branch anywhere
- **THEN** the run reports that to the operator rather than recording the removal as complete

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

The SDD ledger and the review panel record SHALL be **rendered from the store** into the repository,
so that they are readable there without a running daemon and survive the worktree's removal. They
SHALL NOT be copied out of a working directory, because neither is authored as a file in one.

The ledger render SHALL happen during `/myflow-finish` run 1, before it stages the work, so the
records land in the same commit as the implementation they describe and reach the base branch with
it. It SHALL also happen on `/myflow-do`'s commit path — the case where a pull request already exists
and a fix is committed and pushed — so a fix round raised after integration refreshes the records
rather than leaving them stale. The panel record's own render SHALL happen earlier, at panel close;
that firing point belongs to `myflow-run-record` and this requirement does not restate it.

The destination path for a change SHALL be fixed at the first render and reused on every later
render, so repeated runs overwrite in place rather than accumulating one dated file per round.

A kind the store holds no rows of SHALL be reported and SHALL NOT fail the run: a change may
legitimately have no panel record, and a step able to block an integration would be a worse failure
than the gap it closes. That report SHALL be distinguishable from every failure outcome, and — unlike
the file-copy step this replaces — it SHALL mean unambiguously that no record exists, never that one
exists somewhere that was not read.

Every destination the render writes to SHALL be required to resolve inside the repository root. A
path resolving outside it SHALL be refused: reported as a failure of that one render, distinctly from
the report of an empty record, while the remaining records are still rendered. The refusal exists
because the render runs automatically and its result is committed and pushed, so a planted symlink at
a destination would otherwise write an arbitrary file under the repository's name.

A change name SHALL be required to be a single plain path component before any directory is touched.
A name carrying a path separator or a glob metacharacter SHALL be rejected outright, because the
search for an already-rendered file is anchored on the name and a metacharacter in it would let one
change adopt and overwrite another change's rendered record.

Records other than the ledger and the panel record SHALL NOT be rendered. Per-task diffs duplicate
commits already present in git history.

Rendering SHALL NOT change what the ledger may contain. A dispatch whose model the dispatcher could
not observe SHALL still record `unknown (agent-defined)`; durability SHALL NOT be a reason to fill
such an entry with a plausible value.

The script that performed the file copy, and its test harness, SHALL be removed rather than kept as a
wrapper. The pipeline stage that invoked it SHALL keep its existing stage key and its position in
run 1, so that stage runs already recorded under that key remain valid.

The implementation skill SHALL assert the ledger's presence before it hands off, and SHALL report
plainly when it is absent. That assertion SHALL NOT gate the run: it exists so that a run producing no
ledger says so at the point the record must exist, rather than leaving the absence to be discovered at
integration or not at all.

The assertion SHALL be the render's own report — the outcome word saying the store holds no dispatch
rows for this change — read at its call site, never a test for a file at a path. A path-based
assertion could only ever answer whether a file was written where the asserting skill expected it,
which is the ambiguity retiring the copy step removed.

#### Scenario: The ledger survives the change

- **WHEN** a change is integrated and later archived, and its worktree is removed
- **THEN** the SDD ledger is present in the repository at its rendered path
- **AND** a reader can still determine which model implemented each task, with no daemon running

#### Scenario: A missing source is skipped, not fatal

- **WHEN** a change has no review panel finding in the store
- **THEN** the absence is reported in the run's output as "no rows for this change"
- **AND** the integration proceeds and the remaining records are still rendered
- **AND** that report cannot also mean a record was written somewhere that was not read — which is
  what the file-copy step it replaces could never distinguish

#### Scenario: A fix round refreshes rather than duplicates

- **WHEN** a fix is committed to a branch that already has a pull request, after records were already
  rendered
- **THEN** the existing rendered files are overwritten in place
- **AND** no second dated copy is created for the same change

#### Scenario: A destination outside the worktree is refused, not followed

- **WHEN** one of the destination directories under `docs/superpowers/` resolves outside the
  repository root the render was given
- **THEN** nothing is written through it, the refusal is reported as a failure rather than as an empty
  record, and the remaining records are still rendered

#### Scenario: A source outside its own root is refused, not read

- **WHEN** the one remaining copied source — the proposal artifact beside the state file — resolves
  outside the state directory it belongs to
- **THEN** its content is not copied into the repository, the refusal is reported as a failure rather
  than as an absence, and the remaining records are still rendered
- **AND** the rendered records have no source path to escape through at all, because they are read
  from the store rather than from a file

#### Scenario: A change name that is not one plain component is rejected

- **WHEN** the render is invoked with a change name containing a path separator or a glob
  metacharacter
- **THEN** it is rejected before any directory is created or any file is written
- **AND** no other change's rendered record is read or overwritten

#### Scenario: An unobservable model stays unobserved

- **WHEN** a rendered ledger contains a slot dispatched by agent type, whose model the dispatcher
  never read
- **THEN** that entry records `unknown (agent-defined)`
- **AND** rendering does not substitute a guess

#### Scenario: The retired script is gone and nothing names it

- **WHEN** the reference and symlink guards run after the retirement
- **THEN** the copying script and its test harness are absent, and no skill, contract, guard-presence
  list or installer names either of them

#### Scenario: The handoff says the ledger is absent

- **WHEN** a run reaches its handoff having recorded no dispatch, so the store holds no ledger rows
  for the change
- **THEN** the absence is reported at the point it is checked, as the render's own outcome word
- **AND** the run is not blocked by that report

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

### Requirement: Run 1 refuses to integrate silently over unfinished work

Before it asks how the branch should land, and before any git action, `/myflow-finish` SHALL
determine whether the change carries unfinished work, by running a script rather than by asserting
it in prose. The script SHALL be run once per worktree recorded in the state file's `worktrees` map.

Two signals SHALL count as unfinished:

- unchecked items in `openspec/changes/<name>/tasks.md` and in any nested `<name>-fix-N` sub-change
- findings whose recorded status is open in `.superpowers/sdd/final-review-panel.md`

The script SHALL print exactly one verdict line — `CLEAR` when no signal fires, or `OUTSTANDING`
with a per-signal breakdown — and SHALL exit non-zero **with no verdict line** when it cannot read
the worktree. A run that receives no verdict line SHALL stop and ask the operator; it SHALL NOT read
missing output as either verdict, and SHALL check the exit code as well as the line.

A file it cannot find SHALL count as outstanding rather than as clear. Treating silence as clearance
is the failure this requirement exists to prevent.

On `OUTSTANDING` the run SHALL show the breakdown and offer the operator exactly three courses:
continue and integrate anyway; stop so the work can be finished; or file a tracker issue carrying
the outstanding items and then continue. Filing SHALL follow the labelling rules for issues the
pipeline creates.

Choosing to stop SHALL leave the change at `IN_PROGRESS` with no git action performed.

#### Scenario: An unticked plan stops the run before it asks how to land

- **WHEN** run 1 begins for a change whose `tasks.md` has unticked boxes
- **THEN** the breakdown is shown and the three courses are offered before the landing question is
  asked, and before any git command runs

#### Scenario: A clean change is not interrupted

- **WHEN** every signal is clear
- **THEN** run 1 proceeds directly to the landing question with no extra prompt

#### Scenario: An unreadable worktree stops the run

- **WHEN** the script exits non-zero without printing a verdict line
- **THEN** the run stops and asks the operator, and integrates nothing

#### Scenario: A missing plan counts as outstanding

- **WHEN** the change's `openspec/changes/<name>/tasks.md` cannot be found
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

### Requirement: Run 2 invokes self-review after writing FINISHED

Immediately after `FINISHED` is written, run 2 SHALL invoke self-review as step 9. Self-review's own
behavior — the skip prompt, context gathering, the reasoning pass, per-finding Jira filing, the
operator rating, the report and the handoff line — is defined in full by the `myflow-self-review`
capability. This requirement states only the invocation point: after
`FINISHED`, and before the archive branch is pushed.

Self-review SHALL NOT be able to prevent the `FINISHED` write, since it never runs before that write
succeeds, and a failure inside it SHALL NOT move the change off `FINISHED`.

The positioning step shifted every
later run-2 step by one, and self-review is no longer the last thing run 2 does — the archive push
follows it, so that the report and the archive land in one pull request.

#### Scenario: Step 9 follows the FINISHED write

- **WHEN** run 2 completes step 8 and `FINISHED` is written
- **THEN** step 9 (self-review, per the `myflow-self-review` capability) runs before the archive
  branch is pushed

#### Scenario: A run 2 that never reaches FINISHED never runs self-review

- **WHEN** run 2 stops at step 7 on a cleanup leftover
- **THEN** step 9 does not run, and the change remains `IN_PROGRESS`

### Requirement: Run 2 positions the main checkout before it archives

Before the delta-spec sync and the archive move, `/myflow-finish` run 2 SHALL put the main checkout
on an archive branch named `chore/archive-<name>`, cut from a base branch that has been
fast-forwarded to `origin/<base>`. `<base>` SHALL be whatever `resolve-base-branch.sh` printed for
the apply worktree, never a guess and never derived from the currently checked-out branch.

The positioning SHALL be performed by a guard,
`prepare-archive-branch.sh <main-checkout> <base> <archive-branch>`, rather than by prose, for the
same reason `check-finish-preflight.sh` owns the run-1-versus-run-2 verdict: the decision sits in
front of an irreversible step. Its exit codes SHALL be:

| Exit | Meaning |
|------|---------|
| `0` | the main checkout is on `<archive-branch>`, cut from an up-to-date `<base>`; one line on stdout names the branch it started from and the branch it is now on |
| `1` | a named refusal — a dirty working tree, **on `<base>` or off it**, a detached `HEAD`, or an existing `<archive-branch>` that is not descended from `origin/<base>` |
| `2` | `<main-checkout>` is missing, unreadable, or not a git worktree |
| `3` | `<base>` cannot be fast-forwarded to `origin/<base>` — the local branch has diverged |

A main checkout that is already on `<base>` with a **clean** working tree SHALL be fast-forwarded and
used. A main checkout on any other branch with a **clean** working tree SHALL be switched to `<base>`
and fast-forwarded.

**A dirty working tree SHALL be refused wherever it is found — on `<base>` as well as off it.** The
refusal SHALL name both the branch found and the branch required. Refusing only the off-`<base>` case
would let uncommitted changes ride onto the archive branch unremarked, and would leave the
fast-forward to fail on its own terms rather than at a gate that says why.

An existing `<archive-branch>` that is descended from `origin/<base>` SHALL be reused rather than
refused, so that a run 2 which stopped after this step can be re-run.

Anything but exit `0` SHALL stop run 2 before the archive move, with nothing staged, committed,
pushed or removed.

**When the script is absent** — a harness whose repository does not carry it — the same checks SHALL
be performed by hand, in the same order, and the handoff SHALL say the positioning was done
manually. The check is never skipped for want of the script.

#### Scenario: The checkout is already on the base branch

- **WHEN** run 2 starts with the main checkout on `<base>` and a clean working tree
- **THEN** `<base>` is fast-forwarded to `origin/<base>`, `chore/archive-<name>` is created from it,
  and the archive move is staged there

#### Scenario: The checkout is on the base branch and dirty

- **WHEN** run 2 starts with the main checkout on `<base>` and uncommitted changes
- **THEN** run 2 stops with a refusal, exactly as it does for a dirty tree on any other branch, and
  nothing is staged, committed, pushed or removed

#### Scenario: The checkout is elsewhere and clean

- **WHEN** run 2 starts with the main checkout on an unrelated branch and a clean working tree
- **THEN** the checkout is switched to `<base>`, fast-forwarded, and put on `chore/archive-<name>`

#### Scenario: The checkout is elsewhere and dirty

- **WHEN** run 2 starts with the main checkout on an unrelated branch and uncommitted changes
- **THEN** run 2 stops with a refusal naming the branch found and `<base>`, and nothing is staged,
  committed, pushed or removed

#### Scenario: A re-run reuses the existing archive branch

- **WHEN** run 2 is re-run and `chore/archive-<name>` already exists and is descended from
  `origin/<base>`
- **THEN** it is reused rather than refused, and run 2 continues

### Requirement: Run 2 pushes the archive branch and opens its pull request

As its final step, after `FINISHED` is written and after self-review has committed its report onto
the archive branch, run 2 SHALL push `chore/archive-<name>` and open a pull request against
`<base>`. It SHALL open the pull request via a PR CLI when one is usable for the host, and otherwise
SHALL print the forge's create-PR URL and ask whether it was opened — the same shape run 1's
pull-request route already uses.

Run 2 SHALL NOT push to `<base>` at any point.

The push SHALL carry both the archive commit and the self-review report, so there is no window in
which the archive pull request is merged while the report is still unwritten.

A failed push, or a failed pull-request creation, SHALL be reported with the underlying command's
own output and SHALL NOT move the change off `FINISHED`. The handoff SHALL then name the unpushed
archive branch and print the exact commands to complete the landing by hand, and SHALL NOT report
the archive as landed.

After the push, run 2 SHALL restore the main checkout to `<base>`.

#### Scenario: The archive lands via a pull request

- **WHEN** run 2 completes successfully
- **THEN** `chore/archive-<name>` is pushed, a pull request against `<base>` is open, that branch
  carries both the archive commit and the self-review report, and nothing was pushed to `<base>`

#### Scenario: The push fails after FINISHED

- **WHEN** the archive branch cannot be pushed
- **THEN** the failure is reported with git's own output, the change stays `FINISHED`, and the
  handoff names the local archive branch and the commands needed to land it

#### Scenario: The checkout is returned to the base branch

- **WHEN** run 2 finishes, successfully or with a failed push
- **THEN** the main checkout is left on `<base>`

### Requirement: Worktree cleanup verifies the stack actually stopped, before removing anything

`/myflow-finish` run 2 SHALL NOT remove a worktree while a live process is running from it.

After running the project's `## stop` command, and before any removal, run 2 SHALL check each
worktree in the resolved set for processes whose working directory is at or under that worktree's
path. This check SHALL run whether or not the project declares a `## stop` command, and whether or
not that command reported success: a project that declares no stop command can still have a stack
running from its worktree, and a stop command that exits 0 is not evidence that it worked — which
is the whole of what this requirement adds.

The check SHALL be answered from the operating system's own process bookkeeping, never from a file
the project wrote. The records a project keeps about what it started live inside the worktree being
removed, so they are exactly the evidence that a removal destroys.

The check SHALL be project-agnostic. It SHALL NOT read a port list, invoke a build tool, or require
a project to declare anything, so that it protects every project myflow is installed into rather
than only those that opt in.

**A working directory, not a port, SHALL be what identifies the process.** A port proves nothing: a
live workspace legitimately holds one. What a process launched from a worktree has, and what ties it
to the directory about to be removed, is its working directory.

#### Scenario: A live process blocks the removal

- **WHEN** run 2 finds a process whose working directory is at or under a worktree it is about to remove
- **THEN** the check fails, **no** worktree is removed, run 2 stops at `IN_PROGRESS`, and it reports
  each process's pid, its working directory, and the command that reaches it

#### Scenario: A clean worktree proceeds

- **WHEN** no process has a working directory at or under any worktree in the resolved set
- **THEN** the check passes and cleanup proceeds to removal

#### Scenario: The check cannot be answered

- **WHEN** the check cannot run at all — the scanning tool is absent, or a worktree path cannot be read
- **THEN** it is treated as a failed check, not a passed one: no worktree is removed and the reason is
  reported

#### Scenario: The stop command reported success but did not work

- **WHEN** the project's `## stop` command exits 0 and a process from the worktree is still alive
- **THEN** the check fails on the live process, and the stop command's exit code SHALL NOT be read as
  evidence that the stack stopped

### Requirement: A blocked removal is a failure, never a confirmable disclosure

A live process found by the check above SHALL fail the check outright. Run 2 SHALL NOT offer the
operator a prompt to proceed with the removal anyway.

This is deliberately unlike the ignored-files disclosure, which reports what `--force` will destroy
and asks. That disclosure is safe to confirm because the operator can see what is at stake and
decide. A live process is different in kind: confirming it destroys the only records that can reach
the process afterwards, and the resulting orphan holds ports shared across every workspace. Both
occurrences of this incident reached the operator as something to click through.

`/myflow-fast`'s override of the ignored-files ask SHALL NOT extend to this check. That override is
justified by the records worth keeping already being out of the worktree; a live process is not a
preserved record, and an unattended run is the case most likely to orphan one.

#### Scenario: No override is offered

- **WHEN** the check finds a live process, under `/myflow-finish` or `/myflow-fast`
- **THEN** the run stops with no prompt to proceed, and the operator clears the process and re-runs

