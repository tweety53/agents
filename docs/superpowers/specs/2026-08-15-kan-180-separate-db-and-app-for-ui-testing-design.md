# A separate database and app instance for myflow UI testing

**Change:** `kan-180-separate-db-and-app-for-ui-testing`
**Jira:** KAN-180
**Date:** 2026-08-15

## The problem

UI testing of the myflow stats SPA runs against the same PostgreSQL database and the same `myflowd`
instance that hold real recorded stats. There is no second stack to point at, so every manual or
agent-driven smoke run writes permanent rows into production-equivalent data.

The evidence, read out of the live database on 2026-08-15 before it was wiped: eight of the ten rows
in `projects` were test artifacts — `smoke`, `proj`, `proj1`, `reviewproj` at `/a/repo/one`,
`review-repo-d9a8a557`, `verify-f1-1786665626`, `myflow-status-test-repo-1388118d` and
`myflow-status-live-386682ad`, several naming `/tmp` paths that no longer exist. Two synthetic
`changes` rows, `k174-live-test` and `k174-live-test2`, sat at `STARTED`, written by
`myflow stage begin`'s synthetic-change bootstrap. All of it sat beside the genuine
`agents-a740d89c` and `gymie-7c1f238a` projects and the real KAN-167, KAN-172 and KAN-174 changes,
with nothing in the schema distinguishing one from the other. The only remedy available was a full
wipe, which destroyed the real history along with the junk.

## What is not the problem

**The Go test suite already isolates.** Every store-backed test helper creates its own per-test
database — `internal/store/testsupport_test.go`, `internal/harvest/endtoend_test.go`,
`internal/reconcile/testsupport_test.go`, `internal/sweep/testsupport_test.go`,
`internal/api/stats_test.go` and `cmd/myflow/journal_test.go` all build a DSN naming a database they
create for the test and drop afterwards. `go test ./...` never writes to the shared `myflow`
database, and nothing in this change alters that.

The pollution came from the other direction: the `myflow` CLI defaults to
`-addr http://127.0.0.1:4173` (`cmd/myflow/state.go`), and the SPA is served by that same live
daemon. A smoke run, a status check against a throwaway repository, or an agent exercising the UI
therefore lands in production data by default, and the operator has no alternative target to name.

## The two halves

The change has two halves, addressing two different failures with one shared mechanism.

**Per-worktree isolation** is for *development*: a `/myflow-do` apply worktree gets its own database
and its own port, derived from the change name, so two changes in flight never share schema or
collide on a port, and neither touches the main checkout's database.

**The UI-test stack** is for *ad-hoc testing from the main checkout*: a named second stack to point
smoke runs at. Worktree isolation would not have prevented what actually polluted the database —
the junk rows naming `/tmp` paths were written by runs against the main checkout's daemon on 4173,
not from inside any worktree. Conversely a single test stack does nothing for two concurrent worktrees.

`MYFLOW_ADDR` is the mechanism both halves need: it is what points the CLI at the UI-test stack, and
what points it at a worktree's own daemon.

## The design — the UI-test stack

A second, complete myflow stack running alongside the live one. It shares the `myflow-postgres`
container and nothing else.

| Resource | Live stack | UI-test stack |
|----------|-----------|---------------|
| Database | `myflow` | `myflow_uitest`, in the same container on host port 5433 |
| Daemon port | 4173 | 4174 |
| Transcripts root | the real Claude projects directory | an empty temporary directory |
| Fallback state directory | the real per-project state directory | a temporary directory |

**The database is isolated inside the shared service rather than by adding a second service.** This
follows `skills/myflow-contracts/workspace-isolation.md`, which states that what is isolated is the
logical resource — a database inside the shared database server — never the service that holds it.
A second Postgres container would duplicate a service to solve a problem that lives one level in,
and would double what an operator has to remember to bring up.

**The transcripts root and state directory are pointed at temporary directories, not merely left
alone.** This is the difference between a test stack that starts clean and one that immediately
refills itself. `myflowd` runs a harvest loop every five seconds against
`harvest.DefaultTranscriptsRoot()`; pointed at the real directory, a freshly created test database
would begin ingesting the operator's genuine session transcripts within seconds of starting, which
reproduces the very mixing this change exists to prevent. The same argument applies to the fallback
state directory, whose records a test run would otherwise write beside the real ones.

### What already exists

Most of the mechanism is present. `internal/config.FromEnv` already resolves `MYFLOWD_HOST`,
`MYFLOWD_PORT` and `MYFLOWD_DSN`, refusing an unparsable port rather than falling back silently.
`harvest.DefaultTranscriptsRoot` already honours `MYFLOW_TRANSCRIPTS_DIR`, and the fallback state
root already honours `MYFLOW_STATE_DIR`. The daemon can therefore be moved onto a different port,
database and transcript root today, with no code change at all.

### What is added

**1. `MYFLOW_ADDR`, an environment override for the CLI's default daemon address.**

`cmd/myflow` carries `defaultAddr` as a constant and passes it as the default of the `-addr` flag at
three sites — twice in `state.go`, once in `stage.go`. A single `resolveDefaultAddr()` helper reads
`MYFLOW_ADDR`, falling back to the existing constant when it is unset, and the three flag
registrations take their default from that helper instead.

The precedence is: an explicit `-addr` flag wins over `MYFLOW_ADDR`, which wins over the built-in
default. This mirrors the daemon's own resolution and means a whole test session can be targeted by
exporting one variable, rather than by remembering a flag on every invocation. A forgotten flag is
what put the junk in the database in the first place, so removing the need to remember it is the
point rather than a convenience.

**2. A committed fixture seeder, `stats/cmd/uitest-seed/`.**

A test database that is merely empty renders an empty UI, which cannot exercise the views. The
seeder writes a small fixed fixture through the store's own types: two projects, changes spanning
`STARTED`, `IN_PROGRESS` and `FINISHED`, and stage runs carrying token usage so the cost and
statistics views have something to display.

Writing it in Go against the store's exported API, rather than as a SQL file loaded by `psql`, is a
deliberate choice. A SQL fixture drifts silently when a migration changes a column: it keeps loading
until the day it does not, and the failure surfaces as a confusing UI rather than a build error. A
Go seeder compiled against the same types the daemon uses stops compiling the moment the shape
changes, which is the earliest and cheapest place to find out.

**3. `make ui-test-up` and `make ui-test-down` in `stats/Makefile`.**

`ui-test-up` drops and recreates `myflow_uitest`, starts `myflowd` on port 4174 with the temporary
transcript and state roots exported, waits for it to answer, and seeds the fixture. Migrations and
the pricing seed run as part of daemon startup, exactly as they do for the live stack — the test
database is brought up the same way a fresh machine or a clean CI job would bring one up, never by
copying an existing database.

Resetting on every bring-up is what keeps the stack trustworthy. A persistent test database
accumulates the same way the live one did, and the whole value of the second stack is that its
contents are known at the start of every session.

`ui-test-down` stops the daemon and drops the database.

**4. Documentation** in `stats/README.md` and a note under `## run` in `.myflow/project.md`, so the
stack is discoverable from the two places an operator or an agent actually reads.

### The name guard

Both the seeder and the teardown refuse to act unless the database named by their target DSN ends in
`_uitest`. The check runs before any statement is issued.

This is what turns the acceptance criterion from a convention into a structural property. Without
it, "do not point this at the live database" is advice, and advice is exactly what failed here
already. With it, pointing the seeder or the teardown at `myflow` is a refusal with a named reason,
not a discovery made afterwards by reading `projects`.

The guard belongs on the destructive and writing paths specifically. The daemon itself is not
guarded this way — it must be able to serve the live database, which is its normal job.

## The design — per-worktree isolation

`.myflow/project.md` gains a `## workspace isolation` section, declared exactly as
`skills/myflow-contracts/project-configuration.md` specifies. Every variable it names is one the
code already reads, so the section declares behaviour rather than requesting new plumbing.

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| `database` | `MYFLOWD_DSN` | `postgres://myflow:myflow@localhost:5433/myflow?sslmode=disable` | `postgres://myflow:myflow@localhost:5433/myflow_<id_underscored>?sslmode=disable` |
| `port` | `MYFLOWD_PORT` | `4173` | `+<offset>` |
| `url` | `MYFLOW_ADDR` | `http://127.0.0.1:4173` | `http://127.0.0.1:<value:MYFLOWD_PORT>` |

Each `Default` is today's literal value, which is the column where the backwards-compatibility
promise is kept: with no workspace id, every variable resolves to exactly what the main checkout
uses now.

The `url` row is the second place `MYFLOW_ADDR` earns its place. The same variable that points the
CLI at the UI-test stack points it at a worktree's own daemon, so the two halves of this change share
one mechanism rather than each inventing its own.

### What is deliberately not isolated

`MYFLOW_STATE_DIR` and `MYFLOW_TRANSCRIPTS_DIR` get no rows. The contract's `Resource` vocabulary is
closed — `database`, `bucket`, `cache index`, `port`, `url` — and a directory path is none of them.
This is recorded in prose beside the tables, which is where the contract says a project states what
it has chosen not to isolate, so the absence reads as a decision rather than an oversight.

The practical consequence is bounded. A worktree's daemon harvests the same real transcripts the
main checkout's does, but into its own database, so the data is duplicated rather than shared. The
fallback state directory is written only when the daemon is unreachable, and is keyed per project.
Neither crosses the isolation boundary this change is drawn to protect.

There is no `bucket` row and no `cache index` row: the project has no object store and no cache.

### The workspace commands

`scripts/workspace.sh`, with `create`, `remove` and `survivors` subcommands, named in the section's
command table. The contract fixes the working directory of each — `create` runs from the apply
worktree, because whatever starts the applications calls it; `remove` and `survivors` run from the
main checkout, because run 2 calls them after the worktree is gone.

Two properties are load-bearing rather than stylistic, and both come from the contract:

- **`survivors` filters with `awk`, never `grep`.** `grep` exits 1 when nothing matched, and a
  non-zero exit is read as *the check could not run*. A project whose empty report is produced by a
  `grep` can therefore never reach the one result that verifies its cleanup — the removal would be
  reported as skipped forever.
- **`survivors` carries its own timeout inside the container.** The guard's sixty-second bound
  terminates a process group on the host, and work behind `docker exec` is not in that group, so a
  hung query outlives the bound unless the command bounds itself.

### Correcting a stale claim

`skills/myflow-contracts/project-configuration.md` and
`skills/myflow-contracts/workspace-isolation.md` both cite the agents repository as the worked
example of a project with nothing to isolate — "it declares no runnable application". That stopped
being true when `stats/` landed. Adopting the section here makes those sentences wrong, so they are
corrected as part of this change rather than left to mislead the next reader.

## Decisions

The choices this design rests on, and what was considered against each:

- **Both halves, rather than either alone.** The first round of this design scoped to the UI-test
  stack only; the operator clarified that the intent included a separate workspace — application and
  database — for development, which is the workspace-isolation contract. The two address different
  failures and neither subsumes the other: worktree isolation would not have prevented the
  main-checkout smoke runs that actually caused the pollution, and a single test stack does nothing
  for two concurrent worktrees.
- **A separate database inside the shared container, rather than a separate container.** Required by
  the workspace-isolation contract's rule about logical resources, and it halves what an operator
  brings up.
- **Reset on every bring-up, rather than a persistent test database.** A persistent one re-creates
  the accumulation problem; a uniquely-named ephemeral one per session would allow two concurrent
  sessions but leaves orphaned databases behind whenever a teardown is skipped, which is the common
  case when a session ends unexpectedly.
- **`MYFLOW_ADDR`, rather than the existing `-addr` flag or a wrapper script.** The flag already
  works and is exactly what nobody remembered to pass. A wrapper script is a second entry point that
  drifts from the real binary's flags.
- **A committed Go fixture, rather than an empty database or a snapshot of real data.** An empty
  database cannot exercise the views. A snapshot re-imports production content and has to be
  regenerated as the schema moves.

## Open questions

None. Every question this design raised was put to the operator and answered.

## Testing

- `resolveDefaultAddr` — unset, set, and the precedence rule that an explicit `-addr` flag overrides
  `MYFLOW_ADDR`.
- The name guard — refuses a DSN naming `myflow`, accepts one naming a `_uitest` database, on both
  the seeder and the teardown path.
- The seeder — exercised against a per-test database created by the existing helper pattern, so the
  test itself touches neither real database.
- `scripts/workspace.sh` — a Bash test in the repository's existing `scripts/test-*.sh` style,
  covering the round trip (`create` then `survivors` reports the database, `remove` then `survivors`
  is empty and exits 0), the empty-report-exits-0 rule that `awk` rather than `grep` is there to
  satisfy, and the refusal to remove a database whose name does not carry a workspace id.
- `scripts/check-workspace-isolation.sh` against this repository's own configuration, which until now
  has passed by declaring nothing and will from now on actually validate the new section.

Existing suites must stay green: `go test ./... -race -count=1`, `npx tsc -b` in `stats/web`, and
the repository's guard scripts.
