## Why

myflow records **what state a change is in** and nothing about **what reaching that state cost**.
Every efficiency change this repository has landed — KAN-82, KAN-87, KAN-108, KAN-109, KAN-110,
KAN-153 — was argued from estimates, because no measurement existed to argue from. The pipeline's
own cost is the one thing it cannot see.

At the same time, the state record itself is a per-machine JSON file read with `jq`. It cannot be
queried across changes, carries no history, and is invisible outside a terminal.

This change replaces the state file with a PostgreSQL-backed Go service, records per-stage telemetry
for every `/myflow-*` run, and serves both the live pipeline state and aggregated statistics through
a browser UI — so the next efficiency change can be argued from data.

## What Changes

- **BREAKING**: the change state record moves from `/Users/tweety53/Agents/myflow/state/<project-key>/<name>.json`
  into PostgreSQL. Commands read and write it through a `myflow` CLI that speaks HTTP to a
  `myflowd` daemon, instead of reading and writing the file with `jq`.
- **The on-disk JSON format survives as a write-ahead journal.** When the daemon or the database is
  unreachable, a command writes the file and a journal entry, prints one warning line, and **exits
  0** — the pipeline never blocks on this subsystem. The daemon replays the journal on reconnect.
- **New**: every `/myflow-*` skill emits `myflow stage begin` / `myflow stage end` marks at the
  stage boundaries named in **Level 1 — the stages of each command** (`README.md`), so the stage
  vocabulary in the database is the documented one.
- **New**: the daemon harvests Claude Code session transcripts and attributes each message's real
  token usage, model and reasoning effort to the stage window it falls in, splitting main-thread
  from subagent cost. Cursor and Codex record state, marks and wall clock, and record token metrics
  as explicitly unavailable rather than as zero.
- **New**: a statistics API and a React SPA, both served by the same binary on `127.0.0.1`, covering
  eight views: the live state board, cost per change, the stage leaderboard, the cost trend over
  time, cache efficiency, review-panel economics, model comparison, and rework rate. Every view is
  period-parameterised and filterable by project.
- **New toolchain**: Go 1.26.5 and Node/Vite enter this repository, which was Bash + Python +
  Markdown. Go is not currently installed on the development machine and installing it is a
  prerequisite task of this change.

### Fix round 1 — the run drill-down, and a Grafana-shaped interface (2026-08-14)

The interface the eight views above describe is entirely aggregate. Every one of them groups across
changes, and none opens *one* change: there is no route, no component and no click target for a
single myflow run, so the question "what did this change actually cost, stage by stage" — the one a
person asks first — has no screen that answers it. The store and the API already hold the answer
(`GET /api/v1/stage-runs?project=&name=` returns every stage run of one change, with its metrics
bag); only the interface is missing. That is a gap in the original design rather than a new
requirement, and the operator named it at the review gate.

- **New**: a **run detail** route — `#/run/<project>/<change>` — opening one change on a timeline of
  its stage runs across every command that touched it, above a table carrying one row per stage run:
  command, stage, attempt, outcome, duration, tokens in/out/cached, cost, model, effort and
  fast-mode, each row expandable to the raw metrics bag. Its header numbers come from the server's
  own aggregate (`cost-per-change` scoped to the change), never from summing the rows on screen —
  which would silently under-report any change whose runs outgrow one page.
- **New**: the live state board's rows become links into that route, which is the only thing that
  makes it reachable without typing a URL.
- **Changed**: the interface adopts **Grafana's dashboard model** in place of the page-per-table
  shape it has now — a top bar carrying a time-range picker and template-variable dropdowns that
  scope every panel beneath them, and each view recomposed as a dashboard of panels: stat panels
  showing one large number, time-series panels, and table panels, each with its own header. The
  eight views' questions, their server queries and the absence-is-never-zero rule are unchanged —
  this restyles and recomposes the surface, and moves no query.

**The top bar carries exactly the variables the server can answer, and no others.** The statistics
endpoints accept `from`, `to` and `project`, plus `change` on one view. So the bar offers **time
range**, **project** and — once the API gains it, below — **model** as scoping variables, and
**change** as a *navigation* variable, since selecting one opens that change's own dashboard the way
a Grafana variable drives a drilldown. No variable is offered that the server cannot honour: a
dropdown filtering rows after they arrived would report wrong totals the moment a result was paged,
which is precisely what the dashboard-controls requirement forbids.

- **New**: the statistics API gains a **`model` filter**, so the model dropdown above is a real
  server-side restriction rather than a display trick. Every view that aggregates stage runs accepts
  it; the live state board, whose rows are changes rather than stage runs, **rejects** it rather than
  accepting and ignoring it — the same way `breakdown` is already rejected on the views that do not
  offer it. A stage run whose metrics bag recorded no model **does not match** any model filter, and
  the response says how many runs were excluded on that ground, so an unmeasured run is never
  silently folded into or out of a model's totals.
- **New**: `GET /api/v1/models` lists the distinct models recorded in a period, which is what
  populates the dropdown. The alternative — deriving the list client-side from the model-comparison
  view — would make every dashboard depend on a query it does not otherwise need, and would make the
  dropdown's contents depend on which view happened to be open.

### Fix round 2 — the development database purged (2026-08-14)

The store held nine changes and two stage runs, **every one of them test residue** — synthetic
project keys seeded by the guard-script harnesses (`smoke`, `proj`, `proj1`, `review-repo-*`,
`myflow-status-*-repo-*`) plus a refusal-check record. No real `/myflow-*` run was among them. Two
stray daemons during this change's development had also left 2,985 harvest offsets.

That data is harmless to correctness — every view is period- and project-filterable — but it is
noise in every dashboard the operator is about to use for real, and "no real run has ever been
recorded here" is a fact worth making true rather than filtering around.

**Purged, at the operator's instruction:** all `changes`, `change_repos` and `stage_runs` rows, and
all `harvest_offsets`. **Kept:** the schema, the migrations, and the four `pricing` rows, which are
seeded configuration rather than recorded history and are re-seeded at daemon startup anyway.

This is an operational action against a development database, not a code change: it adds no task, no
file and no commit. It is recorded here because the store is this change's own subject, and an
undocumented truncation of it would be indistinguishable later from data loss. **Clearing the
harvest offsets was chosen deliberately over keeping them** — it costs one full re-read of every
transcript on the next daemon start, which attributes nothing (no stage window is open to receive
it) and leaves a genuinely empty slate.

**Not changed by this round**: the schema, the harvester, the CLI, and every contract and skill
file. The run detail route adds no endpoint — it is built from two calls the server already answers.
The model filter does reach `internal/store` and `internal/api`, but as a new query *parameter* on
existing views plus one small list endpoint; no table, column or index changes, and no view's
question changes.

## Capabilities

### New Capabilities

- `myflow-state-store`: the PostgreSQL-backed store that replaces the state file — its schema, the
  daemon/CLI interface every command uses, the write-ahead journal that keeps the pipeline
  non-blocking, and journal reconciliation.
- `myflow-run-telemetry`: how a run's cost is measured — the stage marks skills emit, the transcript
  harvester that attributes real token usage to stage windows, the metrics bag's flexibility rule,
  abandoned-stage sweeping, and honest degradation on harnesses with no transcript.
- `myflow-stats-views`: the aggregation surface — the period-parameterised statistics API, the eight
  views and the questions each answers, project as a first-class filter, and the localhost-only
  exposure of the UI.

### Modified Capabilities

- `myflow-state-machine`: its scenario asserting what "the state file" contains is restated against
  the state **record**, whose primary home becomes the store and whose on-disk form becomes the
  fallback journal. The three states, the command→state mapping, the human-gate-as-state-property
  rule and the re-entrancy rule are all unchanged; only where the record lives changes.

## Impact

**Contracts rewritten**

- `skills/myflow-contracts/state-file.md` — stops describing a JSON file and describes the store,
  keeping its field vocabulary, closed-schema rule, monotonic-write rule and carry-forward rule
  unchanged. The mechanism changes; the contract does not.
- `skills/myflow-contracts/pipeline.md` — gains the stage-mark obligation, citing the README's
  Level 1 list for stage names rather than restating them.

**Skills touched** — `myflow-start`, `myflow-do`, `myflow-finish`, `myflow-fast` each gain mark
calls at their own stage boundaries; `myflow-status` reads the store instead of the state directory.

**New code** — a top-level `stats/` directory: `cmd/myflowd`, `cmd/myflow`, `internal/store`,
`internal/harvest`, and `web/` (React + Vite, embedded into the binary with `go:embed`).

**New infrastructure** — a dedicated `myflow-postgres` container (own port, own volume, independent
of the `gymie-*` stack) and a launchd user agent running `myflowd` at login.

**Dependencies** — Go 1.26.5, PostgreSQL 18.6, React 19.2.8, Vite 8.2.1. Versions verified against
`go.dev/dl`, `endoflife.date`, and the npm registry on 2026-08-13.

**Data read** — `~/.claude/projects/*/*.jsonl`, read-only, for token and model attribution.

**Not affected** — the three pipeline states, the command surface, the git boundaries, the review
panel, and the Jira contract. This change makes the pipeline observable; it does not change what the
pipeline does.
