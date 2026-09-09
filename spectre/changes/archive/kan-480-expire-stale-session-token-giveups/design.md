# design — kan-480-expire-stale-session-token-giveups

## Context

- `session_token_giveups` rows are read once at Watcher start (`seedPersistedGiveUps` →
  `store.PersistedGiveUps`) into `retriedTokens`; `scanRetriedTokens` then runs at the end of
  every `RunOnce` while that set is non-empty, calling each source's `ReadAllCmds` on every
  discovered file — the whole corpus, per cycle.
- Measured corpus here: 7,049 transcript files; one full retry pass ≈ 15 minutes at ~95% CPU.
<!-- measured: flowd PID 48186 sampled twice in scanRetriedTokens 2026-09-10 00:24-00:38 local,
     ~95% CPU from process start to pass completion; 7,049 = count(*) from harvest_offsets -->
- Horizon of 24h: attribution lands only inside stage windows, which close within
  minutes-to-hours (the sweeper closes abandoned runs); a day-old give-up cannot bind anything
  that still attributes, so 24h is generous by an order of magnitude and needs no tuning.
<!-- predicted: 24h is a design constant, not a measurement; the closure claim it rests on is
     measured: 2026-09-09 stage runs' ended_at-started_at spans 18s to 3h53m (n=4,712), and the
     54 cleared give-ups aged 1-18 days all named tokens of archived changes -->

## Decisions

### Age-based expiry at the store's load read

**ID:** giveup-age-expiry-at-load
**Status:** active
**Chosen:** `PersistedGiveUps` deletes rows with `gave_up_at` older than `giveUpRetryHorizon`
(24h, computed in Go, bound as a parameter) and returns only rows inside the horizon — one
function changed, the table stays bounded, and the watcher needs no change at all.
**Considered:** a retry cap on the `retries` column — a cap still leaves dead rows in the table
forever and only bites after `retries` full wasted windows; ruled out. Filtering in
`internal/harvest` at scan time — threads time through the wrong layer and still re-reads rows
into memory; ruled out. Sweep-based expiry (a second background pass) — a whole subsystem for
one predicate; ruled out.

### Expiry deletes, it does not filter only

**ID:** expired-rows-deleted
**Status:** active
**Chosen:** expired rows are deleted, not merely filtered from the result — the measured defect
included unbounded table growth (54 rows and rising), and a filter-only read would re-decide the
same rows on every start forever.
**Considered:** filter-only, keep rows for forensics — the rows carry no forensics value beyond
the token string and a timestamp; ruled out.

## Open questions

_none_
