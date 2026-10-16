# pipeline-fix-kan-954-user-row-loaded-twice-per-request-12

## Context

Every store-backed test in `stats/` gets its database from a helper that runs `CREATE DATABASE`
(copying `template1`) and then `store.RunMigrations` — all 44 embedded migrations, each in its own
transaction with its own commit — under 3-second deadlines
(`stats/internal/store/testsupport_test.go` `newTestDatabase`/`newTestStore`, and five copies:
`internal/reconcile` and `internal/sweep` `newTestStore`, `internal/harvest` `newEndToEndStore`,
`internal/api` `newIntegrationStore`, `cmd/flow` `newTestStoreDSN`). About 200 call sites take that
path. `cmd/uitest-seed`'s helper is left alone: its one test migrates the database itself under a
15-second deadline, as the seed command it tests does.
<!-- measured: ls stats/internal/store/migrations/*.sql | wc -l → 44; grep -rhoE 'newTestStore\(|newTestDatabase\(|newEndToEndStore\(|newIntegrationStore\(|newTestStoreDSN\(' --include='*_test.go' stats | wc -l → 199 @ e777ee03 -->

The live flowd's migration advisory lock is per-database, so it is not the cause. The cause is
setup cost under concurrency. Measured at plan time against `flow-postgres` (Postgres 18.4):
alone, CREATE 63 ms and migrate 567 ms; 24 setups at once, CREATE about 0.95 s and migrate up to
<!-- measured: ad-hoc TestZZTiming (ZZN=1, ZZN=24) in stats/internal/store, removed after; machine-local timings @ e777ee03 -->
2.0 s, so CREATE plus open plus migrate passes 3 s — the deadline the failing tests hit. Copying a
migrated template instead, 24 at once: CREATE 0.63–0.78 s and no migration.
<!-- measured: ad-hoc TestZZTiming (ZZN=1, ZZN=24, ZZN=24 ZZTEMPLATE) in stats/internal/store, removed after; machine-local timings under load average 12–18 @ e777ee03 -->

Reproduced at plan time at e777ee03: `cd stats && go test -count=1 ./...` failed 2 of 2 runs —
`internal/api` (`TestMultiRepoChangeIsOneRowInEveryView`, `TestCostPerChangeAcceptsChangeAlone`),
`internal/reconcile` (four tests) and `internal/store` (`TestPriceFreezesCostAndVersion`), every
failure `run migrations: store: apply migration …: timeout: context deadline exceeded`. With both
tasks' diffs applied: 2 of 2 `go test -count=1 ./...` runs and 2 of 2 `go test -race -count=1
./...` runs green at load averages 17–24, and no database of those runs left behind.
<!-- measured: cd stats && go test -count=1 ./... (and -race), machine-local runs under load average 12–24, cannot be re-run identically @ e777ee03 -->

Seam proof: `TestMigratedDatabasesAreCopiesOfOneTemplate` passes with the fix and fails —
`migration 0018_guard_log.sql applied at … in one database and … in the other` — when
`NewMigratedDatabase` is changed to migrate each database on its own.
<!-- measured: cd stats && go test -count=1 ./internal/storetest/ with NewMigratedDatabase patched to migrate per database, then restored @ e777ee03 -->

Blast radius — the six helper files, every file naming one of the five helper functions outside
`.git`, `node_modules` and the spectre archive (28, the helpers included), plus the three new
`storetest` files — not `big`: 31 files. Only the six helper files and the three new ones are
edited; the call sites keep the helper names.
<!-- measured: grep -rlE 'newTestStore\(|newTestDatabase\(|newEndToEndStore\(|newIntegrationStore\(|newTestStoreDSN\(' --exclude-dir=.git --exclude-dir=archive --exclude-dir=node_modules . | wc -l → 28 @ e777ee03, +3 new files -->

## Decisions

### One migrated template per package process

**ID:** template-per-process
**Status:** active
**Chosen:** `storetest` builds one template database per test process, lazily on the first
migrated database it is asked for (`sync.Once`), and every migrated test database is `CREATE
DATABASE <name> TEMPLATE <template>`. The migration cost is paid once per package instead of once
per test, so the server load that slowed every setup is gone, and the template is always built by
the code under test.
**Considered:** raising the 3-second deadlines — widens a tolerance, and the per-test cost keeps
growing with every migration added; a cross-process lock around database creation — serialises the
same cost, and the queueing time still counts against each deadline; one shared database truncated
between tests — changes the isolation about 200 tests rely on; a persistent template shared across
<!-- measured: grep -rhoE 'newTestStore\(|newTestDatabase\(|newEndToEndStore\(|newIntegrationStore\(|newTestStoreDSN\(' --include='*_test.go' stats | wc -l → 199 @ e777ee03 -->
runs, keyed by a hash of the migration files — a change to `RunMigrations`' own Go code would keep
serving a template built by the old code.

### Deadlines stay 3 seconds, the template build included
<!-- measured: grep -rn 'WithTimeout(context.Background(), 3\*time.Second)' stats/internal/store/testsupport_test.go | wc -l → 3 @ e777ee03 -->

**ID:** deadlines-unchanged
**Status:** active
**Chosen:** setup work steps carry no wall-clock bound; reachability keeps 3 s. Creating, migrating,
copying, opening, seeding and dropping a test database all run on `context.Background()`; only the
reachability ping in `connectAdmin` is bounded, by `reachTimeout` (3 s). How long setup work takes
depends on server load the test does not control — `go test ./...` builds up to GOMAXPROCS templates
in the same second — so a fixed deadline on it races that load: two of four whole-suite runs still
failed at load 20–28 on `migrate flow_template_…: … deadline exceeded`. The bound is not a tolerance
on the code under test: it asserts nothing the specification asks for, and no store defect hides
behind its absence. The one failure it could catch is a hang, which `go test -timeout` still catches,
with a goroutine dump. The reachability bound stays because it is load-independent: a reachable
server answers a ping in milliseconds and a down stack refuses at once.
**Considered:** raising the number for the one-off build, or for every step — widening a timeout to
absorb a slow machine; the race stays and only its odds change. A server-level advisory lock
serialising template builds across processes — the queueing time still counts against each waiter's
deadline, the copies, drops and seeding still race the timer, and it adds a cross-process protocol
to keep the same timer.

### Shared test-support package, not six copies

**ID:** storetest-package
**Status:** active
**Chosen:** a non-test package `stats/internal/storetest` (`AdminDSN`, `DSN`, `Main`,
`NewDatabase`, `NewMigratedDatabase`, `NewStore`), imported only by `_test.go` files; each of the
six helpers keeps its name and body-forwards to it, so no call site changes. The duplicated
`adminDSN`/`testDSN` copies those helpers carried are deleted with them.
**Considered:** applying the template inside each of the six copies — six copies of the template
lifecycle, each a place for it to drift; renaming call sites to `storetest.NewStore` — touches 21
more test files for no behaviour.
<!-- measured: the 28 files of the blast-radius grep minus the six helper files and docs/superpowers/plans/2026-09-12-run-stats.md → 21 @ e777ee03 -->

### Main drops the template and sweeps leftovers

**ID:** main-sweeps-leftovers
**Status:** active
**Chosen:** `storetest.Main(m)`, called from each package's `TestMain`, drops the template once the
tests have run and then every `flow_test_<pid>_<n>` database of this process still present — a test
database whose cleanup drop timed out under load (seen at plan time: test databases from the
failing runs left on the server). The leftover sweep is the cheap fix for that pre-existing leak.
**Considered:** dropping the template only — the leak stays; dropping every `flow_test_*` on the
server — would drop another package process's databases mid-run.

### cmd/uitest-seed stays as it is

**ID:** uitest-seed-untouched
**Status:** active
**Chosen:** `cmd/uitest-seed`'s helper is not switched: its one test checks the seed against an
empty `…_uitest` database it migrates itself under a 15-second deadline, as the seed command does.
**Considered:** switching it too — its database name must end `_uitest` for the seed's own guard, so
it would need a second naming path for one test that never flaked.

### First review on medium

**ID:** first-review-effort
**Status:** active
**Chosen:** the first review, `pipeline-fix-12-review-1`, runs on `flow-medium` (effort `medium`)
on `opus`: a new shared test-support package with a process-level lifecycle (`sync.Once`,
`TestMain`, a sweep that drops databases by pattern) used by six packages.
**Considered:** `low` — the drop-by-pattern sweep and the template lifecycle are worth the deeper
pass.

### One fixer group

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group — Tasks 1 and 2 in series in one `pipeline-fix-12-fix` dispatch; Task 2
depends on Task 1's package and both apply a verified diff.
**Considered:** two groups — a second dispatch for a mechanical diff that cannot start before the
first lands.

## Open questions

## Live check

None: the change touches test support only — no daemon, store schema or runtime state changes.
`flow.verify`'s full `go test ./... -race -count=1` is the check; failure looks like any
`deadline exceeded` in a store-backed package, or `storetest: drop …` lines on stderr.
