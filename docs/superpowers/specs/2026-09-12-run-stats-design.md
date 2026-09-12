# Run stats: plan sessions in the store, per-run observability, main/subagent breakdown

## Goal

Every `/flow-plan` session lands in the stats store and is linked to the `/flow` runs of the
change it seeded. Every run shows, as one row, the numbers needed to tune the pipeline later:
tokens, cost, wall clock, compactions, dynamic decisions, and the observability signals listed
under **Metrics**. Expanding a row shows the main session and every subagent dispatch separately.

## Vocabulary

- **Run** — one skill invocation: the set of `stage_runs` rows sharing one `session_token`
  (`mf-<token>` for `/flow`, `ff-<token>` for `/flow-fast`, `fp-<token>` for `/flow-plan`).
- **Window** — the transcript interval a stage run or dispatch owns; the unit the harvester
  attributes messages to (`internal/harvest/attribute.go`).
- **Declared** model/effort — what the dispatch row says was asked for (`dispatches.model`,
  `dispatches.effort`). **Served** model / **observed** effort — what the transcript's assistant
  messages carry (`Record.Model`, `Record.Effort`).

## 1. Plan sessions in the store

### Schema (`0023_plan_sessions.sql`)

- `stage_runs.change_id` becomes nullable.
- New columns `stage_runs.project_key TEXT NULL REFERENCES projects(project_key)` and
  `stage_runs.jira_key TEXT NULL`, index on `(project_key, jira_key) WHERE change_id IS NULL`.
- Check constraint: `change_id IS NOT NULL OR (project_key IS NOT NULL AND jira_key IS NOT NULL)`.
  Both new columns are set only by a plan-session mark and kept after backfill; the project
  bootstrap `PutChange` already performs is extracted so the plan-session insert shares it.
- `QueryStageRuns` joins `changes` with `LEFT JOIN` and reads `change_id` through `COALESCE(…, 0)`,
  so a harvest window lookup (session_id filter) and the plan-session end mark (jira_key filter)
  both see unattached rows. Every other aggregate keeps its inner join and simply omits them.
- `dispatches.change_id` stays `NOT NULL`; a plan session dispatches nothing.

### Write path

`flow stage begin|end` accepts `-jira-key <KEY>` as an alternative to the positional change
name. With `-jira-key`, `BeginStageInput.ChangeName` is empty, `JiraKey` is set, and the store
inserts the row with `change_id NULL` (or the id of an existing change carrying that
`jira_issue`). Command is `/flow-plan`, stage is `plan.session`, attempt numbered per
`(project_key, jira_key, command, stage)`. `-jira-key` without a change name is accepted only
for stage `plan.session`; any other stage keeps requiring a change name. `plan.session` joins
`internal/stages.Table` and README's Level 1 table as the one stage of a new `/flow-plan`
command, so the documented-stage guard on both the CLI and the daemon admits it.

### Backfill

`PutChange` runs, in the same transaction as the insert of a new change row whose `jira_issue`
is set:

```sql
UPDATE stage_runs SET change_id = $new_id
WHERE change_id IS NULL AND jira_key = $jira_issue AND project_key = $project_key;
```

`jira_key` is kept after backfill; it is the audit trail that the row arrived before the change.
A later plan session for a change that already exists still records with `-jira-key` and is
backfilled at insert time by the same `WHERE` applied once in `BeginStage` when a matching
change already exists. A plan session whose Jira key never becomes a change stays unattached
and is listed under its key alone.

### Harvest

Unchanged. The harvester binds `session_token` → `session_id` and attributes tokens into the
window exactly as for any stage run. `store.Price` prices the row the same way.

### Skill change (`skills/flow-plan/SKILL.md`)

- Generate `fp-<literal-token>` once at entry, like `/flow` generates `mf-<token>`.
- First action after resolving the Jira key:
  `flow stage begin -command '/flow-plan' -stage plan.session -harness <harness> -session-token fp-<literal-token> -jira-key <KEY>`
- Last action, on every exit path including "nothing to stage":
  `flow stage end -session-token fp-<literal-token> -outcome <staged|abandoned>`
- A plan session with no Jira key marks nothing; the contract sentence "records nothing to the
  store" is narrowed to that case. Lines `SKILL.md:41-42` and `:268-269` are rewritten to say so.
- `/flow` kickoff needs no change: the backfill is store-side.

## 2. Metrics

All new metrics are written by the harvester into the existing `metrics` JSONB bag on both
`stage_runs` and `dispatches`, through `CommitHarvestBatch` and `jsonb_deep_add`. No new
tables, no new hooks. They live under one `signals` key, split `main` / `sidechain` exactly as
`tokens` is, so a stage run's main-session figures never include its subagents' and a dispatch
row (sidechain only) can be read on its own. `jsonb_deep_add` sums numbers, replaces strings and
replaces arrays, which fixes three shapes below: every counter is a number, the compaction event
list is an object keyed by the event's RFC 3339 timestamp (two batches never carry the same
event, so keys never collide), and `context_end` is a string so the latest batch wins.

### New keys, per window, under `signals.main` and `signals.sidechain`

| Key | Source record | Shape |
|-----|---------------|-------|
| `compactions` | `type:"system", subtype:"compact_boundary"` | int |
| `compaction_events.<rfc3339nano>` | `compactMetadata` | `{trigger, pre_tokens, post_tokens, duration_ms}` |
| `turns` | `type:"system", subtype:"turn_duration"` | int |
| `turn_duration_ms` | `turn_duration.durationMs` | int, summed |
| `turn_messages` | `turn_duration.messageCount` | int, summed |
| `tool_calls.<name>` | assistant content block `type:"tool_use"`, deduplicated by block `id` | int per tool name |
| `tool_calls_total` | same | int |
| `tool_errors` | user content block `type:"tool_result", is_error:true`, deduplicated by `tool_use_id` | int |
| `denials` | `is_error` result whose text starts with one of the denial phrases | int |
| `api_errors` | assistant line with `isApiErrorMessage:true` | int |
| `context_end` | last assistant `usage` in the window by timestamp | decimal string of `input + cache_read + cache_creation` |
| `served_models.<model>` | assistant `message.model`, lowercased, `<synthetic>` excluded | message count per model |
| `served_efforts.<effort>` | line-level `effort` | message count per effort |

Denial phrases (prefix match, case-sensitive, from live transcripts):
`Permission for this action was denied`, `The user doesn't want to proceed`,
`PreToolUse hook`. Add a phrase only with a transcript sample beside it in the test fixture.

Subagent transcripts (`subagents/agent-<id>.jsonl`) carry the same record shapes; the dispatch
row receives `signals.sidechain`, and the owning stage run's `dispatches.<agentId>.signals`
carries the same figures beside that dispatch's tokens. `compactions` on a dispatch is expected
to stay zero and is recorded anyway.

### Already captured, only surfaced

Input, output, thinking, cache read, cache write 5m/1h, speed, `tokens.main` /
`tokens.sidechain`, per-model buckets, per-dispatch buckets with `AgentType`, `Description`,
`SpawnDepth`, declared model and effort on `dispatches`, cost via `store.Price`, stage wall
clock from `started_at`/`ended_at`, the decision JSON from `decisions`, findings, guard verdicts,
incidents, suite runs.

### Derived at query time

Computed by the new aggregate, never stored:

- **Human gate time** — per run, the summed duration of the stage runs that wait on the
  operator: `flow.design-approval`, `flow.landing-question`, `flow.unfinished-work-gate`.
  Per change, additionally **idle between runs**: sum of `next_run.first.started_at -
  prev_run.last.ended_at` over consecutive runs of the change, which is the review-and-test gate
  the pipeline places between runs. Agent clock is total wall clock minus human gate time; total
  wall clock is `last.ended_at - first.started_at` of the run.
- **Fix iterations** — per change, count of runs whose `command` is `/flow` and whose first stage
  is not `flow.kickoff`.
- **Fan-out width** — per run, the most dispatches open at one instant, from a sweep over the
  run's dispatch intervals.
- **Findings per dispatch** — from `findings.dispatch_id`: raised, and the same rows counted by
  their current `status` (the pipeline's own vocabulary, `open`, `fixed`, …), shown on the
  dispatch row.
- **Suite first-pass** — `suite_runs` carries no change or token, so per run it is the project's
  suite runs whose `ran_at` falls inside the run's span: their count, and whether the earliest
  exited 0. Absent when none fall inside.
- **Cache hit ratio** — `cache_read / (input + cache_read + cache_creation)`.
- **Model mismatch** — declared model absent from `served_models`, or declared effort absent
  from `served_efforts`, on a dispatch whose `served_models` is non-empty; shown as a flag on
  the dispatch row.
- **Main-session cost** — run cost minus the sum of its dispatches' cost, since `store.Price`
  prices a stage run whole and each `dispatches.<agentId>` bucket separately.

## 3. API

`GET /api/v1/stats/runs?from=<rfc3339>&to=<rfc3339>[&project=<key>][&change=<name>]` is a new
`viewName` on the existing stats endpoint, answered in the standard `statsResponse` envelope with
`rows` shaped:

```json
{
  "changes": [{
    "project": "agents", "change": "kan-172-run-stats", "jiraKey": "KAN-172",
    "totals": { ...RunTotals },
    "runs": [{
      "sessionToken": "fp-…", "kind": "plan|flow|flow-fast", "command": "/flow-plan",
      "startedAt": "…", "endedAt": "…",
      "totals": { ...RunTotals },
      "decision": { "execution": "…", "implementer": "…", "panel": "…" } | null,
      "fanOutMax": 2, "suiteRuns": 3, "suiteFirstPass": false | true | null,
      "main": { ...RunTotals },
      "dispatches": [{
        "seq": 1, "role": "…", "slot": "…", "agentType": "flow-medium", "description": "…",
        "depth": 1, "declaredModel": "…", "servedModels": {...}, "declaredEffort": "…",
        "servedEfforts": {...}, "mismatch": false,
        "findingsRaised": 3, "findingsByStatus": {"fixed": 2, "open": 1},
        "startedAt": "…", "endedAt": "…", "totals": { ...RunTotals }
      }]
    }]
  }]
}
```

`RunTotals`: `inputTokens, outputTokens, thinkingTokens, cacheRead, cacheWrite5m,
cacheWrite1h, cacheHitRatio, cost, priced (bool), wallClockMs, agentClockMs, humanGateMs,
idleBetweenRunsMs (change level only), fixIterations (change level only),
compactions, turns, toolCalls, toolErrors, denials, apiErrors, contextEnd`. `priced` is false
when any window in the total was refused by `store.Price`; the cost column then shows the
`Unavailable` component with the refusal reason from `cost-status`.

Unattached plan sessions are returned as a change entry with `change: null` and the `jiraKey`.
`ListRuns` lives in `internal/store/aggregate.go` beside `Decisions()`; the handler joins the
`stats.go` slug table as `runs`.

## 4. SPA

New route `runs` → `views/Runs.tsx`, added to the static view list in `App.tsx` and to the nav.
`RunDetail` stays as the per-stage view and gains a link from each run row.

Layout: one `DataTable` per change, header row = change name, Jira key, change totals. Body =
one row per run. Columns: kind, command, started, tokens in, tokens out, cache hit, cost, wall
clock, human gate, compactions, turns, tool calls, errors (tool errors + api errors, denials in
the tooltip), fan-out, suite first pass, decision (execution / implementer / panel, blank for
plan runs). The change header
row additionally shows idle between runs and fix iterations.

Expand (existing `renderDetail` + `rowKey` on `DataTable`): first row **main session** with the
run's `main` totals; then one row per dispatch: seq, role/slot, agent type, description, depth,
model (declared, with served in parentheses when they differ, flagged red on mismatch), effort
(same treatment), findings (raised, by status), tokens in/out, cache hit, cost, wall clock,
turns, tool calls, errors, compactions, context at end. Nested rows reuse `DispatchTable`'s cell formatters; the table itself is extended
with the new columns rather than duplicated.

Filters: `ProjectFilter`, `ChangeVariable`, `PeriodPicker`, as on the other views.

## 5. Error handling

- A `stage begin` for `plan.session` when the daemon is down journals like every other mark and
  exits 0; reconcile replays it with the `-jira-key` field.
- Backfill is idempotent: re-running `PutChange` for an existing change touches no rows.
- Harvest parse errors on a malformed `compactMetadata` count the compaction and skip the event
  detail.
- Unpriced windows never block the row; `priced:false` is rendered, not an error.

## 6. Testing

- `internal/harvest`: fixture JSONL with one `compact_boundary`, three `turn_duration`, two
  `tool_use` of different names, one `is_error` tool result per denial phrase plus one plain
  error, one `isApiErrorMessage`. Assert every new key. Same fixture shape for a subagent file.
- `internal/store`: migration up/down; `BeginStage` with `-jira-key`; backfill on `PutChange`
  and on a late plan session; `ListRuns` totals, human gate time, fix iterations, mismatch flag.
- `cmd/flow`: `stage begin -jira-key` accepted for `plan.session`, rejected for any other stage.
- `stats/web`: `Runs.tsx` visual baseline from `uitest-seed` extended with one change carrying a
  plan run, a creating run with two dispatches (one mismatched), and a fix run.
- Lint: `gofmt -l`, `go vet`, `npx tsc -b`, `scripts/check-*.sh` all clean.

## Out of scope

- No Claude Code hooks; every signal is transcript-derived.
- No per-stage grouping inside the expanded row (rejected: more clicks for the same numbers;
  `RunDetail` already shows per-stage).
- No separate `plan_sessions` table (rejected: duplicates harvest, pricing and UI code paths for
  one row kind).
- No provisional change row from `/flow-plan` (rejected: the change name needs the slug the plan
  does not have).
