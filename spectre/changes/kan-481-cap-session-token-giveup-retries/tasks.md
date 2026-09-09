# kan-481-cap-session-token-giveup-retries

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Baseline for every `Baseline:` below is the stats Go suite, whole-repo run:
`cd stats && go test ./... -race -count=1`, counted as `go test ./... -race -count=1 -v | grep -c '^=== RUN'` (parents + subtests).
Before task 1 the suite is green with 1072 test cases.
<!-- measured: cd stats && go test ./... -race -count=1 -v | grep -c '^=== RUN' @ main 3bea93f, 2026-09-10 -->

- [x] 1. Cap give-up retries at three failed generations in `PersistedGiveUps`
**Build:** green
**Files:** `stats/internal/store/giveups.go`, `stats/internal/store/giveups_test.go`
**Tests:** `TestPersistedGiveUpsDropsExhaustedRetries`
**Regression:** reverting re-seeds permanently-failed tokens on every daemon start — the whole-corpus retry scan returns for rows that keep refreshing their `gave_up_at`, which is exactly the measured 15-minute stall; the new test fails and the horizon test from kan-480 keeps the age bound pinned
**Baseline:** before=1072 after=1073
<!-- measured: full suite after implementation counts 1073 (+1, the new parent test; no subtests) -->
**Commit:** `fix(store): drop exhausted session-token give-ups from retry seeding`

  - [x] **Step 1: RED** — `TestPersistedGiveUpsDropsExhaustedRetries` (giveups_test.go, real store): record one token four times (`freshAt`, the upsert increments retries to 3) and a low-retries control; `PersistedGiveUps` returns only the control and the table no longer carries the exhausted row. Run `cd stats && go test -race -count=1 -run 'TestPersistedGiveUps' ./internal/store/` and report the failing output.
  - [x] **Step 2: GREEN** — in `giveups.go`: unexported `giveUpMaxRetries = 3` with a doc comment naming the measured history (cleared rows had reached retries 10); the DELETE gains `OR retries >= $2`; the SELECT gains `AND retries < $2`; update doc comments.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l .` (empty), `cd stats && go vet ./...`, `cd stats && go test -race -count=1 -run 'TestPersistedGiveUps|TestRecordSessionTokenGiveUp|TestGiveUp' ./internal/store/`, then the full suite per the header.
