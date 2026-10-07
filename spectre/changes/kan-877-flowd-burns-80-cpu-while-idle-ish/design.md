## Context

`stats/internal/harvest/watcher.go`'s `RunOnce` runs every `harvestInterval` (5 s, `cmd/flowd/main.go`).
Measured on the dev machine on 2026-10-07, before any edit:

- The transcript roots hold 2,643 Claude files (~4 GB) plus the ZCode rollout files.
  <!-- measured: find ~/.claude/projects -name '*.jsonl' | wc -l; du -sh ~/.claude/projects — machine-local, not re-runnable at a ref -->
- One `scanRetriedTokens` pass over them with the 104 pending give-up tokens took 23.6 s to read and
  parse 117,003 commands, and 66.3 s to match them — ~90 s CPU per cycle.
  <!-- measured: throwaway in-package test timing discoverTranscripts + ReadAllCommands + matchSessionTokens against the live token set, 2026-10-07 — machine-local, not re-runnable at a ref -->
- The previous flowd process (started 2026-10-05 12:14) logged retried-scan read failures from 12:19 to
  2026-10-06 22:57 and no give-up in that whole lifetime: the scan stayed live for ~35 h.
  <!-- measured: grep 'retried session token' /private/tmp/flow-live.log, lines 328602-328811 — machine-local log -->
- None of the 104 tokens matched anything: they are `ff-*` zcode runs whose rollout files were pruned.
  <!-- measured: docker exec flow-postgres psql … stage_runs where session_id is null — live dev store -->
- With no retried token pending, a cycle costs ~0.17 s CPU — one `GetHarvestOffset` query and one
  open+read per transcript file.
  <!-- measured: ps -o time= -p <flowd pid> sampled every 2 s — live process -->

It is one change because all three costs live in `RunOnce`'s per-cycle loop over the same file list, in one
file, and share one test file.

## Decisions

### Scan already-read bytes once per process

**ID:** retried-scan-once-per-process
**Status:** active
**Chosen:** `scanRetriedTokens` runs one full pass per Watcher lifetime and then stops — bytes behind a
committed offset never change, and every byte past it (and every new file) already reaches
`matchSessionTokens` through the ordinary per-file path. A pass whose only read failures are vanished files
(`fs.ErrNotExist`) counts as done; any other read failure leaves the pass to run again next cycle. The
per-token `maxSessionTokenResolutionCycles` give-up bound is untouched.
**Considered:** keep the per-cycle scan and only make matching cheaper — still ~24 s of read+parse per
cycle, ~60 cycles per restart; scan once per token rather than per process — same result, since every
retried token is seeded at the same moment, with per-token bookkeeping nobody needs.

### Match commands against a token set

**ID:** match-by-token-set
**Status:** active
**Chosen:** `matchSessionTokens` builds a set of the pending tokens once, tests each command against
`stageMarkInvocationPattern` once, extracts the command's `-session-token`/`--session-token` value(s) with
the same field rules `isSessionMarkCommand` applies, and looks each up in the set. The mention-versus-
invocation rule is unchanged.
**Considered:** leave the O(commands × tokens) loop — it is the 66 s of the 90 s; precompile one regex per
token — still per-token work per command.

### Skip a transcript whose size equals its last-seen committed offset

**ID:** skip-unchanged-by-size
**Status:** active
**Chosen:** the Watcher remembers, per path, the committed offset it last read or committed. When a file's
current size equals that offset, `RunOnce` skips `GetHarvestOffset` and `ReadNew` for it and still runs
`maybeBackfillDispatchMeta`. Any other size — growth, truncation, a path not yet seen — takes the ordinary
path unchanged. A withheld batch (one that revealed a pending token) records nothing, so it is re-read
exactly as today.
**Considered:** compare mtime as well — the ordinary path already treats `offset == size` as "nothing new"
without one, so an mtime adds no correctness; filesystem notifications (fsnotify) — a new dependency and a
new failure mode for a ~3 % idle cost; leave it — the operator asked for it in scope (2026-10-07).

## Open questions

## Live check

Needs flowd restarted on the new build, an operator action — no agent stops the dev flowd (`CLAUDE.md`).

1. Before restarting, confirm persisted give-ups exist:
   `docker exec flow-postgres psql -U flow -d flow -Atc "select count(*) from session_token_giveups where retries < 3 and gave_up_at > now() - interval '24 hours'"` — non-zero.
2. Restart flowd on the new build; record `ps -o time= -p <pid>` at +10 s, +60 s, +120 s, +300 s.
3. Fixed: the CPU time grows by roughly one read+parse pass in the first minute, then by well under 0.2 s per
   5 s cycle; `/private/tmp/flow-live.log` shows the retried tokens giving up again ~5 min after start.
4. Not fixed: CPU time keeps growing by ~5 s per 5 s past the first two minutes, and no give-up appears in
   the log within 10 minutes — the pre-fix shape.
   <!-- predicted: the operator's ps samples after restarting flowd on the new build -->
5. With no give-ups pending, idle growth drops below the per-cycle baseline in `## Context`.
