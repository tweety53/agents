## Context

A standalone self-review fix loop over one finding, KAN-914, in worktree
`/Users/tweety53/Projects/agents-worktrees/self-review-kan-914` on branch `self-review-kan-914`,
based at `b76c2cf1` (origin/main). The brainstorm design is
`.superpowers/sdd/2026-10-21-self-review-kan-914-design.md` in this worktree.

**The defect reproduces at b76c2cf1.** A scratch store test on a fresh migrated database:
`RecordGuardRun(ctx, "ws-0000beef", records.GuardRun{…})` returned
`store: project not found: ws-0000beef`, `errors.Is(err, store.ErrProjectNotFound)` true. Eight
migrations declare `REFERENCES projects(project_key)`; `ensureProject` is called from `PutChange`
and `insertPlanStageRun` only.
<!-- measured: git grep -n 'REFERENCES projects' -- stats/internal/store/migrations | wc -l @ b76c2cf1 -->

**Where the row can come from.** The schema exists only once flowd has run its migrations at
startup (`st.RunMigrations` in `stats/cmd/flowd/main.go`), so `workspace.sh create`, which runs
`createdb` and nothing else, has no table to write into. The record writes carry a project key
but no main checkout path, and `projects.main_checkout_path` is NOT NULL. The daemon itself
already reads its working directory's checkout (`originRemote`), and
`fallback.ProjectKey(dir)` derives both the key and the main checkout from it — the same
derivation every `flow` CLI write uses.

**Live spike.** `scripts/workspace.sh create kan-914-spike`, then the planned change built from
this branch as `flowd` on port 4914 over `flow_kan_914_spike`, run from `stats/`: `select
project_key, main_checkout_path from projects` returned `agents-a740d89c|/Users/tweety53/Projects/agents`.
The daemon was stopped and the database removed (`workspace.sh remove kan-914-spike`; `survivors`
printed nothing). The protected `flowd` on 4173 and the `flow` database were not touched.

**Blast radius:** 4 files — `stats/cmd/flowd/main.go`, `stats/cmd/flowd/wiring_test.go`,
`stats/internal/store/projects.go`, `stats/internal/store/projects_test.go`.
<!-- measured: git diff --stat of the plan-time spike, applied then reverted @ b76c2cf1 -->

## Decisions

### A workspace daemon seeds its own project row at startup

**ID:** seed-at-workspace-daemon-startup
**Status:** active
**Chosen:** flowd, after its migrations, derives its working directory's project key and main checkout with `fallback.ProjectKey(".")` and creates the row with a new `Store.EnsureProject`, which runs the existing `ensureProject` in its own transaction — one place every record write against that database then routes through.
**Considered:** seeding in `scripts/workspace.sh create` — the database is empty until flowd migrates it at startup, so there is no `projects` table to insert into, and running migrations from the script would duplicate flowd's migrator; having every record write upsert the row — their payloads carry no main checkout path, the column is NOT NULL, and `ensureProject` refuses an empty path, so seven writers would each need a new field; seeding from the `flow` CLI on a refused write — a retry path in every record command for what one startup line covers.

### Only a non-UI-test workspace database is seeded

**ID:** seed-scope-workspace-only
**Status:** active
**Chosen:** seed when the DSN's database path starts `/flow_` and does not end `_uitest` — `workspace.sh`'s own `flow_<id_underscored>` naming, minus the UI-test suffix `cmd/uitest-seed` and `workspace.sh remove --force` already key on.
**Considered:** every daemon — the dev daemon on `flow` would gain a startup write it never needed, since its writers bootstrap rows; `!replaysSharedJournal(dsn)` alone — includes `flow_<id>_uitest`, whose rows `cmd/uitest-seed` owns, and an extra `ProjectRoots` entry would put the real checkout's lessons into the UI-test stack's views.

### A failed seed is logged, never fatal

**ID:** seed-failure-logged
**Status:** active
**Chosen:** a working directory outside a git checkout, or an `EnsureProject` error, logs one `Warn` line and the daemon starts as before — the state the fix replaces, never a worse one.
**Considered:** refusing to start — a daemon started from outside a checkout would stop serving reads it serves today.

### Nothing to repair

**ID:** no-repair
**Status:** active
**Chosen:** no repair command — the defect refused writes rather than writing wrong rows, and every affected database is a throwaway `flow_<id>` that cleanup drops.
**Considered:** back-filling `projects` rows into live workspace databases — none outlives its change, and a re-started daemon seeds its own.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — one additive store method over an existing bootstrap and one startup branch gated by a table-tested predicate that keeps the dev daemon's path unchanged.
**Considered:** `flow-medium` — warranted for a change to a path every daemon start or every write takes; this one runs only on a workspace database.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-914-fix`, Task 1.
**Considered:** none — the plan has one task.

## Open questions

## Live check

After the fix commit, from `stats/` in this worktree: `scripts/workspace.sh create kan-914-live`,
`make -s web-dist && go build -o <scratch>/flowd ./cmd/flowd`, start it with
`FLOWD_PORT=4914 FLOWD_DSN='postgres://flow:flow@localhost:5433/flow_kan_914_live?sslmode=disable'`
and scratch `FLOW_TRANSCRIPTS_DIR`/`FLOW_STATE_DIR`, wait for `/api/v1/changes` to answer, then
`docker exec flow-postgres psql -U flow -d flow_kan_914_live -tAc 'select project_key, main_checkout_path from projects'`.
Before the fix that query prints nothing; after it, one row,
`agents-<8 hex>|/Users/tweety53/Projects/agents`. Not working looks like an empty result or a
`flowd did not seed the workspace project row` line in the daemon's log. Stop that daemon by its
own pid and run `scripts/workspace.sh remove kan-914-live`; `survivors kan-914-live` prints
nothing. Never the `flowd` on 4173, the `flow-postgres` container or the `flow` database.
