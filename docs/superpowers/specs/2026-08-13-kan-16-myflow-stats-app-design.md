# KAN-16 — myflow stats app

**Date:** 2026-08-13
**Jira:** [KAN-16](https://tweety53.atlassian.net/browse/KAN-16) — "Myflow stats app"
**Change:** `kan-16-myflow-stats-app`

## Purpose

Replace myflow's machine-local JSON state file with a Go service backed by PostgreSQL, and use the
same store to record per-stage telemetry for every `/myflow-*` run: token cost, wall-clock cost,
model, reasoning effort, fast-mode, and the outcome of each stage. Serve both surfaces — current
pipeline state and aggregated statistics over any period — through a browser UI.

The pipeline today records *what state a change is in* and nothing about *what reaching that state
cost*. Every myflow change since KAN-8 has been an efficiency change argued from estimates, because
no measurement existed. This change makes the cost of the pipeline observable, so the next
efficiency change can be argued from data.

## Constraints

- **The pipeline must never block on this subsystem.** A state write is a gate in a way a Jira call
  is not, so the availability posture has to be stronger than Jira's, not equal to it.
- **The metric structure must absorb frequent additions.** myflow changes often; a new metric must
  not require a schema migration.
- **myflow installs into three harnesses.** Claude Code, Cursor and Codex. Only Claude Code writes
  a machine-readable transcript.
- **Five projects already use myflow**, with 24 live state files between them.

## Decisions

### Scope: all four subsystems land in one change

**ID:** `scope-all-four`
**Status:** active
**Chosen:** State store, telemetry ingestion, aggregation and UI in KAN-16 — the pieces are not
independently useful. A state store with no statistics is a regression from a JSON file (it adds a
network dependency and returns the same data); statistics with no UI cannot be read.
**Considered:**
- *Store and telemetry only, defer aggregation and UI* — produces no readable output at all, so the
  change could not be evaluated by the person who asked for it.
- *Keep the state file, add stats alongside it* — materially lower risk, and it delivers the part
  that is actually looked at. Ruled out because the ticket's first sentence asks for the
  replacement, and running two sources of state indefinitely is the outcome the ticket exists to
  end.

### Postgres is the source of truth, with a local write-ahead journal

**ID:** `source-of-truth-with-journal`
**Status:** active
**Chosen:** Commands write the store; on any failure they write the existing on-disk JSON plus a
journal entry and exit 0. The daemon replays the journal on reconnect. This is a real replacement —
the store is authoritative — without a database outage stopping the pipeline mid-change.
**Considered:**
- *Local file remains the source of truth, mirrored asynchronously into Postgres* — the safest
  option and the one matching the Jira contract's "never a gate" posture, but it is not the
  replacement the ticket asks for; the file would remain the thing every contract describes.
- *Postgres authoritative, the run stops when it is unreachable* — what the ticket literally
  describes. Ruled out because a state write happens at the end of every command, so an outage
  would strand changes at an unwritten state with work already done.

### A long-running daemon owns the database

**ID:** `daemon-owns-db`
**Status:** active
**Chosen:** `myflowd` holds the connection pool and serves both the state API and the stats API; the
CLI and the SPA are both HTTP clients of it. One writer means the journal replay, the harvester and
the stage-timeout sweeper all live in one process with one view of the data.
**Considered:**
- *A single binary in CLI mode connecting to Postgres directly, with a separate serve mode* — no
  daemon for the pipeline to depend on, and the closest fit to how skills already shell out to
  `scripts/*.sh`. Ruled out because the transcript harvester and the abandoned-stage sweeper both
  need to run continuously; with no daemon they would have no host.
- *HTTP API only, skills calling `curl`* — every contract file grows inline `curl` and error
  handling, which is exactly the fragility the CLI exists to contain.

### Stage boundaries are explicit marks; metrics come from the transcript

**ID:** `marks-plus-harvest`
**Status:** active
**Chosen:** Skills emit `myflow stage begin/end` at the README Level 1 boundaries, which defines the
window exactly and harness-independently. The daemon harvests `~/.claude/projects/*/*.jsonl` and
attributes each message's real `usage`, `model` and `effort` to whichever open window its session
and timestamp fall in.
**Considered:**
- *Marks alone, with the end mark carrying the numbers the agent supplies* — an agent cannot observe
  its own cache-read tokens, and cannot observe its subagents' usage at all, so the headline cost
  figure would be wrong in the direction that matters most.
- *Inference from the transcript alone, with no marks* — requires no change to any skill, but stage
  attribution becomes a heuristic that breaks whenever a skill is reworded, and the stage names
  would drift from the documented ones.

### Typed identity, JSONB metrics

**ID:** `typed-core-jsonb-metrics`
**Status:** active
**Chosen:** Stable typed columns for identity and lifecycle (project, change, command, stage,
attempt, timestamps, outcome) and a GIN-indexed JSONB `metrics` column for everything measured.
Adding a metric is writing a new key.
**Considered:**
- *Fully typed columns per metric* — best constraints and query ergonomics, worst fit for a ticket
  whose stated premise is frequent change; each new metric becomes a migration.
- *Append-only event log with everything derived* — maximum flexibility and full replay, but every
  statistic becomes a query to write before it can be shown, which pushes the UI far out.

### Cost is frozen per row and re-priceable

**ID:** `pricing-version`
**Status:** active
**Chosen:** Store token counts as ground truth, a derived USD cost, **and** the `pricing` row
version that produced it. The dashboard is stable, and history can be recomputed when model prices
change.
**Considered:**
- *Store tokens only, price at query time* — survives price changes but makes every historical
  figure move under the reader whenever pricing is edited.
- *Store a frozen dollar figure with no pricing version* — stable, but silently wrong once prices
  change and impossible to correct.

### Claude Code gets full telemetry; other harnesses degrade honestly

**ID:** `harness-degradation`
**Status:** active
**Chosen:** Cursor and Codex record state, stage marks and wall-clock time. Token metrics are
recorded as explicitly unavailable (`metrics.tokens_available: false`), never as zero.
**Considered:**
- *Claude Code only, as a hard requirement* — a Cursor run would then write state the app could not
  interpret.
- *All three harnesses from the start* — neither Cursor nor Codex has a documented transcript
  format, so this is open-ended research inside an already large change.

### The Go application lives in this repository

**ID:** `same-repo`
**Status:** active
**Chosen:** A new top-level `stats/` directory, alongside `skills/` and `scripts/`. The state
contract and the application implementing it version together: renaming a stage mark and updating
the skill that emits it is one commit.
**Considered:**
- *A separate repository* — keeps this one a pure prompt/config repo, but splits every contract
  change across two pull requests.
- *A git submodule* — separate history in one checkout, at the cost of submodule friction in every
  worktree `/myflow-do` creates, which is constant.

### Own docker-compose stack, daemon under launchd

**ID:** `dedicated-stack`
**Status:** active
**Chosen:** A dedicated `myflow-postgres` container with its own port and volume, and `myflowd`
under a launchd user agent so it is running at login.
**Considered:**
- *Reuse the running `gymie-postgres` container with a separate database* — nothing new to run, but
  it couples the pipeline's state store to a project stack that gets stopped, so the write-ahead
  journal would fire routinely rather than exceptionally.
- *Start the daemon on demand from the CLI* — no launchd, but adds first-call latency and a startup
  race when parallel worktrees invoke it simultaneously.

### Localhost-only, no authentication

**ID:** `localhost-no-auth`
**Status:** active
**Chosen:** Bind `127.0.0.1` on a fixed port, no credentials. A single-user dashboard on the user's
own machine, holding no secrets beyond pipeline metadata.
**Considered:**
- *A shared token* — guards against other local processes, at the cost of token handling in every
  client; disproportionate for the data held.
- *LAN-reachable with real authentication* — needs auth, TLS and a hardening pass, which is a
  materially larger change than the one asked for.

### Existing state files are imported and become the journal format

**ID:** `import-and-reuse-format`
**Status:** active
**Chosen:** A one-shot import of all 24 existing files, after which the same on-disk format serves
as the write-ahead journal. A change in flight during the cutover finishes on either path.
**Considered:**
- *Import once, then delete the files* — cleanest end state, but the chosen fallback needs a local
  format regardless, so this would mean inventing a second one.
- *Start empty* — no import; the two paths then coexist with no end date.

## Open questions

*(none — every question raised during brainstorming was answered)*

## Architecture

### Components

Five units under a new top-level `stats/` directory.

| Unit | Path | Responsibility |
|------|------|----------------|
| `myflowd` | `stats/cmd/myflowd` | HTTP daemon: owns the Postgres pool, serves the state API, the stats API and the embedded SPA; hosts the harvester, the journal replayer and the stage sweeper |
| `myflow` | `stats/cmd/myflow` | Thin CLI the skills invoke: `state get`, `state set`, `stage begin`, `stage end`, `import`, `journal flush` |
| `store` | `stats/internal/store` | Schema, migrations and queries — the only package that knows SQL |
| `harvest` | `stats/internal/harvest` | Tails Claude Code transcripts and attributes usage to stage windows |
| `web` | `stats/web` | React SPA, built by Vite and embedded into the binary with `go:embed` |

The SPA is embedded rather than served separately: one binary means no static host to run, no CORS
surface, and no way for the UI and API to be at different versions.

**Boundaries.** `store` exposes typed repository methods and no SQL; the daemon may not build a
query. `harvest` consumes an interface for "which stage window is open for this session" and does
not read the database directly, so it is testable against fixtures with no Postgres. The CLI knows
only HTTP and the on-disk fallback format, so its fallback path is testable with the daemon absent.

### Toolchain

| Component | Version |
|-----------|---------|
| Go | 1.26.5 |
| PostgreSQL | 18.6 (`postgres:18-alpine`) |
| React | 19.2.8 |
| Vite | 8.2.1 |

Go is not currently installed on the development machine; installing it is a prerequisite task of
this change, not an assumption of it.

### Schema

```sql
CREATE TABLE projects (
  project_key        TEXT PRIMARY KEY,
  main_checkout_path TEXT NOT NULL,
  first_seen         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE changes (
  id                  BIGSERIAL PRIMARY KEY,
  project_key         TEXT NOT NULL REFERENCES projects(project_key),
  name                TEXT NOT NULL,
  state               TEXT NOT NULL,
  branch              TEXT,
  worktrees           JSONB NOT NULL DEFAULT '{}'::jsonb,
  artifact_url        TEXT,
  jira_issue          TEXT,
  planning_effort     TEXT,
  models              JSONB,
  review_panel_roster TEXT,
  pr_url              TEXT,
  updated_at          TIMESTAMPTZ NOT NULL,
  updated_by          TEXT NOT NULL,
  UNIQUE (project_key, name)
);

CREATE TABLE stage_runs (
  id          BIGSERIAL PRIMARY KEY,
  change_id   BIGINT NOT NULL REFERENCES changes(id),
  harness     TEXT NOT NULL,
  session_id  TEXT,
  command     TEXT NOT NULL,
  stage       TEXT NOT NULL,
  attempt     INT  NOT NULL,
  started_at  TIMESTAMPTZ NOT NULL,
  ended_at    TIMESTAMPTZ,
  outcome     TEXT,
  metrics     JSONB NOT NULL DEFAULT '{}'::jsonb,
  UNIQUE (change_id, command, stage, attempt)
);

CREATE INDEX stage_runs_metrics_gin ON stage_runs USING GIN (metrics);
CREATE INDEX stage_runs_started_at   ON stage_runs (started_at);

CREATE TABLE pricing (
  model                TEXT NOT NULL,
  effective_from       TIMESTAMPTZ NOT NULL,
  input_per_mtok       NUMERIC NOT NULL,
  output_per_mtok      NUMERIC NOT NULL,
  cache_write_per_mtok NUMERIC NOT NULL,
  cache_read_per_mtok  NUMERIC NOT NULL,
  PRIMARY KEY (model, effective_from)
);
```

`changes` carries the existing state file field for field, which is what keeps the contract rewrite
mechanical rather than a redesign. The unique key on `(change_id, command, stage, attempt)` makes
re-entrancy first-class: a `/myflow-do` fix is attempt 2 of its stages rather than an overwrite of
attempt 1, which is precisely what the rework-rate view counts.

**A state a stage run cannot be attributed to is stored anyway**, against a synthetic change row for
the project, rather than dropped. A mark that arrives for an unknown change is a defect worth seeing
in the data.

### The metrics bag

Keys written per stage run. Every one is optional; absence means not measured, which is distinct
from a recorded zero.

| Key | Source |
|-----|--------|
| `tokens.input`, `tokens.output`, `tokens.cache_creation`, `tokens.cache_read`, `tokens.thinking` | transcript `usage` |
| `tokens.main`, `tokens.sidechain` | transcript, split on `isSidechain` |
| `tokens_available` | `false` on harnesses with no transcript |
| `cost_usd`, `pricing_version` | computed from `pricing` at write time |
| `duration_ms` | mark timestamps |
| `model`, `effort` | transcript `message.model`, `effort` |
| `fast_mode`, `harness` | CLI invocation environment |
| `tool_calls`, `tool_calls_by_name` | transcript tool-use blocks |
| `files_touched`, `lines_added`, `lines_removed` | git, at stage end |
| `fix_rounds`, `panel_rounds`, `findings_by_severity` | `/myflow-do` mark arguments |
| `planning_effort`, `review_panel_roster` | carried from the change |

### Stage marks

Stage names come from **Level 1 — the stages of each command** (`README.md`), so the vocabulary in
the database and the vocabulary in the documentation are one list rather than two that can drift.

```bash
myflow stage begin --change <name> --command /myflow-do --stage "SDD + TDD per task"
myflow stage end   --change <name> --outcome completed
```

A run that dies leaves a stage open. The daemon sweeps stages whose session has been silent past a
timeout, setting `outcome = 'abandoned'`. An abandoned stage is a statistic worth having, not an
error to suppress.

### Harvesting

For each assistant message in `~/.claude/projects/*/*.jsonl` the harvester reads `timestamp`,
`message.model`, `effort`, `isSidechain` and `usage`, and assigns it to the open stage window whose
`session_id` matches and whose `[started_at, ended_at]` contains the timestamp. Main-thread and
sidechain totals accumulate separately, so subagent cost is visible rather than folded into its
parent's.

Attribution is idempotent: the harvester records the last consumed byte offset per transcript file,
so a restart re-reads nothing and double-counts nothing.

### Availability and reconciliation

Every CLI write attempts the daemon first with a short timeout. On any failure — daemon down,
Postgres down, timeout, non-2xx — it writes the existing on-disk JSON to the existing state path,
appends the intent to a journal file beside it, prints one `⚠ myflow: store unreachable — wrote
local journal` line, and **exits 0**. The pipeline continues exactly as it would have.

The daemon replays the journal at startup and on reconnect. Conflicts resolve by `updated_at`, with
the monotonic-state rule as the tiebreaker: a `FINISHED` already in the store is never overwritten
by a stale earlier state from a journal entry.

### API

```
GET    /api/v1/changes?project=&state=
GET    /api/v1/changes/{project}/{name}
PUT    /api/v1/changes/{project}/{name}
POST   /api/v1/stages/begin
POST   /api/v1/stages/end
GET    /api/v1/stats/{view}?from=&to=&project=
```

Every statistics view is period-parameterised at the API, not filtered in the client, so an
arbitrary period costs one query rather than a full download.

### The eight views

| View | Question it answers |
|------|---------------------|
| Live state board | Every open change across every project, with its state and next command |
| Cost per change | End-to-end tokens, dollars and wall clock for one change, broken down by command and stage |
| Stage leaderboard | Which stages cost the most — mean, median and p90 across a period |
| Trend over time | Is the pipeline getting cheaper per change as myflow changes? |
| Cache efficiency | cache-read against cache-creation per stage — the largest single cost lever |
| Panel economics | Findings per token by roster preset — whether `full` earns its cost over `light` |
| Model comparison | The same stage on different models: cost *and* subsequent rework |
| Rework rate | How often `/myflow-do` re-runs as a fix; abandoned-stage rate per stage |

Project is a first-class filter on every view, and cross-project comparison is supported, because
the state directory is already keyed per project and five projects already use it.

## Contract changes

`skills/myflow-contracts/state-file.md` stops describing a JSON file on disk and describes the
store, keeping its field vocabulary, its closed-schema rule, its monotonic-write rule and its
carry-forward rule unchanged. What changes is the mechanism: `jq` reads become `myflow state get`,
and full-object writes become `myflow state set`.

`skills/myflow-contracts/pipeline.md` gains the stage-mark obligation and cites the README's Level 1
list for the stage names rather than restating them.

Each `SKILL.md` gains mark calls at its own stage boundaries. `/myflow-status` reads the store
instead of the state directory.

## Migration

`myflow import` walks `/Users/tweety53/Agents/myflow/state/*/*.json`, inserting each as a `changes`
row — 24 files across 5 projects at time of writing. No stage history is backfilled because nothing
recorded any; statistics begin accumulating at the cutover, and the UI says so rather than showing
an empty chart that looks like zero cost.

The on-disk format survives as the journal format, so a change in flight during the cutover
completes on either path.

## Testing

- `store` is tested against a real PostgreSQL in a throwaway container, never a mock. The JSONB
  round-trip, the GIN-indexed aggregation queries and the unique-key behaviour on re-entrant stages
  are the things under test, and a mock would assert none of them.
- `harvest` is tested against committed transcript fixtures whose token totals are known by
  construction, including a sidechain fixture and a fixture whose messages straddle a window
  boundary.
- The CLI fallback is tested by pointing it at a dead port and asserting exit 0, the expected
  on-disk file, and the journal entry.
- Reconciliation is tested by replaying a journal against a store already advanced past it, and
  asserting the `FINISHED` row survives.
- The API's period parameters are tested at boundaries: a period containing no runs returns an empty
  result rather than an error, and a period wholly inside one stage run attributes it correctly.
