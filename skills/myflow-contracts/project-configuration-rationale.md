# Project configuration — rationale

This file is the reasoning behind `skills/myflow-contracts/project-configuration.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

## Where the agents repository is

### The `.mdc` routing rule

**The `.mdc` extension is what selects the shared library, and nothing else.** Shared opt-in rules
are deliberately **not** installed globally, so a bare filename is the only way to name one — but
every shared rule in `<agents repo>/rules/` is an `.mdc` file, while a project's own standards are
overwhelmingly `<project>/CLAUDE.md`, `<project>/AGENTS.md`, `CONTRIBUTING.md`, and the like. Routing *all* bare
filenames to the agents repo therefore sent a project's real standards to a path that does not
exist, where the drop rule silently discarded them and the principles reviewer lost the whole
project-specific half of its mandate without anything erroring. A project-local `.mdc` is still
nameable — write it as a path (`<project>/.cursor/rules/api.mdc`), which form 3 takes as-is.

### The per-skill link, not the `skills/` directory

**The link to resolve is the per-skill one, never the `skills/` directory above it.** A global
install is a real `~/.claude/skills/` (likewise `~/.cursor/skills/` and `~/.codex/skills/`) holding
**one symlink per skill** — `check-vocabulary.sh` calls that shape "a farm of symlinked
directories" and traverses it deliberately. Resolving the parent lands in `~/.claude`, which
contains no per-skill scripts/ directory and is not a checkout of anything; resolving the skill's own directory lands
in `<agents repo>/skills/<skill>/`, from which step 2 is right. Measured on a global install:
`~/.claude/skills/` is a directory, `~/.claude/skills/myflow-do` is a symlink to
`<agents repo>/skills/myflow-do/`, and the two-step rule yields the checkout root while resolving
the parent yields `~/.claude`.

### Project-local installs need no link

**The same two steps are correct for a project-local install**, where nothing is a symlink at all:
the skill directory is already physical, step 1 is a no-op, and step 2 leaves the checkout root.
This is the shape a project-local skill's own read falls back to, so a rule that depended on a link to
resolve would have no answer here.

### Confirm the derived root before using it

A skill directory that was **copied** into a project rather than linked resolves two levels up to
something that is not the agents repository — an unrelated project root, or, for a global install,
the operator's own configuration directory — and there
is no third step that recovers the real one. That is the absent-script case, and each command says
what it does with it.

### Prose beside the tables is for the reader

**Prose beside the tables is permitted and is never read.** *Not read* is a statement about the
resolver, not a restriction on the author: a project may write whatever a human reader needs — most
usefully, what it has **not** isolated and why, which is the one thing the tables cannot show,
because what a project cannot move is expressed by a row that is absent. An application whose port
is fixed outside this project's repository leaves no trace in either table beyond a `url` row that
happens to keep its `Default`, and a reader who cannot tell that from an oversight will eventually
try to fix it. So the limitation is written down here, in the section it belongs to, and the
resolver ignores it exactly as it ignores every other word outside the two tables.

### The accepted cost of `url`-row duplication

**The duplication that follows is an accepted cost, not an oversight.** A project whose public base
URL and endpoint share a host must write that host in both rows, since neither may reference the
other. Sparing it means multi-level references, which reinstate all three things the single pass
just removed — a cycle to detect, a resolution order to declare rows in, and a partially-resolved
value to reason about when detection fires — in exchange for a repeated hostname. The duplication
this declines to fix is usually already there anyway: the same host is written once per variable in
the project's own configuration, which is where these rows' `Default` values are copied from.

### A `url` row with no token

**A `url` row carrying no token at all is a legitimate declaration, not a mistake.** Its workspace
value is its `Default`, unchanged — a project writes one when a variable its applications read is
genuinely the same in both checkouts and it wants the exported set to be the complete set, rather
than a subset a reader has to diff against the defaults to trust. Nothing removes a `url` row and
nothing is created for it — which is the whole of what the `Resource` word decides here.

### The dropped `url` row is not exempt

**A dropped `url` row is not exempt from the refusal.** The refusal is
keyed on the row having been dropped, never on whether the row names something removable:
a dropped `url` row refuses in an apply worktree exactly as a dropped `database` row does, and so
does a `url` row dropped because its `<value:…>` reference named no row or named another `url` row.
A base URL falling back to its `Default` there points at the project's shared bucket, which is the
same silent wrong answer as falling back to the shared database, reached one step later. The
argument is stated once under **The empty id** (`skills/myflow-contracts/workspace-isolation.md`).

### Working directory for `survivors`, `remove`, and `create`

**Each command runs with a repository root as its working directory, and which root is fixed here
rather than left to the caller.** It has to be stated because a command in this table is routinely
repo-relative: `<project>/gradlew workspaceRemove` and `<project>/scripts/workspace remove` are two of the three
shapes this file writes one in, and both resolve against the working directory, so a project author
cannot write one correctly without being told which directory that is.

- **`survivors` runs from the main checkout**, and that is not a convention invented here.
  `<agents repo>/scripts/check-cleanup-complete.sh` launches it as
  `( cd "$REPO" && exec bash -o pipefail -c "$cmd" )` against
  the repository it was handed, and it is handed the main checkout: by the time the guard asks, the
  change's apply worktree has already been removed.

- **`remove` runs from the main checkout** too. Run 2 calls it after the worktree half of its cleanup
  step, per **Run 2 — the branch is merged** (`skills/myflow-contracts/finish-contract.md`), so the
  apply worktree is by then not a directory anything could run from — and two commands out of one
  table, called by one run of one command, must not disagree about where they run.

- **`create` runs from the apply worktree**, because of who calls it: whatever starts the project's
  applications does, and that is the worktree whose applications need the resources — the same
  first-start-in-a-worktree fact the row in
  **Temporary artifacts registry** (`skills/myflow-contracts/pipeline.md`) records. The asymmetry is
  the rule working rather than an exception to it: each command runs from the checkout that still
  exists when it is called.

### The one-minute bound

**The bound and its outcome are stated here because a `survivors` command is written against
them.** A command that cannot answer in a minute needs a different design — a cached answer, a
narrower query — rather than a longer bound. The outcome deliberately differs from the stop
command's, which shares the bound and *fails* on it: an un-stopped stack is a reason not to remove
a worktree, so blocking is the safe direction there, while blocking an already-merged change over
a slow report is what the skip rule exists to prevent.

**What the bound terminates is a process group on the machine running the guard, and a command
that puts its real work outside that group outlives the bound.** The termination is
`<agents repo>/scripts/check-cleanup-complete.sh`'s, and it signals the process group it created — which reaches
a pipeline, a subshell and any ordinary child, and reaches nothing that has left the group. Two
shapes leave it, and the second is the one a project actually writes: a command that starts its
work under its own job control, and **a command that reaches its service through a container
runtime**, `docker exec …` most of all. There the client the guard signals is a *proxy*: the work
runs in another PID namespace, where no signal sent from the host can name it, so killing the
client ends the guard's wait and leaves the query running inside the container. Measured on this
machine: a group kill reaped the in-group child and left a sibling that had claimed a group of its
own still running afterwards.

**That is a limitation of the bound, not a defect the guard can close**, and it is stated here
because the command is the only place it can be fixed. A guard that stays project-agnostic and
single-file cannot hold one container runtime's `kill` any more than it can hold `psql -l` — the
reason a third verb exists at all, under
**Creation and cleanup** (`skills/myflow-contracts/workspace-isolation.md`). What the bound still
guarantees is what run 2 needs: the *guard* returns, the row is reported as a skip rather than as a
verification, and nothing the escaped process prints later is read as a survivor report. What it
does not guarantee is that the escaped process stopped — and inside a shared container that process
may be holding a lock every other concurrent workspace is waiting on. **So a `survivors` command
that crosses a container boundary carries its own timeout inside the container** — `docker exec`'s
own, the client tool's connect and statement timeouts, or a wrapper that ends the query rather than
the proxy. A command written that way needs nothing from this bound; one written without it can
strand work the guard has already stopped waiting for.

### The three-`url`-rows worked example

The three `url` rows are the shape a project reaches for as soon as one of its variables carries a
URL rather than a port: `MEDIA_BASE_URL` keeps the shared object store's own host and port — that
service is not isolated — while following the bucket the workspace renames, and `WEB_URL` follows
the port block. `DB_CONSOLE_URL` is the case `<value:…>` cannot reach, and is why `<id>` and
`<id_underscored>` are legal here rather than only in a `database` or `bucket` cell: the console
itself is shared, the database it opens is this workspace's, and the `database` row's value is a
whole JDBC URL rather than the bare name — so `<value:DB_URL>` would substitute the connection
string, and the row spells `<id_underscored>` out itself instead. None of the three is expressible
as a `port` or `bucket` row, and none is a resource anything removes.

### No path is isolated inside a command

**No path is isolated inside a command, and that narrowing is deliberate.** A realistic command row
often has no path to check at all: `docker exec <container> dropdb <id_underscored>` names none at
all, `<project>/gradlew workspaceRemove` names the project's own build wrapper, and `<project>/scripts/workspace remove`
names one script — three shapes, one rule needed. Finding "the path" in an arbitrary command means
parsing a shell command, and a parser that is wrong about one command is a containment rule that
reports the wrong thing about all of them. What containment buys elsewhere in this file is specific
and does not carry over: a `## standards` entry is **read as a file whose content lands in a
committed review record**, which is an arbitrary-file-read primitive. Executing a command creates no
such primitive — a repository whose `## run` section starts its stack can already run anything
through that key, and the pull request that could edit a `create` command could edit `## run`
beside it. So the honest scope is the narrow one: containment binds what the pipeline **reads**,
validation binds what it **substitutes**, and a command is **run**. If a key in this section ever
named a file the pipeline reads, that path would be contained exactly as a `## standards` path is —
the rule follows the read, not the section.

### Where enforcement happens

**Where that enforcement actually happens, stated exactly, because "a guard exists" is not "a guard
ran".** The guard runs at the point this section is *read*: `/myflow-do` runs it against each apply
worktree before it resolves or exports a single row, per section 7 of `skills/myflow-do/SKILL.md`,
and a non-zero exit stops that run. That is what makes the enforcement reach every project myflow is
installed into, rather than only the repository the guard ships in — where it is *also* a lint step,
which is a self-check on the agents repository's own configuration and nothing more. A project that
declares no section passes silently in both places, which is the overwhelmingly common case.
`/myflow-finish` deliberately does not repeat the validation: run 2 reads the `survivors` row alone
and already reports every input it cannot resolve as a skip, and a blocking validation there would
strand an already-merged change over text that session cannot correct — the trade
**Creation and cleanup** (`skills/myflow-contracts/workspace-isolation.md`) rejects when it weighs a
change stranded short of its terminal state against stale storage. So a malformed declaration is
caught before it can be used, and never at the point where reporting it would strand a change.
