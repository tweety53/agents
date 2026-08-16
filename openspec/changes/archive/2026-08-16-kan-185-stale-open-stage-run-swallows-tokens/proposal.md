## Why

One `/myflow-fast` run dropped a single `stage end` mark, and every token that session spent for the
next two hours — across a different change entirely — was recorded against the abandoned stage
instead of the stage that spent them. `kan-184-harden-the-release-and-deploy-path` read `MEASURED 0`
with every figure `unavailable` on its own dashboard, while a finish stage of an already-archived
change accumulated 187.8M cache-read tokens. Nothing failed, nothing warned: a stage run left open
contains every timestamp after it, and the tie-break that resolves overlapping windows had nothing
to say about two windows at the same attempt.

Telemetry that a single dropped mark can silently redirect for a whole session is telemetry nobody
can trust a dashboard reading of.

## What Changes

- `bestWindow` (`stats/internal/harvest/attribute.go`) resolves a message that falls inside more
  than one window by preferring the window that **started last**. `Attempt` remains the tiebreaker
  for windows recorded at the same instant — the one case start time cannot decide.
- `BeginStage` (`stats/internal/store/stageruns.go`) closes what a new mark supersedes: in the same
  transaction as the insert, every still-open stage run sharing the new run's `session_token` and
  starting no later than it is closed with `ended_at` set to the new run's `started_at` and
  `outcome` set to `superseded`.
- `superseded` joins `abandoned` as a recorded outcome, and is deliberately distinct from it: the
  rework-rate view reads `abandoned` directly, and a dropped end mark is not rework.

Not changed: the 6-hour abandoned-stage sweeper, which measures silence and would never have fired
on a session that kept working; and the already-recorded metrics on the stage run that absorbed
them, which no code path can split back apart.

## Capabilities

### New Capabilities

*(none)*

### Modified Capabilities

- `myflow-run-telemetry`: attribution of a message that falls inside more than one open window is
  stated, rather than left to iteration order; and a begin mark closes the runs of its own session
  that it supersedes, so two windows of one session are never open at once.

## Impact

- `stats/internal/harvest/attribute.go` — `bestWindow`'s selection rule.
- `stats/internal/store/stageruns.go` — `BeginStage`/`insertStageRun` become one transaction that
  also closes superseded runs.
- `stats/internal/harvest/attribute_test.go`, `stats/internal/store/stageruns_test.go` — regression
  coverage for both.
- `stats/internal/store/migrations/0009_stage_run_open_session_token.sql` — one migration, adding a
  partial index on `session_token` restricted to open runs, so the supersede reaches its rows by
  index rather than by filtering every open run. Added after the review panel measured the cost;
  nothing else needs a schema change, since `outcome` is free text with no CHECK constraint and
  `session_token` already exists (`0008_stage_run_session_token.sql`).
- No API, CLI or SPA surface changes. A `superseded` run reads on every existing view exactly as any
  other closed run does.
