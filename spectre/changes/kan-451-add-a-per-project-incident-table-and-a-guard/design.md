# kan-451-add-a-per-project-incident-table-and-a-guard — design

## Context

The run record is store-native since KAN-258: `stats/internal/store/migrations/0010_run_records.sql`
holds `dispatches` and `findings` keyed by `changes(id)`, `stats/internal/records/types.go` is the
wire shape every layer shares, `stats/internal/store/records.go` writes and reads the rows,
`stats/internal/api/records.go` serves them under `/api/v1/records/{project}/{change}/…`,
`stats/internal/client/client.go` calls them, and `stats/cmd/flow/record.go` is the CLI. A record
write never blocks: on store failure it journals `{"kind","request"}` beside the change's state
file (`recordJournalPath`) and `stats/internal/reconcile/reconcile.go` replays the kinds it knows
(`dispatch`, `dispatch-end`, `finding`, `status`). A record read (`findings`) fails loudly, because
a JSON array a caller could mistake for "no rows" is worse than an error.

`scripts/check-unfinished-work.sh` already calls `flow record findings -change <n> -C <worktree>`
for its second signal, so a guard shelling out to `flow` is established; the project key resolves
from `-C` through `fallback.ProjectKey`. Its harness, `scripts/test-check-unfinished-work.sh`,
places a stub `flow` binary ahead of the real one on `PATH`. `scripts/gather-dispatch-context.sh`
builds the bundle every conductor and panel slot reads, appends a `project commands` section last,
and hashes the body to skip an unchanged rewrite; its harness runs with a restricted `PATH`.

`skills/flow/integrate.md` §1 is the one site where `check-unfinished-work.sh`'s `OUTSTANDING`
reaches an operator, offering exactly three courses. `spectre/specs/` is empty, so no capability
spec is edited.

## Change

### 1. Schema — `stats/internal/store/migrations/0018_guard_log.sql`

```sql unverified:to be written on the change branch
CREATE TABLE guard_verdicts (
  id                    BIGSERIAL PRIMARY KEY,
  change_id             BIGINT NOT NULL REFERENCES changes(id),
  guard                 TEXT NOT NULL,
  worktree              TEXT NOT NULL,
  verdict               TEXT NOT NULL,
  recorded_at           TIMESTAMPTZ NOT NULL,
  false_positive        BOOLEAN NOT NULL DEFAULT false,
  false_positive_reason TEXT,
  flagged_at            TIMESTAMPTZ
);
CREATE INDEX guard_verdicts_change_id ON guard_verdicts (change_id);

CREATE TABLE incidents (
  id           BIGSERIAL PRIMARY KEY,
  project_key  TEXT NOT NULL REFERENCES projects(project_key),
  change_id    BIGINT REFERENCES changes(id),
  guard        TEXT NOT NULL,
  symptom      TEXT NOT NULL,
  recovery     TEXT NOT NULL,
  minutes_lost INT NOT NULL CHECK (minutes_lost >= 0),
  occurred_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX incidents_project_key ON incidents (project_key);
```

`verdict` is the guard's whole verdict line, verbatim — `OUTSTANDING: <worktree> — <breakdown>` —
so the reason the guard gave survives without a second schema for breakdowns. `guard_verdicts`
carries no `project_key`: a verdict always belongs to a change, and the project is
`changes.project_key`. `incidents` carries one because `change_id` is nullable — an incident can be
recorded against a project with no change in flight, or after the change archived.

### 2. Wire types — `stats/internal/records/types.go`

`Verdict{ID, Change, Guard, Worktree, Verdict, RecordedAt, FalsePositive, FalsePositiveReason,
FlaggedAt}`, `VerdictFlag{Guard, Reason}`, `Incident{ID, Change, Guard, Symptom, Recovery,
MinutesLost, OccurredAt}`. `Change` is the change name on the read side (joined from `changes`)
and absent on the write side, where the URL names it.

### 3. Store — `stats/internal/store/records.go`

- `RecordVerdict(ctx, projectKey, change, in Verdict) (Verdict, error)` — insert; `RecordedAt`
  from the caller (the CLI stamps `time.Now()`), unknown change → the same not-found error
  `RecordDispatch` returns.
- `FlagVerdictFalsePositive(ctx, projectKey, change string, in VerdictFlag) (Verdict, error)` —
  `UPDATE … WHERE id = (SELECT id FROM guard_verdicts gv JOIN changes c … WHERE c.project_key=$1
  AND c.name=$2 AND gv.guard=$3 ORDER BY recorded_at DESC, id DESC LIMIT 1)`; zero rows updated →
  a not-found error the API maps to 404. Re-flagging the same row overwrites reason and
  `flagged_at`.
- `ListVerdicts(ctx, projectKey, guard string, falsePositiveOnly bool) ([]Verdict, error)` —
  newest first; `guard == ""` means every guard.
- `RecordIncident(ctx, projectKey string, in Incident) (Incident, error)` — `in.Change` non-empty
  resolves to `change_id` and is a not-found error when unknown; empty leaves it NULL.
  `OccurredAt` zero → the column default.
- `ListIncidents(ctx, projectKey string) ([]Incident, error)` — newest first.

Tests in `records_test.go` against the testcontainer store the existing record tests use.

### 4. API — `stats/internal/api/records.go`, `server.go`

| Route | Store call | Status |
|-------|-----------|--------|
| `POST /api/v1/records/{project}/{change}/verdicts` | `RecordVerdict` | 201 |
| `POST /api/v1/records/{project}/{change}/verdicts/false-positive` | `FlagVerdictFalsePositive` | 200; 404 when no verdict for that change+guard |
| `GET /api/v1/verdicts/{project}?guard=&falsePositive=true` | `ListVerdicts` | 200, JSON array |
| `POST /api/v1/incidents/{project}` | `RecordIncident` | 201 |
| `GET /api/v1/incidents/{project}` | `ListIncidents` | 200, JSON array |

`RecordStore` gains the five methods; the two lists validate `{project}` the way
`runRecord` does. Client methods mirror them one for one.

### 5. CLI — `stats/cmd/flow/record.go`

```text unverified:to be written on the change branch
flow record verdict [-addr url] [-timeout dur] [-C dir]
                    -change name -guard guard -worktree path -verdict line
flow record verdict false-positive [-addr url] [-timeout dur] [-C dir]
                    -change name -guard guard -reason text
flow record verdicts [-addr url] [-timeout dur] [-C dir] [-guard guard] [-false-positive]
flow record incident [-addr url] [-timeout dur] [-C dir]
                    -guard guard -symptom text -recovery text -minutes-lost n [-change name]
flow record incidents [-addr url] [-timeout dur] [-C dir]
```

`verdict`, `verdict false-positive` and `incident` are writes: never block, journal on store
failure with kinds `verdict`, `verdict-false-positive` and `incident`, exit 0. `incident` and
`incidents` take no `-change` identity flag (the project is the key), so they register the
`-addr`/`-timeout`/`-C` flags without `finishRecordIdentityFlags`' `-change` requirement.
`-minutes-lost` must parse as a non-negative integer, checked before the store is contacted, like
`-role`. `verdicts` and `incidents` are reads: JSON array on stdout, `[]` for no rows, stderr and
non-zero when the store cannot answer — `findings`' contract, verbatim. `internal/reconcile`
replays the three new kinds.

### 6. Guard — `scripts/check-unfinished-work.sh`

After `$REASONS` is final and before the verdict line is echoed:

```bash unverified:to be written on the change branch
VERDICT_LINE="CLEAR: $WORKTREE — every plan item is checked and no finding is open"
[ -n "$REASONS" ] && VERDICT_LINE="OUTSTANDING: $WORKTREE — $REASONS"
flow record verdict -change "$NAME" -guard check-unfinished-work -worktree "$WORKTREE" \
  -verdict "$VERDICT_LINE" -C "$WORKTREE" >/dev/null || true
if [ -n "$REASONS" ]; then
  PRIOR="$(flow record verdicts -guard check-unfinished-work -false-positive -C "$WORKTREE" 2>/dev/null)" \
    && N="$(printf '%s' "$PRIOR" | jq 'length' 2>/dev/null)" && [ "${N:-0}" -gt 0 ] \
    && echo "check-unfinished-work: prior false positives for this guard on this project: $N — last: $(printf '%s' "$PRIOR" | jq -r '.[0] | "\(.falsePositiveReason) (\(.change), \(.flaggedAt[:10]))"')" >&2
fi
echo "$VERDICT_LINE"
```

The write's own stdout and stderr are both discarded (`>/dev/null 2>&1`), consistent with the
`flow record findings` call two signals earlier in this same script, which likewise captures its
own stderr (to `$FINDINGS_ERR`) rather than letting it pass through raw, and because every `flow`
invocation also prints a `using FLOW_ADDR=...` diagnostic to stderr whenever that env var is set —
independent of success or failure — which would otherwise leak into any caller (this guard
included) that treats the script's combined output as authoritative. Neither call can change the
verdict line, the exit code, or the header's three exit-code meanings: `|| true` and the `&&` chain
make both advisory. The header gains a paragraph stating that. `test-check-unfinished-work.sh`'s stub `flow` answers `verdict` with exit 0 and `verdicts`
from a fixture file; cases: line printed at N>0 with the newest reason, nothing printed at N=0,
nothing printed and verdict unchanged when `verdicts` fails, verdict unchanged when `verdict`
fails, and the recorded `-verdict` argument equals the printed line.

### 7. Dispatch context — `scripts/gather-dispatch-context.sh`

A `## incidents` section after `project commands`, from
`flow record incidents -C "$WORKTREE_REAL"`, rendered with `jq -r` as:

```markdown unverified:to be written on the change branch
| when | guard | symptom | recovery | minutes lost | change |
|------|-------|---------|----------|--------------|--------|
| 2026-09-03 | check-task-commit-fields | Baseline revert re-entered mid-flight; planning dir wiped | aborted revert, restored dir from stash^3, dropped stashes | 55 | kan-423-… |
```

Census: found when the array is non-empty; `skipped: incidents (none)` on `[]`; `skipped:
incidents (flow unavailable)` when `flow` is off `PATH` or exits non-zero — never `refused`,
since a bundle must build with no store. The section is part of the hashed body, so a new incident
rebuilds the bundle. Harness cases in `test-gather-dispatch-context.sh` for each of the three
outcomes, with a stub `flow` in the case's `PATH` directory.

### 8. Skill text — `skills/flow/integrate.md`, `skills/flow-contracts/finish-contract-run1.md`

In `integrate.md` §1, the `OUTSTANDING:` bullet shows the guard's
`prior false positives …` stderr line alongside the breakdown when it printed, and the
**Continue — integrate anyway** course gains: when the operator's answer says the verdict was
verified structural (a plan held elsewhere, every task ticked, no open finding), run
`flow record verdict false-positive -change <name> -guard check-unfinished-work -reason "<their
reason>" -C <worktree>` once per worktree that reported `OUTSTANDING`, before proceeding.
`finish-contract-run1.md`'s Continue row gains one sentence pointing at that step rather than
restating it. `check-contract-budget.sh` budgets are raised only where the guard trips.

## Decisions

### The guard records its own verdict

**ID:** guard-records-own-verdict
**Status:** active
**Chosen:** `check-unfinished-work.sh` calls `flow record verdict` itself and, on `OUTSTANDING`,
queries prior false positives itself — no skill prose has to remember to log it.
**Considered:** the skill text (`integrate.md` §1) recording the verdict after running the guard —
depends on the conductor following prose, which is how memory notes went unread; every
verdict-emitting guard (`check-finish-preflight.sh`, `check-cleanup-complete.sh`,
`check-base-moved.sh`) via a shared `scripts/lib/record-verdict.sh` — the ticket names one guard
and no second caller exists yet, so the shared helper would be an abstraction with one user.

### Incidents are injected into the conductor dispatch

**ID:** incidents-in-dispatch-bundle
**Status:** active
**Chosen:** `gather-dispatch-context.sh` carries a `## incidents` section, so every conductor and
panel dispatch is handed the project's incidents; the API makes the table queryable.
**Considered:** an SPA per-project incidents view — adds a view, route, DTOs, tests and a visual
baseline for a table the ticket only asks to be queryable; API only — leaves the "memory note
nobody re-reads before dispatching" failure in place, which is the ticket's stated reason for the
table.

### A false-positive flag names the latest verdict for change and guard

**ID:** flag-latest-verdict
**Status:** active
**Chosen:** `flow record verdict false-positive -change <n> -guard <g>` flags the most recent
verdict row for that pair; the operator never looks up a row id.
**Considered:** flagging by verdict id — the id is store-allocated and the operator at the gate
does not have it; flagging every verdict of that change+guard — over-counts one adjudication as
several.

### Reads inside scripts are advisory

**ID:** script-reads-are-advisory
**Status:** active
**Chosen:** a failed or empty `flow record verdicts` / `flow record incidents` call changes
nothing a script answers — the verdict line and exit code in the guard, the bundle's build in the
context script.
**Considered:** treating a store failure as the guard's exit 2 the way signal two does — signal two
needs the store to answer the question; the prior-false-positive line is context beside the
answer, and a gate that cannot fire without a store would block integration for a hint.

### Incident rows are hand-written

**ID:** incident-written-by-hand
**Status:** active
**Chosen:** `flow record incident` is run by the operator or the parent session when an incident
is noticed; no stage prompts for one.
**Considered:** a prompt after `FINISHED` asking whether anything went wrong — the self-review
series ended at kan-380 because that prompt fires after the operator has walked away
(`.flow/project.md` `## self review`).

## Open questions
