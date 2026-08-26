# Workspace isolation — rationale

This file is the reasoning behind `skills/flow-contracts/workspace-isolation.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

## The workspace id

A second copy of a derivation is worse than a second copy of a procedure — two
implementations that disagree by one byte produce two different ids, each of which looks correct on
its own.

It exists for one failure. Two apply worktrees for two different changes run their applications
against the same backing services. The loud half is a port already bound, which stops the second
worktree and is fixed in a minute. The quiet half is the one worth engineering against — one
change's migrations alter the schema the other change is testing against, so a suite that was green
an hour ago fails, and the failure looks like a real defect rather than a collision. The operator
then spends the afternoon debugging code that was never broken.

Duplicating whole services per workspace would address the
same failure one level too high — a second database server, a second cache and a second object store
per concurrent change, at several times the memory and disk, to keep apart two things that were only
ever sharing one name.

A byte is the one unit a shell and a JVM can agree on without either reaching for a
primitive whose answer depends on the ambient locale, and it is already the unit the digest uses.

Without it `tr` is free to consult the ambient locale's case table, and
it does: the same `İstanbul-test` yields `istanbul-6b8a` under `en_US.UTF-8` and `--stanbul-6b8a`
under `C`, from one script on one machine. Pinning the locale makes both `tr` calls byte-wise, which
is what step 2 says they are. The explicit `A-Z`/`a-z` ranges rather than `[:upper:]`/`[:lower:]` say
the same thing twice on purpose — a character class is exactly the construct that would start
meaning something else in another locale.

From
step 2 onward the string is pure ASCII, so `${#prefix}` and `${prefix:0:12}` count the same thing in
every locale — a C-locale shell counts bytes and a UTF-8 one counts characters, and on an ASCII
string those are the same number. Normalising last, as the obvious reading of "truncate then clean
up" would have it, leaves that arithmetic running over the raw name where the two disagree. The
order also makes `.` and `_` real segment boundaries rather than opaque characters inside one
segment, which is why `KAN-99-Fix.Thing` keeps `kan-99-fix` and not `kan-99`.

**A name already within `[a-z0-9-]` passes through step 2 unchanged**, which is why the worked id is
the same with it as without — and why an implementation that omits it looks correct against every
example anyone is likely to try, then diverges on the first change name carrying a capital or a dot.
The three extra comments above are that case, and they are what a second implementation has to
reproduce.

`echo` appends a newline, and the
digest of the name plus a newline is an unrelated value. An implementation that hashes the trailing
newline and one that does not will disagree on every id ever derived, while each looks perfectly
correct in isolation — which is precisely the drift this file being canonical is meant to prevent.
The tool is not part of the contract, only the bytes are: `sha256sum` and `openssl dgst -sha256`
yield the same digest from the same input and are equally acceptable. `shasum -a 256` is named
above because it is present on both stock macOS and Linux, so one written form runs in either
place.

The prefix is what keeps a listing of
derived resources legible — an operator reading `psql -l` can tell at a glance which change a
database belongs to, which a bare digest would not give them. The digest is what keeps two changes
apart when their prefixes collapse together, which is the likely collision rather than the exotic
one: `kan-15-parallel-lanes` and `kan-15-parallel-two` both reduce to the prefix `kan-15` and are
distinguished only by their digests, `42f2` and `77be`.

This is the point the three
  implementations disagreed on, because Unicode folding is where the freedom lives: `İ` (U+0130)
  folds to *two* characters under a JVM's `lowercase`, to `i` under a UTF-8-locale `tr`, and to
  nothing under a C-locale one.

Every value derived from such an id is still legal: the database name
  substitutes `_` for every `-` and stays an unquoted SQL identifier, and the bucket name carries a
  project prefix ahead of the id so it never begins or ends with `-`.

The digest is still taken over the original, unmodified name, so normalising can never merge two
distinct names into one id — it can only make two ids share a prefix, which the digest already
handles. That is why only the prefix is at issue here: the half of the id that carries the
correctness is locale-independent already.

**What determinism does not promise.** Four hex characters is 65 536 values, so the digest makes a
collision unlikely, not impossible. It is drawn over change names, of which a machine has tens
rather than thousands, and a collision only bites when the two colliding changes are running their
stacks *at the same time*. The failure it would then produce is exactly the one this contract
exists to remove, so it is worth saying plainly rather than glossing: if it ever bites, widening the
digest is a one-line change here and a re-derivation everywhere, and every id changes when it
happens.

## What the id derives

Every call site that
exists today may quote correctly; the rule is for the one added next year that does not. An
identifier that only works when quoted is a call site's correctness resting on a convention nobody
is checking, and the cost of removing that dependence entirely is one substitution. A dashed name
would work right up until it did not, and the failure would land in whichever tool was added last.

There is no unquoted-parsing
hazard to protect against, so a second spelling of the same id would be a second rule to remember
with nothing bought by it.

One offset for the whole block, rather than one per port, is what keeps the relationship
between a workspace's ports the same as the relationship between the project's declared ones.

A patched
block is the worst of both schemes: as unpredictable as discovery, yet indistinguishable at a glance
from the deterministic block it no longer is, so nobody can reason about which of a workspace's
ports mean what.

This is the
mitigation for the one cost this contract imposes on an operator: a worktree's ports differ from the
project's documented ones, so opening the documented URL out of habit reaches the main checkout's
applications — a different change's work, answering plausibly and about the wrong thing.

It narrows a race, it does not close one. A port observed
free can be bound by anything on the machine between the check and the bind, and no amount of
re-checking removes that window.

## The cache index

This is the load-bearing half, and it is stated
here because getting it wrong inverts the comparison the section is built on. A claim kept only in
the claiming worktree's own state is invisible to every other worktree: each run then sees an index
nobody has written to yet as free, takes the lowest such index, and two workspaces started minutes
apart collide *reliably* rather than six percent of the time.

**Deriving the index and refusing to start on a collision was considered, and is not what happens.**
It is never silently wrong, which is the property that matters — but it blocks a legitimate second
workspace outright, with renaming the change as the only remedy available to the operator. Probing
gives the same safety and costs nothing at the point of use.

A workspace's stack runs for days and nothing in this contract
is left running to renew a lease, so a lifetime short enough to clean up after a crashed run would
expire under a live one — the silent share again, by a slower route.

That is acceptable for precisely the reason it is safe to reassign in the first
place: what the cache holds is sessions and cache entries, both disposable by construction.

Run 2 has no derivation to repeat and no record to read, so it cannot know which index to sweep, and
an index swept by guess is another workspace's.

What that leaves behind and why it is acceptable are stated once
under **Temporary artifacts registry** (`skills/flow-contracts/artifacts-registry.md`), which is canonical
for every artifact's lifetime.

## The empty id

Putting the isolation behind a request was considered and
rejected on one point: the failure being prevented is the one that *forgetting* the request
reproduces, and reproduces silently — the applications come up, they answer, and only the schema is
wrong. A safeguard that has to be remembered is absent in exactly the runs that needed it, so an
opt-in guard against a silent failure has the failure as its own failure mode.

**Refusing rather than reporting-and-continuing is the same argument that makes isolation automatic
rather than opt-in**, one paragraph above: a safeguard that depends on somebody reading a report is
absent in exactly the run that needed it.

The uniform alternative — every
checkout takes an id, the main one included — has no empty-id branch to get wrong, and breaks every
existing local setup exactly once: databases orphaned, saved connections pointing at nothing,
bookmarks stale. That is a migration nobody asked for, in the one checkout an operator spends most
of their time in.

Every key in a project's configuration is optional; see **Project configuration**
(`skills/flow-contracts/project-configuration.md`), which is canonical for that file and for the
rules an isolation row resolves under.

## Creation and cleanup

Something now exists on the
machine that did not exist before, and a side effect nobody is told about is one nobody thinks to
look for later.

Copying looks faster and is wrong twice over: it carries the source
database's migration state, which need not be the state this branch's migrations expect, and it
carries whatever rows happened to be sitting there, so the workspace's database is subtly unlike a
clean one and the difference surfaces as a result nobody can reproduce. Taking the ordinary path
costs one migration run and buys a database that is exactly what a clean build would produce.

Naming a
mechanism here would fix in the canonical file a choice each project makes for itself, and the first
project to make it differently would make this file wrong.

Blocking an already-merged
change from finishing, over a service that happens to be stopped, trades a real cost — a change
stranded short of its terminal state — for a few megabytes of stale storage. It also matches how the
project-supplied stop check already treats an unreachable stack.

Telling "the service
is stopped" apart from "the survivor report is broken" would need a second exit-code convention that
every project implements identically, and at this point in the pipeline both have the same correct
response: run 2 is past the merge, and blocking an already-merged change over a service that happens
to be down is what the paragraph above forbids. The cost is real and is named rather than hidden — a
genuinely broken survivor report is indistinguishable here from a stopped service — which is why the
skip carries the command's name and its exit code instead of being noted in passing.
