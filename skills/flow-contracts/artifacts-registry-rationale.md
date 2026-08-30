# Temporary artifacts registry — rationale

This file is the reasoning behind `skills/flow-contracts/artifacts-registry.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.**

## Temporary artifacts registry

**This table is the one place a cleanup rule is stated.** Everything else that mentions a removal
points here rather than restating it, in this file and in every skill. **Worktree cleanup**
(`skills/flow-contracts/finish-contract-run2.md`) is the *procedure* for the rows removed there and
not a second statement of the rule: the table says what is removed and when, that section says how
— and a stale second copy of a rule governing `git worktree remove --force` is a copy that deletes
the wrong thing.

**An artifact no row accounts for is a defect in the registry**, corrected by adding the row. It is
never left unaccounted for on the grounds that something probably removes it — that assumption is
exactly how the remote branch went unremoved until it was given a row here.

Run 2 is terminal
and the pull request it opens outlives the run, so no later run exists to delete the branch it was
opened from — this repository already carries five such leftovers, chore/archive-kan-197,
chore/archive-kan-200, chore/archive-kan-209, chore/self-review-kan-201 and chore/self-review-kan-236,
which is the evidence, not a guess, that nothing removes them today. Whether some future run should
gain that duty is design.md's open question `archive-branch-cleanup`, deliberately left open rather
than decided here.

**This is the one row whose removal is verified by asking rather than by looking**, and the reason
is that "ran the removal" is not "verified gone": a removal that reported success against a stale
connection leaves this row's promise broken with nothing having failed. So a survivor is established
from the project's own survivor report, never inferred from the removal's exit code — stated once
under **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`), with the report's
output and exit-code contract under
**Project configuration** (`skills/flow-contracts/project-configuration.md`). A report that could
not reach its service is skipped rather than failed, so this is the one row a stopped service leaves
unverified without stranding an already-merged change; that asymmetry is likewise
**Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`).

**Nothing removes the claimed cache index, and nothing in this pipeline can — which is why the row
says so instead of naming a remover it does not have.** The rule one paragraph above is that an
artifact no row accounts for is a defect in the registry; the row exists for that reason, and an
honest `Removed by` cell is the whole of what it buys. The index is the one workspace value that is
**not** a function of the change name — it is claimed by probing, for the reasons under
**The cache index** (`skills/flow-contracts/workspace-isolation.md`) — and it is not written into
the state file either. So by the time run 2 runs, nothing on the machine can say which index this
change held: there is no derivation to repeat and no record to read, and a run that swept an index
it guessed would flush another workspace's. The project's `remove` command does not touch it for the
same reason, which
**Project configuration** (`skills/flow-contracts/project-configuration.md`) states as a property
of the `cache index` resource word.

**What that leaves behind is bounded, and where it stops being acceptable is named rather than
glossed.** What stays in the index is sessions and cache entries, which are disposable by
construction — the accepted cost under
**The cache index** (`skills/flow-contracts/workspace-isolation.md`) is precisely that nothing
which must survive a restart may be kept there, so leaving them costs a login and never data. The
cost that is *not* free is slot exhaustion: a cache offers sixteen indices, one of which is the
empty-id default, so a probe that reads a non-empty index as taken finds fewer free slots as
finished changes accumulate, and a workspace that can claim none falls back to sharing — the failure
this contract exists to remove. The remedy is the operator flushing the cache, and it is safe for
exactly the reason the leftovers are: nothing durable is in there. A project may ship its own
command to list or flush its stale indices; that is the project's tooling, and this row does not
claim it — the `Removed by` cell stays `nothing in this pipeline` either way.

**Which rows run 2 verifies is read off this table, not listed again.** Every row whose lifetime
ends at run 2 is checked back by `<agents repo>/scripts/check-cleanup-complete.sh`, whose header explains which
rows that leaves it reading and why; step 7 of
**Run 2 — the branch is merged** (`skills/flow-contracts/finish-contract-run2.md`) is where its
verdict is acted on.

**That derivation is declared, not left implicit.** The guard carries one marker line per row of
this table saying whether it checks that row or deliberately does not, with the reason; its harness
reads this table and those markers and fails when the two disagree in either direction. So a row
added here goes nowhere until someone records a decision about it — which is what stops a future
artifact from being confirmed clean by a guard that never looked for it.

The terminal
state file keeps `artifactUrl` indefinitely, so deleting the only source that could republish that
URL would leave it advertised and unrepublishable.
