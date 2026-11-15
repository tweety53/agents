# kan-1124-guard-suite-runtime
> Design record of change kan-1124-guard-suite-runtime (KAN-1124), written at landing. Each decision's own **Status:** line is the status.

## Context

The approved design is round 3 of the planner session
(`.superpowers/sdd/planner/round-3.md`; answers in `answers-1.md`, `answers-2.md` and
`answers-3.md`; the measured evidence in `round-1.md` § Reachability evidence and `round-2.md`
§ Evidence). This file records it. The ticket, KAN-1124, decided all four parts: per-test
recording and fixing the slowest tests, `-race` in the full suite only, copied prepared
repositories, and parallel store, api and harvest tests.

It is one change because the four parts are measured together: the before and after figures are
one per-test profile per Go line and the `flow suite list` medians, and the per-test recording is
what makes the "fix the slowest" part checkable after it lands.

Every edit stays inside the `agents` repository: `stats/` Go code and tests, one migration, and
`.flow/project.md`'s `## test` section. Production behaviour changes in three places only:
`store.Open`'s DSN override (test-pool-size), `flow suite record` and `flow suite list`
(per-test-costs-in-flowd), and `cmd/flow`'s `run` environment parameter (cli-per-call-env), which
is a pure seam — `main` passes `os.Getenv`.

The baseline profile (base f3e458be, this machine, each line's own flags plus `-json`,
`GOMAXPROCS=4`): `guard-tests` 1:28.7 with `-race`, 1:11.1 without; `stats-go` 1:26.2 with
`-race`, 1:06.2 without. `flow suite list` medians since the smoke set: `guard-tests` 1m26s–2m21s,
`stats-go` 1m40s–2m24s. The ≥ 2 s rows are listed in `round-1.md`.

No prototype: nothing a user sees changes beyond `flow suite list`'s terminal output.

## Decisions

### Default runs drop -race; the full suite keeps it

**ID:** race-full-only-switch
**Status:** active
**Chosen:** both Go lines of `.flow/project.md`'s `## test` replace `-race` with
`${FLOW_VERIFY_FULL:+-race}`, so `-race` runs only under `run-verify-list -full`, which already
exports `FLOW_VERIFY_FULL=1`. `stats/internal/repotest/test_list_test.go` (its `-race` assertions
at `:34` and `:91`) and the `## test` prose that names `-race` (`:127-135`, `:159`, `:164-165`)
change with it.
**Considered:** a new `run-verify-list` flag — a second switch beside the `-full` one that already
means "every skip off"; a separate race-only suite line — a fifth suite recorded and run on every
default verify, which is the cost being removed.

### The race-detection core tests stay as they are

**ID:** race-core-tests-stay
**Status:** active
**Chosen:** the four `smoke: core` race-detection tests (sweep, fallback journal, harvest
rollout_subagent, reconcile) are left untouched; they detect races in `-full` runs only, as the
ticket's decision implies.
**Considered:** re-tagging or moving them — they would still need `-race` to detect anything, so
any placement outside `-full` either re-adds `-race` to a default run or proves nothing.

### Per-test costs are recorded in flowd with each suite run

**ID:** per-test-costs-in-flowd
**Status:** active
**Chosen:** operator answer round 1 Q1, option A. Both Go lines gain `-json`. `flow suite record`
reads the child's stdout and, for each line that is a `go test -json` event, prints the event's
`Output` text in its place, so the run log and `run-verify-list`'s `WARNING: DATA RACE` grep see
plain text as today. It computes each top-level test's cost by `flow-guard test-cost`'s rule and
sends the costs as a `Tests` field of the same suite-run request and journal entry; migration
0066 stores them in `suite_run_tests (suite_run_id, package, test, cost_ms)`. `flow suite list
-tests <suite>` prints the latest run's heaviest rows. A line whose stdout carries no `go test
-json` event records its suite row with no test rows; a stream cut short by a crash stores only
the tests that reported.
**Considered:** B — no persistence, profile once into this file and fix the slowest: no one sees a
later regression until someone re-profiles by hand, which is the state the ticket reports; C — a
per-run file beside `run-verify-list`'s report: worktree cleanup deletes it, so there is no
history across runs and no regression shows.

### The test-cost rule moves into its own package

**ID:** testcost-shared-package
**Status:** active
**Chosen:** the event fold and cost rule of `stats/internal/guard/testcost.go` move to a new
package `stats/internal/testcost`; `flow-guard test-cost` keeps its contract, flags and output
and calls it, and `cmd/flow`'s `suite record` calls it. Its tests move with the rule; the guard
keeps a test of its own CLI surface.
**Considered:** `cmd/flow` importing `internal/guard` for the rule — it already imports `guard`,
but the rule would then live behind a guard's private helper; a copy in `cmd/flow` — two copies of
one rule drift.

### A go test -json event is detected line by line

**ID:** json-stream-line-detection
**Status:** active
**Chosen:** `flow suite record` streams the child's stdout line by line (no line-length limit);
a line that decodes as a JSON object with a non-empty `Action` is a `go test -json` event — its
`Output`, when present, is written to stdout and the event is fed to the cost fold — and every
other line is written through unchanged. A Go line's stdout mixes `flow-guard unchanged`,
`smoke-keep` and `make` output with `go test`'s, so the detection is per line, not per stream.
Build failures arrive as `build-output` events and print through the same `Output` rule. Stdin
and stderr stay inherited as today.
**Considered:** a `-json` flag on `flow suite record` declaring the stream — the per-line test
already tells, and a flag is one more thing every caller must keep in step with its command;
buffering the whole stream before printing — the run log loses its live progress.

### Every non-skipped top-level test is stored

**ID:** store-every-test
**Status:** active
**Chosen:** each run stores one `suite_run_tests` row for every top-level test the rule prints
(pass or fail; a skipped test, and one with no terminal event, has none), with no minimum cost.
At about 2,400 rows per pair of Go runs and about 8 suite runs a day, that is roughly 3.6 M rows a
year — within what the dev Postgres holds without a retention policy. Rows cascade-delete with
their `suite_runs` row.
**Considered:** storing only tests above a cost floor — a test growing from below the floor shows
only once it crosses it, which hides the regression's start; a retention job — nobody asked for
one and the volume does not need it.

### flow suite list -tests reads the newest run that carries test rows

**ID:** suite-tests-read
**Status:** active
**Chosen:** `flow suite list -tests <suite>` reads `GET /api/v1/suites/{project}/tests?suite=<s>&limit=<n>`,
served by a store read of the newest `suite_runs` row of that suite that has `suite_run_tests`
rows, and prints one header line in `flow suite list`'s row format followed by that run's test
rows heaviest first, each `<cost seconds to two decimals>\t<package>\t<test>` as `flow-guard
test-cost` prints them; `-limit` (default 20) caps the rows and `-json` prints the response alone.
No run with test rows prints `no test costs recorded for <suite>` and exits 0; an unreachable
store exits 1 as `flow suite list` does.
**Considered:** extending the existing list route with every run's tests — every summary read
would carry thousands of rows nobody asked for; the newest run regardless of test rows — a
non-JSON line recorded after a JSON one would print nothing.

### Guard fixtures copy a prepared repository per shape

**ID:** fixture-template-helper
**Status:** active
**Chosen:** one shared helper in `stats/internal/guard/helpers_test.go` builds each recurring
repository shape once per package run under `execFixtures.dir` (`sync.OnceValues`,
`fixtureGitEnv`) and copies it per test with mode and mtime preserved. The shapes are bare init,
init plus an empty commit, and main with a bare origin and a clone, whose `origin` URL is
rewritten after the copy. Every file that builds one of those shapes step by step moves to it, and
the per-file masters (`pdsMaster`, `tmMaster`, `trBase`, `psF15Master`, `puc`) with a matching
shape fold in. Worktrees and per-test commits are still made after the copy, because gitdir paths
are absolute.
**Considered:** per-file masters for the slowest tests only — leaves the long tail of exec-bound
sub-second tests untouched.

### "The slowest" means every test at 2 s or more

**ID:** slowest-threshold
**Status:** active
**Chosen:** every top-level test costing ≥ 2 s in the post-change default profile of either Go
line (11 guard rows and 2 stats-go rows today) is cut below 2 s, or this file records why the cost
is the operation the test proves.
**Considered:** ≥ 1 s — 31 rows, most of them already sub-second-adjacent after the fixture and
race changes; a top-N count — an arbitrary N says nothing about whether a test is slow.

### Which tests run under t.Parallel

**ID:** parallel-scope-rule
**Status:** active
**Chosen:** every top-level test in the parallel packages calls `t.Parallel()`, except tests that
call `t.Setenv` or `t.Chdir`, directly or through a helper — Go panics on either under
`t.Parallel` — (harvest `TestDefaultZcodeRolloutRoot`; in `cmd/flow` the five `PATH` tests and
the three `t.Chdir` tests of `stage_test.go`) and storetest's own internal tests. Subtests
sharing the parent's store, server or fake stay serial. A test that flakes under parallel is fixed
at its source; no sleep or timeout is widened.
**Considered:** leaving timing-sensitive tests serial — keeps the flake's cause in place and the
package's serial floor with it.

### cmd/flow joins the parallel set

**ID:** cmd-flow-parallel
**Status:** active
**Chosen:** operator answer round 1 Q2, "Yes": `stats/cmd/flow` joins store, api and harvest,
under parallel-scope-rule.
**Considered:** exactly the three packages the ticket names — `stats-go` stays bounded by
`cmd/flow`'s ~40 s serial run.

### cmd/flow's run takes a per-call environment

**ID:** cli-per-call-env
**Status:** active
**Chosen:** operator answer round 2 Q3, "Yes". `run` takes a `getenv func(string) string`, and
`main` passes `os.Getenv`. Every CLI env read — `cmd/flow/state.go` (`:44`, `:68`, `:102`,
`:106`), `stage.go` (`:185`, `:196`, `:199`), `record.go:2077`, `fallback.StateRoot` — and every
`guard.Env{Getenv: …}` hand-off (`interactive.go:111`, `request.go:197`, `:329`,
`record.go:1238`, `:3847`, `stage.go:708`, `:892`, `:1021`) reads through it. This seam lands
before any test is rewritten. In the tests, `isolatedStateRoot` and the direct `t.Setenv` calls
become a per-test env map passed to `run` (`isolatedStateRoot`'s 327 call sites and the direct
calls across 18 test files), and those tests call `t.Parallel()`; the `PATH` tests stay serial,
because exec'd children inherit the process's `PATH`.
**Considered:** No — only the 120 tests that reach no `t.Setenv` call `t.Parallel()`:
<!-- measured: go test -json -count=1 ./cmd/flow/ split by whether a test reaches t.Setenv (planner round-2 evidence) @ merge-base f3e458be -->
`cmd/flow` stays at about 40 s serial and keeps bounding `stats-go`.

### The environment is an explicit parameter down to each reader

**ID:** getenv-explicit-param
**Status:** active
**Chosen:** `getenv` is passed as a parameter from `run` to each function that reads the
environment, and `fallback.StateRoot` and the `fallback` helpers built on it
(`StateFilePath`, the journal path, the state directory) take the root or the `getenv` from their
caller; `flowd`'s caller (`cmd/flowd/main.go:360`) passes `os.Getenv`.
**Considered:** a package-level `getenv` variable tests swap — parallel tests would race on it;
a `context.Value` — an untyped, hidden dependency, and several readers (`journalRecordWrite`,
`journalStageMark`, `policyCachePath`) take no context.

### The test pool is sized down through the DSN

**ID:** test-pool-size
**Status:** active
**Chosen:** `store.Open` keeps `MaxConns = 10` unless the DSN carries `pool_max_conns`, and
storetest's DSN carries `pool_max_conns=2`; `MinConns` never exceeds `MaxConns`. The worst case
is then 3–4 packages × 4 parallel tests × (2 + 1 admin) ≈ 48 connections, under the server's 100.
**Considered:** capping `-parallel` or `-p` — also slows the tests that never touch Postgres.

### Template copies take no lock

**ID:** template-copy-unlocked
**Status:** active
**Chosen:** no mutex around `CREATE DATABASE … TEMPLATE`. Postgres refuses a copy only while a
session is connected to the template, and copies run from the admin database while the
template's pool is already closed (`storetest.go` `buildTemplate`). The parallel run is the
proof; a refusal seen there gets the mutex.
**Considered:** serialising the copies pre-emptively — a lock against a refusal nobody has seen.

### No guard requires t.Parallel in new tests

**ID:** no-parallel-lint
**Status:** active
**Chosen:** no `check-*` guard enforces `t.Parallel` in new tests.
**Considered:** a new guard — nobody asked for it.

## Open questions

## Recommended approach

Five steps, in this order:

1. Record per-test costs (per-test-costs-in-flowd): the shared `internal/testcost` package
   (testcost-shared-package), then migration 0066 with the store, API, client and reconcile
   paths, then `flow suite record`'s line-by-line decode (json-stream-line-detection) and
   `flow suite list -tests` (suite-tests-read). Only once `flow suite record` decodes the stream
   do the Go lines gain `-json`, in the same edit that makes `-race` conditional
   (race-full-only-switch).
2. Size the test pool down (test-pool-size), then put store, api and harvest under `t.Parallel`
   (parallel-scope-rule, template-copy-unlocked).
3. Thread `getenv` through `cmd/flow` (cli-per-call-env, getenv-explicit-param), then rewrite
   its tests onto per-test env maps and make them parallel (cmd-flow-parallel).
4. Move `stats/internal/guard`'s tests onto copied repository shapes (fixture-template-helper).
5. Re-profile both Go lines from the recorded costs and fix every test still at ≥ 2 s
   (slowest-threshold), then update the docs.

Rejected alternatives are each decision's `**Considered:**` line: no persistence or a per-run
file (per-test-costs-in-flowd), the three-package parallel set (cmd-flow-parallel), the literal
`t.Setenv` exclusion (cli-per-call-env), per-file masters for the slowest tests only
(fixture-template-helper), a `-parallel` cap (test-pool-size).

## Critical files

- `.flow/project.md` — `## test`: both Go lines (`-json`, `${FLOW_VERIFY_FULL:+-race}`) and the
  prose naming `-race` and the re-profiling command.
- `stats/internal/repotest/test_list_test.go` — asserts the Go lines' flags.
- `stats/internal/guard/testcost.go`, `stats/internal/guard/test_cost_test.go` — the cost rule's
  current home and tests.
- `stats/internal/testcost/` — new package holding the cost rule.
- `stats/internal/store/migrations/0066_suite_run_tests.sql` — new table.
- `stats/internal/records/types.go` — `SuiteRun` gains `Tests`; the per-test row type.
- `stats/internal/store/suiteruns.go`, `stats/internal/store/suiteruns_test.go` — insert of the
  test rows with the suite row; the newest-run-with-tests read.
- `stats/internal/api/suites.go`, `stats/internal/api/server.go` — `ApplySuiteRunRecord`
  validation of `Tests`; the new `GET /api/v1/suites/{project}/tests` route.
- `stats/internal/client/suites.go` — client method for the tests read.
- `stats/internal/reconcile/reconcile.go` — journal replay of a suite run (carries `Tests`).
- `stats/cmd/flow/suite.go`, `stats/cmd/flow/suite_test.go` — `suite record` decode and cost
  send; `suite list -tests`; usage text.
- `stats/internal/store/store.go` — `store.Open` honours `pool_max_conns`.
- `stats/internal/storetest/storetest.go` — DSN carries `pool_max_conns=2`.
- `stats/internal/store/*_test.go`, `stats/internal/api/*_test.go`,
  `stats/internal/harvest/*_test.go` — `t.Parallel()`.
- `stats/cmd/flow/main.go` — `run`'s `getenv` parameter; `main` passes `os.Getenv`.
- `stats/cmd/flow/state.go`, `stage.go`, `record.go`, `interactive.go`, `request.go`,
  `policy.go`, `journal.go` — env reads and `guard.Env` hand-offs through `getenv`.
- `stats/internal/fallback/statefile.go` — `StateRoot` and the paths built on it take their
  root from the caller.
- `stats/cmd/flowd/main.go` — passes `os.Getenv` to the state-root lookup.
- `stats/cmd/flow/*_test.go` — the 18 files using `isolatedStateRoot` or `t.Setenv` move to
  per-test env maps; `t.Parallel()` everywhere parallel-scope-rule allows.
  <!-- measured: grep -l 'isolatedStateRoot(t)\|t.Setenv' stats/cmd/flow/*_test.go | wc -l @ merge-base f3e458be -->
- `stats/internal/guard/helpers_test.go` — the repository-shape template helper.
- `stats/internal/guard/*_test.go` — every file building a bare, empty-commit or
  origin-plus-clone repository step by step, and the files holding `pdsMaster`, `tmMaster`,
  `trBase`, `psF15Master` and `puc` (`check_panel_diff_size_test.go`,
  `check_plan_unchanged_test.go`, `check_plan_shape_test.go`, `check_tree_markers_test.go`,
  `check_task_records_test.go`).
- `README.md` — the `test-cost` mention under **Verify** and `flow suite list -tests`.

## Existing code to reuse

- `stats/internal/guard/testcost.go` `testCost` — the cost rule (own Elapsed plus paused
  descendants; skip and no-terminal excluded) that `internal/testcost` takes over.
- `stats/cmd/flow/suite.go` `runSuiteRecord`, `suiteJournalName`, `journalRecordWrite`,
  `callRecord`, `registerRecordConnFlags` — the record path the `Tests` field rides, store
  failure semantics included.
- `stats/internal/api/suites.go` `ApplySuiteRunRecord` — the one apply path for live and
  replayed suite runs.
- `stats/internal/store/suiteruns.go` `InsertSuiteRun`, `scanSuiteRunRow` — the suite row insert
  and scan the test rows join.
- `stats/internal/storetest` `NewStore`, `buildTemplate`, `createDatabase` — per-test databases
  copied from a migrated template.
- `stats/internal/guard/helpers_test.go` `execFixtures`, `fixtureGitEnv`, `writeExec` — the
  once-per-package fixture directory and fixed git identity the shape templates build under.
- `guard.Env{Getenv: …}` — the per-call environment seam the guards already use, which
  cli-per-call-env mirrors.
- `scripts/flow-guard run-verify-list` — already exports `FLOW_VERIFY_FULL=1` under `-full`,
  which `${FLOW_VERIFY_FULL:+-race}` reads.
- `flow suite list` and `flow-guard test-cost -min 1` — the before and after measurements.

## Verification

- Before any task: the baseline is the profile and medians in `## Context`.
- After the recording tasks: run `scripts/flow-guard run-verify-list <worktree> test` with `flow`
  on `PATH`; the log shows plain `go test` text, `flow suite list` shows the two Go suite rows,
  and `flow suite list -tests guard-tests` and `-tests stats-go` print the heaviest rows of that
  run. A reconcile test replays a journalled suite run carrying `Tests` and finds the test rows
  stored.
- A `-race` check: `scripts/flow-guard run-verify-list -full <worktree> test` runs both Go lines
  with `-race` (`ps`/log shows the flag) and records them as `guard-tests-full` and
  `stats-go-full`; the default run's command carries no `-race`.
- Parallel: `cd stats && go test -count=3 -short ./internal/store/ ./internal/api/ ./internal/harvest/ ./cmd/flow/`
  passes three times, and `go test -race -count=1 -short` over the same packages passes, so
  parallel tests are race-free; no run reports Postgres's `too many clients` error.
- Fixtures: `cd stats && go test -short -count=1 ./internal/guard/...` passes, and its `TestMain`
  leaves `git status` clean.
- After all tasks: `flow suite list -tests guard-tests` and `-tests stats-go` from a default run
  show no row at ≥ 2.00 s other than those `## Decisions` records under slowest-threshold, and
  both Go lines' `flow suite list` medians fall below the baseline in `## Context`.
- `## lint` from `.flow/project.md` and the full `## test` list pass.

## Live check

The change adds migration 0066 and a route to `flowd` and changes what `flow suite record` sends,
so the check runs the branch's `flowd` and `flow` against a copy of the real store, never the store
itself, from the change's worktree `<wt>`:

1. `scripts/flow-guard live-store-copy up kan-1124-live` from `<wt>`, then export the
   `FLOW_ADDR`/`FLOW_RECORDS_ADDR` lines it prints. Its `flowd` is the branch's, so starting it
   applies 0066 over the copy's real `suite_runs` rows.
2. `cd <wt>/stats && go build -o "$(flow state dir -C <wt>)/live-bin/flow" ./cmd/flow`, and put
   that directory first on `PATH`, so `run-verify-list` wraps both Go lines in the branch's
   `flow suite record`.
3. Before figures: `flow suite list -suite guard-tests -limit 1` (the copy's newest real row) and
   `flow suite list -tests guard-tests`, which prints `no test costs recorded for guard-tests` —
   no real row carries test rows yet.
4. `scripts/flow-guard run-verify-list <wt> test`. After figures: `flow suite list -tests
   guard-tests` and `flow suite list -tests stats-go` each print the new run's header line and
   its heaviest rows; `flow suite list -tests stats-go -json` holds one `tests` entry per
   non-skipped top-level test of that run; the two Go lines' kept logs (named in the run's
   `## Report`) hold no line starting `{"Action"`; the medians `flow suite list` prints for the
   two Go suites, recorded beside the `## Context` baseline.
5. `scripts/flow-guard live-store-copy down kan-1124-live`.

The change **not** working looks like: step 4's `-tests` read still printing `no test costs
recorded` (the decode or the send is broken, or 0066 never ran), a kept log carrying raw `go test
-json` events (the decode never reached the stream), `flowd` failing to start on the copy (0066
refuses the real rows), or a default run's Go command line carrying `-race`.

## Non-goals

- No web view of per-test costs; the SPA is untouched.
- No retention or pruning of `suite_run_tests` (store-every-test).
- No guard requiring `t.Parallel` in new tests (no-parallel-lint).
- No change to which tests the smoke set marks heavy or core, nor to the four race-detection core
  tests (race-core-tests-stay).
- No `t.Parallel` in packages outside store, api, harvest and `cmd/flow`; `repotest` and
  `setuptest` stay serial, and their slow tests are handled by slowest-threshold alone.
- No cwd seam for `cmd/flow`: the three `t.Chdir` tests stay serial.

## Risks

- **A parallel test exposes hidden shared state** (a package global, a shared fake server, a
  cache file under a shared root) and flakes. Limit: parallel-scope-rule fixes it at the source;
  the `-count=3` and `-race` runs in `## Verification` surface it before landing.
- **A CLI path execs a child that reads a `FLOW_*` variable from the process environment** rather
  than the per-call map (`stage.go:1279`'s wrapped child, `suite.go`'s child). A test relying on
  `t.Setenv` reaching that child would see the real environment once rewritten. Limit: such a
  test passes the variable through the child's `Env` built from `getenv`, or stays serial with
  `t.Setenv`; the rewrite task checks each exec site.
- **A Go line run without `flow` on `PATH`** (the `run-verify-list` fallback) writes raw
  `go test -json` into its log. The `WARNING: DATA RACE` grep still matches, since JSON keeps the
  phrase literal; the log is only harder to read. Limit: accepted — every flow machine has `flow`
  installed.
- **Postgres connection exhaustion** under parallel store tests. Limit: test-pool-size's ≈ 48
  connection ceiling; a refusal seen in the parallel run is fixed in storetest, not by capping
  `-parallel`.
- **Template copy refusal** under concurrent `CREATE DATABASE … TEMPLATE`. Limit:
  template-copy-unlocked's proof run; a refusal gets the mutex.
- **Copied repositories carry absolute paths** (`origin` URL, worktree gitdirs). Limit:
  fixture-template-helper rewrites `origin` after the copy and makes worktrees after it.
- **The measured gains vary by machine and load.** Limit: every figure is read from
  `flow suite list` and the recorded per-test costs on the same host, never compared across hosts.
- **`TestIntegrateChangeHappyPath` stays at 2 s or more** (decision slowest-threshold): measured 5.14/6.20 s at the merge-base (tasks.md Task 18's profile); no task changes its fixture, and its cost is the two-repository prepare, commit and land it proves, one chain no split can shorten. Limit: Task 18 Step 2 writes the measured median into its doc comment.
- Accepted: Task 20 Step 2 appends tasks during implementation for any row still at 2 s or more, an open-ended tail not sized here (plan-review-2 F2) — Task 18 Step 2 already cuts or justifies every known row, so the tail covers only rows the re-profile surfaces.
- Accepted: a Go line run without `flow` on `PATH` writes raw JSON (plan-review-2 F3, decoder in the repo) — every flow machine has `flow` installed; a second decoder adds code for a path no machine takes.
