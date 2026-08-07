# Per-workspace database, cache, storage and ports

**Change:** `kan-15-parallel-myflow-do-task-lanes`
**Jira:** KAN-15 — "Run independent myflow-do task groups in parallel lanes"
**Date:** 2026-08-06 (fourth revision, same day)
**Status:** approved design, pre-implementation

**The change name names lanes, and this design does not build them.** The name is kept because the
proposal artifact's source path derives from it, and a revision round must republish to the same URL
rather than mint a new one. The mismatch is recorded rather than hidden.

---

## How this change arrived here

Four designs preceded this one. Each is preserved in the change's `design.md` as superseded
decisions rather than deleted, because the reasoning is worth keeping.

1. **In-run lanes joined by patch** — one plan split into concurrent task groups in scratch
   worktrees. Rejected as more complexity than the problem warrants.
2. **Lanes as separate branches with a PR each** — rejected once it was clear that a lane with its
   own branch, worktree, PR, panel, gate and finish run *is* a change.
3. **Workspace isolation with host processes and a parameterized compose stack** — approved, and
   implementation began, then reconsidered in favour of containers.
4. **The dev stack in containers** — four dev Dockerfiles, a compose profile, bind-mounted source
   and volume-backed caches, two run modes. Rejected as too complex for the problem.

**The problem was never the whole stack.** It is that two changes share one Postgres database, so
one change's Flyway migrations land under another change's tests. This design fixes that, and the
two adjacent resources that fail the same way once changes really do run concurrently.

## What is actually shared, and what actually breaks

The data services are **containers shared by every workspace**, and this design leaves them, their
ports and `docker-compose.yml` **entirely untouched**. What it namespaces is the *logical* resource
inside each one.

| Resource | Today | Fails how | This design |
|---|---|---|---|
| Postgres **database** | one, `gymie` | one change's migrations alter the schema another is testing | one database per workspace |
| Redis | one instance, index 0 | sessions and cache mix between changes | one index per workspace |
| MinIO | one bucket, `gymie-media` | uploads mix, and one change's cleanup removes another's | one bucket per workspace |
| App ports | 8080 / 8081 / 3000 / 3001 | a second `devStart` cannot bind at all | one offset block per workspace |
| Mailpit | one inbox | emails interleave | **shared, deliberately** |
| pgAdmin | one UI | — | shared; it can connect to any database |

**Ports are in scope because without them concurrency is not real.** Isolating the database alone
would let a change switch safely between branches, but a second `devStart` would still fail on 8080.
That was raised explicitly during brainstorming and the answer was to include them.

## Constraints established from the repository

- **`application.yml` already parameterises the datasource**: `url: ${DB_URL:jdbc:postgresql://localhost:5432/gymie}`.
  But **`application-local.yml` hardcodes** `jdbc:postgresql://localhost:5432/gymie` with no
  placeholder, and `local` is the profile development runs — so the local profile overrides the hook
  that already exists. That one line is the pivot of the whole gymie change.
- **`dbReset` already takes `-PpgDb`**, defaulting to `gymie`, and already runs
  `DROP DATABASE IF EXISTS … WITH (FORCE)` followed by `CREATE DATABASE "$pgDb"` — correctly quoted.
  The per-database machinery exists; it is only ever pointed at one name.
- **The gateway has no datasource.** Only the backend touches Postgres, which halves the surface.
- **`dbSeed` starts `:app:bootRun --spring.profiles.active=local`**, so it inherits exactly the
  hardcoded URL above.
- **Gymie's integration tests use Testcontainers**, so the test database is already ephemeral and
  per-JVM. The test suite is untouched.

## Goals / Non-Goals

**Goals:**

- Stop one change's migrations from reaching another change's database.
- Let two changes run their apps at the same time.
- Leave the main checkout behaving exactly as it does today.

**Non-Goals:**

- Containerizing the applications. Designed in full and rejected; see the superseded decisions.
- In-run parallelism. `/myflow-do` stays sequential.
- Changing `docker-compose.yml`, the production images, the test suite, or either frontend
  repository.
- Isolating mailpit.

## Design

### 1. The workspace id — unchanged

The id already written and guard-verified in `skills/myflow-contracts/workspace-isolation.md`: a
readable prefix plus the first four hex characters of the SHA-256 of the **full** change name.
`kan-15-parallel-myflow-do-task-lanes` yields `kan-15-55a6`. Deterministic, unrecorded,
uncoordinated. **The main checkout has no id.**

### 2. What the id derives

| Derived | Worktree | Main checkout |
|---|---|---|
| Database | `gymie_kan_15_55a6` | `gymie` |
| MinIO bucket | `gymie-media-kan-15-55a6` | `gymie-media` |
| App ports | offset block from the digest | 8080 / 8081 / 3000 / 3001 |
| Redis index | **probed, not derived** — see below | 0 |

**Underscores in the database name, dashes in the bucket.** `dbReset` already quotes its identifier
so a dashed database name would work, but an unquoted call site anywhere would break on it;
underscores are safe unquoted and cost nothing. Bucket names are not SQL identifiers and take the
id verbatim.

### 3. The Redis index is probed, not derived

Every other value is derived from the digest. The Redis index cannot be, because Redis offers
**16** indices and a digest modulo 16 collides at roughly 6% for two concurrent workspaces — and a
collision silently shares sessions and cache, which is the exact failure class this change exists to
remove.

`devStart` therefore **probes which indices are in use and claims the lowest free one**, recording it
in that workspace's generated `.env` so it is stable for the run. It costs one round-trip at start
and is not stable across restarts, which does not matter: Redis holds sessions and cache, and both
are disposable.

### 4. Isolation is automatic in a worktree

Any apply worktree gets its own database, index, bucket and port block without being asked; the main
checkout always gets the defaults. **Nothing is opt-in**, because the failure being prevented —
running against another change's schema — is one that forgetting a flag would reproduce silently.

### 5. Creation and seeding

`devStart` creates the workspace's database and bucket **if absent**, and says so loudly the first
time, so the side effect is visible rather than silent. The database starts **empty**: Flyway
migrates it, `dbSeed` seeds it. That is the same path a fresh machine and CI take, so a workspace
database is never subtly different from a clean one.

`CREATE DATABASE … TEMPLATE gymie` was considered and rejected: Postgres refuses a template copy
while any session is connected to the source, so it would fail whenever anything is running against
`gymie`, and it would copy that database's migration state, which may not match the branch.

`dbReset` and `dbSeed` **default to the workspace's database** in a worktree. That is both consistent
and safer: `dbReset -Pforce` from a worktree destroys that change's database rather than the shared
one.

### 6. The frontends follow the ports

The generated `.env` currently pins `ADMIN_API_URL=http://localhost:8080`. Every app URL it writes
is derived from the workspace's port block instead, so a worktree's frontends talk to that
worktree's backend. Without this, the frontends would silently reach whichever backend owns 8080 —
the collision the port offsets exist to remove.

### 7. Cleanup

The workspace's database and bucket are dropped at `/myflow-finish` run 2. They get a row in the
temporary artifacts registry and a matching marker in `scripts/check-cleanup-complete.sh`.

**A stopped Postgres is *skipped, not failed*.** An absent container means there is nothing to drop
right now; the run reports it and continues to `FINISHED`. That matches how the existing `## stop`
check already treats an unreachable stack, and a stale database costs a few megabytes — where the
stricter reading would block finishing a merged change because a container happened to be down.

## What this change touches

**agents repository** — `skills/myflow-contracts/workspace-isolation.md` (reduced to the id, the
derived-name rules, the empty-id rule and the probe rule); `project-configuration.md`
(`## workspace isolation`); `skills/myflow-do/SKILL.md` (compute and export the id; the test guide
names the worktree's URLs); `pipeline.md` (one registry row);
`scripts/check-cleanup-complete.sh` (one marker and its check); the four contract index rows.

**gymie repository** — `src/app/src/main/resources/application-local.yml` (the placeholders);
`gradle/dev-lifecycle.gradle.kts` (derive the block, probe Redis, create database and bucket, port
the `ManagedService` ports, generate the `.env` URLs, default `pgDb`); `.myflow/project.md`.

**`docker-compose.yml` is not touched**, and neither frontend repository is. Two worktrees, two
review panels, two pull requests.

## Risks / Trade-offs

**The Redis probe is racy between two simultaneous starts** → two `devStart`s launched in the same
second could claim the same index. The window is small and the recovery is to restart one, but it is
a real gap and no lock is proposed; a lock would be more machinery than the failure justifies.

**Mailpit stays shared** → two changes' emails interleave in one inbox. Confusing when testing email
flows concurrently, not corrupting, and mailpit has no natural namespace to use.

**A digest collision would share a database** → two concurrent changes whose full names collide in
four hex characters. Far less likely than the Redis case, and widening the digest is a one-line
change if it ever bites.

**Two app stacks cost real machine resources** → two JVMs, two node processes, two Gradle daemons.
Nothing schedules or limits them.

**The port block is bounded by what is free** → the block is checked before use and falls back to
discovery as a whole, per the existing contract rule; a machine with many workspaces running will
eventually run out of predictable blocks.

**`application-local.yml` gaining placeholders changes a file every developer uses** → but every
placeholder carries today's value as its default, so a checkout with nothing set behaves exactly as
before.

## Migration Plan

Nothing changes in the main checkout: same database, same ports, same bucket, same Redis index. A
worktree becomes isolated automatically the first time `devStart` runs there, creating its database
and bucket and saying so. There is nothing to roll back beyond reverting the change.

## Open questions

None. Every question raised across four brainstorming sessions was put to the operator and answered
before this design was written.

## Success criteria

1. `devStart` in the main checkout uses `gymie`, Redis index 0, `gymie-media` and ports
   8080 / 8081 / 3000 / 3001 — unchanged.
2. `devStart` in an apply worktree creates `gymie_<id>` and `gymie-media-<id>` if absent, says so,
   claims a free Redis index, and binds its own port block.
3. Flyway migrates the workspace database from empty and `dbSeed` seeds it.
4. Two worktrees can run their apps simultaneously, each reaching its own database, index, bucket
   and backend.
5. The generated `.env` in a worktree names that worktree's backend port.
6. `dbReset -Pforce` in a worktree targets that workspace's database, not `gymie`.
7. `./gradlew test` passes unchanged, using Testcontainers.
8. `/myflow-finish` run 2 drops the workspace database and bucket; a stopped Postgres is reported
   and skipped rather than blocking `FINISHED`.
9. `docker-compose.yml` is byte-identical to its pre-change state.
