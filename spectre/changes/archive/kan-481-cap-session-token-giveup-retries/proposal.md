# kan-481-cap-session-token-giveup-retries

## Why

- Follow-up to kan-480, found by measuring its deployed behavior (2026-09-10): the 24-hour
  horizon expires rows that stop being re-recorded, but `RecordSessionTokenGiveUp`'s upsert
  refreshes `gave_up_at` on every re-give-up — tokens that keep failing stay forever-fresh and
  re-seed the whole-corpus retry scan on every daemon start. Live: the pre-fix binary
  re-recorded 44 rows minutes before the kan-480 binary deployed; the first post-deploy start
  went straight back to ~95% CPU in `scanRetriedTokens`.
- A token that failed three full bounded retry generations is dead by construction: `retries`
  increments on every re-give-up and never resets, and the cleared historical rows had reached
  10.

## What changes

- `internal/store`: `PersistedGiveUps` also drops rows with `retries >= giveUpMaxRetries` (3) —
  deleted like expired rows, never returned. Age and attempts bound the table and the retry scan
  together.
