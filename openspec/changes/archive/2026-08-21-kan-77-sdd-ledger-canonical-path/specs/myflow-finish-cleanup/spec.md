## ADDED Requirements

### Requirement: The SDD ledger has one canonical path, stated once

The SDD ledger SHALL live at exactly one path inside a change's worktree:
`<abs-worktree>/.superpowers/sdd/tasks/progress.md`.

That path SHALL be the one `superpowers:subagent-driven-development`'s own workspace script
produces — it derives `<repo-root>/.superpowers/sdd/<plan-basename>/` from the plan file, and under
myflow the plan is always `tasks.md`. myflow SHALL NOT declare a second, flat path of its own
alongside it, because the upstream skill creates the plan-scoped directory on every run regardless,
and two ledgers in one workspace reproduce the ambiguity this requirement exists to remove.

Three components SHALL state that path, and SHALL state it by citing one source of record rather
than by each carrying its own copy:

- **the writer** — the skill that dispatches implementers SHALL name the ledger's path where it
  invokes the upstream skill, alongside the flat `.superpowers/sdd/` paths it already names for the
  dispatch-context bundle, the panel record and the per-task diffs;
- **the registry** — the Temporary artifacts registry's `SDD ledger` row SHALL name the **file**,
  not the directory that contains it;
- **the preservation step** — the script that copies the record into the repository SHALL cite the
  registry row for the path it reads.

A registry row naming a directory SHALL NOT be treated as stating the path. A directory admits every
filename under it, so a row of that shape settles nothing and leaves each run free to choose.

#### Scenario: The writer names the ledger where it names the other artifacts

- **WHEN** a reader of the implementation skill looks for where the ledger is written
- **THEN** the canonical path is named there, next to the other `.superpowers/sdd/` artifacts
- **AND** no second, flat ledger path is offered as an alternative

#### Scenario: The registry names a file, not a directory

- **WHEN** the Temporary artifacts registry's `SDD ledger` row is read
- **THEN** its Location cell names `<abs-worktree>/.superpowers/sdd/tasks/progress.md`
- **AND** a reader can tell from that row alone which file the preservation step will look for

### Requirement: A record found at a known-wrong path is rescued, not skipped

When a session record is absent from its canonical path, the preservation step SHALL search an
**ordered allowlist of known-wrong paths** for that record, and SHALL copy the first one that exists
to the canonical destination.

The allowlist SHALL be a list of literal relative paths, and the match SHALL be by name. It SHALL
NOT be a glob over the workspace minus a denylist of the pipeline's own artifacts: such a denylist
must track the writer's artifact list, and the two drift the moment an artifact is added —
reintroducing this defect class in a new place. It SHALL NOT be a content check on the record's first
line either: a record written without the expected header is exactly the case a rescue is for.

A rescued file SHALL pass **every** path protection an ordinary source passes — source-root
containment, destination-root containment, and acting through the resolved path rather than the
argument. A fallback path resolving outside its root SHALL be **refused**, reported on stderr and
non-zero for that source, and SHALL NOT be reported as an absent source. This restates no new
protection; it requires that the existing ones apply to the new source.

The rescue SHALL apply to the SDD ledger and to the review panel record. It SHALL NOT apply to the
proposal artifact source: `/myflow-fast` publishes no proposal artifact by design, so a rescue and
its absent-record outcome would fire on every fast run, and an alarm that fires on the most-used
command is one nobody reads.

#### Scenario: A ledger written to a flat path is rescued

- **WHEN** a change's ledger was written to `.superpowers/sdd/ledger.md` rather than the canonical
  path
- **THEN** it is copied to the canonical destination in the repository
- **AND** the output names the wrong path it was found at, so the writer's defect is visible

#### Scenario: The allowlist order decides

- **WHEN** two of a source's fallback paths both exist in the worktree
- **THEN** the one earlier in the allowlist is the one rescued
- **AND** the choice does not depend on filesystem order

#### Scenario: A rescued source outside the worktree is refused, not skipped

- **WHEN** a fallback path is a symlink resolving outside the worktree
- **THEN** that source is reported on stderr and fails, distinctly from an absent source
- **AND** the remaining records are still preserved

#### Scenario: The proposal artifact keeps the plain skip

- **WHEN** a change has no proposal artifact source, as every `/myflow-fast` change does
- **THEN** the absence is reported as an ordinary skip
- **AND** no rescue is attempted and no absent-record alarm is raised

### Requirement: A record absent from every known path is reported as missing, not skipped

When a record subject to the rescue above is found at neither its canonical path nor any path in its
allowlist, the preservation step SHALL report it with an outcome **distinct from the ordinary skip**,
naming every path it tried.

That outcome SHALL exit zero. Non-zero SHALL keep its single existing meaning — a copy was attempted
and was refused or could not be written. A source that was never there is not that, and overloading
the non-zero column would make a caller unable to tell the two apart. Visibility SHALL be bought with
a distinct outcome word instead.

The outcome table governing the preservation step's results SHALL carry a row for the rescue and a
row for the missing record, each stating what the outcome means and what the caller does. For both,
the caller SHALL report the outcome and **proceed with the integration**: a preservation step able to
block an integration is a worse failure than the gap it closes.

The ordinary skip SHALL remain the outcome for a source that legitimately has none, and SHALL remain
the only outcome meaning **nothing was wrong and nothing need be done**. It SHALL NOT be described as
the only outcome that is *silent*: `preserved:`, `rescued:` and `MISSING:` all print to stdout and
exit zero too, so silence does not distinguish them. What distinguishes the skip is that it alone
asks nothing of the caller.

#### Scenario: No ledger anywhere is reported as missing

- **WHEN** a change's ledger is at neither the canonical path nor any allowlisted path
- **THEN** the output distinguishes this from an ordinary skip and names every path tried
- **AND** the step exits zero and the integration proceeds

#### Scenario: A caller can still tell a failed copy from an absent record

- **WHEN** a caller inspects the preservation step's exit status
- **THEN** non-zero means only that a copy was attempted and refused or failed
- **AND** neither a rescue nor a missing record produces a non-zero exit

## MODIFIED Requirements

### Requirement: Session records are preserved in the repository

The SDD ledger and the review panel record SHALL be copied out of the change's gitignored working
directory and into the repository, so that they survive the worktree's removal.

The copy SHALL happen during `/myflow-finish` run 1, before it stages the work, so the records land
in the same commit as the implementation they describe and reach the base branch with it. It SHALL
also happen on `/myflow-do`'s commit path — the case where a pull request already exists and a fix is
committed and pushed — so a fix round raised after integration refreshes the records rather than
leaving them stale.

The destination path for a change SHALL be fixed at the first copy and reused on every later copy,
so repeated runs overwrite in place rather than accumulating one dated file per round. A record
recovered by the rescue above SHALL be written to that same canonical destination, so a rescue and an
ordinary copy are indistinguishable to every later reader of the repository.

A source that does not exist SHALL be reported and skipped. It SHALL NOT fail the run: a change may
legitimately have no panel record, and a preservation step able to block an integration would be a
worse failure than the gap it closes. For a record subject to the rescue, "does not exist" SHALL mean
absent from the canonical path **and** from every path in its allowlist, and that case SHALL be
reported as missing rather than skipped, per **A record absent from every known path is reported as
missing, not skipped**.

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

The implementation skill SHALL assert the ledger's presence at its canonical path before it hands
off, and SHALL report plainly when it is absent. That assertion SHALL NOT gate the run: it exists so
that a run producing no ledger says so at the point the record must exist, rather than leaving the
absence to be discovered at preservation time or not at all.

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

#### Scenario: A rescued record reuses the canonical destination

- **WHEN** a ledger is rescued from a wrong path for a change that already has a preserved ledger
- **THEN** the existing preserved file is overwritten in place
- **AND** no second dated copy is created

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

#### Scenario: The handoff says the ledger is absent

- **WHEN** a run reaches its handoff with no ledger at the canonical path
- **THEN** the absence is reported at the point it is checked
- **AND** the run is not blocked by that report
