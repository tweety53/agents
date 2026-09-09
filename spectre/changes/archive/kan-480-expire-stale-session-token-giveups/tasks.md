# kan-480-expire-stale-session-token-giveups

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Baseline for every `Baseline:` below is the stats Go suite, whole-repo run:
`cd stats && go test ./... -race -count=1`, counted as `go test ./... -race -count=1 -v | grep -c '^=== RUN'` (parents + subtests).
Before task 1 the suite is green with 1071 test cases.
<!-- measured: cd stats && go test ./... -race -count=1 -v | grep -c '^=== RUN' @ main 2f1ac1f, 2026-09-10; first written as 1036 from the kan-479 branch's own count, corrected to the measured 1071 (main's suite had grown past that branch point) -->

- [x] 1. Expire give-ups past a 24-hour retry horizon in `PersistedGiveUps`
**Build:** green
**Files:** `stats/internal/store/giveups.go`, `stats/internal/store/giveups_test.go`
**Tests:** `TestPersistedGiveUpsExpiresPastHorizon`
**Regression:** reverting re-seeds every dead give-up on every daemon start — the retry scan re-reads the whole transcript corpus per cycle for a full bounded window (measured 15+ min at ~95% CPU here), stalling binding and attribution; `TestPersistedGiveUpsExpiresPastHorizon` fails and the fresh-row passthrough case keeps every existing give-up behavior pinned
**Baseline:** before=1071 after=1072
<!-- measured: full suite after implementation counts 1072 (+1, the new parent test; no subtests) -->
**Commit:** `fix(store): expire session-token give-ups past a 24-hour retry horizon`

  - [x] **Step 1: RED** — `TestPersistedGiveUpsExpiresPastHorizon` (giveups_test.go, real store): record one give-up with `at = now-25h` and one with `at = now-1h` (retries 2 and 1) via `RecordSessionTokenGiveUp`; `PersistedGiveUps` returns only the fresh token with retries 1; the table no longer carries the expired row. Run `cd stats && go test -race -count=1 -run 'TestPersistedGiveUps' ./internal/store/` and report the failing output.
  - [x] **Step 2: GREEN** — in `giveups.go`: unexported `giveUpRetryHorizon = 24 * time.Hour` with a doc comment naming the measured cost it bounds; `PersistedGiveUps` computes `cutoff := time.Now().Add(-giveUpRetryHorizon)`, deletes rows with `gave_up_at < cutoff`, selects rows with `gave_up_at >= cutoff`; update the function's doc comment.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l .` (empty), `cd stats && go vet ./...`, `cd stats && go test -race -count=1 -run 'TestPersistedGiveUps|TestRecordSessionTokenGiveUp|TestGiveUp' ./internal/store/`, then the full suite per the header.
