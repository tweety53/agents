## Why

Two myflow changes share one Postgres database. One change's Flyway migrations alter the schema
another change is testing against, and the failure looks like a real test result rather than a
collision. That is what KAN-15's addition names: *"if I run multiple myflow tasks in parallel, the
runned app/storage etc. should be isolated too, so they won't interact and break each other
behavior (for example when db schema is different between them it can produce false positive
testing results)."*

Two adjacent resources fail the same way as soon as changes really do run at once. **Redis** holds
sessions and cache in one index, so two backends mix them. **MinIO** has one bucket, so uploads mix
and one change's cleanup removes another's. And the applications bind fixed ports — 8080, 8081,
3000, 3001 — so a second `devStart` cannot start at all, which is what makes concurrency impossible
today regardless of what the database is called.

**The change name still says "lanes", and this change does not build them.** An in-run lane
mechanism was designed and rejected, then a containerized dev stack was designed and rejected. The
name is kept because the proposal artifact's source path derives from it and a revision round must
not mint a new URL. Every superseded design is preserved in `design.md` rather than deleted.

## What Changes

- **A workspace id, derived from the change name** — a readable prefix plus the first four hex
  characters of the SHA-256 of the full name, e.g. `kan-15-55a6`. Deterministic, so the same change
  always gets the same id; unique, so two changes sharing a Jira key do not collide. Nothing is
  recorded and no session coordinates with another.
- **A database per workspace** — `gymie_kan_15_55a6`. Underscored, so no call site has to quote it.
- **A MinIO bucket per workspace** — `gymie-media-kan-15-55a6`.
- **An app port block per workspace** — backend, gateway, admin **and the KMP frontend**, offset
  from the same digest, checked free before use. All four move; none is an exception. The frontend's
  port was briefly treated as immovable, on a claim that its dev server offered no override outside
  its own repository — see `kmp-frontend-port-rotates` in `design.md` for why that was wrong and what
  believing it cost.
- **A Redis index per workspace, probed rather than derived.** Redis offers only 16 indices, and a
  digest modulo 16 collides at roughly 6% for two concurrent workspaces — silently sharing sessions
  and cache, which is the failure class this change removes. `devStart` claims the lowest free index
  and records it for the run.
- **Isolation is automatic in any apply worktree**, and the main checkout keeps `gymie`, index 0,
  `gymie-media` and today's ports. Nothing is opt-in, because forgetting a flag would silently
  reproduce the bug.
- **`devStart` creates the database and bucket if absent** and says so loudly the first time. The
  database starts empty; Flyway migrates it and `dbSeed` seeds it — the same path a fresh machine
  takes.
- **`dbReset` and `dbSeed` default to the workspace's database** in a worktree, which is both
  consistent and safer: `dbReset -Pforce` there destroys that change's data, not the shared
  database.
- **The generated `.env` derives every app URL from the workspace's ports**, so a worktree's
  frontends reach that worktree's backend rather than whichever one owns 8080.
- **`/myflow-finish` run 2 drops the workspace's database and bucket**, with a registry row and a
  cleanup-guard marker. A stopped Postgres is **skipped, not failed** — an absent container means
  nothing to drop, and a stale database costs a few megabytes where the strict reading would block
  finishing a merged change.
- **`docker-compose.yml` is not changed at all.** The data services stay shared containers on their
  current ports; only the logical resources inside them are namespaced.
- **The four dev commands act on a chosen subset of the services.** `devStart`, `devRestart`,
  `devStop` and `devStatus` take `-Pservices=<names>`, comma-separated, from `backend`, `gateway`,
  `frontend` and `admin-frontend`. Omitting it means all four, so nothing an operator types today
  changes. `devStart` already stops before it starts, so naming a subset is how a single service is
  restarted — which is what makes the flag worth having: a backend restart currently drags a
  Kotlin/JS compile along with it.
- **A workspace records the ports it resolved, rather than inferring them again later.** A run that
  binds every service writes down the block it resolved; a partial start reads that record instead
  of recomputing the block and guessing which services have since moved off it. A recorded value can
  go stale, and that is the point of preferring it — a stale record is visible and deletable, where
  an inference over a pid file, a port file and `lsof` announces a guess as an observation. Three
  fix rounds were spent making that inference precise before it was deleted instead.
- **A partial start does not disturb the workspace's port block.** The block-free check treats a
  port held by this checkout's own running service as free rather than as a clash, since under a
  partial start the holder is the stack the block belongs to. Foreign contention while this
  checkout has services running is **refused** rather than resolved by discovering a new block: a
  started service on discovered ports while its siblings hold the old ones is a split stack that
  reports one set of ports and serves on another.
- **NOT BREAKING.** Every default reproduces current behaviour.

## Capabilities

### New Capabilities

- `myflow-workspace-isolation`: how a workspace id is derived, what it derives (database, bucket,
  ports) and what it deliberately does not, why the Redis index is probed rather than derived, the
  empty-id rule, the automatic-in-a-worktree rule, the creation and seeding path, the cleanup rule,
  and the `## workspace isolation` project configuration key.

### Modified Capabilities

None. The registry row this change adds is an instance of `myflow-finish-cleanup`'s existing "Every
artifact the pipeline creates is enumerated in one registry" requirement rather than a change to it.

## Impact

**agents repository** — `skills/myflow-contracts/workspace-isolation.md`;
`skills/myflow-contracts/project-configuration.md`; `skills/myflow-do/SKILL.md`;
`skills/myflow-contracts/pipeline.md`; `scripts/check-cleanup-complete.sh`; the contract index rows
in `skills/myflow-contracts/SKILL.md`, `CLAUDE.md`, `AGENTS.md` and
`rules/myflow-manual-review.mdc`.

**gymie repository** — `src/app/src/main/resources/application-local.yml`, which currently hardcodes
`jdbc:postgresql://localhost:5432/gymie` and so overrides the `${DB_URL:…}` placeholder
`application.yml` already has; `gradle/dev-lifecycle.gradle.kts`; `.myflow/project.md`; the pure
decisions the lifecycle script defers to, under `buildSrc/src/main/kotlin/com/gymie/gradle/`, each
with its own test; and `README.md`, whose dev-lifecycle section documents the commands.

**gymie-frontend repository** — one file, `webApp/webpack.config.d/`, a handful of lines, giving the dev
server's port an environment override beside the `historyApiFallback` snippet already there.

**Not touched** — `docker-compose.yml`, the production Dockerfiles, the helm chart, the test suite
and its Testcontainers setup, and the admin frontend repository.

Three worktrees, three review panels, three pull requests — the third being a single small file that
removes a whole class of defect from the second. The machinery it retires is 469 lines.
<!-- measured: wc -l buildSrc/src/{main,test}/kotlin/com/gymie/gradle/SharedPort*.kt @ the gymie apply worktree, staged -->

No new runtime dependency. `psql`, the MinIO client and Redis are already in the stack.
