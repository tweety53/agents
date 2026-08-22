# Store-native run record — design

**Jira:** KAN-258 (umbrella; subsumes KAN-237, KAN-218, KAN-212, KAN-155; removes or reshapes
KAN-77, KAN-65, KAN-205, KAN-179)
**Change:** `kan-258-store-native-run-record`
**Date:** 2026-08-22

## Problem

The pipeline's derived records are files in a gitignored worktree. A file has a path, and no
component agrees on it: the writer never names the SDD ledger's path, `preserve-session-records.sh`
reads one, the artifacts registry names a directory. When they disagree, preservation prints
`skipped: <src> (absent)` and exits 0 — which is also the correct output for a change that
legitimately has no such record. **"Never written" and "written somewhere else" are indistinguishable
to every caller**, and the record dies with the worktree.

Measured in this repository: 34 preserved ledgers against 40 preserved panel records; two ledgers
carrying hand-typed filenames the self-review gather's search pattern cannot match;
`kan-13-myflow-planning-and-status-fixes` lost its ledger outright.

In a store there is no path to disagree about, and "no rows for this run" is a different value from
"rows exist and nobody looked in the right place".

## Scope

**In:** all four record types the umbrella names — the SDD ledger, review-panel findings,
per-dispatch cost, and task → commit → model. The first, third and fourth are one record, not three;
see decision `dispatches-is-one-table`.

**Out:**

- `tasks.md`, `proposal.md`, `design.md` and the delta specs. They are authored artifacts, git-
  versioned, PR-reviewed and read as prose by implementer subagents. They stay Markdown.
- Task **completion** state. `tasks.md`'s checkbox stays the single source, and
  `check-unfinished-work.sh` keeps reading it.
- The SPA. No new view, no new route, no change to `RunDetail.tsx`.
- Backfill. The store starts empty; the 34 archived ledgers and the preserved reviews stay as
  committed Markdown, unparsed.

## Data model

One migration, `stats/internal/store/migrations/0010_run_records.sql`, adding two tables.

### `dispatches`

One row per subagent dispatch. This single table is the SDD ledger, KAN-237's task → commit → model
and KAN-212's per-dispatch cost.

| Column | Type | Notes |
|--------|------|-------|
| `id` | `BIGSERIAL PRIMARY KEY` | |
| `change_id` | `BIGINT NOT NULL REFERENCES changes(id)` | |
| `stage_run_id` | `BIGINT REFERENCES stage_runs(id)` | nullable — a dispatch outside a marked stage still records |
| `seq` | `INT NOT NULL` | append order within the change; `UNIQUE (change_id, seq)` |
| `task_id` | `TEXT` | dotted id from the `tasks.md` heading; NULL for roles that run against no task |
| `role` | `TEXT NOT NULL` | `implementer` · `reviewer` · `panel-fix` · `red-partner` |
| `slot` | `TEXT` | panel slot name; NULL for implementers |
| `model` | `TEXT NOT NULL` | recorded intent, or the literal `unknown (agent-defined)` |
| `commit_sha` | `TEXT` | |
| `outcome` | `TEXT` | |
| `session_token` | `TEXT` | the run's own token, for harvest binding |
| `started_at` | `TIMESTAMPTZ NOT NULL` | |
| `ended_at` | `TIMESTAMPTZ` | |
| `metrics` | `JSONB NOT NULL DEFAULT '{}'` | the harvester's bag, same shape as `stage_runs.metrics` |
| `notes` | `TEXT` | |

Indexes: `(change_id)`, `(model)`, `(role)`, `(task_id)`, `(started_at)`, and a GIN index on
`metrics` — the same set `stage_runs` carries, for the same reasons.

`UNIQUE (change_id, seq)` is the race detector, exactly as `stage_runs_attempt_key` is for
`BeginStage`: `RecordDispatch` allocates the next `seq` inside a single `INSERT … SELECT`, and two
concurrent inserts that compute the same next value collide rather than one silently overwriting the
other.

### `findings`

One row per review-panel finding.

| Column | Type | Notes |
|--------|------|-------|
| `id` | `BIGSERIAL PRIMARY KEY` | |
| `change_id` | `BIGINT NOT NULL REFERENCES changes(id)` | |
| `dispatch_id` | `BIGINT REFERENCES dispatches(id)` | the slot that raised it |
| `ref` | `TEXT NOT NULL` | `F1`, `F2`, … ; `UNIQUE (change_id, ref)` |
| `round` | `INT NOT NULL` | `0` = the initial panel; `1..n` = fix rounds |
| `slot` | `TEXT NOT NULL` | |
| `severity` | `TEXT NOT NULL` | |
| `location` | `TEXT` | |
| `note` | `TEXT NOT NULL` | |
| `status` | `TEXT NOT NULL` | `open` · `fixed` · the record's other status words |
| `reproducer` | `TEXT` | the command, or `none — <reason>` |

Indexes: `(change_id)`, `(severity)`, `(status)`, `(slot)`.

`ref` is unique per change, not per round: a fix round **updates** a row's `status` in place. Nothing
accumulates, which is why KAN-65 (`final-review-panel.md` is cumulative and so can never satisfy its
own guard) stops being reachable.

## Write path

Four layers, identical to the ones `stage begin` already traverses: `cmd/myflow` →
`internal/client` → `internal/api` → `internal/store`.

```
myflow record dispatch  -change <name> -task <id> -role <role> [-slot <s>] -model <m> \
                        [-commit <sha>] [-outcome <o>] -session-token <tok> \
                        -started-at <ts> [-C <dir>]
myflow record finding    -change <name> -ref F<n> -round <r> -slot <s> -severity <sev> \
                        [-location <l>] -status <st> [-reproducer <cmd>] -note <text>
myflow record status     -change <name> -ref F<n> -status <st>
myflow record render     -change <name> -kind ledger|panel|all -repo <abs-repo-root>
```

**One call per dispatch, at dispatch close.** The parent already waits for a dispatch's commit sha
before dispatching the next implementer into the same worktree, so at close it knows the task, the
role, the model, the commit and the outcome. A begin/end pair would buy nothing and double the call
count.

**Endpoints:** `POST /api/v1/records/dispatches`, `POST /api/v1/records/findings`,
`PATCH /api/v1/records/findings/{ref}`, `GET /api/v1/records/{project}/{change}`.

**Store unreachable → journal, warn once, exit 0.** `fallback.AppendJournalEntry` is reused
unchanged; the journal file is `<state-dir>/<project-key>/<name>.journal.record`, following
`stage.go`'s existing `.journal.stage` suffix. `internal/reconcile` gains `replayRecordFile`
alongside its existing `replayFile` and `replayStageFile`. **Never branch on `myflow record`'s exit
code as a signal about the record** — a write that could not reach the store still exits 0, the same
guarantee `state set` and `stage begin` already give.

**The silence is closed at the handoff, not at the call.** `/myflow-do`'s handoff block gains a
`Records:` line naming how many record writes for this change are still sitting in the journal. A
never-block guarantee that is also silent is precisely the failure this change exists to remove.

**Caller mistakes still exit non-zero.** An unknown role, a missing required flag, or a
`-session-token` carrying a shell substitution is a defect in the call, not an outcome of the run —
the same split `stage begin` already draws, and `check-stage-mark-calls.sh`'s substitution check is
extended to cover `myflow record` call sites in skill source.

## Cost attribution

`internal/harvest` already attributes a transcript's token usage to a `Window` —
`(session_id, [started_at, ended_at))`, half-open — and already splits main-thread from sidechain and
buckets per model.

A dispatch is a sidechain inside a stage's window. Attribution therefore adds a **second, independent
pass** over the same records, against dispatch windows resolved the same way stage windows are.
`stage_runs.metrics` is written exactly as it is today and its values do not change; `dispatches.metrics`
is new. The two are different grains over the same usage, not a double count, and nothing about the
existing pass is modified.

`WindowSource` gains a sibling `DispatchWindowSource` rather than growing a kind discriminator on
`Window`: the consumer-side interface stays one method, and `internal/harvest` continues to import
nothing from `internal/store`.

## Rendering

`myflow record render` writes Markdown **straight into the repository** — no worktree copy anywhere
in the path:

| Kind | Destination |
|------|-------------|
| `ledger` | `<repo>/docs/superpowers/ledgers/<YYYY-MM-DD>-<change>.md` |
| `panel` | `<repo>/docs/superpowers/reviews/<YYYY-MM-DD>-<change>-panel.md` |

The date is fixed at the first render for a change: an existing file is reused, so a fix round
overwrites in place rather than leaving one dated duplicate per round. That rule is carried over
verbatim from `preserve-session-records.sh`, which is the only thing about it that survives.

**`-kind panel` always writes; `MISSING:` never applies to it.** A panel that raised nothing produces
no finding rows, and reporting `MISSING:` would leave no record — which `check-unfinished-work.sh`
reads as OUTSTANDING for a genuinely clean change. `myflow-review-panel-economics` already settles it:
a panel that raised no finding says so with `findings-total: 0`, a declaration, where silence is not.
The command runs at panel close, so the invocation is itself the evidence a panel ran.

**When each render fires:**

- **Panel record — at panel close**, before `check-unfinished-work.sh` runs. The guard reads the same
  rendered marker block it reads today and needs no change at all. This is also KAN-205 ("validate
  the panel record's marker block at panel close, not at finish run 1") satisfied as a consequence.
- **Ledger — at finish run 1**, where preservation used to happen.
- **Either, on demand**, via the command directly.

**Outcome words**, printed one per kind, replacing `preserve-session-records.sh`'s four:

| Word | Means | Exit |
|------|-------|------|
| `rendered: <dest>` | rows existed and the file was written | 0 |
| `MISSING: <kind> — no rows for <change>` | the store holds nothing of this kind | 0 |
| `journalled: <kind>` | the store was unreachable; nothing was written | 0 |
| a message on stderr | the destination was refused or could not be written | non-zero |

`MISSING:` now means what it says. The ambiguity this whole change exists to close was that
`skipped: (absent)` could not distinguish a change with no record from a record written elsewhere;
there is no "elsewhere" for a table.

**The path protections `preserve-session-records.sh` carried move with the duty, in Go.** The change
name is validated against the same allowlist (one leading alphanumeric, then letters, digits, `.`,
`_`, `-`), and the destination is required to be contained within the repository root — both now
enforced in the renderer rather than by an accident of shell string concatenation.

## Contract and skill changes

`preserve-session-records.sh` and `scripts/test-preserve-session-records.sh` are **deleted**. What
moves with them:

- **`pipeline.md`** — the "Preserving the session records" section and its outcome table are replaced
  by a render-outcome table; the `Panel record` and `SDD ledger` rows in the **Temporary artifacts
  registry** change their "Removed by" column, since neither is preserved out of the worktree any
  more; **Model policy**'s sentence naming the preserved ledger as the audit trail is repointed at
  the store.
- **`finish-contract.md`** — run 1's preservation duty becomes a render duty.
- **`myflow-do/SKILL.md`** — sections 4, 5 and 7: the ledger and panel-record *writing* instructions
  become `myflow record …` calls; the `prUrl` push path's preservation call becomes a render call;
  section 7's "confirm this run produced a ledger" check becomes a store query. The instruction that
  `superpowers:subagent-driven-development`'s workspace script writes the ledger to
  `.superpowers/sdd/tasks/progress.md` is removed — the parent records each dispatch instead.
- **`myflow-finish/SKILL.md`** — run 1's `finish.preserve-sessions` stage keeps its key and its
  place; only what it invokes changes.
- **`handoff-blocks.md`** — the `IN_PROGRESS` block gains the `Records:` line, so `/myflow-status`
  renders it too rather than only the command that produced it. **`myflow-fast/SKILL.md`**'s own
  `IN_PROGRESS`-with-no-artifact block — the one shape no other skill prints — gains it alongside.
- **Guard-presence lists** in `myflow-do`, `myflow-finish` and `myflow-fast` drop the retired script.
- **`setup.sh`** and **`check-guard-symlinks.sh`** drop its symlink.
- **`check-contract-budget.sh`** — every contract file edited above needs its budget row revisited;
  raising a budget for a genuine addition is the correct response, narrowing the guard is not.

The `finish.preserve-sessions` **stage key is unchanged**. Renaming it would invalidate every
recorded stage run carrying it and force a matching edit to `README.md`'s Level 1 table, which
`stats/internal/stages/names_test.go` parses — cost with no benefit, since the stage still does the
same job at the same point.

## Decisions

### The four records are two tables, not four

**ID:** `dispatches-is-one-table`
**Status:** active
**Chosen:** One `dispatches` table carries the SDD ledger, task → commit → model, and per-dispatch
cost — they are three views of one row, not three records. `findings` is the second table because a
finding has a genuinely different lifecycle (it is raised by a dispatch, then updated by later
rounds).
**Considered:** Four tables, one per subsumed ticket — rejected: three of them would share the same
primary key and be joined on every read. One generic `run_records` table with a `kind` column and a
JSONB payload — rejected: severity, status, model and role are exactly the columns the ticket's own
search cases filter on, and this buries them in JSON.

### Scope is the whole umbrella, in one change

**ID:** `full-umbrella-scope`
**Status:** active
**Chosen:** All four record types land together — the operator's call at the design gate.
**Considered:** Foundation plus the SDD ledger alone, with KAN-218/212/237 following as small
changes — rejected by the operator. Foundation plus ledger and findings — same.

### The Markdown record is generated output; the agent stops writing it

**ID:** `store-is-the-only-writer`
**Status:** active
**Chosen:** The skill instructs the agent to write records through the CLI only. Every Markdown
record becomes a render of the store.
**Considered:** Writing both, with the file as the render target — rejected: two sources that can
disagree is the problem this change exists to remove, and keeping the hand-write keeps it.

### `tasks.md`'s checkbox stays canonical for completion

**ID:** `completion-stays-in-the-file`
**Status:** active
**Chosen:** Store rows are a record, never a gate. `check-unfinished-work.sh` and `pipeline.md`'s
single-source rule are untouched.
**Considered:** Making the store canonical for completion — rejected: it puts a hard gate behind a
daemon that is explicitly allowed to be unreachable, and the guard would then have to distinguish "no
rows" from "cannot reach the store" on the one path where guessing wrong is unrecoverable.

### The store starts empty

**ID:** `no-backfill`
**Status:** active
**Chosen:** No import of the 34 archived ledgers or the preserved reviews. They stay as committed
Markdown.
**Considered:** A best-effort `myflow record import` — rejected: the archived format was never
stable (two ledgers carry hand-typed filenames the existing search pattern cannot match), so the
parser would be written against drift. A required full backfill — rejected for the same reason,
plus it makes an unparseable historical file block a change about new runs.

### The panel record renders at panel close

**ID:** `render-panel-at-close`
**Status:** active
**Chosen:** Rendering fires when the panel closes, so `check-unfinished-work.sh` reads the same
marker block it reads today.
**Considered:** Rendering only at finish run 1 — rejected: the guard runs during `/myflow-do`, long
before run 1, so it would have nothing to read.

### `preserve-session-records.sh` is retired outright

**ID:** `retire-preservation-script`
**Status:** active
**Chosen:** The skills call `myflow record render` directly; the script and its test harness are
deleted, and `pipeline.md`'s outcome table is replaced.
**Considered:** Keeping it as a thin wrapper that invokes the render and reports the same four words
— rejected by the operator: it preserves an indirection that no longer indirects anything. Rendering
into the worktree and letting the script keep copying — rejected: that keeps the file-path
indirection this change is about.

### The SPA is untouched

**ID:** `no-spa-in-this-change`
**Status:** active
**Chosen:** API and CLI only. `GET /api/v1/records/{project}/{change}` ships; nothing in
`stats/web/` changes.
**Considered:** Extending `RunDetail.tsx` with dispatch and finding panels — rejected by the
operator. A cross-change findings view — same.

### Cost comes from the harvester, intent from the dispatch

**ID:** `harvester-owns-cost`
**Status:** active
**Chosen:** The parent writes task, role, model and slot at dispatch close; `internal/harvest`
attributes token usage against that row afterwards, in a second pass over the same transcript
records.
**Considered:** The dispatching agent reporting tokens explicitly — rejected: Cursor and Codex write
no transcript, so the figure would be absent or guessed on two of three harnesses. Deriving the model
from the transcript too — rejected: `pipeline.md`'s Model policy requires recorded intent and forbids
a plausible-looking guess.

## Open questions

*(none — every question raised at the design gate was answered)*

## Testing

- **Migration** — `0010` applies twice idempotently, against `postgres:18-alpine`, alongside the
  existing migration tests.
- **Store** — `RecordDispatch` allocates `seq` under concurrency without overwriting (the same
  concurrent-insert test shape `BeginStage` already has); `UpsertFinding` updates status in place;
  `MergeDispatchMetrics` deep-merges rather than replacing.
- **Fallback** — a record write with the daemon down journals to `<name>.journal.record` and exits 0;
  `replayRecordFile` replays it.
- **Harvest** — a sidechain record inside a dispatch window lands in `dispatches.metrics`, and
  `stage_runs.metrics` for the same session is byte-identical to what the existing pass produces.
- **Render** — a change with rows renders both kinds; a change with none prints `MISSING:` and exits
  0; a change name carrying `/` or a glob metacharacter is refused; a destination outside the repo
  root is refused.
- **Guard** — `check-unfinished-work.sh` passes unchanged against a rendered panel record, proving
  the marker block survived the move.
- **Removal** — `check-guard-symlinks.sh` passes with the retired script gone, and no skill still
  names it (`check-references.sh`).
