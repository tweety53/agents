## Why

`/myflow-do`'s review panel has two re-run modes. **Targeted** already hands every re-running slot
`fix-round-N.diff`. **Full** — the escalation mode — hands every required slot the rewritten
`final-review.diff`: the whole `git diff <merge-base>`.

*Three or more fix rounds have already run* is itself an auto-escalate trigger, so from round 4
onward every round is Full.

KAN-315 ran twelve rounds; three slots re-read a diff that reached 3,300 lines every round, while
each round's actual subject was a 200-to-900-line delta.
<!-- measured: KAN-315's self-review, as reported in KAN-327's description -->

That delta was already written to `<worktree>/.superpowers/sdd/fix-round-N.diff` and already named
as the slot's primary target. By round 6 the production files had been unchanged for several rounds
and the new work was confined to test guards, but every slot was still paying to re-read the compose
file, the workflows and the scripts.

The deltas exist and the panel step already writes them. What is missing is the rule that after
round 1 the delta is a slot's **input**, not an emphasis inside a full diff.

## What Changes

- The **Full** row of **Panel re-runs** (`skills/myflow-do/SKILL.md` §5) stops giving every required
  slot the rewritten `final-review.diff`. Slot 0 (Primary) keeps it; Principles, Code review (low),
  Adversarial and the extra principles lenses each get a **per-slot delta** instead.
- A slot's delta is `git diff <the HEAD sha that slot last reviewed> HEAD` — anchored at that slot's
  own last read, not at the current round, so a slot that skipped a Targeted round still sees what
  it missed.
- A required delta-slot whose delta is empty is not dispatched; the panel record says
  `not re-run — nothing new since its last read`.
- Each slot's dispatch prompt states which diff it is holding and, for a delta, the sha it starts
  from.
- `myflow record dispatch begin` gains an optional `-diff-base <sha>`, stored on the dispatch row
  and rendered into the SDD ledger, so the audit trail says what each panel slot actually read.
- **Targeted** mode is unchanged, and Bugbot and Security are untouched — they are dispatched by
  `subagent_type` with `Diff: uncommitted changes` and read no diff file.

## Capabilities

### New Capabilities

*(none)* — the rule belongs to the panel's existing economics capability, and the recorded base is
one more field on a dispatch row that `myflow-run-record` already defines.

### Modified Capabilities

- `myflow-review-panel-economics`: which diff each slot reads in a Full-mode re-run, and the third
  not-re-run disposition.
- `myflow-run-record`: a dispatch row records the diff base its subagent read from.

## Impact

- `skills/myflow-do/SKILL.md` — §5's **Panel re-runs** table and the dispatch-recording block;
  `skills/myflow-do/SKILL-rationale.md` — why the anchor is per-slot rather than per-round.
- `stats/internal/store/migrations/0014_dispatch_diff_base.sql` — new, one column.
- `stats/internal/store/records.go` — the dispatch column list and the insert.
- `stats/internal/records/types.go` — `Dispatch.DiffBase`.
- `stats/internal/records/render.go` — `RenderLedger`'s `- Diff base:` line. `RenderPanel` is not
  touched, so the marker contract the two panel guards parse is unchanged.
- `stats/cmd/myflow/record.go` — the `-diff-base` flag.
- No new dependency, no new guard.
