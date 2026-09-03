# kan-375-stats-two-go-tests-fail-when-a-workspace-s

**Title:** stats: two Go tests fail when a workspace's FLOWD_PORT and FLOW_ADDR are exported
**Jira:** KAN-375

## Why

`/flow`'s verify stage runs `cd stats && go test ./... -race -count=1` with the worktree's
workspace variables exported, exactly as `prepare-workspace.sh` prints them (`FLOWD_DSN`,
`FLOWD_PORT`, `FLOW_ADDR`). Under that environment two tests fail; with the variables unset the
suite is green. Surfaced by KAN-374's verify stage on 2026-09-01, in files that branch did not touch.

- `stats/internal/config/config_test.go` — `TestFromEnvDefaultsPortWhenUnset` reads the real
  process environment and sees `FLOWD_PORT=5753` where it expects the variable absent
  (`got port 5753, want default 4173`).
- `stats/cmd/flow/record_test.go` — `TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard`
  runs the real `scripts/check-unfinished-work.sh`, which shells out to `flow record findings` on
  `PATH`; that subprocess honours the ambient `FLOW_ADDR` and dials the workspace port, where no
  daemon runs (`dial tcp 127.0.0.1:5753: connect: connection refused`, surfaced as the guard's exit
  2). With `FLOW_ADDR` unset the same subprocess dials `127.0.0.1:4173` — the dev daemon — so the
  test passes today only while the dev stack is up.

Both are test-environment assumptions, not product defects. The result is that the pipeline's own
verify stage cannot run the `## test` list as written without first unsetting two of the variables
it just exported.

## What changes

- `TestFromEnvDefaultsPortWhenUnset` pins `FLOWD_PORT` empty with `t.Setenv`; `config.FromEnv`
  already treats an empty variable as unset.
- `TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard` pins `FLOW_ADDR` to its own
  `httptest` daemon's URL with `t.Setenv`, so the guard's `flow record findings` subprocess reads
  from the test daemon rather than from whatever the caller exported or the dev daemon.
- No other test under `stats/` reads `FLOWD_PORT` or `FLOW_ADDR` from the ambient environment
  without pinning it first; the sweep found none.
- `cd stats && go test ./... -race -count=1` passes both bare and with `FLOWD_DSN`, `FLOWD_PORT`
  and `FLOW_ADDR` exported.

Design, decisions and rejected alternatives: `design.md` beside this file.
