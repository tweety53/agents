# kan-451-add-a-per-project-incident-table-and-a-guard

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** two store tables — a guard verdict log with a false-positive flag, and a per-project
incident table — written through `flow record`, read back by `check-unfinished-work.sh` on
`OUTSTANDING` and by every dispatch bundle `gather-dispatch-context.sh` builds.

**Architecture:** one migration; the wire types, store methods, API routes, client methods and CLI
verbs follow the `dispatches`/`findings` layering already in place (`records.Dispatch` →
`store.RecordDispatch` → `api.recordHandler` → `client.RecordDispatch` → `flow record dispatch`).
Writes take the never-block journal; reads fail loudly. Two scripts and one skill file consume the
verbs.

**Tech Stack:** Go (`stats/`), PostgreSQL migrations, bash guard scripts with fixture harnesses,
Markdown skill text.

**Spec:** `spectre/changes/kan-451-add-a-per-project-incident-table-and-a-guard/design.md`

## Global Constraints

- `check-unfinished-work.sh`'s verdict line, its exit codes and their meanings are unchanged; the
  two `flow` calls it gains are advisory (design decision `script-reads-are-advisory`).
- `gather-dispatch-context.sh` builds with no `flow` on `PATH` and with the store down; the
  `## incidents` section is `skipped`, never `refused`.
- `0010_run_records.sql` and every applied migration are untouched; the new tables are a new file
  (the reason `0011_dispatch_agent_id.sql`'s header gives). `EmbeddedMigrationCount` derives the
  count, so no literal is bumped.
- `skills/flow/integrate.md` stays under its 18602-byte budget (15442 bytes today) and
  `skills/flow-contracts/finish-contract-run1.md` under 30090 (24433 today) in
  `scripts/check-contract-budget.sh`; new prose carries no `SHALL`/`MUST` sentence.
  <!-- measured: wc -c skills/flow/integrate.md skills/flow-contracts/finish-contract-run1.md; grep -n 'integrate.md\|finish-contract-run1.md' scripts/check-contract-budget.sh @ main 16a7c33 -->
- Every guard listed under `## lint` in `.flow/project.md` and `cd stats && go test ./... -race
  -count=1` exit clean before a task is reported done. `stats/internal/store` tests need the
  `flow-postgres` stack on port 5433 (`stats/README.md` "Running the tests"); they create and drop
  per-test databases and never touch the default `flow` database.

---

- [x] 1. Migration, wire types and store methods

**Build:** green
**Files:** `stats/internal/store/migrations/0018_guard_log.sql`, `stats/internal/records/types.go`,
`stats/internal/store/records.go`, `stats/internal/store/records_test.go`
**Tests:** `TestGuardLogMigrationAppliesTwiceIdempotently`, `TestRecordVerdictRoundTrips`,
`TestFlagVerdictFalsePositiveFlagsTheLatestForChangeAndGuard`,
`TestFlagVerdictFalsePositiveReportsNoVerdict`, `TestListVerdictsFiltersByGuardAndFlag`,
`TestRecordIncidentWithAndWithoutAChange`, `TestListIncidentsNewestFirst`
**Regression:** revert this commit and the seven tests do not compile — `records.Verdict`,
`records.Incident` and the five store methods are gone — and `store.Migrate` applies the
embedded set without `0018`, which the migration-count test no longer accepts.
**Baseline:** before=162 after=169
<!-- measured: cd stats && go test -count=1 -v ./internal/store | grep -c '^--- PASS' @ main 16a7c33 -->
<!-- predicted: the same command after task 1 — seven new top-level tests, 162 + 7 -->
**Commit:** feat(store): add guard verdict and incident tables with record and list methods

  - [x] **Step 1: Write the migration**

    Create `stats/internal/store/migrations/0018_guard_log.sql` with the two `CREATE TABLE`
    statements and two indexes from `design.md` §1, preceded by a header comment in the style of
    `0011_dispatch_agent_id.sql`: why `guard_verdicts` derives its project through `changes` while
    `incidents` stores `project_key` (nullable `change_id`), why `verdict` is the verbatim line,
    and why the file is new rather than an edit to `0010`. `//go:embed migrations/*.sql`
    (`stats/internal/store/migrations.go` line 16) picks it up with no code change.
    <!-- measured: grep -n 'go:embed' stats/internal/store/migrations.go @ main 16a7c33 -->

  - [x] **Step 2: Add the wire types**

    In `stats/internal/records/types.go`, after `Finding` and before `Run`:

```go unverified:field tags must match the JSON the CLI and jq in tasks 4–5 read (falsePositiveReason, flaggedAt, change)
type Verdict struct {
    ID                  int64      `json:"id"`
    Change              string     `json:"change,omitempty"`
    Guard               string     `json:"guard"`
    Worktree            string     `json:"worktree"`
    Verdict             string     `json:"verdict"`
    RecordedAt          time.Time  `json:"recordedAt"`
    FalsePositive       bool       `json:"falsePositive"`
    FalsePositiveReason string     `json:"falsePositiveReason,omitempty"`
    FlaggedAt           *time.Time `json:"flaggedAt,omitempty"`
}

type VerdictFlag struct {
    Guard  string `json:"guard"`
    Reason string `json:"reason"`
}

type Incident struct {
    ID          int64     `json:"id"`
    Change      string    `json:"change,omitempty"`
    Guard       string    `json:"guard"`
    Symptom     string    `json:"symptom"`
    Recovery    string    `json:"recovery"`
    MinutesLost int       `json:"minutesLost"`
    OccurredAt  time.Time `json:"occurredAt"`
}
```

    Doc comments in the style of `Dispatch`'s: what `Change` means on each side, why `Verdict` is
    the whole line, why `FlaggedAt` is a pointer.

  - [x] **Step 3: Write the failing store tests**

    In `stats/internal/store/records_test.go`, after `TestMarkDispatchesUnattributedByIDRepeatStampIsIdempotent`
    (line 1721), using the same `newTestDatabase`/`PutChange` fixture the file's other tests use:
    - `TestGuardLogMigrationAppliesTwiceIdempotently` — in the shape of
      `TestRunRecordsMigrationAppliesTwiceIdempotently` (line 94), asserting both tables exist
      after a second `Migrate`.
    - `TestRecordVerdictRoundTrips` — `RecordVerdict` returns an `ID`, and `ListVerdicts(project,
      "", false)` returns the row with `Change` set to the change name and `FalsePositive` false.
    - `TestFlagVerdictFalsePositiveFlagsTheLatestForChangeAndGuard` — two verdicts for the same
      change and guard, one for another guard; `FlagVerdictFalsePositive` flags only the newest of
      the pair (by `recorded_at` then `id`), returns it with `FlaggedAt` set, and re-flagging
      overwrites the reason.
    - `TestFlagVerdictFalsePositiveReportsNoVerdict` — a change with no verdict for that guard
      returns an error that `errors.Is(err, store.ErrNotFound)` (or whichever sentinel
      `EndDispatch` returns for an unknown key — reuse it, do not add one).
    - `TestListVerdictsFiltersByGuardAndFlag` — `guard` filter, `falsePositiveOnly` filter, newest
      first.
    - `TestRecordIncidentWithAndWithoutAChange` — `Change` empty stores NULL and reads back empty;
      `Change` naming a known change reads back that name; an unknown change name is the same
      not-found error `RecordDispatch` returns; `OccurredAt` zero takes the column default.
    - `TestListIncidentsNewestFirst` — two rows, project filter excludes another project's row.
    <!-- measured: grep -n 'func Test' stats/internal/store/records_test.go @ main 16a7c33 -->

  - [x] **Step 4: Run the tests and see them fail**

    `cd stats && go test -count=1 -run 'GuardLog|Verdict|Incident' ./internal/store` — compile
    failure on the missing methods.

  - [x] **Step 5: Implement the store methods**

    In `stats/internal/store/records.go`, after `RunRecord`'s helpers (`readFindings`, line 712),
    the five methods of `design.md` §3. Resolve `change_id` with the same `SELECT id FROM changes
    WHERE project_key=$1 AND name=$2` `insertDispatch` uses (line 220) and map zero rows to the
    not-found sentinel it maps to. `FlagVerdictFalsePositive` is one `UPDATE … WHERE id = (SELECT
    … ORDER BY recorded_at DESC, id DESC LIMIT 1) RETURNING …`; `RowsAffected()==0` → not found.
    Lists `JOIN changes` for the name and `ORDER BY recorded_at DESC, id DESC` /
    `occurred_at DESC, id DESC`.
    <!-- measured: grep -n '^func ' stats/internal/store/records.go @ main 16a7c33 -->

  - [x] **Step 6: Run the store tests green, then gofmt and vet**

    `cd stats && gofmt -w . && go vet ./... && go test -count=1 -race ./internal/store`.

  - [x] **Step 7: Commit**

    `feat(store): add guard verdict and incident tables with record and list methods`

- [x] 2. API routes and client methods

**Build:** green
**Files:** `stats/internal/api/records.go`, `stats/internal/api/server.go`,
`stats/internal/api/records_test.go`, `stats/internal/client/client.go`,
`stats/internal/client/records_test.go`
**Allowed-collateral:** call-site updates to `api.RecordStore`'s five new methods --
`stats/internal/api/changes_test.go`, `stats/internal/client/client_test.go`,
`stats/internal/web/embed_test.go` (each holds a fake/stub implementing `api.RecordStore` for an
unrelated test file's own server; every one needs the five new methods stubbed to keep compiling,
same shape as kan-326's task 2 `Allowed-collateral:` for `api.New`'s new store parameter).
**Tests:** `TestRecordVerdictRouteAnswers201`, `TestFlagVerdictRouteAnswers200And404`,
`TestListVerdictsRouteFiltersByQuery`, `TestRecordIncidentRouteAnswers201AndValidatesMinutes`,
`TestListIncidentsRouteReturnsNewestFirst`, `TestRecordVerdictReturnsTheRecordedRow`,
`TestFlagVerdictDistinguishesNoVerdictFromAnUnreachableStore`,
`TestListVerdictsAndIncidentsReadArrays`
**Regression:** revert this commit and the five routes answer 404 from the API-prefix fallback,
the client methods do not exist, and every test above fails to compile or asserts a 404.
**Baseline:** before=124 after=132
<!-- measured: cd stats && for p in ./internal/api ./internal/client; do go test -count=1 -v "$p" | grep -c '^--- PASS'; done @ main 16a7c33 (76 + 48) -->
<!-- predicted: the same command after task 2 — five api tests and three client tests, 124 + 8 -->
**Commit:** feat(api): serve guard verdicts and incidents

  - [x] **Step 1: Write the failing API tests**

    In `stats/internal/api/records_test.go`, after `TestEndDispatchRouteClosesTheRowItsBeginOpened`
    (line 608), in the shape of `TestRecordDispatchRouteAllocatesSeqAndAnswers201` (line 344) —
    the file's in-memory `RecordStore` fake gains the five methods:
    - `POST /api/v1/records/{p}/{c}/verdicts` → 201 with the row.
    - `POST …/verdicts/false-positive` → 200 with the flagged row; 404 when the store reports
      no verdict; 400 on an empty `reason`.
    - `GET /api/v1/verdicts/{p}?guard=x&falsePositive=true` → the store was called with
      `("x", true)`; no query → `("", false)`; `falsePositive=maybe` → 400.
    - `POST /api/v1/incidents/{p}` → 201; `minutesLost` negative or `guard`/`symptom`/`recovery`
      empty → 400 without reaching the store.
    - `GET /api/v1/incidents/{p}` → 200 with the array the store returned.
    <!-- measured: grep -n 'func Test' stats/internal/api/records_test.go @ main 16a7c33 -->

  - [x] **Step 2: Write the failing client tests**

    In `stats/internal/client/records_test.go`, after
    `TestEndDispatchReadsTheClosedRowAndDistinguishesAnUnknownKey` (line 227), in the shape of
    `TestSetFindingStatusDistinguishesAnUnknownRefFromAnUnreachableStore` (line 106): the
    three client tests named above, using the file's `httptest` daemon-header server.

  - [x] **Step 3: Run both packages and see the new tests fail to compile**

    `cd stats && go test -count=1 ./internal/api ./internal/client`.

  - [x] **Step 4: Implement the routes**

    `stats/internal/api/records.go`: add the five methods to `RecordStore` (line 39); handlers
    `recordVerdict`, `flagVerdict`, `listVerdicts`, `recordIncident`, `listIncidents` beside
    `runRecord` (line 301), decoding with the same `DisallowUnknownFields` decoder
    `recordFinding` uses and mapping store errors through `mapStoreError`. `server.go`: five
    `mux.HandleFunc` lines after line 307, paths per `design.md` §4. Validation in the handler,
    before the store: non-empty `guard`, `worktree`, `verdict` / `reason` / `guard`, `symptom`,
    `recovery`; `minutesLost >= 0`; `falsePositive` query is `true`, `false` or absent.
    <!-- measured: grep -n 'mux.HandleFunc' stats/internal/api/server.go @ main 16a7c33 -->

  - [x] **Step 5: Implement the client methods**

    `stats/internal/client/client.go`, after `GetCostStatus` (line 757): `RecordVerdict`,
    `FlagVerdictFalsePositive`, `ListVerdicts(ctx, project, guard string, falsePositiveOnly bool)`,
    `RecordIncident`, `ListIncidents`, through `writeRecord` for writes and the `GetRunRecord`
    read shape for lists; `FlagVerdictFalsePositive` classifies a daemon-headed 404 as the
    not-found error `SetFindingStatus` returns, so the CLI journals a network failure and
    refuses a genuine "no verdict".

  - [x] **Step 6: Run green, gofmt, vet, commit**

    `cd stats && gofmt -w . && go vet ./... && go test -count=1 -race ./internal/api ./internal/client`,
    then `feat(api): serve guard verdicts and incidents`.

- [x] 3. `flow record` verbs and journal replay

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`,
`stats/internal/reconcile/reconcile.go`, `stats/internal/reconcile/record_test.go`
**Allowed-collateral:** `stats/internal/api/records.go` -- adds `ApplyVerdictRecord`,
`ApplyVerdictFlag` and `ApplyIncidentRecord` beside `ApplyFindingStatus`, which design.md §5 step 5
requires so the live route and `internal/reconcile`'s replay of a journalled verdict/incident entry
share one validation path, exactly as the four existing `Apply*` functions do for the other kinds.
**Tests:** `TestRecordVerdictWritesAndFallsBackToJournal`,
`TestRecordVerdictFalsePositiveRefusesNoVerdictAndJournalsUnreachable`,
`TestRecordVerdictsPrintsArrayAndFailsLoudly`,
`TestRecordIncidentRejectsNegativeMinutesWithoutContactingStore`,
`TestRecordIncidentsPrintsArray`, `TestReplayAppliesVerdictAndIncidentEntries`
**Regression:** revert this commit and `flow record verdict` prints `unknown record command` at
exit 2, so task 4's guard records nothing and task 5's bundle skips incidents on every run; a
journalled `verdict` entry is refused by `applyRecordEntry` as an unknown kind.
**Baseline:** before=161 after=167
<!-- measured: cd stats && for p in ./cmd/flow ./internal/reconcile; do go test -count=1 -v "$p" | grep -c '^--- PASS'; done @ main 16a7c33 (140 + 21) -->
<!-- predicted: the same command after task 3 — five cmd/flow tests and one reconcile test, 161 + 6 -->
**Commit:** feat(flow): add record verdict, verdicts, incident and incidents verbs

  - [x] **Step 1: Write the failing CLI tests**

    In `stats/cmd/flow/record_test.go`, after `TestRecordFindingRejectsWithdrawnWithNoReason`
    (line 1555), in the shape of `TestRecordWritePrintsOneLineAndExitsZero` (line 52) and
    `TestRecordWriteFallsBackToJournalAndExitsZero` (line 249):
    - `verdict` with all four flags → POST body carries them and `recordedAt` non-zero; store
      down → `⚠ flow: store unreachable — wrote local journal`, exit 0, journal entry kind
      `verdict`.
    - `verdict false-positive` → a daemon 404 exits non-zero with the message and journals
      nothing; a dead port journals kind `verdict-false-positive` and exits 0; empty `-reason`
      exits 2 without contacting the store.
    - `verdicts` → prints the array verbatim, `[]` for no rows, stderr + exit 1 on a dead port;
      `-guard` and `-false-positive` land in the query string.
    - `incident` with `-minutes-lost -1` or `abc` → exit 2, no request; `-change` omitted → body
      omits `change`.
    - `incidents` → prints the array.
    <!-- measured: grep -n 'func Test' stats/cmd/flow/record_test.go @ main 16a7c33 -->

  - [x] **Step 2: Write the failing reconcile test**

    In `stats/internal/reconcile/record_test.go`, after
    `TestRecordStatusEntryWithAnEmptyRefStillReachesTheStore` (line 431), in the shape of
    `TestReplayAppliesPendingRecordEntries` (line 218): a journal holding one `verdict`, one
    `verdict-false-positive` and one `incident` entry is replayed with three applied, zero
    refused, and the fake store saw each call.

  - [x] **Step 3: See both fail**

    `cd stats && go test -count=1 -run 'Verdict|Incident' ./cmd/flow ./internal/reconcile`.

  - [x] **Step 4: Implement the verbs**

    `stats/cmd/flow/record.go`: five `case` arms in `runRecord`'s switch (after `"findings"`,
    line ~197); `runRecordVerdict` dispatches its own `false-positive` sub-verb the way
    `runRecordDispatchVerb` splits `begin`/`end`. Writes go through `callRecord` (line 340) with
    the journal kinds from `design.md` §5; `incident`/`incidents` register `-addr`, `-timeout`,
    `-C` without `-change` being required, resolving the project through `fallback.ProjectKey`
    as line 488 does. `-minutes-lost` parses with `strconv.Atoi` and refuses `< 0` before any
    request. Extend `recordUsage` with the five usage lines and one paragraph per verb pair, in
    the style of the `findings` paragraph.
    <!-- measured: grep -n 'case "findings"\|^func callRecord\|fallback.ProjectKey' stats/cmd/flow/record.go @ main 16a7c33 -->

  - [x] **Step 5: Implement replay**

    `stats/internal/reconcile/reconcile.go` `applyRecordEntry` (line ~676): three `case` arms
    after `"status"`, decoding the request into the wire type and calling the corresponding
    `api.Apply…` function — add `ApplyVerdictRecord`, `ApplyVerdictFlag`, `ApplyIncidentRecord`
    beside `ApplyFindingStatus` in `stats/internal/api/records.go` (line 126) so the handler and
    the replay share one validation path, as the existing four kinds do. Update the kind list in
    the comment at line 595.

  - [x] **Step 6: Run green, gofmt, vet, commit**

    `cd stats && gofmt -w . && go vet ./... && go test -count=1 -race ./cmd/flow ./internal/reconcile ./internal/api`,
    then `feat(flow): add record verdict, verdicts, incident and incidents verbs`.

- [x] 4. `check-unfinished-work.sh` records its verdict and prints prior false positives

**Build:** green
**Files:** `scripts/check-unfinished-work.sh`, `scripts/test-check-unfinished-work.sh`
**Allowed-collateral:** `stats/cmd/flow/record_test.go` -- the pre-existing
`TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard` shells out to this guard, which now
also POSTs its advisory verdict write to the same `FLOW_ADDR` the test's render-only mock daemon
answers on; that mock's shared `renderDaemon` helper hard-fails any non-GET request, so this one
test's own local handler now tolerates the write instead of reusing that helper.
**Tests:** Case 20: the `-verdict` argument the stub `flow` received equals the printed verdict
line on both CLEAR and OUTSTANDING; Case 21: OUTSTANDING with two prior false positives prints the
`prior false positives … 2 — last: <newest reason> (<change>, <date>)` stderr line after the
verdict; Case 22: OUTSTANDING with `[]` prints no such line; Case 23: `verdicts` failing prints
no such line and the verdict and exit 0 are unchanged; Case 24: `verdict` failing leaves the
verdict line and exit 0 unchanged; Case 25: CLEAR never calls `verdicts`
**Regression:** revert this commit and the harness's stub `flow` is never invoked with `verdict`
(case 20 finds no recorded argument), and no stderr line appears in case 21 — the recurrence the
ticket names is rediscovered by hand again.
**Baseline:** before=78 after=84
<!-- measured: bash scripts/test-check-unfinished-work.sh | grep -c '^ok:' @ main 16a7c33 -->
<!-- predicted: the same command after task 4 — six new cases, one ok line each, 78 + 6 -->
**Commit:** feat(scripts): record check-unfinished-work verdicts and surface prior false positives

  - [x] **Step 1: Extend the stub and write the failing cases**

    In `scripts/test-check-unfinished-work.sh`, extend `new_fixture`'s stub `flow` (line 212)
    to switch on `$2`: `findings` keeps its behaviour; `verdict` appends its arguments to
    `$WT/bin/verdict.args` and exits 0; `verdicts` cats `$WT/bin/verdicts.json` (default `[]`).
    Add helpers `set_prior_false_positives <json>`, `set_verdicts_unreachable` and
    `set_verdict_unreachable` beside `set_findings_unreachable` (line 193) — the last two make
    only that verb exit 1. Append cases 20–25 before the `if [ "$FAILURES" -ne 0 ]` tail, in the
    shape of the existing cases, using the `flaggedAt`/`falsePositiveReason`/`change` JSON keys
    task 1 fixed.
    <!-- measured: grep -n 'set_findings_unreachable\|new_fixture\|FAILURES" -ne 0' scripts/test-check-unfinished-work.sh @ main 16a7c33 -->

  - [x] **Step 2: Run the harness and see cases 20–25 fail**

    `bash scripts/test-check-unfinished-work.sh` — cases 1–19 pass, 20–25 fail.

  - [x] **Step 3: Add the two calls to the guard**

    Replace the `if [ -z "$REASONS" ] … fi; exit 0` tail (lines 335–340) with the block from
    `design.md` §6; add a header paragraph (after "SIGNAL TWO READS THE STORE") titled "THE
    VERDICT IS RECORDED, AND PRIOR FALSE POSITIVES ARE ADVISORY", stating that neither `flow`
    call can move the verdict line or the exit code, why the read runs only on OUTSTANDING, and
    that the write's own stdout and stderr are both discarded, so a store outage leaves no visible
    trace from that call.
    <!-- measured: sed -n 335,340p scripts/check-unfinished-work.sh @ main 16a7c33 -->

  - [x] **Step 4: Run the harness green and the lint list**

    `bash scripts/test-check-unfinished-work.sh`, then every `## lint` command in
    `.flow/project.md` (the vocabulary and references guards read the new comment).

  - [x] **Step 5: Commit**

    `feat(scripts): record check-unfinished-work verdicts and surface prior false positives`

- [x] 5. `gather-dispatch-context.sh` carries `## incidents`

**Build:** green
**Files:** `scripts/gather-dispatch-context.sh`, `scripts/test-gather-dispatch-context.sh`
**Tests:** Case 44: with a stub `flow` returning two incidents, the bundle carries a
`## incidents` section after `## project commands` holding a six-column table with the newer row
first and the census counts it as found; Case 45: `[]` yields `skipped: incidents (none)` and no
section; Case 46: no `flow` on `PATH` yields `skipped: incidents (flow unavailable)`, no section,
exit 0; Case 47: a new incident changes the body hash so the bundle is rebuilt; Case 48: a
principles-path with a mid-string `(` that does not end in `)` still gets the `(absent)` suffix;
Case 49: a principles-path shaped `notes (v2)` and one shaped `foo(bar)` (both end in `)`, one with
a leading space before `(` and one without) still get the `(absent)` suffix — `render_body` no
longer pattern-matches the skip label's shape at all, since every label is now formatted in full at
the point it's pushed onto `SKIPPED_LABELS`.
**Regression:** revert this commit and the bundle carries no `## incidents` section in case 44,
and the census line in cases 45–46 names no `incidents` entry.
**Baseline:** before=67 after=75
<!-- measured: bash scripts/test-gather-dispatch-context.sh | grep -c '^ok:' @ main 16a7c33 -->
<!-- predicted: the same command after task 5 — eight new `ok:` lines across cases 44-49, 67 + 8 -->
**Commit:** feat(scripts): hand every dispatch the project's incident table

  - [x] **Step 1: Write the failing cases**

    In `scripts/test-gather-dispatch-context.sh`, appended as cases 44-47, in the shape of the
    restricted-`PATH` cases (the `PATH="<dest-dir>"` pattern at lines 108–129): a stub `flow` in
    the case's `PATH` directory that prints a fixed JSON array for `record incidents`, and the
    four cases above. Numbered 44-47 rather than 27-30 as originally drafted here: by the time
    this task ran, the file already had cases 27 through 43 taken by unrelated prior work, and
    this file's own no-renumbering convention (recorded elsewhere in its header comments) means
    existing case numbers are never shifted to make room, so the four new ones were appended at
    the end instead. The before=67/after=71 baseline in the Tests field above is unaffected. The
    expected table row for case 44 is exactly
    `| 2026-09-03 | check-task-commit-fields | Baseline revert re-entered mid-flight | aborted revert, restored dir from stash^3 | 55 | kan-423 |`.
    <!-- measured: grep -c '^ok:' output and grep -n 'PATH=' scripts/test-gather-dispatch-context.sh @ main 16a7c33 -->

  - [x] **Step 2: See them fail**

    `bash scripts/test-gather-dispatch-context.sh`.

  - [x] **Step 3: Add the section**

    In `scripts/gather-dispatch-context.sh`, after the project-commands block (line ~443) and
    before `render_body`: run `flow record incidents -C "$WORKTREE_REAL"` with stderr to a temp
    file when `command -v flow` succeeds; on exit 0 and `jq 'length' > 0`, render
    `INCIDENTS_BODY` with `jq -r` as the header row plus one row per element
    (`.occurredAt[:10]`, `.guard`, `.symptom`, `.recovery`, `.minutesLost`, `.change // "—"`),
    piping cell text through `gsub("|"; "\\|")`, push `incidents` onto `FOUND_LABELS` with
    `FOUND_PATHS+=("@incidents")`; on `[]` push `SKIPPED_LABELS+=("incidents (none)")`; on any
    failure or no `flow`, `SKIPPED_LABELS+=("incidents (flow unavailable)")`. `render_body`'s
    `case` (line 475) gains `"@incidents") printf '%s\n' "$INCIDENTS_BODY" ;;`, and its skipped
    loop prints a label already carrying its parenthesised reason without appending `(absent)`
    — match `*"("*` to decide. Header comment: one paragraph under the project-commands one,
    stating the three outcomes and that the section is hashed.
    <!-- measured: sed -n 445,500p scripts/gather-dispatch-context.sh @ main 16a7c33 -->

  - [x] **Step 4: Run green and the lint list, then commit**

    `bash scripts/test-gather-dispatch-context.sh` and the `## lint` list; then
    `feat(scripts): hand every dispatch the project's incident table`.

- [x] 6. `integrate.md` shows the line and records the flag on Continue

**Build:** green
**Files:** `skills/flow/integrate.md`, `skills/flow-contracts/finish-contract-run1.md`
**Tests:** none — prose; `scripts/check-contract-budget.sh`, `scripts/check-references.sh` and
`scripts/check-vocabulary.sh` are the checks
**Regression:** revert this commit and an OUTSTANDING gate never records a false positive, so
task 4's stderr line stays at zero forever.
**Baseline:** not applicable — no test count changes
**Commit:** feat(flow): record a hand-verified OUTSTANDING as a guard false positive

  - [x] **Step 1: Edit `integrate.md` §1**

    In the `**OUTSTANDING:**` bullet (line ~45): "show the breakdown — and the guard's
    `prior false positives for this guard on this project` stderr line when it printed —". After
    the prompt's three courses, one paragraph: when the operator chooses **Continue** and their
    answer says the verdict was verified structural (the plan held in another worktree with every
    task ticked and no open finding), run, once per worktree that reported OUTSTANDING and before
    proceeding to **2**:

```bash verified:the flag set is design.md §5's; the call shape mirrors the flow record findings call at scripts/check-unfinished-work.sh:317
flow record verdict false-positive -change <name> -guard check-unfinished-work \
  -reason "<the operator's reason, verbatim>" -C <worktree>
```

    A write that falls back to the journal is one warning line and the run continues.
    <!-- measured: sed -n 39,60p skills/flow/integrate.md @ main 16a7c33 -->

  - [x] **Step 2: Edit the contract's Continue row**

    `skills/flow-contracts/finish-contract-run1.md` line 91, the **Continue** row: append "— and,
    where the operator called the verdict structural, records it as a guard false positive per
    **1. Check for unfinished work** (`skills/flow/integrate.md`)". No second copy of the command.
    <!-- measured: sed -n 84,92p skills/flow-contracts/finish-contract-run1.md @ main 16a7c33 -->

  - [x] **Step 3: Run the lint list and commit**

    Every `## lint` command in `.flow/project.md`; raise the two files' rows in
    `scripts/check-contract-budget.sh` only if the guard trips. Then
    `feat(flow): record a hand-verified OUTSTANDING as a guard false positive`.
