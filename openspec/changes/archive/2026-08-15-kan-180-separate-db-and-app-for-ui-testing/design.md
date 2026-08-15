# Design

Adapted from the approved design at
`docs/superpowers/specs/2026-08-15-kan-180-separate-db-and-app-for-ui-testing-design.md`.

## The two halves

Two failures, one shared mechanism.

**Per-worktree isolation** is for *development*. A `/myflow-do` apply worktree currently runs against
the main checkout's database and port, because this repository declares no `## workspace isolation`
section. Two changes in flight share a schema and collide on 4173.

**The UI-test stack** is for *ad-hoc testing from the main checkout*. Worktree isolation would not
have prevented the pollution that motivated this change: the junk rows naming `/tmp` paths were
written by runs against the main checkout's daemon, not from inside any worktree. Conversely, a
single test stack does nothing for two concurrent worktrees.

`MYFLOW_ADDR` is what both halves need — it points the CLI at the UI-test stack, and at a worktree's
own daemon.

## What already exists

`internal/config.FromEnv` resolves `MYFLOWD_HOST`, `MYFLOWD_PORT` and `MYFLOWD_DSN`, refusing an
unparsable port rather than falling back silently. `harvest.DefaultTranscriptsRoot` honours
`MYFLOW_TRANSCRIPTS_DIR`; the fallback state root honours `MYFLOW_STATE_DIR`. The daemon can already
be moved onto another port, database and transcript root with no code change.

The gap is entirely on the client side and in the project's own declaration.

## Per-worktree isolation

`.myflow/project.md` gains a `## workspace isolation` section, declared exactly as
`skills/myflow-contracts/project-configuration.md` specifies.

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| `database` | `MYFLOWD_DSN` | `postgres://myflow:myflow@localhost:5433/myflow?sslmode=disable` | `postgres://myflow:myflow@localhost:5433/myflow_<id_underscored>?sslmode=disable` |
| `port` | `MYFLOWD_PORT` | `4173` | `+<offset>` |
| `url` | `MYFLOW_ADDR` | `http://127.0.0.1:4173` | `http://127.0.0.1:<value:MYFLOWD_PORT>` |

Each `Default` is today's literal value. That column is where the backwards-compatibility promise is
kept: with no workspace id, every variable resolves to what the main checkout uses now.

`scripts/workspace.sh` supplies `create`, `remove` and `survivors`. The contract fixes each one's
working directory — `create` from the apply worktree, `remove` and `survivors` from the main
checkout — and two properties of `survivors` are load-bearing:

- **Filter with `awk`, never `grep`.** `grep` exits 1 on no match, and a non-zero exit is read as
  *the check could not run*, so a `grep`-filtered report can never reach the result that verifies
  the cleanup.
- **Carry a timeout inside the container.** The guard's sixty-second bound terminates a host process
  group; work behind `docker exec` is not in it.

### Not isolated, deliberately

`MYFLOW_STATE_DIR` and `MYFLOW_TRANSCRIPTS_DIR` get no rows — the `Resource` vocabulary is closed
(`database`, `bucket`, `cache index`, `port`, `url`) and a directory path is none of them. The
consequence is bounded: a worktree harvests the same transcripts into its *own* database, so data is
duplicated rather than shared, and the fallback state directory is written only when the daemon is
unreachable. Neither crosses the boundary this change protects. No `bucket` or `cache index` row —
the project has neither.

## The UI-test stack

| Resource | Live | UI test |
|----------|------|---------|
| Database | `myflow` | `myflow_uitest`, same container, host port 5433 |
| Daemon port | 4173 | 4174 |
| Transcripts root | the real projects directory | an empty temporary directory |
| Fallback state directory | the real one | a temporary directory |

Pointing the transcript root away from the real one is load-bearing, not tidiness: the daemon
harvests every five seconds, so a test database left on the real root refills itself with genuine
session data within seconds of starting.

`make ui-test-up` drops and recreates `myflow_uitest`, starts the daemon on 4174 with the temporary
roots exported, waits for it to answer, and seeds the fixture. Resetting on every bring-up is what
keeps the stack trustworthy — a persistent test database accumulates exactly the way the live one
did. `make ui-test-down` stops it and drops the database.

`stats/cmd/uitest-seed/` writes a fixed fixture through the store's own types: two projects, changes
spanning `STARTED`, `IN_PROGRESS` and `FINISHED`, and stage runs carrying token usage so the cost
views render. Go rather than SQL, so a migration that changes a column breaks the build instead of
loading into a shape the views render wrongly.

## The name guard

The seeder and the teardown refuse to act unless the database named by their target ends in
`_uitest`, checked before any statement is issued. This turns the acceptance criterion from a
convention into a structural property: the live database was polluted while "do not point this at
the live database" was in force as advice.

The daemon carries no such restriction — serving the live database is its ordinary job.

## Decisions

### Both halves, rather than either alone

**ID:** both-halves
**Status:** active
**Chosen:** per-worktree isolation and a main-checkout UI-test stack — the operator's stated intent
was a separate workspace (application and database) for development, and the ticket's own evidence
is main-checkout pollution; neither half covers the other's failure.
**Considered:** UI-test stack alone — chosen in the first design round and corrected, since it leaves
two concurrent worktrees sharing a schema. Isolation alone — leaves the main-checkout smoke runs
that caused the recorded pollution with nowhere safe to go. A CLI refusal on throwaway repo roots —
guesses at which roots are synthetic, and redirects nothing.

### The database is isolated inside the shared container

**ID:** database-inside-shared-service
**Status:** active
**Chosen:** a separate database in the existing `myflow-postgres` container — the isolation contract
states that what is isolated is the logical resource, never the service holding it.
**Considered:** a second Postgres container — duplicates a service to solve a problem one level in,
and doubles what an operator brings up.

### Reset on every bring-up

**ID:** reset-on-bring-up
**Status:** active
**Chosen:** drop and recreate the test database each time — its contents are then known at the start
of every session, which is the whole value of a separate stack.
**Considered:** a persistent test database wiped on request — recreates the accumulation problem. A
uniquely-named ephemeral database per session — allows two concurrent sessions but orphans a
database whenever teardown is skipped, which is the common case when a session ends unexpectedly.

### `MYFLOW_ADDR` rather than the existing flag

**ID:** myflow-addr-env
**Status:** active
**Chosen:** an environment override for the client's default address — one export directs a whole
session, and the same row serves the isolation section's `url` value.
**Considered:** the existing `-addr` flag — it already worked and is precisely what nobody
remembered to pass. A wrapper script — a second entry point that drifts from the real binary's flags.

### A committed Go fixture

**ID:** go-fixture-seeder
**Status:** active
**Chosen:** a seeder written against the store's exported types — a schema change breaks the build,
which is the earliest and cheapest place to find out.
**Considered:** a SQL fixture loaded by `psql` — drifts silently until the day it does not. A
redacted snapshot of live data — re-imports production content and needs regenerating as the schema
moves.

### `MYFLOW_STATE_DIR` and `MYFLOW_TRANSCRIPTS_DIR` are not isolated

**ID:** paths-not-isolated
**Status:** active
**Chosen:** no rows for either — the contract's `Resource` vocabulary is closed and a directory path
matches none of the five words. Recorded in prose beside the tables, which is where the contract says
a project states what it chose not to isolate.
**Considered:** widening the vocabulary with a `path` word — a contract change far larger than this
one, for a consequence that is bounded (duplicated harvest data, not shared state).

## Open questions

None. Every question this design raised was put to the operator and answered, including the
re-scoping round that widened it from one half to both.
