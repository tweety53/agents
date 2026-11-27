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
exactly how a remote branch goes unremoved.

Run 2 is terminal, and before step 10 merged its own pull request with `--delete-branch` that pull
request outlived the run, so no later run existed to delete the branch it was
opened from . The archive branch itself is gone: run 1 archives on the change branch, which run 2 removes (**one task, one
worktree**, `docs/rationale/flow-contracts/finish-contract.md`).

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

The terminal
state file keeps `artifactUrl` indefinitely, so deleting the only source that could republish that
URL would leave it advertised and unrepublishable.
