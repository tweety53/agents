## MODIFIED Requirements

### Requirement: Run 2 removes the change's worktrees and branches

After the archive is committed and pushed, `/myflow-finish` SHALL remove every worktree belonging
to the change. The set SHALL be the keys of the state file's `worktrees` map; when that is **absent
or empty**, it SHALL be found by scanning `git worktree list` in each affected repository for branch
`openspec/<name>`.

**An empty map SHALL be treated exactly as an absent one, and never as "there are no worktrees".**
`/myflow-finish`'s preflight verdict and its unfinished-work gate are each defined as *once per
worktree in the resolved set*, per **Resolving a change's worktrees**
(`skills/myflow-contracts/pipeline.md`) — never a raw read of the map, because a map carrying zero
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
