# kan-480-expire-stale-session-token-giveups

## Why

- Measured (2026-09-10, this machine): with 54 stale `session_token_giveups` rows (ages 1–18
  days, tokens from long-archived changes, retries up to 10), every flowd start re-seeded all of
  them for retry and `Watcher.scanRetriedTokens` read the entire 7,049-file transcript corpus
  whole once per 5-second harvest cycle — 15+ minutes of ~95% CPU per pass, stalling
  session-token binding and token attribution the whole time.
- The retries are hopeless by construction: a give-up only helps re-find a session mark whose
  usage can still attribute inside a stage run's `[started_at, ended_at)` window; those windows
  close within minutes-to-hours. A give-up older than a day can never bind anything useful.

## What changes

- `internal/store`: `PersistedGiveUps` expires give-ups past a 24-hour retry horizon — rows
  older than the horizon are deleted (bounded hygiene write; the table stops growing forever)
  and only rows inside the horizon are returned.
- With no give-ups inside the horizon, `scanRetriedTokens` returns immediately and a restart
  reaches binding and attribution in seconds.
