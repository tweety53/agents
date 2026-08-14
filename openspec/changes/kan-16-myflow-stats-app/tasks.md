> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Replace myflow's per-machine JSON state file with a PostgreSQL-backed Go service, record
per-stage telemetry for every `/myflow-*` run from the harness's own session transcripts, and serve
both the live pipeline state and aggregated statistics through a browser interface — without the
pipeline ever blocking on any of it.

**Architecture:** A new top-level `stats/` directory holding a Go module: a `myflowd` daemon that
owns the PostgreSQL pool and serves the state API, the statistics API and an embedded React SPA; a
thin `myflow` CLI the skills shell out to; an `internal/store` package that is the only thing that
knows SQL; and an `internal/harvest` package that attributes real token usage to stage windows. Full
design in `design.md`.

**Tech Stack:** Go 1.26.5, PostgreSQL 18.6 (`postgres:18-alpine`), React 19.2.8, Vite 8.2.1.
Verification is `go test ./...` inside `stats/`, the SPA's own test run, plus this repository's
existing prose guards for the contract and skill edits.

## Global Constraints

- **The pipeline never blocks on this subsystem.** Every CLI path that touches the store falls back
  to the on-disk JSON plus a journal entry and exits 0. A task that introduces a store call without
  that fallback is incomplete.
- **`internal/store` is the only package that builds SQL.** The daemon calls typed repository
  methods; it never assembles a query.
- **No metric is ever recorded as zero to stand in for "not measured."** Absence and zero are
  distinct facts and stay distinct through the schema, the API and the interface.
- **Stage names come from `README.md`'s Level 1 table.** No skill invents one.
- **The three pipeline states, the command surface, the git boundaries, the review panel and the
  Jira contract are untouched.** This change makes the pipeline observable; it does not change what
  it does.
- **No task edits `openspec/` or `docs/superpowers/`** — `/myflow-finish` commits those.

## Baseline

The `stats/` directory does not exist, so every Go and TypeScript test in this plan is new. The Go
toolchain is absent from the development machine.

**Every `**Baseline:**` figure in this plan counts _top-level_ test functions**, measured as
`cd stats && go test ./... -count=1 -v | grep -c '^--- PASS'`. Counting `PASS:` instead includes
subtests and yields a substantially larger number — the two measures diverge by roughly a third once
table-driven tests appear, and comparing one against the other reads as a stale baseline when
nothing is wrong. Use the command above.

<!-- measured: `which go` returned nothing, and `ls stats` did not exist, on 2026-08-13; a machine-local observation that cannot be re-run from a commit -->

| Measure | Now | After this change |
|---------|-----|-------------------|
| Go toolchain | absent | 1.26.5 installed |
| `stats/` | absent | a Go module with an embedded SPA |

---

### 1 Prerequisites — Go toolchain, the module skeleton, and the dedicated Postgres stack

**Build:** green

**Files:**
- Add: `stats/go.mod`
- Add: `stats/go.sum`
- Add: `stats/health_test.go`
- Add: `stats/docker-compose.yml`
- Add: `stats/README.md`
- Add: `stats/Makefile`
- Modify: `.gitignore`

**Allowed-collateral:** `stats/go.sum`

**Interfaces:**
- Consumes: nothing.
- Produces: a buildable Go module and a running PostgreSQL every later task tests against.

- [x] **Step 1: Install Go 1.26.5**

Install with `brew install go`, which resolved to exactly 1.26.5 on this machine. Confirm with
`go version` before proceeding — every later task's verification command depends on it.

- [x] **Step 2: Initialise the module**

```bash unverified:confirm the module path suits the repository's remote once the first import lands
cd stats && go mod init github.com/tweety53/agents/stats
```

- [x] **Step 3: A dedicated PostgreSQL, independent of the gymie stack**

A separate container, its own volume, and a port that does not collide with the running
`gymie-postgres` on 5432.

```yaml verified:brought up with `docker compose up -d` against postgres:18-alpine on 2026-08-13; port 5433 confirmed free beforehand
services:
  myflow-postgres:
    image: postgres:18-alpine
    container_name: myflow-postgres
    environment:
      POSTGRES_USER: myflow
      POSTGRES_PASSWORD: myflow
      POSTGRES_DB: myflow
    ports:
      - "5433:5432"
    volumes:
      # Postgres 18 images mount at /var/lib/postgresql, NOT the older
      # /var/lib/postgresql/data — the image places a version-specific
      # subdirectory under it and refuses to start against the old path.
      - myflow-pgdata:/var/lib/postgresql
volumes:
  myflow-pgdata:
```

- [x] **Step 4: Ignore build output**

Add `stats/bin/`, `stats/web/node_modules/` and `stats/web/dist/` to `.gitignore`.

**Tests:** `stats/health_test.go` — one test asserting the module compiles and that a connection to
the compose stack's PostgreSQL succeeds, skipped with a clear message when the stack is not running.

**Regression:** Reverting this task removes the module and the compose file, so `go test ./...`
inside `stats/` fails to find a module and every later task's verification is unrunnable.

**Baseline:** before=0 after=1 Go tests.
<!-- predicted: `cd stats && go test ./... 2>&1 | grep -c ^ok` after task 1 -->

**Commit:** `feat(kan-16-myflow-stats-app): add the stats Go module, its Postgres stack, and the Go toolchain prerequisite`

---

### 2 `internal/store` — schema, migrations, and the change repository

**Build:** green

**Files:**
- Add: `stats/internal/store/store.go`
- Add: `stats/internal/store/migrations.go`
- Add: `stats/internal/store/migrations/0001_init.sql`
- Add: `stats/internal/store/changes.go`
- Add: `stats/internal/store/store_test.go`
- Add: `stats/internal/store/changes_test.go`
- Add: `stats/internal/store/testsupport_test.go`
- Modify: `stats/Makefile`

**Interfaces:**
- Consumes: task 1's module and PostgreSQL.
- Produces: `store.Store` with `GetChange`, `PutChange`, `ListChanges`, and the migration runner
  every later store task extends.

- [x] **Step 1: The `projects` and `changes` tables**

Exactly the DDL in `design.md`'s "Data model" section for those two tables. Migrations are embedded
with `go:embed` and applied in lexical filename order, recorded in a `schema_migrations` table so a
second run is a no-op.

- [x] **Step 2: The change repository**

`PutChange` renders the whole record — every field, never a partial merge — matching the state file
contract's write rule. It enforces the monotonic-state rule in SQL, refusing a write whose state is
earlier in the pipeline than the stored one, and returns a typed error the caller can distinguish
from a transport failure.

- [x] **Step 3: Test against a real PostgreSQL**

`testsupport_test.go` connects to the compose stack, creates a uniquely-named database per test run
and drops it afterwards. No mock: the JSONB round-trip and the unique constraint are the things
under test, and a mock asserts neither.

**Tests:** `TestMigrationsAreIdempotent`, `TestPutChangeRoundTripsEveryField`,
`TestPutChangeRefusesMonotonicViolation`, `TestPutChangeIsWholeObject`,
`TestListChangesFiltersByProject`, `TestSameNameInTwoProjectsCoexist`,
`TestGetChangeUnknownReturnsNotFound`, plus one test per review finding fixed in this task's fix
round.

**Regression:** Reverting this task removes the schema and the change repository, so the daemon in
task 4 has nothing to read or write and `TestPutChangeRefusesMonotonicViolation` cannot run.

**Baseline:** before=1 after=12 Go tests.
<!-- predicted: `cd stats && go test ./... -count=1 -v | grep -c ^=== RUN` after task 2 -->

**Commit:** `feat(kan-16-myflow-stats-app): add the store schema, migrations, and the change repository`

---

### 2.1 `internal/store` — a change is one unit of work across every repository it touches

**Build:** green

**Files:**
- Add: `stats/internal/store/migrations/0002_change_repos.sql`
- Add: `stats/internal/store/repos.go`
- Add: `stats/internal/store/repos_test.go`
- Modify: `stats/internal/store/changes.go`
- Modify: `stats/internal/store/changes_test.go`
- Modify: `stats/internal/store/store_test.go`

**Interfaces:**
- Consumes: task 2's `store.Store` and change repository.
- Produces: a `Repos` field on `Change`, written transactionally by `PutChange`; `ListChangeRepos`;
  and the `change_repos` rows task 3's `stage_runs` references and task 11's views break down by.

  **No standalone exported `PutChangeRepos`**, deliberately. The repository set must be written in
  the same transaction as the change row, and an independently-callable exported method could honour
  that only by opening its own transaction — which would not be the change's — or by taking a
  transaction handle as a parameter, which leaks the driver into the package's public surface. The
  set is therefore written through `PutChange` and an unexported helper. This was found while
  implementing rather than while planning: the original line named a method whose signature this
  task's own non-negotiable requirement rules out.

A myflow change may span more than one repository: the state record already carries **one** scalar
`branch` and a `worktrees` map with one entry per affected repository, per **Multi-repo shape**
(`skills/myflow-contracts/state-file.md`). It is already one unit of work, and the store must read
that way — a two-repository change is one row here, never two.

- [x] **Step 1: The `change_repos` table**

```sql unverified:confirm the FK cascade behaviour against postgres:18-alpine before relying on it in task 3
CREATE TABLE change_repos (
  change_id   BIGINT NOT NULL REFERENCES changes(id) ON DELETE CASCADE,
  repo_root   TEXT   NOT NULL,
  merge_base  TEXT,
  PRIMARY KEY (change_id, repo_root)
);

CREATE INDEX change_repos_repo_root ON change_repos (repo_root);
```

`merge_base` is nullable, and a `NULL` means **no merge base recorded** for that repository — the
state file contract already states that a `null` merge base is legal and is a refusal to infer one,
never a licence to compute one here.

- [x] **Step 2: The owning project stays scalar, and that is deliberate**

`changes.project_key` names the project whose state directory holds the record, exactly as the state
directory is already keyed. It is **not** the list of affected repositories — `change_repos` is.
Document this on the `changes` type so a later reader cannot mistake the owning project for the
full set.

- [x] **Step 3: Repositories are written with the change, in one transaction**

`PutChange` writes the change row and its `change_repos` rows together, so a reader never sees a
change whose repository set is half-written. A repository absent from the payload is removed from
the set — the same whole-object rule `PutChange` already follows for the change's own fields.

**Tests:** `TestPutChangeWritesRepoSet`, `TestTwoRepoChangeIsOneRow`,
`TestPutChangeReplacesRepoSetWholesale`, `TestNullMergeBaseRoundTrips`,
`TestListChangeReposOrdersStably`, `TestRepoSetIsTransactionalWithTheChange`.

**Regression:** Reverting this task collapses a multi-repository change back to a single scalar
project key, so `TestTwoRepoChangeIsOneRow` fails and a change spanning two repositories is either
duplicated into two rows or silently loses a repository.

**Baseline:** before=12 after=18 Go tests.
<!-- predicted: `cd stats && go test ./... -count=1 -v | grep -c '^--- PASS'` after task 2.1 -->

**Commit:** `feat(kan-16-myflow-stats-app): model a change as one unit of work across every repository it touches`

---

### 3 `internal/store` — stage runs, pricing, and the aggregation queries

**Build:** green

**Files:**
- Add: `stats/internal/store/migrations/0003_stage_runs.sql`
- Add: `stats/internal/store/stageruns.go`
- Add: `stats/internal/store/pricing.go`
- Add: `stats/internal/store/aggregate.go`
- Add: `stats/internal/store/stageruns_test.go`
- Add: `stats/internal/store/aggregate_test.go`
- Modify: `stats/internal/store/store_test.go`
- Modify: `stats/internal/store/migrations.go`

**Interfaces:**
- Consumes: task 2's `store.Store` and migration runner, and task 2.1's `change_repos`.
- Produces: `BeginStage`, `EndStage`, `MergeMetrics`, `SweepAbandoned`, `Price`, and one typed
  aggregation method per statistics view, consumed by task 11.

- [x] **Step 1: The `stage_runs` and `pricing` tables**

The DDL in `design.md`, including the GIN index on `metrics` and the unique constraint on
`(change_id, command, stage, attempt)`.

`stage_runs` also carries a **nullable** `repo_root`, referencing `change_repos` by
`(change_id, repo_root)`. It is set when a stage ran inside one repository, and `NULL` when the
stage belongs to the change as a whole. `NULL` here means *the whole unit of work*, and is never
read as *unknown repository* — the same absence-is-not-a-value rule the metrics bag follows.

- [x] **Step 2: Attempt allocation**

`BeginStage` allocates the next attempt number for its `(change, command, stage)` triple inside the
insert, so two concurrent worktrees cannot claim the same attempt.

- [x] **Step 3: `MergeMetrics` merges keys, never replaces the bag**

The harvester writes token keys while a stage-end mark writes outcome keys; neither may erase the
other's. Merging is a JSONB concatenation at the top level with per-key last-write-wins.

- [x] **Step 4: Pricing and the frozen cost**

`Price` resolves the `pricing` row in effect at a stage run's start, computes `cost_usd`, and writes
it alongside `pricing_version`. Token counts remain stored, so a later re-price is possible.

- [x] **Step 5: The aggregation queries**

One method per view in `design.md`'s "The views" table. Each takes a period and an optional project.
Every one restricts by period in SQL. Runs whose metrics record tokens as unavailable are excluded
from token figures and counted separately, never averaged in as zeros.

**Tests:** `TestBeginStageAllocatesAttempts`, `TestConcurrentBeginStageDoesNotCollide`,
`TestMergeMetricsPreservesOtherKeys`, `TestEndStageRecordsOutcome`,
`TestSweepAbandonedClosesSilentStages`, `TestPriceFreezesCostAndVersion`,
`TestAggregateRestrictsByPeriodInSQL`, `TestAggregateEmptyPeriodReturnsEmptyNotError`,
`TestAggregateExcludesUnavailableTokensFromAverages`, `TestAggregateSeparatesMainFromSidechain`.

**Regression:** Reverting this task removes stage runs entirely, so
`TestBeginStageAllocatesAttempts` fails and no statistics view has a source.

**Baseline:** before=18 after=28 Go tests.
<!-- predicted: `cd stats && go test ./... -count=1 -v | grep -c ^=== RUN` after task 3 -->

**Commit:** `feat(kan-16-myflow-stats-app): add stage runs, pricing, and the aggregation queries to the store`

---

### 3.1 `internal/store` — filtering, searching, sorting and pagination over every reasonable field

**Build:** green

**Files:**
- Add: `stats/internal/store/query.go`
- Add: `stats/internal/store/query_test.go`
- Add: `stats/internal/store/migrations/0004_query_indexes.sql`
- Modify: `stats/internal/store/changes.go`
- Modify: `stats/internal/store/stageruns.go`
- Modify: `stats/internal/store/store_test.go`
- Modify: `stats/internal/store/repos.go`

**Interfaces:**
- Consumes: task 2's changes, task 2.1's repositories, task 3's stage runs.
- Produces: a `Query` value — filters, a free-text term, a sort key and direction, a page — that
  `ListChanges` and `ListStageRuns` both accept, and that task 11 builds from request parameters.

Every field a reader would reasonably want to slice by is filterable, sortable and searchable. That
breadth is the requirement; the allowlist below is how it is delivered **safely**, not a narrowing
of it.

- [x] **Step 1: An allowlist maps request field names to real columns**

A sort key and a filter field are **identifiers**, and an identifier cannot be a bound parameter —
so it can never be interpolated from request text. `query.go` holds a fixed map from the field names
the API accepts to the columns they mean, and a field absent from that map is a rejected request,
never a query built from the caller's string. This is the whole injection surface of the statistics
layer; treat any dynamic SQL fragment outside this map as a defect.

Fields covered: on changes — project, name, state, branch, jira issue, planning effort, review panel
roster, PR URL presence, updated at, updated by; on stage runs — command, stage, attempt, harness,
outcome, started at, ended at, duration, repo root; plus the owning change's fields by join.

- [x] **Step 2: Metrics keys are filterable and sortable too**

The metrics bag is where the interesting numbers live, so a key inside it is addressable as a field
(for example `metrics.tokens.cache_read`). The **key path** is validated against a syntax rule and
bound as a parameter to a JSONB path operator — it is never concatenated into SQL. A run whose
metrics lack the key sorts as *absent*, and absent is ordered distinctly from a recorded zero, per
this change's global constraint.

- [x] **Step 3: Free-text search across identity fields**

One search term matches across change name, Jira issue, branch, command, stage, updated by, and
repository root. Case-insensitive, substring, and bound as a parameter.

- [x] **Step 4: Pagination, with a stable order**

Every list is paginated, and every sort is made total by appending a unique tiebreaker, so a row
cannot appear on two pages or on neither as the caller pages through.

- [x] **Step 5: Indexes for the orderings that will actually be used**

Btree indexes for the common sort keys, and the GIN index task 3 already created for metrics
containment.

**Tests:** `TestSortByEveryAllowlistedField`, `TestUnknownSortFieldIsRejected`,
`TestUnknownFilterFieldIsRejected`, `TestSortInjectionAttemptIsRejected`,
`TestFilterByMetricsKeyPath`, `TestMalformedMetricsKeyPathIsRejected`,
`TestAbsentMetricSortsDistinctlyFromZero`, `TestSearchMatchesAcrossIdentityFields`,
`TestPaginationIsStableUnderEqualSortKeys`, `TestPageBoundaryLosesNoRow`.

**Regression:** Reverting this task removes the allowlist and the query builder, so
`TestSortInjectionAttemptIsRejected` cannot run and every list endpoint falls back to a fixed order
with no filtering.

**Baseline:** before=49 after=60 Go tests.
<!-- predicted: `cd stats && go test ./... -count=1 -v | grep -c '^--- PASS'` after task 3.1 -->

**Commit:** `feat(kan-16-myflow-stats-app): filter, search, sort and paginate over every reasonable field`

---

### 4 `cmd/myflowd` — the daemon and the state API

**Build:** green

**Files:**
- Add: `stats/cmd/myflowd/main.go`
- Add: `stats/internal/api/server.go`
- Add: `stats/internal/api/changes.go`
- Add: `stats/internal/api/changes_test.go`
- Add: `stats/internal/config/config.go`
- Add: `stats/internal/config/config_test.go`
- Modify: `stats/internal/store/changes.go`
- Modify: `stats/internal/store/changes_test.go`

**Interfaces:**
- Consumes: task 2's change repository.
- Produces: an HTTP server on loopback exposing the change endpoints in `design.md`'s API section,
  consumed by task 5.

- [x] **Step 1: Bind loopback only**

The listener binds `127.0.0.1` on a configured port. Binding any other interface is a configuration
error the daemon refuses at startup rather than a runtime option.

- [x] **Step 2: The change endpoints**

`GET /api/v1/changes`, `GET /api/v1/changes/{project}/{name}`, `PUT /api/v1/changes/{project}/{name}`.
A monotonic refusal from the store returns a distinct status the CLI can tell apart from a transport
failure, because the two need opposite handling: one is a correct refusal, the other is a fallback
trigger.

- [x] **Step 3: Graceful shutdown**

In-flight requests complete; the pool closes after them.

**Tests:** `TestServerBindsLoopbackOnly`, `TestPutChangeReturnsMonotonicRefusalDistinctly`,
`TestGetChangeUnknownReturns404`, `TestListChangesFiltersByProjectAndState`,
`TestShutdownDrainsInFlight`.

**Regression:** Reverting this task leaves the CLI in task 5 with no server to reach, so every
non-fallback CLI test fails.

**Baseline:** before=66 after=82 Go tests.
<!-- predicted: `cd stats && go test ./... -count=1 -v | grep -c ^=== RUN` after task 4 -->

**Commit:** `feat(kan-16-myflow-stats-app): add the myflowd daemon and its state API`

---

### 5 `cmd/myflow` — the CLI, its state commands, and the non-blocking fallback

**Build:** green

**Files:**
- Add: `stats/cmd/myflow/main.go`
- Add: `stats/cmd/myflow/state.go`
- Add: `stats/internal/client/client.go`
- Add: `stats/internal/client/client_test.go`
- Add: `stats/internal/fallback/journal.go`
- Add: `stats/internal/fallback/statefile.go`
- Add: `stats/internal/fallback/journal_test.go`
- Add: `stats/cmd/myflow/state_test.go`
- Modify: `stats/internal/api/server.go`
- Modify: `stats/cmd/myflow/stage.go`
- Modify: `stats/internal/api/changes.go`
- Modify: `stats/internal/api/changes_test.go`
- Modify: `skills/myflow-contracts/state-file.md`

**Interfaces:**
- Consumes: task 4's API.
- Produces: `myflow state get` and `myflow state set`, the two commands every skill will call, and
  the journal format tasks 6 and 7 read.

- [x] **Step 1: `state get` and `state set`**

`set` sends the whole object. Both resolve the project key exactly as the state file contract
computes it, from `git rev-parse --git-common-dir`, so a worktree and its main checkout address the
same record.

- [x] **Step 2: The fallback, on every failure mode**

Daemon unreachable, connection refused, timeout, or any non-2xx that is not the monotonic refusal:
write the on-disk JSON at the existing state path, append the intent to the journal beside it, print
one warning line, and **exit 0**.

A monotonic refusal is not a fallback trigger — the store was reached and answered correctly. The
CLI reports it and exits non-zero, exactly as a rejected write should.

- [x] **Step 3: Reads fall back too**

`state get` with no reachable store returns the on-disk record and says the value came from the
fallback rather than the store.

**Tests:** `TestStateSetFallsBackOnDeadPort`, `TestStateSetFallsBackOnTimeout`,
`TestStateSetFallbackExitsZero`, `TestStateSetFallbackWritesStateFileAndJournalEntry`,
`TestMonotonicRefusalIsNotAFallback`, `TestStateGetFallsBackAndSaysSo`,
`TestProjectKeyMatchesStateFileContract`, `TestProjectKeyIsIdenticalFromAWorktree`.

**Regression:** Reverting this task removes the CLI, so no skill has a way to reach the store and
`TestStateSetFallbackExitsZero` — the test that encodes the never-block guarantee — cannot run.

**Baseline:** before=82 after=124 Go tests.
<!-- predicted: `cd stats && go test ./... -count=1 -v | grep -c ^=== RUN` after task 5 -->

**Commit:** `feat(kan-16-myflow-stats-app): add the myflow CLI, its state commands, and the non-blocking journal fallback`

---

### 6 Journal reconciliation

**Build:** green

**Files:**
- Add: `stats/internal/reconcile/reconcile.go`
- Add: `stats/internal/reconcile/reconcile_test.go`
- Add: `stats/internal/reconcile/testsupport_test.go`
- Add: `stats/cmd/myflow/journal.go`
- Add: `stats/cmd/myflow/journal_test.go`
- Modify: `stats/internal/api/changes.go`
- Modify: `stats/internal/store/store.go`
- Add: `stats/internal/fallback/lock.go`
- Add: `stats/internal/fallback/lock_test.go`
- Modify: `stats/internal/fallback/journal.go`
- Modify: `stats/cmd/myflowd/main.go`
- Modify: `stats/cmd/myflow/main.go`

**Interfaces:**
- Consumes: task 5's journal format, task 2's change repository.
- Produces: automatic replay at daemon startup and on reconnect, plus a `myflow journal flush`
  command for replaying on demand.

- [x] **Step 1: Replay at startup and on reconnect**

The daemon replays every journal entry it finds. Conflicts resolve by `updatedAt`, with the
monotonic rule as the tiebreaker.

- [x] **Step 2: A stale entry never regresses a stored state**

An entry recording an earlier state than the one stored is retired, not applied and not retried
forever.

- [x] **Step 3: Interrupted replay repeats rather than loses**

An entry is removed from the journal only once the store has accepted or explicitly refused it.

**Tests:** `TestReplayAppliesPendingEntries`, `TestStaleEntryCannotRegressFinished`,
`TestInterruptedReplayResumesWithoutDuplicating`, `TestReplayRetiresRefusedEntries`,
`TestJournalFlushCommandReplaysOnDemand`.

**Regression:** Reverting this task means a journal written during an outage is never applied, so
`TestStaleEntryCannotRegressFinished` fails and state written during an outage is silently lost on
reconnect.

**Baseline:** before=124 after=136 Go tests.
<!-- predicted: `cd stats && go test ./... -count=1 -v | grep -c ^=== RUN` after task 6 -->

**Commit:** `feat(kan-16-myflow-stats-app): reconcile the write-ahead journal into the store`

---

### 8 Stage marks — the CLI commands and the daemon endpoints

**Build:** green

**Files:**
- Add: `stats/cmd/myflow/stage.go`
- Add: `stats/internal/api/stages.go`
- Add: `stats/internal/api/stages_test.go`
- Add: `stats/cmd/myflow/stage_test.go`
- Add: `stats/internal/stages/names.go`
- Add: `stats/internal/stages/names_test.go`
- Modify: `stats/cmd/myflow/main.go`
- Modify: `stats/cmd/myflowd/main.go`
- Modify: `stats/internal/api/server.go`
- Modify: `stats/internal/api/changes_test.go`
- Modify: `stats/internal/client/client.go`
- Modify: `stats/internal/client/client_test.go`
- Modify: `stats/internal/reconcile/reconcile.go`
- Modify: `stats/internal/reconcile/reconcile_test.go`
- Add: `stats/internal/reconcile/stage_test.go`
- Modify: `stats/cmd/myflow/journal.go`

**Interfaces:**
- Consumes: task 3's `BeginStage`/`EndStage`, task 5's fallback.
- Produces: `myflow stage begin` and `myflow stage end`, called by every skill in task 15.

- [x] **Step 1: The documented stage names, in code**

`internal/stages/names.go` holds the stage names from `README.md`'s Level 1 table, keyed by command.
A mark naming a stage absent from that table is rejected by the CLI with a message naming the
documented alternatives — this is what keeps the two vocabularies one list.

- [x] **Step 2: The mark commands**

`stage begin` records command, stage, change, harness, session id and start instant. `stage end`
records the end instant, the outcome, and any metrics passed as arguments (`--fix-rounds`,
`--panel-rounds`, `--findings`).

- [x] **Step 3: A mark never blocks**

Identical fallback to task 5: warning line, journal entry, exit 0.

- [x] **Step 3a: A journalled mark is replayed, like a journalled state write**

A stage mark that falls back during an outage is written to its own journal, separate from the state
journal — the state journal's entries are decoded as whole change records with unknown fields
rejected, so a mark's body mixed in would break replay for every entry after it. That separation is
correct; what it must not do is leave the mark unreplayed. Reconciliation walks the stage journal
too, decoding each entry with its own decoder, so a mark survives an outage exactly as a state write
does. A mark journalled to a file nothing reads is a mark lost — silently, and precisely during the
outages the statistics would most want to show.

- [x] **Step 4: A mark for an unknown change is stored**

Against a synthetic change row for the project, rather than dropped. A mark for a change nobody
created is a defect worth seeing.

**Tests:** `TestStageNamesMatchReadmeLevelOne`, `TestUndocumentedStageNameIsRejected`,
`TestStageBeginRecordsIdentityAndInstant`, `TestStageEndRecordsOutcomeAndMetrics`,
`TestStageMarkFallsBackAndExitsZero`, `TestMarkForUnknownChangeIsStored`.

**Regression:** Reverting this task removes stage marks, so `TestStageNamesMatchReadmeLevelOne`
fails and no stage window exists for the harvester in task 9 to attribute usage to.

**Baseline:** before=136 after=176 Go tests.
<!-- predicted: `cd stats && go test ./... -count=1 -v | grep -c ^=== RUN` after task 8 -->

**Commit:** `feat(kan-16-myflow-stats-app): add stage mark commands and their daemon endpoints`

---

### 9 `internal/harvest` — transcript parsing and attribution

**Build:** green

**Files:**
- Add: `stats/internal/harvest/transcript.go`
- Add: `stats/internal/harvest/attribute.go`
- Add: `stats/internal/harvest/watcher.go`
- Add: `stats/internal/harvest/transcript_test.go`
- Add: `stats/internal/harvest/attribute_test.go`
- Add: `stats/internal/harvest/watcher_test.go`
- Modify: `stats/internal/store/query.go`
- Modify: `stats/internal/store/query_test.go`
- Modify: `stats/internal/store/stageruns.go`
- Modify: `stats/internal/store/stageruns_test.go`
- Add: `stats/internal/store/migrations/0005_jsonb_deep_add.sql`
- Add: `stats/internal/store/migrations/0006_harvest_offsets.sql`
- Add: `stats/internal/store/harvest.go`
- Add: `stats/internal/store/harvest_test.go`
- Modify: `stats/internal/harvest/transcript.go`
- Modify: `stats/internal/harvest/transcript_test.go`
- Modify: `stats/cmd/myflowd/main.go`
- Add: `stats/testdata/transcripts/` (fixtures: a main-thread session, a session with sidechain
  messages, a session whose messages straddle a stage boundary, a truncated final line)
- Modify: `stats/cmd/myflowd/main.go`

**Allowed-collateral:** `stats/testdata/transcripts/*.jsonl`

**Interfaces:**
- Consumes: an interface answering *which stage window is open for this session* — **not** the store
  directly, so this package is testable with no PostgreSQL running.
- Produces: token, model and effort metrics merged into stage runs via task 3's `MergeMetrics`.

- [x] **Step 1: Parse the transcript**

Each assistant record yields `timestamp`, `message.model`, `effort`, `isSidechain`, and `usage` with
`input_tokens`, `cache_creation_input_tokens`, `cache_read_input_tokens`, `output_tokens` and
`output_tokens_details.thinking_tokens`.

```json verified:read from a live session file under ~/.claude/projects/ on 2026-08-13
{"usage":{"input_tokens":2,"cache_creation_input_tokens":5901,"cache_read_input_tokens":101190,"output_tokens":1072,"output_tokens_details":{"thinking_tokens":696}}}
```

A truncated final line — the file is being appended to as it is read — is not an error: stop at the
last complete record and resume from there.

- [x] **Step 2: Attribute to the open window**

A message belongs to the open stage run whose session matches and whose `[started_at, ended_at]`
interval contains its timestamp. Main-thread and sidechain totals accumulate under separate keys.

- [x] **Step 3: Idempotence by byte offset**

The last consumed offset per transcript file is persisted, so a restart re-reads nothing and
double-counts nothing.

- [x] **Step 4: Watch, from the daemon**

The daemon watches the transcript directory and harvests continuously.

**Tests:** `TestParseUsageFromLiveFixture`, `TestTruncatedFinalLineIsResumedNotFailed`,
`TestAttributeToOpenWindow`, `TestMessagesOutsideAnyWindowAreNotAttributed`,
`TestSidechainAccumulatesSeparately`, `TestBoundaryStraddlingMessagesSplitByTimestamp`,
`TestConsecutiveRunsOverUnchangedFileAddNothing`,
`TestOutageAcrossSeveralCyclesThenRecoveryMatchesCleanRun`, `TestHarvestNeedsNoDatabase`.

**Regression:** Reverting this task removes the only source of real token numbers, so
`TestRestartDoesNotDoubleCount` fails and every cost view has nothing but wall-clock time.

**Baseline:** before=176 after=205 Go tests.
<!-- predicted: `cd stats && go test ./... -count=1 -v | grep -c ^=== RUN` after task 9 -->

**Commit:** `feat(kan-16-myflow-stats-app): harvest session transcripts and attribute usage to stage windows`

---

### 10 The abandoned-stage sweeper and honest unavailability

**Build:** green

**Files:**
- Add: `stats/internal/sweep/sweep.go`
- Add: `stats/internal/sweep/sweep_test.go`
- Add: `stats/internal/sweep/testsupport_test.go`
- Add: `stats/internal/sweep/export_test.go`
- Modify: `stats/internal/api/stages.go`
- Modify: `stats/internal/api/stages_test.go`
- Modify: `stats/cmd/myflow/stage_test.go`
- Modify: `stats/cmd/myflow/stage.go`
- Modify: `stats/internal/harvest/attribute.go`
- Modify: `stats/cmd/myflowd/main.go`

**Interfaces:**
- Consumes: task 3's `SweepAbandoned`.
- Produces: closed stage runs for dead sessions, and the `tokens_available` flag every view in
  task 11 reads.

- [x] **Step 1: Sweep silent sessions**

A stage whose session has been silent past a configured timeout is closed with an abandoned outcome.

- [x] **Step 2: Record unavailability, never zero**

A mark emitted from a harness that writes no readable transcript sets `tokens_available: false` on
its stage run. Nothing anywhere writes a zero token count to stand in for an unmeasured one.

**Tests:** `TestSweepClosesSilentStages`, `TestSweepLeavesActiveStagesOpen`,
`TestSweepIsIdempotent`, `TestNonClaudeHarnessMarksTokensUnavailable`,
`TestUnavailableIsNeverWrittenAsZero`.

**Regression:** Reverting this task leaves crashed runs' stages open forever, so
`TestSweepClosesSilentStages` fails and the rework-rate view cannot count abandonment.

**Baseline:** before=205 after=213 Go tests.
<!-- predicted: `cd stats && go test ./... -count=1 -v | grep -c ^=== RUN` after task 10 -->

**Commit:** `feat(kan-16-myflow-stats-app): sweep abandoned stages and record unmeasured metrics as unavailable`

---

### 11 The statistics API — all eight views

**Build:** green

**Files:**
- Add: `stats/internal/api/stats.go`
- Add: `stats/internal/api/stats_test.go`
- Modify: `stats/internal/api/server.go`
- Modify: `stats/internal/api/changes.go`
- Modify: `stats/cmd/myflowd/main.go`
- Modify: `stats/internal/api/changes_test.go`
- Modify: `stats/internal/api/stages_test.go`
- Modify: `stats/internal/client/client_test.go`

**Interfaces:**
- Consumes: task 3's aggregation methods.
- Produces: `GET /api/v1/stats/{view}`, consumed by task 13.

- [x] **Step 1: One route, eight views**

Each view accepts `from`, `to` and an optional `project`. Every one restricts by period in the
query.

- [x] **Step 2: The boundary convention, stated once**

A stage run straddling a period boundary is attributed by its start instant, and the response says
so, so every view answers the question the same way.

- [x] **Step 3: Absence is visible**

A period before the store held anything reports that no data was recorded for that span,
distinctly from a recorded zero. The store starts empty; nothing is imported.

- [x] **Step 4: Every list endpoint exposes task 3.1's query surface**

`filter`, `q`, `sort`, `order` and page parameters on the changes and stage-run listings, built into
a `store.Query` and validated there. An unknown field is a 400 naming the field and the fields that
are accepted — never a silently ignored parameter, and never a query built from the caller's string.

- [x] **Step 5: Multi-repo aggregation is the default**

A change spanning repositories is one row in every view, with per-repository breakdown available on
request rather than as the default shape.

**Tests:** `TestEveryViewAcceptsAPeriod`, `TestEmptyPeriodReturnsEmptyNotError`,
`TestBoundaryConventionIsConsistentAcrossViews`, `TestProjectFilterRestrictsResults`,
`TestNoProjectFilterAggregatesAcrossProjects`, `TestPeriodBeforeAnyDataReportsNotRecorded`,
`TestLiveStateBoardMatchesStatusOutput`, `TestReworkRateReadsAttemptsNotTiming`,
`TestListEndpointsAcceptFilterSortSearchPage`, `TestUnknownQueryFieldReturns400NamingIt`,
`TestMultiRepoChangeIsOneRowInEveryView`, `TestPerRepoBreakdownAvailableOnRequest`.

**Regression:** Reverting this task removes every statistics endpoint, so
`TestEmptyPeriodReturnsEmptyNotError` fails and the interface in task 13 has nothing to read.

**Baseline:** before=213 after=231 Go tests.
<!-- predicted: `cd stats && go test ./... -count=1 -v | grep -c ^=== RUN` after task 11 -->

**Commit:** `feat(kan-16-myflow-stats-app): serve the eight statistics views over the API`

---

### 12 The web application — scaffold, build, and embedding

**Build:** green

**Files:**
- Add: `stats/web/package.json`
- Add: `stats/web/vite.config.ts`
- Add: `stats/web/tsconfig.json`
- Add: `stats/web/index.html`
- Add: `stats/web/src/main.tsx`
- Add: `stats/web/src/api.ts`
- Add: `stats/web/src/api.test.ts`
- Add: `stats/internal/web/embed.go`
- Add: `stats/internal/web/embed_test.go`
- Modify: `stats/Makefile`
- Modify: `stats/internal/api/server.go`
- Modify: `stats/cmd/myflowd/main.go`
- Modify: `.gitignore`
- Add: `stats/web/package-lock.json`

**Interfaces:**
- Consumes: task 11's statistics API.
- Produces: a built SPA embedded in the daemon binary, and the typed API client task 13's views use.

- [x] **Step 1: Scaffold**

React 19.2.8 and Vite 8.2.1. `make build` runs the Vite build and then `go build`, in that order, so
the embedded assets are always the current ones.

- [x] **Step 2: Embed**

`go:embed` over the Vite output. A build with no `dist/` present fails loudly at compile time rather
than serving an empty page at runtime.

- [x] **Step 3: Serve**

The daemon serves the SPA at `/`, with unknown non-API paths falling through to `index.html` for
client-side routing.

**Tests:** `TestEmbeddedAssetsArePresent`, `TestUnknownPathServesIndexNotFound`,
`TestApiPathsAreNotSwallowedBySpaFallback`, and the SPA's `api.test.ts` covering period and project
query construction.

**Regression:** Reverting this task removes the interface entirely, so `TestEmbeddedAssetsArePresent`
fails and the statistics have no reader.

**Baseline:** before=231 after=234 Go tests, plus 16 SPA tests.
<!-- predicted: `cd stats && go test ./... -count=1 -v | grep -c ^=== RUN` and `cd stats/web && npm test` after task 12 -->

**Commit:** `feat(kan-16-myflow-stats-app): scaffold the web application and embed it in the daemon`

---

### 13 The web application — the eight views and the project filter

**Build:** green

**Files:**
- Add: `stats/web/src/views/StateBoard.tsx`
- Add: `stats/web/src/views/CostPerChange.tsx`
- Add: `stats/web/src/views/StageLeaderboard.tsx`
- Add: `stats/web/src/views/Trend.tsx`
- Add: `stats/web/src/views/CacheEfficiency.tsx`
- Add: `stats/web/src/views/PanelEconomics.tsx`
- Add: `stats/web/src/views/ModelComparison.tsx`
- Add: `stats/web/src/views/ReworkRate.tsx`
- Add: `stats/web/src/components/PeriodPicker.tsx`
- Add: `stats/web/src/components/ProjectFilter.tsx`
- Add: `stats/web/src/components/Unavailable.tsx`
- Add: `stats/web/src/App.tsx`
- Add: `stats/web/src/components/ViewFrame.tsx`
- Add: `stats/web/src/hooks/useStatsView.ts`
- Add: `stats/web/src/format.ts`
- Add: `stats/web/src/viewTypes.ts`
- Add: `stats/web/src/styles.css`
- Add: `stats/web/src/setupTests.ts`
- Modify: `stats/web/src/main.tsx`
- Modify: `stats/web/src/api.ts`
- Modify: `stats/web/src/api.test.ts`
- Modify: `stats/internal/api/stats.go`
- Modify: `stats/internal/api/stats_test.go`
- Modify: `stats/web/package.json`
- Modify: `stats/web/package-lock.json`
- Modify: `stats/web/tsconfig.json`
- Modify: `stats/web/vite.config.ts`
- Add: `stats/web/src/components/DataTable.tsx`
- Add: `stats/web/src/components/SearchBox.tsx`
- Add: `stats/web/src/components/FilterBar.tsx`
- Add: `stats/web/src/views/views.test.tsx`
- Add: `stats/web/src/components/DataTable.test.tsx`

**Interfaces:**
- Consumes: task 12's API client and task 11's query parameters.
- Produces: the interface the operator actually reads.

- [x] **Step 1: The eight views**

One route each, every one driven by the period picker and the project filter.

- [x] **Step 1a: One table component carries filter, search and sort**

`DataTable` renders sortable column headers, a filter bar and paging controls, and drives them
through task 11's query parameters rather than filtering client-side — the server already does this
work, and doing it twice would make a paginated list wrong. `SearchBox` drives the free-text term.
Every view that lists rows uses this component, so sorting behaves identically across all of them.

- [x] **Step 1b: A multi-repo change reads as one row**

With its repositories shown as a detail of that row, never as separate rows.

- [x] **Step 2: Unavailable is rendered as unavailable**

The `Unavailable` component renders a metric marked unavailable distinctly from a measured zero.
This is the interface half of the constraint the schema and the API already hold.

- [x] **Step 3: The state board is the default route**

It is the view that replaces reading a state file, so it is what opens first.

**Tests:** `views.test.tsx` — one case per view asserting it renders from a fixture response, plus
`cache efficiency: a real zero ratio reads as a value, an unmeasured ratio reads as unavailable`,
`a period before the store held anything is stated, not rendered as an empty measured table`,
`the live state board is the default route`, and `an unrecognised route also falls back to the
state board`; plus `DataTable.test.tsx` for the shared table's sort, filter and paging.

**Regression:** Reverting this task leaves the daemon serving an empty shell, so the views test file
is absent and no statistic is readable by a human.

**Baseline:** before=16 after=39 SPA tests.
<!-- predicted: `cd stats/web && npm test` after task 13 -->

**Commit:** `feat(kan-16-myflow-stats-app): add the eight statistics views and the project filter`

---

### 14 Rewrite the state contract against the store

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/state-file.md`
- Modify: `scripts/check-contract-budget.sh`
- Modify: `stats/internal/api/changes.go`
- Modify: `stats/internal/reconcile/reconcile.go`
- Modify: `stats/internal/reconcile/reconcile_test.go`

**Interfaces:**
- Consumes: task 5's CLI surface.
- Produces: the contract every skill reads before touching state.

- [x] **Step 1: Describe the store, keep the vocabulary**

The field list, the closed-schema rule, the monotonic-write rule and the carry-forward rule are
unchanged in substance. What changes is the mechanism: a `jq` read becomes `myflow state get`, and a
full-object write becomes `myflow state set`.

- [x] **Step 2: State the never-block guarantee in the contract**

The fallback is a property of the contract, not an implementation detail of the CLI, because every
skill depends on it being true.

- [x] **Step 3: Re-anchor the budget row**

The file's budget row in `scripts/check-contract-budget.sh` is re-anchored to its new size plus the
guard's standard headroom. Raising a budget for a genuine addition is the correct response; narrowing
the guard is not.

**Tests:** No automated test — this task edits contract prose. Verification is the repository's
own reference, markdown-integrity, contract-budget and vocabulary guards all exiting clean, plus
two read-throughs: the rewritten contract against every scenario in this change's state-store
delta spec, and against the code backing each claim in the CLI, client, fallback and reconcile
packages.


**Regression:** Reverting this task restores a contract describing a JSON file, so every skill's
state step describes a mechanism the CLI no longer implements.

**Baseline:** before=0 after=0 automated cases; the three named guards exit clean.
<!-- predicted: the three guard commands in `.myflow/project.md`'s lint section, after task 14 -->

**Commit:** `refactor(kan-16-myflow-stats-app): rewrite the state contract against the store`

---

### 15 Emit stage marks from every skill

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `skills/myflow-start/SKILL.md`
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-finish/SKILL.md`
- Modify: `skills/myflow-fast/SKILL.md`
- Modify: `scripts/check-contract-budget.sh`
- Modify: `README.md`
- Modify: `stats/internal/stages/names.go`

**Interfaces:**
- Consumes: task 8's mark commands and stage-name table.
- Produces: the stage runs every statistics view aggregates.

- [x] **Step 1: The obligation, in the pipeline contract**

A new section stating that each command marks its stages, citing `README.md`'s Level 1 table for the
names rather than restating them, and stating that a failed mark never blocks.

- [x] **Step 2: Mark calls in each skill**

At each skill's own stage boundaries, matching its row in the Level 1 table exactly.

- [x] **Step 3: Re-anchor the budget rows**

For every contract or skill file this task grows.

**Tests:** No automated test for the prose itself. The mechanical tie is the stages package's
own test, which re-parses the README's Level 1 table and compares it against the encoded list, so a
mark naming an undocumented stage fails there. Verification is that test plus the repository's
reference, markdown-integrity, contract-budget and vocabulary guards all exiting clean, and a
read-through of each edited skill confirming the marks sit at the documented boundaries.


**Regression:** Reverting this task removes every mark call, so no stage run is ever created and
every statistics view is permanently empty despite the whole store working.

**Baseline:** before=0 after=0 automated cases; the three named guards exit clean.
<!-- predicted: the three guard commands in `.myflow/project.md`'s lint section, after task 15 -->

**Commit:** `feat(kan-16-myflow-stats-app): emit stage marks from every pipeline command`

---

### 16 `/myflow-status` reads the store

**Build:** green

**Files:**
- Modify: `skills/myflow-status/SKILL.md`
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `stats/internal/client/client.go`
- Modify: `stats/internal/client/client_test.go`
- Modify: `stats/cmd/myflow/main.go`
- Modify: `stats/cmd/myflow/state.go`
- Modify: `stats/cmd/myflow/state_test.go`
- Modify: `stats/internal/fallback/statefile.go`
- Modify: `stats/internal/fallback/journal_test.go`
- Modify: `scripts/check-contract-budget.sh`

**Interfaces:**
- Consumes: task 5's `state get`, task 11's live state board.
- Produces: a status report whose source is the store, with the fallback path stated.

- [x] **Step 1: Read through the CLI**

Change name resolution enumerates from the store rather than from the state directory, falling back
to the directory when the store is unreachable — and saying which source it used.

- [x] **Step 2: Keep the unreadable-record rule**

A record that cannot be read is still reported by name and skipped, never rebuilt by inference.

**Tests:** No automated test for the skill prose itself. The mechanical ties are the new CLI list
command's own tests — store-backed listing, the fallback path exiting zero with its source named,
and a response lacking the daemon header treated as unavailable — plus the statistics API test
asserting the live state board agrees with what this command reports. Verification is those tests
plus the repository's reference, markdown-integrity, contract-budget and vocabulary guards all
exiting clean, and running the command live with the daemon up and stopped.


**Regression:** Reverting this task leaves `/myflow-status` reading a directory that is no longer
authoritative, so it reports stale state for any change written while the directory was only a
journal.

**Baseline:** before=0 after=0 automated cases; the three named guards exit clean.
<!-- predicted: the three guard commands in `.myflow/project.md`'s lint section, after task 16 -->

**Commit:** `refactor(kan-16-myflow-stats-app): read /myflow-status from the store`

---

### 17 Run the daemon at login, and declare the project's new commands

**Build:** green

**Files:**
- Add: `stats/launchd/com.tweety53.myflowd.plist`
- Modify: `.myflow/project.md`
- Modify: `stats/README.md`
- Modify: `CLAUDE.md`
- Modify: `AGENTS.md`

**Interfaces:**
- Consumes: everything above.
- Produces: a running daemon at login and a project configuration naming the new lint and test
  commands, so `/myflow-do` runs them for every future change.

- [x] **Step 1: The launchd user agent**

Starts `myflowd` at login, restarts it on failure, logs to a known path. Loading it is documented in
`stats/README.md` as an operator step, not performed by a skill.

- [x] **Step 2: Declare the new commands**

`.myflow/project.md` gains the Go and SPA test commands under `## test`, the Go vet and formatter
checks under `## lint`, and a `## run` entry describing the compose stack and the daemon — replacing
that section's current statement that this repository has nothing to run.

- [x] **Step 3: Name the commands in the agent instructions**

`CLAUDE.md` and `AGENTS.md` name the auto-fix and check commands the Lint Fix Priority rule refers
to, which they currently leave as placeholders.

**Tests:** No automated test — configuration and documentation. Verification is
`scripts/check-workspace-isolation.sh` exiting clean against the edited `.myflow/project.md`, the
full `## test` list running green, and the daemon answering on its port after the agent is loaded.

**Regression:** Reverting this task means the daemon is not running after a reboot, so every command
takes the journal fallback path until it is started by hand, and `/myflow-do` never runs the Go
tests for a future change.

**Baseline:** before=0 after=0 automated cases; the full `## test` list exits clean.
<!-- predicted: every command in `.myflow/project.md`'s test section, after task 17 -->

**Commit:** `chore(kan-16-myflow-stats-app): run the daemon at login and declare the project's new lint and test commands`

---

## Fix round 2 — the development database purged (2026-08-14)

**No task, no file, no commit.** At the operator's instruction the development store's recorded
history was truncated — `changes`, `change_repos`, `stage_runs` and `harvest_offsets` — keeping the
schema, the migrations and the four seeded `pricing` rows. Every record removed was test residue
from the guard-script harnesses and from two stray daemons; no real `/myflow-*` run was among them.
Rationale in `proposal.md`'s **Fix round 2** section. Recorded here so the plan and the store's
actual contents do not silently disagree.

## Fix round 1 — the run drill-down, and a Grafana-shaped interface

Recorded 2026-08-14, at the `IN_PROGRESS` review gate, per **3. Documenting a fix, before
implementing it** (`skills/myflow-do/SKILL.md`). Rationale and scope in `proposal.md`'s
**Fix round 1** section; the two requirements these tasks satisfy are **One change opens on its own
dashboard** and **The interface is a dashboard of panels under shared controls**
(`specs/myflow-stats-views/spec.md`).

**These three tasks touch `stats/web/` and nothing else.** No schema, no store, no daemon handler,
no CLI, no contract and no skill file changes in this round. A task here that finds itself editing
`internal/` has left its scope and is wrong.

**Baseline for this round:** 38 SPA test cases, measured as `cd stats/web && npm test` on
2026-08-14 against commit `c2357c5`.
<!-- measured: `cd stats/web && npm test` reported "Tests  38 passed (38)" on 2026-08-14 -->

---

### 18 The run detail dashboard, and the board rows that reach it

**Build:** green

**Files:**
- Add: `stats/web/src/views/RunDetail.tsx`
- Add: `stats/web/src/views/RunDetail.test.tsx`
- Add: `stats/web/src/components/StageTimeline.tsx`
- Add: `stats/web/src/components/StageRunTable.tsx`
- Add: `stats/web/src/hooks/useRunDetail.ts`
- Add: `stats/web/src/metrics.ts`
- Add: `stats/web/src/metrics.test.ts`
- Modify: `stats/web/src/api.ts`
- Modify: `stats/web/src/api.test.ts`
- Modify: `stats/web/src/App.tsx`
- Modify: `stats/web/src/views/StateBoard.tsx`
- Modify: `stats/web/src/views/views.test.tsx`
- Modify: `stats/web/src/format.ts`
- Modify: `stats/web/src/styles.css`

**Interfaces:**
- Consumes: `GET /api/v1/stage-runs` (task 3.1's filter/sort/page surface, given an HTTP caller by
  task 12) and `GET /api/v1/stats/cost-per-change` (task 11), both already served — this task adds
  no endpoint and changes no handler.
- Produces: a `#/run/<project>/<change>` route, and a `StageRunDTO` in the typed client.

- [x] **Step 1: Type the stage-run list endpoint in the client**

`api.ts` gains `StageRunDTO` mirroring `internal/api/stats.go`'s `stageRunDTO` field for field —
`stageRunId`, `repoRoot`, `harness`, `sessionId`, `command`, `stage`, `attempt`, `startedAt`,
`endedAt`, `outcome`, `metrics` — and a `listStageRuns(query)` built on the existing
`buildListQuery`, exactly as `listChanges` is. `metrics` stays `unknown`: it is a deliberately open
JSONB bag (`design.md`, the metrics-bag decision), and typing it as a closed interface here would
re-close in the client what the schema went out of its way to leave open.

The test asserts the query string `listStageRuns` builds for a change — `project`, `name`, a sort,
a limit — and that a reserved-name collision is rejected the same way `listChanges` rejects it.

- [x] **Step 2: Read known metrics out of the open bag, without closing it**

New `metrics.ts` exposes readers for the keys the aggregation SQL already names — `cost_usd`,
`model`, `tokens.input`, `tokens.output`, `tokens.cache_read`, `tokens.cache_creation`,
`tokens.main`, `tokens.sidechain` — plus `effort` and `fast_mode`, and one function returning every
*other* key in the bag so an unrecognised metric is still displayed rather than dropped.

**Every reader returns `null` for an absent key and never `0`,** and returns `null` for a key whose
value is the wrong JSON type rather than coercing it. This is the absence-is-never-zero rule at the
last layer that can still break it, and it is what the round's own spec scenario ("a stage run that
recorded no token metrics") asserts. Tests cover: a present numeric key, an absent key, a key
present but `null`, a key holding a string where a number is expected, a nested `tokens` object
missing entirely, and the pass-through of an unknown key.

- [x] **Step 3: The stage-run table**

`StageRunTable.tsx` renders one row per stage run — command, stage, attempt, outcome, duration,
tokens in/out/cached, cost, model — with each row expandable to the run's whole raw metrics bag,
including keys this build does not know about. A `null` from any reader renders through the existing
`Unavailable` component, never as `0` or `—`-that-could-be-zero.

Duration comes from `startedAt`/`endedAt`; a run with no `endedAt` is *open*, not zero-length, and
renders as still running.

- [x] **Step 4: The timeline**

`StageTimeline.tsx` draws each stage run as a bar positioned by `startedAt` and `endedAt` across the
change's own span, grouped by command. It is presentation over the same rows the table already has
— it issues no request of its own, so the two can never disagree.

An open run (no `endedAt`) draws to the right edge and is marked as open rather than being given a
fabricated end.

- [x] **Step 5: The route, and its summary numbers**

`useRunDetail.ts` fetches two things for `#/run/<project>/<change>`: the change's stage runs, and
the `cost-per-change` aggregate scoped to that change. **The header's totals come from the
aggregate, never from summing the fetched stage runs** — the round's own spec scenario ("a change
with more stage runs than one page holds") is exactly this, and summing a page is the defect it
names. The test seeds a stage-run page shorter than the aggregate's own run count and asserts the
header still shows the aggregate's figures.

`App.tsx`'s hash router gains the parametrised route alongside its eight static ones, and
`RunDetail.tsx` composes header, timeline and table.

- [x] **Step 6: Make it reachable from the board**

`StateBoard.tsx`'s change column becomes a link to that change's route. The test asserts the link's
`href` for a row, including that a project key or change name containing a URL-significant character
survives the round trip — the board is the only navigation path to the route, so a wrongly built
link is the whole feature failing silently.

**Tests:** `stats/web/src/metrics.test.ts` covers the readers' absence, wrong-type and unknown-key
behaviour; `stats/web/src/api.test.ts` gains the `listStageRuns` query-building cases;
`stats/web/src/views/RunDetail.test.tsx` covers the route rendering, the open-run case, the
unmeasured-metrics case, and the aggregate-not-a-page-sum case; `stats/web/src/views/views.test.tsx`
gains the state-board link case. Verification is `cd stats/web && npm test` and
`cd stats/web && npx tsc -b`.

**Regression:** Reverting this task returns the interface to having no way to open a single change —
the exact gap this round exists to close — and leaves `GET /api/v1/stage-runs` with no caller in the
interface at all.

**Baseline:** before=38 after=38+ SPA cases (`cd stats/web && npm test`).
<!-- predicted: the new cases in the four test files this task names, none of them written yet -->

**Commit:** `feat(18): open one change on its own dashboard of stage runs`

---

### 19 The Grafana chrome — panels, and the controls that scope them

**Build:** green

**Files:**
- Add: `stats/web/src/components/Panel.tsx`
- Add: `stats/web/src/components/Panel.test.tsx`
- Add: `stats/web/src/components/StatPanel.tsx`
- Add: `stats/web/src/components/TimeSeriesPanel.tsx`
- Add: `stats/web/src/components/DashboardBar.tsx`
- Add: `stats/web/src/components/DashboardBar.test.tsx`
- Add: `stats/web/src/components/ChangeVariable.tsx`
- Add: `stats/web/src/components/ModelVariable.tsx`
- Add: `stats/web/src/theme.css`
- Modify: `stats/web/src/App.tsx`
- Modify: `stats/web/src/api.ts`
- Modify: `stats/web/src/api.test.ts`
- Modify: `stats/web/src/components/PeriodPicker.tsx`
- Modify: `stats/web/src/components/ProjectFilter.tsx`
- Modify: `stats/web/src/hooks/useStatsView.ts`
- Modify: `stats/web/src/viewTypes.ts`
- Modify: `stats/web/src/styles.css`

**Interfaces:**
- Consumes: the existing `Period` and project state `App.tsx` already owns and passes to every view,
  and **task 21's `model` parameter and `GET /api/v1/models`** — this task is implemented after task
  21, not in plan order, because its model variable has nothing to call until that task lands.
- Produces: `Panel`, `StatPanel` and `TimeSeriesPanel` primitives, and a `DashboardBar` carrying the
  time range and the template variables.

- [x] **Step 1: The dark palette, as tokens**

`theme.css` declares the palette, spacing scale, panel border, radius and typography as custom
properties, and `styles.css` consumes them. No component hard-codes a colour. Grafana's own dark
theme is the reference for the values.

- [x] **Step 2: The panel primitive**

`Panel.tsx` is a bordered region with a heading, an optional description, and a body — the unit the
round's spec calls "one self-contained result with its own heading". It renders the loading, error
and **not-recorded** branches `ViewFrame` renders today, so recomposing a view into panels cannot
lose the not-recorded distinction on the way. Its test asserts exactly that: a panel handed a
`recorded: false` response renders the not-recorded banner and never its children.

- [x] **Step 3: Stat and time-series panels**

`StatPanel.tsx` renders one large number with a label and an optional sparkline.
`TimeSeriesPanel.tsx` renders a line chart from `{ day, value }` points as inline SVG — no charting
dependency is added, because a single line with axes is well inside what SVG does directly, and
pulling in a library for it is the "config switch with one implementation" KISS violation
`engineering-principles.md` names.

**A `StatPanel` given `null` renders "unavailable", never `0`,** and its test asserts that against
`0`, `null` and `undefined` separately — the three must not collapse.

- [x] **Step 4: The dashboard bar**

`DashboardBar.tsx` holds the time-range picker (the existing `PeriodPicker`, restyled) and the
template variables: **project** (the existing `ProjectFilter`, restyled), **model**
(`ModelVariable.tsx`, populated from task 21's `GET /api/v1/models` for the current period and
project — never a hard-coded list, so a model this build has never heard of appears the first time
it is used), and **change** (`ChangeVariable.tsx`, a navigation variable — selecting a change routes
to its dashboard rather than filtering the current one).

`useStatsView` and `ViewProps` carry the selected model through to the request, so a view's rows are
the server's answer for the model as it currently reads. **No view filters rows by model after
receiving them** — that is the defect the round's spec forbids, and it under-reports the moment a
result is paged.

**The model variable is disabled, with a stated reason, on a dashboard that cannot accept it.** The
live state board's rows are changes rather than stage runs, and task 21 has the API reject a model
restriction there rather than ignore it; the bar must not offer a control whose only possible
outcome on that dashboard is a 400. The test asserts the disabled state on the state board and the
enabled state on a stage-run dashboard.

The bar is rendered once by `App.tsx`, above the active dashboard, so every panel beneath it reads
the same period and the same project by construction — the spec's "no two panels can be showing
different periods" is a property of there being one control, not of a synchronisation step.

Its test asserts that changing the range calls the change handler with the new range, and that
choosing a change navigates rather than filtering.

**Tests:** `stats/web/src/components/Panel.test.tsx` covers the panel's loading/error/not-recorded
branches and the null-versus-zero rule in `StatPanel`;
`stats/web/src/components/DashboardBar.test.tsx` covers the range and variable handlers, the model
variable's population from the server and its disabled state on the live state board;
`stats/web/src/api.test.ts` gains the `model` parameter's presence in a built stats query and its
absence when no model is selected. Verification is `cd stats/web && npm test` and
`cd stats/web && npx tsc -b`.

**Regression:** Reverting this task removes the panel primitives task 20 composes with, and returns
the chrome to a per-page heading with no shared bar.

**Baseline:** before=38+ after=38++ SPA cases (`cd stats/web && npm test`).
<!-- predicted: the new cases in the two test files this task names, none of them written yet -->

**Commit:** `feat(19): add the dashboard chrome — panels and the controls that scope them`

---

### 20 Recompose the nine dashboards onto the panels

**Build:** green

**Files:**
- Modify: `stats/web/src/views/StateBoard.tsx`
- Modify: `stats/web/src/views/CostPerChange.tsx`
- Modify: `stats/web/src/views/StageLeaderboard.tsx`
- Modify: `stats/web/src/views/Trend.tsx`
- Modify: `stats/web/src/views/CacheEfficiency.tsx`
- Modify: `stats/web/src/views/PanelEconomics.tsx`
- Modify: `stats/web/src/views/ModelComparison.tsx`
- Modify: `stats/web/src/views/ReworkRate.tsx`
- Modify: `stats/web/src/views/RunDetail.tsx`
- Modify: `stats/web/src/views/views.test.tsx`
- Modify: `stats/web/src/components/ViewFrame.tsx`
- Modify: `stats/web/src/components/DataTable.tsx`
- Modify: `stats/web/src/components/DashboardBar.tsx`
- Modify: `stats/web/src/hooks/useRunDetail.ts`
- Modify: `stats/web/src/views/RunDetail.test.tsx`
- Modify: `stats/web/src/App.tsx`
- Modify: `stats/web/src/styles.css`

**Interfaces:**
- Consumes: task 19's `Panel`, `StatPanel` and `TimeSeriesPanel`.
- Produces: nine dashboards — the eight views plus the run detail — each a panel composition.

- [x] **Step 1: Give each dashboard its summary panels**

Each of the eight views gains stat panels above its table, computed from **the rows the server
already returned for that view** — a count, a total and a mean where the view's own question has
one. Where a view's question has no meaningful single number, it gets no stat panel rather than an
invented one.

**A stat panel derived from rows is only correct where those rows are the whole answer.** Every
statistics view returns its complete result set for the period — none of the eight is paged — which
is why summarising them here is sound, and why the run detail dashboard (whose stage-run list *is*
paged) takes its numbers from the server aggregate instead, exactly as task 18 built it. The test
asserts that difference directly, so the two cannot be quietly harmonised into the wrong one.

- [x] **Step 2: Trend becomes a time-series panel**

`Trend.tsx` renders its daily points through `TimeSeriesPanel` and keeps its table beneath. A day
with a `null` cost is a gap in the line, not a point at zero.

- [x] **Step 3: Tables become table panels**

`ViewFrame.tsx` is reduced to what `Panel` does not cover, or removed where `Panel` covers all of
it; `DataTable` renders inside a panel body. The not-recorded branch stays exactly one
implementation — whichever file owns it after this step, it is not duplicated across two.

- [x] **Step 4: The run detail dashboard, recomposed**

`RunDetail.tsx`'s header becomes stat panels, its timeline a panel, its table a panel.

**Settle the model variable on this route.** Task 19 renders `DashboardBar` unconditionally, so the
model dropdown appears here while `RunDetail` consumes nothing from it — a control that silently does
nothing, which is the defect this round exists to remove. Choose one and make it true: either the
run detail honours the model (its stage-run table showing only runs that used it, and its header
numbers that model's own buckets), **or** the variable is disabled here with a stated reason, the
same way it is on the live state board. Do not leave it rendering and inert.

Honouring it is the better answer where it is cheap — a stage run's `models` bag already says which
models it used — but a disabled control with an honest reason is a correct outcome too. What is not
available is the third option.

**Tests:** `stats/web/src/views/views.test.tsx` gains, per view, an assertion that its panels render
the server's own figures; the run-detail case from task 18 is extended to assert its header still
reads from the aggregate after the recomposition. Verification is `cd stats/web && npm test` and
`cd stats/web && npx tsc -b`.

**Regression:** Reverting this task leaves the panel primitives unused and the views as bare tables
— the round's second requirement unmet while its first still holds.

**Baseline:** before=38++ after=38+++ SPA cases (`cd stats/web && npm test`).
<!-- predicted: the new per-view cases in views.test.tsx, none of them written yet -->

**Commit:** `refactor(20): recompose every dashboard onto the panel primitives`

---

### 21 The statistics API gains a model filter

**Build:** green

**Implemented before task 19**, not in plan order: task 19's model variable has nothing to call
until this lands. Plan order is the reading order; the dispatch order is 18, 21, 19, 20.

**Files:**
- Modify: `stats/internal/store/aggregate.go`
- Modify: `stats/internal/store/aggregate_test.go`
- Modify: `stats/internal/api/stats.go`
- Modify: `stats/internal/api/stats_test.go`
- Modify: `stats/internal/api/server.go`
- Modify: `stats/web/src/api.ts`
- Modify: `stats/web/src/api.test.ts`
- Modify: `stats/web/src/hooks/useRunDetail.ts`
- Modify: `stats/internal/client/client_test.go`
- Modify: `stats/internal/web/embed_test.go`
- Modify: `stats/internal/store/harvestshape_test.go`

The last three were not anticipated and are recorded after the fact. Adding `model *string` to the
seven store methods changes the `StatsStore` interface, so every fake and call site implementing it
moves with it — mechanical signature updates carrying no behaviour. Recording them is the correct
repair for an incomplete declaration.

**Interfaces:**
- Consumes: task 22's per-model buckets (`metrics->'models'`) — the only correct record of which
  models a stage run used.
- Produces: a `model` parameter on the seven stage-run views, `GET /api/v1/models`, and an
  `excludedNoModel` count in the statistics envelope.

- [x] **Step 1: The store learns to scope by model**

The seven aggregations over stage runs — `CostPerChange`, `StageLeaderboard`, `TrendOverTime`,
`CacheEfficiency`, `PanelEconomics`, `ModelComparison`, `ReworkRate` — take a `model *string`
alongside the `project *string` they already take, and restrict on the **per-model buckets task 22
introduced** (`sr.metrics->'models' ? $n`) when it is non-nil.

**Not `metrics->>'model'`.** That scalar was retired in task 22 and was dead code before that — it
was never written by anything. A filter against it would match nothing at all, and would silently
return empty results that read as "this model was never used". `LiveStateBoard` does **not**: its rows are changes, not stage runs, and giving it a
parameter it cannot honour is exactly the silently-inert control the round's spec forbids.

**An explicit parameter, not a new `Scope` struct.** One filter does not earn a container, and a
struct would change every call site for no behaviour. When a third scoping dimension arrives, that
is when the struct earns its place — not in anticipation of it.

**The parameter is bound, never interpolated.** `internal/store`'s existing rule holds: the model
value is a bind parameter like `project` is. A model name is caller-supplied text and must never
reach an identifier position — see `query.go`'s allowlist rationale for the standard this repeats.

Tests, against the seeded database the package's other aggregation tests already use: a filter
matching some runs returns only those; a filter matching none returns an empty result rather than an
error; a filter is honoured on every one of the seven methods, iterated rather than spot-checked, so
a method added later without the filter fails here.

- [x] **Step 2: Absence is not a match, and is counted**

A stage run whose metrics bag has no `model` key, or has it as JSON `null`, matches **no** model
filter — `metrics->>'model' = $n` is already NULL-safe in that direction, and the test asserts it
rather than assuming it.

New `CountRunsWithoutModel(ctx, period, project)` returns how many stage runs in scope recorded no
model. It is called **only when a model filter is set**, so an unfiltered request pays nothing for
it.

The test seeds runs with and without a recorded model and asserts the count covers exactly the
latter — and that it is unaffected by which model was filtered for, since it counts runs that could
not have matched *any* model.

- [x] **Step 3: The view endpoint accepts it, and the state board rejects it**

`statsQueryParams` gains `"model"`. The handler passes it to the seven, and **rejects it on
`state-board` with 400** naming the view — the same shape `breakdown` is already rejected with on
the views that do not offer it. Accepting and ignoring it is the one outcome not available: a
control that silently does nothing is worse than one that says it cannot apply.

The envelope gains `excludedNoModel`, an integer present only on a response that applied a model
filter, absent otherwise — absent meaning "no filter was applied", never "zero were excluded", which
is this change's own absence-is-never-zero rule applied to its newest field.

Tests: a filtered view returns only that model's rows; `state-board` with a model returns 400 naming
the view; an unfiltered response carries no `excludedNoModel`; a filtered response carries the count
including when it is genuinely `0`.

- [x] **Step 3b: A filtered view reports that model's numbers, not the whole run's**

Restricting to a model selects runs that **used** it — and the token and currency figures those rows
report are **that model's own bucket**, not the run's total. A review-panel run costing $61.10 across
Opus and Sonnet, filtered to Sonnet, reports Sonnet's $19.90: not $61.10 (which would attribute the
Opus parent to Sonnet) and not $0 (which would read as "Sonnet was never used"). This is the whole
point of task 22's buckets, and without it the filter is decorative.

Where a run used the model but that bucket carries no priced cost, the row's cost is **unavailable**
rather than zero, exactly as everywhere else.

The test asserts the $19.90 case directly against a seeded two-model run — both wrong answers
($61.10 and $0) must fail it.

- [x] **Step 4: `GET /api/v1/models`**

A new route on the existing mux, returning the distinct non-null models recorded in a period,
optionally scoped by project, sorted. It takes `from`, `to` and `project` and rejects any other
parameter, exactly as the view endpoint does. It goes through `withDaemonHeader` like every other
route, so a caller can still tell the daemon's own answer from a proxy's.

The test asserts: the distinct set for a seeded period; that a model recorded only outside the
period is absent; that runs with no model contribute nothing; and that an unknown parameter is
rejected.

- [x] **Step 5: Let `change` scope `cost-per-change` on its own**

`buildStatsViewQuery` currently rejects `change` unless `breakdown=repo` accompanies it, so task 18's
run-detail header fetches **every** change's rows for the project and filters to one in the browser.
That is arithmetically correct only because `cost-per-change` is unpaged — the moment anyone adds a
limit there, the header under-reports and nothing fails. Accept `change` alone on `cost-per-change`,
server-side, and have `useRunDetail` send it.

The pairing rule for `breakdown=repo` is unchanged: it still requires `change`, `command` and
`stage` together. This step widens what `change` may appear *without*, not what `breakdown` needs.

Files this step adds to the list: `stats/web/src/api.ts`, `stats/web/src/api.test.ts`,
`stats/web/src/hooks/useRunDetail.ts`.

**Tests:** `stats/internal/store/aggregate_test.go` covers the filter across all seven methods, the
no-model exclusion, and `CountRunsWithoutModel`; `stats/internal/api/stats_test.go` covers the
parameter's acceptance, the state-board rejection, the `excludedNoModel` presence rule, and the new
route. Verification is `cd stats && go test ./... -race -count=1`, `cd stats && gofmt -l .` and
`cd stats && go vet ./...`.

**Regression:** Reverting this task leaves the model variable in the dashboard bar with no server
parameter to send, which would either break the request or push the filter into the client — the
row-discarding shape the round's spec forbids.

**Baseline:** before=253 after=253+ top-level Go test functions, measured as
`cd stats && go test ./... -count=1 -v | grep -c '^--- PASS'`.
<!-- measured: that command reported 253 against commit `c2357c5` on 2026-08-14 -->

**Commit:** `feat(21): restrict the statistics views to one model, server-side`

---

### 22 Per-model token buckets, and a cost path that prices them

**Build:** green

**Implemented before tasks 21, 19 and 20.** The dispatch order for this round is 18, 22, 23, 21,
19, 20; plan order is reading order. See the round preamble.

**Files:**
- Modify: `stats/internal/harvest/attribute.go`
- Modify: `stats/internal/harvest/attribute_test.go`
- Modify: `stats/internal/harvest/watcher.go`
- Modify: `stats/internal/harvest/watcher_test.go`
- Modify: `stats/internal/store/pricing.go`
- Modify: `stats/internal/store/stageruns_test.go`
- Modify: `stats/internal/store/aggregate.go`
- Modify: `stats/internal/store/aggregate_test.go`
- Modify: `stats/internal/api/stats_test.go`

The ninth file was not anticipated when this task was written and is recorded here after the fact,
which is the correct repair for a declaration that turned out incomplete — never a silent
discrepancy between the plan and the commit. `TestBoundaryConventionIsConsistentAcrossViews` seeded
the retired scalar `model` key, invisible to `ModelComparison` once step 2 lands, so its fixture had
to move to the bucket shape. No production code in `internal/api` was touched.

**Interfaces:**
- Consumes: `jsonb_deep_add` (migration `0005`), which sums numeric leaves at any depth — the reason
  this needs no schema change.
- Produces: a `models` object in the metrics bag, keyed by the model string, and a `Price` that
  computes each bucket's cost separately.

**Why this task exists.** A stage run's `model` is written last-write-wins per harvest batch
(`attribute.go`'s `MetricsPatch` doc comment), so a run records exactly one model. Under myflow that
is wrong for the common case rather than an edge case: `/myflow-do`'s review panel runs its parent
on one model and every reviewer slot on another, and the implementer and panel-fix roles are
separately configurable (`models.implementation`, `models.reviewPanel`, `models.panelFix` —
**Model policy**, `skills/myflow-contracts/pipeline.md`). Every panel stage run is therefore a
mixed-model run reported as single-model, which makes `ModelComparison`'s per-model mean cost wrong
today and would make task 21's model filter wrong on arrival.

- [x] **Step 1: Attribute tokens per model, not just per stage run**

`Attribute` already buckets tokens by main versus sidechain. It gains a second dimension: the model
the record's own message carried. `MetricsPatch` gains

```json unverified:the bag shape this task introduces, not one read from existing code; confirm jsonb_deep_add sums these leaves at this depth before relying on it
models: { "<model>": { tokens: { input, output, cache_read, cache_creation, main, sidechain } } }
```

and the existing top-level `tokens` is **unchanged** — it stays the whole-run total, so every
aggregation reading it keeps working and this task changes no view's meaning by accident.

`jsonb_deep_add` sums numeric leaves at any depth, so two batches touching the same run and the same
model add rather than replace, with no merge code and no migration. The test asserts exactly that:
two batches, overlapping models, summed leaves.

**A record whose message carried no model contributes to the top-level `tokens` and to no bucket.**
It is not filed under an `"unknown"` key — inventing a model name to hold unmeasured usage is the
absence-is-never-zero rule broken at its source, and it would make that fabricated model appear in
task 21's dropdown as though it were real.

- [x] **Step 2: Retire the scalar `model` from aggregation**

`MetricsPatch.Model` stops being written. `ModelComparison` reads the buckets — one row per
(model, command, stage), with each model's own token totals — instead of grouping by
`metrics->>'model'`. Its question is unchanged; only its source is now correct for mixed-model runs.

The test that matters: a single stage run carrying two models produces **two** rows whose token
totals are each model's own, and which sum to the run's top-level `tokens`. Under the old scalar
that run produced one row attributing everything to whichever model spoke last, so this test fails
against the previous implementation — which is what makes it a real test rather than a restatement.

`MetricsPatch.Effort` stays as it is: effort is a property of the run's own configuration, not of
each message, and nothing aggregates by it.

- [x] **Step 3: Price each bucket separately**

`Price` currently reads one model and one token bag. It now prices **each** bucket against the rate
in effect for that bucket's model at the run's start, writes `models.<model>.cost_usd`, and writes
the top-level `cost_usd` as their sum — so every existing aggregation over `cost_usd` keeps working
and now reports a mixed-model run's real total rather than one model's rate applied to every token.

**A run with one unpriceable bucket is not priced at zero for that bucket, and its total is not
silently short.** Where any bucket has no rate in effect, `Price` writes the buckets it could price,
omits the ones it could not, and **omits the top-level `cost_usd` entirely** rather than writing a
partial sum that reads like a complete one. `ErrPricingNotFound` names the models it could not
price. A partial total is the most dangerous possible output here: it is indistinguishable from a
correct one at every layer above.

Tests: a two-model run priced against two different rates; a run where one model has no rate,
asserting the partial buckets are written and the top-level total is absent; the existing
effective-dating behaviour still holding per bucket.

**Tests:** `attribute_test.go` covers per-model bucketing, the no-model record, and cross-batch
summing; `aggregate_test.go` covers the two-row mixed-model `ModelComparison` case;
`stageruns_test.go` covers the per-bucket pricing and the absent-total rule. Verification is
`cd stats && go test ./... -race -count=1`, `cd stats && gofmt -l .`, `cd stats && go vet ./...`.

**Regression:** Reverting this task restores single-model attribution, which reports every review
panel's whole cost under one model and makes task 21's filter and `ModelComparison` both wrong.

**Baseline:** before=253 after=253+ top-level Go test functions, measured as
`cd stats && go test ./... -count=1 -v | grep -c '^--- PASS'`.
<!-- measured: that command reported 253 against commit `c2357c5` on 2026-08-14 -->

**Commit:** `fix(22): attribute tokens and cost per model, not per stage run`

---

### 23 Wire the cost path, and price it against the rates actually charged

**Build:** green

**Files:**
- Add: `stats/internal/store/migrations/0007_pricing_rate_shape.sql`
- Add: `stats/internal/store/pricing_seed.go`
- Add: `stats/internal/store/pricing_seed_test.go`
- Modify: `stats/internal/store/pricing.go`
- Modify: `stats/internal/store/stageruns_test.go`
- Modify: `stats/internal/harvest/transcript.go`
- Modify: `stats/internal/harvest/attribute.go`
- Modify: `stats/internal/harvest/transcript_test.go`
- Modify: `stats/internal/harvest/attribute_test.go`
- Modify: `stats/internal/harvest/watcher.go`
- Modify: `stats/internal/harvest/watcher_test.go`
- Modify: `stats/cmd/myflowd/main.go`
- Modify: `stats/README.md`

**Interfaces:**
- Consumes: task 22's per-model buckets.
- Produces: the first `cost_usd` that has ever existed outside a test.

**Why this task exists — three defects, each of which alone makes every currency figure wrong.**

**(a) Nothing calls `Price`, and no rate is ever seeded.** `Store.Price` and `Store.PutPricing` are
called from `stageruns_test.go` and nowhere else. In the running daemon the `pricing` table is empty
and `Price` is never invoked, so no stage run ever receives a `cost_usd` and every currency figure in
every view is permanently unavailable.

**(b) The cache-write rate is one column, but two rates are charged.** A 5-minute cache write costs
**1.25x** base input; a 1-hour cache write costs **2x**. The `pricing` table has a single
`cache_write_per_mtok`, and `internal/harvest` reads only the collapsed
`usage.cache_creation_input_tokens`, discarding the `usage.cache_creation.{ephemeral_5m,ephemeral_1h}`
split that says which rate applies. **This machine's sessions are ~100% 1-hour writes**, and cache
creation dominates the token mix, so pricing them at the 5m rate understates the dominant cost
component by 37.5%.

<!-- measured: every sampled assistant entry in this session's transcript carried ephemeral_1h_input_tokens > 0 and ephemeral_5m_input_tokens == 0, on 2026-08-14; one sampled entry showed cache_creation 25984 against input 2 -->

**(c) Fast mode doubles the rate and is not recorded.** Opus 5 and Opus 4.8 in fast mode are
$10/$50 per MTok against $5/$25 standard. `usage.speed` carries `"standard"` or `"fast"` and nothing
reads it — while "fast mode used" is one of the metrics the linked issue asks for by name.

**How twenty-one review passes missed (a).** Every pricing test seeds its own rates and calls `Price`
directly, so the unit tests pass and prove the function correct in isolation. Nothing asserted that
anything in production calls it. Step 5 is the durable repair: an assertion about the wiring, not
about the function.

- [x] **Step 1: Record what determines the rate**

`transcript.go` reads `usage.cache_creation.ephemeral_5m_input_tokens`,
`usage.cache_creation.ephemeral_1h_input_tokens` and `usage.speed`. `attribute.go`'s `Bucket` splits
`cache_creation` into `cache_creation_5m` and `cache_creation_1h`, and `MetricsPatch` records the
speed.

**The existing `cache_creation` total stays and keeps its meaning.** Task 24's aggregations read it,
and it is the sum of the two new fields. Adding the split alongside it changes no view.

**A transcript that carries the collapsed total but no split is not silently priced at the cheaper
rate.** It records the total with neither sub-field, and step 4 prices it as unavailable rather than
guessing which cache it was — guessing here is a 60% error in the dominant token class.

- [x] **Step 2: The pricing table learns the rates that exist**

Migration `0007` adds `cache_write_5m_per_mtok`, `cache_write_1h_per_mtok`,
`fast_input_per_mtok` and `fast_output_per_mtok`, and backfills the existing
`cache_write_per_mtok` into the 5m column so no seeded row loses meaning. The old column is left in
place and stops being read — dropping a column is a separate decision from ceasing to depend on it.

- [x] **Step 3: Seed the real rates**

`pricing_seed.go` carries the published rates, **read from Anthropic's pricing page on 2026-08-14**
(`https://platform.claude.com/docs/en/about-claude/pricing`) with that URL and date in a comment
beside the table:

<!-- verified: read from https://platform.claude.com/docs/en/about-claude/pricing on 2026-08-14 -->

| Model | input | output | 5m write | 1h write | cache read | fast in | fast out |
|-------|-------|--------|----------|----------|------------|---------|----------|
| `claude-opus-5` | 5 | 25 | 6.25 | 10 | 0.50 | 10 | 50 |
| `claude-opus-4-8` | 5 | 25 | 6.25 | 10 | 0.50 | 10 | 50 |
| `claude-sonnet-5` | 2 | 10 | 2.50 | 4 | 0.20 | — | — |
| `claude-haiku-4-5` | 1 | 5 | 1.25 | 2 | 0.10 | — | — |

`myflowd` upserts them at startup through the existing `PutPricing`, whose `ON CONFLICT` already
makes re-seeding a no-op.

**`effective_from` is set early enough to cover this store's whole recorded history**, and the seed
says so in a comment: the store started empty in August 2026, so one row per model covers all of it
at the current rates. The table is keyed `(model, effective_from)` so a future rate change is an
insert, not a migration.

**A model absent from the seed prices as unavailable rather than being given an invented rate.**

- [x] **Step 4: Price each bucket against the rate that applies to it**

`Price` charges 5m and 1h cache writes at their own rates, and uses the fast-mode input/output rates
when the run recorded `speed: "fast"`. A run whose cache split is unknown, or whose model has no fast
rate while recording fast speed, is **unpriceable for that bucket** — and task 22's rule already
holds: the buckets that could be priced are written, the ones that could not are omitted, and the
top-level `cost_usd` is omitted entirely rather than written as a partial sum.

- [x] **Step 5: Call it, and test that it is called**

After `CommitHarvestBatch` reports `applied`, `RunOnce` prices every stage run in that batch's
deltas — **after** the commit, never inside it, because the commit's atomicity is what makes
harvesting exactly-once and pricing is a pure recomputation from stored metrics. A pricing failure is
logged and skipped, never fatal.

The test runs a **whole harvest cycle** over a seeded transcript and asserts the touched runs carry a
`cost_usd`. **Write it first and watch it fail against current code.** A test that calls `Price`
directly cannot fail when nothing calls `Price` — which is exactly why this defect survived.

**Tests:** `transcript_test.go` and `attribute_test.go` cover the cache split and speed capture,
including the absent-split case; `pricing_seed_test.go` asserts every seeded row round-trips and that
re-seeding is idempotent; `stageruns_test.go` covers 1h-versus-5m pricing and fast-mode pricing;
`watcher_test.go` carries the end-to-end wiring assertion and the pricing-failure-is-not-fatal case.
Verification is `cd stats && go test ./... -race -count=1`, `gofmt -l .`, `go vet ./...`, plus a live
check that a real stage run acquires a `cost_usd`.

**Regression:** Reverting this task returns every currency figure to permanently unavailable while
every token figure keeps working — and reverting step 1 alone silently underprices the dominant token
class by 37.5%, which is worse, because it looks like a number.

**Baseline:** before=271 after=271+ top-level Go test functions, measured as
`cd stats && go test ./... -count=1 -v | grep -c '^--- PASS'`.
<!-- measured: task 21's own run reported 271 against commit `c0b4ed6` on 2026-08-14 -->

**Commit:** `fix(23): price harvested runs against the rates actually charged`

---

### 24 The aggregation SQL reads a metrics shape the harvester has never written

**Build:** green

**Implemented before task 23.** Dispatch order for this round: 18, 22, 24, 23, 21, 19, 20.

**Files:**
- Modify: `stats/internal/store/aggregate.go`
- Modify: `stats/internal/store/aggregate_test.go`
- Modify: `stats/internal/store/stageruns_test.go`
- Add: `stats/internal/store/harvestshape_test.go`

`pricing.go` was declared here when this task was written and is **removed from the list after the
fact**: task 22 had already rewired `Price` to decode `models.<model>.tokens` against the real
nested `Bucket` shape, so it reads no top-level `tokens` path at all and the claim below that it
"still reads one" was wrong. Correcting the declaration is the right repair — never a no-op edit to
make a file list come true.

**Interfaces:**
- Consumes: `internal/harvest`'s own `MetricsPatch` shape — as a *value produced by that package*,
  not as a hand-written literal.
- Produces: aggregations that read the shape actually stored.

**Why this task exists — a Critical defect in what already shipped.** `internal/harvest` writes

```json verified:read from internal/harvest/attribute.go's TokenDelta and Bucket types on 2026-08-14
{"tokens": {"main": {"input": …, "output": …, "cache_creation": …, "cache_read": …, "thinking": …}, "sidechain": {…}}}
```

`internal/store/aggregate.go` reads `tokens->>'input'`, `tokens->>'cache_read'`,
`tokens->>'cache_creation'` — paths that **do not exist** at that level — and `tokens->>'main'`,
`tokens->>'sidechain'`, which exist but are **objects**. Against real harvested data:

- `CostPerChange` **fails outright**: `('{"tokens":{"main":{"input":5}}}'::jsonb->'tokens'->>'main')::numeric`
  raises `invalid input syntax for type numeric`, so the whole view errors rather than degrading.
- `CacheEfficiency` returns NULL for every row.
- `PanelEconomics` totals 0 tokens, making findings-per-Mtok meaningless rather than absent.
- `CostPerChange`'s input totals are NULL.

<!-- verified: the cast error above was reproduced against the running myflow-postgres container on 2026-08-14 -->

**How it survived twenty-one review passes.** Every store test hand-seeds the flat shape its own SQL
expects (`{"tokens":{"input":100}}`, `{"tokens":{"main":1000,"sidechain":4000}}`), and every harvest
test asserts against the nested shape its own code produces. Both suites are internally consistent
and neither has ever seen the other's output. This is the same seam class as task 23's unwired
`Price` — a component correct in isolation, connected to nothing that checks it — and the same class
as the seven vacuous tests this change already found. **Fixing only the SQL would leave the seam
uncovered and the next drift equally invisible**, which is why step 3, not step 1, is the point of
this task.

- [x] **Step 1: Read the shape that is actually written**

`total_tokens_input` sums across both buckets rather than reading a key that does not exist;
`main_tokens` and `sidechain_tokens` sum each bucket's own fields; `CacheEfficiency` and
`PanelEconomics` read their token types inside the buckets. Where a bucket is absent the result is
NULL, not 0 — the absence-is-never-zero rule holds here exactly as everywhere else, and
`PanelEconomics`'s current `COALESCE(…, 0)` is part of the defect rather than a defence against it.

- [x] **Step 2: Fix the fixtures that encoded the wrong shape**

Every hand-written `{"tokens": …}` literal in `aggregate_test.go` and `stageruns_test.go` moves to
the real shape. **A fixture asserting the old flat shape is not evidence of anything** — it is the
thing that hid this defect, so none of them is kept "for coverage".

- [x] **Step 3: Close the seam so it cannot drift again**

New `harvestshape_test.go` builds its fixture **by marshalling `internal/harvest`'s own
`MetricsPatch`** rather than by writing JSON by hand, commits it through the ordinary store path, and
asserts every aggregation returns real numbers over it.

**This is the durable repair.** A hand-written literal is a second, private copy of the contract
between two packages, and this defect is what happens when the copies drift. Sourcing the fixture
from the producing package means a future change to `MetricsPatch` that breaks an aggregation fails
here, in this repository, rather than silently in a view nobody queried.

Write this test **first** and watch it fail against the current `aggregate.go` — the failure is the
proof it guards the defect. A version of this test that passes before step 1 is testing the wrong
thing.

**Tests:** `harvestshape_test.go` is the new seam test; `aggregate_test.go` and `stageruns_test.go`
have their fixtures corrected and keep their existing assertions. Verification is
`cd stats && go test ./... -race -count=1`, `cd stats && gofmt -l .`, `cd stats && go vet ./...`.

**Regression:** Reverting this task returns the statistics half of the application to erroring or
reporting NULL against every real harvested run, while its whole test suite stays green — the exact
condition this task was found in.

**Baseline:** before=260 after=260+ top-level Go test functions, measured as
`cd stats && go test ./... -count=1 -v | grep -c '^--- PASS'`.
<!-- measured: task 22's own run reported 260 against commit `567a70f` on 2026-08-14 -->

**Commit:** `fix(24): read the metrics shape the harvester actually writes`

---

### 25 A parameter that cannot apply says so

**Build:** green

**Implemented last in this round**, after task 20 — it touches files tasks 19 and 20 are both editing.

**Files:**
- Modify: `stats/internal/api/stats.go`
- Modify: `stats/internal/api/stats_test.go`
- Modify: `stats/web/src/hooks/useRunDetail.ts`
- Modify: `stats/web/src/views/RunDetail.test.tsx`
- Modify: `stats/web/src/metrics.ts`
- Modify: `stats/web/src/metrics.test.ts`
- Modify: `stats/web/src/components/StageRunTable.tsx`
- Modify: `stats/web/src/views/RunDetail.tsx`

`RunDetail.tsx` was not anticipated and is recorded after the fact: it is the only caller of
`StageRunTable`, so step 4's shortfall notice had to be forwarded through it or it would have been
computed and never rendered.

**Interfaces:**
- Consumes: task 21's `model` rejection on `state-board`, whose shape this generalises.
- Produces: one rule for every scoping parameter, rather than one rule for the newest one.

**Why this task exists.** Task 21 established that a view rejects a scoping parameter it cannot
honour, and the spec requirement **Views are restrictable to one model** states it — but only for
`model`. Two older cases still break it, both found while implementing this round:

- **`breakdown` is silently ignored on the six views that do not support it.** It is in
  `statsQueryParams`, so the unknown-parameter guard passes it, and only `cost-per-change` ever reads
  it. `GET /api/v1/stats/trend?breakdown=repo` returns an ordinary trend and reports nothing about
  the parameter it discarded. The plan for task 21 asserted the opposite was already true; it was
  not.
- **`useRunDetail` still filters `cost-per-change` rows by change name in the browser**, left in
  place when task 21 moved that scoping server-side because an out-of-scope test mocks the fetch
  layer directly. It is a no-op today — the server already returns only that change — which is
  exactly what makes it worth removing: dead defensive code that appears to be load-bearing invites
  the next reader to preserve it.

- [x] **Step 1: Reject `breakdown` where it cannot apply**

`breakdown` on any view other than `cost-per-change` returns 400 naming the view, exactly as `model`
does on `state-board`. The existing `breakdown=repo` pairing rules on `cost-per-change` are
unchanged.

**This is a deliberate behaviour change to an endpoint that previously accepted the request**, and it
is the right one: a caller who sent `breakdown` believed it applied, and returning an unbreakdowned
answer with a 200 tells them it did.

- [x] **Step 2: Remove the client-side change filter**

`useRunDetail` stops filtering by `changeName`. `RunDetail.test.tsx`'s fetch mock is updated to
return what the server would actually return for a scoped request, so the test exercises the real
contract rather than compensating for a filter that no longer exists.

**A test whose mock returns rows the server could not return is not covering the server's
behaviour** — it was what made the dead filter look necessary.

- [x] **Step 3: `readModel` reads a key nothing writes**

`metrics.ts`'s `readModel()` reads the top-level scalar `model`, which task 22 retired and which was
dead code even before that — nothing ever wrote it. The run-detail table's model column is therefore
always empty. It reads the `models` bucket keys instead, which is where a run's models actually live,
and reports **all** of them for a mixed-model run rather than picking one.

A run using two models is the common case for a review-panel stage; showing one of them, or none,
was the original misattribution this round set out to fix, resurfacing at the last layer.

The test that matters is the one that would have caught it: a fixture built from the shape the
harvester writes, asserting a model name reaches the column. The existing `metrics.test.ts` cases
pass against a hand-written bag carrying the retired key — the same private-copy problem task 24's
seam test was written to end.

- [x] **Step 4: The stage-run table's model filter cannot silently shorten a page**

Task 20 filters the fetched stage-run page against each run's `models` keys in the browser. Where a
change's stage runs exceed one page, that shows only the matching runs *within the page fetched* —
so a change with thirty Sonnet runs can display three, with nothing saying so.

Either request the restriction server-side, or state the shortfall in the panel. **Do not leave a
count that looks complete and is not** — the header numbers come from the server aggregate and will
disagree with the visible rows, which is the confusing half of the failure.

**Tests:** `stats_test.go` covers the rejection on a view that does not support `breakdown` and its
continued acceptance on `cost-per-change`; `RunDetail.test.tsx` covers the header numbers against a
correctly-scoped mock. Verification is `cd stats && go test ./... -race -count=1`,
`cd stats/web && npm test`, `cd stats/web && npx tsc -b`.

**Regression:** Reverting this task returns `breakdown` to being silently discarded on six views and
restores a dead filter that hides whether server-side scoping actually works.

**Baseline:** before=286 Go / 98 SPA, after=288 Go / 102 SPA, measured as
`cd stats && go test ./... -count=1 -v | grep -c '^--- PASS'` and `cd stats/web && npm test`.
<!-- measured: both figures reported by this task's own run against commit `4b8cf92` on 2026-08-14 -->

**Commit:** `fix(25): reject a scoping parameter a view cannot honour`

---

### 26 The interface sends field names the server has never accepted

**Build:** green

**Found by running the application**, after the review panel closed on three clean slots — not by any
test. Recorded here because the fix is a task, not a patch.

**Files:**
- Modify: `stats/web/src/components/ChangeVariable.tsx`
- Modify: `stats/web/src/hooks/useRunDetail.ts`
- Add: `stats/web/src/testdata/queryFields.json`
- Add: `stats/internal/store/queryfields_test.go`
- Modify: `stats/web/src/api.ts`
- Modify: `stats/web/src/api.test.ts`

**Interfaces:**
- Consumes: `store.AllowedChangeFields()` and `store.AllowedStageRunFields()`, which already exist and
  were built for exactly this — a caller that needs the allowlist rather than a copy of it.
- Produces: the SPA's sort/filter field names, derived from the server's allowlist instead of guessed.

**The defect.** `ChangeVariable.tsx:38` sorts by `updatedAt` and `useRunDetail.ts:121` sorts by
`startedAt` — DTO field names. The server's allowlist takes column names: `updated_at`, `started_at`.
Both requests return **400** against the running daemon:

```text verified:both reproduced with curl against myflowd on 127.0.0.1:4173 on 2026-08-14; the snake_case forms return 200
store: unknown query field: sort "updatedAt" not recognised; accepted: branch, jira_issue, name, …, updated_at, updated_by
store: unknown query field: sort "startedAt" not recognised; accepted: attempt, branch, command, …, started_at, …
```

<!-- verified: both reproduced with curl against myflowd on 127.0.0.1:4173 on 2026-08-14; the snake_case forms return 200 -->

So the **change template variable renders empty** and the **run detail dashboard cannot load its
stage runs at all** — the feature this whole fix round was opened to build.

**Why nothing caught it.** The SPA's tests mock `fetchStatsView`/`listChanges` at the module boundary,
so no test has ever sent one of these strings to a server. The Go tests exercise the allowlist
thoroughly and never see what the interface sends. **Two vocabularies, each tested against itself** —
the same shape as the harvester/SQL mismatch (task 24), the SPA reader mismatch (panel F2), and the
unwired `Price` (task 23). Fourth instance.

- [x] **Step 1: Correct the two call sites**

`updatedAt` → `updated_at`, `startedAt` → `started_at`. Verify against the running daemon, not against
a mock.

- [x] **Step 2: Stop it being a private copy**

`internal/store` already exports `AllowedChangeFields()` and `AllowedStageRunFields()`. A new Go test
writes them to `stats/web/src/testdata/queryFields.json` and fails when the committed file disagrees
— the same generated-fixture-plus-drift-test pattern the wire shape now uses, which is already proven
to catch a rename.

`api.ts` exposes the imported names as the typed vocabulary for `SortKey.field` and `ListQuery.filters`,
and a SPA test asserts every field name the interface actually sends appears in that list. **A test
that asserts a literal string equals itself is not this test** — it must check membership of the
server-derived set, or it is the same private copy again with extra steps.

**Tests:** `queryfields_test.go` for the drift guard; `api.test.ts` for the membership assertion.
Verification is `go test ./... -race -count=1`, `npm test`, `npx tsc -b`, **and a live check against
the daemon that both requests return 200 and the run detail renders its rows**.

**Regression:** Reverting this returns the run-detail dashboard to loading no stage runs and the
change variable to rendering empty, with a green test suite in both cases.

**Baseline:** before=295 Go / 102 SPA, after=295+ / 102+, measured as
`cd stats && go test ./... -count=1 -v | grep -c '^--- PASS'` and `cd stats/web && npm test`.
<!-- measured: 295 Go and 106 SPA against commit `d5f4fc7` on 2026-08-14 -->

**Commit:** `fix(26): send the field names the server's allowlist actually accepts`

---

### 27 The dashboard renders light, and every stat panel prints its label twice

**Build:** green

**Found by opening the page**, after task 26. Neither defect is visible to any test — one is a
default-theme choice and the other is duplicated text that renders correctly in isolation.

**Files:**
- Modify: `stats/web/src/theme.css`
- Modify: `stats/web/src/components/StatPanel.tsx`
- Modify: `stats/web/src/components/Panel.tsx`
- Modify: `stats/web/src/components/Panel.test.tsx`
- Modify: `stats/web/src/views/views.test.tsx`
- Modify: `stats/web/src/styles.css`

**Interfaces:**
- Consumes: task 19's panel primitives and token palette.
- Produces: a dark-by-default dashboard, and one heading per panel.

**Defect 1 — it is not dark.** `theme.css`'s `:root` declares `color-scheme: light` and a white
surface palette; the dark tokens apply only under `@media (prefers-color-scheme: dark)`. On a
light-mode machine the dashboard renders as a white page with plain headings. **Grafana's dashboard
model is dark-first, and that is most of what the operator asked for** — the proposal's own wording
is "a dark ground with slightly-lighter bordered panels". A theme that inverts with the viewer's OS
setting is a reasonable default for a general-purpose app and is the wrong default here.

Make the dark palette the base `:root`, and keep a light variant under
`@media (prefers-color-scheme: light)` and `:root[data-theme="light"]` so a light-mode viewer is
still served rather than forced. The token names do not change; only which values are the default.

**Defect 2 — every stat panel prints its label twice.** `StatPanel` renders its own
`stat-panel-label`, and the `Panel` wrapping it renders the same string as the panel heading, so the
live board shows `CHANGES / CHANGES / 10`. Each renders correctly on its own, which is why no test
caught it — the duplication only exists in the composition.

Decide which owns the heading and remove the other. The panel heading is the better owner: every
other panel type gets its title that way, so a stat panel that carried its own would be the one
inconsistent case. Add a test asserting a stat panel inside a panel renders its label **once**.

- [x] **Step 1: Dark by default**
- [x] **Step 2: One heading per panel**
- [x] **Step 3: Look at it**

Rebuild the SPA, reload `http://127.0.0.1:4173`, and confirm both by eye against a screenshot before
claiming the task done. **Neither defect is detectable from the test suite**, so a green suite is not
evidence here.

**Tests:** `Panel.test.tsx` gains the single-heading assertion; `views.test.tsx` keeps its existing
panel assertions. Verification is `npm test`, `npx tsc -b`, and a screenshot.

**Regression:** Reverting returns a white dashboard with doubled panel labels.

**Baseline:** before=296 Go / 110 SPA, after=296 / 110+.
<!-- measured: 296 Go and 110 SPA against commit `10129aa` on 2026-08-14 -->

**Commit:** `fix(27): make the dashboard dark by default and stop stat panels doubling their label`
