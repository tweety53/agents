# Workspace isolation

**This file is canonical for workspace isolation.** Skills, guards, and the projects that declare
isolation reference it by name; none of them restate the derivation, the port rule, or the empty-id
promise. A second copy of a derivation is worse than a second copy of a procedure — two
implementations that disagree by one byte produce two different ids, each of which looks correct on
its own. If a rule below and a skill, a guard, or a project's own configuration ever
disagree, this file wins.

It exists for one failure. Two apply worktrees for two different changes run their applications
against the same backing services. The loud half is a port already bound, which stops the second
worktree and is fixed in a minute. The quiet half is the one worth engineering against — one
change's migrations alter the schema the other change is testing against, so a suite that was green
an hour ago fails, and the failure looks like a real defect rather than a collision. The operator
then spends the afternoon debugging code that was never broken.

**What is isolated is the logical resource, never the service that holds it.** A workspace gets its
own database inside the shared database server, its own index inside the shared cache, and its own
bucket inside the shared object store. Duplicating whole services per workspace would address the
same failure one level too high — a second database server, a second cache and a second object store
per concurrent change, at several times the memory and disk, to keep apart two things that were only
ever sharing one name.

## The workspace id

**An apply worktree has a workspace id; the main checkout has none.** The id is derived from the
change name and from nothing else. It is **deterministic** — the same change name always yields the
same id, in this session and in one a week later — so two sessions that never communicate cannot
produce the same id for two different changes, or two different ids for one change. It is
**never recorded in the state file** and never negotiated: there is no registry, no lock, and no
allocation step that a second session could lose a race to.

**Every step is defined over the change name's UTF-8 bytes, and nothing in it is defined over
characters.** A byte is the one unit a shell and a JVM can agree on without either reaching for a
primitive whose answer depends on the ambient locale, and it is already the unit the digest uses.
The derivation, stated precisely enough that two independent implementations agree:

1. **The digest** is the first four characters of the lowercase hexadecimal **SHA-256** of the
   change name's **UTF-8** bytes, with **no trailing newline**. It is taken over the **full** change
   name — never over the prefix below, and never over a truncated or otherwise normalised form.
2. **The normalised name** replaces each of the change name's UTF-8 bytes with one character: a
   byte in `A`–`Z` is lowered by 32, a byte in `a`–`z` or `0`–`9` is kept, and **every other byte
   becomes a single `-`**. The result is always pure `[a-z0-9-]`, exactly one character per input
   byte.
3. **The prefix** is the longest leading run of whole `-`-separated segments of the **normalised
   name** whose joined length is at most 12 characters. If the first segment alone exceeds 12
   characters, the prefix is that segment's first 12 characters. Every trailing `-` is then removed.
4. **The id** is the prefix and the digest joined by a single `-`, prefix first: `<prefix>-<digest>`.

```bash verified:run in this worktree on macOS (Darwin 25.5.0) with /usr/bin/shasum, /usr/bin/sed and bash 3.2, under LC_ALL of C, en_US.UTF-8, ru_RU.UTF-8 and tr_TR.UTF-8; every id in a comment is that run's output, identical under all four
name="kan-15-parallel-myflow-do-task-lanes"

prefix="$(printf '%s' "$name" | LC_ALL=C tr 'A-Z' 'a-z' | LC_ALL=C tr -c 'a-z0-9' '-')"
while [ "${#prefix}" -gt 12 ] && [ "${prefix%-*}" != "$prefix" ]; do prefix="${prefix%-*}"; done
prefix="$(printf '%s' "${prefix:0:12}" | LC_ALL=C sed 's/-*$//')"

digest="$(printf '%s' "$name" | shasum -a 256 | cut -c1-4)"
id="$prefix-$digest"                 # kan-15-55a6
                                     # Demo_X           -> demo-x-5197
                                     # KAN-99-Fix.Thing -> kan-99-fix-feef
                                     # İstanbul-test    -> --stanbul-6b8a
```

**`LC_ALL=C` on both `tr` calls is the whole of step 2's locale independence, and it is in the block
rather than left to prose.** Without it `tr` is free to consult the ambient locale's case table, and
it does: the same `İstanbul-test` yields `istanbul-6b8a` under `en_US.UTF-8` and `--stanbul-6b8a`
under `C`, from one script on one machine. Pinning the locale makes both `tr` calls byte-wise, which
is what step 2 says they are. The explicit `A-Z`/`a-z` ranges rather than `[:upper:]`/`[:lower:]` say
the same thing twice on purpose — a character class is exactly the construct that would start
meaning something else in another locale.

**The normalisation runs first, before any length is counted, and that order is load-bearing.** From
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

**`printf '%s'` rather than `echo` is the load-bearing detail.** `echo` appends a newline, and the
digest of the name plus a newline is an unrelated value. An implementation that hashes the trailing
newline and one that does not will disagree on every id ever derived, while each looks perfectly
correct in isolation — which is precisely the drift this file being canonical is meant to prevent.
The tool is not part of the contract, only the bytes are: `sha256sum` and `openssl dgst -sha256`
yield the same digest from the same input and are equally acceptable. `shasum -a 256` is named
above because it is present on both stock macOS and Linux, so one written form runs in either
place.

**The prefix is for humans; the digest is for correctness.** The prefix is what keeps a listing of
derived resources legible — an operator reading `psql -l` can tell at a glance which change a
database belongs to, which a bare digest would not give them. The digest is what keeps two changes
apart when their prefixes collapse together, which is the likely collision rather than the exotic
one: `kan-15-parallel-lanes` and `kan-15-parallel-two` both reduce to the prefix `kan-15` and are
distinguished only by their digests, `42f2` and `77be`.

**A change name outside `[a-z0-9-]`, stated rather than left to be discovered.** myflow change names
are `<lowercased-jira-key>-<slug>` in practice, so this is a boundary case rather than a daily one,
and a boundary case nobody has written down is where two implementations drift apart. Step 2 defines
it completely, and these are its consequences:

- **Upper case is folded by the ASCII rule only**, never by Unicode case folding. `A`–`Z` lower;
  every other byte, `İ` and `Ä` included, is a separator. This is the point the three
  implementations disagreed on, because Unicode folding is where the freedom lives: `İ` (U+0130)
  folds to *two* characters under a JVM's `lowercase`, to `i` under a UTF-8-locale `tr`, and to
  nothing under a C-locale one.
- **One byte, one `-`**, so a multi-byte character yields as many `-` as it has UTF-8 bytes: `İ` is
  two bytes and becomes `--`, and a three-byte character such as `€` becomes `---`. Runs of `-` are
  **not** collapsed. `İstanbul-test` therefore normalises to `--stanbul-test` and derives
  `--stanbul-6b8a`, which is the worked example in the block above.
- **Only trailing `-` are trimmed, and all of them are.** Leading and interior runs stay, so an id
  may begin with `-` — a name that normalises to nothing but separators leaves the prefix empty and
  the id is then `-<digest>`. Every value derived from such an id is still legal: the database name
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

**Three values, and no service among them:**

- **a database name**, so one change's migrations cannot reach another change's schema;
- **an object-store bucket name**, so two changes' uploads do not land in one place;
- **a block of application ports**, so two changes' applications can be up at the same time.

A project names the variable that carries each of these and the default each falls back to, per
**Project configuration** (`skills/myflow-contracts/project-configuration.md`). The cache index is
deliberately absent from this list; it is not derived at all, for the reason given under
**The cache index** below.

**The shared services are not changed, and neither is the configuration that declares them.** Their
containers keep the names they have, their host ports stay where they are, and the container-runtime
configuration file is left byte-identical. Nothing in this contract adds a service, renames one, or
moves a published port. What is namespaced is the logical resource *inside* each service, which is
the level the failure actually lives at.

**A derived database name must be safe unquoted.** It is spelled with `_` separators rather than the
`-` the id itself uses, so it is a legal SQL identifier with no quoting at all. Every call site that
exists today may quote correctly; the rule is for the one added next year that does not. An
identifier that only works when quoted is a call site's correctness resting on a convention nobody
is checking, and the cost of removing that dependence entirely is one substitution. A dashed name
would work right up until it did not, and the failure would land in whichever tool was added last.

**`<id_underscored>` is that spelling, and the substitution is total: every `-` in the id becomes
`_`.** It applies to the whole id — the separators inside the prefix and the `-` joining the prefix
to the digest alike — so the worked name above, `kan-15-parallel-myflow-do-task-lanes`, derives the
id `kan-15-55a6` and the underscored spelling `kan_15_55a6`. Replacing only the joiner would leave
`kan-15_55a6`, which still needs quoting and so buys none of what the paragraph above is for, and
the two readings differ on every id whose prefix has more than one segment. Nothing else is
replaced: the prefix rule already confines the id to `[a-z0-9-]`, so `-` is the only character a
substitution can have to reach.

```bash verified:run in this worktree on macOS (Darwin 25.5.0) with bash 3.2; continues the id block above, whose $id is kan-15-55a6
id_underscored="${id//-/_}"          # kan_15_55a6
```

**A bucket name is not a SQL identifier, so it takes the id verbatim.** There is no unquoted-parsing
hazard to protect against, so a second spelling of the same id would be a second rule to remember
with nothing bought by it. The bucket carries a project prefix for the same reason the database name
does — an operator listing buckets should be able to tell which project a workspace's bucket belongs
to — and then the id exactly as this file derives it.

**The port block is one offset, derived from the same digest:**

```bash verified:run in this worktree on macOS (Darwin 25.5.0) with digest 55a6, the digest derived above
offset=$(( (16#$digest % 400 + 1) * 10 ))    # 3270 for digest 55a6
```

That yields a multiple of 10 between 10 and 4000, so a change's ports are stable across sessions and
can be bookmarked with no registry and no coordination — the same property the id has, for the same
reason. One offset for the whole block, rather than one per port, is what keeps the relationship
between a workspace's ports the same as the relationship between the project's declared ones.

**The block is checked free before use.** Every port in the block, not merely the first:

```bash verified:run in this worktree on macOS (Darwin 25.5.0) with /usr/sbin/lsof; exit 1 with no listener, exit 0 with one
lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1    # 0 = bound, 1 = free
```

**Any bound port discards the entire block.** If a single port in the block is already held — by
another workspace, or by an unrelated process that knows nothing about this pipeline — the **whole**
block is abandoned in favour of free-port discovery. It is never repaired port by port. A patched
block is the worst of both schemes: as unpredictable as discovery, yet indistinguishable at a glance
from the deterministic block it no longer is, so nobody can reason about which of a workspace's
ports mean what. A workspace's ports are therefore always either wholly its deterministic block or
wholly discovered.

**The ports actually bound are written into the change's manual test guide.** The guide names the
URLs of the worktree that resolved them, never the project's declared defaults. This is the
mitigation for the one cost this contract imposes on an operator: a worktree's ports differ from the
project's documented ones, so opening the documented URL out of habit reaches the main checkout's
applications — a different change's work, answering plausibly and about the wrong thing.

**What the free check does not promise.** It narrows a race, it does not close one. A port observed
free can be bound by anything on the machine between the check and the bind, and no amount of
re-checking removes that window. What the contract requires is that the failure stay loud: a bind
that fails is surfaced and the block rediscovered, never silently reassigned to whatever happened to
be free at the time.

## The cache index

**The cache index is claimed by probing, and is never derived from the digest.** A run asks the
cache which indices are already taken, claims a free one, and uses it for every process that run
starts. It is the single value in this contract that is not a function of the change name, and the
exception is deliberate.

**The reason is the size of the space.** A cache offers **sixteen** indices, so a derived value is
the digest modulo sixteen and two concurrent workspaces land on the same index one time in sixteen —
roughly **six percent**. Everywhere else in this contract a collision is either vanishingly unlikely
or loud: a database name is drawn from 65 536 digests, and a port already bound refuses the bind and
discards the block. A shared cache index is neither. Two changes would quietly share sessions and
cache entries, which is exactly the class of silent wrong answer this contract exists to remove, and
a six-percent silent failure is worse than a rarer loud one.

**That argument holds only while a probe can see the previous claim, so the claim is written where
the next probe looks — inside the cache itself.** This is the load-bearing half, and it is stated
here because getting it wrong inverts the comparison the section is built on. A claim kept only in
the claiming worktree's own state is invisible to every other worktree: each run then sees an index
nobody has written to yet as free, takes the lowest such index, and two workspaces started minutes
apart collide *reliably* rather than six percent of the time. **A probing scheme whose claim is
private is the derived scheme with its collision promoted to the default outcome.** A claim is
therefore recorded in the shared cache, under an entry naming the workspace that holds it, so it is
visible to the next claimant at the moment it is made.

**The claim is taken atomically, and an index is verified otherwise empty before it is kept.** Two
runs reaching the cache at the same instant must not both be told they have it, so the claim is
written with an operation that fails when an entry is already present, never a read followed by a
write. And an index may hold keys while holding no claim — written before this contract existed, or
by anything else pointed at that cache — so an index that turns out not to be empty once the claim
is in it is given back. Giving it back removes the claim and **nothing else**: whatever made the
index unfree belongs to whoever wrote it.

**A run that cannot get an index of its own says so and stops.** Fifteen claimable indices is a real
ceiling and abandoned changes consume it, so exhaustion is an outcome rather than a hypothetical.
The one response that is not available is sharing: a run with nothing left to claim names every
index and who holds it, and refuses. **Index 0 is never claimed** — it is the empty-id case's
declared default, so a workspace taking it would be sharing with the main checkout the moment that
checkout started.

**Deriving the index and refusing to start on a collision was considered, and is not what happens.**
It is never silently wrong, which is the property that matters — but it blocks a legitimate second
workspace outright, with renaming the change as the only remedy available to the operator. Probing
gives the same safety and costs nothing at the point of use.

**There is no expiry, deliberately.** A workspace's stack runs for days and nothing in this contract
is left running to renew a lease, so a lifetime short enough to clean up after a crashed run would
expire under a live one — the silent share again, by a slower route. Identity replaces it: a claim
naming its holder can be released, reported and listed by name. The claim also lives exactly as long
as what it protects, since it is kept in the index it reserves — a cache that restarts without
persistence loses the claim and the entries together, which is correct, because there is then
nothing left to keep apart.

**The accepted cost is that the index is not stable across restarts.** Every other value a workspace
derives is the same one a week later; this one depends on what else happened to be running when the
run first claimed. That is acceptable for precisely the reason it is safe to reassign in the first
place: what the cache holds is sessions and cache entries, both disposable by construction. A
workspace that lands on a different index loses a login, not data — so nothing that must survive a
restart may be kept there.

**This pipeline releases nothing at finish, and the claimed index has a registry row saying so.**
Run 2 has no derivation to repeat and no record to read, so it cannot know which index to sweep, and
an index swept by guess is another workspace's. What that leaves behind and why it is acceptable are
stated once under **Temporary artifacts registry** (`skills/myflow-contracts/pipeline.md`), which is
canonical for every artifact's lifetime. **A project whose claim is visible can do better than the
pipeline can**, and the ceiling above is the reason to: releasing the claim in its own `remove`
command, reporting a claim that outlived cleanup through `survivors`, and listing every claim on the
machine without needing an id — which is what an abandoned change does not leave behind. None of
that is required here, and none of it changes the registry row.


## The empty id

**Every value derived from a workspace id resolves to the project's declared default when the id is
empty.** This is a rule the configuration must satisfy, not a description of any one project's file:
with no id set, the database name, the bucket name and every application port are exactly the ones
the project had before it adopted this contract. In practice each derived value is written as an
interpolation with an unset-or-empty default — `${VAR:-<base>}`, whose base is today's literal
value — so the id-less case is not a branch anybody maintains but the absence of a substitution.

**The main checkout is the empty-id case**, and so is every project that declares no isolation at
all. A project with no isolation section in its configuration behaves exactly as it does today
**everywhere, including in apply worktrees**, and that is not reported as a misconfiguration: for a
repository with no runnable application it is the correct state. Every key in a project's
configuration is optional; see **Project configuration** (`skills/myflow-contracts/project-configuration.md`),
which is canonical for that file and for the rules an isolation row resolves under.

**Isolation in an apply worktree is automatic, never opt-in.** A worktree that declares isolation
receives its own database, cache index, bucket and port block without anybody asking for them, and
the main checkout receives the defaults. Putting the isolation behind a request was considered and
rejected on one point: the failure being prevented is the one that *forgetting* the request
reproduces, and reproduces silently — the applications come up, they answer, and only the schema is
wrong. A safeguard that has to be remembered is absent in exactly the runs that needed it, so an
opt-in guard against a silent failure has the failure as its own failure mode. An operator who
genuinely wants the shared default still has one: the project's declared defaults are reachable
through the same variables the empty-id case resolves to, so choosing them is a deliberate act
rather than an omission.

**A malformed row does not fall back to its declared default in an apply worktree, and the tension
there is worth stating rather than resolving quietly.** A malformed row is reported by name and
dropped rather than repaired — repairing it would be a guess about a resource that may not exist.
Which shapes fail, how a failure is reported, and what *row* and *cell* each name are stated under
**Project configuration** (`skills/myflow-contracts/project-configuration.md`). But the value a
dropped row would fall back to is its declared default, and that default is by construction the
project's **shared** resource: it is precisely the value the empty-id case is built on. The two
checkouts therefore resolve a drop differently, and the asymmetry is a decision rather than an
inconsistency:

- **In the main checkout there is no workspace id, the declared default is the correct value, and
  nothing changes.** The value a dropped row would have produced *is* the default, so the drop
  costs nothing. The row is still reported by name, which puts the misconfiguration in front of
  someone where it is cheap to fix rather than only where it is expensive.
- **In an apply worktree a dropped row means that value cannot be isolated, so the run refuses to
  proceed with it.** Using the shared default there is the silent failure named one paragraph above,
  reached by a different route. The run reports the row, the cell that failed validation and the
  shared value it is declining to use, and stops before starting anything that would read that
  variable. It neither exports the default nor leaves the variable unset, since an unset variable
  resolves to the same shared value one level further in. Correcting the row and re-running is the
  whole remedy — a dropped row moves no state.

**The refusal is keyed on the row having been dropped,
never on what kind of resource the row names.** A row naming a plain value rather than something
that exists to be removed is not exempt: a dropped `url` row refuses exactly as a dropped `database`
row does, and so does a `url` row dropped because its reference
named no row or named another `url` row. A public base URL that falls back to its default in a
worktree points at the **shared** bucket, so the applications come up and answer about another
change's uploads — one step later than the schema case and identical in kind. Which rows a resource
word selects is a cleanup question and only that; it decides nothing about isolation.

**Refusing rather than reporting-and-continuing is the same argument that makes isolation automatic
rather than opt-in**, one paragraph above: a safeguard that depends on somebody reading a report is
absent in exactly the run that needed it.

**Why the empty id is the default rather than a special case.** The uniform alternative — every
checkout takes an id, the main one included — has no empty-id branch to get wrong, and breaks every
existing local setup exactly once: databases orphaned, saved connections pointing at nothing,
bookmarks stale. That is a migration nobody asked for, in the one checkout an operator spends most
of their time in. The backwards-compatibility promise is therefore load-bearing, and a change to a
derived value that does not preserve it is a defect in that change, not a limitation of this
contract.

## Creation and cleanup

**The resources are created on demand, and the first creation is reported.** A run that finds no
database or bucket for its workspace creates them and says so, once. Something now exists on the
machine that did not exist before, and a side effect nobody is told about is one nobody thinks to
look for later. Later runs find the resources already there and print nothing, so the notice carries
real information — it means *this is new*, not merely *this is here*.

**A new database starts empty and is brought up to date by the project's normal migration and
seeding path**, which is the same path a fresh machine and a clean CI job take. It is never created
by copying an existing database. Copying looks faster and is wrong twice over: it carries the source
database's migration state, which need not be the state this branch's migrations expect, and it
carries whatever rows happened to be sitting there, so the workspace's database is subtly unlike a
clean one and the difference surfaces as a result nobody can reproduce. Taking the ordinary path
costs one migration run and buys a database that is exactly what a clean build would produce.

**At finish, the resources are removed.** Each is a temporary artifact and therefore has a row
naming what creates it, where it lives and what removes it, in
**Temporary artifacts registry** (`skills/myflow-contracts/pipeline.md`). What performs the removal
is the command the project declares for it, given the workspace id so the teardown targets this
change's resources rather than the project's defaults. **Which command that is belongs to the
project, not to this contract** — a project names its create and remove commands in its own
configuration, per **Project configuration** (`skills/myflow-contracts/project-configuration.md`),
and whether it reuses a command it already had or adds one is its decision to record there. Naming a
mechanism here would fix in the canonical file a choice each project makes for itself, and the first
project to make it differently would make this file wrong.

**Removal is not verification, so a project declares a third command that reports survivors.** The
registry row promises that a workspace's database and bucket are *gone*,
and "ran the removal" is not *verified gone* — a removal that reported success against a stale
connection, or a bucket a policy refused to delete, both leave the promise broken with nothing
having failed. A survivor is therefore established by asking, never inferred from the removal's exit
code. Why a third verb rather than two, rather than reading the removal's own result: the guard that
checks the row lives in the agents repository and must stay project-agnostic, so it cannot hold
`psql -l` or one project's object-store client. The project owns the question and answers it in its
own configuration, exactly as it owns creation and removal —
see **Project configuration** (`skills/myflow-contracts/project-configuration.md`), canonical for how
the three commands are written, what the survivor report prints, what its exit code means, what a
non-empty report does to the terminal state, and what a project that declares no survivor report at
all gets in place of the verification.

**A service that is not running is reported and skipped, rather than failed.** If the database
server is down when run 2 reaches cleanup, there is nothing to remove at that moment: the skip is
reported by name and the run continues. This is deliberately unlike a reported survivor, which
blocks the terminal state under **Project configuration** (`skills/myflow-contracts/project-configuration.md`),
and the asymmetry is stated here rather than left for a reader to find and mistake for an
oversight. Blocking an already-merged
change from finishing, over a service that happens to be stopped, trades a real cost — a change
stranded short of its terminal state — for a few megabytes of stale storage. It also matches how the
project-supplied stop check already treats an unreachable stack. A change whose project declares no
isolation passes the same way: a step whose artifact is already absent is a success, which is the
re-entrancy rule run 2 follows everywhere else.

**That skip is signalled by one non-zero exit rather than two, deliberately.** Telling "the service
is stopped" apart from "the survivor report is broken" would need a second exit-code convention that
every project implements identically, and at this point in the pipeline both have the same correct
response: run 2 is past the merge, and blocking an already-merged change over a service that happens
to be down is what the paragraph above forbids. The cost is real and is named rather than hidden — a
genuinely broken survivor report is indistinguishable here from a stopped service — which is why the
skip carries the command's name and its exit code instead of being noted in passing.
