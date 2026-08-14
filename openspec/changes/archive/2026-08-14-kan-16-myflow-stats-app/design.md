# Design — myflow stats app

Source: the approved brainstorming design at
`docs/superpowers/specs/2026-08-13-kan-16-myflow-stats-app-design.md`. That document is the
narrative record of the dialogue; this one is the technical design the tasks are derived from.

## Context

myflow records what state a change is in and nothing about what reaching that state cost. The state
record is a per-machine JSON file read with `jq`: it cannot be queried across changes, carries no
history, and is invisible outside a terminal. This change replaces it with a PostgreSQL-backed Go
service, records per-stage telemetry, and serves both surfaces through a browser UI.

## Goals

- The state record moves into PostgreSQL, with the same field vocabulary and the same closed-schema,
  monotonic-write and carry-forward rules the current contract states.
- Every `/myflow-*` stage is timed and costed, with real token numbers rather than self-reported
  ones.
- Statistics are queryable over an arbitrary period and filterable by project.
- **The pipeline never blocks on this subsystem**, at any point, for any reason.

## Non-goals

- Changing the three pipeline states, the command surface, the git boundaries, the review panel or
  the Jira contract. This change makes the pipeline observable; it does not change what it does.
- Telemetry parity across harnesses. Only Claude Code writes a machine-readable transcript.
- Remote access, multi-user access, or authentication.

## Architecture

### Components

Five units under a new top-level `stats/` directory.

| Unit | Path | Responsibility |
|------|------|----------------|
| `myflowd` | `stats/cmd/myflowd` | HTTP daemon: owns the Postgres pool; serves the state API, the stats API and the embedded SPA; hosts the harvester, the journal replayer and the abandoned-stage sweeper |
| `myflow` | `stats/cmd/myflow` | Thin CLI the skills invoke: `state get`, `state set`, `stage begin`, `stage end`, `import`, `journal flush` |
| `store` | `stats/internal/store` | Schema, migrations and queries — the only package that knows SQL |
| `harvest` | `stats/internal/harvest` | Tails Claude Code transcripts and attributes usage to stage windows |
| `web` | `stats/web` | React SPA, built by Vite, embedded into the binary with `go:embed` |

The SPA is embedded rather than served separately: one binary means no static host to run, no CORS
surface, and no way for UI and API to reach different versions.

### Boundaries

- `store` exposes typed repository methods and returns no SQL; the daemon may not build a query.
- `harvest` depends on an interface answering *which stage window is open for this session*, not on
  the database, so it is testable against fixtures with no PostgreSQL running.
- The CLI knows only HTTP and the on-disk fallback format, so its fallback path is testable with the
  daemon absent. **The one exception is `journal flush`**: it replays the journal directly into the
  store through `internal/store` and `internal/reconcile`, bypassing `myflowd`'s HTTP API entirely.
  Requiring a running daemon for a command whose whole purpose is to reconcile a journal written
  *because* the daemon (or the database) was unreachable would make it useless in exactly the
  situation it exists for — an operator asking, right now, "apply whatever is pending" has no daemon
  to ask through. This is a deliberate, narrow exception: it names `journal flush` alone, nothing else
  in the CLI depends on `internal/store`, and every other command keeps the boundary above intact.

Each unit answers the three questions the design brief asks: what it does, how it is used, and what
it depends on — and none of them requires reading another's internals.

### Toolchain

| Component | Version | Verified |
|-----------|---------|----------|
| Go | 1.26.5 | `go.dev/dl` on 2026-08-13 |
| PostgreSQL | 18.6 | `endoflife.date/api/postgresql.json` on 2026-08-13 |
| React | 19.2.8 | npm registry on 2026-08-13 |
| Vite | 8.2.1 | npm registry on 2026-08-13 |

Go is not installed on the development machine — `which go` returns nothing — so installing it is a
prerequisite task, not an assumption.

## Data model

```sql verified:authored in-tree for this change; validated by the store package's migration test against postgres:18-alpine
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
CREATE INDEX stage_runs_started_at  ON stage_runs (started_at);

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

`changes` carries the current state file field for field, which is what makes the contract rewrite
mechanical rather than a redesign.

### A change is one unit of work, across every repository it touches

A myflow change may span more than one repository, and the state record already treats that as
**one** record: a single scalar `branch` plus a `worktrees` map with one entry per affected
repository, per **Multi-repo shape** (`skills/myflow-contracts/state-file.md`). The store must read
the same way — a two-repository change is one row, never two.

`changes.project_key` therefore names the project whose **state directory holds the record**, exactly
as that directory is already keyed. It is not the list of affected repositories; `change_repos` is.

```sql unverified:confirm the FK cascade and the composite reference against postgres:18-alpine when task 2.1 lands
CREATE TABLE change_repos (
  change_id   BIGINT NOT NULL REFERENCES changes(id) ON DELETE CASCADE,
  repo_root   TEXT   NOT NULL,
  merge_base  TEXT,
  PRIMARY KEY (change_id, repo_root)
);

CREATE INDEX change_repos_repo_root ON change_repos (repo_root);
```

`stage_runs` carries a **nullable** `repo_root` referencing `(change_id, repo_root)`: set when a
stage ran inside one repository, `NULL` when the stage belongs to the change as a whole. `NULL`
means *the whole unit of work* and is never read as *unknown repository* — the same
absence-is-not-a-value rule the metrics bag follows.

A `NULL` `merge_base` means **no merge base recorded** for that repository. The state file contract
already states that a null merge base is a refusal to infer one, and nothing here computes one.

Every statistics view aggregates the unit by default, with a per-repository breakdown available on
request rather than as the default shape.

### Filtering, searching and sorting

Every field a reader would reasonably slice by is filterable, sortable and searchable: the typed
columns on both tables — project, name, state, branch, Jira issue, planning effort, roster, updated
at and by; command, stage, attempt, harness, outcome, started and ended at, duration, repository
root — **and** arbitrary keys inside the metrics bag, addressed by key path.

That breadth is the requirement. The mechanism that delivers it safely is an **allowlist**, and the
reason is specific: a sort key and a filter field are SQL **identifiers**, and an identifier cannot
be a bound parameter. Anything derived from request text and placed where an identifier goes is an
injection. So `internal/store` holds a fixed map from accepted field names to real columns, a field
absent from that map is a rejected request, and a metrics key path is validated against a syntax
rule and then bound as a parameter to a JSONB path operator. **That map is the entire dynamic-SQL
surface of the statistics layer**; a dynamic fragment built anywhere else is a defect, not a
shortcut.

Free-text search matches across the identity fields. Every list is paginated, and every sort is made
total with a unique tiebreaker, so paging cannot show a row twice or skip one. A run whose metrics
lack the sorted key orders as *absent*, distinctly from a recorded zero.

The unique key on `(change_id, command, stage, attempt)` makes re-entrancy first-class: a
`/myflow-do` fix is attempt 2 of its stages rather than an overwrite of attempt 1 — which is exactly
what the rework-rate view counts.

**A mark for an unknown change is stored, not dropped**, against a synthetic change row for the
project. A mark arriving for a change nobody created is a defect worth seeing in the data rather
than a record to discard.

### A replayed begin mark can open a second attempt

`BeginStage` has no idempotency key. A `stage begin` that times out on the client **after** the
server committed it is journalled and replayed, opening a genuine second attempt for the same
command and stage. The following `stage end` resolves to the highest open attempt, which is the
right one, so nothing is lost and no state moves backwards — but the earlier attempt stays open
until the sweeper closes it as abandoned.

The cost is therefore **an inflated rework count**, not a correctness hazard, and it is stated here
because the rework-rate view reads attempt numbers and abandoned outcomes directly. A reader
comparing rework across changes should know that a store outage during a run can add an attempt that
no operator caused.

Closing it needs a begin-mark idempotency key, its own schema and its own tests. That is deliberately
not carried here: the failure requires a timeout that lands after a successful commit, and the
resulting record is honest about what happened — an attempt really was opened.

### The metrics bag

Every key is optional. Absence means *not measured*, which is a different fact from a recorded zero
and is never collapsed into one.

| Key | Source |
|-----|--------|
| `tokens.input`, `tokens.output`, `tokens.cache_creation`, `tokens.cache_read`, `tokens.thinking` | transcript `usage` |
| `tokens.main`, `tokens.sidechain` | transcript, split on `isSidechain` |
| `tokens_available` | `false` on harnesses that write no transcript |
| `cost_usd`, `pricing_version` | computed from `pricing` at write time |
| `duration_ms` | mark timestamps |
| `model`, `effort` | transcript `message.model` and `effort` |
| `fast_mode`, `harness` | the CLI invocation's environment |
| `tool_calls`, `tool_calls_by_name` | transcript tool-use blocks |
| `files_touched`, `lines_added`, `lines_removed` | git, at stage end |
| `fix_rounds`, `panel_rounds`, `findings_by_severity` | `/myflow-do` mark arguments |
| `planning_effort`, `review_panel_roster` | carried from the change |

Adding a metric is writing a key. No migration, which is what the ticket's "flexible because I
expect frequent myflow updates" requires.

**Merging is deep, and that is load-bearing rather than a detail.** The bag is nested — every token
figure lives under one `tokens` object — and it is written by more than one party at different
times: the harvester merges token counts as it reads the transcript, a stage-end mark merges the
outcome, and pricing merges the cost. A **shallow** merge (Postgres's `||`, which concatenates only
at the top level) would let the later writer's `tokens` object replace the earlier writer's entirely,
silently discarding every sibling key under it. That is not a hypothetical: it was found by review on
task 3, reproduced against the live database, and would have destroyed `input`, `output` and
`cache_read` on every stage the moment the harvester landed.

A merge therefore combines objects **recursively**, and only a non-object value replaces. A writer
must never be required to know what other keys exist under a parent in order to avoid destroying
them — requiring that would make correctness depend on every future caller remembering an unwritten
rule, which is the property this contract exists to remove.

## Stage marks

Stage names are taken from **Level 1 — the stages of each command** (`README.md`), so the vocabulary
in the database is the documented one rather than a second list that can drift from it.

```bash unverified:confirm the final flag names once the CLI's argument parser lands in task 4
myflow stage begin --change kan-16-myflow-stats-app \
                   --command /myflow-do \
                   --stage "SDD + TDD per task"
myflow stage end   --change kan-16-myflow-stats-app --outcome completed
```

A run that dies leaves a stage open. The daemon sweeps stages whose session has been silent past a
timeout and sets `outcome = 'abandoned'`. An abandoned stage is a statistic worth having, not an
error to suppress — the rework-rate view reads it directly.

**`/myflow-fast` marks only its own four coarse Level 1 stages**, never the finer-grained
`/myflow-start` and `/myflow-do` stages it chains by citation. A change driven end to end by
`/myflow-fast` therefore has no per-stage detail in the store, and shows up in the stage leaderboard
and cost-per-change views at a much coarser grain than a change driven by the individual commands —
the two are not directly comparable, and a view mixing both kinds of run should say so rather than
imply they measure the same thing.

## Harvesting

For each assistant message in `~/.claude/projects/*/*.jsonl` the harvester reads `timestamp`,
`message.model`, `effort`, `isSidechain` and `usage`, and assigns it to the open stage window whose
`session_id` matches and whose `[started_at, ended_at]` interval contains the timestamp.

Main-thread and sidechain totals accumulate separately, so subagent cost is visible rather than
folded into its parent's.

**Subagent messages are not in the parent's file.** They are written to
`<session>/subagents/agent-*.jsonl`, a sibling directory beside the session transcript — while
carrying the **same** top-level `sessionId` as the parent. Attribution therefore matches on session
id *across files*, and discovery walks the directory rather than reading one file per session. This
was established by reading live transcripts, not from the format's documentation; a synthetic
fixture would have reproduced the `isSidechain` flag and missed the layout entirely. This is the reason marks alone were rejected as the metric source: an
agent cannot observe its own cache-read tokens, and cannot observe its subagents' usage at all.

The transcript's shape was confirmed by reading a live session file: each assistant record carries
`usage` with `input_tokens`, `cache_creation_input_tokens`, `cache_read_input_tokens`,
`output_tokens` and `output_tokens_details.thinking_tokens`, alongside top-level `timestamp`,
`sessionId`, `isSidechain`, `effort`, `gitBranch` and `cwd`.

Attribution is idempotent: the harvester records the last consumed byte offset per transcript file,
so a restart re-reads nothing and double-counts nothing.

## Availability and reconciliation

Every CLI write attempts the daemon first, with a short timeout. On any failure — daemon down,
database down, timeout, non-2xx — the CLI:

1. writes the existing on-disk JSON to the existing state path;
2. appends the intent to a journal file beside it;
3. prints one `⚠ myflow: store unreachable — wrote local journal` line;
4. **exits 0**.

The pipeline continues exactly as it would have. This is a stronger posture than the Jira contract's
"never a gate", and it has to be: a state write happens at the end of every command, so an outage
that stopped the write would strand changes at an unwritten state with the work already done.

The daemon replays the journal at startup and on reconnect. Conflicts resolve by `updated_at`, with
the monotonic-state rule as the tiebreaker: a `FINISHED` already in the store is **never**
overwritten by an earlier state arriving from a stale journal entry.

## API

```text unverified:confirm the final route shapes against the router once task 6 lands
GET  /api/v1/changes?project=&state=
GET  /api/v1/changes/{project}/{name}
PUT  /api/v1/changes/{project}/{name}
POST /api/v1/stages/begin
POST /api/v1/stages/end
GET  /api/v1/stats/{view}?from=&to=&project=
GET  /api/v1/stage-runs?filter=&q=&sort=&order=&limit=&offset=
```

`/stage-runs` is the stage-run counterpart of `/changes`: the same allowlisted filter, search, sort
and pagination surface, over the other table. It exists because the query layer's stage-run
allowlist would otherwise have no caller at all — a guarded surface nothing reaches is a surface
nobody notices breaking.

**`from` and `to` are required on every view, including the live state board.** Uniformity is worth
more here than saving a parameter on one view: a caller never has to remember which views take a
period, and a board rendered without one cannot silently mean something different from the same
board rendered with one.

**The per-repository breakdown is available on `cost-per-change` alone**, requested with
`breakdown=repo`. That is not an arbitrary limit: it is the only view keyed to a single change, and
therefore the only one with a single repository set to break down. The other seven aggregate across
many changes over a period, where "per repository" would mean something different and unstated.
Because it aggregates in Go rather than in SQL, it must scope by **project and name together** — a
change is keyed by both, and filtering on name alone silently blends same-named changes from
different projects into one answer.

Every statistics view is period-parameterised at the API rather than filtered in the client, so an
arbitrary period costs one query instead of a full download.

## The views

| View | Question it answers |
|------|---------------------|
| Live state board | Every open change across every project, with its state and its next command |
| Cost per change | End-to-end tokens, dollars and wall clock for one change, broken down by command and stage |
| Stage leaderboard | Which stages cost the most — mean, median and p90 across a period |
| Trend over time | Is the pipeline getting cheaper per change as myflow changes? |
| Cache efficiency | cache-read against cache-creation per stage — the largest single cost lever |
| Panel economics | Findings per token by roster preset — whether `full` earns its cost over `light` |
| Model comparison | The same stage on different models: cost *and* subsequent rework |
| Rework rate | How often `/myflow-do` re-runs as a fix; the abandoned-stage rate per stage |

Project is a first-class filter on every view, and cross-project comparison is supported: the state
directory is already keyed per project and five projects already use it.

## Contract changes

`skills/myflow-contracts/state-file.md` stops describing a JSON file and describes the store,
keeping its field vocabulary, its closed-schema rule, its monotonic-write rule and its carry-forward
rule unchanged. What changes is the mechanism — `jq` reads become `myflow state get`, and
full-object writes become `myflow state set`.

`skills/myflow-contracts/pipeline.md` gains the stage-mark obligation and cites the README's Level 1
list for stage names rather than restating them.

Each `SKILL.md` gains mark calls at its own stage boundaries. `/myflow-status` reads the store
instead of the state directory.

## Starting empty

The store starts with no records. Nothing imports the state files that existed before it, and no
statistics predate it.

That is a deliberate simplification, not an omission. An importer would have to reproduce the old
shell key derivation byte for byte, interpret files written under retired field spellings, and stay
correct while both mechanisms coexisted — a body of one-shot code whose only purpose is to carry
history that has no measurements attached to it anyway. No stage run was ever recorded before this
change, so an import would populate identity rows and nothing else.

Statistics therefore begin accumulating at first use, and a view covering a period before that says
so rather than rendering an empty chart that reads like zero cost.

The on-disk JSON keeps the state-file contract's shape because that is the simplest available format
for the write-ahead journal, **not** for compatibility with anything already on disk.

## Testing strategy

- `store` is tested against a real PostgreSQL in a throwaway container, never a mock. The JSONB
  round-trip, the GIN-indexed aggregation queries and the unique-key behaviour on re-entrant stages
  are the things under test, and a mock would assert none of them.
- `harvest` is tested against committed transcript fixtures whose token totals are known by
  construction — including a sidechain fixture and one whose messages straddle a window boundary.
- The CLI fallback is tested by pointing it at a dead port and asserting exit 0, the expected
  on-disk file, and the journal entry.
- Reconciliation is tested by replaying a journal against a store already advanced past it, and
  asserting the `FINISHED` row survives.
- The statistics API's period parameters are tested at their boundaries: a period containing no runs
  returns an empty result rather than an error, and a period wholly inside one stage run attributes
  it correctly.

## Decisions

### Scope: all four subsystems land in one change

**ID:** `scope-all-four`
**Status:** active
**Chosen:** State store, telemetry, aggregation and UI together — the pieces are not independently
useful. A state store with no statistics is a regression from a JSON file; statistics with no UI
cannot be read.
**Considered:** *Store and telemetry only* — produces no readable output, so the change could not be
evaluated by the person who asked for it. *Keep the state file and add statistics beside it* —
materially lower risk and delivers the part actually looked at, but running two sources of state
indefinitely is the outcome the ticket exists to end.

### Postgres is the source of truth, with a local write-ahead journal

**ID:** `source-of-truth-with-journal`
**Status:** active
**Chosen:** Commands write the store; on failure they write the on-disk JSON plus a journal entry
and exit 0, and the daemon replays the journal on reconnect. A real replacement without a database
outage stopping the pipeline mid-change.
**Considered:** *File remains authoritative, mirrored asynchronously* — safest, and matches the Jira
contract's posture, but it is not the replacement the ticket asks for. *Store authoritative, the run
stops on an outage* — what the ticket literally describes; ruled out because a state write ends every
command, so an outage strands changes with work already done.

### A long-running daemon owns the database

**ID:** `daemon-owns-db`
**Status:** active
**Chosen:** `myflowd` holds the pool and serves both APIs; CLI and SPA are clients. One writer means
the journal replayer, the harvester and the stage sweeper share one view of the data.
**Considered:** *CLI connecting to PostgreSQL directly, with a separate serve mode* — the closest fit
to how skills already shell out to `scripts/*.sh`, but the harvester and the sweeper must run
continuously and would have no host. *Skills calling `curl` directly* — puts inline HTTP and error
handling into every contract file.

### Stage boundaries are explicit marks; metrics come from the transcript

**ID:** `marks-plus-harvest`
**Status:** active
**Chosen:** Marks define the window exactly and harness-independently; the harvester supplies the
real numbers.
**Considered:** *The end mark carries numbers the agent supplies* — an agent cannot see its own
cache reads or its subagents' usage, so the headline figure would be wrong in the direction that
matters most. *Inference from the transcript alone* — needs no skill changes, but stage attribution
becomes a heuristic that breaks on any rewording and drifts from the documented stage names.

### Typed identity, JSONB metrics

**ID:** `typed-core-jsonb-metrics`
**Status:** active
**Chosen:** Typed columns for identity and lifecycle, a GIN-indexed JSONB bag for measurement.
**Considered:** *Fully typed columns per metric* — best constraints, worst fit for a ticket whose
premise is frequent change. *Append-only event log with everything derived* — maximum flexibility and
full replay, but every statistic becomes a query to write before it can be shown.

### Cost is frozen per row and re-priceable

**ID:** `pricing-version`
**Status:** active
**Chosen:** Store token counts, a derived USD cost, and the `pricing` row version that produced it.
The dashboard is stable and history can be recomputed when prices change.
**Considered:** *Tokens only, priced at query time* — survives price changes but makes every
historical figure move under the reader. *A frozen figure with no version* — stable, but silently
wrong once prices change and impossible to correct.

### Claude Code gets full telemetry; other harnesses degrade honestly

**ID:** `harness-degradation`
**Status:** active
**Chosen:** Cursor and Codex record state, marks and wall clock; token metrics are recorded as
explicitly unavailable, never as zero.
**Considered:** *Claude Code only, as a hard requirement* — a Cursor run would write state the app
could not interpret. *All three harnesses now* — neither of the other two has a documented transcript
format, so this is open-ended research inside an already large change.

### The Go application lives in this repository

**ID:** `same-repo`
**Status:** active
**Chosen:** A new top-level `stats/` directory. The state contract and the application implementing
it version together: renaming a stage mark and updating the skill that emits it is one commit.
**Considered:** *A separate repository* — keeps this one a pure prompt/config repo, but splits every
contract change across two pull requests. *A submodule* — separate history in one checkout, at the
cost of submodule friction in every worktree `/myflow-do` creates, which is constant.

### Own docker-compose stack, daemon under launchd

**ID:** `dedicated-stack`
**Status:** active
**Chosen:** A dedicated `myflow-postgres` container with its own port and volume, and `myflowd` under
a launchd user agent.
**Considered:** *Reuse the running `gymie-postgres` container* — nothing new to run, but it couples
the pipeline's store to a project stack that gets stopped, so the journal fallback would fire
routinely rather than exceptionally. *Start the daemon on demand from the CLI* — no launchd, but adds
first-call latency and a startup race when parallel worktrees invoke it at once.

### Localhost-only, no authentication

**ID:** `localhost-no-auth`
**Status:** active
**Chosen:** Bind `127.0.0.1` on a fixed port, no credentials. A single-user dashboard on the user's
own machine, holding no secrets beyond pipeline metadata.
**Considered:** *A shared token* — guards against other local processes, at the cost of token
handling in every client; disproportionate for the data held. *LAN-reachable with real auth* — needs
auth, TLS and a hardening pass, a materially larger change than the one asked for.

### Existing state files are imported and their format becomes the journal

**ID:** `import-and-reuse-format`
**Status:** superseded by `start-empty`
**Chosen:** A one-shot import of every existing file, after which the same on-disk format serves as
the write-ahead journal. A change in flight during the cutover finishes on either path.
**Considered:** *Import once, then delete the files* — cleanest end state, but the chosen fallback
needs a local format regardless, so this would mean inventing a second one. *Start empty* — no
import; the two paths then coexist with no end date.

### The store starts empty; nothing is imported

**ID:** `start-empty`
**Status:** active
**Chosen:** No importer. The store begins with no records, statistics accumulate from first use, and
a period before that reports *not recorded* rather than zero.
**Considered:** *A one-shot import of the existing state files, with their format doubling as the
journal* — this was the original decision (`import-and-reuse-format`, superseded here). It was
withdrawn because the history it would carry has no measurements attached: no stage run existed
before this change, so an import would populate identity rows alone, while the importer would have to
reproduce the retired shell key derivation exactly, tolerate retired field spellings, and remain
correct while both mechanisms coexisted. *Start empty but keep reading old files as a fallback* —
same coupling, spread across every read path instead of one command.

## Open questions

*(none — every question raised during brainstorming was answered before the design was written)*
