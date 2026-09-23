# flow project configuration — agents

Read by globally installed flow skills.

## apps

This repository is mostly the source of the flow skills, commands, and rules, installed
elsewhere by `setup.sh` — that half has no port and no URL. It also holds `stats/`, a
PostgreSQL-backed Go service (`flowd`) with an embedded React SPA, plus a thin `flow` CLI. Both
halves live in the one repo and are covered below.

| App | Repo root | Kind | URL | Notes |
|-----|-----------|------|-----|-------|
| flow sources | `/Users/tweety53/Projects/agents` | Bash + Python + Markdown | — | The skills/commands/rules half. Verification is the guard scripts below plus a sandboxed `setup.sh` run. |
| flow stats daemon | `/Users/tweety53/Projects/agents/stats` | Go + React/Vite | `http://127.0.0.1:4173` | `flowd`, loopback-only. Backed by a dedicated `flow-postgres` container on host port 5433, independent of any other Postgres stack on this machine. Also the one application `## apps` names that a fix run's reload rule (`skills/flow/verify-and-handoff.md`) never reloads — the same protection, not a separate one. |

**This repository is Bash + Python, not Bash-only.** `scripts/check-plan-provenance.sh` is a thin
wrapper that execs `scripts/check-plan-provenance.py` (Python 3, standard library only —
`/usr/bin/python3`, no third-party imports, no pip, no network) so its fence/container classifier
can be a real block-structure parser instead of a hand-rolled Bash ERE allowlist.

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
scripts/run-guard-tests.sh
cd stats && go test ./... -race -count=1
cd stats/web && npm test
```

`scripts/run-guard-tests.sh` discovers every `scripts/test-*.sh` by glob and runs them concurrently
through `scripts/lib/parallel.sh`; a new harness added to `scripts/` is picked up automatically, with
no edit needed here.

**Measured runtime: recorded per run in the flow store — query it with `flow suite list`.** The
canonical suites are `guard-tests` (`scripts/run-guard-tests.sh`), `stats-go`
(`go test ./... -race -count=1`) and `stats-spa` (`stats/web` `npm test`); each row carries the
machine that ran it, and the per-(suite, host) summary line is the median of the last 10 passing
runs. Whether the whole `## test` list still fits inside this harness's default tool timeout in one
invocation is answered by those recorded figures, not by a number pasted here. Record a run with
`flow suite record -suite <name> -- <command>`; a store that cannot be reached costs one warning
line and never changes the suite's own exit code.

**`check-installed-citations.sh` (named in `## lint` below) is unlike every other guard in that
list: it shells out to a sandboxed `setup.sh` twice per invocation** — once for `global`, once for
`all` — to derive the installed set it classifies citations against, rather than only reading
files already on disk. A single invocation measures about 0.84s — worth naming here since it is the one guard in this repository paying for a subprocess
rather than a plain file scan.
<!-- measured: time scripts/check-installed-citations.sh >/dev/null @ branch kan-102-citations-resolve-to-installed-paths -->

## worktree setup

```bash
cd stats && make web-build
```

`stats/internal/web/dist/` is gitignored and `stats/internal/web/embed.go`'s `//go:embed all:dist`
refuses to compile without it, so a fresh worktree's first `go test ./...` or `go build ./...`
fails until the SPA is built once. `make web-build` is `npm ci && npm run build` in `stats/web`
(`stats/Makefile`), the same prerequisite `make test` and `make build` already carry.

## lint

```bash
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-python-suppressions.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
scripts/check-plan-shape.sh
scripts/check-task-records.sh
scripts/check-workspace-isolation.sh
scripts/check-visual-verification.sh .
printf 'stats/web/src/App.tsx\n' | scripts/check-visual-trigger.sh .
scripts/check-spec-reach.sh .
scripts/resolve-visual-screenshots.sh . baseline.spec.ts
scripts/check-uitest-overrides.sh
scripts/check-contract-budget.sh
scripts/check-markdown-integrity.py
scripts/check-stage-mark-calls.sh
scripts/check-guard-symlinks.sh
scripts/check-dispatch-paragraphs.sh
scripts/check-mutation-reproducer-pin.sh
scripts/check-self-review-report.sh
scripts/check-installed-citations.sh
scripts/check-installed-rules.sh
scripts/check-normative-inventory.sh
scripts/check-model-keys.sh
scripts/check-model-resolution-shell.sh
scripts/check-worktree-location.sh "$(git worktree list --porcelain | awk '/^worktree /{print substr($0,10); exit}')"
cd stats && gofmt -l .
cd stats && go vet ./...
cd stats/web && npx tsc -b
```

**There is no auto-fix command for the guard scripts** (`scripts/check-*`) — every one of them
reports `file:line` and is fixed by editing the offending line, never by weakening the guard or
adding a suppression marker to silence a real hit. **`stats/` does have one**: `cd stats && gofmt
-w .` reformats Go source before the `gofmt -l .` check above is run, per the Lint Fix Priority
rule's "run the auto-fix command first" step. There is no equivalent for the SPA — `web/package.json`
carries no lint or format script, only `tsc -b`'s type check, so a TypeScript violation is fixed by
hand like a guard-script one.

**`check-contract-budget.sh` is a ratchet, not a target.** It fails when an owned `.md` or `.mdc`
file — every one under `skills/`, `rules/`, `spectre/specs/`, `commands/`, `commands-claude/`,
`.flow/` and the repository root, resolved through `scripts/lib/owned-corpus.sh` — outgrows the
budget declared for it in the guard's own `budgets()` table, or carries no budget at all. The table
is keyed on the path relative to the repository root, not on the bare basename, because every skill
directory has a file literally named `SKILL.md` and a basename key would collide across skills. Each
budget is the size its file had when the change that added its row landed, plus 25% — so ordinary
edits pass and a real section addition trips it, forcing a deliberate edit to the table rather than a
silent regrowth of a file every `/flow*` command loads. Raising a budget is the correct response
to a genuine addition; narrowing the guard's scope or deleting a row is not.

**`check-normative-inventory.sh` reports a set rather than a verdict.** It prints every sentence in
this repository's owned Markdown that carries `SHALL`, `SHALL NOT`, `MUST` or `MUST NOT` as a whole
word — one per line, whitespace-normalised, sorted — and its exit codes are only `0` the inventory
printed and `2` it cannot answer. It has no violation code deliberately: comparing two inventories
is the caller's act, so a change that edits prose in bulk captures the output before its first edit
and diffs the output after its last against it, and resolves any difference by restoring the
sentence rather than by accepting the new inventory. Unlike every other guard here it writes its
payload to stdout (its file and sentence counts go to stderr), so a lint run sees roughly a
thousand lines from it and no verdict line. It resolves the corpus through
`scripts/lib/owned-corpus.sh`, which `check-contract-budget.sh` calls too, so the two guards
cannot disagree about which files this repository owns.

**`check-workspace-isolation.sh` is a lint step where the other `## workspace isolation` guard is
not.** It takes a project root, defaults to this repository when given none, and answers a question
about the text of `.flow/project.md` — so it runs against a bare tree like every other guard here.
`check-cleanup-complete.sh` reads the same section and is excluded below for the opposite reason: it
needs a change in flight.

**`check-installed-rules.sh` is the one guard in this list that reads outside the repository.**
It compares the always-on rules this checkout declares against what `setup.sh global` last installed
under `$HOME` — the symlinks in `~/.claude/rules/` and the per-rule markers in the two managed
blocks — because a rule can merge and stay unreadable by every session. It still runs against a bare tree and takes no
change-in-flight state, so it belongs here for `check-installed-citations.sh`'s reason. A machine
with no global install is not a failure: it reports `INSTALLED-RULES-NONE` and exits 0, which is what
keeps it runnable in CI and a fresh clone. A partial install is a failure.

**Every guard in the list is currently expected to exit 0.** `check-workspace-isolation.sh` reports
`ISOLATION-OK` and validates this repository's own declared section; its own header carries its
full exit-code contract: 0 every project checked is well formed, 1 violations found, 2 it cannot
answer at all. `check-plan-provenance.sh` reports
"all provenance stated" — see the script's own header for the full exit-code contract: 0
clean/nothing-in-flight, 1 violations found, 2 environment, 3 containment, 4
content-classification; a caller that treats "non-zero" uniformly, as this repository's own lint
step does, is unaffected by that split. A future non-zero exit is a real hit on a plan in flight:
fix the offending line by stating its provenance, never by narrowing the guard's scope or adding a
suppression marker.

## stop

**This key declares no command, and that is the answer rather than an omission.** `/flow`'s
worktree-cleanup check 5 runs whatever this key declares before removing a worktree; there is
nothing here for it to run.

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

**A workspace's own derived resources are not protected, and removing them is correct.** The
`flow_<id_underscored>` database and the bucket an apply worktree derives are per-change artifacts:
`scripts/workspace.sh remove <id>` drops them during archive cleanup exactly as the registry
requires, and that must keep working. The line is the one **Workspace isolation**
(`skills/flow-contracts/workspace-isolation.md`) already draws — what is isolated is the logical
resource, never the service that holds it. This section protects the service and the dev workspace's
own logical resources; it says nothing about anyone else's.

**Nothing worktree-local exists for check 5 to stop.** A worktree gets its own database and port, and
no `/flow` step starts a daemon against them. Check 6, `check-worktree-processes.sh`, is what proves
no process holds a worktree, and it remains a gate.

Stopping the dev stack is an **operator** action, deliberately. The commands live where the operator
starts it: `## run` above, and `stats/README.md`'s "Running the daemon at login" section for the
launchd agent.

## review panel citation check

```bash
scripts/check-references.sh
```

## standards

- `CLAUDE.md`
- `AGENTS.md`

## jira

`KAN`

## default landing route

`merge and push`

## handoff

`none`

## self review

`defer`

## self review model

`fable`

## toggles

All three toggles below are declared `dynamic` for this repository: execution mode, implementer
model and the review panel roster are each handed to the plan's class (and, for the panel, its
rolls) rather than running as this run would without the toggle.

## execution mode

`dynamic`

## implementer model

`dynamic`

## review panel

`dynamic`

## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| `database` | `FLOWD_DSN` | `postgres://flow:flow@localhost:5433/flow?sslmode=disable` | `postgres://flow:flow@localhost:5433/flow_<id_underscored>?sslmode=disable` |
| `port` | `FLOWD_PORT` | `4173` | `+<offset>` |
| `url` | `FLOW_ADDR` | `http://127.0.0.1:4173` | `http://127.0.0.1:<value:FLOWD_PORT>` |
| `url` | `FLOW_RECORDS_ADDR` | `http://127.0.0.1:4173` | `http://127.0.0.1:4173` |

**The `FLOW_RECORDS_ADDR` row is deliberately not isolated, and its token-free workspace cell is
the statement of that.** The record family (`flow record`, `flow self-review bundle`) resolves
its store address from it, so an apply worktree's dispatch and finding rows land in the
persistent store the main checkout serves, and a deferred self-review bundle still finds them
after `scripts/workspace.sh remove` has dropped the `database` row's resource. The cell could not use the `database` word: that row is
taken, and cleanup removes what it names — the persistent store must never be a removal target.

| Command | Runs |
|---------|------|
| `create` | `scripts/workspace.sh create <id>` |
| `remove` | `scripts/workspace.sh remove <id>` |
| `survivors` | `scripts/workspace.sh survivors <id>` |

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
stops" paragraph is the answer to where that stage gets its stack from, and is not restated here.

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
reused stack pays the restart.

| Setting | Value |
|---------|-------|
| `ui paths` | `stats/web/src/**` |
| `screenshots` | `stats/web/tests/visual` |

| Command | Runs |
|---------|------|
| `setup` | `cd stats/web && npm install && npx playwright install chromium` |
| `verify` | `cd stats/web && npm run test:visual` |
| `capture` | `cd stats/web && npx playwright test <spec> --update-snapshots` |
| `fingerprint` | `cd stats/web && npm run build && curl -sf http://127.0.0.1:4174/ \| cmp -s - ../internal/web/dist/index.html` |
