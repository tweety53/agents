# myflow-base-branch-resolution

## Purpose

How the base branch is resolved for every step of the pipeline that needs it: one guard owning the
fetch, both resolution paths, the name validation and the three assertions; its stdout/stderr and
exit contract; the rule that every consumer invokes it rather than re-deriving it; and which
directory a consumer passes it.

## Requirements

### Requirement: One guard owns base-branch resolution

The repository SHALL carry exactly one implementation of base-branch resolution:
`resolve-base-branch.sh`, invoked by basename per **Guard resolution**
(`skills/myflow-contracts/pipeline.md`). It SHALL own the wrapped fetch, both resolution paths
(`refs/remotes/origin/HEAD`, then `git remote show origin`), the name validation, and all three
assertions.

No skill, contract, or other guard SHALL carry a second copy of that sequence, whether as executable
code or as a fenced block a run is expected to retype. Prose explaining **why** a rule exists is not
a second copy and SHALL be kept.

#### Scenario: The finish contract states the invocation, not the sequence

- **WHEN** `skills/myflow-contracts/finish-contract.md` tells a run to resolve the base branch
- **THEN** it names `resolve-base-branch.sh` in an invoking position and states its exit contract
- **AND** it carries no fenced block that performs the fetch, either resolution path, or any of the
  three assertions

#### Scenario: The reasoning survives the block's removal

- **WHEN** the fenced block is replaced by the invocation
- **THEN** the `HEAD@{upstream}` warning and the no-remote message remain in the contract, each
  naming the guard as the place the rule is enforced

#### Scenario: The preflight guard still takes the base ref as an argument

- **WHEN** `check-finish-preflight.sh` runs
- **THEN** it takes `<base-ref>` as its second argument and resolves no base branch itself
- **AND** its header comment names `resolve-base-branch.sh` as the one owner of that resolution

### Requirement: The resolver prints a bare name and separates its refusals by exit code

`resolve-base-branch.sh <dir>` SHALL take exactly one argument, the directory to resolve in.

On success it SHALL write the base branch name to stdout and nothing else, so a caller may use it
directly in a command substitution. On any refusal it SHALL write the reason to stderr and
**nothing to stdout** — an empty stdout SHALL never be readable as a resolved base.

| Exit | Meaning |
|------|---------|
| `0` | Resolved; the name is on stdout |
| `1` | A named refusal — detached `HEAD`, no base resolved, base equal to the current branch, or a base name that fails validation |
| `2` | Cannot answer — the argument is missing, `<dir>` is absent, unreadable, or not a git worktree, or `HEAD`'s own ref cannot be read |
| `3` | The repository has no `origin` remote at all |

Exit `3` SHALL be distinct from exit `2` so a caller can print the finish contract's no-remote
message — *"this repository has no remote, so there is nothing to push to or merge into"* — rather
than a base-branch failure, which sends the operator debugging the wrong thing. A caller SHALL NOT
distinguish the two by matching stderr text.

#### Scenario: A resolvable repository prints the name alone

- **WHEN** the resolver runs in a worktree whose `origin/HEAD` names `main`
- **THEN** stdout is exactly `main`, stderr is empty, and the exit status is `0`

#### Scenario: The fallback path answers when `origin/HEAD` is absent

- **WHEN** `refs/remotes/origin/HEAD` does not resolve but `git remote show origin` reports a
  `HEAD branch:` line
- **THEN** that branch is printed and the exit status is `0`

#### Scenario: A refusal leaves stdout empty

- **WHEN** any refusal occurs
- **THEN** stdout is empty, the reason is on stderr, and the exit status is `1`, `2` or `3` per the
  table above

#### Scenario: A repository with no remote is distinguishable

- **WHEN** the resolver runs in a git worktree with no `origin` remote configured
- **THEN** the exit status is `3` and the message names the missing remote

#### Scenario: An unreadable target cannot answer

- **WHEN** the argument is missing, names a path that does not exist, or names a directory that is
  not a git worktree
- **THEN** the exit status is `2`

#### Scenario: An unreadable `HEAD` ref is distinguished from a detached `HEAD`

- **WHEN** `HEAD`'s own ref cannot be read — corrupt, or permission-denied — so the current-branch
  lookup fails rather than succeeding with empty output
- **THEN** the exit status is `2`, not `1`, and the message names the unreadable ref rather than
  reporting a detached `HEAD`

### Requirement: The three assertions are unconditional and ordered

The resolver SHALL assert, in this order: the current branch is not detached; the resolved base is
non-empty; and the resolved base differs from the current branch. Each failure SHALL be reported by
name.

The third assertion is what makes the `HEAD@{upstream}` class of misresolution impossible rather
than merely unlikely, and it SHALL NOT be skippable — by a flag, an environment variable, or any
other mode.

The resolved name SHALL be validated before it is printed: it arrives from the remote and flows into
refs and git commands, so a value that is empty, begins with `-`, or carries a character outside
`[A-Za-z0-9._/-]` SHALL be refused rather than printed.

#### Scenario: A detached HEAD is refused

- **WHEN** the resolver runs in a worktree with a detached `HEAD`
- **THEN** it refuses with exit `1` and names the detached `HEAD`

#### Scenario: A base equal to the current branch is refused

- **WHEN** resolution returns the branch that is currently checked out
- **THEN** it refuses with exit `1` and names the self-comparison

#### Scenario: A hostile branch name is refused rather than printed

- **WHEN** the remote reports a `HEAD branch:` value beginning with `-` or carrying a control
  character
- **THEN** the resolver refuses with exit `1` and prints nothing to stdout

### Requirement: Every consumer invokes the resolver against the apply worktree

Every place in the pipeline that needs the base branch SHALL obtain it from the resolver rather than
re-deriving it: the preflight's `<base-ref>` argument, `/myflow-finish` run 1's merge-and-push
route, run 2's archive commit, and `/myflow-status`'s merge-status ancestor test.

Because the third assertion is unconditional, every consumer SHALL pass an **apply worktree** as the
resolver's directory argument — where `HEAD` is the change's own branch by construction — and never
the main checkout, where the base branch may legitimately be the one checked out. `/myflow-finish`
run 2 SHALL resolve the base before worktree cleanup removes that worktree.

**A resolved base belongs to the worktree it was resolved in, and SHALL NOT be reused across
worktrees.** A change may affect more than one repository, and two repositories need not share a
base-branch name. Any step that iterates a change's resolved worktree set SHALL therefore resolve the
base once per worktree, inside that iteration — never once, before the loop, from whichever worktree
came first.

#### Scenario: Run 2 resolves before it cleans up

- **WHEN** `/myflow-finish` run 2 needs the base branch to commit the archive on
- **THEN** it resolves it from the apply worktree, which cleanup has not yet removed

#### Scenario: A multi-repo change resolves a base per worktree

- **WHEN** a cleanup check runs for each worktree in a change whose repositories have differently
  named base branches
- **THEN** each worktree's check tests against the base resolved in that worktree, and a worktree
  merged into its own base is never reported unmerged because another worktree's base was used

#### Scenario: Status resolves in the worktree it is visiting

- **WHEN** `/myflow-status` answers merge status for a worktree in the resolved set
- **THEN** it obtains that worktree's base branch from the resolver rather than deriving it

### Requirement: A resolver refusal never blocks `/myflow-status`

`/myflow-status` is read-only and never blocks. A non-zero exit from the resolver SHALL make that
worktree's merge status **inconclusive** — the same disposition its existing steps give a git
failure or an unresolvable base ref — and SHALL NOT stop the report. The refusal's own message SHALL
be what the detail view reports.

#### Scenario: A refusal reads as inconclusive

- **WHEN** the resolver refuses for a worktree `/myflow-status` is reporting on
- **THEN** that worktree's merge status is inconclusive, the report continues, and the detail view
  carries the resolver's message

#### Scenario: A refusal is never read as merged or as not merged

- **WHEN** the resolver refuses
- **THEN** the merge status is neither *merged* nor *not merged*, and no ancestor test is run for
  that worktree

### Requirement: `/myflow-status` ships the guard it invokes

Invoking a guard makes `/myflow-status` subject to the same rules every other command skill follows,
per **A guard a command invokes SHALL be reachable from that command's installed skill**
(`openspec/specs/myflow-contract-distribution/spec.md`). Its `SKILL.md` SHALL carry a real
**Check guard presence.** paragraph naming `resolve-base-branch.sh`, and
`skills/myflow-status/scripts/` SHALL exist and carry a relative symlink to it.

Every statement in the repository asserting that `/myflow-status` invokes no guard or carries no
`scripts/` directory SHALL be corrected — including `check-guard-symlinks.sh`'s
`declare_if_present "myflow-status"` call, whose declared zero would otherwise be false once the
skill's required set is non-empty.

#### Scenario: The guard-symlink check computes a non-empty required set for status

- **WHEN** `check-guard-symlinks.sh` runs against this repository
- **THEN** `myflow-status` appears in its coverage report with a non-zero required-guard count and
  no `(declared)` marker

#### Scenario: The union skill carries the guard too

- **WHEN** `/myflow-fast`'s delegating presence paragraph is resolved
- **THEN** `resolve-base-branch.sh` is among the guards required in `skills/myflow-fast/scripts/`,
  and the symlink is present there
