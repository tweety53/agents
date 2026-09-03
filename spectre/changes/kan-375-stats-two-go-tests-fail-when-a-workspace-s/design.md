# stats: two Go tests fail when a workspace's FLOWD_PORT and FLOW_ADDR are exported — design

**Change:** `kan-375-stats-two-go-tests-fail-when-a-workspace-s`
**Jira:** KAN-375
**Date:** 2026-09-03

## Context

### The failures, reproduced

```bash verified:ran at aeb23f6 on 2026-09-03 — exactly these two tests fail; every other package is ok
cd stats && FLOWD_PORT=5753 FLOW_ADDR=http://127.0.0.1:5753 \
  FLOWD_DSN='postgres://flow:flow@localhost:5433/flow_x?sslmode=disable' \
  go test ./... -race -count=1
# --- FAIL: TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard
#     record_test.go:1035: check-unfinished-work.sh exited non-zero (exit status 2)
# --- FAIL: TestFromEnvDefaultsPortWhenUnset
#     config_test.go:52: got port 5753, want default 4173
```

The ticket names "the findings test" in `stats/cmd/flow` as the second failure. It is not the
`record findings` verb tests — those pass `-addr` explicitly and never read the environment. It is
the real-guard panel test: `exec.Command("bash", guard, repo, "demo")` inherits the process
environment, `scripts/check-unfinished-work.sh:317` runs `flow record findings -change demo -C
<repo>` with no `-addr`, and `resolveDefaultAddr` in `stats/cmd/flow/state.go` returns `FLOW_ADDR`
when set. Unset, it returns `http://127.0.0.1:4173` — the dev daemon — which is why the test is
green today and why that green depends on the dev stack being up.

## The fix

Two `t.Setenv` lines. No source file changes.

**`stats/internal/config/config_test.go`** — `TestFromEnvDefaultsPortWhenUnset` opens with
`t.Setenv("FLOWD_PORT", "")`. `config.FromEnv` guards each read with `if v := os.Getenv(...); v !=
""`, so empty is unset by construction, and `t.Setenv` restores the caller's value when the test
ends. `FLOWD_HOST` and `FLOWD_DSN` are not asserted by that test and stay untouched.

**`stats/cmd/flow/record_test.go`** — `TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard`
adds `t.Setenv("FLOW_ADDR", srv.URL)` directly after `defer srv.Close()`. The test's `renderDaemon`
answers every GET with `{"change":"demo","dispatches":[],"findings":[]}` under the genuine-daemon
header, which is exactly what `record findings` reads through `cl.GetRunRecord`; the guard then
sees `[]`, counts zero open findings, and prints `CLEAR:`. The `record render` call in the same
test keeps its explicit `-addr srv.URL`, so the pin changes only the subprocess.

Spiked both lines and ran the two tests with and without the variables exported — green both ways;
reverted before this design was written.

## Sweep

Every `os.Getenv` / `t.Setenv` in `stats/**/*_test.go`, at `aeb23f6`:

| Variable | Where | Pinned? |
|----------|-------|---------|
| `FLOWD_PORT` | `internal/config/config_test.go` ×3 | two already `t.Setenv`; the third is this change |
| `FLOW_ADDR` | `cmd/flow/state_test.go:1389–1406` | already `t.Setenv` (table test and flag-beats-env test) |
| `FLOW_ADDR` (via subprocess) | `cmd/flow/record_test.go` real-guard test | this change |
| `FLOW_STATE_DIR`, `FLOW_HARNESS`, `TMPDIR` | `cmd/flow`, `internal/fallback`, `cmd/flowd` | already `t.Setenv` |
| `FLOW_STATS_DSN`, `FLOW_STATS_ADMIN_DSN`, `FLOW_STATS_SKIP_SLOW_TESTS`, `UPDATE_*_FIXTURE` | store/api/harvest/sweep/reconcile helpers | opt-in switches, deliberately ambient; not workspace variables |

`FLOWD_DSN` is read by no test; the full-suite run above exported it to a non-existent database and
nothing else failed.

## Testing

`cd stats && go test ./... -race -count=1`, run twice from the change branch: once bare, once with
`FLOWD_DSN`, `FLOWD_PORT=5753`, `FLOW_ADDR=http://127.0.0.1:5753` exported. Both green is the
acceptance.

## Decisions

### Pin FLOW_ADDR to the test daemon rather than clearing it

**ID:** pin-flow-addr-to-test-daemon
**Status:** active
**Chosen:** `t.Setenv("FLOW_ADDR", srv.URL)` — the guard's `flow record findings` subprocess reads
from the same `httptest` daemon the test already stands up, so the test is self-contained.
**Considered:** clearing `FLOW_ADDR` (`t.Setenv("FLOW_ADDR", "")`) — same one-line cost, but the
subprocess then dials `127.0.0.1:4173` and the test stays green only while the dev daemon answers,
which is the hidden dependency that made the ticket misattribute the failure in the first place.

### No shared env-isolation helper

**ID:** no-shared-env-helper
**Status:** active
**Chosen:** one `t.Setenv` in each of the two tests, each pinning a different variable to a
different value.
**Considered:** a `TestMain` or helper that clears `FLOWD_*`/`FLOW_ADDR` for a whole package —
rejected: the two pins want different values, no third caller exists, and a package-wide clear
would silently mask a future test that genuinely depends on the environment.

## Open questions

None.

## Out of scope

The real-guard test exercises whichever `flow` binary is on `PATH`, not the one built from the tree
under test. Pre-existing, unrelated to the two failures, and left as is.
