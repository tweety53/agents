## Context

The brainstorming record — the problem, the scope and the rule — is
`docs/superpowers/specs/2026-08-24-kan-327-review-per-round-delta-after-round-1-design.md`. This
file carries the decisions, the open questions and the implementation mechanism.

## The per-slot delta

A delta-slot's input in a Full-mode re-run is

```bash unverified:the shape of the range; the sha is resolved at dispatch time and cannot be written literally here
git diff <the HEAD sha that slot last reviewed> HEAD
```

written to `<abs-worktree>/.superpowers/sdd/slot-delta-<round>-<slot>.diff`, beside the existing
`fix-round-N.diff` the Targeted row already uses.

**Anchored per slot, not per round.** In Targeted mode only slot 0 and the slots that raised
findings re-run, so when round 5 escalates to Full, Principles may not have read rounds 2 through 4.
`fix-round-5.diff` alone would leave that gap, and *targeting is a cost optimization, never a
coverage waiver* forbids it.

**Tree-to-tree, deliberately.** `git diff A B` compares two trees and needs no ancestry between
them, so a base recorded before a `git rebase --autosquash` stays valid after the rebase rewrote the
task commit. `A..B` and a merge-base-anchored range would not.

**Coverage holds by construction.** Pass 1 always runs the full roster against the full
`final-review.diff`, so every slot has a real starting sha and every change made since sits inside
some slot's delta.

**An unknown base means the full diff.** A slot whose last-reviewed sha the dispatcher does not
hold reads `final-review.diff`. An unknown base is not a small delta.

## Recording the base

The base is recorded on the panel slot's own dispatch row — the row keyed `panel-<round>-<slot>`
that already says which slot ran in which round.

| Layer | Change |
|-------|--------|
| `stats/internal/store/migrations/0014_dispatch_diff_base.sql` | `ALTER TABLE dispatches ADD COLUMN diff_base TEXT` — the same single-column shape `0011_dispatch_agent_id.sql` and `0012_dispatch_key.sql` use |
| `stats/internal/store/records.go` | `diff_base` in `dispatchColumns`, the scan, and `insertDispatch` |
| `stats/internal/records/types.go` | `Dispatch.DiffBase string \`json:"diffBase,omitempty"\`` |
| `stats/internal/records/render.go` | `RenderLedger` emits `- Diff base:` beside `- Commit:`, through `neutraliseMarkers` |
| `stats/cmd/myflow/record.go` | `-diff-base <sha>` on `dispatch begin`, optional |

`RenderPanel` is not touched. The marker contract `check-panel-reproducers.sh` and
`check-unfinished-work.sh` parse is unchanged, and dispatch rows already render into the SDD ledger
rather than the panel record.

**Scheduling reads the dispatcher's in-session value, never the store.** A panel runs inside one
`/myflow-do` invocation, and a re-run of `/myflow-do` starts a fresh pass 1, so no cross-run read is
needed and no read path has to be added. The row is the durable audit trail. This keeps the write on
the same never-block guarantee every `myflow record` write has: an unreachable store journals the
intent and the panel proceeds unaffected — which is only safe *because* nothing schedules off it.

## The empty delta

A required delta-slot whose delta is empty is not dispatched, and the panel record says
`not re-run — nothing new since its last read`.

This is a third disposition beside the two the record already distinguishes — `not re-run — subject
unchanged` for a conditional slot whose trigger did not fire, and a slot the operator declined. It
is a **different mechanism** from trigger-scoping, which still reaches conditional slots alone: a
required delta-slot's bounded region is its own delta, and an empty delta means its previous clean
result is not stale rather than that it was waived. The zero-open-findings bar is untouched, and not
re-running a slot closes, softens and expires nothing it has already raised.

## Decisions

### Which diff each slot gets in a Full-mode re-run

**ID:** full-mode-per-slot-diff
**Status:** active
**Chosen:** Primary keeps the whole `final-review.diff`; every other diff-reading slot gets a
per-slot delta — the only option that keeps the plan-alignment and history remit Primary owns while
removing the re-reading the ticket measured.
**Considered:** *Primary and Principles both keep the whole* — Principles owns hard invariants
(layer purity, new suppressions, weakened lint config), but it judges those against the code it is
shown and pass 1 already showed it the whole change; ruled out as paying for a second full read to
buy nothing. *Every slot goes delta, Primary included* — cheapest, ruled out because it drops the
artifact-and-history remit the ticket names as the reason the split is per-slot rather than global.

### What a delta-slot's delta is anchored at

**ID:** delta-anchored-per-slot
**Status:** active
**Chosen:** the HEAD sha that slot itself last reviewed — exact, no coverage gap, no re-reading, at
the cost of one recorded sha per panel dispatch.
**Considered:** *the current round's `fix-round-N.diff`*, the ticket's literal suggestion and the
cheapest — ruled out because a slot that skipped a Targeted round never sees that round, which the
coverage-waiver rule forbids. *The cumulative union of every fix round since pass 1* — no gap and no
bookkeeping, ruled out because a slot that re-ran every round then re-reads earlier rounds, which is
the cost this change exists to remove.

### Where the base sha is recorded

**ID:** base-sha-on-dispatch-row
**Status:** active
**Chosen:** the store, on the slot's own dispatch row, rendered into the SDD ledger — durable,
survives the worktree, and reuses a row that already exists per slot per round.
**Considered:** *the pass log* (`<abs-worktree>/.superpowers/sdd/final-review-panel.md`) — no schema
change at all, ruled out because it is worktree-lifetime and reaches no durable record.
*Derive it: read which slots ran in which round back out of the panel record and concatenate
`fix-round-k+1.diff … fix-round-N.diff`* — stores nothing, ruled out because concatenated diffs
double-count churned lines and are larger than the single range they stand in for, and because the
panel record is rendered only when the panel closes, so mid-panel there is nothing to derive from.

### Whether the closing pass reads the full diff

**ID:** no-closing-pass-exemption
**Status:** active
**Chosen:** no exemption — the delta rule applies to every pass after 1.
**Considered:** *re-run the required slots against the whole diff once a pass comes back clean* — a
safety net against a defect visible only across two rounds' deltas, ruled out because coverage
already holds by construction and it costs one extra full pass on every change to insure against a
case nothing has observed.

## Open questions

*(none)*
