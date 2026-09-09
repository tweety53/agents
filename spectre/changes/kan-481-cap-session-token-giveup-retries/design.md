# design — kan-481-cap-session-token-giveup-retries

## Context

- kan-480's horizon expires give-ups by age; the upsert in `RecordSessionTokenGiveUp` refreshes
  `gave_up_at` on every re-give-up, so rows re-recorded within a daemon's lifetime never reach
  it. Measured 2026-09-10: 44 rows re-recorded minutes before the kan-480 binary deployed put
  the first post-deploy start back into the full corpus scan.

## Decisions

### Retry cap at three failed generations, enforced in the same load read

**ID:** giveup-retry-cap-at-load
**Status:** active
**Chosen:** `PersistedGiveUps` deletes rows with `retries >= giveUpMaxRetries` (3) alongside
kan-480's age expiry, and returns only rows failing neither bound. Three full bounded windows
(60 cycles each) without a bind is measured-hopeless, not bad luck.
**Considered:** raising the horizon instead — does not stop freshly-refreshed rows; ruled out.
Never re-seeding, keep rows forever — unbounded table; ruled out.

## Open questions

_none_

<!-- approved: the operator's 2026-09-10 "fix it" instruction named both mechanisms — "bound the
     retry scan or expire old give-ups" — and pre-approved landing without a further gate -->
