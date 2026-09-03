# kan-375-stats-two-go-tests-fail-when-a-workspace-s

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

Two independent one-line test fixes, one per package. `design.md` is canonical for every decision;
`proposal.md` for the failures and the sweep. Neither task changes a non-test file.

**Baseline, measured before any edit:**

- Both tests fail, and only those two, when the workspace variables are exported.
  <!-- measured: cd stats && FLOWD_PORT=5753 FLOW_ADDR=http://127.0.0.1:5753 FLOWD_DSN='postgres://flow:flow@localhost:5433/flow_x?sslmode=disable' go test ./... -race -count=1 @ aeb23f6 -->
- `stats/internal/config` has 3 top-level tests; `stats/cmd/flow` has 136.
  <!-- measured: go test ./internal/config -count=1 -v | grep -cE '^--- (PASS|FAIL)' and the same for ./cmd/flow, from stats/ @ aeb23f6 -->

**Each task's verification runs the same command twice, bare and with the variables exported.** The
exported form is exactly what `prepare-workspace.sh` prints and `/flow`'s verify stage uses; the
bare form proves the pin did not break the run everyone had before.

---

- [x] 1. Pin FLOWD_PORT in TestFromEnvDefaultsPortWhenUnset

`TestFromEnvDefaultsPortWhenUnset` in `stats/internal/config/config_test.go` calls
`config.FromEnv()` with no pin and asserts `DefaultPort`, so it reads whatever `FLOWD_PORT` the
caller exported. Pin it empty: `FromEnv` reads with `if v := os.Getenv("FLOWD_PORT"); v != ""`, so
an empty value is unset by construction, and `t.Setenv` restores the caller's value on cleanup.

**Files:** `stats/internal/config/config_test.go`
**Tests:** `TestFromEnvDefaultsPortWhenUnset` — existing, now isolated from the ambient environment; none added
**Regression:** reverting this commit makes `TestFromEnvDefaultsPortWhenUnset` fail again with
`got port 5753, want default 4173` whenever `FLOWD_PORT=5753` is exported.
**Baseline:** before=3 after=3
<!-- measured: cd stats && go test ./internal/config -count=1 -v | grep -cE '^--- (PASS|FAIL)' @ aeb23f6; the count is unchanged because no test is added -->
**Commit:** `fix(config): pin FLOWD_PORT in the unset-port test`
**Build:** green

  - [x] **Step 1: Reproduce the failure**

Run, from `stats/`:

```bash verified:ran at aeb23f6 — fails with "got port 5753, want default 4173"
FLOWD_PORT=5753 go test ./internal/config -race -count=1 -run TestFromEnvDefaultsPortWhenUnset
```

Expected: FAIL, `config_test.go:52: got port 5753, want default 4173`.

  - [x] **Step 2: Add the pin**

In `stats/internal/config/config_test.go`, make the first line of
`TestFromEnvDefaultsPortWhenUnset`'s body:

```go verified:spiked at aeb23f6 — the test passes with and without FLOWD_PORT exported
func TestFromEnvDefaultsPortWhenUnset(t *testing.T) {
	t.Setenv("FLOWD_PORT", "")

	cfg, err := config.FromEnv()
```

Extend the doc comment above the function by one sentence saying the pin exists because the
caller's shell may export `FLOWD_PORT` (a `/flow` worktree does), and that `FromEnv` treats an
empty value as unset.

  - [x] **Step 3: Verify, both ways**

Run, from `stats/`:

```bash verified:ran at aeb23f6 with the spike applied — both invocations ok
FLOWD_PORT=5753 go test ./internal/config -race -count=1
go test ./internal/config -race -count=1
```

Expected: `ok` twice. Then `gofmt -l .` from `stats/` prints nothing and `go vet ./internal/config`
exits 0.

  - [x] **Step 4: Commit**

`git add stats/internal/config/config_test.go` and commit with the subject in `**Commit:**`.

---

- [x] 2. Pin FLOW_ADDR to the test daemon in the real-guard panel test

`TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard` in `stats/cmd/flow/record_test.go`
runs the real `scripts/check-unfinished-work.sh` through `exec.Command`, which inherits the
process environment. The guard's line 317 runs `flow record findings -change demo -C <repo>` with
no `-addr`, so `resolveDefaultAddr` in `stats/cmd/flow/state.go` hands that subprocess whatever
`FLOW_ADDR` the caller exported — and, unset, the dev daemon at `127.0.0.1:4173`. Pin it to the
test's own `httptest` server: `renderDaemon` already answers every GET with a record carrying
`"findings":[]` under the genuine-daemon header, which is exactly what `record findings` reads
through `GetRunRecord`, so the guard sees `[]` and prints `CLEAR:`. The `record render` call in
the same test keeps its explicit `-addr srv.URL` and is untouched.

**Files:** `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard` — existing, now isolated from
the ambient environment; none added
**Regression:** reverting this commit makes
`TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard` fail with
`check-unfinished-work.sh exited non-zero (exit status 2)` whenever `FLOW_ADDR` names a port with
no daemon, and leaves it dependent on the dev daemon at `4173` when `FLOW_ADDR` is unset.
**Baseline:** before=136 after=136
<!-- measured: cd stats && go test ./cmd/flow -count=1 -v | grep -cE '^--- (PASS|FAIL)' @ aeb23f6; the count is unchanged because no test is added -->
**Commit:** `fix(flow): pin FLOW_ADDR to the test daemon in the real-guard panel test`
**Build:** green

  - [x] **Step 1: Reproduce the failure**

Run, from `stats/`:

```bash verified:ran at aeb23f6 — fails with "check-unfinished-work.sh exited non-zero (exit status 2)"
FLOW_ADDR=http://127.0.0.1:5753 go test ./cmd/flow -race -count=1 -run TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard
```

Expected: FAIL, `record_test.go:1035: check-unfinished-work.sh exited non-zero (exit status 2)`,
with `dial tcp 127.0.0.1:5753: connect: connection refused` in the guard's captured stderr.

  - [x] **Step 2: Add the pin**

In `stats/cmd/flow/record_test.go`, inside
`TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard`, directly after `defer srv.Close()`:

```go verified:spiked at aeb23f6 — the test passes with and without FLOW_ADDR exported
	srv := httptest.NewServer(renderDaemon(t, `{"change":"demo","dispatches":[],"findings":[]}`))
	defer srv.Close()
	t.Setenv("FLOW_ADDR", srv.URL)
```

Add a comment above the `t.Setenv` line saying the guard shells out to `flow record findings`
with no `-addr`, that the subprocess inherits this environment, and that pinning it to `srv.URL`
keeps the guard reading from this test's daemon rather than from the caller's `FLOW_ADDR` or the
dev daemon on `4173`. Do not restate `design.md`'s rejected alternative there.

  - [x] **Step 3: Verify, both ways**

Run, from `stats/`:

```bash verified:ran at aeb23f6 with the spike applied — both invocations ok
FLOW_ADDR=http://127.0.0.1:5753 go test ./cmd/flow -race -count=1
go test ./cmd/flow -race -count=1
```

Expected: `ok` twice. Then `gofmt -l .` from `stats/` prints nothing and `go vet ./cmd/flow`
exits 0.

  - [x] **Step 4: Whole-suite acceptance**

Run, from `stats/`, the acceptance command KAN-375 names, then the bare form:

```bash verified:ran at aeb23f6 with both spikes applied — every package ok, both invocations
FLOWD_PORT=5753 FLOW_ADDR=http://127.0.0.1:5753 FLOWD_DSN='postgres://flow:flow@localhost:5433/flow_x?sslmode=disable' go test ./... -race -count=1
go test ./... -race -count=1
```

Expected: every package `ok`, both times. Confirm the sweep still holds:

```bash verified:ran at aeb23f6 — the only hits are the pins design.md's sweep table lists
grep -rn --include='*_test.go' -E 'FLOWD_PORT|FLOW_ADDR' .
```

Expected: hits only in `internal/config/config_test.go` (three `t.Setenv`) and
`cmd/flow/state_test.go:1389-1406` plus this task's own `t.Setenv` in `cmd/flow/record_test.go`.

  - [x] **Step 5: Commit**

`git add stats/cmd/flow/record_test.go` and commit with the subject in `**Commit:**`.
