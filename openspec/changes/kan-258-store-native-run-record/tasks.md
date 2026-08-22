# Tasks — the run record becomes store-native

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** move the pipeline's derived records — every subagent dispatch and every review-panel
finding — out of worktree files and into the stats store, and render their Markdown from it.

**Architecture:** two tables (`dispatches`, `findings`) behind the four layers `myflow stage` already
uses (`cmd/myflow` → `internal/client` → `internal/api` → `internal/store`), with the existing
journal-on-failure path reused verbatim. Cost is derived by a second, independent
`internal/harvest` attribution pass, never self-reported. The Markdown record becomes a render, and
`preserve-session-records.sh` is retired.

**Spec:** `openspec/changes/kan-258-store-native-run-record/design.md`, and the delta specs beside it.

## Global constraints

- **Go 1.x as the repository already pins it** — this change adds no module dependency. Everything it
  needs (`pgx/v5`, `net/http`, `encoding/json`) is already in `stats/go.mod`.
- **The CLI knows only HTTP and the on-disk fallback format.** `cmd/myflow` SHALL NOT import
  `internal/store`. `journal flush` is the one existing exception and this change adds no second one
  — stated in `stats/cmd/myflow/main.go`'s own header comment.
- **`internal/harvest` imports nothing from `internal/store`.** Window sources are consumer-side
  interfaces, satisfied by the daemon's wiring in `cmd/myflowd/main.go`.
- **A record write never blocks.** Every store failure journals, warns once on stderr, and exits 0.
  The only non-zero exit is a caller mistake.
- **Never weaken a guard to make a task pass.** `.myflow/project.md`'s `## lint` list is the check
  set; `cd stats && gofmt -w .` is the only auto-fix in this repository.
- **The session-token rule holds for the new verb too**: `-session-token` takes a literal, never a
  shell substitution.

Names fixed once here so no task re-derives them:

```go unverified:the package does not exist yet — this is the shape tasks 2, 4, 6 and 12 build against
// stats/internal/records/types.go — the wire shape, shared by api, client and the renderer.
package records

type Dispatch struct {
    ID           int64           `json:"id"`
    AgentID      string          `json:"agentId,omitempty"`
    Seq          int             `json:"seq"`
    StageRunID   *int64          `json:"stageRunId,omitempty"`
    TaskID       string          `json:"taskId,omitempty"`
    Role         string          `json:"role"`
    Slot         string          `json:"slot,omitempty"`
    Model        string          `json:"model"`
    CommitSHA    string          `json:"commitSha,omitempty"`
    Outcome      string          `json:"outcome,omitempty"`
    SessionToken string          `json:"sessionToken,omitempty"`
    StartedAt    time.Time       `json:"startedAt"`
    EndedAt      *time.Time      `json:"endedAt,omitempty"`
    Metrics      json.RawMessage `json:"metrics,omitempty"`
    Notes        string          `json:"notes,omitempty"`
}

type Finding struct {
    Ref         string `json:"ref"`
    DispatchSeq *int   `json:"dispatchSeq,omitempty"`
    Round      int    `json:"round"`
    Slot       string `json:"slot"`
    Severity   string `json:"severity"`
    Location   string `json:"location,omitempty"`
    Note       string `json:"note"`
    Status     string `json:"status"`
    Reproducer string `json:"reproducer,omitempty"`
}

type Run struct {
    Change     string     `json:"change"`
    Dispatches []Dispatch `json:"dispatches"`
    Findings   []Finding  `json:"findings"`
}
```

Routes, all scoped by project and change:

```text unverified:the routes do not exist yet — task 4 adds them to internal/api/server.go
POST   /api/v1/records/{project}/{change}/dispatches
POST   /api/v1/records/{project}/{change}/findings
PATCH  /api/v1/records/{project}/{change}/findings/{ref}
GET    /api/v1/records/{project}/{change}
```

Constraint names, referenced by the retry logic in task 2 and asserted in task 1:
`dispatches_seq_key`, `findings_ref_key`.

**Ordering.** Task 24 was added after tasks 16–17's review and runs before task 23. Task 23 was added
after task 17 landed and runs immediately after it, before task 18 —
it closes a blocker that would make every finish run 1 report OUTSTANDING. Task 22 was added after
tasks 13–15 landed and runs immediately after them, before
task 16 — it closes gaps the retirement opened, and leaving them open would make task 19's own grep
fail. Tasks 20 and 21 were added after task 10 landed and run **immediately after it**,
before task 11 — they change the `dispatches` shape, and doing that before the renderer reads it
avoids rewriting the renderer. Otherwise tasks 1–12 are the Go half and run in order; each red/green pair is dispatched
together. Tasks 13–19 are the contract half. Task 13 must land **after** task 12, because retiring
the script before its replacement exists leaves run 1 with neither. Tasks 14–19 touch disjoint files
and may run in any order among themselves after 13.

---

### 1 Store tests for `dispatches` and `findings`

**Build:** red

**Squash-with:** Task 2

**Files:**
- Create: `stats/internal/store/records_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the executable statement of what `RecordDispatch`, `UpsertFinding`, `SetFindingStatus`,
  `MergeDispatchMetrics` and `RunRecord` must do. Task 2 satisfies it.

Written before the implementation so the cases assert what the delta spec's **One dispatch row
carries task, commit, model and cost together** and **A finding is a row whose status is updated in
place** require, rather than what the implementation happens to do.

Follow `stats/internal/store/stageruns_test.go`'s existing idiom exactly — the same
`newTestStore(t)` helper, the same `postgres:18-alpine` container fixture, the same `t.Cleanup`
truncation. Do **not** introduce a second fixture shape.

- [x] **Step 1: The migration applies twice idempotently**

Assert `EmbeddedMigrationCount()` rises by exactly one and that a second `RunMigrations` call records
no additional row in `schema_migrations`, mirroring the existing migration idempotency test.

- [x] **Step 2: `RecordDispatch` allocates `seq` and returns the row**

Two dispatches for one change come back with `Seq` 1 and 2. A dispatch for a *different* change also
starts at 1 — `seq` is per change, not global.

- [x] **Step 3: Concurrent `RecordDispatch` calls do not overwrite each other**

Fire sixteen goroutines against one change, wait, then assert sixteen rows exist with sixteen
distinct `seq` values. This is the case that stops the implementation being written as a
read-then-insert; it is the same shape `TestBeginStageAllocatesAttemptsUnderConcurrency` already
uses.

- [x] **Step 4: A duplicate finding ref is refused**

`UpsertFinding` for `F1` twice with different notes updates the row rather than inserting a second;
`RunRecord` then reports exactly one `F1` carrying the second note. Assert the constraint name
`findings_ref_key` appears in the error when a raw insert bypasses the upsert, so the retry logic in
task 2 has something to key on.

- [x] **Step 5: `SetFindingStatus` updates in place**

Set `F1` to `fixed`; assert one row, status `fixed`, every other column unchanged.

- [x] **Step 6: `MergeDispatchMetrics` deep-merges**

Write `{"tokens":{"input":10}}`, then `{"tokens":{"output":5}}`; assert both keys survive. A
top-level `||` concatenation loses `input` here, which is exactly what `jsonb_deep_merge` exists to
prevent — the same defect `stage_runs.metrics` already had to design around.

- [x] **Step 7: `RunRecord` returns dispatches in `seq` order and findings in `ref` order**

- [x] **Step 8: Run the tests and confirm they fail**

```bash verified:the package exists today and `go test` is .myflow/project.md's ## test command for it
cd stats && go test ./internal/store -run 'Record|Finding|Dispatch' -count=1
```

Expected: FAIL — `undefined: RecordDispatch` and siblings.

**Tests:** the cases enumerated in steps 1–7, in `stats/internal/store/records_test.go`.

**Regression:** Reverting this task leaves the store's record methods with no executable statement of
their contract, so the concurrency guarantee in particular becomes unverifiable.

**Baseline:** `./internal/store` before=128 after=135 top-level tests.
<!-- measured: cd stats && go test ./internal/store -list '.*' | grep -c '^Test' @ branch main -->

**Commit:** `test(store): assert the run record's dispatch and finding contracts`

---

### 2 Migration `0010_run_records.sql` and `internal/store/records.go`

**Build:** green

**Files:**
- Create: `stats/internal/records/types.go`
- Create: `stats/internal/store/migrations/0010_run_records.sql`
- Create: `stats/internal/store/records.go`

`internal/records/types.go` is created **here**, not in task 4: the store signatures below name
`records.Dispatch` and `records.Finding`, so the package must exist before this task compiles. Task 4
extends the same file; it does not create it.

**Interfaces:**
- Consumes: task 1's test cases.
- Produces, for tasks 4, 10 and 12:

```go unverified:signatures fixed by this plan; the file does not exist yet
func (s *Store) RecordDispatch(ctx context.Context, projectKey, change string, in records.Dispatch) (records.Dispatch, error)
func (s *Store) UpsertFinding(ctx context.Context, projectKey, change string, in records.Finding) (records.Finding, error)
func (s *Store) SetFindingStatus(ctx context.Context, projectKey, change, ref, status string) error
func (s *Store) MergeDispatchMetrics(ctx context.Context, dispatchID int64, patch json.RawMessage) error
func (s *Store) RunRecord(ctx context.Context, projectKey, change string) (records.Run, error)
```

`DispatchWindowsForSession` is **not** this task's: its return type is `harvest.DispatchWindow`,
which task 10 creates, and `internal/harvest` is outside this task's file list. Task 10 adds the type,
the source interface and this store method together.

- [x] **Step 1: Write the migration**

Both tables exactly as **1. Schema** in `design.md` specifies, with the two named constraints and the
indexes listed there. Follow `0003_stage_runs.sql`'s commenting convention: the file's header
explains *why* each non-obvious choice is what it is, not what the SQL says.

- [x] **Step 2: `RecordDispatch`, with the attempt-allocation idiom**

Allocate `seq` inside a single `INSERT … SELECT COALESCE(MAX(seq),0)+1 …`, with
`dispatches_seq_key` as the race detector and a bounded retry on that constraint name — the exact
shape `insertStageRunAndSupersede` already uses. Reuse `isUniqueViolation(err, constraint)` rather
than writing a second copy.

- [x] **Step 3: `UpsertFinding` and `SetFindingStatus`**

`INSERT … ON CONFLICT ON CONSTRAINT findings_ref_key DO UPDATE`. `SetFindingStatus` is a plain
`UPDATE … WHERE change_id = … AND ref = …` returning `pgx.ErrNoRows` as a typed
`ErrFindingNotFound` so task 4 can map it to a 404 rather than a 500.

- [x] **Step 4: `MergeDispatchMetrics`**

`UPDATE dispatches SET metrics = jsonb_deep_merge(metrics, $2)`, with the same `ErrNilMetricsPatch`
guard `MergeMetrics` already carries — a nil patch must be refused in Go, not surfaced as a NOT NULL
violation.

- [x] **Step 5: `RunRecord` and `DispatchWindowsForSession`**

- [x] **Step 6: Run the tests and confirm they pass**

```bash verified:the package exists today and `go test` is .myflow/project.md's ## test command for it
cd stats && gofmt -w . && go test ./internal/store -count=1 && go vet ./...
```

Expected: PASS.

**Tests:** none added here — task 1 carries them.

**Regression:** Reverting this task removes both tables and every store method, so no record can be
written at all and task 1's suite fails wholesale.

**Baseline:** `./internal/store` before=135 after=135 top-level tests; all pass.

**Commit:** `feat(store): add the dispatches and findings tables and their write path`

---

### 3 API handler tests for the record routes

**Build:** red

**Squash-with:** Task 4

**Files:**
- Create: `stats/internal/api/records_test.go`
- Modify: `stats/internal/api/changes_test.go` — `fakeStore`'s record bookkeeping. Go cannot split a
  struct across files, and this is where the stage and stats bookkeeping already live.
- Modify: `stats/internal/store/records_test.go`

**Interfaces:**
- Consumes: task 2's store methods, through a fake satisfying the handler's own `RecordStore`
  interface.
- Produces: the executable statement of the four routes' status codes and bodies. Task 4 satisfies
  it.

Follow `stats/internal/api/stages_test.go`'s idiom: a fake store, `httptest.NewRequest`, assertions
on status code and decoded body. No database.

- [x] **Step 1: A dispatch POST returns 201 and the allocated `seq`**

- [x] **Step 2: A finding POST returns 201; a second POST for the same ref returns 200**

201 for created, 200 for updated — so a caller can tell an upsert that inserted from one that
replaced, which the CLI's own output distinguishes.

- [x] **Step 3: A PATCH to an unknown ref returns 404, not 500**

Assert the handler maps `ErrFindingNotFound` rather than letting it surface as a server error.

- [x] **Step 4: A GET returns dispatches in `seq` order and findings in `ref` order**

- [x] **Step 5: A malformed body returns 400 and does not reach the store**

Assert the fake store recorded zero calls — a decode failure must never be a partial write.

- [x] **Step 6: The `Myflow-Daemon` response header is set on every route**

The client's `do` checks it to tell the daemon from something else answering on the port; a new route
that omits it would be classified as "not the daemon" and take the fallback path on every call.

Three further cases go in `stats/internal/store/records_test.go`, stating the corrections tasks 1
and 2 surfaced:

- [x] **Step 7: Findings order naturally, not lexically**

Insert `F1`, `F2` and `F10`; assert `RunRecord` returns them in that order. `ORDER BY ref` alone
returns `F1, F10, F2`, so this is the case that stops the ordering being written as a plain lexical
sort — a ten-finding panel would otherwise render out of order.

- [x] **Step 8: A finding resolves its raising dispatch**

Record a dispatch, then a finding carrying that dispatch's `Seq`; assert the row's `dispatch_id` is
that dispatch's id and that `RunRecord` returns the `Seq` back. A finding whose `DispatchSeq` is nil
leaves `dispatch_id` NULL, which is the legitimate case for a finding no single dispatch raised.

- [x] **Step 9: `RecordDispatch` returns the row's id**

Assert the returned `records.Dispatch` carries a non-zero `ID`, so `MergeDispatchMetrics`'s key is
reachable through the typed API rather than only through a raw pool.

- [x] **Step 10: Run the tests and confirm they fail**

```bash verified:both packages exist today and `go test` is .myflow/project.md's ## test command for them
cd stats && go test ./internal/api -run Record -count=1 && go test ./internal/store -run 'Order|DispatchSeq|ReturnsID' -count=1
```

Expected: FAIL — the routes are not registered, and the three store cases assert behaviour task 2
deliberately left out.

**Tests:** the cases in steps 1–6, in `stats/internal/api/records_test.go`; the cases in steps 7–9,
in `stats/internal/store/records_test.go`.

**Regression:** Reverting this task leaves the four routes with no executable statement of their
status codes, lets the 404-vs-500 mapping regress silently, and removes the only assertion that a
ten-finding record renders in order.

**Baseline:** `./internal/api` before=63 after=69 top-level tests; `./internal/store` before=135
after=138.
<!-- measured: cd stats && go test ./internal/api -list '.*' | grep -c '^Test' @ branch main -->

**Commit:** `test(api): assert the record routes' status codes and ordering`

---

### 4 `internal/records`, `internal/api/records.go`, the routes, and the client methods

**Build:** green

**Files:**
- Modify: `stats/internal/records/types.go`
- Modify: `stats/internal/store/records.go`
- Create: `stats/internal/api/records.go`
- Modify: `stats/internal/api/server.go` — the route table at lines 291–299
- Modify: `stats/internal/client/client.go`
- Modify: `stats/cmd/myflowd/main.go` — wire the handler

**Interfaces:**
- Consumes: task 2's store methods; task 3's test cases.
- Produces, for task 6 and task 12:

```go unverified:signatures fixed by this plan; the methods do not exist yet
func (c *Client) RecordDispatch(ctx context.Context, project, change string, in records.Dispatch) (records.Dispatch, error)
func (c *Client) RecordFinding(ctx context.Context, project, change string, in records.Finding) (out records.Finding, created bool, err error)
func (c *Client) SetFindingStatus(ctx context.Context, project, change, ref, status string) error
func (c *Client) GetRunRecord(ctx context.Context, project, change string) (records.Run, error)
```

- [x] **Step 1: Extend `internal/records/types.go` to the three types in full**

Task 2 created the file with the fields its own signatures needed. Bring it to exactly the three
types in **Global constraints** above — `Dispatch.ID` and `Finding.DispatchSeq` are the two additions.
This package holds data and, from task 12, rendering — nothing else. It imports only `encoding/json`
and `time`.

- [x] **Step 2: The three store corrections**

In `stats/internal/store/records.go`:

- `RecordDispatch` returns the inserted row's `id` in `records.Dispatch.ID`, so
  `MergeDispatchMetrics`'s key is reachable through the typed API. Task 2's tests reach it through a
  raw pool; that helper goes away with this step.
- `UpsertFinding` resolves `DispatchSeq` to `dispatch_id` in the same statement — a scalar subquery
  against `dispatches` for `(change_id, seq)` — so the column the design calls "the slot that raised
  it" is actually written. A nil `DispatchSeq` leaves it NULL, which is the legitimate case.
- `RunRecord` orders findings naturally rather than lexically:

```sql unverified:the ordering clause is fixed by this plan; task 2 shipped a plain ORDER BY ref
ORDER BY NULLIF(regexp_replace(ref, '\D', '', 'g'), '')::int NULLS LAST, ref
```

  A ref carrying no digits sorts last by `ref`, so a non-conforming value is ordered rather than
  crashing the cast.

- [x] **Step 3: The handler and its consumer-side `RecordStore` interface**

Define `RecordStore` in `internal/api`, at the consumer, naming only the methods the handler calls —
**four**, one per route. `MergeDispatchMetrics` is deliberately absent: it has no route, because the
harvester writes through it directly (task 10). The placement is the one `StageStore` already uses,
and is what keeps `internal/api` testable with a fake and no database.

- [x] **Step 4: Register the four routes**

Add them to `server.go`'s mux beside the existing `records`-free routes, keeping the file's existing
ordering convention (reads, then writes, then the SPA fallback).

- [x] **Step 5: The four client methods**

Reuse `c.do(req)` unchanged — it already owns the `Myflow-Daemon` header check, the timeout, and the
unreachable/refused classification. Do not add a second copy of that HTTP handling; that is exactly
the outcome `design.md`'s `daemon-owns-db` decision rejects.

- [x] **Step 6: Test the four client methods**

Task 3's cases cover the handler; these cover the client, in
`stats/internal/client/records_test.go`: the status-code mapping for each method, both success codes
for the finding upsert, and the unreachable path for all four. Landing four production methods with
no test of their own would be worse than the small widening of this task's stated scope.

- [x] **Step 7: Run the tests and confirm they pass**

```bash verified:go test is .myflow/project.md's ## test command for these packages
cd stats && gofmt -w . && go test ./internal/api ./internal/client ./internal/records ./internal/store -count=1
```

Expected: PASS. `go vet ./...` is **not** run here: it fails in this worktree until the SPA's
`dist` is built, for a reason unrelated to this change (`internal/web/embed.go`'s `//go:embed
all:dist`). Task 19 builds it and runs the full vet.
<!-- measured: reported by task 1+2's implementer against this worktree @ commit 01bf42a -->

**Tests:** step 6's cases, in `stats/internal/client/records_test.go`. Task 3 carries the handler and
store cases.

**Regression:** Reverting this task removes the four routes and the client methods, so the CLI has
nothing to call and every record write takes the fallback path unconditionally; it also restores the
lexical finding order and the unwritable `dispatch_id`.

**Baseline:** `./internal/api` before=69 after=69, `./internal/store` before=138 after=139, and
`./internal/client` before=40 after=45 top-level tests; all pass. The store's extra test is
`TestConcurrentUpsertFindingReportsCreatedExactlyOnce`, added by this task's review fix — the
`created` flag had no concurrency test at all, and the defect it now covers reproduced in the
review.
<!-- measured: cd stats && go test ./internal/client -list '.*' | grep -c '^Test' @ commit 01bf42a -->

**Commit:** `feat(api): serve the run record routes and add their client methods`

---

### 5 CLI tests for `myflow record`, including the journal fallback

**Build:** red

**Squash-with:** Task 6

**Files:**
- Create: `stats/cmd/myflow/record_test.go`

**Interfaces:**
- Consumes: task 4's client methods.
- Produces: the executable statement of the verb's argument parsing, exit codes and fallback
  behaviour. Task 6 satisfies it.

Follow `stats/cmd/myflow/stage_test.go`'s idiom exactly: drive `run(ctx, args, stdin, stdout,
stderr)` directly with in-memory streams, point `MYFLOW_STATE_DIR` at `t.TempDir()`, and use an
`httptest` server for the daemon. No subprocess.

- [x] **Step 1: A successful dispatch write prints one line and exits 0**

- [x] **Step 2: An unreachable store journals and still exits 0**

Point the client at a closed port. Assert exit 0, one `⚠ myflow: store unreachable — wrote local
journal` line on stderr, and that `<state-dir>/<project-key>/<change>.journal.record` exists
carrying the request body. This is the never-block guarantee; it is the case that must not be
allowed to regress into a non-zero exit.

- [x] **Step 3: A missing required flag exits 2 and writes no journal**

A caller mistake is not a store failure. Assert the journal file does **not** exist.

- [x] **Step 4: An unrecognised role exits 2, naming the accepted roles**

The CLI validates `role` against `implementer`, `reviewer`, `panel-fix`, `red-partner` before ever
contacting the store, mirroring how `stage begin` validates its stage key against
`internal/stages`.

- [x] **Step 5: A session token carrying a substitution exits 2**

Reuse `validateSessionToken` — do not write a second copy. Assert `-session-token 'mf-$(date +%s)'`
is refused and names why.

- [x] **Step 6: `myflow record` with no subcommand prints usage and exits 2**

- [x] **Step 7: Run the tests and confirm they fail**

```bash verified:the package exists today and `go test` is .myflow/project.md's ## test command for it
cd stats && go test ./cmd/myflow -run Record -count=1
```

Expected: FAIL — `unknown command "record"`.

**Tests:** the cases enumerated in steps 1–6, in `stats/cmd/myflow/record_test.go`.

**Regression:** Reverting this task removes the only executable statement that a record write never
blocks, which is the single guarantee this whole change most depends on.

**Baseline:** `./cmd/myflow` before=38 after=44 top-level tests.
<!-- measured: cd stats && go test ./cmd/myflow -list '.*' | grep -c '^Test' @ branch main -->

**Commit:** `test(myflow): assert the record verb's exit codes and journal fallback`

---

### 6 `cmd/myflow/record.go` and the usage block

**Build:** green

**Files:**
- Create: `stats/cmd/myflow/record.go`
- Modify: `stats/cmd/myflow/main.go` — the `usage` constant and `run`'s switch

**Interfaces:**
- Consumes: task 4's client methods; task 5's test cases.
- Produces, for task 12: `recordJournalPath(projectKey, name string) string`, returning
  `fallback.JournalFilePath(projectKey, name) + ".record"`.

- [x] **Step 1: `runRecord` and the four subcommands**

```text unverified:the flag set is fixed by this plan; the file does not exist yet
myflow record dispatch  -change <name> [-task <id>] -role <role> [-slot <s>] -model <m> \
                        [-commit <sha>] [-outcome <o>] -session-token <tok> \
                        -started-at <ts> [-C <dir>]
myflow record finding    -change <name> -ref F<n> [-round <r>] -slot <s> -severity <sev> \
                        [-location <l>] [-dispatch-seq <n>] -status <st> [-reproducer <cmd>] -note <text>
myflow record status     -change <name> -ref F<n> -status <st>
myflow record render     -change <name> -kind ledger|panel|all -repo <abs-repo-root>
```

`render` is declared here and returns "not implemented" until task 12 — declaring the whole verb in
one place keeps `main.go`'s switch from being edited twice. It exits **non-zero**: it rendered
nothing, and the never-block guarantee covers an unreachable store, not an absent subcommand.

**Three flags differ from a first reading of design.md, each for a stated reason.** `-task` is
optional: `design.md`'s schema makes `task_id` NULL "where the role runs against no task", and a
required flag would make that case unreachable — the same argument `dispatch_id` gets there,
inverted. `-dispatch-seq` exists because task 4 taught `UpsertFinding` to resolve `DispatchSeq` to
`dispatch_id`, and without a flag nothing could ever populate it; `0` is an unambiguous "not given",
since `seq` starts at 1. `-round` defaults to `0` with no required check, because `0` is the initial
panel and a real value — there is no "not given" to distinguish, unlike `stage end`'s negative
sentinels.

`myflow record finding` prints `recorded: F<n>` when the client reports `created` and `updated: F<n>`
when it does not, so a panel run can tell a new finding from a restated one. That is what the 201/200
split task 3 asserts and the `created bool` task 4 threads out exist for; a CLI that printed one word
either way would make both dead weight.

- [x] **Step 2: The journal fallback**

Model it on `journalStageMark` exactly: marshal a body carrying the subcommand's kind and the wire
request, append via `fallback.AppendJournalEntry`, print the one warning line, swallow the journal's
own error, exit 0. Do not invent a second journal mechanism.

- [x] **Step 3: Extend `usage` and `run`'s switch**

- [x] **Step 4: Run the tests and confirm they pass**

```bash verified:both commands are in .myflow/project.md's ## lint and ## test lists
cd stats && gofmt -w . && go test ./cmd/myflow -count=1 && go vet ./...
```

Expected: PASS.

**Tests:** none added here — task 5 carries them.

**Regression:** Reverting this task leaves the skills with no command to call; every `myflow record`
invocation in the contract half becomes `unknown command`.

**Baseline:** `./cmd/myflow` before=44 after=44 top-level tests; all pass.

**Commit:** `feat(myflow): add the record verb and its journal fallback`

---

### 7 Reconciler tests for `replayRecordFile`

**Build:** red

**Squash-with:** Task 8

**Files:**
- Create: `stats/internal/reconcile/record_test.go`

A file of its own, not an addition to `reconcile_test.go`: stage-mark replay already has its own
`stage_test.go` rather than living in that 747-line state-journal file, and record replay follows
the same structure.

**Interfaces:**
- Consumes: task 6's journal body shape.
- Produces: the executable statement of record replay. Task 8 satisfies it.

- [x] **Step 1: A `.journal.record` file replays into the store**

Write **three** entries by hand in the exact shape task 6 emits — one per kind: `dispatch`, `finding`
and `status`. Assert all three reach the fake store in file order and the file is retired. `status`
is a third kind with its own journalled shape (the ref sits in the body because the wire PATCH puts
it in the URL), and a test covering only two kinds would let replay mishandle it silently. Order
across kinds is what this case exists to assert, so one file carrying all three is the stronger
test.

- [x] **Step 2: A partial trailing line is left for the next run**

Assert the existing `splitCompleteLines` behaviour holds for this file kind too — a half-written
final line is not decoded and not retired.

- [x] **Step 3: An undecodable entry is refused, and the rest still apply**

The same shape `errChangeEntryDecodeFailed` already establishes for change entries: one bad line does
not cost the file.

- [x] **Step 4: Run the tests and confirm they fail**

```bash verified:the package exists today and `go test` is .myflow/project.md's ## test command for it
cd stats && go test ./internal/reconcile -run Record -count=1
```

Expected: FAIL. Not on `replayRecordFile` being undefined — these tests are in `package
reconcile_test` and can never name an unexported function. The real RED is a compile failure on
`reconcile.RecordStore` being undefined and `reconcile.New` taking too few arguments, which is the
shape the stage-mark predecessor produced too.

**Tests:** the cases enumerated in steps 1–3, in `stats/internal/reconcile/reconcile_test.go`.

**Regression:** Reverting this task leaves record replay unverified, so a journalled write could
silently never reach the store — the exact failure this change exists to close.

**Baseline:** `./internal/reconcile` before=14 after=19 top-level tests.
<!-- measured: cd stats && go test ./internal/reconcile -list '.*' | grep -c '^Test' @ branch main -->

**Commit:** `test(reconcile): assert record journal entries replay in file order`

---

### 8 `replayRecordFile` in `internal/reconcile`

**Build:** green

**Files:**
- Modify: `stats/internal/reconcile/reconcile.go`
- Modify: `stats/cmd/myflowd/main.go` — pass the record store to the reconciler
- Modify: `stats/cmd/myflow/journal.go` — `reconcile.New`'s third call site. `myflow journal flush`
  is the CLI's one sanctioned `internal/store` import, so widening `New`'s signature reaches it too.

**Interfaces:**
- Consumes: task 7's test cases; task 2's store methods.
- Produces: nothing later tasks call directly.

- [x] **Step 1: Add the third replay branch**

`Run`'s `WalkDir` already dispatches by suffix; add `.journal.record` beside `.journal` and
`.journal.stage`, calling a new `replayRecordFile` written in the same shape as `replayStageFile`.

- [x] **Step 2: Wire the record store into `New`**

- [x] **Step 3: Run the tests and confirm they pass**

```bash verified:both commands are in .myflow/project.md's ## lint and ## test lists
cd stats && gofmt -w . && go test ./internal/reconcile ./cmd/myflowd -count=1 && go vet ./...
```

Expected: PASS.

**Tests:** none added here — task 7 carries them.

**Regression:** Reverting this task strands every journalled record write on disk forever, which
makes the never-block guarantee a data-loss guarantee instead.

**Baseline:** `./internal/reconcile` before=19 after=19 top-level tests; all pass.

**Commit:** `feat(reconcile): replay record journal entries into the store`

---

### 9 Harvest tests for dispatch-window attribution

**Build:** red

**Squash-with:** Task 10

**Files:**
- Modify: `stats/internal/harvest/attribute_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks — this package imports no store.
- Produces: the executable statement of the second attribution pass. Task 10 satisfies it.

Reuse the existing `mainThreadRecords(t)` and `sidechainRecords(t)` fixtures. Do **not** add a new
fixture file; the two that exist already carry both a main-thread and a sidechain model.

- [x] **Step 1: Sidechain usage inside a dispatch window lands on that dispatch**

Build one `DispatchWindow` spanning the sidechain fixture's timestamps; assert the returned map keys
on the dispatch id and that its `Sidechain` bucket carries the fixture's own figures.

- [x] **Step 2: Main-thread usage is not attributed to a dispatch**

The parent's own tokens belong to the stage run, not to the subagent it dispatched. Assert the
dispatch's `Main` bucket is zero.

- [x] **Step 3: Stage attribution is unchanged**

Run the existing `Attributor` over the same records and assert its result is byte-identical to what
`TestSidechainAccumulatesSeparately` already asserts. This is the case that proves the second pass
is additive rather than a change to the first.

- [x] **Step 4: A record outside every dispatch window is attributed to none**

- [x] **Step 5: The half-open interval holds for dispatch windows too**

A record timestamped exactly at one window's `EndedAt`, which equals the next window's `StartedAt`,
resolves to the window that is *starting* — the same convention `Window` already documents.

- [x] **Step 6: Run the tests and confirm they fail**

```bash verified:the package exists today and `go test` is .myflow/project.md's ## test command for it
cd stats && go test ./internal/harvest -run Dispatch -count=1
```

Expected: FAIL — `DispatchAttributor` is undefined.

**Tests:** the cases enumerated in steps 1–5, in `stats/internal/harvest/attribute_test.go`.

**Regression:** Reverting this task leaves per-dispatch cost unverified, and in particular removes
the assertion that stage attribution did not change.

**Baseline:** `./internal/harvest` before=71 after=76 top-level tests.
<!-- measured: cd stats && go test ./internal/harvest -list '.*' | grep -c '^Test' @ branch main -->

**Commit:** `test(harvest): assert dispatch windows attribute sidechain usage`

---

### 10 `DispatchWindowSource` and the second attribution pass

**Build:** green

**Files:**
- Modify: `stats/internal/harvest/attribute.go`
- Modify: `stats/internal/harvest/watcher.go` — run the second pass beside the first
- Modify: `stats/internal/store/records.go` — add `DispatchWindowsForSession`
- Modify: `stats/cmd/myflowd/main.go` — satisfy the new source from that store method

**Interfaces:**
- Consumes: task 9's test cases. **Not** a store method from an earlier task: `DispatchWindowsForSession`
  returns `harvest.DispatchWindow`, a type this task creates, so the type, the source interface and
  the store method land together here. Task 2 deliberately omitted it for that reason.
- Produces, for the daemon's wiring:

```go unverified:signatures fixed by this plan; the types do not exist yet
type DispatchWindow struct {
    DispatchID int64
    SessionID  string
    StartedAt  time.Time
    EndedAt    *time.Time
}
type DispatchWindowSource interface {
    DispatchWindowsForSession(ctx context.Context, sessionID string) ([]DispatchWindow, error)
}
func NewDispatchAttributor(w DispatchWindowSource) *DispatchAttributor
func (a *DispatchAttributor) Attribute(ctx context.Context, records []Record) (map[int64]Delta, error)
```

- [x] **Step 1: The window type, the source interface and the attributor**

`DispatchAttributor.Attribute` accumulates **only** records whose `IsSidechain` is true, into the
`Sidechain` bucket, reusing `Delta`, `Bucket` and `TokenDelta` unchanged.

- [x] **Step 2: Window selection**

Write the selection inline rather than bending `bestWindow` to serve two window types. This is
deliberate: `bestWindow`'s tie-break reads `Window.Attempt`, which has no dispatch analogue, and an
abstraction bent to fit a second, differently-shaped caller costs more than the small amount of
logic repeated here — the same reasoning `views/RunDetail.tsx`'s `RunPanel` records for its own
duplication.

- [x] **Step 3: Run both passes in the watcher**

The second pass reads the same `records` slice the first does, in the same batch, and merges its
deltas through `MergeDispatchMetrics`. A failure in the second pass must not cost the first: log it
and continue, exactly as the existing pass already handles a merge failure.

- [x] **Step 4: `store.DispatchWindowsForSession`, then wire the source in the daemon**

The store method joins `dispatches` to `changes` for the session token bound to `sessionID`, exactly
as `QueryStageRuns` already resolves a session's stage windows, and returns `[]harvest.DispatchWindow`.
`internal/store` importing `internal/harvest` is the direction that already holds for the stage
window source; `internal/harvest` still imports nothing from `internal/store`.

- [x] **Step 5: Run the tests and confirm they pass**

```bash verified:go test is .myflow/project.md's ## test command for these packages
cd stats && gofmt -w . && go test ./internal/harvest ./internal/store ./cmd/myflowd -count=1
```

Expected: PASS.

**Tests:** none added here — task 9 carries them.

**Regression:** Reverting this task leaves every dispatch's `metrics` bag permanently empty, so
per-dispatch cost — one of the four records this change exists to make store-native — is recorded
nowhere.

**Baseline:** `./internal/harvest` before=76 after=76 top-level tests; all pass.

**Commit:** `feat(harvest): attribute sidechain usage to dispatch windows`

---

### 20 Tests for agent-id attribution and a stable tie-break

**Build:** red

**Squash-with:** Task 21

**Files:**
- Modify: `stats/internal/harvest/attribute_test.go`
- Modify: `stats/internal/store/records_test.go`
- Modify: `stats/cmd/myflow/record_test.go`

**Interfaces:**
- Consumes: task 10's `DispatchAttributor`, `DispatchWindow` and `bestDispatchWindow`.
- Produces: the executable statement of what agent-id matching must do. Task 21 satisfies it.

**Why this pair exists.** A review panel dispatches its slots concurrently against one parent
session, so their dispatch windows overlap and every sidechain record in the overlap is credited to
whichever slot started last. The change's total stays correct; its split across slots does not — and
per-slot cost is exactly what this record is for. Transcript records already carry an `agentId` that
separates them cleanly; `dispatches` has no column to match it against. Found by task 9+10's
implementer, and the tie-break half of it by that pair's reviewer.
<!-- measured: reported against commit 872cef2 by the implementer and the reviewer of tasks 9+10 -->

- [x] **Step 1: A sidechain record matches the dispatch carrying its agent id**

Two dispatch windows overlapping the same instant, with different agent ids. Assert each record lands
on the dispatch whose `AgentID` equals the record's own, **not** on whichever window started last.

- [x] **Step 2: A record whose agent id matches no dispatch falls back to the window rule**

Cursor and Codex expose no agent id, and a record harvested before this column existed carries none.
The fallback is not a legacy path to be removed later — it is the normal path on two of three
harnesses, and it must keep working.

- [x] **Step 3: A dispatch with no agent id still receives records by the window rule**

The mirror of step 2: the column is nullable, and an empty value means "not reported", never "matches
the empty agent id".

- [x] **Step 4: An exact `StartedAt` tie resolves the same way every time**

Two windows with byte-identical `StartedAt` and no agent id. Assert the same dispatch wins across
repeated runs. Today `bestDispatchWindow`'s `After` comparison is false for both and the
first-encountered window wins, while `DispatchWindowsForSession`'s `ORDER BY d.started_at` has no
secondary key — so Postgres does not guarantee which row that is. The stage grain already has
`TestSameInstantWindowsFallBackToHighestAttempt` for its own version of this; the dispatch grain has
nothing.

- [x] **Step 5: `myflow record dispatch` accepts and sends `-agent-id`**

Optional. Absent means the harness reported none.

- [x] **Step 6: Run the tests and confirm they fail**

```bash verified:go test is .myflow/project.md's ## test command for these packages
cd stats && go test ./internal/harvest ./internal/store ./cmd/myflow -count=1
```

Expected: FAIL — `AgentID` is undefined and the tie-break is unstable.

**Tests:** the cases in steps 1–5, in the three files above.

**Regression:** Reverting this task removes the only proof that concurrent panel slots are costed
separately, and the only proof that a tie resolves reproducibly.

**Baseline:** `./internal/harvest` before=76 after=80, `./internal/store` before=140 after=142,
`./cmd/myflow` before=44 after=45.

**Commit:** `test(harvest): assert agent-id attribution and a stable dispatch tie-break`

---

### 21 Agent-id matching and the stable tie-break

**Build:** green

**Files:**
- Create: `stats/internal/store/migrations/0011_dispatch_agent_id.sql`
- Modify: `stats/internal/records/types.go`
- Modify: `stats/internal/store/records.go`
- Modify: `stats/internal/api/records.go`
- Modify: `stats/internal/client/client.go`
- Modify: `stats/cmd/myflow/record.go`
- Modify: `stats/internal/harvest/attribute.go`

**Interfaces:**
- Consumes: task 20's test cases.
- Produces: `records.Dispatch.AgentID`, and `harvest.DispatchWindow.AgentID`.

- [x] **Step 1: Migration 0011 adds a nullable `agent_id`**

A **new** migration, never an edit to `0010` — that file is already applied to real databases, and
migrations are tracked by filename with no checksum, so editing one silently diverges an existing
database from a fresh one.

**Do not index the column.** Nothing queries by it: matching happens in Go, in `bestDispatchWindow`,
over the row set `DispatchWindowsForSession` has already fetched by session token. An index would be
charged on every dispatch insert and never read — `EXPLAIN` on that query plans a hash semi join over
a sequential scan and references no `agent_id` index. Add one only alongside a query that would use
it.

- [x] **Step 2: Thread `AgentID` through the wire type, store, API and client**

Nullable end to end. An absent value is `""` and means "not reported" — never a value that matches
another absent one.

- [x] **Step 3: `-agent-id` on `myflow record dispatch`**

Optional, for the same reason: two of the three supported harnesses expose no such id. A dispatch
recorded without one is not degraded, it is ordinary.

- [x] **Step 4: Match on agent id first, fall back to the window rule**

`bestDispatchWindow` prefers a window whose `AgentID` equals the record's own. With no match on
either side it falls back to exactly today's behaviour, so nothing that works now stops working.

- [x] **Step 5: Make the tie-break stable**

Add a secondary sort key to `DispatchWindowsForSession` — `ORDER BY d.started_at, d.id` — and state
in `bestDispatchWindow`'s own comment that "started last" is undefined at an exact tie and is broken
by the lower id. A comment claiming a temporal order that does not exist at a tie is worse than none.

- [x] **Step 6: Run the tests and confirm they pass**

```bash verified:go test is .myflow/project.md's ## test command for these packages
cd stats && gofmt -w . && go test ./internal/harvest ./internal/store ./internal/api ./internal/client ./cmd/myflow -count=1 && gofmt -l .
```

Expected: PASS, and `gofmt -l .` prints nothing.

**Tests:** none added here — task 20 carries them.

**Regression:** Reverting this task restores per-slot cost that is wrong whenever panel slots run
concurrently, which is the ordinary case, and an unstable tie-break.

**Baseline:** `./internal/harvest` before=80 after=80, `./internal/store` before=142 after=142,
`./cmd/myflow` before=45 after=45; all pass.

**Commit:** `feat(harvest): attribute sidechain usage by agent id, with a stable tie-break`

---
### 11 Renderer tests

**Build:** red

**Squash-with:** Task 12

**Files:**
- Create: `stats/internal/records/render_test.go`
- Modify: `stats/cmd/myflow/record_test.go`

**Interfaces:**
- Consumes: task 4's `records.Run` type.
- Produces: the executable statement of both renderings and of `render`'s outcome words. Task 12
  satisfies it.

- [x] **Step 1: The panel rendering satisfies the guard**

Render a `records.Run` carrying three findings, write it to a temp file, and run
`scripts/check-unfinished-work.sh` against it from the test. Assert it reports clear when every
status is `fixed` and outstanding when one is `open`. **This is the load-bearing case of the whole
change:** it proves the marker block survived the move into the store, against the real guard rather
than against a re-implementation of it.

- [x] **Step 2: `findings-total:` equals the marker-line count**

- [x] **Step 3: A note containing a marker label is neutralised**

Render a finding whose note is `finding-status: F9 fixed`. Assert the guard still counts the record's
real findings and not a fourth — a validly-formatted marker inside prose reads the same as a real
one, which is why the record's own format rule forbids quoting it.

- [x] **Step 4: The ledger rendering names each dispatch's model**

Including a dispatch whose model is `unknown (agent-defined)`, asserted to render verbatim. No step
may substitute a plausible slug.

- [x] **Step 5: `render` with rows prints `rendered:` and writes the file**

- [x] **Step 6: `render` with no rows prints `MISSING:` and exits 0**

Assert no file is created.

- [x] **Step 7: A change name carrying `/` or a glob metacharacter is refused**

Assert exit non-zero, a message naming why, and that nothing was written.

- [x] **Step 8: A destination outside the repo root is refused**

Point `-repo` at a temp dir and place a symlink at `docs/superpowers/ledgers` pointing outside it.
Assert the render is refused rather than following it.

- [x] **Step 9: A second render reuses the first render's date**

Create `docs/superpowers/ledgers/2020-01-01-demo.md`, render again, assert that file was overwritten
and no second dated file exists.

- [x] **Step 10: Run the tests and confirm they fail**

```bash verified:the packages exist after task 4; `go test` is .myflow/project.md's ## test command for them
cd stats && go test ./internal/records ./cmd/myflow -run 'Render|Record' -count=1
```

Expected: FAIL — `RenderLedger` is undefined and `record render` reports "not implemented".

**Tests:** the cases enumerated in steps 1–9, in `stats/internal/records/render_test.go` and
`stats/cmd/myflow/record_test.go`.

**Regression:** Reverting this task removes the only proof that a rendered panel record still
satisfies `check-unfinished-work.sh`, which is the assertion that lets the guard stay unchanged.

**Baseline:** `./cmd/myflow` before=45 after=52 top-level tests; `./internal/records` before=0
after=9; `./internal/store` before=142 after=143.
<!-- measured: cd stats && go test ./cmd/myflow ./internal/records ./internal/store -list '.*' | grep -c '^Test' @ base b61c5fb -->
<!-- the ./internal/store row is this task's store-backed renderer test: internal/records imports
     nothing but encoding/json and time, so a rendering asserted against a value that came out of a
     real database can only be asserted from the far side of the store. -->

**Commit:** `test(records): assert both renderings and the render command's outcomes`

---

### 12 `internal/records/render.go` and `myflow record render`

**Build:** green

**Files:**
- Create: `stats/internal/records/render.go`
- Modify: `stats/cmd/myflow/record.go` — implement the `render` subcommand

**Interfaces:**
- Consumes: task 11's test cases; task 4's `GetRunRecord`.
- Produces, for the contract half:

```go unverified:signatures fixed by this plan; the file does not exist yet
func RenderLedger(r Run) string
func RenderPanel(r Run) string
func Destination(repoRoot, kind, change string, today time.Time) (string, error)
```

- [x] **Step 1: `RenderLedger`**

One section per dispatch in `seq` order: the task id, the role, the model verbatim, the commit sha,
the outcome, and the token figures from `Metrics` where present. A dispatch whose `Metrics` is empty
renders the figures as `not measured`, never as zero — zero is a measurement.

- [x] **Step 2: `RenderPanel`**

The findings table, then the `findings-total:` line and the unbroken `finding-status:` span, then a
separate `reproducers-total:` block with one `finding-reproducer:` line per finding. The two marker
blocks stay apart because `check-unfinished-work.sh` requires the `finding-status:` lines to occupy
one unbroken span.

- [x] **Step 3: Neutralise marker labels in free text**

A note or location whose text would form a marker label has its label-forming colon replaced by a
lookalike that is not the marker character, and the substitution is documented in the function's own
comment. Neutralise on the way out, never on the way in — the store keeps the operator's text.

- [x] **Step 4: `Destination`, with both path protections**

Validate the change name against `^[A-Za-z0-9][A-Za-z0-9._-]*$` and require the resolved destination
to be contained within `repoRoot` after `filepath.EvalSymlinks`. Reuse an existing dated file for the
change when one exists, so the date is fixed at the first render.

- [x] **Step 4a: `-kind panel` always writes; `MISSING:` is `ledger`-only**

A panel that raised nothing produces no rows, and reporting `MISSING:` for it writes no record — which
`check-unfinished-work.sh` reads as OUTSTANDING for a genuinely clean change. `RenderPanel` already
handles the zero case correctly (`findings-total: 0`, no markers); it is the command's MISSING rule
that shadows it. The invocation happens at panel close, so it is itself the evidence a panel ran — no
sentinel row is stored. Settled by **The panel record declares how many findings it carries**
(`openspec/specs/myflow-review-panel-economics/spec.md`), which already states that zero findings is a
declaration and silence is not.

**The filename is `<date>-<change>-panel.md`**, matching the roughly forty records already in
`docs/superpowers/reviews/` and the script this change retires — not a convention of its own.
<!-- measured: ls docs/superpowers/reviews/ | wc -l @ branch main -->

- [x] **Step 5: The `render` subcommand**

Fetch through `GetRunRecord`, render, write, print one outcome line per kind. An unreachable store
prints `journalled: <kind>` and exits 0 — there is nothing to journal for a read, so the word records
that the render did not happen and why, rather than pretending an empty record.

- [x] **Step 6: Run the tests and confirm they pass**

```bash verified:both commands are in .myflow/project.md's ## lint and ## test lists
cd stats && gofmt -w . && go test ./... -race -count=1 && go vet ./... && gofmt -l .
```

Expected: PASS, and `gofmt -l .` prints nothing.

**Tests:** none added here — task 11 carries them.

**Regression:** Reverting this task leaves the records readable only through the API, so the archive
stops being self-contained — the regression `design.md` names explicitly.

**Baseline:** `stats` before=435 after=452 top-level tests across all packages; all pass.
<!-- measured: cd stats && go test ./... -list '.*' | grep -c '^Test' @ base b61c5fb; the count
     excludes ./cmd/myflowd and ./internal/web, which cannot build until task 19 produces
     stats/internal/web/dist -->

**Commit:** `feat(records): render the ledger and panel record from the store`

---

### 13 Retire `preserve-session-records.sh`

**Build:** green

**Files:**
- Delete: `scripts/preserve-session-records.sh`
- Delete: `scripts/test-preserve-session-records.sh`
- Modify: `setup.sh` — the symlink set
- Modify: `.myflow/project.md` — the `## test` list

**Interfaces:**
- Consumes: task 12's working `myflow record render`.
- Produces: an absence every later task's guard run asserts.

**This task must not run before task 12.** Retiring the script while its replacement does not exist
leaves finish run 1 with neither.

- [x] **Step 1: Delete both files**

- [x] **Step 2: Remove the symlink from `setup.sh`**

- [x] **Step 3: Remove `scripts/test-preserve-session-records.sh` from `.myflow/project.md`'s `## test` list**

- [x] **Step 4: Confirm the symlink guard is green**

```bash verified:both commands are in .myflow/project.md's ## lint list
scripts/check-guard-symlinks.sh && scripts/check-references.sh
```

Expected: exit 0 from both. `check-references.sh` will still fail here if any skill names the script
— that is tasks 14–17's work, so run it again at the end of task 19 rather than treating a hit here
as this task's defect.

**Tests:** none added — this task removes a suite rather than adding one. Its own verification is the
two guards above.

**Regression:** Reverting this task restores a script that duplicates the render's duty and can
disagree with it, which is the two-sources problem this change exists to remove.

**Baseline:** `scripts/test-preserve-session-records.sh` before=169 `ok:` assertions, after=the file
does not exist. The repository's remaining harnesses are unchanged.
<!-- measured: bash scripts/test-preserve-session-records.sh | grep -c '^ok:' @ branch main -->

**Commit:** `chore(scripts): retire the session-record preservation script`

---

### 14 `pipeline.md` — the render-outcome table, the registry rows, the model-policy pointer

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`

**Interfaces:**
- Consumes: task 12's outcome words.
- Produces: the contract rows tasks 15, 16 and 17 cite rather than restate.

Four edits, all in one file:

- [x] **Step 1: Replace "Preserving the session records"**

The section becomes **Rendering the session records**, carrying this table and nothing that
restates the render's own procedure:

| Outcome | What it means | What you do |
|---------|---------------|-------------|
| `rendered: <dest>`, exit 0 | The record was written to that path from the store's rows. | Nothing. Proceed. |
| `MISSING: <kind> — no rows for <change>`, exit 0 | The store holds no rows of that kind for this change. | **Report it.** It means no record exists, never that one was written elsewhere. Proceed. |
| `journalled: <kind>`, exit 0 | The store could not be reached, so nothing was rendered. | **Report it, naming the kind.** Proceed with the integration. |
| A message on stderr, **exit non-zero** | A destination was refused or could not be written. | **Report it, with the command's own message.** Then proceed. |

Keep the existing paragraph's two guarantees, reworded for the new words: a non-zero exit is never
silent and never a stop, and the remaining kinds are still attempted after any one failure.

**Delete** the "Do not harmonise the two orderings for symmetry" paragraph — it describes an
asymmetry between a copy step and a staging step that no longer exists.

- [x] **Step 2: The Temporary artifacts registry**

The `Panel record` and `SDD ledger` rows change. Their Location cell becomes `the store`, and their
"Removed by" cell becomes `nothing — the store is the terminal record`, matching how the `State file`
row already reads. Add one row for the rendered files:

| Artifact | Created by | Location | Removed by |
|----------|-----------|----------|-----------|
| Rendered ledger and panel record | `myflow record render` | `<project>/docs/superpowers/` | nothing — they are committed and archived with the change |

- [x] **Step 3: Model policy's persistence sentence**

The sentence naming the preserved ledger under `docs/superpowers/ledgers/` as what makes a dispatch's
model answerable after archiving is repointed at the store, and gains the fact the store adds: the
question is now answerable **across** changes, by query, not only by reading one change's file.

- [x] **Step 4: Guards green**

```bash verified:both commands are in .myflow/project.md's ## lint list
scripts/check-references.sh && scripts/check-contract-budget.sh
```

`check-contract-budget.sh` may report `pipeline.md` over budget; raising its row is task 19's work,
not this task's — leave it and record the number.

**Tests:** none — contract prose, covered by the two guards above and by task 19's full run.

**Regression:** Reverting this task leaves the canonical contract describing a copy step that no
longer exists, so a run reading it would call a deleted script.

**Baseline:** `pipeline.md`'s `check-contract-budget.sh` figure before and after, recorded in the
commit body; the guard's verdict is what task 19 acts on.

**Commit:** `docs(pipeline): replace session-record preservation with rendering from the store`

---

### 15 `finish-contract.md` and `myflow-finish/SKILL.md`

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/finish-contract.md`
- Modify: `skills/myflow-finish/SKILL.md`

**Interfaces:**
- Consumes: task 14's outcome table, cited rather than restated.
- Produces: nothing later tasks consume.

- [x] **Step 1: Run 1's preservation duty becomes a render duty**

In `finish-contract.md`, the invocation of the retired script becomes:

```bash unverified:the command ships in task 12; this is its call site
myflow record render -change <name> -kind ledger -repo <abs-repo-root>
```

Keep the citation to `pipeline.md`'s outcome table rather than restating it — that citation is the
property that let task 14 change the table in one place.

- [x] **Step 2: `myflow-finish/SKILL.md`'s `finish.preserve-sessions` stage**

The stage key, its position and its mark are unchanged. Only the command it runs changes. State
explicitly, in one sentence, that the key is deliberately kept — otherwise the next reader renames it
and invalidates every stage run already recorded under it.

- [x] **Step 3: The guard-presence list**

Drop the retired script from `myflow-finish/SKILL.md`'s named guard set.

- [x] **Step 4: Guards green**

```bash verified:both commands are in .myflow/project.md's ## lint list
scripts/check-references.sh && scripts/check-contract-budget.sh
```

**Tests:** none — contract prose, covered by the guards above and task 19's full run.

**Regression:** Reverting this task leaves run 1 invoking a deleted script, so every integration
fails at the preservation step.

**Baseline:** both files' `check-contract-budget.sh` figures recorded in the commit body.

**Commit:** `docs(finish-contract): render the ledger at run 1 instead of copying it`

---

### 22 Close the retirement's loose ends

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/finish-contract.md` — the proposal-artifact copy
- Modify: `skills/myflow-contracts/pipeline.md` — the `Proposal artifact source` registry row
- Modify: `skills/myflow-finish/SKILL-rationale.md`
- Modify: `skills/myflow-contracts/pipeline-rationale.md`
- Modify: `stats/cmd/myflow/record.go` — one stale comment
- Modify: the eleven `scripts/` files whose comments cite the retired script

**Interfaces:**
- Consumes: tasks 13–15's retirement.
- Produces: a tree in which nothing cites a deleted file, and no registry row can never fire.

**Why this task exists.** The retired script copied **three** sources; `myflow record render` replaces
**two**. Found by tasks 13–15's implementer, who stated the gap rather than inventing a replacement.
<!-- measured: reported against commit 8e773bd by the implementer of tasks 13-15 -->

- [x] **Step 1: Run 1 copies the proposal artifact source**

`/myflow-start` writes the published proposal's HTML beside the state file so a revision round can
republish to the same URL, and `pipeline.md`'s registry deletes that source at run 2 **only if a
preserved copy exists**. Nothing produces that copy any more, so the row can never fire: run 2 leaves
the source behind forever, and no durable copy of the approved proposal reaches the repository.

Run 1 copies it directly, stated in `finish-contract.md` beside the render — not by a new script,
which would reintroduce what tasks 13–15 retired, and not through `myflow record render`, which
renders **from the store** and has no business copying a file that was never in it.

```bash unverified:the call site is fixed by this plan; the contract does not state it yet
cp "<state-dir>/<name>-proposal-artifact.html" \
   "<repo>/docs/superpowers/artifacts/<YYYY-MM-DD>-<name>.html"
```

The change name is validated against `^[A-Za-z0-9][A-Za-z0-9._-]*$` **before** the path is built, for
the reason `records.Destination` carries the same rule: a glob metacharacter in a name once matched
and overwrote a *different* change's preserved record. A change with no published artifact — every
`/myflow-fast` run — is skipped and said so, never treated as a failure.

- [x] **Step 2: The registry row states what performs it**

`pipeline.md`'s `Proposal artifact source` row names run 1's copy as what produces the preserved copy
its own condition tests, so the row's condition is reachable.

- [x] **Step 3: Repoint the citations at `records.Destination`**

Eleven `scripts/` files carry comments citing the retired script's header as the canonical statement
of the change-name allowlist. The rule did not disappear — it moved into `records.Destination`
(`stats/internal/records/render.go`). Repoint each citation there. Do **not** delete the reasoning;
the allowlist's history is why it exists.
<!-- measured: grep -rn 'preserve-session-records' scripts/ @ commit 8e773bd -->

- [x] **Step 4: Two rationale files and one Go comment**

`skills/myflow-finish/SKILL-rationale.md` says the script "still runs before the first `add`,
unchanged" — false now. `skills/myflow-contracts/pipeline-rationale.md`'s "Preserving the session
records" section **keeps its heading**, because `check-references.sh` matches citing bold tokens
against it and `pipeline.md` cites it inline by that former name; its ordering-asymmetry paragraph is
stale prose about a script that no longer exists and is rewritten. **No grep finds this one** — it
never names the script literally; tasks 13–15's reviewer found it by reading, which is the only way it
is findable, and is why it is written out here rather than left to a guard. `stats/cmd/myflow/record.go`'s
`resolveRenderKinds` comment says `all` is "finish run 1", but run 1 renders `-kind ledger` only —
the panel rendered at panel close.

- [x] **Step 5: The grep is clean**

```bash verified:grep is a standard tool; the paths are this repository's own
grep -rn 'preserve-session-records' skills/ rules/ commands/ commands-claude/ scripts/ setup.sh .myflow/ README.md AGENTS.md CLAUDE.md || echo "clean"
```

Expected: every path clean **except** `skills/myflow-do/SKILL.md`, which is task 17's file and still
carries two hits at this point in the order. Task 19 step 4 runs the same grep, after task 17, and
expects `clean` outright.

**Tests:** none added — contract prose and comments, covered by the guards in step 5 and task 19.

**Regression:** Reverting this task leaves a registry row that can never fire, a published proposal
that reaches the repository nowhere, and eleven files citing a deleted script as canonical.

**Baseline:** every guard in `.myflow/project.md`'s `## lint` list exits 0 after this task, except any
budget row task 19 raises.

**Commit:** `docs(finish-contract): copy the proposal artifact at run 1 and repoint the retired citations`

---

### 16 `handoff-blocks.md` and the `Records:` line

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/handoff-blocks.md`
- Modify: `skills/myflow-fast/SKILL.md` — the `IN_PROGRESS`-with-no-artifact block

**Interfaces:**
- Consumes: task 6's journal path, which is what the line counts.
- Produces: the block shape task 17's handoff renders.

- [x] **Step 1: Add the line to the `IN_PROGRESS` block**

```text unverified:the line does not exist yet; this is the shape this task adds
**Records:** all writes reached the store | N write(s) journalled — the store was unreachable | unknown — the journal could not be counted
```

The third alternative is not optional. Task 24's `journal-count` prints `unknown` when it cannot
produce a number, and a value the command is required to produce must have a rendering — otherwise the
contract has no shape for it and an agent invents one.

State that the line is present **whether or not** any write was journalled: its absence must never
be mistakable for a clean run, which is exactly the silence this change exists to remove.

- [x] **Step 2: The composite command's own variant**

`myflow-fast/SKILL.md` prints the one `IN_PROGRESS` shape no other skill prints, so it carries its
own copy of the block and needs the line added there too.

- [x] **Step 3: Guards green**

```bash verified:both commands are in .myflow/project.md's ## lint list
scripts/check-references.sh && scripts/check-contract-budget.sh
```

**Tests:** none — contract prose, covered by the guards above and task 19's full run.

**Regression:** Reverting this task restores a never-block guarantee that is also silent, so a
journalled record write leaves the operator no trace at the gate they read.

**Baseline:** both files' `check-contract-budget.sh` figures recorded in the commit body.

**Commit:** `docs(handoff-blocks): report journalled record writes at the IN_PROGRESS gate`

---

### 17 `myflow-do/SKILL.md` sections 4, 5 and 7

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-do/SKILL-rationale.md` — it carries the last `check-references.sh` hit
  (line 251, citing `pipeline.md`'s renamed section) and no other task owns it.

**Interfaces:**
- Consumes: tasks 6 and 12's commands; task 16's block shape.
- Produces: nothing later tasks consume.

The largest prose edit in the change, and the one that actually stops the double-write.

- [x] **Step 1: Section 4 — the ledger**

Remove the paragraph stating that `superpowers:subagent-driven-development`'s workspace script writes
the ledger to `.superpowers/sdd/tasks/progress.md`, and the warning about the flat path. In its
place: the parent records each dispatch as it closes, with

```bash unverified:the command ships in task 6; this is its call site
myflow record dispatch -change <name> -task <n> -role implementer -model <m> \
  -commit <sha> -outcome completed -session-token mf-<literal-token> -started-at <ts>
```

State that the model named here is the recorded intent required by **Model policy**, and that a slot
whose model the dispatcher cannot read records `unknown (agent-defined)` and never a guess.

- [x] **Step 2: Section 5 — the panel record**

The findings table and both marker blocks stop being hand-written. Each finding is recorded with
`myflow record finding` as its slot raises it; a fix round updates it with `myflow record status`.
The record is rendered at panel close, before `check-unfinished-work.sh` runs:

```bash unverified:the command ships in task 12; this is its call site
myflow record render -change <name> -kind panel -repo <abs-worktree>
```

A panel that raised nothing still renders, declaring `findings-total: 0` — see task 12 step 4a. Do not
add a "skip the render when there were no findings" shortcut here; that is the exact shape of the
defect step 4a exists to close.

Keep every existing rule about the record's **format** exactly as written — the marker anchoring, the
unbroken span, the separate reproducer block, the label-collision prohibition. They now bind the
renderer rather than the agent, and the guard is unchanged. Say that explicitly, so the next reader
does not delete a format rule on the grounds that nobody types it any more.

- [x] **Step 3: Section 7 — the ledger assert and the handoff**

The `test -f <abs-worktree>/.superpowers/sdd/tasks/progress.md` check becomes a store query through
`myflow record render -kind ledger`'s own `MISSING:` word, reported at its call site and still never
gating. Add the `Records:` line to the section's handoff block, matching task 16's shape.

- [x] **Step 4: The `prUrl` push path and the guard-presence list**

The preservation call on `/myflow-do`'s push path becomes a render call. Drop the retired script from
the named guard set.

- [x] **Step 5: Guards green**

```bash verified:all three commands are in .myflow/project.md's ## lint list
scripts/check-references.sh && scripts/check-contract-budget.sh && scripts/check-stage-mark-calls.sh
```

**Tests:** none — contract prose, covered by the guards above and task 19's full run.

**Regression:** Reverting this task restores the hand-written ledger and panel record alongside the
store rows, so two sources exist again and can disagree — the defect the whole change removes.

**Baseline:** `myflow-do/SKILL.md`'s `check-contract-budget.sh` figure recorded in the commit body.

**Commit:** `docs(myflow-do): record dispatches and findings in the store instead of files`

---

### 24 `myflow record journal-count`, so the handoff line can be produced

**Build:** green

**Files:**
- Modify: `stats/cmd/myflow/record.go`
- Modify: `stats/cmd/myflow/main.go` — the usage block
- Modify: `stats/cmd/myflow/record_test.go`
- Modify: `skills/myflow-contracts/handoff-blocks.md`
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `skills/myflow-fast/SKILL.md`

**Interfaces:**
- Consumes: task 6's `recordJournalPath`.
- Produces: `myflow record journal-count -change <name> [-C <dir>]`, printing one decimal count.

**Why this task exists.** Task 16 requires the `IN_PROGRESS` handoff to name how many record writes are
still journalled, and nothing exposes that number. `recordJournalPath` is an internal Go function, so
an agent would have to hand-derive `<state-dir>/<project-key>/<name>.journal.record` — which is the
exact recipe **State file** (`skills/myflow-contracts/state-file.md`) says was moved into the CLI so
that skills stop running it by hand. Found by tasks 16–17's reviewer.
<!-- measured: reported against commit 891fc64 by the reviewer of tasks 16-17 -->

- [x] **Step 1: The subcommand**

Counts complete entries in the record journal for the resolved project and change. An absent journal
is `0`, not an error — the ordinary case is that every write reached the store. Reuse
`splitCompleteLines`' rule for a partial trailing line rather than counting bare newlines: a
half-written final line is not an entry.

- [x] **Step 2: It never blocks either**

A count that cannot be produced prints `unknown` and exits 0. This command exists to make a handoff
line honest; it must not become the reason a handoff does not print.

- [x] **Step 3: The three contracts call it**

`handoff-blocks.md` names the command as the line's source rather than describing a path, and
`myflow-do` and `myflow-fast` follow it. Remove the hand-derived path from all three.

- [x] **Step 4: Correct the journal-sibling sentence**

`handoff-blocks.md` calls `<name>.journal.record` "the third journal beside the fallback file and the
state journal", which omits `.journal.stage` and coins "the state journal" for a file
`state-file.md` calls simply "the journal". `internal/reconcile` names exactly three: `.journal`,
`.journal.stage`, `.journal.record`. Say that.

- [x] **Step 5: Tests**

An absent journal counts `0`; a journal with two complete entries and a partial third counts `2`; an
unreadable journal prints `unknown` and exits 0.

```bash verified:go test is .myflow/project.md's ## test command for this package
cd stats && gofmt -w . && go test ./cmd/myflow -count=1 && gofmt -l .
```

**Tests:** the three cases in step 5, in `stats/cmd/myflow/record_test.go`.

**Regression:** Reverting this task leaves task 16's handoff line unproducible except by a recipe the
state-file contract withdrew, so agents would either hand-derive a path or omit the line — and an
omitted line is indistinguishable from a clean run, which is what task 16 exists to prevent.

**Baseline:** `./cmd/myflow` before=52 after=55.

**Commit:** `feat(myflow): report the journalled record-write count`

---

### 23 The guards read the panel record where it is now rendered

**Build:** green

**Files:**
- Modify: `scripts/check-unfinished-work.sh`
- Modify: `scripts/check-panel-reproducers.sh`
- Modify: `scripts/test-check-unfinished-work.sh`
- Modify: `scripts/test-check-panel-reproducers.sh`
- Modify: `stats/internal/records/render_test.go`
- Modify: `skills/myflow-fast/SKILL.md`
- Modify: `openspec/changes/kan-258-store-native-run-record/specs/myflow-review-panel-economics/spec.md`

**Interfaces:**
- Consumes: task 12's `records.Destination`, task 17's panel-close render.
- Produces: a tree in which the record the guards read is the record something writes.

**Why this task exists — it is a blocker, not a tidy-up.** The renderer writes the panel record to
`<repo>/docs/superpowers/reviews/<date>-<change>-panel.md`; both guards read
`<worktree>/.superpowers/sdd/final-review-panel.md`. After task 17 nothing writes a findings table or
a marker block into that file — it survives only as the pass log, which carries no `findings-total:`
line. So `check-unfinished-work.sh` would report **every** change OUTSTANDING at finish run 1, and
`check-panel-reproducers.sh` would fail **every** fix round. Found by task 17's implementer, who
declined to invent a bridge.
<!-- measured: reported against commit 891fc64 by the implementer of tasks 16-17 -->

**The ruling: one file, at the rendered destination.** Writing both paths would reintroduce exactly
the two-locations-that-can-disagree problem this whole change exists to remove, and
`docs/superpowers/reviews/` is where finish run 1 commits it from anyway.

- [x] **Step 1: Both guards resolve the rendered panel record**

`check-unfinished-work.sh` already takes the worktree and the change name; **`check-panel-reproducers.sh`
takes no change name at all** and gains one as a required second argument, carrying its sibling's
change-name allowlist byte for byte — the name now reaches a path and arrives from a PR-editable state
file. That forces one edit outside this task's file list, to `skills/myflow-do/SKILL.md` section 5,
the guard's only call site: leaving it would make the documented invocation exit 2. Resolve
`<worktree>/docs/superpowers/reviews/*-<change>-panel.md` with the **anchored** search — the same rule
`records.existingDatedFile` applies, and for the same reason: an unanchored glob once let one change
match and overwrite another's record. A change name is validated before it reaches a glob.

The **pass log stays where it is.** `<worktree>/.superpowers/sdd/final-review-panel.md` keeps the mode,
the slots, the diff path, the `fix-mutation:` lines and the bounces — it is worktree-lifetime by the
registry and is not a findings record. Do not merge the two.

- [x] **Step 2: Both harnesses assert the new location**

And assert the old one is **not** read: a guard that silently fell back to the sdd path would pass
these harnesses while reading a file nothing writes.

- [x] **Step 3: `render_test.go` stops masking the mismatch**

Its guard test copies the rendering into the sdd path, which is why the mismatch survived task 12's
review. Render to the real destination and let the guard find it there.

- [x] **Step 4: Amend the delta spec**

`specs/myflow-review-panel-economics/spec.md`'s **The rendered record satisfies the marker contract
the guard reads** says the guard "SHALL require no change". The marker **contract** is unchanged — the
guard's **path** is not. State both: the format the guard parses is untouched, and the location it
reads moves to where the record is now written. A requirement that contradicts the shipped code is the
spec's defect, not the code's.

- [x] **Step 5: `myflow-fast/SKILL.md`'s stale mechanism**

It says `finish.preserve-sessions` "copies the panel record and the SDD ledger into
`<project>/docs/superpowers/`". The stage key is deliberately unchanged, but the panel record is now
rendered by `/myflow-do` at panel close and the ledger by run 1. The sentence's mechanism is stale; it
does not trip the grep because it names the stage key rather than the script.

- [x] **Step 6: Guards and harnesses green**

```bash verified:both harnesses are named in .myflow/project.md's ## test list
scripts/test-check-unfinished-work.sh && scripts/test-check-panel-reproducers.sh \
  && scripts/check-references.sh && scripts/check-contract-budget.sh
cd stats && go test ./internal/records ./cmd/myflow -count=1
```

**Tests:** the cases in steps 2 and 3, in the two harnesses and `render_test.go`.

**Regression:** Reverting this task makes every finish run 1 report OUTSTANDING and every fix round
fail its reproducer guard — the change would be unusable end to end while every unit test stayed
green.

**Baseline:** `test-check-unfinished-work.sh` 128 → 136 `ok:` cases, `test-check-panel-reproducers.sh`
58 → 67; `./internal/records` unchanged at 21.
<!-- measured: reported against commit 2024800 by the implementer of task 23 -->

**Commit:** `fix(check-unfinished-work): read the panel record where it is now rendered`

---

### 18 `check-stage-mark-calls.sh` covers `myflow record` call sites

**Build:** green

**Files:**
- Modify: `scripts/check-stage-mark-calls.sh`
- Modify: `scripts/test-check-stage-mark-calls.sh`

**Interfaces:**
- Consumes: task 17's call sites, which are what the guard now scans.
- Produces: the guard verdict task 19's full run depends on.

The guard already rejects a hardcoded `-harness` literal and a `-session-token` carrying a
substitution in skill source. `myflow record dispatch` takes a session token too, so the same two
rules must reach it — otherwise the one mechanism that makes transcript binding work is enforced on
one verb and not the other.

- [x] **Step 1: Extend the scan to `myflow record` call sites**

- [x] **Step 2: Add the harness cases**

One case per rule, in the existing harness's idiom: a `myflow record dispatch` line carrying
`-session-token "mf-$(date +%s)"` is rejected and named; one carrying a literal passes.

- [x] **Step 3: Run the harness**

```bash verified:both files are named in .myflow/project.md's ## test and ## lint lists
scripts/test-check-stage-mark-calls.sh && scripts/check-stage-mark-calls.sh
```

Expected: exit 0 from both.

**Tests:** the two cases in step 2, in `scripts/test-check-stage-mark-calls.sh`.

**Regression:** Reverting this task lets a `myflow record` call site carry a shell substitution as its
session token, which lands in every transcript as the same unexpanded string and discriminates
between no two sessions — silently breaking the binding the record depends on.

**Baseline:** `scripts/test-check-stage-mark-calls.sh` before and after `ok:` counts recorded in the
commit body; the harness's own final line is the verdict.

**Commit:** `feat(check-stage-mark-calls): cover the record verb's session token`

---

### 19 Budget rows and a full guard run

**Build:** green

**Files:**
- Modify: `scripts/check-contract-budget.sh` — the `budgets()` table

**Interfaces:**
- Consumes: every earlier task's edits.
- Produces: a green lint run.

- [x] **Step 1: Raise the budget row for every contract file this change grew**

`pipeline.md`, `finish-contract.md`, `handoff-blocks.md`, `myflow-do/SKILL.md`,
`myflow-finish/SKILL.md`, `myflow-fast/SKILL.md`. Set each to its new size plus 25%, the rule the
guard's own table already follows. **Do not** narrow the guard's scope or delete a row; raising a
budget is the correct response to a genuine addition.

- [x] **Step 1a: Build the SPA before anything vets**

`internal/web/embed.go` carries `//go:embed all:dist`, so `go vet ./...` and `go build ./...` fail in
any worktree where `stats/internal/web/dist` was never built. This is not a defect in this change; it is what
a fresh worktree looks like.
<!-- measured: reported by task 1+2's implementer against this worktree @ commit 01bf42a -->

```bash verified:make build is named in .myflow/project.md's ## run section as what builds the SPA then verifies the Go build
cd stats && make build
```

- [x] **Step 2: Run the whole lint list**

```bash verified:this is .myflow/project.md's ## lint list, run as three invocations per its own runtime note
scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-plan-provenance.sh \
  && scripts/check-task-build-green.sh && scripts/check-workspace-isolation.sh \
  && scripts/check-uitest-overrides.sh && scripts/check-contract-budget.sh \
  && scripts/check-markdown-integrity.py && scripts/check-stage-mark-calls.sh \
  && scripts/check-guard-symlinks.sh && scripts/check-self-review-report.sh \
  && scripts/check-installed-citations.sh && scripts/check-installed-rules.sh
```

```bash verified:this is .myflow/project.md's ## lint list, Go and SPA halves
cd stats && gofmt -w . && gofmt -l . && go vet ./...
```

```bash verified:this is .myflow/project.md's ## lint list, SPA half
cd stats/web && npx tsc -b
```

**A lint-only check is not enough, and this change has the evidence.** `test-check-cleanup-complete.sh`
was left failing by task 14's registry edit and survived that task's own review, because the harness
lives in `## test` and the review ran `## lint`. Every guard that *declares* something about a
contract has a harness, and the harness is the only thing that reads the declaration.
<!-- measured: reported against commit 8e773bd by task 22's implementer -->

- [x] **Step 3: Run the whole test list**

Split across three invocations, per `.myflow/project.md`'s own runtime note — roughly 144s total
against a 120000ms default tool timeout, so one invocation reliably times out.
<!-- measured: .myflow/project.md's ## test section, "Measured runtime" note @ branch main -->

```bash verified:this is .myflow/project.md's ## test list, guard half
for t in scripts/test-*.sh; do "$t" || exit 1; done
```

```bash verified:this is .myflow/project.md's ## test list, Go half
cd stats && go test ./... -race -count=1
```

```bash verified:this is .myflow/project.md's ## test list, SPA half
cd stats/web && npm test
```

- [x] **Step 4: Confirm nothing names the retired script**

```bash verified:grep is a standard tool; the paths are this repository's own
grep -rn 'preserve-session-records' skills/ rules/ commands/ commands-claude/ scripts/ setup.sh .myflow/ || echo "clean"
```

Expected: `clean`. A hit here is a missed edit in tasks 14–17, fixed there rather than suppressed.

**Tests:** none added — this task runs the existing suites.

**Regression:** Reverting this task leaves at least one contract file over its budget, so the lint run
fails and the change cannot land.

**Baseline:** every guard exits 0; `cd stats && go test ./... -race -count=1` before=393 after=459
top-level tests, all passing. The `after` figure was written before tasks 20-24 added their own
tests and read 411; measured at the end of task 19 it is 459.

**Commit:** `chore(scripts): raise contract budgets for the store-native run record`
