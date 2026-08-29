## Context

spectre already has half of this. `spectre/peers` declares neighbouring trees as
`<name> <relative-path>`, `tree.ResolvePeer` classifies a name as found / not declared / not
present / unreadable, and citations already carry a peer scope — `(@web:plans#R4)` — which
`validate` and `refs` resolve. What is missing is the same idea one level down: a *change* that
spans trees.

flow already models the multi-repo case too. The state file's `worktrees` map is keyed by absolute
worktree path, one entry per affected repository, with a single scalar `branch`; **A change
spanning repositories is one record** (`skills/flow-contracts/state-file.md`) is explicit that a
two-repository change is one record. What it does not carry is an ordering, and no repository but
the canonical one carries anything at all.

The guards are what break, and they break in the same place: each resolves
`<worktree>/<spec-root>/changes/` per worktree, finds nothing in a secondary one, and reports that
absence as a verdict — `OUTSTANDING` for a plan that will never be there, or exit 2 with no verdict
at all.

This change is itself cross-repo: `spectre` gains the tree format and the command, `agents` gains
the guard and contract changes that call into it.

## Decisions

### What lives in the second repo

**ID:** pointer-tree-not-full-tree
**Status:** active
**Chosen:** A pointer tree — `<spec-root>/changes/<id>/link.md` and nothing else in that directory —
because the plan stays in exactly one place and cannot drift, while the second repo still answers
"what is this branch part of" to someone who cloned only it.
**Considered:** A full `spectre new` scaffold per repo, rejected because the task list would then
exist in two places with nothing keeping them equal, and `validate` could not tell a satellite from
a change whose plan is merely missing. Leaving the tree in one repo and passing the canonical
worktree path to every guard, rejected because it is the cheapest change but does not meet the
requirement at all — the second repo still carries no record.

### Who owns the link mechanism

**ID:** spectre-owns-the-link
**Status:** active
**Chosen:** spectre. The reference becomes a property of the tree format, so `validate` and `list`
understand it and any tool reading the tree sees it, not just flow.
**Considered:** flow generating the pointer at isolate-workspace from the state file, rejected
because the pointer would then be invisible to every spectre command and unvalidated. Hand-authoring
it during writing-plans, rejected because it duplicates facts the state file already holds with
nothing keeping it correct.

### How a link names its counterpart

**ID:** link-names-a-peer
**Status:** active
**Chosen:** `<peer>:<change-id>`, where `<peer>` is a name declared in `spectre/peers`. It reuses
the resolution and the four outcomes spectre already has, stays relative to the tree's parent
directory, and commits nothing machine-specific.
**Considered:** An absolute repository path, as KAN-343's `tasks.md` records today
(`**Repository:** /Users/tweety53/Projects/gymie-frontend`), rejected because it breaks on any other
clone — which is the "clone the frontend alone" case this change exists for. A `## Part of` heading
inside a normal `proposal.md`, rejected with pointer-tree-not-full-tree above.

### Which direction the reference runs

**ID:** two-way-link
**Status:** active
**Chosen:** Both. The satellite carries `## Part of`; the canonical carries `## Parts`. Reading
either tree alone tells you every repository involved, with no peer scan and no dependence on the
other repositories being present on disk.
**Considered:** Satellite → canonical only, with the canonical's part set derived by asking each
declared peer, rejected because the derivation needs every peer checked out and present, and returns
an incomplete answer when one is not.

### What writes the canonical side

**ID:** one-command-writes-both-sides
**Status:** active
**Chosen:** `spectre link` writes the peer's side and its own in one run. Two sides that must agree
should not depend on the operator remembering a second command in a second directory.
**Considered:** A command per tree, each writing only its own root, with `validate` reporting a
one-sided link, rejected because it makes the common case two steps to prevent a failure mode that
guarded-cross-tree-write below already closes.

### How guarded the cross-tree write is

**ID:** guarded-cross-tree-write
**Status:** active
**Chosen:** Refuse unless the peer resolves, the canonical change exists there, the canonical is not
itself a satellite, the link does not already exist, and the peer's own change directory has no
uncommitted modifications. `--force` overrides the uncommitted-modifications refusal alone. The peer
side is written first and the local side second, so a refused peer write leaves nothing half-created.
**Considered:** Checking only that the peer is declared and carries the change, as `archive` does
when it lets `git mv`'s own error surface, rejected because this is spectre's first write *outside*
the tree it was pointed at: the failure it would allow is silently overwriting an edit in a working
tree the operator is not looking at.

### Archive skew

**ID:** independent-archive
**Status:** active
**Chosen:** Each tree archives on its own timeline. A link resolves against `changes/<id>/` and then
`changes/archive/<id>/`, so it never dangles. Skew — canonical archived, part still open, or the
reverse — is a `validate` finding, not a refusal.
**Considered:** `spectre archive` in the canonical tree refusing while any part is open, rejected
because it blocks the canonical repo behind the other repo's merge, which is the friction this
change is trying to remove. Archiving the parts along with the canonical, rejected because it is a
`git mv` inside another repository's working tree that can fail halfway and leave exactly the split
state it was meant to prevent.

### The branch is recorded, and recorded strictly

**ID:** branch-recorded-strictly
**Status:** active
**Chosen:** `## Branch` carries exactly one non-empty line, matching the git ref-name shape
`<agents repo>/scripts/resolve-base-branch.sh` already enforces — first character in
`[A-Za-z0-9._]`, every character in `[A-Za-z0-9._/-]`. Both sides of a link must carry the identical
branch name; disagreement is a finding. spectre validates the shape and the agreement, and resolves
nothing against git.
**Considered:** Omitting the branch, on the grounds that it is flow's concern and already in the
state file's `branch` field, rejected because a bare clone of the second repo cannot read flow's
state file and so still could not say which branch carries the work. Recording it as uninterpreted
free text, superseded by no-free-text-in-link-md below.

### No field in `link.md` is free text

**ID:** no-free-text-in-link-md
**Status:** active
**Chosen:** Every section has a grammar and `validate` enforces it. `## Part of` and `## Parts`
carry `<peer>:<change-id>` with the peer declared and the id matching spectre's own change-id rule;
`## Branch` is the ref-name shape above; `## Merge order` is a numbered list of `.` and peer names;
`## Tasks here` is a task-number list — comma-separated positive integers and ascending ranges
(`4-12`, `1,3,7-9`), ascending overall, no overlaps, every listed number matching a real task in the
canonical `tasks.md`.
**Considered:** Leaving `## Branch` and `## Tasks here` as uninterpreted prose spectre stores and
prints, rejected on the operator's instruction: an unvalidated field in a file the guards read is a
field that is wrong without anything saying so, and `## Tasks here` in particular is checkable
against the canonical plan, which turns it from a comment into a claim.

### Who runs `spectre link` on a real change

**ID:** flow-runs-link-automatically
**Status:** active
**Chosen:** `/flow`'s isolate-workspace step, for every worktree beyond the canonical one, from the
worktree set it already resolves. No change can gain a second repository without gaining the link,
which is what lets the guards rely on the link being there.
**Considered:** A planned task emitted by writing-plans, rejected because a plan that omits it
leaves the guards in today's state. Leaving it to the operator entirely, rejected for the same
reason.

### Where the merge order lives

**ID:** merge-order-in-canonical-link
**Status:** active
**Chosen:** A `## Merge order` section in the canonical `link.md` — a numbered list, one entry per
participating tree, in landing order, `.` naming the canonical tree itself and every other entry a
peer name. `validate` requires `.` plus every declared part to appear exactly once.
**Considered:** Using the order of the `## Parts` list with the canonical always first, rejected
because it hardcodes "canonical merges first", which happened to be right for KAN-343 and is not
right in general. A per-part rank, rejected because ranks drift and collide and `validate` then has
to police a numbering nobody maintains.

### Run 1 lands the repos in that order

**ID:** run-1-enforces-merge-order
**Status:** active
**Chosen:** `/flow`'s run 1 iterates the resolved worktree set in `## Merge order` and takes the
chosen route for each in that sequence, stopping and reporting on the first failure. The three
routes are unchanged; only their iteration order becomes defined.
**Considered:** Reporting the order and leaving the landing to the operator, rejected because that
is what happens today — the ordering would be written down and still executed by hand. Gating each
repository on the previous one's branch actually being merged into its base, rejected because it
makes run 1 wait on a pull request that may take days, turning one run into several.

### Merge atomicity is out of scope; ordering is what this change buys

**ID:** atomicity-is-ordering-only
**Status:** active
**Chosen:** Two repositories cannot be merged atomically without a forge feature neither has. What
this change delivers is a recorded, enforced *ordering*, which is what KAN-343 actually did by hand.
**Considered:** Nothing further — the alternative is a cross-forge atomic merge that does not exist.

### This change links its own trees

**ID:** dogfood-once-green
**Status:** active
**Chosen:** Once spectre's tasks are green and the binary reinstalled, `spectre link` runs for real
against this change's own two trees — `agents` canonical, `spectre` the part — and the remaining
agents-side tasks run under the mechanism they just built. Genuine end-to-end proof before the
guards depend on it.
**Considered:** Shipping without using it, letting the next cross-repo change be the first user,
rejected because the guards' new behaviour would first run in anger on someone else's change.

### A peer that is not present on disk is not a finding

**ID:** peer-absence-is-not-a-finding
**Status:** active
**Chosen:** A link check requires the peer to be **declared** in `peers` — a tree-level fact
committed with the tree, so it is true in a worktree too. It requires the counterpart change to
exist and both sides to agree **only when the peer tree is actually present on disk**; a peer that
is declared but not present is reported as not checked, never as a finding.
**Considered:** Treating not-present as a finding, the way `ResolvePeer` classifies it for citation
resolution today, rejected because `peers` paths are relative to the tree's parent directory
(`agents ../agents`) and every `/flow` run works in a worktree. From
`~/Projects/spectre-worktrees/spectre-<name>/` that path resolves to
`~/Projects/spectre-worktrees/agents`, which does not exist — the sibling worktree is under
`~/Projects/agents-worktrees/` instead. A finding there would fire on every worktree of every
linked change, which is the over-firing `<agents repo>/scripts/check-references.sh`'s header
records as the condition that manufactures suppression markers. It is also correct on its own
terms: a link legitimately outlives the other repository being checked out at all.

### The guards take the canonical worktree path, and fall back to the peer

**ID:** guards-take-the-canonical-worktree-path
**Status:** active
**Chosen:** A guard that needs the canonical `tasks.md` accepts the canonical worktree's absolute
path as an argument, supplied by `/flow` from the worktree set it has already resolved. Only when
that argument is absent does the guard fall back to resolving the link's peer name through `peers`;
when neither answers, it exits 2 — "cannot determine anything" — rather than guessing.
**Considered:** Peer resolution alone, rejected for the reason peer-absence-is-not-a-finding gives:
inside a worktree it resolves to a path that does not exist, so the guards would be back to
reporting an absence as a verdict, which is the defect this change exists to remove. Reading the
state file's `worktrees` map directly from the guard, rejected because **Resolving a change's
worktrees** (`skills/flow-contracts/worktree-resolution.md`) forbids a raw read of that map and
requires the caller to resolve the set first — the guard is the callee, so the path arrives as an
argument.

### An unreadable peer IS a finding

**ID:** unreadable-peer-is-a-finding
**Status:** active
**Chosen:** `PeerUnreadable` — the declared path exists but the tree could not be opened, or its
specs could not be read — is a finding. Only `PeerNotPresent` is quiet.
**Considered:** Folding `PeerUnreadable` in with `PeerNotPresent`, which task 2's first
implementation did. Rejected on review for two reasons. `internal/check/refs.go` already treats
`PeerUnreadable` as a real finding for the same `ResolvedPeer` type in the same package, so
silencing it here makes one package answer the same question two ways. And the two conditions are
not the same: absent is the ordinary worktree case `peer-absence-is-not-a-finding` exists for,
while unreadable means a peer tree that is there and broken — exactly the state a validator should
say out loud. This decision narrows `peer-absence-is-not-a-finding` rather than superseding it:
that decision's reasoning is about a path that legitimately does not exist, and it never reached
the unreadable case.

### `list` degrades, never aborts — and unresolved progress says so in JSON

**ID:** list-degrades-and-says-so
**Status:** active
**Chosen:** A malformed `link.md` in one change directory makes `list` render that row without its
link information and carry on; it never aborts the listing. And a row whose progress could not be
resolved carries an explicit marker in `--json`, rather than serialising as `done: 0, total: 0`.
**Considered:** Aborting on the first parse error, which task 4's first implementation did. Rejected
for two reasons found on review. It makes a read-only listing fail where it previously always
succeeded — one broken change hides every healthy one — and it reports a content problem as exit 2
when `validate` reports the identical file as exit 1, so the two commands disagree about the
repository's own exit-code contract.

The JSON half is not cosmetic, and the reason lies outside spectre: `/flow`'s implementation phase
reads `spectre list --json` and treats `total == 0` as "no plan spectre can read — stop and resume
at brainstorming". A satellite whose peer is simply not checked out resolves no progress, so an
unmarked `0` routes the pipeline back to brainstorming for a change whose plan is intact in the
canonical repository. The text renderer already distinguishes the two as `-` against `0/0`; the JSON
contract must carry the same distinction rather than being the one surface that loses it. Marking
the unresolved case — rather than omitting `done`/`total` — keeps every existing consumer working.

### `## Tasks here` is checked for validity, never for completeness

**ID:** tasks-here-is-not-complete
**Status:** active
**Chosen:** `validate` checks that every number listed in `## Tasks here` names a real task in the
canonical `tasks.md`. It does **not** check that the list names every task belonging to that
repository. A stale list that merely omits a newly added task is not a finding.
**Considered:** Checking completeness too. Rejected because spectre cannot do it without learning a
vocabulary that is not its own: which repository a task belongs to is carried by `**Files:**` and
the `**Repository:**` field, both `/flow` conventions in the `agents` repository, and spectre knows
nothing of either. Teaching spectre to read them would make a general tree format depend on one
pipeline's plan grammar — a far worse trade than an incomplete list.

**Measured, not assumed.** Task 13 ran `spectre validate` against real linked trees with
`## Tasks here` deliberately left at `1-5` after task 12 had been added to the `spectre` side, with
the peer resolvable. It reported `no findings` at exit 0. The check behaves exactly as this decision
now records, and the gap is real rather than theoretical — a satellite's task list can drift from the
plan and nothing will say so.

**What this means for a reader of a satellite tree:** `## Tasks here` is a signpost, not an
inventory. The canonical `tasks.md` is the only authority on which tasks exist and where each one
lives.

### A cross-repo task's commit subject describes what that commit did

**ID:** subject-per-repository
**Status:** active
**Chosen:** When one task lands a commit in each repository, each commit's subject names what that
commit actually did there. A shared subject is right only when both sides perform the literally same
action.
**Considered:** One subject per task, reused in every repository it touches. This change did it both
ways before noticing — task 6 wrote `chore(spectre): declare the peer and link this change's two
trees` in both repositories, and task 13 wrote two different subjects. Neither is a defect, and the
rule above is what reconciles them: task 6's two sides each declared a peer and wrote a `link.md`,
the same action mirrored, so one subject was true in both places. Task 13's sides did different
things — one flipped a peer path, the other corrected a tracking list — so a shared subject would
have been false in one repository. A subject that is false where it sits is worse than two subjects
that differ.

### `link.md` is implementation, not planning

**ID:** link-md-is-implementation
**Status:** active
**Chosen:** Carve `link.md` out of the planning exclusion by file, alongside the exemption
`<project>/spectre/specs/` already has, so it lands in a task commit rather than the integrate
phase's planning commit.
**Considered:** Following the rule as written and letting the planning commit carry it. Rejected
because in a satellite repository `link.md` is the change's **only** content, so that repository's
branch would carry no implementation commit at all, and the file that says what the branch is part
of would arrive only at integration — after the moment someone cloning that repository alone would
want it, which is the case this whole change exists for. Also considered leaving the commits as they
stand and filing a follow-up, rejected because the contract and the practice would then disagree and
the next cross-repo change would meet the same question with no answer.

**The reasoning is the one that already exempts a capability spec.** A spec is implementation
because it states what the system must do. `link.md` is implementation because it states which
repositories the change spans, which branch carries it, and in what order they land — structural
fact about the change, not the planning narrative `proposal.md`, `design.md` and `tasks.md` carry.
The boundary within the spectre tree stops being purely directory-shaped, and the one file that
makes it so is named.

**Found by this change's own commits**, not by review of the contract: tasks 6 and 13 committed
`spectre/changes/<id>/link.md` in task commits because the plan told them to, and the collision only
became visible when the staged diff was checked against **Git boundaries**
(`skills/flow-contracts/git-boundaries.md`) at the verify stage.

## Open questions

None. Every question raised during brainstorming was answered before this stage closed.
