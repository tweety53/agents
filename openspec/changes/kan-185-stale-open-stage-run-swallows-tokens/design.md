# Design — a stale open stage run must not swallow a session's telemetry

The brainstormed design this file is adapted from lives at
`docs/superpowers/specs/2026-08-16-kan-185-stale-open-stage-run-swallows-tokens-design.md`, which
carries the full incident walk-through. This file states the technical decisions.

## Context

A stage window is `[started_at, ended_at)`, with `ended_at` nil meaning still open
(`harvest.Window`, `stats/internal/harvest/attribute.go`). An open window therefore contains every
timestamp after its start, forever.

`/myflow-fast` run 2 on `kan-175-more-ui-ux-fixes` marked `finish.verify-merge` begin at 10:11:01Z
(stage run 146) and never marked it end — the fallback journal holds nothing, so the mark was never
issued rather than lost in transit. From that instant every assistant message in that session fell
inside two windows: the orphan's and whichever stage was actually running.

`bestWindow` resolves that overlap by preferring the highest `Attempt`. Both windows sit at attempt
1, so the tie-break has nothing to say and the winner falls out of iteration order —
`storeWindowSource.WindowsForSession` (`stats/cmd/myflowd/main.go`) passes no sort key, and
`Query.buildOrderSQL` appends the primary key as the total-order tiebreaker, so windows arrive
oldest-first. The orphan won every message for two hours: 187.8M cache-read, 227k output, 105k
thinking tokens, while nine stage runs of `kan-184-harden-the-release-and-deploy-path` recorded
nothing and its dashboard read `MEASURED 0`.

Constraints this design works inside:

- **A mark never blocks, delays or alters the stage it marks** (`skills/myflow-contracts/pipeline.md`).
  Whatever a begin mark does extra must stay inside the write it already makes.
- **`session_id` is bound after the fact** (KAN-172): the harvester resolves a session token to a
  session id once it finds the token in a transcript, so at begin time `session_id` is routinely
  NULL and `session_token` is the only identity a mark carries.
- **Absence is never zero** — no view may be handed a fabricated figure to paper over a
  misattribution.

## Goals / Non-Goals

**Goals**

- A dropped end mark costs its own stage's measurement and nothing further.
- Two windows of one session are never open simultaneously in the ordinary case, so nothing has to
  arbitrate between them.
- The orphan stops reading as "still running" on the dashboard for hours after it ended.

**Non-Goals**

- Repairing metrics already committed to stage run 146. The bytes are consumed, the offsets
  advanced, and the batch boundaries that would say which tokens belong to which change are gone.
- Detecting or reporting the dropped mark itself. Marks are best-effort by design.
- Changing the abandoned-stage sweeper, which addresses a different failure correctly.

## Decisions

### Attribution prefers the window that started last

**ID:** `latest-start-wins`
**Status:** active
**Chosen:** Among the windows containing a message, prefer the latest `StartedAt`; `Attempt` breaks
a remaining tie between windows starting at the same instant — of two windows that both contain a
message, the one that started later is the one the session is actually in.
**Considered:**
- *Keep `Attempt` as the only tie-break* — it is silent on the case that caused this, two windows at
  attempt 1, which is precisely the ordinary shape of a dropped end mark.
- *Refuse to attribute an ambiguous message at all* — turns a misattribution into a permanent
  under-count, and the absence-is-never-zero rule then reports honest emptiness for work that was
  really done and really measurable.
- *Order `WindowsForSession` by `started_at DESC` and keep taking the first match* — fixes this one
  caller by convention rather than stating the rule where attribution is decided; the next
  `WindowSource` implementation reintroduces the bug.

`Attempt` is not displaced by this, and the replay case is precisely why. A journalled begin carries
the **original** mark's start instant, not a fresh one: `stats/cmd/myflow/stage.go` captures
`StartedAt` once, before the RPC attempt, journals that same request on failure, and
`stats/internal/reconcile/reconcile.go` replays it as `StartedAt: req.StartedAt`. A replayed attempt
therefore starts at the *same* instant as the one it replays, never after it, so start time cannot
decide between the two at all.

**And yet the tie-break is still not reached on that path** — the correction has two halves, and the
first round of it stopped at the first. `begin-supersedes-open-runs` closes the replayed attempt's
predecessor with `ended_at` equal to the replay's own `started_at`, which makes that window the
empty interval `[T, T)`, and `Window.contains` rejects an empty interval. So the ordinary replay
never presents two containing windows at all. `Attempt` survives as a **defensive** tiebreaker, for
the windows the store no longer produces on that path: a run recorded before this change, a run
carrying no session token, and the moment before a supersede commits.

*(Corrected twice — in fix round 2, from the panel's reading of the replay path, and again in fix
round 4, when the panel pointed out that the round-2 correction was right about the timestamps and
wrong about the consequence. The decision itself never changed; both corrections were to the
sentence justifying it. This paragraph is the third statement of it and the first that matches the
code.)*

### A begin mark closes what it supersedes, in its own transaction

**ID:** `begin-supersedes-open-runs`
**Status:** active
**Chosen:** `BeginStage` inserts the new run and closes every still-open run sharing its
`session_token` that started no later than it, in one transaction. A session's marks are sequential,
so a begin is proof that whatever that session had open is over.
**Considered:**
- *Attribution fix alone* — leaves the orphan row open in every view until the sweeper reaches it
  six hours later, and leaves the store emitting overlapping open windows for anything else that
  reads them.
- *A separate call after the insert* — two writes with a gap between them, in a path whose whole
  guarantee is that a mark never blocks; a failure between them leaves the session with two open
  runs, which is the state being eliminated.
- *Shortening `sweepSilenceTimeout`* — the sweeper measures silence and never fires on a session
  that keeps working, so no timeout short of absurd would have helped here.

### Matched by session token, never by session id

**ID:** `supersede-keyed-on-token`
**Status:** active
**Chosen:** Close open runs matching the new run's `session_token`. It is always present — a `stage
begin` without one is rejected before it reaches the store — and it means exactly "the runs of this
session". A NULL token matches no row, so a run recorded without one is never touched.
**Considered:** *Matching on `session_id`* — routinely NULL at begin time, so the rule would silently
do nothing on exactly the marks that need it.

### The end instant is the new run's start, not `now()`

**ID:** `supersede-ends-at-successor-start`
**Status:** active
**Chosen:** `ended_at` is the superseding run's `started_at`. It is the truthful boundary — the
superseded stage ended no later than the moment the next one began — and it leaves the windows
non-overlapping, so attribution is correct from the store's shape alone, independent of
`latest-start-wins`.
**Considered:** *`now()`, as `SweepAbandoned` uses* — correct for the sweeper, which knows only that
the run is stale as of now; here the boundary is known exactly, and `now()` would leave the window
overlapping every stage that ran in between.

### Guarded on `started_at <=` the new run

**ID:** `supersede-guarded-on-start-order`
**Status:** active
**Chosen:** Only close runs whose `started_at` is at or before the new run's. A journalled begin
replayed later carries its original, older start instant; without the guard it would close runs that
legitimately started after it.
**Considered:** *Closing every open run of the token unconditionally* — a replay would then close the
live stage the session is actually in, which is the same class of damage this change exists to stop.

### `superseded` is a new outcome, distinct from `abandoned`

**ID:** `superseded-outcome-distinct`
**Status:** active
**Chosen:** Record `superseded`. `abandoned` means the daemon closed a run because its session went
silent, and the rework-rate view reads that outcome directly
(`stats/internal/store/aggregate.go`); a superseded run is neither rework nor silence.
**Considered:** *Reusing `abandoned`* — puts every dropped end mark into the rework rate, corrupting
a statistic to avoid adding a word. *Leaving `outcome` NULL* — indistinguishable from a run still
open in every query that reads the outcome.

### Begins for one session serialise on an advisory lock

**ID:** `supersede-serialised-per-session`
**Status:** active
**Chosen:** `insertStageRunAndSupersede` takes `pg_advisory_xact_lock(hashtext(<session token>))` as
its transaction's first statement, and only when the token is non-empty. Atomicity is not
serialisability: under READ COMMITTED neither of two concurrent begins for one session sees the
other's uncommitted insert, so neither supersedes it and both stay open — the state this change
exists to prevent, reachable through the reconciler replaying a journalled begin beside a live one.
The lock is transaction-scoped, so it releases on commit or rollback with no unlock path to forget,
and the store already uses a Postgres advisory lock for its migration runner.
**Considered:**
- *SERIALIZABLE with a retry* — `BeginStage` already retries on the attempt unique-violation, so the
  machinery exists; but it converts a routine, expected interleaving into a serialisation failure on
  a path required never to block or delay the mark, and it widens the isolation level of a
  transaction that is otherwise a plain insert.
- *A unique index enforcing one open run per token* — refuses the second begin outright rather than
  superseding the first, which turns an ordinary sequence of marks into an error.
- *Leaving it* — the panel measured the race as reachable; the bounded blast radius (a lingering
  open row, attribution unaffected because `bestWindow` orders on start time rather than on
  `outcome`) is an argument for its severity, not for its absence.

### A late end mark cannot reopen what a supersede closed

**ID:** `end-mark-cannot-resurrect`
**Status:** active
**Chosen:** `EndStage`'s UPDATE gains `AND ended_at IS NULL`, and a row that exists but is already
closed returns `ErrStageRunAlreadyClosed`, which `ApplyEndStageMark` translates into the
`ErrNoOpenStageRun` it already treats as a definitive, loudly-logged refusal. `ApplyEndStageMark`
resolves an open run and closes it in two separate calls; a supersede landing in that gap was
silently overwritten, putting the session back to two overlapping windows.
**Considered:**
- *Making the lookup and the close one transaction* — larger, and it would still need the guard to
  say what happens when the row is already closed.
- *Guarding `MergeMetrics` the same way* — deliberately not done. Its write is additive telemetry
  about work that really happened, so landing it on a since-closed run records the truth, where
  refusing it would discard measured usage. Only the fields that say whether a window is open need
  the guard.

## Risks / Trade-offs

- **A legitimately concurrent open run in one session would now be closed early** → marks are
  sequential within a session by construction, and the `started_at <=` guard keeps a replay from
  reaching a live run. Subagents write to their own transcripts and mark nothing.
- **A late end mark for a run this rule already closed now finds nothing open** → `ApplyEndStageMark`
  returns `ErrNoOpenStageRun`, which is already a definitive, loudly-logged outcome rather than a
  retry loop — the same path a duplicate replay or a swept run already takes.
- **`outcome` has no CHECK constraint, so a new value cannot be rejected by the schema** → the two
  readers of the column are the rework-rate filter (`outcome = 'abandoned'`, unaffected) and the
  SPA, which renders whatever string it is given.

## Migration Plan

One migration, `0009_stage_run_open_session_token.sql`, adding a partial index on
`stage_runs (session_token) WHERE ended_at IS NULL` — the supersede UPDATE's own predicate. It was
not in the original design: the review panel measured that UPDATE reaching its rows through
`stage_runs_ended_at` and then filtering every open run, and the operator chose to index it rather
than withdraw the finding. The existing `stage_runs_unresolved_session_token` cannot serve it,
being partial on `session_id IS NULL` — a predicate the supersede does not carry.

Nothing else needs a schema change: `session_token` exists (`0008_stage_run_session_token.sql`) and
`outcome` is free text with no CHECK constraint. No column, constraint or row is altered, so the
migration is additive and safe to apply while the daemon is running.

Deployment is a daemon rebuild and restart, which applies the migration on startup; the CLI is
unchanged. Rollback is reverting the binary — the index is harmless if left in place, and rows
already closed as `superseded` stay closed, which is a correct record either way.

**"Safe to apply while running" means the writes it blocks are brief, not that it blocks none.**
`applyMigration` (`stats/internal/store/migrations.go`) wraps every migration in one transaction,
and a plain `CREATE INDEX` inside a transaction holds a write-blocking lock on `stage_runs` until
that transaction commits. `CREATE INDEX CONCURRENTLY` is not an option: Postgres refuses it inside a
transaction block, so the runner's own design forecloses it. The lock is therefore held for the
length of the index build — proportional to the table, milliseconds at this tool's scale, where the
live store holds a few hundred stage runs — and a marking session that lands inside that window
waits rather than fails. The panel raised this against the unqualified claim the section used to
make; the mechanism is unchanged, the sentence is now honest about it.

## Open Questions

*(none — the two decisions the design turned on were put to the operator and answered: both parts
rather than attribution alone, and no backfill of the misattributed history.)*
