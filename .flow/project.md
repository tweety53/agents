# flow project configuration — agents

Read by globally installed flow skills.

## apps

This repository is mostly the source of the flow skills, commands, and rules, installed
elsewhere by `setup.sh` — that half has no port and no URL. It also holds `stats/`, a
PostgreSQL-backed Go service (`flowd`) with an embedded React SPA, plus a thin `flow` CLI. Both
halves live in the one repo and are covered below.

| App | Repo root | Kind | URL | Notes |
|-----|-----------|------|-----|-------|
| flow sources | `/Users/tweety53/Projects/agents` | Bash + Python + Go + Markdown | — | The skills/commands/rules half. Verification is the guard scripts below plus a sandboxed `setup.sh` run. |
| flow stats daemon | `/Users/tweety53/Projects/agents/stats` | Go + React/Vite | `http://127.0.0.1:4173` | `flowd`, loopback-only. Backed by a dedicated `flow-postgres` container on host port 5433, independent of any other Postgres stack on this machine. Also the one application `## apps` names that a fix run's reload rule (`skills/flow/verify-and-handoff.md`) never reloads — the same protection, not a separate one. |

**This repository is Go + Python, with two bash files.** The `hooks/*.py` hooks run on Python 3, standard library only — `/usr/bin/python3`, no third-party imports, no pip, no network. The guards ported into `flow-guard` — among them `check-unfinished-work`, `check-guard-calls`, `check-finish-preflight`, `recover-guard-incident`, `resolve-base-branch`, `check-visual-trigger`, `resolve-visual-screenshots`, `compose-mockup-frames`, `measure-visual-properties`, `check-foreign-staged`, `check-main-checkout-drift`, `refresh-main-checkout`, `break-and-prove`, and `prepare-workspace` — are Go (`stats/internal/guard/`), every one called as `flow-guard <name>` through `scripts/flow-guard`, the one bash bootstrap, which runs the binary built from this checkout; `setup.sh` is the only other bash file and stays bash.

**Every new guard, and any new logic added to an existing guard, is written in Go in `flow-guard`** — `stats/internal/guard/<name>.go`, registered in `guard.Registry`, tested beside it, called as `flow-guard <name>`; there is no per-guard script. A new `check-*` guard lands with its Go test, `stats/internal/guard/<name, - replaced by _>_test.go`, in the same change — `check_guard_calls_test.go` for `flow-guard check-guard-calls` — and `flow-guard check-guard-calls`'s rule C fails lint while it is missing.

### artifact tree

The pipeline's artifact tree is `spectre/`.

## run

To exercise the installer without touching the real home directory:

```bash
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
```

**The stats daemon.** Bring up the dedicated Postgres stack, then build and run `flowd`:

```bash
cd stats && docker compose up -d          # flow-postgres on host port 5433
cd stats && make build                    # builds the SPA, then both binaries into bin/
cd stats && ./bin/flowd
```

`flowd` binds `127.0.0.1:4173` (override with `FLOWD_PORT`) and refuses to start on any other
interface or on an unparsable `FLOWD_PORT` rather than defaulting silently. Running it this way is
for manual, foreground verification only — for a daemon that survives logout and restarts on
failure, see `stats/README.md`'s "Running the daemon at login" section and its launchd agent.
**No skill loads that agent**; loading it is an operator step, deliberately. **No automated stage runs these three commands either** — see the
UI-test stack below for what an automated stage uses instead.

**The UI-test stack**, for ad-hoc testing from the main checkout rather than from an apply worktree:
`make ui-test-up` and `make ui-test-down` (`stats/Makefile`) bring a second, disposable `flowd` up
on port 4174 against `flow_uitest`, seeded with a fixed fixture, and tear it down again. Point a
session at it with `FLOW_ADDR=http://127.0.0.1:4174` and
`FLOW_RECORDS_ADDR=http://127.0.0.1:4174` — the record family resolves its own address (see the
isolation section below), so pointing only `FLOW_ADDR` at 4174 leaves `flow record` reads on the
protected daemon. Neither target is isolated by the
`## workspace isolation` section below — that section covers apply worktrees, and this stack is a
single, main-checkout-only fixture instead.

**A live check against real data runs on a copy of the dev store**, never on the store itself:
`scripts/flow-guard live-store-copy up <name> [<port>]` copies the dev `flow` database into
`flow_<name>_live`, builds `flowd` and `flow` from the checkout it runs from and starts it on `<port>`
(default 4175) against the copy, printing the `FLOW_ADDR`/`FLOW_RECORDS_ADDR` exports and a `PATH` led by the copy's directory, so its `flow` is the one run, that point a
session at it and registers a scratch change named `<name>` in the copy — a STARTED state record and `.superpowers/sdd/<name>/` in that checkout — so `flow record -change <name>` writes there; `scripts/flow-guard live-store-copy down <name>` stops that daemon and drops only that copy.
It refuses ports 4173 and 4174, and its Go file's leading comment (`stats/internal/guard/livestorecopy.go`) carries its exit codes. A design's `## Live check`
names it rather than re-specifying the copy.

**This is the stack `flow.visual-verify` (`skills/flow/visual-verify.md`) starts and stops.**
That stage's step 5 probes a URL and, if nothing answers, "starts the stack from `## run`" — a
project with more than one candidate stack has to say which one that means, so this paragraph is
the answer: `make ui-test-up` / `make ui-test-down` against `http://127.0.0.1:4174`, matching the
`## visual verification` section below and `stats/web/playwright.config.ts`'s pinned `baseURL`.
**Never `flowd` on `127.0.0.1:4173`** — that is the dev workspace's protected daemon (`## stop`
below), its data changes on every run so no baseline over it could ever be stable, and `CLAUDE.md`
forbids any agent action touching it at all.

## test

```bash
# suite: guard-tests
scripts/flow-guard unchanged . ':!docs/' ':!spectre/' ':!stats/web/' || { keep=$(scripts/flow-guard smoke-keep stats/go.mod=stats/ stats/go.sum=stats/ hooks/=stats/internal/repotest/ hooks/=stats/internal/setuptest/ scripts/=stats/internal/repotest/ scripts/=stats/internal/setuptest/ setup.sh=stats/internal/setuptest/) && cd stats && FLOW_SMOKE_KEEP="$keep" GOMAXPROCS=$(( ($(getconf _NPROCESSORS_ONLN) + 2) / 3 )) go test -json ${FLOW_VERIFY_FULL:+-race} -short -count=1 ./internal/guard/...; }
# suite: stats-go
scripts/flow-guard unchanged . ':!docs/' ':!spectre/' ':!stats/web/' || { keep=$(scripts/flow-guard smoke-keep stats/go.mod=stats/ stats/go.sum=stats/ hooks/=stats/internal/repotest/ hooks/=stats/internal/setuptest/ scripts/=stats/internal/repotest/ scripts/=stats/internal/setuptest/ setup.sh=stats/internal/setuptest/) && cd stats && make -s web-dist && FLOW_SMOKE_KEEP="$keep" GOMAXPROCS=$(( ($(getconf _NPROCESSORS_ONLN) + 2) / 3 )) go test -json ${FLOW_VERIFY_FULL:+-race} -short -count=1 $(go list ./... | grep -v '/internal/guard$'); }
# suite: stats-spa
# inputs: stats/web/ stats/Makefile
scripts/flow-guard unchanged stats/web/ stats/Makefile || { cd stats && make -s web-dist && cd web && npm test; }
# suite: flow-task-list
# inputs: mods/flow-task-list/
scripts/flow-guard unchanged mods/flow-task-list/ || claude plugin test mods/flow-task-list
```

The four suites run at once, so each Go suite runs on a third of the cores, rounded up —
`GOMAXPROCS` bounds the go command's build parallelism, its default `-p`, and each test binary's
threads and `-parallel` — beside the SPA's two worker threads (`stats/web/vite.config.ts`), so no
suite starves another.

Each suite runs only when the branch changed, since its merge-base on `origin/main`, a path it
reads (`scripts/flow-guard unchanged`): a change to `docs/` or `spectre/` alone runs no suite, and one to
`stats/web/` alone runs only the SPA's. A suite its guard skipped prints the guard's line in the
`## Report` and records no suite row.

**This list is the smoke set — the default run.** Both Go lines run under `-short`, and three kinds
of test run in it:

- **every light test.** A test is heavy when its cost is 1 s or more — its own `Elapsed` plus that
  of every descendant that paused for `t.Parallel()`, under the line's own flags less `-short` on an
  idle machine. A heavy test opens with `smoke.Heavy(t)` (`stats/internal/smoke`), which skips it
  in a `-short` run; a new test carries no mark and runs.
- **the core set.** A core test carries `// smoke: core — <the capability it proves>` as the last
  line of the comment directly above its `func` and never calls `smoke.Heavy`;
  `git grep -n '^// smoke: core' -- stats ':!stats/internal/smoke'` lists the set with its reasons,
  and the static test of `stats/internal/smoke` fails a test carrying both the tag and the call.
- **the heavy tests of what the branch touched.** `flow-guard smoke-keep` reads the diff
  `flow-guard unchanged` reads and prints the paths `FLOW_SMOKE_KEEP` carries to `smoke.Heavy`: a
  changed `_test.go` file, the test files paired by name with a changed Go source, and the kept
  path of each `<path prefix>=<kept path>` argument a changed path falls under — the rows on the
  two Go lines. A changed Go source that pairs with no test file keeps its package directory and is named on
  stderr, at the top of its suite's block.

**A core test is gathered by one rubric.** Keep a heavy test only when it is the cheapest
success-path proof — through the registered command, the CLI's `run` or a real script — of
something every run executes, of a gate whose failure is silent, or of an operation that rewrites
or destroys; when no lighter test covers that path and `## lint` does not already run that guard
over the real tree; and when no other core test already runs it for real. A heavy test that is the
only reader of a repository input nearly every change touches is kept under the same rule, in
place of a keep row on that input.

**The full suite is `scripts/flow-guard run-verify-list -full <worktree> test`**, the same list
with every skip off: it exports `FLOW_VERIFY_FULL=1`, under which `flow-guard unchanged` answers
"changed" without looking and `smoke.Heavy` skips nothing, skips no command on a green run and
runs the visual `verify` row whole; under it both Go lines also run `-race`. It runs only when the operator asks for it, never unasked as
part of a `/flow` run. A command it fails that the default list passed at the same HEAD prints
`escaped the default list — it passed at <sha>` in its block: the fix tags the failing test core,
or a cheaper sibling proving the same behaviour, by the rubric above.

**Re-profiling is `flow suite list -tests guard-tests` and `flow suite list -tests stats-go`
after a default run**: every non-skipped top-level test's cost, recorded with the run, heaviest
first, read against the 1 s heavy mark and the 2 s ceiling — a test at 2 s or more is cut below it,
or its doc comment records why its cost is the operation it proves. A hand profile is still the
line's command piped into `scripts/flow-guard test-cost -min 1`. A marked test is skipped and so
not printed; what prints is the core set plus whatever grew heavy — each new row meets the rubric
and is tagged, or gains the `smoke.Heavy(t)` call.

**`/flow-fast` and `/flow-interactive` scope these commands by hand**, and `smoke.Heavy` follows
the scope: a command naming tests with `-run` runs every test it names whatever else it carries,
and one scoped to a whole package drops `-short`.

`stats/internal/web/dist/` is gitignored and `stats/internal/web/embed.go`'s `//go:embed all:dist`
refuses to compile without it, so `make -s web-dist` (`stats/Makefile`) opens every line here and in
`## lint` that compiles `internal/web` or needs `stats/web/node_modules`: it runs `make web-build`,
`npm ci && npm run build` in `stats/web`, only when the build is absent.

Only the SPA and `flow-task-list` lines declare `# inputs:` (`skills/flow-contracts/project-configuration.md`): the Go
tests read files across the whole repository — `stats/internal/guard`'s tests
open `skills/`, `scripts/` and `setup.sh` through `../../../` — so both always run.

The `flow-task-list` suite runs the mod's tests on the harness's own runner: `claude plugin test <dir>` runs every `*.test.ts` and `*.test.tsx` under `<dir>` in an environment like the one the mod's hooks run in and supplies the `claude-code/testing` module they import, which nothing else in this repository resolves.

The `guard-tests` suite is `go test` over `stats/internal/guard/` — every guard's tests, a new one
picked up with no edit here; the tests that exec a guard build `flow-guard` once per package run
into a temp `FLOW_GUARD_CACHE_DIR`, never the operator's cache. The package's `TestMain` fails the run when it leaves the
repository's `git status` changed — a test or reproducer writing into the checkout. The `stats-go` line excludes
`internal/guard`, which `guard-tests` runs, so the package runs once per verify. A new `check-*` guard lands with its Go test in the same change — plans declare it and
review panels look for it, and `flow-guard check-guard-calls`'s rule C fails `## lint` while it is
missing.

**Measured runtime: recorded per run in the flow store — query it with `flow suite list`.** The
canonical suites are `guard-tests` (`go test -json ${FLOW_VERIFY_FULL:+-race} -short -count=1 ./internal/guard/...`), `stats-go`
(`go test -json ${FLOW_VERIFY_FULL:+-race} -short -count=1 $(go list ./... | grep -v '/internal/guard$')`), `stats-spa` (`stats/web` `npm test`) and `flow-task-list` (`claude plugin test mods/flow-task-list`); each row carries the
machine that ran it, and the per-(suite, host) summary line is the median of the last 10 passing
runs. Both Go lines print test2json, which `flow suite record` turns back into plain output and
folds into the run's per-test costs. Whether the whole `## test` list still fits inside this harness's default tool timeout in one
invocation is answered by those recorded figures, not by a number pasted here. Each command a `# suite:` line names is
recorded so on every `flow-guard run-verify-list <worktree> test` that runs it, and as `<name>-full`
— the visual row as `visual-full` — on a `-full` run, so no median mixes the two; record any other run with
`flow suite record -suite <name> -- <command>` — a store that cannot be reached costs one warning
line and never changes the suite's own exit code.

## lint

```bash
printf 'stats/web/src/App.tsx\n' | scripts/flow-guard check-visual-trigger .
scripts/flow-guard resolve-visual-screenshots . baseline.spec.ts
scripts/flow-guard check-guard-calls
cd stats && gofmt -l .
cd stats && make -s web-dist && go vet ./...
cd stats && make -s web-dist && cd web && npx tsc -b
```

**There is no auto-fix command for the guard scripts** (`scripts/check-*`) — every one of them
reports `file:line` and is fixed by editing the offending line, never by weakening the guard or
adding a suppression marker to silence a real hit. **`stats/` does have one**: `cd stats && gofmt
-w .` reformats Go source before the `gofmt -l .` check above is run, per the Lint Fix Priority
rule's "run the auto-fix command first" step. There is no equivalent for the SPA — `web/package.json`
carries no lint or format script, only `tsc -b`'s type check, so a TypeScript violation is fixed by
hand like a guard-script one.

**Every guard in the list is currently expected to exit 0.**

## stop

**This key declares no command, and that is the answer rather than an omission.** `/flow`'s
no `/flow` step runs this key before removing a worktree; there is
nothing here for a stack start or restart to run.

**What is protected is the dev workspace's service and its storage, and only those** — the empty-id
case of `## workspace isolation` below, which is what the main checkout resolves:

| Protected — never stopped, dropped or removed by any agent action | Why |
|---|---|
| `flowd` on `127.0.0.1:4173` | the store every `flow` call in every project writes through |
| the `flow-postgres` container on host port 5433 | the **service**, shared by every workspace |
| the default `flow` database inside it | the dev workspace's own storage |

Not `docker compose down`, not `launchctl unload`, not a `kill` on the daemon's pid, and not to make
a later step succeed. Stopping any of them mid-run silently degrades the rest of that run's writes to
the on-disk journal and takes the store away from every other session on this machine; restarting
afterwards does not repair it, because whatever fell through to the journal while it was down stays
there.

**The one sanctioned restart is the one a landing runs.** After a landing that moved `stats/`,
`scripts/flow-guard pipeline-fix land` and `scripts/flow-guard refresh-main-checkout` run `make -C <main>/stats
restart` — after the push, from the landed code, at the end of the run that landed it — so the live
flowd and the installed `flow` CLI never keep the build the landing replaced.

**A workspace's own derived resources are not protected, and removing them is correct.** The
`flow_<id_underscored>` database and the bucket an apply worktree derives are per-change artifacts:
`scripts/flow-guard workspace remove <id>` drops them during the cleanup run exactly as the registry
requires, and that must keep working. The line is the one **Workspace isolation**
(`skills/flow-contracts/workspace-isolation.md`) already draws — what is isolated is the logical
resource, never the service that holds it. This section protects the service and the dev workspace's
own logical resources; it says nothing about anyone else's.

**Nothing worktree-local exists to stop.** A worktree gets its own database and port, and
no `/flow` step starts a daemon against them.

Stopping the dev stack is an **operator** action, deliberately. The commands live where the operator
starts it: `## run` above, and `stats/README.md`'s "Running the daemon at login" section for the
launchd agent.

## review panel citation check

```bash
```

## scripts

Every new script in this repository is written in Go: its logic goes in the `flow-guard` binary
(`stats/internal/guard/<name>.go`, registered in `guard.Registry`, with its cases in a Go test
beside it), called as `flow-guard <name>`. `scripts/flow-guard` is the one bash bootstrap that
builds and runs that binary; it and `setup.sh` stay bash, and the patches landed changes carried
are readable only in history.

## standards

- `CLAUDE.md`

## jira

KAN

## default landing route

merge and push

## handoff

none

## decisions

recommended

## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| `database` | `FLOWD_DSN` | `postgres://flow:flow@localhost:5433/flow?sslmode=disable` | `postgres://flow:flow@localhost:5433/flow_<id_underscored>?sslmode=disable` |
| `port` | `FLOWD_PORT` | `4173` | `+<offset>` |
| `url` | `FLOW_ADDR` | `http://127.0.0.1:4173` | `http://127.0.0.1:<value:FLOWD_PORT>` |
| `url` | `FLOW_RECORDS_ADDR` | `http://127.0.0.1:4173` | `http://127.0.0.1:4173` |

**The `FLOW_RECORDS_ADDR` row is deliberately not isolated, and its token-free workspace cell is
the statement of that.** The record family (`flow record`, `flow stage`, `flow state`, `flow policy`, `flow jira`, `flow self-review finding`, `flow self-review findings`, `flow self-review bundle`, `flow suite record`, `flow suite list`, `flow change save`, `flow change file`, `flow interactive land`) resolves
its store address from it, so an apply worktree's dispatch, finding and suite-runtime rows land in the
persistent store the main checkout serves, and a deferred self-review bundle still finds them
after `scripts/flow-guard workspace remove` has dropped the `database` row's resource; stage marks, state
rows, settings and Jira transitions must outlive the worktree's daemon as the records do. The cell could not use the `database` word: that row is
taken, and cleanup removes what it names — the persistent store must never be a removal target.

| Command | Runs |
|---------|------|
| `create` | `scripts/flow-guard workspace create <id>` |
| `remove` | `scripts/flow-guard workspace remove <id>` |

**`FLOW_STATE_DIR` and `FLOW_TRANSCRIPTS_DIR` are deliberately not isolated.** The `Resource`
column above is a closed vocabulary — `database`, `bucket`, `cache index`, `port`, `url`, and no
other word — and a directory path is none of those five, so neither variable gets a row; this is a
decision, not an oversight. The consequence is bounded: every worktree's `flowd` still harvests
from the one real transcripts root, but each writes the result into its *own* isolated database from
the table above, so the harvested data ends up duplicated across worktrees rather than shared through
one database. Neither variable crosses the boundary this section exists to protect.

## visual verification

`stats/web` is the one SPA this repository carries. Its baseline covers the dashboard and
run-detail views, per `stats/web/tests/visual/baseline.spec.ts`, against the disposable UI-test
stack on `http://127.0.0.1:4174` — `## run`'s "This is the stack `flow.visual-verify` … starts and
stops" paragraph is the answer to where that stage gets its stack from, and is not restated here. `ui paths` also names `skills/flow-contracts/model-policy.md`, which `stats/Makefile`'s `policy-copy` embeds into flowd for the Agent roles view.

**The `--app` argument is `<worktree>:stats/web=http://127.0.0.1:4174`.** `flow-guard
visual-verify-pass prepare` takes the app root
`stats/web`, never `stats`: Go sources under `stats/` name the `flowd` default
`http://127.0.0.1:4173`, so check 2 fails on `stats` by construction.

**No `regression checkout` row** — this repository commits its own baselines to the change's own
branch, per `design.md`'s `agents-has-no-regression-checkout` decision. **No `push to default
branch` row** either, so nothing declared here can push anywhere.

**`capture` carries `--update-snapshots`, measured rather than assumed.** `capture` creates this
change's baseline, so it must be the Playwright invocation that succeeds when writing a PNG that
does not yet exist — bare `npx playwright test <spec>` does not: Playwright writes the file but
still exits non-zero, "A snapshot doesn't exist … writing actual", which would block every single
run on the intended path. Measured directly against the installed `@playwright/test` 1.62.1: the
bare `--update-snapshots` flag (no explicit mode) is what actually exits 0 on that first-run write
— its documented "changed" preset. The explicit `--update-snapshots=missing` mode was tried too and
still exits non-zero on the identical case, so it is not the mechanism this project uses.
<!-- measured: rm -rf tests/visual/<name>.spec.ts-snapshots; npx playwright test tests/visual/<name>.spec.ts [flags]; echo $? @ 2026-08-27, stats/web, @playwright/test 1.62.1 -->
`verify` above carries no such flag — it must keep failing on real drift against the committed
baseline, which is the whole point of a regression gate.

**`fingerprint` compares `index.html`, rebuilt first.** Vite content-hashes its asset filenames and
rewrites `index.html` to reference them, so the served `index.html` names the exact bundle the
daemon embeds, and `web.Handler` serves that embedded file byte-for-byte at `/`. The `npm run
build` in front makes the comparison mean "the worktree's source", not "whatever `dist/` last
held"; `make ui-test-up` rebuilds too, so a stack this stage started matches first time and only a
reused stack pays the restart. It then requires `skills/flow-contracts/model-policy.md`'s bytes inside
`/tmp/flow-uitest-flowd`, the binary `make ui-test-up` built and runs (`UITEST_BIN` in
`stats/Makefile`): `//go:embed` stores the file verbatim, and no `index.html` byte changes when
that file does, so without it a reused stack would render the Agent roles view from the old
policy. It last compares `make go-sources-hash` with the hash `make
ui-test-up` recorded in `/tmp/flow-uitest-go.sha256`, since neither `index.html` nor the policy
changes when a Go source or the seed does, so a stack built before a Go-side fix reads stale.

**`capture` and `verify` check the stack first.** Each opens on `flow-guard check-dev-stack-fresh`, the declared `fingerprint` run, and restarts the disposable UI-test stack with `make ui-test-up` when it is stale, so an implementer's or fixer's capture never reads a pre-fix build as a failure; `make ui-test-up` touches only that stack, never the `flowd` on 4173.

**`verify` runs the smoke set.** No `specs` row is declared, so `flow-guard run-verify-list` fills
`<specs>` with `--grep @smoke`: the tests of `stats/web/tests/visual/baseline.spec.ts`, each
describe tagged `@smoke`. The whole SPA suite, `<specs>` empty, runs only when the operator asks
(**Inline verify**, `skills/flow/verify-and-handoff.md`).

**`capture` rebuilds `full-app-suite.zip` after Playwright.** The row runs Playwright inside a shell
function, so the arguments a caller appends — `--update-snapshots=all`, `-g '<titles>'` — still
reach Playwright, and on its exit 0 runs `flow-guard rebuild-full-app-suite-zip` with the worktree and the
`screenshots` root resolved absolute, so no recapture leaves the zip trailing its PNGs
(group-allowed-pairs-roles self-review).

| Setting | Value |
|---------|-------|
| `ui paths` | `stats/web/src/**`, `skills/flow-contracts/model-policy.md` |
| `screenshots` | `stats/web/tests/visual` |

| Command | Runs |
|---------|------|
| `setup` | `cd stats/web && npm install && npx playwright install chromium` |
| `smoke` | `cd stats/web && npx playwright test --list` |
| `verify` | `cd stats/web && { ../../scripts/flow-guard check-dev-stack-fresh ../.. \|\| make -C .. ui-test-up; } && npx playwright test <specs>` |
| `capture` | `cd stats/web && { ../../scripts/flow-guard check-dev-stack-fresh ../.. \|\| make -C .. ui-test-up; } && capture() { npx playwright test "$@" && ../../scripts/flow-guard rebuild-full-app-suite-zip "$(cd ../.. && pwd)" "$PWD/tests/visual"; } && capture <spec> --update-snapshots` |
| `fingerprint` | `cd stats/web && npm run build && curl -sf http://127.0.0.1:4174/ \| cmp -s - ../internal/web/dist/index.html && python3 -c 'import sys; sys.exit(open(sys.argv[1], "rb").read() not in open(sys.argv[2], "rb").read())' ../../skills/flow-contracts/model-policy.md /tmp/flow-uitest-flowd && make -s --no-print-directory -C .. go-sources-hash \| cmp -s - /tmp/flow-uitest-go.sha256` |

## progress

quiet
