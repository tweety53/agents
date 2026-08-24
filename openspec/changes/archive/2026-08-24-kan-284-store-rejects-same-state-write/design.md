# Design — kan-284-store-rejects-same-state-write

The diagnosis this design rests on — why a same-state write is refused, why the ticket's stated
cause is not the real one, and the live evidence for both — is recorded once in
`docs/superpowers/specs/2026-08-24-kan-284-store-rejects-same-state-write-design.md` and is not
restated here.

## Where the timestamp is stamped

`runStateSet` (`stats/cmd/myflow/state.go`) gains one step between the `isJSONObject` check and the
project-key resolution: rewrite the decoded body's `updatedAt` to
`time.Now().UTC().Format(time.RFC3339Nano)`, unconditionally.

The rewrite lands on `body` — the variable the fallback path writes to the on-disk state file and
appends to the journal, and the variable `putChange` wraps with `mainCheckoutPath`. One instant
therefore reaches all three destinations, and a journal entry replayed hours later carries the
instant its write actually happened at rather than the instant of the replay.

It is a sibling of the existing `withMainCheckoutPath`: same shape, same position in the flow, same
reason — a value the CLI owns that no caller should have to supply.

`time.RFC3339Nano` is chosen because `toDTO` already emits the field in exactly that format, so the
value a `state get` prints and the value a `state set` sends are the same shape.

## The monotonic rule is not touched

`changes.go`'s `WHERE state_rank(...) > ... OR (... AND EXCLUDED.updated_at > changes.updated_at)`
is left exactly as it is. It was never wrong; it was being fed values from two clocks at two
precisions. With one writer at one precision, the strict `>` is satisfied by every live write,
including one that follows a synthetic bootstrap inside the same second.

This is why the existing same-state and duplicate-retry tests must stay green **unchanged** — they
are the regression net for the rule itself, and editing them to accommodate this change would remove
the only evidence that the rule still holds.

**Residual:** two writes reading an identical nanosecond from the clock would still be refused. On
darwin `time.Now()` is nanosecond-resolution, so this is unreachable in practice, and it is not
worth relaxing `>` to `>=`, which would re-admit the stale overwrite and the duplicate replay the
rule exists to stop.

## Where a merge base is refused

Two enforcement points, deliberately not one.

**CLI**, in `runStateSet`, before the store is touched — the same position as the existing
`isJSONObject` check. On a violation: one stderr line naming the offending worktree path and the
rejected value, exit 2, nothing written locally and nothing journalled.

Failing loudly here rather than taking the never-block fallback is the whole point. The fallback
exists for a store outage; a malformed value is a caller mistake, and journalling it would hide the
bad value until `check-finish-preflight.sh` refuses at the finish gate — which is the failure being
fixed, not a tolerable version of it.

**Store**, in `PutChange`, as `ErrInvalidMergeBase` alongside the existing `ErrInvalidState`,
checked against `c.Repos` after `reposFromWorktrees` has derived them. For a live write this is
unreachable, because the CLI refused first. It covers the one path that bypasses the CLI: `reconcile`
replaying a hand-edited fallback file, which the state file contract already names as a real case.

`IsDefinitiveChangeOutcome` had to be **extended** to reach the new sentinel. Its `switch` is an
explicit allowlist, so an unrecognised error falls to `default: return false` — a journal entry
carrying a bad merge base would never retire, and `replayFile` would stop on it and block every
valid entry behind it. This design originally asserted the classification already held; task 3
disproved that with a RED test (`{Journals:1 Applied:0 Refused:0}`) before adding the one `case`.
The assumption was load-bearing, and the correction is recorded here rather than quietly absorbed.

**No new CLI reporting path is needed, and the 400-versus-409 handling under "The pipeline never
blocks" is unchanged** — the CLI validates first, so it never sees this error on either path.

## Contract changes

`skills/myflow-contracts/state-file.md`:

- The `updatedAt` bullet becomes CLI-owned: stamped by `myflow state set` at full precision on every
  write; skills neither read the clock for it nor emit it. The `date -u +%Y-%m-%dT%H:%M:%SZ`
  instruction and the "never invent or placeholder it" warning both go — a field a skill cannot
  write cannot be fabricated. The field stays accepted-and-ignored on the wire.
- The `worktrees` bullet gains the value constraint and names where it is refused. The existing
  paragraph on a `null` value is untouched.
- **Writes are monotonic in both dimensions** keeps its rule verbatim and gains one sentence
  recording that the instant it orders by now comes from a single writer.

`state-file.md` is the only file under `skills/` that mentions reading the clock for this field, so
no other skill file changes.

## Decisions

### Where `updatedAt` comes from

**ID:** updated-at-stamped-by-cli
**Status:** active
**Chosen:** The CLI stamps it — one clock, one value shared by the store row, the fallback file and
the journal entry, and a hand-written or fabricated timestamp becomes impossible rather than merely
forbidden.
**Considered:** *The daemon stamps it* — also closes the class, but one write then carries two
different instants, the store's and the fallback file's, for no gain: the daemon is loopback-only,
so client and server share a clock and there is no skew to remove. *Align precision only*, by
truncating the synthetic bootstrap to the second — narrows the window without closing it, since two
same-state writes inside one second are still refused and a fabricated timestamp stays possible.
*Accept `>=` at the same state* — re-admits the stale overwrite and the duplicate replay the
monotonic rule exists to stop.

### Where an invalid merge base is refused

**ID:** merge-base-refused-at-cli-and-store
**Status:** active
**Chosen:** Both the CLI and the store. The CLI covers every write the pipeline makes and fails
loudly at the point of the mistake; the store covers journal replay of a hand-edited fallback file,
which bypasses the CLI entirely.
**Considered:** *CLI only* — smaller, and covers every write the pipeline actually makes, but leaves
a hand-edited fallback file able to reach the store through replay. *Store only* — would require
also teaching the CLI to report the refusal rather than treat it as an outage, because a daemon-side
400 is swallowed by the never-block fallback path; strictly more work for strictly less coverage.

### How far the contract change reaches

**ID:** updated-at-is-cli-owned-in-contract
**Status:** active
**Chosen:** `state-file.md` declares the field CLI-owned and skills stop writing it, while the field
stays accepted-and-ignored on the wire so a journal entry already on disk still replays.
**Considered:** *Keep it skill-written and let the CLI silently overwrite* — least churn, but the
contract would keep instructing skills to read the clock for a value that is then discarded, a
documented step that does nothing. *Reject a body still carrying the field* — strictest, but breaks
every journal entry already on disk and any body written before this change.

### Whether the new sentinel gets an HTTP status mapping

**ID:** merge-base-error-maps-to-400
**Status:** active
**Chosen:** Map `ErrInvalidMergeBase` to 400 in `mapStoreError`, alongside its three caller-content
siblings `ErrInvalidState`, `ErrInvalidMainCheckoutPath` and `ErrDuplicateRepoRoot`.
**Considered:** *Leave it unmapped*, on the reading that this design's "the 400-versus-409 handling
is unchanged" forbade touching the mapping. That sentence is about the CLI's fallback behaviour, and
mapping to 400 leaves that behaviour byte-for-byte identical, because the CLI validates before it
sends and never reaches the store's refusal on either path. What leaving it unmapped costs is real:
an out-of-band HTTP PUT carrying a bad merge base gets a 500 asserting the daemon broke when the
caller's content was wrong, and writes a `logger.Error` line per client mistake.

### Whether the sha pattern is shared between the two enforcement points

**ID:** merge-base-pattern-not-shared
**Status:** active
**Chosen:** Write `^[0-9a-f]{40}$` again in `internal/store` rather than exporting one constant.
**Considered:** *Share it*, following the `store.ProjectKeySuffixPattern` precedent. Rejected
because that constant is exported for a different reason — another package interpolates it into
SQL — while here `cmd/myflow` does not import `internal/store` at all, and importing it for one
regular expression would pull the store package and pgx into the CLI's dependency graph. The two
enforcement points are deliberately independent defences, which is the same reason they are two.

### What this change does not repair

**ID:** no-migration-of-existing-rows
**Status:** active
**Chosen:** No migration. The stray `.claude-485b1ff8` project row, the rows parked at synthetic
`STARTED`, and any backfill of `change_repos` are left alone.
**Considered:** *Include the stray project row* — a genuine defect, but a different one: it is a
project-key resolution fault, not a write-ordering or value-shape fault, and bundling it would put
two unrelated root causes behind one review.

## Open questions

*(none)*
