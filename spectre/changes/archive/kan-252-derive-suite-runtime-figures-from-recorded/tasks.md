# kan-252-derive-suite-runtime-figures-from-recorded

Implementation plan — suite runtimes recorded in the store, cited by `.flow/project.md`.

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** Measured suite runtimes live as recorded rows in the myflow store (per run, per
suite, per machine), and `.flow/project.md`'s `## test` prose cites `flow suite list`
instead of pasting a number.

**Architecture:** A `suite_runs` table (migration 0020) behind store accessors, HTTP routes
beside the verdicts/incidents pair, and a `flow suite record`/`flow suite list` CLI in the
`flow hazard` pattern — one end-to-end precedent (hazards) followed at every layer. The CLI
wrapper times any command and passes stdio and the exit code through untouched; the reader
summarises per (suite, host) with the median of the last 10 passing runs.

**Tech Stack:** Go (pgx, net/http, flag, os/exec), PostgreSQL migration. No new dependencies.

**Spec:** `spectre/changes/kan-252-derive-suite-runtime-figures-from-recorded/design.md`
(canonical); `docs/superpowers/specs/2026-09-09-kan-252-derive-suite-runtime-figures-from-recorded-design.md`
(the approved design record).

## Global Constraints

- Migration files are tracked by filename with no checksum — write a NEW file
  `0020_suite_runs.sql`; never edit an applied migration.
- No new Go dependencies. `gofmt -w` before `gofmt -l` claims clean; `go vet ./...` clean.
- An unreachable store never alters a wrapped run's outcome: `flow suite record` journals
  nothing and records nothing, prints one warning line, and exits with the child's own exit
  code — never a gate. `flow suite list` is a read: it prints the error and exits 1.
- The dev daemon on `127.0.0.1:4173` and the `myflow-postgres` container are protected
  (`.flow/project.md` `## stop`): an agent never stops or restarts them. Writing suite-run
  rows through the dev daemon is the feature working, not a violation.
- Never `git add` `<project>/spectre/changes/` or `<project>/docs/superpowers/` in a task
  commit; never `--no-verify`; never push or merge.
- Seeding (task 5) runs the binary built from THIS worktree (`stats/bin/flow`), never the
  globally installed `flow`, which predates the `suite` subcommand.

---

- [x] 1. suite_runs table and store accessors

**Build:** green
**Files:**
- Create: `stats/internal/store/migrations/0020_suite_runs.sql`
- Modify: `stats/internal/records/types.go`
- Create: `stats/internal/store/suiteruns.go`
- Create: `stats/internal/store/suiteruns_test.go`
**Tests:** `TestInsertSuiteRunReturnsRecordedRow`, `TestListSuiteRunsFiltersBySuiteAndLimit`, `TestSuiteRunsRejectUnknownProject` — run with the selector in Step 2
**Regression:** reverting this task's commit drops the only coverage of the suite-run accessors; the table no longer migrates, so every later task's store-touching tests fail.
**Baseline:** before=0 after=4
<!-- predicted: go test ./internal/store/ -run 'TestInsertSuiteRun|TestListSuiteRunsFiltersBySuiteAndLimit|TestSuiteRunsRejectUnknownProject' after task 1 -->
**Commit:** feat(store): suite runtimes table and accessors

  - [x] **Step 1: Write the failing store tests**

    Follow `stats/internal/store/hazards_test.go`'s use of the per-test database harness
    (`newTestDatabase` from `testsupport_test.go`; tests needing Postgres skip with a clear
    message when the compose stack is down). Three tests:

    - `TestInsertSuiteRunReturnsRecordedRow` — insert against a seeded project row; the
      returned `records.SuiteRun` carries the suite, host, duration, exit code and a
      non-zero `ran_at`; a second insert is a second row (no upsert).
    - `TestListSuiteRunsFiltersBySuiteAndLimit` — seed three suites across two hosts; with
      `suite` set, only that suite's rows come back, newest first; with `suite` empty, all
      suites come back; `limit` caps the row count.
    - `TestSuiteRunsRejectUnknownProject` — an insert naming a project_key with no `projects`
      row fails with a foreign-key error the store surfaces, not a silent row.

    The wire type lands in `stats/internal/records/types.go` beside `Verdict`:

```go verified:authored in-tree for this change
// SuiteRun is one timed suite execution on one machine, recorded by
// `flow suite record`. ExitCode is the child's own exit status, so a
// failed run's duration can be stored without ever presenting itself
// as a runtime figure.
type SuiteRun struct {
ID         int64     `json:"id"`
Suite      string    `json:"suite"`
Host       string    `json:"host"`
DurationMs int64     `json:"durationMs"`
ExitCode   int       `json:"exitCode"`
RanAt      time.Time `json:"ranAt"`
}
```

  - [x] **Step 2: Run the tests to verify they fail**

    Run: `cd stats && go test ./internal/store/ -run 'TestInsertSuiteRunReturnsRecordedRow|TestListSuiteRunsFiltersBySuiteAndLimit|TestSuiteRunsRejectUnknownProject' -v`
    Expected: FAIL — `InsertSuiteRun` / `ListSuiteRuns` undefined (compile error).

  - [x] **Step 3: Write the migration and the accessors**

    `stats/internal/store/migrations/0020_suite_runs.sql`, new file, header citing 0018's
    reasoning for project-scoping (suites run outside changes too):

```sql verified:authored in-tree for this change
CREATE TABLE suite_runs (
id          BIGSERIAL PRIMARY KEY,
project_key TEXT NOT NULL REFERENCES projects(project_key),
suite       TEXT NOT NULL,
host        TEXT NOT NULL,
duration_ms BIGINT NOT NULL CHECK (duration_ms >= 0),
exit_code   INT NOT NULL,
ran_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX suite_runs_project_suite ON suite_runs (project_key, suite, ran_at);
```

    `stats/internal/store/suiteruns.go` beside `hazards.go`:

```go verified:authored in-tree for this change
// InsertSuiteRun records one timed suite execution and returns the row
// as stored, timestamps included.
func (s *Store) InsertSuiteRun(ctx context.Context, projectKey string, run records.SuiteRun) (records.SuiteRun, error)

// ListSuiteRuns returns a project's suite runs newest first. An empty
// suite lists every suite; limit caps the row count and must be > 0.
func (s *Store) ListSuiteRuns(ctx context.Context, projectKey, suite string, limit int) ([]records.SuiteRun, error)
```

    Both validate inputs before contacting Postgres the way `hazard.go` validates its
    closed vocabulary: `suite` and `host` non-empty, `duration_ms >= 0`, `limit > 0`.

  - [x] **Step 4: Run the tests to verify they pass, then lint**

    Run: `cd stats && go test ./internal/store/ -run 'TestInsertSuiteRunReturnsRecordedRow|TestListSuiteRunsFiltersBySuiteAndLimit|TestSuiteRunsRejectUnknownProject' -v`
    Expected: PASS, 3 tests (skip with a clear message only if the compose stack is down).
    <!-- predicted: the selector names exactly this task's three new store tests -->
    Run: `cd stats && gofmt -w . && gofmt -l . && go vet ./...`
    Expected: `gofmt -l` prints nothing; `go vet` exits 0.

  - [x] **Step 5: Commit**

```bash verified:authored in-tree for this change
git add stats/internal/store/migrations/0020_suite_runs.sql stats/internal/records/types.go stats/internal/store/suiteruns.go stats/internal/store/suiteruns_test.go
git commit -m "feat(store): suite runtimes table and accessors"
```

- [x] 2. Suite-run client methods

**Build:** green
**Files:**
- Create: `stats/internal/client/suites.go`
- Create: `stats/internal/client/suites_test.go`
**Tests:** `TestRecordSuiteRunPostsAndReadsTheRecordedRow`, `TestListSuiteRunsReadsTheArray`, `TestSuiteRunCallsFallBackWhenNothingTrustworthyAnswers` — run with the selector in Step 2
**Regression:** reverting this task's commit drops the only coverage of the suite-run HTTP client, including its unreachable-store fallback; the CLI task would then be the first thing to exercise these routes.
**Baseline:** before=0 after=3
<!-- predicted: go test ./internal/client/ -run 'TestRecordSuiteRunPostsAndReadsTheRecordedRow|TestListSuiteRunsReadsTheArray|TestSuiteRunCallsFallBackWhenNothingTrustworthyAnswers' after task 2 -->
**Commit:** feat(client): suite-run record and list methods

  - [x] **Step 1: Write the failing client tests**

    Follow `stats/internal/client/records_test.go`'s `httptest.Server` pattern, including
    `TestRecordCallsFallBackWhenNothingTrustworthyAnswers`'s fallback assertions. Three
    tests:

    - `TestRecordSuiteRunPostsAndReadsTheRecordedRow` — asserts the request is
      `POST /api/v1/suites/proj/runs` with the JSON body, and the decoded returned row
      equals the server's response.
    - `TestListSuiteRunsReadsTheArray` — asserts `GET /api/v1/suites/proj/runs` carries
      `suite` and `limit` as query parameters when set, and decodes the JSON array.
    - `TestSuiteRunCallsFallBackWhenNothingTrustworthyAnswers` — a server without the
      `Flow-Daemon` trust header, and a dead URL, produce the same wrapped error shape the
      record calls produce for verdicts; nothing panics, nothing retries in a loop.

  - [x] **Step 2: Run the tests to verify they fail**

    Run: `cd stats && go test ./internal/client/ -run 'TestRecordSuiteRunPostsAndReadsTheRecordedRow|TestListSuiteRunsReadsTheArray|TestSuiteRunCallsFallBackWhenNothingTrustworthyAnswers' -v`
    Expected: FAIL — `RecordSuiteRun` / `ListSuiteRuns` undefined.

  - [x] **Step 3: Write the client methods**
    `stats/internal/client/suites.go` beside `records.go`, same trust-header and fallback
    plumbing the record calls use:

```go verified:authored in-tree for this change
// RecordSuiteRun posts one timed suite execution and returns the row
// as the store recorded it.
func (c *Client) RecordSuiteRun(ctx context.Context, project string, run records.SuiteRun) (records.SuiteRun, error)

// ListSuiteRuns reads a project's recorded suite runs, newest first.
// Empty suite lists every suite; limit must be > 0.
func (c *Client) ListSuiteRuns(ctx context.Context, project, suite string, limit int) ([]records.SuiteRun, error)
```

  - [x] **Step 4: Run the tests to verify they pass, then lint**

    Run: `cd stats && go test ./internal/client/ -run 'TestRecordSuiteRunPostsAndReadsTheRecordedRow|TestListSuiteRunsReadsTheArray|TestSuiteRunCallsFallBackWhenNothingTrustworthyAnswers' -v`
    Expected: PASS, 3 tests.
    <!-- predicted: the selector names exactly this task's three new client tests -->
    Run: `cd stats && gofmt -w . && gofmt -l . && go vet ./...`
    Expected: clean.

  - [x] **Step 5: Commit**

```bash verified:authored in-tree for this change
git add stats/internal/client/suites.go stats/internal/client/suites_test.go
git commit -m "feat(client): suite-run record and list methods"
```

- [x] 3. Suite-run API routes

**Build:** green
**Files:**
- Modify: `stats/internal/api/server.go`
- Modify: `stats/internal/api/records.go`
- Modify: `stats/internal/api/changes_test.go`
- Modify: `stats/internal/reconcile/record_test.go`
- Modify: `stats/internal/web/embed_test.go`
- Modify: `stats/internal/client/client_test.go`
- Create: `stats/internal/api/suites.go`
- Create: `stats/internal/api/suites_test.go`
**Tests:** `TestPostSuiteRunReturnsTheRecordedRow`, `TestGetSuiteRunsReadsTheArray`, `TestPostSuiteRunRejectsAnEmptySuite` — run with the selector in Step 2
**Regression:** reverting this task's commit drops the only coverage of the suite-run routes and unregisters them, so the CLI's record and list both lose their server side.
**Baseline:** before=0 after=4
<!-- predicted: go test ./internal/api/ -run 'TestPostSuiteRunReturnsTheRecordedRow|TestGetSuiteRuns|TestPostSuiteRunRejectsAnEmptySuite' after task 3 -->
**Commit:** feat(api): suite-run record and list routes

  - [x] **Step 1: Write the failing handler tests**

    Follow `stats/internal/api/records_test.go`'s harness pattern. Three tests:

    - `TestPostSuiteRunReturnsTheRecordedRow` — `POST /api/v1/suites/proj/runs` against a
      seeded project returns the stored row as JSON.
    - `TestGetSuiteRunsReadsTheArray` — `GET /api/v1/suites/proj/runs?suite=s&limit=2`
      returns the filtered, capped array; absent query parameters use the defaults
      (every suite, limit 20).
    - `TestPostSuiteRunRejectsAnEmptySuite` — a body with an empty `suite` (and one with a
      negative `durationMs`) is a caller mistake: the same 4xx shape the record handlers
      use for a bad payload, with nothing written.

  - [x] **Step 2: Run the tests to verify they fail**

    Run: `cd stats && go test ./internal/api/ -run 'TestPostSuiteRunReturnsTheRecordedRow|TestGetSuiteRunsReadsTheArray|TestPostSuiteRunRejectsAnEmptySuite' -v`
    Expected: FAIL — routes unregistered / handler types undefined.

  - [x] **Step 3: Write the handlers and register the routes**

    `stats/internal/api/suites.go` beside `records.go`'s verdict handlers; registered in
    `server.go` beside the verdicts/incidents lines:

```go verified:authored in-tree for this change
mux.HandleFunc("POST /api/v1/suites/{project}/runs", sh.recordSuiteRun)
mux.HandleFunc("GET /api/v1/suites/{project}/runs", sh.listSuiteRuns)
```

    Handlers validate the payload before the store is contacted (suite and host non-empty,
    duration >= 0), resolve `{project}` the way the incidents handler does, and encode the
    store's rows verbatim — no second representation of the row.

  - [x] **Step 4: Run the tests to verify they pass, then lint**

    Run: `cd stats && go test ./internal/api/ -run 'TestPostSuiteRunReturnsTheRecordedRow|TestGetSuiteRunsReadsTheArray|TestPostSuiteRunRejectsAnEmptySuite' -v`
    Expected: PASS, 3 tests.
    <!-- predicted: the selector names exactly this task's three new api tests -->
    Run: `cd stats && gofmt -w . && gofmt -l . && go vet ./...`
    Expected: clean.

  - [x] **Step 5: Commit**

```bash verified:authored in-tree for this change
git add stats/internal/api/server.go stats/internal/api/suites.go stats/internal/api/suites_test.go
git commit -m "feat(api): suite-run record and list routes"
```

- [x] 4. flow suite record and list

**Build:** green
**Files:**
- Modify: `stats/cmd/flow/main.go`
- Create: `stats/cmd/flow/suite.go`
- Create: `stats/cmd/flow/suite_test.go`
**Tests:** `TestSuiteRecordRunsChildAndRecordsRow`, `TestSuiteRecordPassesChildExitThrough`, `TestSuiteRecordWarnsWithoutAlteringExitWhenStoreDown`, `TestSuiteRecordRefusesMissingSuiteOrCommand`, `TestSuiteListRendersRowsAndSummary`, `TestSuiteListJSON` — run with the selector in Step 2
**Regression:** reverting this task's commit removes the only user-facing surface of the feature — nothing can record or read a suite runtime — and drops the wrapper's exit-code and fallback coverage.
**Baseline:** before=0 after=8
<!-- predicted: go test ./cmd/flow/ -run 'TestSuiteRecord|TestSuiteList' after task 4 -->
**Commit:** feat(stats): flow suite record and list

  - [x] **Step 1: Write the failing CLI tests**

    Follow `stats/cmd/flow/hazard_test.go`'s pattern (route `run` through a test server
    URL; assert on the written stdout/stderr buffers and the returned exit code). Six
    tests:

    - `TestSuiteRecordRunsChildAndRecordsRow` — `suite record -suite s -- <child>` runs the
      child with inherited stdio, records one row (suite, this host, a positive duration,
      the child's exit code), and exits with the child's code.
    - `TestSuiteRecordPassesChildExitThrough` — a child exiting 3 makes record exit 3.
    - `TestSuiteRecordWarnsWithoutAlteringExitWhenStoreDown` — the store URL is dead: one
      warning line on stderr, the child still runs, the exit code is still the child's.
    - `TestSuiteRecordRefusesMissingSuiteOrCommand` — no `-suite`, or no `--` and command
      words: usage on stderr, exit 2, nothing executed, nothing recorded.
    - `TestSuiteListRendersRowsAndSummary` — seeded rows across two hosts render newest
      first with a host column, plus one summary line per (suite, host) carrying the median
      of the last 10 passing runs (even count: the lower middle of the sorted slice —
      `sorted[(n-1)/2]`, no float); failed runs are excluded from the median but still
      rendered as rows.
    - `TestSuiteListJSON` — `-json` emits the raw array and no summary lines.

  - [x] **Step 2: Run the tests to verify they fail**

    Run: `cd stats && go test ./cmd/flow/ -run 'TestSuiteRecord|TestSuiteList' -v`
    Expected: FAIL — `runSuite` undefined.

  - [x] **Step 3: Write the subcommand and register it**

    `stats/cmd/flow/suite.go` beside `hazard.go`; dispatch `case "suite":` in `main.go`'s
    switch, with a `suiteUsage` block in `hazardUsage`'s shape — flags first, then the
    failure-semantics paragraph. The wrapper's core:

```go verified:authored in-tree for this change
// runSuiteRecord executes the child with inherited stdio, wall-times
// it, and records the row. The store never gates the child: a record
// failure prints one warning line and changes no exit code. The only
// non-zero exits are caller mistakes (usage, 2) and a child that
// could not start (127, nothing recorded).
func runSuiteRecord(ctx context.Context, args []string, stdout, stderr io.Writer) int
```

    Project and address resolution follow `hazard.go` exactly (`-C dir`, `-addr`,
    `-timeout`, `FLOW_ADDR`); the host is `os.Hostname()`'s short name. `list` prints rows
    then summaries; `-json` prints only the array.

  - [x] **Step 4: Run the tests to verify they pass, then lint**

    Run: `cd stats && go test ./cmd/flow/ -run 'TestSuiteRecord|TestSuiteList' -v`
    Expected: PASS, 6 tests.
    <!-- predicted: the selector names exactly this task's six new cli tests -->
    Run: `cd stats && gofmt -w . && gofmt -l . && go vet ./...`
    Expected: clean.

  - [x] **Step 5: Commit**

```bash verified:authored in-tree for this change
git add stats/cmd/flow/main.go stats/cmd/flow/suite.go stats/cmd/flow/suite_test.go
git commit -m "feat(stats): flow suite record and list"
```

- [x] 5. project.md cites the app; seed real recordings

**Build:** green
**Files:**
- Modify: `.flow/project.md`
**Tests:** none
**Regression:** reverting this task's commit restores the pasted "52 to 56s" prose and its provenance comments — the stale-figures failure mode KAN-252 exists to remove.
**Baseline:** before=0 after=0
<!-- predicted: no tests in this task — markdown prose plus live recordings through the wrapper -->
**Commit:** docs: cite recorded suite runtimes from the test section

  - [x] **Step 1: Build this worktree's flow binary and record the three suites**

```bash verified:commands read from .flow/project.md ## test section
cd stats && go build -o bin/flow ./cmd/flow
./bin/flow suite record -suite guard-tests -C . -- scripts/run-guard-tests.sh
cd stats && ../bin/flow suite record -suite stats-go -C .. -- go test ./... -race -count=1
cd stats/web && ../bin/flow suite record -suite stats-spa -C ../.. -- npm test
```

    Each command passes its output and exit code through untouched; each suite runs in one
    invocation, well inside this harness's default tool timeout. On store failure the line
    is a warning and the exit code is the suite's own — the recording is not a gate.
    <!-- predicted: each of the three seeded suites completes in one tool call; confirm at this step -->
    <!-- predicted: stats/bin/flow suite list -C . after task 5, one summary line per suite -->

  - [x] **Step 2: Read the figures back and rewrite the `## test` prose**

    Run: `stats/bin/flow suite list -C .`
    Replace `.flow/project.md`'s "Measured runtime … 52 to 56s wall" paragraph and the two
    `<!-- measured: … -->` comments after it with prose in this shape (figures read, not
    pasted — quote none of them):

> **Measured runtime: recorded per run in the myflow store — query it with
> `flow suite list`.** The canonical suites are `guard-tests`
> (`scripts/run-guard-tests.sh`), `stats-go` (`go test ./... -race -count=1`) and
> `stats-spa` (`stats/web` `npm test`); each row carries the machine that ran it, and
> the per-(suite, host) summary line is the median of the last 10 passing runs.
> Whether the whole `## test` list still fits inside this harness's default tool
> timeout in one invocation is answered by those recorded figures, not by a number
> pasted here — which is the point: a written duration went stale twice
> before this paragraph stopped carrying one.

    Keep the runner description, the TMPDIR incident note, and the 0.84 s
    `check-installed-citations` note byte-for-byte; the figure those paragraphs carry is
    out of scope (design.md, cite-app-only).

  - [x] **Step 3: Lint the markdown and the budget**

    Run: `scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-contract-budget.sh && scripts/check-markdown-integrity.py`
    Expected: all exit 0 — `.flow/project.md` shrinks, so its budget entry passes.

  - [x] **Step 4: Commit**

```bash verified:authored in-tree for this change
git add .flow/project.md
git commit -m "docs: cite recorded suite runtimes from the test section"
```
