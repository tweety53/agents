# kan-472-flow-dynamic-review-panel-roster-repo-scoped

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Twenty-three tasks in eight groups, dependency order; `design.md` is canonical for every rule, table
and shape a task implements — a task names the section it implements and what the edit must say,
never a second copy of it. Group 8 was appended after groups 1–7 had landed (KAN-478 joined,
`design.md`'s `kan-478-joined-eighth-group`); it carries its own measured baseline.

**Baseline, measured before any edit:**

- `cd stats && go test ./... -count=1` passes 646 top-level tests.
  <!-- measured: go test ./... -count=1 -json | jq -r 'select(.Action=="pass" and .Test!=null and (.Test|contains("/")|not)) | .Test' | wc -l @ branch spectre/kan-472-flow-dynamic-review-panel-roster-repo-scoped -->
- `cd stats/web && npm test` passes 157 tests in 9 files.
  <!-- measured: npx vitest run @ branch spectre/kan-472-flow-dynamic-review-panel-roster-repo-scoped -->
- `scripts/run-guard-tests.sh` discovers 55 `scripts/test-*.sh` harnesses.
  <!-- measured: ls scripts/test-*.sh | wc -l @ branch spectre/kan-472-flow-dynamic-review-panel-roster-repo-scoped -->
- 18 migrations are embedded, the last `0018_guard_log.sql`.
  <!-- measured: ls stats/internal/store/migrations | wc -l @ branch spectre/kan-472-flow-dynamic-review-panel-roster-repo-scoped -->
- Byte sizes against `scripts/check-contract-budget.sh`'s `budgets()` rows: `skills/flow/SKILL.md`
  14,658 of 16,278; `skills/flow/brainstorm.md` 12,067 of 35,015; `skills/flow/brainstorm-planner.md`
  21,102 of 23,248; `skills/flow/implement.md` 33,422 of 41,777; `skills/flow/review-panel.md`
  50,979 of 58,732; `skills/flow/verify-and-handoff.md` 31,780 of 33,199;
  `skills/flow-contracts/project-configuration.md` 47,438 of 48,175;
  `skills/flow-contracts/handoff-blocks.md` 16,387 of 20,240; `.flow/project.md` 25,647 of 26,450.
  <!-- measured: wc -c on each file and the budgets() rows at scripts/check-contract-budget.sh:146-224 @ branch spectre/kan-472-flow-dynamic-review-panel-roster-repo-scoped -->
- Line anchors on the clean branch. `stats/cmd/flow/record.go`: 50–66 `validateFindingStatus`,
  218–247 `runRecord`'s verb switch, 528–616 `runRecordDispatchBegin`. `stats/internal/store/records.go`:
  343–398 `UpsertFinding`, 399–413 `SetFindingStatus`. `stats/internal/api/records.go`: 23–28
  `RecordWriter`, 39–54 `RecordStore`, 288–311 `recordFinding`. `stats/internal/api/server.go`:
  294–314 the route table. `stats/internal/reconcile/reconcile.go`: 694–706 the replay switch.
  `stats/internal/api/stats.go`: 45–70 `StatsStore`, 87–103 the `viewName` constants and
  `knownViews`, 444–501 `rowsFor`. `stats/internal/store/aggregate.go`: 199–257 `StageLeaderboard`.
  `stats/internal/stages/names.go`: 65 the `flow.writing-plans` row. `README.md`: 125 the same row.
  `stats/web/src/api.ts`: 40–42 `ViewName`/`VIEW_NAMES`, 251–259 `StageLeaderboardRow`.
  `stats/web/src/App.tsx`: 26–38 `VIEW_LABELS`/`VIEW_COMPONENTS`. `skills/flow/SKILL.md`: 47–54 the
  stage-key table, 56–80 **Model resolution**'s block, 191–218 **Guardrails**.
  `skills/flow/brainstorm.md`: 197–206 "Once `## Plan` returns". `skills/flow/brainstorm-planner.md`:
  249–338 section **D**. `skills/flow/implement.md`: 19–104 **Dispatch the conductor** (70–81 the
  handshake), 243–530 section **4**, 495–512 **Turn discipline**. `skills/flow/review-panel.md`:
  113–135 **The roster**, 258–283 the dispatch record, 543–625 **Panel re-runs** (567–571 the Minor
  rule), 776–796 the fix subagent dispatch, 798–807 the handback.
  `skills/flow/verify-and-handoff.md`: 116 the verifier handshake.
  `skills/flow-contracts/handoff-blocks.md`: 100–135 the `IN_PROGRESS` block.
  `skills/flow-contracts/project-configuration.md`: 33–36 the single-literal key rows, 87–96 the
  match rule. `scripts/check-unfinished-work.sh`: 338; `scripts/check-panel-findings-closed.sh`: 80
  (the two `jq` open-finding selects). `scripts/check-dispatch-paragraphs.sh`: 155–200 the entry and
  site tables. `scripts/check-model-resolution-shell.sh`: 126–143 its cases.
  <!-- measured: grep -n / sed -n on each file @ branch spectre/kan-472-flow-dynamic-review-panel-roster-repo-scoped -->

**Every task that grows an owned `.md` file runs `scripts/check-contract-budget.sh` in its verify
step**; a trip is answered by raising that file's row to its new byte size plus 25%, in the same
commit, with `scripts/check-contract-budget.sh` in `**Allowed-collateral:**`. Task 6 is expected
to trip on `project-configuration.md`; every other trip is possible, none is planned.

**Every Go task's verify step is** `cd stats && gofmt -l . && go vet ./...` **plus its own
`-run`**; every SPA task's is `cd stats/web && npx tsc -b` plus its own `-t`; every prose task's is
the guard list its own verify step names.

## Group 1 — store and CLI

- [x] 1. `decisions` table, store methods and the `records.Decision` type

Per `design.md` **Store › `decisions`**. Add `stats/internal/store/migrations/0019_decisions.sql`
carrying exactly the DDL that section shows, with a header comment stating why one JSONB row
rather than columns (`design.md`'s `decisions-jsonb-row`). Add to `stats/internal/records/types.go`:

```go verified:authored in-tree for this change; field shapes mirror Dispatch/Finding in the same file
// Decision is one run's dynamic decision: the whole `## Decision` block as JSON.
type Decision struct {
	ID           int64           `json:"id"`
	SessionToken string          `json:"sessionToken"`
	RecordedAt   time.Time       `json:"recordedAt"`
	Decision     json.RawMessage `json:"decision"`
}
```

Add to `stats/internal/store/records.go`, after `SetFindingStatus`:
`RecordDecision(ctx, projectKey, change string, in records.Decision) (records.Decision, bool, error)` — the bool is `created`, from `xmax = 0` exactly as `UpsertFinding` reads it —
`INSERT … SELECT c.id … FROM changes c WHERE c.project_key = $1 AND c.name = $2 ON CONFLICT ON
CONSTRAINT decisions_session_key DO UPDATE SET decision = EXCLUDED.decision, recorded_at = now()
RETURNING id, session_token, recorded_at, decision` (`ErrChangeNotFound` on no row, the
`UpsertFinding` shape) — and `ListDecisions(ctx, projectKey, change string) ([]records.Decision,
error)`, newest first. Reject an empty `SessionToken` or an empty/invalid-JSON `Decision` with a
wrapped `ErrInvalidRecord`-style sentinel `ErrInvalidDecision` declared in `records.go`.

  - [x] **Step 1: RED** — in `stats/internal/store/records_test.go` add `TestRecordDecisionUpsertsPerSession`
    (record twice under one token: one row, second body wins) and `TestListDecisionsNewestFirst`
    (two tokens, order by `recorded_at DESC`); run `cd stats && go test ./internal/store -run
    'TestRecordDecision|TestListDecisions' -count=1 | tail -5`, both fail to compile.
  - [x] **Step 2: GREEN** — migration, type, methods; re-run, both pass. Confirm the
    embedded-count tests (`TestMigrationsApplyCleanly`-shaped, `records_test.go:116`,
    `store_test.go:29`) still pass: `go test ./internal/store -count=1 | tail -3`.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l . && go vet ./... && go test ./internal/store
    -run 'TestRecordDecision|TestListDecisions' -count=1`.

**Files:** `stats/internal/store/migrations/0019_decisions.sql`, `stats/internal/records/types.go`,
`stats/internal/store/records.go`, `stats/internal/store/records_test.go`
**Tests:** `TestRecordDecisionUpsertsPerSession`, `TestListDecisionsNewestFirst`
**Regression:** reverting this commit removes the table; both tests fail on a missing relation.
**Baseline:** before=646 after=648
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(stats): add the decisions table and store methods`
**Build:** green
**After:** none

- [x] 2. Decision routes, client methods and journal replay

`stats/internal/api/records.go`: add `RecordDecision` and `ListDecisions` to `RecordWriter`
(the replay path needs the write) and `RecordStore`; an `ApplyDecisionRecord(ctx, w RecordWriter,
project, change string, in records.Decision) (records.Decision, error)` beside `ApplyFindingRecord`
that maps an empty token or body to `ErrInvalidRecord`; handlers `recordDecision` (`201` created /
`200` replaced, from task 1's `created` bool) and
`listDecisions`. `stats/internal/api/server.go` route table: `POST
/api/v1/records/{project}/{change}/decisions` and `GET /api/v1/records/{project}/{change}/decisions`.
`stats/internal/client/client.go`: `RecordDecision` and `ListDecisions` through `writeRecord`/`send`,
the `RecordFinding`/`GetRunRecord` shapes. `stats/internal/reconcile/reconcile.go` replay switch:
`case "decision":` decoding `records.Decision` and calling `ApplyDecisionRecord`; extend the
kind list in the comment at 595–597.

  - [x] **Step 1: RED** — `stats/internal/api/records_test.go`: `TestRecordDecisionRouteCreatesThenReplaces`,
    `TestListDecisionsRouteNewestFirst`, `TestRecordDecisionRejectsEmptyBody` (400);
    `stats/internal/client/records_test.go`: `TestClientRecordDecisionRoundTrip`;
    `stats/internal/reconcile/record_test.go` (create the file if `reconcile/` holds no
    `record_test.go` — it does not on the clean branch; put the case in `reconcile_test.go`
    instead): `TestReplayDecisionKind`. Run each package's `-run` selector; all fail.
  - [x] **Step 2: GREEN** — implement; re-run, all pass.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l . && go vet ./... && go test ./internal/api
    ./internal/client ./internal/reconcile -run 'Decision' -count=1`.

**Files:** `stats/internal/api/records.go`, `stats/internal/api/server.go`,
`stats/internal/client/client.go`, `stats/internal/reconcile/reconcile.go`,
`stats/internal/api/records_test.go`, `stats/internal/client/records_test.go`,
`stats/internal/reconcile/record_test.go`
**Allowed-collateral:** `stats/internal/store/records.go` (plan-predicted path corrected: the
`reconcile_test.go` the plan named does not carry this case — `reconcile/record_test.go` already
existed on the clean branch (KAN-451) and is where `TestReplayDecisionKind` actually lives, per
the task's own conditional); `stats/internal/api/changes_test.go`, `stats/internal/client/client_test.go`,
`stats/internal/web/embed_test.go` — compile-time `api.RecordStore`/`RecordWriter` interface
satisfiers that needed matching stub methods once the interface grew two methods, unavoidable
fallout, not a design choice.
**Tests:** `TestRecordDecisionRouteCreatesThenReplaces`, `TestListDecisionsRouteNewestFirst`,
`TestRecordDecisionRejectsEmptyBody`, `TestClientRecordDecisionRoundTrip`, `TestReplayDecisionKind`
**Regression:** reverting this commit leaves the table unreachable over HTTP; every test above
fails on a 404 route or an unknown replay kind.
**Baseline:** before=648 after=653
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(stats): serve and replay decision records`
**Build:** green
**After:** Task 1

- [x] 3. `flow record decision` and `flow record decisions`

`stats/cmd/flow/record.go` verb switch (218–247): `case "decision":` → `runRecordDecision` with
flags `-change`, `-session-token`, `-file <path>` (the JSON body, read whole; `-` for stdin), plus
the shared `-addr`/`-timeout`/`-C`; validates the token non-empty and the body as JSON before any
network call (exit 2 on either, the `validateFindingStatus` posture); on an unreachable store
journals the entry with kind `decision` via `fallback.AppendJournalEntry` and exits 0 printing the
one warning line every other record verb prints; prints `recorded: decision <id>` on success.
`case "decisions":` → `runRecordDecisions` with `-change` and `-C`, printing `ListDecisions` as a
JSON array (what a resumed run reads; empty array when none). Extend `recordUsage` (line 89) and
`runRecordJournalCount` so a journalled `decision` counts in the handoff's `Records:` line.

  - [x] **Step 1: RED** — `stats/cmd/flow/record_test.go`: `TestRunRecordDecisionRefusesBadBody`
    (exit 2, no request), `TestRunRecordDecisionJournalsWhenUnreachable` (exit 0, one journal entry
    of kind `decision`), `TestRunRecordDecisionsPrintsJSON`. Run `cd stats && go test ./cmd/flow -run
    'TestRunRecordDecision' -count=1 | tail -5`; all fail.
  - [x] **Step 2: GREEN** — implement; re-run, pass.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l . && go vet ./... && go test ./cmd/flow -run
    'TestRunRecordDecision|TestRunRecordJournalCount' -count=1`.

**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Allowed-collateral:** `stats/internal/fallback/journal.go`; `stats/cmd/flow/main.go` — usage-text
rows for the two new verbs, and threading `stdin` into `runRecord` for `-file -`.
**Tests:** `TestRunRecordDecisionRefusesBadBody`, `TestRunRecordDecisionJournalsWhenUnreachable`,
`TestRunRecordDecisionsPrintsJSON`
**Regression:** reverting this commit makes `flow record decision` an unknown verb; all three
tests fail on the usage exit.
**Baseline:** before=653 after=656
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(flow): add record decision and record decisions verbs`
**Build:** green
**After:** Task 2

- [x] 4. `dispatches.effort` and `-effort`

`stats/internal/store/migrations/0020_dispatch_effort.sql`: `ALTER TABLE dispatches ADD COLUMN
effort TEXT NOT NULL DEFAULT 'default';` with a header stating the vocabulary (`low`, `medium`,
`high`, `default`) and why `default` rather than NULL (an absent effort is a known fact — the
dispatcher set none — never an unknown one, `design.md` **Enforcement**). `records.Dispatch` gains
`Effort string \`json:"effort,omitempty"\``; `insertDispatch`, `RunRecord`'s scan and
`DispatchWindowsForSession` (if it selects columns by name) carry it; `runRecordDispatchBegin`
gains `-effort` validated against the four words (exit 2 otherwise), defaulting to `default`; the
ledger renderer (`RenderLedger`, `stats/internal/records/render.go:321`) prints `effort=<v>` after
the model on each dispatch line.

  - [x] **Step 1: RED** — `stats/internal/store/records_test.go`: `TestRecordDispatchStoresEffort`;
    `stats/cmd/flow/record_test.go`: `TestRunRecordDispatchBeginRejectsUnknownEffort`;
    `stats/internal/records/render_test.go`: `TestLedgerRendersEffort`. Run each `-run`; all fail.
  - [x] **Step 2: GREEN** — implement; re-run, pass. Run the full store and cmd packages once:
    `go test ./internal/store ./cmd/flow ./internal/records -count=1 | tail -5`.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l . && go vet ./... && go test ./internal/store
    ./cmd/flow ./internal/records -run 'Effort' -count=1`.

**Files:** `stats/internal/store/migrations/0020_dispatch_effort.sql`,
`stats/internal/records/types.go`, `stats/internal/records/render.go`,
`stats/internal/store/records.go`, `stats/cmd/flow/record.go`,
`stats/internal/store/records_test.go`, `stats/cmd/flow/record_test.go`,
`stats/internal/records/render_test.go`
**Tests:** `TestRecordDispatchStoresEffort`, `TestRunRecordDispatchBeginRejectsUnknownEffort`,
`TestLedgerRendersEffort`
**Regression:** reverting this commit drops the column; the store test fails on an unknown column,
the CLI test on an unknown flag, the render test on a missing `effort=`.
**Baseline:** before=656 after=659
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(stats): record dispatch effort`
**Build:** green
**After:** Task 3

- [x] 5. `deferred <reason>` finding status, Minor-only

Per `design.md` **Store › `deferred <reason>`**. `validateFindingStatus` (`record.go:55–66`)
accepts `deferred` carrying a non-empty reason, the exact `withdrawn` shape, and its error message
lists four forms. `SetFindingStatus` (`store/records.go:399`) — when `status` opens with
`deferred`, the `UPDATE` adds `AND f.severity ILIKE 'Minor'` (round-3 panel fix F11/F13:
case-insensitive, matching this store's own severity convention); zero rows affected then distinguishes
`ErrFindingNotFound` from a new `ErrDeferredNotMinor` by a second `SELECT severity`; the API maps
`ErrDeferredNotMinor` to 409 in `mapStoreError` (`server.go:461`) and the client treats 409 as a
refusal, never a journal candidate (the same reason 404 is not). `RenderPanel` already prints
whatever status the row holds; `render_test.go` gets the case. Both guards: change the `jq` select
at `check-unfinished-work.sh:338` and `check-panel-findings-closed.sh:80` to also exclude
`startswith("deferred")`, and update each guard's header comment (307, 335; 77) naming the third
closed form.

  - [x] **Step 1: RED** — `record_test.go`: `TestValidateFindingStatusDeferred` (bare `deferred`
    refused, `deferred x` accepted); `store/records_test.go`: `TestSetFindingStatusDeferredMinorOnly`
    (Major → `ErrDeferredNotMinor`, Minor → ok); `render_test.go`: `TestPanelRendersDeferred`;
    `scripts/test-check-unfinished-work.sh` and `scripts/test-check-panel-findings-closed.sh`: one
    case each, a `deferred x` finding counted closed. Run all; fail.
  - [x] **Step 2: GREEN** — implement; re-run, pass.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l . && go vet ./... && go test ./cmd/flow
    ./internal/store ./internal/records ./internal/api -run 'Deferred' -count=1`, then from the
    worktree root `scripts/test-check-unfinished-work.sh` and
    `scripts/test-check-panel-findings-closed.sh`.

**Files:** `stats/cmd/flow/record.go`, `stats/internal/store/records.go`,
`stats/internal/api/server.go`, `stats/internal/client/client.go`,
`stats/cmd/flow/record_test.go`, `stats/internal/store/records_test.go`,
`stats/internal/records/render_test.go`, `scripts/check-unfinished-work.sh`,
`scripts/check-panel-findings-closed.sh`, `scripts/test-check-unfinished-work.sh`,
`scripts/test-check-panel-findings-closed.sh`
**Tests:** `TestValidateFindingStatusDeferred`, `TestSetFindingStatusDeferredMinorOnly`,
`TestPanelRendersDeferred`, `scripts/test-check-unfinished-work.sh`,
`scripts/test-check-panel-findings-closed.sh`,
`TestSetFindingStatusDeferredAcceptsLowercaseMinor` (round-3 panel fix F11/F13: the deferred
guard's severity match was case-sensitive, contradicting this store's own case-insensitive
severity convention; fixed to a case-insensitive match, with this regression test pinning a
lowercase-severity Minor finding can now be deferred)
**Regression:** reverting this commit makes `deferred` an invalid status at the CLI and an open
finding to both guards; every test above fails.
**Baseline:** before=659 after=663
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(stats): add the deferred finding status, Minor only`
**Build:** green
**After:** Task 4

## Group 2 — toggles, classifier, decision tree

- [x] 6. The three toggle keys

`skills/flow-contracts/project-configuration.md`: three rows after the `## self review model` row
(line 36), one per key from `design.md` **Toggles**, each stating the two literals, absent =
`default`, and report-by-name-and-drop; one sentence at 92–96 extending the match rule to the three
keys. `skills/flow/SKILL.md` **Model resolution** block (60–80): three `project-get.sh` reads into
`EXECUTION_MODE_TOGGLE`, `IMPLEMENTER_MODEL_TOGGLE`, `REVIEW_PANEL_TOGGLE`, each `default` when
absent or when the body is neither literal (with the `⚠ flow: …` stderr line the block already
uses for a bad model), and a paragraph after the `VERIFY_MODEL` one stating what each toggle
governs, citing `design.md`'s section by name only. **Guardrails** (191–218): the first bullet's
"the roster is resolved from the settings store, never asked" gains "or from the recorded
decision when `## review panel` is `dynamic`", and the "never add a slot" bullet gains "a dynamic
roster is the decision's roster, and the docs-only reduction still only removes".
`scripts/check-model-resolution-shell.sh`: the extracted block must now also print the three
toggles; add one case at 126–143 with all three keys `dynamic` and one with `## review panel`
holding `sometimes`, expecting `default` and the warning. `.flow/project.md`: no toggle is
declared for this repository in this change — say so in one line under `## self review model`
(line 341–344), so the absence is recorded rather than an oversight.

  - [x] **Step 1: Edit** the four files as above.
  - [x] **Step 2: Verify** — `scripts/check-model-resolution-shell.sh`,
    `scripts/test-check-model-resolution-shell.sh`, `scripts/check-model-keys.sh`,
    `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`,
    `scripts/check-contract-budget.sh` — all exit 0. `project-configuration.md` is 737 bytes under
    its row and this task adds more than that: raise the row per the header rule.
    <!-- predicted: wc -c skills/flow-contracts/project-configuration.md after this task -->

**Files:** `skills/flow-contracts/project-configuration.md`, `skills/flow/SKILL.md`,
`scripts/check-model-resolution-shell.sh`, `.flow/project.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** none — prose plus a guard whose cases are inline; the verify step's guards are the check
**Regression:** reverting this commit leaves no toggle a run can read; tasks 7–15 resolve every
toggle as `default` and the dynamic paths never run.
**Baseline:** before=662 after=662 — no Go change; the guard harness count stays 55
<!-- predicted: ls scripts/test-*.sh | wc -l after this task -->
**Commit:** `feat(flow-contracts): add the execution mode, implementer model and review panel toggles`
**Build:** green
**After:** none

- [x] 7. `plan-class.sh` and the planner's `## Decision` step

`scripts/plan-class.sh <tasks.md> <repos>` — Bash, `set -euo pipefail`, shipped: symlinked into
`skills/flow/scripts/`. Prints exactly three lines in the shape `design.md`'s `## Decision` block
shows — `inputs: tasks=N files=N repos=N migration=yes|no spec=yes|no red=yes|no unverified=yes|no`,
`class: <small|regular|big>`, `rolls: compact N · experimental N` — from the rules under
**Inputs and the class** and **The rolls**: tasks = column-0 `- [ ] <n>.`/`- [x] <n>.` lines;
files = the union of backticked paths on `**Files:**` lines; migration/spec/red/unverified by the
patterns that section names; rolls from `basename "$(dirname "$1")"` as the change name, via
`shasum -a 256`, first 8 hex digits, `mod 100`. Exit 2 on a missing file or a non-integer
`<repos>`. `scripts/test-plan-class.sh`: fixtures for each class boundary (5/12 small, 6 tasks
regular, 15 tasks big, 1 migration + 8 tasks big, 1 migration + 7 tasks regular), the override
never appearing (the script has none), and a fixed name's two rolls asserted equal across two runs.
`skills/flow/brainstorm-planner.md` section **D** (249–338): after the guards paragraph (328–330),
a **Decide** subsection: run `plan-class.sh <changeRoot>/tasks.md <repos>`; apply the one-step
override rule; read the three toggles from the parent's dispatch prompt (the parent passes them —
task 8); walk the chain and the tree as `design.md` **The decision chain** states, by name; write
`<abs-worktree>/.superpowers/sdd/decision.json` in the shape `design.md` **Store › `decisions`**
shows; append the `## Decision` block, in `design.md`'s exact shape, as the last section of the
`## Plan` return. `skills/flow/SKILL.md` guard-presence list (165–170) gains `plan-class.sh`.
`scripts/check-guard-symlinks.sh` rule 2 finds the invocation in `brainstorm-planner.md` and
requires the symlink — create it in this task.

  - [x] **Step 1: RED** — write `scripts/test-plan-class.sh` with the cases above; run it; every
    case fails on a missing script.
  - [x] **Step 2: GREEN** — write `scripts/plan-class.sh`, `chmod +x`, `ln -s ../../../scripts/plan-class.sh
    skills/flow/scripts/plan-class.sh`; re-run the harness, all pass. Run it against this change's
    own `tasks.md` with `1` and confirm `class: big` (17 tasks).
    <!-- predicted: scripts/plan-class.sh spectre/changes/kan-472-flow-dynamic-review-panel-roster-repo-scoped/tasks.md 1 after this task -->
  - [x] **Step 3: Edit** `brainstorm-planner.md` and `SKILL.md` as above.
  - [x] **Step 4: Verify** — `scripts/test-plan-class.sh`, `scripts/check-guard-symlinks.sh`,
    `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`,
    `scripts/check-contract-budget.sh`; all exit 0 (`brainstorm-planner.md` has 2,146 bytes of
    room; raise the row if the subsection exceeds it).

**Files:** `scripts/plan-class.sh`, `scripts/test-plan-class.sh`, `skills/flow/scripts/plan-class.sh`,
`skills/flow/brainstorm-planner.md`, `skills/flow/SKILL.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** `scripts/test-plan-class.sh`
**Regression:** reverting this commit removes the classifier; the harness fails on a missing
script, and no planner can produce a `## Decision` block.
**Baseline:** before=662 after=662 — no Go change; harness count 55 → 56
<!-- predicted: ls scripts/test-*.sh | wc -l after this task -->
**Commit:** `feat(scripts): add plan-class.sh and the planner's decision step`
**Build:** green
**After:** Task 6

- [x] 8. The parent prints and records the decision

`skills/flow/brainstorm.md` "Dispatch the planner" (112–121): the planner prompt also states the
three resolved toggles and the resolved worktree count. "Once `## Plan` returns" (197–206): before
the `dispatch end`, print the `## Decision` block verbatim, then the `flow stage begin flow.decide`
/ `flow record decision` / `flow stage end` sequence `design.md` **The `## Decision` block** shows;
a `## Plan` carrying no block is reported as a planner defect and the run continues on `default`
for all three (recorded as such through the same command, with `"planner": "no block"` in the
JSON). Add a **Resume and fix runs** paragraph citing `design.md`'s section: **Resuming at
`STARTED`** (70–90) reads `flow record decisions -change <name>` and states the decision in force
or re-dispatches the planner's decide step alone when none exists; `flow.document-fix`
(`implement.md` 194–241) hands the appended plan through the same step. `skills/flow/SKILL.md`
stage-key table row for `brainstorm.md` (49) gains `flow.decide` after `flow.writing-plans`.
`stats/internal/stages/names.go` (after line 65) and `README.md` (after line 125): the
`flow.decide` row, name "Decide — execution, models, panel". `skills/flow/implement.md` **Dispatch
the conductor** (39–44): the conductor prompt also states the decision JSON path and the three
toggles.

  - [x] **Step 1: Edit** the five files as above.
  - [x] **Step 2: Verify** — `cd stats && go test ./internal/stages -run
    'TestStagesMatchReadmeLevelOne' -count=1`; from the worktree root
    `scripts/check-stage-mark-calls.sh`, `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`,
    `scripts/check-contract-budget.sh`; all exit 0.

**Files:** `skills/flow/brainstorm.md`, `skills/flow/SKILL.md`, `skills/flow/implement.md`,
`stats/internal/stages/names.go`, `README.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** none — the existing `TestStagesMatchReadmeLevelOne` pins the new row
**Regression:** reverting this commit leaves the block unprinted and unrecorded — `flow record
decisions` returns `[]` for every run and a resumed run has nothing to follow.
**Baseline:** before=662 after=662
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(flow): print and record the planner's decision`
**Build:** green
**After:** Task 3, 6, 7

## Group 3 — enforcement

- [x] 9. Nine agent definitions and their install

`agents/flow-<model>-<effort>.md` for sonnet/opus/haiku × low/medium/high, frontmatter and body
per `design.md` **Enforcement**:

```markdown unverified:confirm `effort:` is the frontmatter key Claude Code's agent definitions read for reasoning effort, and its accepted values, before writing the nine files
---
name: flow-sonnet-high
description: General-purpose flow role on sonnet at high effort; the dispatch prompt carries every instruction.
model: sonnet
effort: high
---
General-purpose flow role; the dispatch prompt carries every instruction.
```

`setup.sh`: an `install_agents <dest>` function beside `install_commands`, linking every
`agents/*.md` into `<dest>` through `link_into`; called from the global mode (near line 1053) for
`~/.claude/agents/` and from the project mode (near 227) for `.claude/agents/`; the two summary
banners name the new target. `scripts/test-setup-agents.sh`: a sandboxed `HOME=$(mktemp -d)
./setup.sh global` then asserts nine symlinks under `$HOME/.claude/agents/`, each resolving, each
frontmatter carrying `model:` and `effort:`. `scripts/check-installed-citations.sh` runs
`setup.sh` sandboxed already — confirm it still passes.

  - [x] **Step 1: Establish the frontmatter key** — dispatch nothing; read the harness's own agent
    definition documentation via the `claude-code-guide` agent or Context7 and record the answer
    in the REPORT FILE. If the key is not `effort`, use the documented one in all nine files and
    report the plan's `unverified:` block as resolved.
  - [x] **Step 2: RED** — write `scripts/test-setup-agents.sh`; run; fails on zero links.
  - [x] **Step 3: GREEN** — nine files, `install_agents`, both call sites; re-run, passes.
  - [x] **Step 4: Verify** — `scripts/test-setup-agents.sh`, `scripts/check-installed-citations.sh`,
    `scripts/check-vocabulary.sh`, `scripts/check-references.sh`; all exit 0.

**Files:** `agents/flow-sonnet-low.md`, `agents/flow-sonnet-medium.md`, `agents/flow-sonnet-high.md`,
`agents/flow-opus-low.md`, `agents/flow-opus-medium.md`, `agents/flow-opus-high.md`,
`agents/flow-haiku-low.md`, `agents/flow-haiku-medium.md`, `agents/flow-haiku-high.md`, `setup.sh`,
`scripts/test-setup-agents.sh`
**Allowed-collateral:** `scripts/check-installed-citations.sh` — the nine new agent files enter
the guard's own owned corpus; it needs a `declare_if_present` entry per file to keep passing.
**Tests:** `scripts/test-setup-agents.sh`
**Regression:** reverting this commit leaves no `flow-<model>-<effort>` type to dispatch by; the
harness fails on zero links.
**Baseline:** before=662 after=662 — harness count 56 → 57
<!-- predicted: ls scripts/test-*.sh | wc -l after this task -->
**Commit:** `feat(agents): ship the nine flow model-effort agent definitions`
**Build:** green
**After:** none

- [x] 10. The universal `Model:` handshake

Per `design.md` **Enforcement › Universal handshake**. A **MODEL HANDSHAKE** paragraph, verbatim at
every dispatch site, in the "also carries" shape the sites use:

> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.

Sites: `skills/flow/implement.md` (conductor 34–44, implementer 407–431), `skills/flow/review-panel.md`
(slot 305–350, fix subagent 721–775), `skills/flow/verify-and-handoff.md` (verifier, beside 84–92),
`skills/flow/brainstorm.md` (planner 112–126). The handshake rule, stated once in
`skills/flow/implement.md` **Dispatch the conductor** replacing 70–81 and cited from every other
site: compare; mismatch → `dispatch end -outcome fallback`, re-dispatch once under `<key>-retry`
on the same type and model; second mismatch → `dispatch end -outcome fallback` and `## Question`
with the two options `design.md` names. `verify-and-handoff.md:116`'s own rule becomes a citation.
`brainstorm.md`'s planner handshake keeps its `opus` fallback and cites this rule for the retry key
shape. Every dynamic dispatch (implementer, slot, fix, verifier stays `sonnet`) passes
`subagent_type: flow-<model>-<effort>` and `model: <model>` and records `-effort`. `scripts/check-dispatch-paragraphs.sh`:
entry `[handshake]="**MODEL HANDSHAKE:**"` with shared phrases `Model: <the model named in your own
system prompt>` / `before any tool call`, sites `implement.md` min 2, `review-panel.md` min 2,
`brainstorm.md` min 1, `verify-and-handoff.md` min 1; `scripts/test-check-dispatch-paragraphs.sh`:
one case per new site.

  - [x] **Step 1: RED** — add the harness cases; run `scripts/test-check-dispatch-paragraphs.sh`;
    the new cases fail.
  - [x] **Step 2: GREEN** — the guard entry and the six site paragraphs; re-run, pass.
  - [x] **Step 3: Edit** the handshake rule and the dispatch shape as above.
  - [x] **Step 4: Verify** — `scripts/test-check-dispatch-paragraphs.sh`,
    `scripts/check-dispatch-paragraphs.sh`, `scripts/check-vocabulary.sh`,
    `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-installed-citations.sh`,
    `scripts/check-contract-budget.sh`; all exit 0 (`verify-and-handoff.md` has 1,419 bytes of
    room; raise the row if tripped).

**Files:** `skills/flow/implement.md`, `skills/flow/review-panel.md`,
`skills/flow/verify-and-handoff.md`, `skills/flow/brainstorm.md`,
`scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** `scripts/test-check-dispatch-paragraphs.sh`
**Regression:** reverting this commit removes the paragraph from every site; the guard's new entry
reports each site, and the harness's new cases fail.
**Baseline:** before=662 after=662
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(flow): handshake every dispatched role's model, retry once, then stop`
**Build:** green
**After:** Task 8, 9

## Group 4 — dynamic panel

- [x] 11. `review-panel.md` dispatches the decided roster

Per `design.md` **Dynamic panel dispatch** and **The tree**. Opening paragraph (3–9): the roster is
`REVIEWERS` when `REVIEW_PANEL_TOGGLE` is `default`, the decision's `panel.roster` when `dynamic`.
**The roster** table (117–124): the `Model` column reads "`DEFAULT_MODEL`, or the decision's
model/effort for this slot"; a paragraph after 135 states the dynamic dispatch shape (type
`flow-<model>-<effort>`, `model`, `-effort`) and that a compact roster is recorded in
`final-review-panel.md` as `compact — <rolled value>`. The dispatch record (258–283): `-effort`.
**Panel re-runs** (543–625): a **Rerun policy `full`** paragraph after 612 stating the final
whole-branch pass exactly as `design.md`'s bullet does, and that `delta` is the section as it
stands. The docs-only reduction (210–236) gains one sentence: it applies to a dynamic roster
unchanged. The `IN_PROGRESS` handoff's `Panel:` line (`handoff-blocks.md` 104) gains
`· <default|dynamic — class, compact?, rerun policy>`.

  - [x] **Step 1: Edit** both files as above.
  - [x] **Step 2: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-installed-citations.sh`,
    `scripts/check-normative-inventory.sh`, `scripts/check-contract-budget.sh`; all exit 0.

**Files:** `skills/flow/review-panel.md`, `skills/flow-contracts/handoff-blocks.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh`; `skills/flow/verify-and-handoff.md` —
panel round 1's F5-F7 fix, folded into this commit by fixup: `/flow`'s own live-printed `Panel:`
template (distinct from `handoff-blocks.md`'s `/flow-status`-only copy this task already updates)
was never brought into the same dynamic-roster shape — swept to match.
**Tests:** none — prose; the verify step's guards are the check
**Regression:** reverting this commit leaves the panel reading `REVIEWERS` alone; a `dynamic`
toggle changes nothing the panel does.
**Baseline:** before=662 after=662
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(flow): dispatch the decided panel roster, models, effort and rerun policy`
**Build:** green
**After:** Task 10

- [x] 12. Experimental reviewers

`skills/flow/experimental/failure-modes.md`: line 1 `description: What each changed behaviour does
under error return, timeout, partial write and concurrent re-entry.`, then a reviewer prompt in the
shape of `skills/flow/security-reviewer-prompt.md` — the angle, what to read
(`final-review.diff`, the context bundle), the finding format with severity and reproducer, and the
standards-as-data clause `principles-reviewer-prompt.md` carries. `skills/flow/review-panel.md`,
after **The roster**'s dynamic paragraph (task 11): an **Experimental slot** subsection per
`design.md` **Dynamic panel dispatch** — dispatched once, general-purpose by the decided type, prompt
file by absolute path, `-slot exp-<name>` on the dispatch and every finding, report file
`panel-report-<round>-exp-<name>.md`, re-run rule as any diff-reading slot, never in the docs-only
reduced roster. `skills/flow/brainstorm-planner.md` **Decide** (task 7): the experimental roll
resolves against `ls <agents repo>/skills/flow/experimental/*.md` sorted, index `experimental_roll
mod count`; none present records `experimental: none available`. `scripts/check-guard-symlinks.sh`
rule 5 forbids a top-level symlink in a skill directory — a real directory of real files is not
one; confirm by running it.

  - [x] **Step 1: Write** the prompt file and the two prose edits.
  - [x] **Step 2: Verify** — `scripts/check-guard-symlinks.sh`, `scripts/check-vocabulary.sh`,
    `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-installed-citations.sh`, `scripts/check-contract-budget.sh` (the new file needs
    a `budgets()` row: its own size plus 25%); all exit 0.

**Files:** `skills/flow/experimental/failure-modes.md`, `skills/flow/review-panel.md`,
`skills/flow/brainstorm-planner.md`, `scripts/check-contract-budget.sh`
**Tests:** none — a prompt file and prose; the verify step's guards are the check
**Regression:** reverting this commit leaves every experimental roll resolving to `none
available`; no `exp-` slot is ever dispatched.
**Baseline:** before=662 after=662
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(flow): add the experimental reviewer slot and its first prompt`
**Build:** green
**After:** Task 11

## Group 5 — inline execution

- [x] 13. The inline path

Per `design.md` **Inline execution**. `skills/flow/implement.md`: a new **Inline — the parent
implements** section after **Dispatch the conductor** (before line 106), entered when the decision's
`execution` is `inline`: the parent runs sections **1**, **2** and **4**, `review-panel.md` and
`verify-and-handoff.md` itself with the four substitutions `design.md` lists (no
conductor/implementer/fix dispatch; every implementer- and fixer-facing paragraph binds the parent;
slots and verifier dispatch unchanged; the inline `dispatches` rows with `-agent-id inline` and the
parent's model/effort). Section **4**'s wave paragraph (303–331) gains "inline runs bundles in plan
order, never in waves". `skills/flow/SKILL.md` **Guardrails** last bullet (216–218): "never run …
in the parent session" becomes "… unless the recorded decision's execution is `inline`, per
**Inline — the parent implements**". `skills/flow/review-panel.md` fix-subagent dispatch (776–796):
one sentence — inline, the parent applies the fix under the same paragraphs and records
`-role panel-fix -agent-id inline`.

  - [x] **Step 1: Edit** the three files as above.
  - [x] **Step 2: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-installed-citations.sh`,
    `scripts/check-normative-inventory.sh`, `scripts/check-contract-budget.sh`; all exit 0.

**Files:** `skills/flow/implement.md`, `skills/flow/SKILL.md`, `skills/flow/review-panel.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** none — prose; the verify step's guards are the check
**Regression:** reverting this commit leaves an `inline` decision with no path to run on; the
parent dispatches the conductor regardless.
**Baseline:** before=662 after=662
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(flow): add the inline execution path`
**Build:** green
**After:** Task 10

- [x] 14. The context ceiling

Per `design.md` **The context ceiling**. `skills/flow/implement.md` **Inline** section (task 13): a
**Context ceiling** paragraph — where the remaining figure comes from, the two thresholds, the
six-bundle fallback, the stop's mechanics (`-outcome stopped` on the open stage, no state write),
and that the block below is printed. `skills/flow-contracts/handoff-blocks.md`: a new `## Context
ceiling — clear and resume` block after the `IN_PROGRESS` block (after 192), in the exact shape
`design.md` shows, with the one-paragraph rule that it is printed only by an inline run and that
the next `/flow <name>` resumes on the recorded decision. `skills/flow/review-panel.md`: before
pass 1 (before 238), one sentence citing the pre-panel check.

  - [x] **Step 1: Edit** the three files as above.
  - [x] **Step 2: Verify** — the same guard list as task 13; all exit 0.

**Files:** `skills/flow/implement.md`, `skills/flow-contracts/handoff-blocks.md`,
`skills/flow/review-panel.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** none — prose; the verify step's guards are the check
**Regression:** reverting this commit leaves an inline run with no stop rule; KAN-475's 1M
constraint is unenforced.
**Baseline:** before=662 after=662
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(flow): stop an inline run at the context ceiling with a resume handoff`
**Build:** green
**After:** Task 13

## Group 6 — severity-gated minors

- [x] 15. Fix or defer each Minor

Per `design.md` **Severity-gated minors**. `skills/flow/review-panel.md` **Panel re-runs**: the
paragraph at 567–571 ("A Minor finding blocks and is fixed") becomes the fix-or-defer rule — the two
fix conditions, the `deferred <reason>` record call, the 10% sentence as guidance with the explicit
"never computed, 0% is acceptable", same in both modes; the handback line at 798 ("A minor finding
blocks the handoff exactly as a critical one does") becomes "a Minor either fixed or deferred blocks
nothing; a Minor left `open` blocks exactly as a Critical does". `check-panel-findings-closed.sh`
already accepts `deferred` (task 5). `skills/flow-contracts/handoff-blocks.md` `IN_PROGRESS` block:
a `### Deferred minors` list after the `Review the diff` lines, `F<n> <location> — <note> —
<reason>` per row, `none` when empty, and a `Deferred:` count line beside `Records:`.
`skills/flow/verify-and-handoff.md` **Write `IN_PROGRESS`** (423–454): fill both from `flow record
findings -change <name>` filtered on `startswith("deferred")`. `skills/flow/SKILL.md` **Guardrails**
"never hand off with an open finding" bullet (206–208): "a deferred Minor is not open".

  - [x] **Step 1: Edit** the four files as above.
  - [x] **Step 2: Verify** — the same guard list as task 13 plus `scripts/check-self-review-report.sh`;
    all exit 0.

**Files:** `skills/flow/review-panel.md`, `skills/flow-contracts/handoff-blocks.md`,
`skills/flow/verify-and-handoff.md`, `skills/flow/SKILL.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** none — prose; the store rule is task 5's and already tested
**Regression:** reverting this commit returns every Minor to the fix round; `deferred` is legal in
the store but no run ever writes it.
**Baseline:** before=662 after=662
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(flow): fix or defer each Minor finding by severity`
**Build:** green
**After:** Task 5, 12, 14

## Group 7 — stats views

- [x] 16. The `reviewers` view

Per `design.md` **Stats views › `reviewers`**. `stats/internal/store/aggregate.go`: `ReviewerRow`
{Slot, Experimental bool, Description, Dispatches, Changes, Critical, Major, Minor int, FindingsPerDispatch
float64, DeferredShare, WithdrawnShare float64} and `Reviewers(ctx, period, project, model *string)
([]ReviewerRow, error)` — findings joined to dispatches on `(change_id, slot)` and to `stage_runs`
for the period, the `StageLeaderboard` query's shape; `experimental = slot LIKE 'exp-%'`;
`description` from the newest `decisions` row whose `decision->'panel'->'roster'` contains an
entry with that slot (`jsonb_path_query_first`). `stats/internal/api/stats.go`: `viewReviewers`
in the constants and `knownViews`, `Reviewers` on `StatsStore`, a `rowsFor` arm and
`toReviewerDTOs`. `stats/web/src/api.ts`: `"reviewers"` in `ViewName`/`VIEW_NAMES`, `ReviewerRow`.
`stats/web/src/views/Reviewers.tsx`: `ViewFrame` + `DataTable`, an `Experimental` badge cell and an
info icon (`<span role="img" aria-label="experimental reviewer" title={description}>`) on
experimental rows. `App.tsx`: label "Reviewers" and component.

  - [x] **Step 1: RED** — `aggregate_test.go`: `TestReviewersCountsBySeverityAndMarksExperimental`;
    `stats_test.go`: extend `TestEveryViewCarriesItsRealNumbersThrough`'s table with the view;
    `views.test.tsx`: `reviewers shows severity counts and badges an exp- slot with its description`.
    Run `go test ./internal/store ./internal/api -run 'Reviewers|EveryView' -count=1` and `npx
    vitest run -t reviewers`; fail.
  - [x] **Step 2: GREEN** — implement; re-run, pass.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l . && go vet ./... && go test ./internal/store
    ./internal/api -run 'Reviewers|EveryView' -count=1`; `cd stats/web && npx tsc -b && npx vitest
    run -t reviewers`.

**Files:** `stats/internal/store/aggregate.go`, `stats/internal/api/stats.go`,
`stats/web/src/api.ts`, `stats/web/src/views/Reviewers.tsx`, `stats/web/src/App.tsx`,
`stats/internal/store/aggregate_test.go`, `stats/internal/api/stats_test.go`,
`stats/web/src/views/views.test.tsx`
**Allowed-collateral:** `stats/internal/api/changes_test.go`, `stats/internal/web/embed_test.go`,
`stats/internal/client/client_test.go` — one-line `Reviewers` pass-throughs on the `api.StatsStore`
test stand-ins, needed to keep them compiling once the interface grew that method;
`stats/web/src/views/RunDetail.test.tsx` — its `VIEW_NAMES` length assertion (`stats-ui-cut`'s
four-survivor count) goes stale the moment a fifth view is registered; updated to the real count
in the same fixup.
**Tests:** `TestReviewersCountsBySeverityAndMarksExperimental`, `TestEveryViewCarriesItsRealNumbersThrough`,
`reviewers shows severity counts and badges an exp- slot with its description`
**Regression:** reverting this commit removes the view; the Go tests fail on an unknown method and
view, the vitest case on a missing route.
**Baseline:** before=662 after=663 — vitest 157 → 158
<!-- predicted: go test … | wc -l and npx vitest run after this task -->
**Commit:** `feat(stats): add the reviewers view`
**Build:** green
**After:** Task 5

- [x] 17. The `decisions` view

Per `design.md` **Stats views › `decisions`**. `aggregate.go`: `DecisionRow` {Project, Change,
RecordedAt, Class, Overridden bool, Execution, ImplementerModel, ImplementerEffort, RosterSize int,
Compact bool, ExperimentalSlot, Rerun, WallClockSeconds float64, InputTokens, OutputTokens,
CacheReadTokens int64, CostUsd float64, Critical, Major, Minor, FixRounds, Fallbacks, TimedOut int}
and `Decisions(ctx, period, project *string) ([]DecisionRow, error)` — one row per `decisions` row
joined by `(change_id, session_token)` to `stage_runs` (min begin, max end), `dispatches`
(metrics sums via the same `tokens` keys `CostPerChange` reads; `outcome` counts) and `findings`;
the model filter does not apply (a decision spans models). `stats.go`: `viewDecisions`, the
`StatsStore` method, the `rowsFor` arm, `toDecisionDTOs`. `api.ts`: `"decisions"`, `DecisionRow`.
`stats/web/src/views/Decisions.tsx`: a summary `DataTable` grouped client-side by class ×
execution (count, mean wall-clock, mean cost, mean findings) above the per-run `DataTable`.
`App.tsx`: label "Decisions".

  - [x] **Step 1: RED** — `aggregate_test.go`: `TestDecisionsJoinsRunTotals`; `stats_test.go`: the
    view in `TestEveryViewCarriesItsRealNumbersThrough`'s table; `views.test.tsx`: `decisions
    shows a run's class, execution and cost, and a class × execution summary`. Run the selectors;
    fail.
  - [x] **Step 2: GREEN** — implement; re-run, pass.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l . && go vet ./... && go test ./internal/store
    ./internal/api -run 'Decisions|EveryView' -count=1`; `cd stats/web && npx tsc -b && npx vitest
    run -t decisions`. Then the full `## test` list once, in the foreground, as the last bundle's
    FULL SUITE paragraph requires.

**Files:** `stats/internal/store/aggregate.go`, `stats/internal/api/stats.go`,
`stats/web/src/api.ts`, `stats/web/src/views/Decisions.tsx`, `stats/web/src/App.tsx`,
`stats/internal/store/aggregate_test.go`, `stats/internal/api/stats_test.go`,
`stats/web/src/views/views.test.tsx`
**Allowed-collateral:** `stats/internal/api/changes_test.go`, `stats/internal/web/embed_test.go`,
`stats/internal/client/client_test.go` — one-line `Decisions` pass-throughs on the `api.StatsStore`
test stand-ins, needed to keep them compiling once the interface grew that method a second time
(task 16 needed the identical fix-up for its own new method).
**Tests:** `TestDecisionsJoinsRunTotals`, `TestEveryViewCarriesItsRealNumbersThrough`,
`decisions shows a run's class, execution and cost, and a class × execution summary`
**Regression:** reverting this commit removes the view; the stored decisions are reachable only by
SQL, and every test above fails on an unknown view.
**Baseline:** before=663 after=664 — vitest 158 → 159
<!-- predicted: go test … | wc -l and npx vitest run after this task -->
**Commit:** `feat(stats): add the decisions view`
**Build:** green
**After:** Task 16

## Group 8 — bundling

**Baseline for this group, measured after task 17 landed (merge base `4ddc51c`):**

- `cd stats && go test ./... -count=1` passes 678 top-level tests.
  <!-- measured: go test ./... -count=1 -json | jq -r 'select(.Action=="pass" and .Test!=null and (.Test|contains("/")|not)) | .Test' | wc -l @ branch spectre/kan-472-flow-dynamic-review-panel-roster-repo-scoped after task 17 -->
- `cd stats/web && npm test` passes 159 tests in 9 files.
  <!-- measured: npx vitest run @ branch spectre/kan-472-flow-dynamic-review-panel-roster-repo-scoped after task 17 -->
- `scripts/run-guard-tests.sh` discovers 57 `scripts/test-*.sh` harnesses.
  <!-- measured: ls scripts/test-*.sh | wc -l @ branch spectre/kan-472-flow-dynamic-review-panel-roster-repo-scoped after task 17 -->
- Byte sizes against `budgets()`: `skills/flow/review-panel.md` 57,816 of 58,732;
  `skills/flow/implement.md` 38,535 of 41,777; `skills/flow/brainstorm-planner.md` 25,765 of
  31,664; `skills/flow/security-reviewer-prompt.md` 1,540 of 1,540.
  <!-- measured: wc -c on each file and the budgets() rows in scripts/check-contract-budget.sh @ branch spectre/kan-472-flow-dynamic-review-panel-roster-repo-scoped after task 17 -->
- Line anchors after task 17. `skills/flow/review-panel.md`: 120–128 **The roster** table, 134–146
  the vocabulary and dynamic-dispatch paragraphs, 148–181 **Experimental slot**, 183–219 **An
  unspawnable id is substituted, not skipped**, 303–307 the "dispatch **separate** review
  subagents … Never merge two slots into one prompt" paragraph, 309–334 the dispatch record and its
  `-model`/`-effort` paragraph, 378–381 the MODEL HANDSHAKE carve-out for Bugbot and Security,
  462–475 the throwaway-worktree dispatch rule, 735 the `subagent_type` reproducer exemption.
  `skills/flow/implement.md`: 353–364 "Dispatch one implementer per bundle", 366–394 **Waves**,
  396–435 the gather call, 518–529 FULL SUITE, 531–578 **The next implementer overlaps the guard**.
  `skills/flow/brainstorm-planner.md`: 343–415 **Decide** (362–364 step 3, 367–373 the tree,
  375–379 the bugbot/security and rolls paragraph, 389–397 the JSON list, 399–411 the block).
  `scripts/plan-class.sh`: 9, 13, 37 the header's roll lines, 120 the `rolls:` printf.
  `stats/internal/store/aggregate.go`: 494–575 `Reviewers`, 601–620 `DecisionRow`, 650–760
  `Decisions` (713–727 the `panel` projections). `stats/internal/api/stats.go`: 772 `RosterSize`
  in `decisionRowDTO`, 793–800 `toDecisionDTOs`. `stats/web/src/api.ts`: 299 `DecisionRow`.
  `stats/web/src/views/views.test.tsx`: 187 the reviewers case, 206 the decisions case.
  `skills/flow-contracts/handoff-blocks.md`: 103 the `Panel:` line.
  `scripts/check-dispatch-paragraphs.sh`: 183–205 the `handshake` entry rows, 218 `SITE_ENTRY`.
  `scripts/check-contract-budget.sh`: 191–198 the `skills/flow/` rows.
  <!-- measured: grep -n / sed -n on each file @ branch spectre/kan-472-flow-dynamic-review-panel-roster-repo-scoped after task 17 -->

`skills/myflow-do/security-reviewer-prompt.md` is the legacy skill's copy and is not touched.

- [x] 18. Bugbot and Security as prompt-driven roles

Per `design.md` **Dynamic panel dispatch › Bugbot and Security are prompts**. Write
`skills/flow/bugbot-reviewer-prompt.md` in the shape of `skills/flow/experimental/failure-modes.md`
minus its `description:` line: the angle (a defect hunt over `final-review.diff` — wrong
conditions, missing guards, boundary errors, resource and error-path mistakes — proven by mutation:
**The mutation-testing brief**'s five moves, each against the copies the dispatch names, never
`<worktree>`), what to read, the finding format with severity and reproducer, the standards-as-data
clause `principles-reviewer-prompt.md` carries. Rewrite `skills/flow/security-reviewer-prompt.md`
into the same shape — the Cursor `security-review` framing, the `generalPurpose` substitute
sentence and the fenced `Subagent (generalPurpose):` block go; the threat list, the read-only rule
and the per-finding fields stay. `skills/flow/review-panel.md`: the `bugbot` and `security` rows
of **The roster** (125–126) become "general-purpose + `bugbot-reviewer-prompt.md`, own throwaway
worktree copy per repository" / "general-purpose + `security-reviewer-prompt.md`", Model column
"`DEFAULT_MODEL`, or the decision's model/effort for this slot"; 136–139 lose the `subagent_type`
sentence; 183–219 (**An unspawnable id is substituted, not skipped**) are removed whole, the
operator-named-id check at 199–205 kept as its own paragraph under **The roster**; 328–334 become
"`-model` is `DEFAULT_MODEL` (or this run's override) on `default` and the decision's model on
`dynamic`, for every slot; `-effort` likewise" — no exception; 378–381's carve-out goes (every slot
carries the handshake); 466–470 cite the bugbot prompt instead of a substituted slot; 735's
`subagent_type` exemption goes. Both prompt files are passed by absolute path, resolved like
`[PRINCIPLES_PATH]`. `skills/flow/brainstorm-planner.md` 375–376: the bugbot/security sentence
becomes `design.md` **The tree**'s bullet. `scripts/check-contract-budget.sh`: a row for the new
file (its size plus 25%) and the `security-reviewer-prompt.md` row raised to its new size plus 25%.

  - [x] **Step 1: Write** the two prompt files and the prose edits.
  - [x] **Step 2: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh`,
    `scripts/check-installed-citations.sh`, `scripts/check-normative-inventory.sh`,
    `scripts/check-contract-budget.sh`; all exit 0. `grep -n 'subagent_type: bugbot\|security-review\|unknown (agent-defined)\|unspawnable' skills/flow/review-panel.md skills/flow/brainstorm-planner.md` prints nothing.

**Files:** `skills/flow/bugbot-reviewer-prompt.md`, `skills/flow/security-reviewer-prompt.md`,
`skills/flow/review-panel.md`, `skills/flow/brainstorm-planner.md`, `scripts/check-contract-budget.sh`
**Allowed-collateral:** `scripts/check-installed-citations.sh` — a `declare_if_present` entry for
the new prompt file, the same fallout task 9 recorded for the agent definitions;
`scripts/check-references.sh` — `security-reviewer-prompt.md`'s rewrite legitimately cites
`review-panel.md`, so its old `EXPECTED_ZERO_REVIEWER_PROMPTS` entry needed removing to keep the
guard honest; `skills/flow/implement.md`, `skills/flow/SKILL.md`,
`skills/flow/verify-and-handoff.md`, `README.md`, `skills/flow-contracts/model-policy.md`,
`skills/flow-contracts/model-policy-rationale.md` — panel round 0's F1-F4 fix, folded into this
commit by fixup: these six files described the `subagent_type` Bugbot/Security dispatch mechanism
this task retires as still current, and the "never merged into one" claim task 20's bundling makes
false; swept to match `review-panel.md`'s own corrected state.
**Tests:** none — prompt files and prose; the verify step's guards are the check
**Regression:** reverting this commit returns `bugbot` and `security` to `subagent_type` dispatch;
no prompt-driven Bugbot exists and a bundle carrying either role has nothing to run.
**Baseline:** before=678 after=678 — no Go change; harness count 57
<!-- predicted: ls scripts/test-*.sh | wc -l after this task -->
**Commit:** `feat(flow): dispatch bugbot and security as prompt-driven roles`
**Build:** green
**After:** none

- [x] 19. The bundle roll, the grouping decision and implementer groups

Per `design.md` **The rolls**, **Bundled dispatch › Grouping**, **Implementer merging**, **The
`## Decision` block** and **Store › `decisions`**. `scripts/plan-class.sh`: a third roll
`bundle_roll = sha256("<name>bundle") mod 100`, computed exactly as the other two; the `rolls:`
line (120) becomes `rolls: compact N · experimental N · bundle N`; the header's three roll lines
(9, 13, 37) name it. `scripts/test-plan-class.sh`: the three-roll line shape asserted, the fixed
name's three rolls equal across two runs, and one fixture whose name rolls `bundle < 30` and one
`≥ 30` (find each by trying names; record the names in the harness). `skills/flow/brainstorm-planner.md`
**Decide**: step 3 (362–364) gains "and its grouping — `bundle_roll < 30` the class's static row,
else free within ≤2 dispatches × ≤3 roles with a one-line `grouping_reason`; a compact roster
dispatches separately; an experimental slot that fits no group is `skipped — bundle cap`"; a new
step 4 after it — implementer groups, per `design.md` **Implementer merging**: run
`plan-dispatch-bundles.sh <changeRoot>/tasks.md`, group its bundles freely, record `groups` and
`groups_reason`, or `null` when step 1 is inline; the tree (367–373) gains a **static grouping**
column carrying `design.md`'s three rows; the JSON list (389–397) gains `rolls.bundle`,
`panel.grouping`, `panel.dispatches`, `panel.grouping_reason`, `panel.experimental` (the skipped
string), `groups`, `groups_reason`; the block (399–411) gains the `dispatches:` suffix on the
review-panel row, the `grouping: free — <reason>` line, the `implementer groups` row and the
`bundle N (<interpretation>)` roll — `design.md`'s exact shape. `skills/flow/brainstorm.md`
**Resume and fix runs** paragraph: a fix run's second decision row re-groups on the same
name-derived roll; only `groups` (the appended plan's bundles) and a free grouping can differ.

  - [x] **Step 1: RED** — add the harness cases; run `scripts/test-plan-class.sh`; the new cases fail.
  - [x] **Step 2: GREEN** — the script; re-run, pass. Run it against this change's own `tasks.md`
    with `1` and record the printed `bundle` roll in the REPORT FILE.
    <!-- predicted: scripts/plan-class.sh spectre/changes/kan-472-flow-dynamic-review-panel-roster-repo-scoped/tasks.md 1 after this task -->
  - [x] **Step 3: Edit** `brainstorm-planner.md` and `brainstorm.md` as above.
  - [x] **Step 4: Verify** — `scripts/test-plan-class.sh`, `scripts/check-guard-symlinks.sh`,
    `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`,
    `scripts/check-contract-budget.sh`; all exit 0 (`brainstorm-planner.md` has 5,899 bytes of
    room; raise the row if tripped).

**Files:** `scripts/plan-class.sh`, `scripts/test-plan-class.sh`, `skills/flow/brainstorm-planner.md`,
`skills/flow/brainstorm.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** `scripts/test-plan-class.sh`
**Regression:** reverting this commit removes the third roll; the harness's three-roll cases fail
and no decision carries a grouping or implementer groups.
**Baseline:** before=678 after=678 — no Go change; harness count 57
<!-- predicted: ls scripts/test-*.sh | wc -l after this task -->
**Commit:** `feat(scripts): roll the panel grouping and record implementer groups`
**Build:** green
**After:** none

- [x] 20. `review-panel.md` bundles the roster into at most two dispatches

Per `design.md` **Bundled dispatch** — every subsection. `skills/flow/review-panel.md`: the
paragraph at 303–307 becomes "dispatch the decision's `panel.dispatches` — at most two per round,
each one to three roles — in the canonical worktree, each reading the whole combined file; a role
is never dispatched once per worktree: one pass reads every worktree's section (design.md's
`combined-diff-per-round`)", and a **Bundled dispatch** subsection follows it stating, by name from
`design.md`: the cap on both toggle values and the `default`-path grouping rule; one `dispatches`
row per bundle with the `+`-joined `-slot`, key and highest model/effort (the record at 309–334
shows `-slot <slot|slot+slot+slot>` and `-key panel-<round>-<that>`); the bundle prompt — shared
paragraphs once, one **PASS `<id>`** section per role in roster order carrying that role's brief
and its own REPORT FILE line, mutating passes last in their copies, one findings summary per role
in the return; the INDEPENDENT PASSES paragraph verbatim; no de-duplication; re-runs re-grouped by
the same grouping with only the re-running roles. **Experimental slot** (148–181): it joins the
group with room, last among reading passes, and is skipped when the decision recorded
`skipped — bundle cap`. **The docs-only reduction**: one dispatch. `final-review-panel.md` names the
dispatches as groups. `scripts/check-dispatch-paragraphs.sh`: entry `[independent]="**INDEPENDENT
PASSES:**"` with shared phrases `starts from \`final-review.diff\`` / `raise it again under this
pass` / `before beginning the next pass`, site `review-panel.md` min 1, appended to `SITE_ENTRY`
(218); `scripts/test-check-dispatch-paragraphs.sh`: one case. `skills/flow-contracts/handoff-blocks.md`
103: the `Panel:` line's dynamic suffix gains `, dispatches: <group> · <group>`.

  - [x] **Step 1: RED** — add the harness case; run `scripts/test-check-dispatch-paragraphs.sh`; it fails.
  - [x] **Step 2: GREEN** — the guard entry and the paragraph in `review-panel.md`; re-run, pass.
  - [x] **Step 3: Edit** the rest of `review-panel.md` and `handoff-blocks.md` as above.
  - [x] **Step 4: Verify** — `scripts/test-check-dispatch-paragraphs.sh`,
    `scripts/check-dispatch-paragraphs.sh`, `scripts/check-vocabulary.sh`,
    `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-installed-citations.sh`,
    `scripts/check-normative-inventory.sh`, `scripts/check-contract-budget.sh`; all exit 0
    (`review-panel.md` has 916 bytes of room and this task adds more: raise the row per the
    header rule). `grep -n 'Never merge two slots' skills/flow/review-panel.md` prints nothing.

**Files:** `skills/flow/review-panel.md`, `skills/flow-contracts/handoff-blocks.md`,
`scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** `scripts/test-check-dispatch-paragraphs.sh`
**Regression:** reverting this commit returns the panel to one subagent per slot with no cap; the
guard's new entry reports `review-panel.md` and the harness case fails.
**Baseline:** before=678 after=678
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(flow): bundle panel roles into at most two dispatches per round`
**Build:** green
**After:** Task 18, 19

- [x] 21. Implementer groups and the two-in-flight cap

Per `design.md` **Implementer merging**. `skills/flow/implement.md` section **4**: 353–364 —
the unit dispatched is the decision's `groups` entry (bundle ids from the same
`plan-dispatch-bundles.sh` output; a `null` decision field, which only inline produces, never
reaches this section), worked in plan order, one commit per task, the red/partner rule unchanged;
**Waves** (366–394) — readiness by the union after-set, **at most two implementer dispatches in
flight per wave** with the rest queued in plan order and launched as picks land, the throwaway
worktree per group named `<worktree>-wave-group-<g>`; the gather call (396–435) — one call over the
union of the group's task ids into `dispatch-context-group-<g>.md`, the CONTEXT BUNDLE paragraph
naming it; FULL SUITE (518–529) — the plan-last group; the overlap rule (531–578) — group N+2.
**Dispatch the conductor** (39–44): the conductor prompt already names the decision JSON path; add
that `groups` is read from it. `skills/flow/SKILL.md` **Guardrails**: the "never add a slot"
bullet's sibling on implementers, if any, gains "at most two in flight"; otherwise one new bullet.

  - [x] **Step 1: Edit** the two files as above.
  - [x] **Step 2: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-installed-citations.sh`,
    `scripts/check-normative-inventory.sh`, `scripts/check-contract-budget.sh`; all exit 0
    (`implement.md` has 3,242 bytes of room; raise the row if tripped).

**Files:** `skills/flow/implement.md`, `skills/flow/SKILL.md`
**Allowed-collateral:** `scripts/check-contract-budget.sh`
**Tests:** none — prose; the verify step's guards are the check
**Regression:** reverting this commit leaves the conductor dispatching one implementer per bundle
with unbounded waves; a recorded `groups` field changes nothing.
**Baseline:** before=678 after=678
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(flow): merge implementer bundles into planner-decided groups, two in flight`
**Build:** green
**After:** Task 19

- [x] 22. The `reviewers` view splits a bundled slot

Per `design.md` **Stats views › `reviewers`**. `stats/internal/store/aggregate.go` `Reviewers`
(494–575): `scoped_dispatches` selects `d.id, d.change_id, unnest(string_to_array(d.slot, '+')) AS
slot`; `scoped_findings` joins `findings f` to `scoped_dispatches sd` on `sd.id = f.dispatch_id`
(`findings.dispatch_id`, `0010_run_records.sql:85`, the column `-dispatch-seq` fills) `AND
sd.slot = f.slot`; the doc comment's `(change_id, slot)` sentence is replaced by the new
join and why (a bundle is one row; `design.md`'s `compound-slot-one-row-per-bundle`). A finding
recorded with no dispatch seq is not counted — state that in the comment.

  - [x] **Step 1: RED** — `aggregate_test.go`: `TestReviewersSplitsBundledSlots` — one dispatch
    `-slot primary+principles` with one `primary` finding and none for `principles`; expect two
    rows, `principles` with `dispatches=1` and zero findings, `primary` with one. Run `cd stats &&
    go test ./internal/store -run 'TestReviewersSplitsBundledSlots' -count=1 | tail -5`; fails.
  - [x] **Step 2: GREEN** — the query; re-run, pass. Run the existing
    `TestReviewersCountsBySeverityAndMarksExperimental` — still passes.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l . && go vet ./... && go test ./internal/store
    ./internal/api -run 'Reviewers|EveryView' -count=1`.

**Files:** `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`
**Tests:** `TestReviewersCountsBySeverityAndMarksExperimental` (the splits-bundled-slots case this
task originally added was, post-round-2's multi-commit fixup, folded by rebase replay into task
16's own commit instead — a rebase-replay attribution artifact, not a missing test; it is real,
passes, and is exercised by this task's own verify step regardless of which commit's diff contains
its declaration).
**Regression:** reverting this commit counts a bundled dispatch under one compound slot; the new
test fails on a missing `principles` row.
**Baseline:** before=678 after=679
<!-- predicted: go test ./... -count=1 -json | jq … | wc -l after this task -->
**Commit:** `feat(stats): count a bundled dispatch once per role in the reviewers view`
**Build:** green
**After:** Task 20

- [x] 23. The `decisions` view shows grouping and dispatches

Per `design.md` **Stats views › `decisions`**. `aggregate.go` `DecisionRow` (601–620) gains
`Grouping string`, `Dispatches string`, `ImplementerGroups string`; `Decisions` (650–760) projects
`panel->>'grouping'` (`default` when `panel` is a string), the `panel->'dispatches'` arrays
rendered as `+`-joined groups ` · `-joined (`array_to_string` over `jsonb_array_elements_text`
per group), and `groups` the same way over bundle ids (empty when `null`). `stats.go`:
`decisionRowDTO` and `toDecisionDTOs` (772–800) carry the three; `api.ts` `DecisionRow` (299);
`Decisions.tsx`: three columns after `rerun`, and the client-side summary additionally groups by
`grouping`. `views.test.tsx` 206: the fixture carries a grouping and the assertion names it.

  - [x] **Step 1: RED** — `aggregate_test.go`: `TestDecisionsRendersGrouping` (a `free` row with two
    groups and `groups [[1,2],[3]]` → `primary+principles · code-review-low+mutation` and
    `1+2 · 3`; a `default` panel string → `default`, empty dispatches); `stats_test.go`: the view's
    row in `TestEveryViewCarriesItsRealNumbersThrough` carries the new columns; `views.test.tsx`:
    the decisions case asserts the grouping cell. Run the selectors; fail.
  - [x] **Step 2: GREEN** — implement; re-run, pass.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l . && go vet ./... && go test ./internal/store
    ./internal/api -run 'Decisions|EveryView' -count=1`; `cd stats/web && npx tsc -b && npx vitest
    run -t decisions`. Then the full `## test` list once, in the foreground, as the last bundle's
    FULL SUITE paragraph requires.

**Files:** `stats/internal/store/aggregate.go`, `stats/internal/api/stats.go`,
`stats/web/src/api.ts`, `stats/web/src/views/Decisions.tsx`,
`stats/internal/store/aggregate_test.go`, `stats/internal/api/stats_test.go`,
`stats/web/src/views/views.test.tsx`
**Tests:** `TestDecisionsRendersGrouping`, `TestEveryViewCarriesItsRealNumbersThrough`,
`decisions shows a run's class, execution and cost, and a class × execution summary` (round 4's F14
fix moved the grouping-rendering case back into this commit, along with the group_sources/
flattened_groups SQL it exercises: round 3's F12 fixup had landed both in task 17's commit
instead, where the SQL selected three columns the Decisions row type did not yet declare and the
test itself predated the Decisions store method and its Grouping/Dispatches fields — both broke
task 17's own standalone build; task 22's splits-bundled-slots rebase-replay artifact in task 16's
commit is unrelated and untouched).
**Regression:** reverting this commit hides the grouping from the view; the stored field is
reachable only by SQL and the new test fails on a missing column.
**Baseline:** before=679 after=680 — vitest 159 → 159 (the existing case is extended)
<!-- predicted: go test … | wc -l and npx vitest run after this task -->
**Commit:** `feat(stats): show grouping and dispatches in the decisions view`
**Build:** green
**After:** Task 19, 22

## Group 9 — simple-reviewer

- [x] 24. The simple-reviewer slot

Small class's compact-roster code-quality reviewer is a brand-new persistent slot,
`simple-reviewer`, never a rename or reuse of `code-review-low` (`design.md`'s
`simple-reviewer-new-slot` decision). New prompt file `skills/flow/simple-reviewer-prompt.md`, same
shape as `bugbot-reviewer-prompt.md`/`security-reviewer-prompt.md`: general-purpose + prompt, no
`subagent_type`, high-confidence defects only, its own persona in its own words. `ValidReviewers`
(`stats/internal/store/settings.go`) gains the id; `skills/flow/review-panel.md`'s **The roster**
table gains its row, and the "six entries, never a seventh" sentence is corrected to seven.
`scripts/check-contract-budget.sh` gains the new file's budget row. Confirmed by reading
`stats/internal/store/aggregate.go`'s `Reviewers` query that the `reviewers` stats view counts by
slot generically (no hardcoded slot list) and needs no change. Confirmed
`scripts/check-installed-citations.sh` needs no new `declare_if_present` entry — the new prompt
file's citation is checked cleanly without one.

  - [x] **Step 1: RED** — `settings_test.go`: `TestPutSettingsAcceptsSimpleReviewer` asserts
    `PutSettings` accepts a `Reviewers` list containing `simple-reviewer`. Run `cd stats && go test
    ./internal/store -run TestPutSettingsAcceptsSimpleReviewer -count=1`; fails against the
    unmodified `ValidReviewers`.
  - [x] **Step 2: GREEN** — add `"simple-reviewer": true` to `ValidReviewers`; re-run, pass.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l . && go vet ./... && go test ./internal/store
    -run 'ValidReviewer|Reviewer' -count=1`; from the worktree root: `scripts/check-vocabulary.sh`,
    `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-dispatch-paragraphs.sh`, `scripts/check-installed-citations.sh` (2 pre-existing
    unrelated `spectre/changes/` root-prefix failures, confirmed via `git stash` to predate this
    task), `scripts/check-normative-inventory.sh`, `scripts/check-contract-budget.sh`.

**Files:** `skills/flow/simple-reviewer-prompt.md`, `stats/internal/store/settings.go`,
`stats/internal/store/settings_test.go`, `skills/flow/review-panel.md`,
`scripts/check-contract-budget.sh`
**Tests:** `TestPutSettingsAcceptsSimpleReviewer`
**Regression:** reverting this commit drops `simple-reviewer` from `ValidReviewers`; the new test
fails on `ErrInvalidReviewer`.
**Commit:** `feat(flow): add the simple-reviewer slot for small class's compact roster`
**Build:** green
**After:** Task 19
