# kan-450-record-dispatch-end-timestamps-duration-and

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Three layers in dependency order: the cost figure onto the wire (task 1), the end/duration lines
the ledger already has data for (task 2), and the totals section that consumes both (task 3).
`design.md` is canonical for every decision.

**Baseline, measured before any edit:**

- `go test ./internal/records/ ./internal/store/ -count=1`: 191 tests pass.
  <!-- measured: go test ./internal/records/ ./internal/store/ -count=1 -v | grep -c '^--- PASS' @ branch spectre/kan-450-record-dispatch-end-timestamps-duration-and, worktree at merge-base e08e8ee -->

---

- [x] 1. Carry per-dispatch cost on the run record

Add `CostUSD *float64` (`json:"costUsd,omitempty"`) to `records.Dispatch` in
`stats/internal/records/types.go` — a pointer, so an unpriced dispatch stays distinguishable from
one priced at zero, the same rule `Metrics`' pointer-shaped keys already follow.

Join the figure in `stats/internal/store/records.go`'s `readDispatches`:

```sql unverified:confirm the jsonb path against Price's writer (internal/store/pricing.go) and the dispatches table shape (0010_run_records.sql, 0011, 0012, 0014)
LEFT JOIN stage_runs ON stage_runs.id = dispatches.stage_run_id
-- cost column, alongside the existing dispatch columns:
(stage_runs.metrics -> 'dispatches' -> dispatches.agent_id ->> 'cost_usd')::numeric
```

`NULL` when there is no stage run, no agent id, or no priced bucket — scan into the pointer as
nil, never zero. `CostUSD` is derived and read-only: `insertDispatch` and `EndDispatch` never
write it. Extend the round-trip scan helpers (`qualifiedDispatchColumns`/`scanDispatchRow`) so
every reader of a dispatch row sees the new column.

TDD: failing store test first, in `stats/internal/store/records_test.go` — a change whose
dispatch names a priced stage-run bag yields `CostUSD` on `RunRecord`'s row; a dispatch with no
`stage_run_id`, an empty agent id, or a bag without that key yields nil.

**Files:** `stats/internal/records/types.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`
**Tests:** `TestRunRecordJoinsDispatchCost`, `TestRunRecordUnpricedDispatchCostStaysNil`
**Regression:** reverting this commit leaves every ledger cost figure nil — task 3's totals
section then reports every dispatch unpriced.
**Baseline:** before=191 after=193 tests in `./internal/records/ + ./internal/store/`
<!-- measured: go test ./internal/records/ ./internal/store/ -count=1 -v | grep -c '^--- PASS' @ branch spectre/kan-450-record-dispatch-end-timestamps-duration-and, merge-base e08e8ee -->
<!-- predicted: go test ./internal/records/ ./internal/store/ -count=1 after this task -->
**Commit:** `feat(stats): carry per-dispatch cost on the run record`
**Build:** green

Verify step:

  - [x] **Step 1: `cd stats && gofmt -l .` and `go vet ./...` exit clean**
  - [x] **Step 2: `cd stats && go test ./internal/records/ ./internal/store/ -run 'TestRunRecord' -count=1` passes, and the full two-package run matches the Baseline after-count**

- [x] 2. Render dispatch end and duration in the ledger

`stats/internal/records/render.go`'s `RenderLedger` prints, directly after `- Started:`:

- `- Ended: <RFC3339 UTC>` when `EndedAt` is non-nil, else `- Ended: not recorded` — a still-open
  dispatch is ordinary, the same voice `orElse` already uses.
- `- Duration: <(EndedAt − StartedAt).Round(time.Second).String()>` only when `EndedAt` is
  non-nil **and** `StartedAt` is non-zero — a duration needs both instants, and one missing is an
  absence, never zero.

TDD: failing render tests first in `stats/internal/records/render_test.go` — closed dispatch
renders both lines; open dispatch renders `not recorded` and no Duration line; zero `StartedAt`
with `EndedAt` set renders Ended without Duration.

**Files:** `stats/internal/records/render.go`, `stats/internal/records/render_test.go`
**Tests:** `TestRenderLedgerEndedAndDuration`, `TestRenderLedgerOpenDispatchHasNoDuration`
**Regression:** reverting this commit hides every end instant and duration from the ledger while
the store still holds them — the reader is back to querying by hand.
**Baseline:** before=193 after=195 tests in `./internal/records/ + ./internal/store/`
<!-- predicted: go test ./internal/records/ ./internal/store/ -count=1 -v | grep -c '^--- PASS' after this task, continuing from task 1's after-count -->
**Commit:** `feat(stats): render dispatch end and duration in the ledger`
**Build:** green

Verify step:

  - [x] **Step 1: `cd stats && gofmt -l .` and `go vet ./...` exit clean**
  - [x] **Step 2: `cd stats && go test ./internal/records/ -run 'TestRenderLedger' -count=1` passes, and the full two-package run matches the Baseline after-count**

- [x] 3. Total ledger cost by role and task

New pure aggregation in `stats/internal/records/cost.go`: `CostByRoleTask(r Run) []RoleTaskCost`
groups `r.Dispatches` by (role, taskID — raw values, no fallbacks) and returns, per group:
dispatch count, the four token sums (`input`, `output`, `cache_read`, `cache_creation`, each
main+sidechain, over rows whose metrics carry a `tokens` key at all), the `CostUSD` sum over
non-nil figures, and counts of unmeasured and unpriced rows. Absence is never zero: a row with no
`tokens` key adds to no token sum, and a nil `CostUSD` adds to no dollar sum — both surface as
their count instead. `RoleTaskCost` lives beside it; the group order is the rows' seq order of
first appearance, deterministic and test-friendly.

`RenderLedger` appends, after the dispatch sections:

```markdown verified:authored in-tree for this change; shape only, rendered by task 3's code
## Cost by role and task

| Role | Task | Dispatches | Input | Output | Cache read | Cache creation | Cost USD |
|---|---|---|---|---|---|---|---|
| implementer | 3 | 2 | 1200 | 340 | 900 | 10 | 0.42 |

Unmeasured dispatches: 1
Unpriced dispatches: 1
```

The section is omitted when the change has no dispatches (the early return already covers that
ledger). Per-group unmeasured/unpriced counts are **not** rendered — only the ledger-level
totals — so the table stays a totals table; a reader drills into the per-dispatch sections above
for the why.

TDD: failing tests first — `stats/internal/records/cost_test.go` (grouping, sums, absence
counts, ordering) and `stats/internal/records/render_test.go` (section shape, omission when no
dispatches).

**Files:** `stats/internal/records/cost.go`, `stats/internal/records/cost_test.go`, `stats/internal/records/render.go`, `stats/internal/records/render_test.go`
**Tests:** `TestCostByRoleTaskGroupsAndSums`, `TestCostByRoleTaskAbsenceIsNotZero`, `TestRenderLedgerCostSection`, `TestRenderLedgerCostSectionOmittedWithoutDispatches`
**Regression:** reverting this commit removes the cost-per-change figure entirely — KAN-450's
outlier visibility is gone and cost again requires manual summing across sections.
**Baseline:** before=195 after=199 tests in `./internal/records/ + ./internal/store/`
<!-- predicted: go test ./internal/records/ ./internal/store/ -count=1 -v | grep -c '^--- PASS' after this task, continuing from task 2's after-count -->
**Commit:** `feat(stats): total ledger cost by role and task`
**Build:** green

Verify step:

  - [x] **Step 1: `cd stats && gofmt -l .` and `go vet ./...` exit clean**
  - [x] **Step 2: `cd stats && go test ./internal/records/ ./internal/store/ -count=1` passes, and the count matches the Baseline after-count**
