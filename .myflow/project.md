# myflow project configuration — agents

Read by globally installed myflow skills. Every key is optional; anything absent is
auto-detected from the repository instead.

## apps

This repository is mostly the source of the myflow skills, commands, and rules, installed
elsewhere by `setup.sh` — that half has no port and no URL. It also now holds `stats/`, a
PostgreSQL-backed Go service (`myflowd`) with an embedded React SPA, plus a thin `myflow` CLI. Both
halves live in the one repo and are covered below.

| App | Repo root | Kind | URL | Notes |
|-----|-----------|------|-----|-------|
| myflow sources | `/Users/tweety53/Projects/agents` | Bash + Python + Markdown | — | The skills/commands/rules half. Verification is the guard scripts below plus a sandboxed `setup.sh` run. |
| myflow stats daemon | `/Users/tweety53/Projects/agents/stats` | Go + React/Vite | `http://127.0.0.1:4173` | `myflowd`, loopback-only. Backed by a dedicated `myflow-postgres` container on host port 5433, independent of any other Postgres stack on this machine. |

**This repository is Bash + Python, not Bash-only.** `scripts/check-plan-provenance.sh` is a thin
wrapper that execs `scripts/check-plan-provenance.py` (Python 3, standard library only —
`/usr/bin/python3`, no third-party imports, no pip, no network) so its fence/container classifier
can be a real block-structure parser instead of a hand-rolled Bash ERE allowlist. This followed five
review panel passes and seven fix waves that found defect class after defect class in the Bash
version — canonical enumeration and full history in `check-plan-provenance.py`'s own module
docstring (this file does not restate the count, since a copied number is exactly what let an
earlier, wrong count survive six review passes) and
`openspec/changes/kan-14-plan-provenance/design.md`'s "Post-review reshape" section. Every other
guard in this repository remains Bash-only; adding Python here was a deliberate, recorded widening
of the toolchain, not a drift.

## run

To exercise the installer without touching the real home directory:

```bash
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
```

**The stats daemon.** Bring up the dedicated Postgres stack, then build and run `myflowd`:

```bash
cd stats && docker compose up -d          # myflow-postgres on host port 5433
cd stats && make build                    # builds the SPA, then verifies the Go build
cd stats && go build -o bin/myflowd ./cmd/myflowd && ./bin/myflowd
```

`myflowd` binds `127.0.0.1:4173` (override with `MYFLOWD_PORT`) and refuses to start on any other
interface or on an unparsable `MYFLOWD_PORT` rather than defaulting silently. Running it this way is
for manual, foreground verification only — for a daemon that survives logout and restarts on
failure, see `stats/README.md`'s "Running the daemon at login" section and its launchd agent.
**No skill loads that agent**; loading it is an operator step, deliberately, because an agent left
running unattended during this change's development harvested 2,961 transcript offsets into the
database before anyone noticed.

**The UI-test stack**, for ad-hoc testing from the main checkout rather than from an apply worktree:
`make ui-test-up` and `make ui-test-down` (`stats/Makefile`) bring a second, disposable `myflowd` up
on port 4174 against `myflow_uitest`, seeded with a fixed fixture, and tear it down again. Point a
session at it with `MYFLOW_ADDR=http://127.0.0.1:4174`. Neither target is isolated by the
`## workspace isolation` section below — that section covers apply worktrees, and this stack is a
single, main-checkout-only fixture instead.

## test

```bash
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-check-plan-provenance.sh
scripts/test-check-finish-preflight.sh
scripts/test-commit-split.sh
scripts/test-prepare-workspace.sh
scripts/test-preserve-session-records.sh
scripts/test-check-unfinished-work.sh
scripts/test-check-cleanup-complete.sh
scripts/test-gather-self-review-context.sh
scripts/test-gather-dispatch-context.sh
scripts/test-check-task-build-green.sh
scripts/test-check-task-commit-fields.sh
scripts/test-check-workspace-isolation.sh
scripts/test-workspace.sh
scripts/test-check-contract-budget.sh
scripts/test-check-vocabulary.sh
scripts/test-check-panel-diff-size.sh
scripts/test-plan-dispatch-bundles.sh
scripts/test-check-panel-reproducers.sh
scripts/test-run-reproducer.sh
scripts/test-check-markdown-integrity.sh
scripts/test-check-uitest-overrides.sh
scripts/test-check-guard-symlinks.sh
scripts/test-check-stage-mark-calls.sh
scripts/test-check-self-review-report.sh
scripts/test-lib-coverage.sh
cd stats && go test ./... -race -count=1
cd stats/web && npm test
```

**Measured runtime: roughly 144s total** — 118.63s for the Bash/Python guard tests above, plus
23.93s for `go test ./... -race -count=1` and 1.80s for the SPA's `npm test` (both measured against
the already-running `myflow-postgres` compose stack, with `stats/internal/web/dist` already built)
— against this harness's 120000ms default tool timeout. This was already close to the edge before
`stats/` existed; adding the Go and SPA suites moves it from "close" to "reliably over" for a single
invocation. Split the run across more than one invocation (the guard tests as one call,
`cd stats && go test ./...` as a second, `cd stats/web && npm test` as a third) or raise the tool
timeout, rather than reading a timeout here as one of these commands failing. This note cites no
count of the list on purpose: a written count goes stale the first time a command is added to it,
and nothing here checks it against the list above.

## lint

```bash
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
scripts/check-workspace-isolation.sh
scripts/check-uitest-overrides.sh
scripts/check-contract-budget.sh
scripts/check-markdown-integrity.py
scripts/check-stage-mark-calls.sh
scripts/check-guard-symlinks.sh
scripts/check-self-review-report.sh
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
hand like a guard-script one. The list is cited by count nowhere in this file, deliberately: a
written count went stale the first time a guard was added to it, and the same sentence would go
stale again on the next.

**`check-contract-budget.sh` is a ratchet, not a target.** It fails when a file under
`skills/myflow-contracts/`, or a `skills/*/SKILL.md` or `skills/*/SKILL-rationale.md`, outgrows the
budget declared for it in the guard's own `budgets()` table, or carries no budget at all. The table
is keyed on the path relative to the repository root, not on the bare basename, because every skill
directory has a file literally named `SKILL.md` and a basename key would collide across skills. Each
budget is the size its file had when the change that added its row landed, plus 25% — so ordinary
edits pass and a real section addition trips it, forcing a deliberate edit to the table rather than a
silent regrowth of a file every `/myflow-*` command loads. Raising a budget is the correct response
to a genuine addition; narrowing the guard's scope or deleting a row is not.

**`check-workspace-isolation.sh` is a lint step where the other `## workspace isolation` guard is
not.** It takes a project root, defaults to this repository when given none, and answers a question
about the text of `.myflow/project.md` — so it runs against a bare tree like every other guard here.
`check-cleanup-complete.sh` reads the same section and is excluded below for the opposite reason: it
needs a change in flight.

**Its place in this list is a self-check on this repository, not how it covers the projects myflow
is installed into** — and conflating the two made a permanently vacuous lint step read as
enforcement. The run here checks this repository's own `## workspace isolation` section below; a
green lint run therefore says nothing whatever about any other project's declaration. What covers
those is `/myflow-do`, which runs this guard against
each apply worktree before it resolves the section, per section 7 of `skills/myflow-do/SKILL.md` —
so a declaration is validated where it is read, in whichever repository holds it. The lint entry
stays because this repository's own configuration is one more configuration worth checking, and
because it keeps the guard runnable from a bare tree.

**`check-finish-preflight.sh`, `preserve-session-records.sh`, `check-unfinished-work.sh` and
`check-cleanup-complete.sh` are deliberately not lint steps.** All four are `/myflow-finish` helpers
that need a change in flight and a real worktree, a repository or a state directory passed in as
arguments; they answer a question about one change, not about the state of the repository's text. A
lint step that cannot run against a bare tree would fail on every unrelated invocation, so the
omission is a decision, not an oversight. They are covered instead by their harnesses under
`## test`. `check-panel-diff-size.sh`, `plan-dispatch-bundles.sh`, `check-panel-reproducers.sh` and
`run-reproducer.sh` are excluded for the same reason: they are `/myflow-do` helpers that likewise
need a change in flight and a worktree passed in, so they are covered by their own harnesses under
`## test` instead.

**Every guard in the list is currently expected to exit 0.** `check-workspace-isolation.sh` reports
`ISOLATION-OK` and validates this repository's own declared section; its own header carries its
full exit-code contract: 0 every project checked is well formed, 1 violations found, 2 it cannot
answer at all. `check-plan-provenance.sh` reports
"all provenance stated" — see the script's own header for the full exit-code contract: 0
clean/nothing-in-flight, 1 violations found, 2 environment, 3 containment, 4
content-classification; a caller that treats "non-zero" uniformly, as this repository's own lint
step does, is unaffected by that split. The long-standing exception recorded here previously — a
block of unattributed fenced snippets in `kan-8-myflow-updates`'s plan — cleared on its own when
that change archived, exactly as predicted, because the guard excludes `openspec/changes/archive/`
by design. There is no known exception left. A future non-zero exit is a real hit on a plan in
flight: fix the offending line by stating its provenance, never by narrowing the guard's scope or
adding a suppression marker.

## stop

The manual, foreground `myflowd` started above stops with Ctrl-C or `kill` on its PID — it shuts
down gracefully, draining in-flight requests before its connection pool closes.

```bash
cd stats && docker compose down           # stops myflow-postgres; does not touch other stacks
launchctl unload ~/Library/LaunchAgents/com.tweety53.myflowd.plist   # only if the login agent is loaded
```

The launchd agent line is a no-op if the agent was never loaded — see `stats/README.md`. This key
is present rather than omitted so that `/myflow-finish`'s stack-stopped check runs against a real
answer, not a recorded "nothing to stop" that stopped being true when `stats/` landed.

## standards

- `CLAUDE.md`
- `AGENTS.md`

## jira

`KAN`

## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| `database` | `MYFLOWD_DSN` | `postgres://myflow:myflow@localhost:5433/myflow?sslmode=disable` | `postgres://myflow:myflow@localhost:5433/myflow_<id_underscored>?sslmode=disable` |
| `port` | `MYFLOWD_PORT` | `4173` | `+<offset>` |
| `url` | `MYFLOW_ADDR` | `http://127.0.0.1:4173` | `http://127.0.0.1:<value:MYFLOWD_PORT>` |

| Command | Runs |
|---------|------|
| `create` | `scripts/workspace.sh create <id>` |
| `remove` | `scripts/workspace.sh remove <id>` |
| `survivors` | `scripts/workspace.sh survivors <id>` |

**`MYFLOW_STATE_DIR` and `MYFLOW_TRANSCRIPTS_DIR` are deliberately not isolated.** The `Resource`
column above is a closed vocabulary — `database`, `bucket`, `cache index`, `port`, `url`, and no
other word — and a directory path is none of those five, so neither variable gets a row; this is a
decision, not an oversight. The consequence is bounded: every worktree's `myflowd` still harvests
from the one real transcripts root, but each writes the result into its *own* isolated database from
the table above, so the harvested data ends up duplicated across worktrees rather than shared through
one database. Neither variable crosses the boundary this section exists to protect.
