# Workspace isolation

**This file is canonical for workspace isolation.** Skills, guards, and the projects that declare
isolation reference it by name; none of them restate the derivation, the port rule, or the empty-id
promise. See **The workspace id** (`skills/flow-contracts/workspace-isolation-rationale.md`) for
why a second copy of a derivation is worse than a second copy of a procedure. If a rule below and a
skill, a guard, or a project's own configuration ever disagree, this file wins.

See **The workspace id** (`skills/flow-contracts/workspace-isolation-rationale.md`) for the
failure this contract exists to prevent.

**What is isolated is the logical resource, never the service that holds it.** A workspace gets its
own database inside the shared database server, its own index inside the shared cache, and its own
bucket inside the shared object store. See **The workspace id**
(`skills/flow-contracts/workspace-isolation-rationale.md`) for why duplicating whole services per
workspace was not the design taken instead.

## The workspace id

**An apply worktree has a workspace id; the main checkout has none.** The id is derived from the
change name and from nothing else. It is **deterministic** — the same change name always yields the
same id, in this session and in one a week later — so two sessions that never communicate cannot
produce the same id for two different changes, or two different ids for one change. It is
**never recorded in the state file** and never negotiated: there is no registry, no lock, and no
allocation step that a second session could lose a race to.

**Every step is defined over the change name's UTF-8 bytes, and nothing in it is defined over
characters.** See **The workspace id** (`skills/flow-contracts/workspace-isolation-rationale.md`)
for why a byte, rather than a character, is the unit two independent implementations can agree on.
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
rather than left to prose.** See **The workspace id**
(`skills/flow-contracts/workspace-isolation-rationale.md`) for the measured locale divergence this
guards against.

**The normalisation runs first, before any length is counted, and that order is load-bearing.** See
**The workspace id** (`skills/flow-contracts/workspace-isolation-rationale.md`) for why the order
matters and what it produces for a name carrying a dot.

See **The workspace id** (`skills/flow-contracts/workspace-isolation-rationale.md`) for why a name
already within `[a-z0-9-]` passes through step 2 unchanged, and what that means for testing a second
implementation.

**`printf '%s'` rather than `echo` is the load-bearing detail.** See **The workspace id**
(`skills/flow-contracts/workspace-isolation-rationale.md`) for why, and for which tools are
acceptable.

**The prefix is for humans; the digest is for correctness.** See **The workspace id**
(`skills/flow-contracts/workspace-isolation-rationale.md`) for the collision example this
separates.

**A change name outside `[a-z0-9-]`, stated rather than left to be discovered.** myflow change names
are `<lowercased-jira-key>-<slug>` in practice, so this is a boundary case rather than a daily one,
and a boundary case nobody has written down is where two implementations drift apart. Step 2 defines
it completely, and these are its consequences:

- **Upper case is folded by the ASCII rule only**, never by Unicode case folding. `A`–`Z` lower;
  every other byte, `İ` and `Ä` included, is a separator. See **The workspace id**
  (`skills/flow-contracts/workspace-isolation-rationale.md`) for the three-implementation
  disagreement this rule settles.
- **One byte, one `-`**, so a multi-byte character yields as many `-` as it has UTF-8 bytes: `İ` is
  two bytes and becomes `--`, and a three-byte character such as `€` becomes `---`. Runs of `-` are
  **not** collapsed. `İstanbul-test` therefore normalises to `--stanbul-test` and derives
  `--stanbul-6b8a`, which is the worked example in the block above.
- **Only trailing `-` are trimmed, and all of them are.** Leading and interior runs stay, so an id
  may begin with `-` — a name that normalises to nothing but separators leaves the prefix empty and
  the id is then `-<digest>`. See **The workspace id**
  (`skills/flow-contracts/workspace-isolation-rationale.md`) for why every value derived from such
  an id is still legal.

See **The workspace id** (`skills/flow-contracts/workspace-isolation-rationale.md`) for why the
digest, unlike the prefix, is locale-independent already and cannot be merged by normalising.

See **The workspace id** (`skills/flow-contracts/workspace-isolation-rationale.md`) for what
determinism does not promise about a digest collision, and the remedy if one ever bites.

## What the id derives

**Three values, and no service among them:**

- **a database name**, so one change's migrations cannot reach another change's schema;
- **an object-store bucket name**, so two changes' uploads do not land in one place;
- **a block of application ports**, so two changes' applications can be up at the same time.

A project names the variable that carries each of these and the default each falls back to, per
**Project configuration** (`skills/flow-contracts/project-configuration.md`). The cache index is
deliberately absent from this list; it is not derived at all, for the reason given under
**The cache index** below.

**The shared services are not changed, and neither is the configuration that declares them.** Their
containers keep the names they have, their host ports stay where they are, and the container-runtime
configuration file is left byte-identical. Nothing in this contract adds a service, renames one, or
moves a published port. What is namespaced is the logical resource *inside* each service, which is
the level the failure actually lives at.

**A derived database name must be safe unquoted.** It is spelled with `_` separators rather than the
`-` the id itself uses, so it is a legal SQL identifier with no quoting at all. See **What the id
derives** (`skills/flow-contracts/workspace-isolation-rationale.md`) for why this is required of
the one call site added next year, not only the ones that exist today.

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

**A bucket name is not a SQL identifier, so it takes the id verbatim.** See **What the id derives**
(`skills/flow-contracts/workspace-isolation-rationale.md`) for why no second spelling is used
here. The bucket carries a project prefix for the same reason the database name does — an operator
listing buckets should be able to tell which project a workspace's bucket belongs to — and then the
id exactly as this file derives it.

**The port block is one offset, derived from the same digest:**

```bash verified:run in this worktree on macOS (Darwin 25.5.0) with digest 55a6, the digest derived above
offset=$(( (16#$digest % 400 + 1) * 10 ))    # 3270 for digest 55a6
```

That yields a multiple of 10 between 10 and 4000, so a change's ports are stable across sessions and
can be bookmarked with no registry and no coordination — the same property the id has, for the same
reason. See **What the id derives** (`skills/flow-contracts/workspace-isolation-rationale.md`) for
why one offset covers the whole block rather than one per port.

**The block is checked free before use.** Every port in the block, not merely the first:

```bash verified:run in this worktree on macOS (Darwin 25.5.0) with /usr/sbin/lsof; exit 1 with no listener, exit 0 with one
lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1    # 0 = bound, 1 = free
```

**Any bound port discards the entire block.** If a single port in the block is already held — by
another workspace, or by an unrelated process that knows nothing about this pipeline — the **whole**
block is abandoned in favour of free-port discovery. It is never repaired port by port. See **What
the id derives** (`skills/flow-contracts/workspace-isolation-rationale.md`) for why a patched
block is worse than either alternative. A workspace's ports are therefore always either wholly its
deterministic block or wholly discovered.

**The ports actually bound are printed in the `/myflow-do` handoff.** The handoff names the
URLs of the worktree that resolved them, never the project's declared defaults. See **What the id
derives** (`skills/flow-contracts/workspace-isolation-rationale.md`) for the cost this mitigates
and the failure it prevents.

**What the free check does not promise.** What the contract requires is that the failure stay loud:
a bind that fails is surfaced and the block rediscovered, never silently reassigned to whatever
happened to be free at the time. See **What the id derives**
(`skills/flow-contracts/workspace-isolation-rationale.md`) for why the check only narrows the race
rather than closing it.

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
the next probe looks — inside the cache itself.** See **The cache index**
(`skills/flow-contracts/workspace-isolation-rationale.md`) for why a private claim would invert
the comparison the section is built on. **A probing scheme whose claim is private is the derived
scheme with its collision promoted to the default outcome.** A claim is therefore recorded in the
shared cache, under an entry naming the workspace that holds it, so it is visible to the next
claimant at the moment it is made.

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

See **The cache index** (`skills/flow-contracts/workspace-isolation-rationale.md`) for the
deriving-and-refusing alternative that was considered and rejected for this design.

**There is no expiry, deliberately.** See **The cache index**
(`skills/flow-contracts/workspace-isolation-rationale.md`) for why a lease-based lifetime would
expire under a live workspace. Identity replaces it: a claim naming its holder can be released,
reported and listed by name. The claim also lives exactly as long as what it protects, since it is
kept in the index it reserves — a cache that restarts without persistence loses the claim and the
entries together, which is correct, because there is then nothing left to keep apart.

**The accepted cost is that the index is not stable across restarts.** Every other value a workspace
derives is the same one a week later; this one depends on what else happened to be running when the
run first claimed. See **The cache index**
(`skills/flow-contracts/workspace-isolation-rationale.md`) for why that is an accepted cost. A
workspace that lands on a different index loses a login, not data — so nothing that must survive a
restart may be kept there.

**This pipeline releases nothing at finish, and the claimed index has a registry row saying so.**
See **The cache index** (`skills/flow-contracts/workspace-isolation-rationale.md`) for why run 2
cannot know which index to sweep. What that leaves behind and why it is acceptable are stated once
under **Temporary artifacts registry** (`skills/flow-contracts/artifacts-registry.md`).
**A project whose claim is visible can do better than the pipeline
can**, and the ceiling above is the reason to: releasing the claim in its own `remove` command,
reporting a claim that outlived cleanup through `survivors`, and listing every claim on the machine
without needing an id — which is what an abandoned change does not leave behind. None of that is
required here, and none of it changes the registry row.


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
configuration is optional; see **Project configuration** (`skills/flow-contracts/project-configuration.md`).

**Isolation in an apply worktree is automatic, never opt-in.** A worktree that declares isolation
receives its own database, cache index, bucket and port block without anybody asking for them, and
the main checkout receives the defaults. See **The empty id**
(`skills/flow-contracts/workspace-isolation-rationale.md`) for the opt-in alternative this rejects
and why. An operator who genuinely wants the shared default still has one: the project's declared
defaults are reachable through the same variables the empty-id case resolves to, so choosing them is
a deliberate act rather than an omission.

**A malformed row does not fall back to its declared default in an apply worktree, and the tension
there is worth stating rather than resolving quietly.** A malformed row is reported by name and
dropped rather than repaired — repairing it would be a guess about a resource that may not exist.
Which shapes fail, how a failure is reported, and what *row* and *cell* each name are stated under
**Project configuration** (`skills/flow-contracts/project-configuration.md`). But the value a
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

See **The empty id** (`skills/flow-contracts/workspace-isolation-rationale.md`) for why refusing,
rather than reporting and continuing, follows the same argument that makes isolation automatic.

See **The empty id** (`skills/flow-contracts/workspace-isolation-rationale.md`) for why the empty
id is the default rather than a special case, and the uniform alternative it rejects. The
backwards-compatibility promise is therefore load-bearing, and a change to a derived value that does
not preserve it is a defect in that change, not a limitation of this contract.

## Creation and cleanup

**The resources are created on demand, and the first creation is reported.** A run that finds no
database or bucket for its workspace creates them and says so, once. See **Creation and cleanup**
(`skills/flow-contracts/workspace-isolation-rationale.md`) for why the first creation is reported.
Later runs find the resources already there and print nothing, so the notice carries real
information — it means *this is new*, not merely *this is here*.

**A new database starts empty and is brought up to date by the project's normal migration and
seeding path**, which is the same path a fresh machine and a clean CI job take. It is never created
by copying an existing database. See **Creation and cleanup**
(`skills/flow-contracts/workspace-isolation-rationale.md`) for why copying was rejected.

**At finish, the resources are removed.** Each is a temporary artifact and therefore has a row
naming what creates it, where it lives and what removes it, in **Temporary artifacts registry**
(`skills/flow-contracts/artifacts-registry.md`). What performs the removal is the command the project
declares for it, given the workspace id so the teardown targets this change's resources rather than
the project's defaults. **Which command that is belongs to the project, not to this contract** — a
project names its create and remove commands in its own configuration, per **Project configuration**
(`skills/flow-contracts/project-configuration.md`), and whether it reuses a command it already had
or adds one is its decision to record there. See **Creation and cleanup**
(`skills/flow-contracts/workspace-isolation-rationale.md`) for why no mechanism is named here.

**Removal is not verification, so a project declares a third command that reports survivors.** The
registry row promises that a workspace's database and bucket are *gone*,
and "ran the removal" is not *verified gone* — a removal that reported success against a stale
connection, or a bucket a policy refused to delete, both leave the promise broken with nothing
having failed. A survivor is therefore established by asking, never inferred from the removal's exit
code. Why a third verb rather than two, rather than reading the removal's own result: the guard that
checks the row lives in the agents repository and must stay project-agnostic, so it cannot hold
`psql -l` or one project's object-store client. The project owns the question and answers it in its
own configuration, exactly as it owns creation and removal —
see **Project configuration** (`skills/flow-contracts/project-configuration.md`), canonical for how
the three commands are written, what the survivor report prints, what its exit code means, what a
non-empty report does to the terminal state, and what a project that declares no survivor report at
all gets in place of the verification.

**A service that is not running is reported and skipped, rather than failed.** If the database
server is down when run 2 reaches cleanup, there is nothing to remove at that moment: the skip is
reported by name and the run continues. This is deliberately unlike a reported survivor, which
blocks the terminal state under **Project configuration**
(`skills/flow-contracts/project-configuration.md`), and the asymmetry is stated here rather than
left for a reader to find and mistake for an oversight. See **Creation and cleanup**
(`skills/flow-contracts/workspace-isolation-rationale.md`) for the cost this asymmetry trades
against, and why it matches the project-supplied stop check. A change whose project declares no
isolation passes the same way: a step whose artifact is already absent is a success, which is the
re-entrancy rule run 2 follows everywhere else.

**That skip is signalled by one non-zero exit rather than two, deliberately.** See **Creation and
cleanup** (`skills/flow-contracts/workspace-isolation-rationale.md`) for why one exit code, and
what cost that accepts.