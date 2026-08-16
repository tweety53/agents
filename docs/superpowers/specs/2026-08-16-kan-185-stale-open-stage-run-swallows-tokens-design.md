# KAN-185 — one missing end mark, a whole session's telemetry gone

**Date:** 2026-08-16
**Change:** `kan-185-stale-open-stage-run-swallows-tokens`

## The observation

The stats dashboard for `kan-184-harden-the-release-and-deploy-path` read `RUNS 9`, `MEASURED 0`,
total cost `unavailable`, total input tokens `unavailable`. Every one of its nine stage runs carried
a bound `session_id` and an empty metrics bag.

Nothing had failed. `harvest_offsets` for that session's transcript stood at 11,911,042 bytes —
byte-for-byte the file's current size, updated seconds earlier. The harvester was reading the
transcript, attributing it, and committing the result. It was committing it to the wrong stage run.

## The mechanism, end to end

1. A `/myflow-fast` run 2 on `kan-175-more-ui-ux-fixes` marked `finish.verify-merge` begin at
   10:11:01Z (stage run 146) and never marked it end. The next mark in that session was
   `finish.sync-archive` begin, eight seconds later. No entry sat in the fallback journal, so the
   end mark was never issued at all rather than lost on its way to the store.
2. Stage run 146 therefore stayed open — `ended_at IS NULL` — which is exactly how
   `harvest.Window` spells "still running". An open window contains **every** timestamp after its
   start (`Window.contains`, `stats/internal/harvest/attribute.go`).
3. Every later stage of that session opened its own window, each one also containing the messages
   being written at the time. So from 10:11 onward, every assistant message fell inside at least two
   windows: the orphan's and the live stage's.
4. `bestWindow` resolves that overlap by preferring the highest `Attempt`. Both windows are
   attempt 1 — the orphan is not a retry of anything — so the tie-break has nothing to say, and the
   winner falls out of iteration order. `storeWindowSource.WindowsForSession`
   (`stats/cmd/myflowd/main.go`) passes no sort key, and `Query.buildOrderSQL` appends the row's own
   primary key as the total-order tiebreaker, so the windows arrive `ORDER BY id ASC`: oldest first.
   `bestWindow` keeps the first match and only replaces it on a strictly higher attempt, so the
   orphan wins, every time, forever.
5. The result is silent and total. Run 146 accumulated 187.8M cache-read, 227k output and 105k
   thinking tokens, priced and attributed to a finish stage of a change that had been archived two
   hours earlier. Nine stage runs of a different change recorded nothing at all.
6. Every dashboard read exactly what the store held. The absence-is-never-zero rule did its job:
   kan-184's figures read `unavailable`, not `0`. The rule was working; the number it was honestly
   reporting the absence of had been deposited somewhere else.

## Why nothing caught it

The abandoned-stage sweeper is the mechanism that exists for stage runs that never end, and it is
not a backstop here on any useful timescale. `sweepSilenceTimeout` is 6 hours
(`stats/cmd/myflowd/main.go`), and `SweepAbandoned` closes runs whose `started_at` precedes that
cutoff — it measures silence. A session that keeps working keeps producing marks, and the orphan
sits inside the live window the whole time. The damage is done in the first minute; the sweeper
would arrive, at best, six hours later.

`Window`'s half-open interval does not help either. It exists to resolve a message landing exactly
on the boundary between two **adjacent, closed** windows, and its own doc comment says so. Two
windows that are open at the same instant are a different situation, which is why `Attempt` exists
as a separate tie-break — and `Attempt` addresses only the case it was written for, a replayed
begin opening a second attempt of the same stage.

## The design

Two changes, one in each layer that got this wrong. Either alone would have prevented the incident;
together they close it from both ends — the store stops producing overlapping open windows, and the
harvester stops mis-resolving them where they already exist.

### 1. `bestWindow` prefers the window that started last

`stats/internal/harvest/attribute.go`. Among the windows containing a message, prefer the one with
the latest `StartedAt`; `Attempt` breaks a remaining tie between windows that start at the same
instant.

This is the rule the tie-break should always have carried: of two windows that both contain a
message, the one that started later is the one the session is actually in. It subsumes the case
`Attempt` was written for rather than displacing it — a replayed begin opens attempt 2 *after*
attempt 1, so latest-start already returns the window highest-attempt returned. `Attempt` stays as
the tiebreaker for the one case latest-start cannot decide: two windows recorded at the identical
instant.

Messages written between an orphan's start and the next stage's start still attribute to the
orphan, which is correct — during those seconds the session really was in that stage, whatever the
missing mark says.

### 2. A begin mark closes the open runs it supersedes

`stats/internal/store/stageruns.go`. `BeginStage` inserts the new run and, in the **same
transaction**, closes every still-open stage run that shares the new run's `session_token` and
started no later than it does, setting `ended_at` to the new run's own `started_at` and `outcome`
to `superseded`.

A session's marks are sequential: a begin is proof that whatever that session had open is over,
whether or not its end mark was ever issued. Recording that proof where it is discovered is what
keeps two overlapping open windows from existing in the first place — and what stops the orphan
sitting in the dashboard as "still running" for six hours.

**Keyed on `session_token`, never `session_id`.** `session_id` is bound after the fact, by the
harvester, once it finds the token in a transcript (KAN-172); at begin time it is routinely NULL.
The token is always present — a `stage begin` without one is rejected before it reaches the store —
and it means exactly "the runs of this session", which is the set to close. A NULL token matches no
row, so a legacy run without one is never touched.

**`ended_at` is the new run's `started_at`, not `now()`.** It is the truthful boundary: the
superseded stage ended no later than the moment the next one began. It also leaves the windows
non-overlapping, so attribution is correct from the store's shape alone, without relying on part 1
— which is what makes the two parts independent rather than redundant.

**Guarded by `started_at <= <new run's started_at>`.** A journalled begin replayed later carries
its original, older `started_at`; without the guard it would close runs that legitimately started
after it. With it, a replay closes only what it really preceded.

**`superseded` is a new outcome, distinct from `abandoned`.** `abandoned` means the daemon closed a
run because its session went silent, and the rework-rate view reads that outcome directly
(`stats/internal/store/aggregate.go`). A superseded run is not rework and is not silence: it is a
run whose end mark never came, discovered by the next mark in the same session. Recording them as
the same outcome would put every dropped end mark into the rework rate.

## Decisions

- **Both parts, not one.** The operator chose the two-part fix over attribution-only. Part 1 alone
  leaves the orphan row open in the dashboard until the sweeper reaches it, and leaves the store
  producing overlapping windows for anything else that reads them. Part 2 alone leaves every stage
  run already open — including ones open right now — resolved by the old tie-break.
- **The 6-hour sweeper timeout is unchanged.** Shortening it was considered and rejected: the
  sweeper measures silence, so no timeout short of absurd would have fired on a session that was
  actively working. It addresses a different failure and addresses it correctly.
- **The misattributed history stays as it is.** The two hours already committed to run 146 cannot be
  moved without hand-written SQL, and the batch boundaries that would say which tokens belong to
  which change are gone — any split would be an estimate. The record keeps what was measured.

## Testing

- `stats/internal/harvest/attribute_test.go` — two open windows in one session, the stale one first
  in the returned order, a message timestamped inside both: it attributes to the later-started
  window. Plus the existing attempt tie-break case, which must keep passing unchanged.
- `stats/internal/store/stageruns_test.go` — a begin closes an earlier open run sharing its session
  token, with `ended_at` equal to the new run's `started_at` and outcome `superseded`; a begin whose
  `started_at` precedes an open run leaves that run open; a begin with no session token closes
  nothing.
