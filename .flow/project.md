# myflow project configuration — agents

Read by globally installed myflow skills.

## apps

This repository is mostly the source of the myflow skills, commands, and rules, installed
elsewhere by `setup.sh` — that half has no port and no URL. It also now holds `stats/`, a
PostgreSQL-backed Go service (`myflowd`) with an embedded React SPA, plus a thin `myflow` CLI. Both
halves live in the one repo and are covered below.

| App | Repo root | Kind | URL | Notes |
|-----|-----------|------|-----|-------|
| myflow sources | `/Users/tweety53/Projects/agents` | Bash + Python + Markdown | — | The skills/commands/rules half. Verification is the guard scripts below plus a sandboxed `setup.sh` run. |
| myflow stats daemon | `/Users/tweety53/Projects/agents/stats` | Go + React/Vite | `http://127.0.0.1:4173` | `myflowd`, loopback-only. Backed by a dedicated `myflow-postgres` container on host port 5433, independent of any other Postgres stack on this machine. Also the one application `## apps` names that a fix run's reload rule (`skills/flow/verify-and-handoff.md`) never reloads — the same protection, not a separate one. |

**This repository is Bash + Python, not Bash-only.** `scripts/check-plan-provenance.sh` is a thin
wrapper that execs `scripts/check-plan-provenance.py` (Python 3, standard library only —
`/usr/bin/python3`, no third-party imports, no pip, no network) so its fence/container classifier
can be a real block-structure parser instead of a hand-rolled Bash ERE allowlist. This followed five
review panel passes and seven fix waves that found defect class after defect class in the Bash
version — canonical enumeration and full history in `check-plan-provenance.py`'s own module
docstring (this file does not restate the count, since a copied number is exactly what let an
earlier, wrong count survive six review passes) and in
`openspec/changes/archive/2026-07-29-kan-14-plan-provenance/design.md` — frozen and historical, a
record of why the wrapper exists rather than a live artifact — under its "Post-review reshape"
section. Every other guard in this repository remains Bash-only; adding Python here was a
deliberate, recorded widening of the toolchain, not a drift.

### artifact tree

The pipeline's artifact tree is `spectre/`. `openspec/` is frozen at the 2026-08-25 cutover and is
never written to again. The two changes left open in it —
`kan-295-cut-pipeline-load-cost-split-by-consumer` and `kan-327-review-per-round-delta-after-round-1`
— are abandoned.

## run

To exercise the installer without touching the real home directory:

```bash
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
```

**The stats daemon.** Bring up the dedicated Postgres stack, then build and run `myflowd`:

```bash
cd stats && docker compose up -d          # myflow-postgres on host port 5433
cd stats && make build                    # builds the SPA, then both binaries into bin/
cd stats && ./bin/myflowd
```

`myflowd` binds `127.0.0.1:4173` (override with `FLOWD_PORT`) and refuses to start on any other
interface or on an unparsable `FLOWD_PORT` rather than defaulting silently. Running it this way is
for manual, foreground verification only — for a daemon that survives logout and restarts on
failure, see `stats/README.md`'s "Running the daemon at login" section and its launchd agent.
**No skill loads that agent**; loading it is an operator step, deliberately, because an agent left
running unattended during this change's development harvested 2,961 transcript offsets into the
database before anyone noticed. **No automated stage runs these three commands either** — see the
UI-test stack below for what an automated stage uses instead.

**The UI-test stack**, for ad-hoc testing from the main checkout rather than from an apply worktree:
`make ui-test-up` and `make ui-test-down` (`stats/Makefile`) bring a second, disposable `myflowd` up
on port 4174 against `myflow_uitest`, seeded with a fixed fixture, and tear it down again. Point a
session at it with `FLOW_ADDR=http://127.0.0.1:4174`. Neither target is isolated by the
`## workspace isolation` section below — that section covers apply worktrees, and this stack is a
single, main-checkout-only fixture instead.

**This is the stack `flow.visual-verify` (`skills/flow/verify-and-handoff.md`) starts and stops.**
That stage's step 4 probes a URL and, if nothing answers, "starts the stack from `## run`" — a
project with more than one candidate stack has to say which one that means, so this paragraph is
the answer: `make ui-test-up` / `make ui-test-down` against `http://127.0.0.1:4174`, matching the
`## visual verification` section below and `stats/web/playwright.config.ts`'s pinned `baseURL`.
**Never `myflowd` on `127.0.0.1:4173`** — that is the dev workspace's protected daemon (`## stop`
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

**Measured runtime: the guard-test harnesses now run in about 52 to 56s wall through the runner**,
down from roughly 216 to 218s running the same 42 harnesses sequentially, one after another.
<!-- measured: time scripts/run-guard-tests.sh, three clean runs of 52.5s/52.9s/52.0s wall @ branch spectre/kan-362-myflow-guard-test-suite-takes-119s-sequentially -->
<!-- measured: for f in scripts/test-*.sh; do bash "$f"; done, two runs of 218.44s/216.10s wall @ branch spectre/kan-362-myflow-guard-test-suite-takes-119s-sequentially -->
Both figures are for the 42 Bash/Python guard harnesses alone; the two `stats` commands above are
untouched by this change and are not part of either figure. Plus those two commands, the whole
`## test` list now comfortably fits inside this harness's 120000ms default tool timeout in one
invocation — no more splitting the run across several calls, and no need to raise the timeout.

One run through the runner, out of five taken while measuring this, reported
`test-check-installed-citations.sh` as `FAIL` with no case-level failure in its replayed output. The
failure was reproduced (3 of 8 full-suite runs) and root-caused: **not** oversubscription, but a
shared temp-directory namespace. `test-lib-parallel.sh` (cases 7–8) and
`test-check-installed-citations.sh` (trap-chain sub-cases) each identified their own child's
directory by snapshotting the global `$TMPDIR` for `parallel-lib.*`; under
`scripts/run-guard-tests.sh`'s concurrent harnesses, a sibling process's live directory could be
picked up instead, producing empty paths, "could not observe", and false "leaked" reports. Fixed in
`2e3ed4f` by giving each observing child its own private `TMPDIR`. Since the fix, 24 consecutive
clean full-suite runs.
<!-- measured: scripts/run-guard-tests.sh, 24 consecutive clean runs after 2e3ed4f @ branch spectre/kan-362-myflow-guard-test-suite-takes-119s-sequentially -->

**`check-installed-citations.sh` (named in `## lint` below) is unlike every other guard in that
list: it shells out to a sandboxed `setup.sh` twice per invocation** — once for `global`, once for
`all` — to derive the installed set it classifies citations against, rather than only reading
files already on disk. A single invocation measures about 0.84s, negligible against either total
above, but worth naming here since it is the one guard in this repository paying for a subprocess
rather than a plain file scan.
<!-- measured: time scripts/check-installed-citations.sh >/dev/null @ branch openspec/kan-102-citations-resolve-to-installed-paths -->

## lint

```bash
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
scripts/check-plan-shape.sh
scripts/check-workspace-isolation.sh
scripts/check-visual-verification.sh .
printf 'stats/web/src/App.tsx\n' | scripts/check-visual-trigger.sh .
scripts/resolve-visual-screenshots.sh . baseline.spec.ts
scripts/check-uitest-overrides.sh
scripts/check-contract-budget.sh
scripts/check-markdown-integrity.py
scripts/check-stage-mark-calls.sh
scripts/check-guard-symlinks.sh
scripts/check-dispatch-paragraphs.sh
scripts/check-self-review-report.sh
scripts/check-installed-citations.sh
scripts/check-installed-rules.sh
scripts/check-normative-inventory.sh
scripts/check-model-keys.sh
scripts/check-model-resolution-shell.sh
cd stats && gofmt -l .
cd stats && go vet ./...
cd stats/web && npx tsc -b
```

**`check-plan-shape.sh` sits beside `check-plan-provenance.sh` and `check-task-build-green.sh` in
this list but is not one of them.** Those two are project-configured — resolved through a project's
own `.flow/project.md` and run only where a project declares them — while `check-plan-shape.sh` is
**shipped**, symlinked into `skills/flow/scripts/` and cited by basename per **Guard resolution**
(`skills/flow-contracts/pipeline.md`), because the guard it protects
(`check-task-commit-fields.sh`) is itself shipped and runs in every project `/flow` touches. It
answers a bare-tree question exactly like `check-plan-provenance.sh` and `check-task-build-green.sh`
do — no arguments scans every non-archived `<spec-root>/changes/*/tasks.md` — which is why it
belongs in this list at all, for the same reason those two do.

**There is no auto-fix command for the guard scripts** (`scripts/check-*`) — every one of them
reports `file:line` and is fixed by editing the offending line, never by weakening the guard or
adding a suppression marker to silence a real hit. **`stats/` does have one**: `cd stats && gofmt
-w .` reformats Go source before the `gofmt -l .` check above is run, per the Lint Fix Priority
rule's "run the auto-fix command first" step. There is no equivalent for the SPA — `web/package.json`
carries no lint or format script, only `tsc -b`'s type check, so a TypeScript violation is fixed by
hand like a guard-script one. The list is cited by count nowhere in this file, deliberately: a
written count went stale the first time a guard was added to it, and the same sentence would go
stale again on the next.

**`check-contract-budget.sh` is a ratchet, not a target.** It fails when an owned `.md` or `.mdc`
file — every one under `skills/`, `rules/`, `spectre/specs/`, `commands/`, `commands-claude/`,
`.flow/` and the repository root, resolved through `scripts/lib/owned-corpus.sh` — outgrows the
budget declared for it in the guard's own `budgets()` table, or carries no budget at all. The table
is keyed on the path relative to the repository root, not on the bare basename, because every skill
directory has a file literally named `SKILL.md` and a basename key would collide across skills. Each
budget is the size its file had when the change that added its row landed, plus 25% — so ordinary
edits pass and a real section addition trips it, forcing a deliberate edit to the table rather than a
silent regrowth of a file every `/myflow-*` command loads. Raising a budget is the correct response
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

**Its place in this list is a self-check on this repository, not how it covers the projects myflow
is installed into** — and conflating the two made a permanently vacuous lint step read as
enforcement. The run here checks this repository's own `## workspace isolation` section below; a
green lint run therefore says nothing whatever about any other project's declaration. What covers
those is `/flow`, which runs this guard against
each apply worktree before it resolves the section, per **Verify** in `skills/flow/verify-and-handoff.md` —
so a declaration is validated where it is read, in whichever repository holds it. The lint entry
stays because this repository's own configuration is one more configuration worth checking, and
because it keeps the guard runnable from a bare tree.

**`check-finish-preflight.sh`, `check-unfinished-work.sh`, `check-cleanup-complete.sh` and
`check-worktree-processes.sh` are deliberately not lint steps.** All four are `/flow` integrate/archive
helpers that need a change in flight and a real worktree, a repository or a state directory passed
in as arguments; they answer a question about one change, not about the state of the repository's
text. A lint step that cannot run against a bare tree would fail on every unrelated invocation, so
the omission is a decision, not an oversight. They are covered instead by their harnesses under
`## test`.
`check-panel-diff-size.sh`, `plan-dispatch-bundles.sh`, `check-panel-reproducers.sh` and
`run-reproducer.sh` are excluded for the same reason: they are `/flow` implementation helpers that
likewise need a change in flight and a worktree passed in, so they are covered by their own
harnesses under `## test` instead.

**`check-installed-citations.sh` belongs in the list for the opposite reason those are excluded.**
It takes no change-in-flight state — it derives the installed set by running a sandboxed `setup.sh`
itself, twice per invocation, rather than being handed one — so it scans the same bare tree every
other guard above does and exits identically regardless of what change, if any, is in flight. It is
the only guard in this list that shells out to the installer rather than only reading files already
on disk, which costs it real time; see the runtime note under `## test`, next to the entry its
harness added there.

**`check-dispatch-paragraphs.sh` keeps a required dispatch paragraph from silently
disappearing.** It is argument-free and self-scoped exactly like `check-guard-symlinks.sh`: the
scan root is this repository's own root, resolved from the script's own location. It checks a table
of required paragraphs — REPRODUCE, DON'T READ at `skills/flow/review-panel.md` and
`skills/flow/implement.md`, and VERBATIM REPORT — THE FACT at `skills/flow/review-panel.md` — for
each paragraph's label and a handful of load-bearing phrases per variant, held as short literals
rather than a copy of the whole blockquote, so a body reworded around those phrases passes clean. A
green run proves only that each paragraph is present at its required sites and carries its phrases —
never that any reviewer, implementer, dispatcher or fix agent actually obeyed it.

**`check-installed-rules.sh` is the one guard in this list that reads outside the repository.**
It compares the always-on rules this checkout declares against what `setup.sh global` last installed
under `$HOME` — the symlinks in `~/.claude/rules/` and the per-rule markers in the two managed
blocks — because a rule can merge and stay unreadable by every session, which is how KAN-202's
commit-scope rule spent a day with a dangling pointer. It still runs against a bare tree and takes no
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
step does, is unaffected by that split. The long-standing exception recorded here previously — a
block of unattributed fenced snippets in `kan-8-myflow-updates`'s plan — cleared on its own when
that change archived, exactly as predicted, because the guard excludes `changes/archive/` by
design — `spectre/changes/archive/` today, `openspec/changes/archive/` in the frozen tree where
that change actually sits. There is no known exception left. A future non-zero exit is a real hit
on a plan in flight: fix the offending line by stating its provenance, never by narrowing the
guard's scope or adding a suppression marker.

## stop

**This key declares no command, and that is the answer rather than an omission.** `/flow`'s
worktree-cleanup check 5 runs whatever this key declares before removing a worktree; there is
nothing here for it to run.

**What is protected is the dev workspace's service and its storage, and only those** — the empty-id
case of `## workspace isolation` below, which is what the main checkout resolves:

| Protected — never stopped, dropped or removed by any agent action | Why |
|---|---|
| `myflowd` on `127.0.0.1:4173` | the store every `myflow` call in every project writes through |
| the `myflow-postgres` container on host port 5433 | the **service**, shared by every workspace |
| the default `myflow` database inside it | the dev workspace's own storage |

Not `docker compose down`, not `launchctl unload`, not a `kill` on the daemon's pid, and not to make
a later step succeed. Stopping any of them mid-run silently degrades the rest of that run's writes to
the on-disk journal and takes the store away from every other session on this machine; restarting
afterwards does not repair it, because whatever fell through to the journal while it was down stays
there.

**A workspace's own derived resources are not protected, and removing them is correct.** The
`myflow_<id_underscored>` database and the bucket an apply worktree derives are per-change artifacts:
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

## self review model

`fable`

## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| `database` | `FLOWD_DSN` | `postgres://flow:flow@localhost:5433/flow?sslmode=disable` | `postgres://flow:flow@localhost:5433/flow_<id_underscored>?sslmode=disable` |
| `port` | `FLOWD_PORT` | `4173` | `+<offset>` |
| `url` | `FLOW_ADDR` | `http://127.0.0.1:4173` | `http://127.0.0.1:<value:FLOWD_PORT>` |

| Command | Runs |
|---------|------|
| `create` | `scripts/workspace.sh create <id>` |
| `remove` | `scripts/workspace.sh remove <id>` |
| `survivors` | `scripts/workspace.sh survivors <id>` |

**`FLOW_STATE_DIR` and `FLOW_TRANSCRIPTS_DIR` are deliberately not isolated.** The `Resource`
column above is a closed vocabulary — `database`, `bucket`, `cache index`, `port`, `url`, and no
other word — and a directory path is none of those five, so neither variable gets a row; this is a
decision, not an oversight. The consequence is bounded: every worktree's `myflowd` still harvests
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

| Setting | Value |
|---------|-------|
| `ui paths` | `stats/web/src/**` |
| `screenshots` | `stats/web/tests/visual` |

| Command | Runs |
|---------|------|
| `setup` | `cd stats/web && npm install && npx playwright install chromium` |
| `verify` | `cd stats/web && npm run test:visual` |
| `capture` | `cd stats/web && npx playwright test <spec> --update-snapshots` |
