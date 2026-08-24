# Review the per-round delta, not the whole diff, after round 1

**Change:** `kan-327-review-per-round-delta-after-round-1`
**Jira:** KAN-327
**Date:** 2026-08-24

## Problem

`/myflow-do`'s review panel has two re-run modes. **Targeted** already hands every re-running slot
`fix-round-N.diff`. **Full** — the escalation mode — hands every required slot the rewritten
`final-review.diff`, the whole `git diff <merge-base>`.

*Three or more fix rounds have already run* is itself an auto-escalate trigger, so from round 4
onward every round is Full. KAN-315 ran twelve rounds: three slots re-read a diff that reached
3,300 lines, every round, when each round's actual subject was a 200-to-900-line delta already
written to `<worktree>/.superpowers/sdd/fix-round-N.diff`.

The deltas exist and are already written by the panel step. What is missing is the rule that after
round 1 the delta is a slot's **input**, not an emphasis inside a full diff.

## Scope

The **Full** row of **Panel re-runs** (`skills/myflow-do/SKILL.md` §5) and nothing else.
Targeted mode is untouched: it is already the cheap path, and Primary has always acted as the
delta's integration check there.

Only slots that read `final-review.diff` are affected — Primary, Principles, Code review (low),
Adversarial, Lens B and Lens C. Bugbot and Security are dispatched by `subagent_type` with
`Diff: uncommitted changes` and a repository path; they read no diff file and this change does not
reach them.

## The rule

In a Full-mode re-run:

| Slot | Diff it is given |
|------|------------------|
| 0 Primary | the rewritten `final-review.diff` — the whole change |
| 2 Principles, 3 Code review (low), 5 Adversarial, 6+ Lens B / Lens C | its own **per-slot delta** |

Primary keeps the whole diff because its remit is plan alignment and the change's history, which
a delta cannot carry. The other slots judge the code in front of them, and the code in front of
them is what has changed since they last looked.

Every slot's dispatch prompt states which of the two it is holding, naming the path and, for a
delta, the sha the delta starts from.

## The per-slot delta

A slot's delta is

```
git diff <the HEAD sha that slot last reviewed> HEAD
```

**Per-slot, not per-round.** In Targeted mode only slot 0 and the slots that raised findings
re-run, so when round 5 escalates to Full, Principles may not have read rounds 2 through 4.
`fix-round-5.diff` alone would leave that gap, and *targeting is a cost optimization, never a
coverage waiver* forbids it. A range anchored at the slot's own last read closes the gap exactly,
with no re-reading.

**Tree-to-tree, deliberately.** `git diff A B` compares two trees and needs no ancestry between
them, so a base sha recorded before a `git rebase --autosquash` stays valid after the rebase
rewrote the task commit. A range anchored at a merge base or expressed as `A..B` would not.

**Coverage holds by construction.** Pass 1 always runs the full roster against the full
`final-review.diff`, so every slot has a real starting sha, and every change made since is inside
some slot's delta. There is no closing-pass exemption: a pass that closes the panel follows the
same rule as any other pass after 1.

**A slot with no recorded base gets the full diff.** An unknown base is not a small delta; the
safe reading is the whole change.

## Recording the base

The base is recorded in the store, on the panel slot's own dispatch row — the row that already
says which slot ran in which round, keyed `panel-<round>-<slot>`.

- A `diff_base TEXT` column on `dispatches`, added by a migration in the same single-column shape
  `0011_dispatch_agent_id.sql` and `0012_dispatch_key.sql` already use.
- `-diff-base <sha>` on `myflow record dispatch begin`, carrying the sha the slot is about to read
  from. Optional: a dispatch that reads a full diff records none.
- `RenderLedger` emits it as a `- Diff base:` line beside the existing `- Commit:` line. Dispatch
  rows render into the SDD ledger; `RenderPanel` renders findings and markers and is not touched,
  so the marker contract the two panel guards parse is unchanged.

**Scheduling reads the dispatcher's own in-session value, never the store.** A panel runs inside
one `/myflow-do` invocation and a re-run of `/myflow-do` starts a fresh pass 1, so no cross-run
read is needed. The store row is the durable audit trail. This keeps the record write on the same
never-block guarantee every other `myflow record` write has: an unreachable store journals the
intent and the panel proceeds unaffected.

## Empty delta

A required delta-slot whose delta comes back empty has nothing new to read. It is not dispatched,
and the panel record states `not re-run — nothing new since its last read`.

This is a third disposition beside the two the record already distinguishes — `not re-run —
subject unchanged` for a conditional slot whose trigger did not fire, and a slot the operator
declined. It closes, softens or expires no finding: the zero-open-findings bar still governs every
slot in the roster, and an empty delta means the slot's previous clean result is not stale rather
than that it was waived.

## Decisions, open questions and the implementation mechanism

Recorded once, in this change's OpenSpec `design.md`
(`openspec/changes/kan-327-review-per-round-delta-after-round-1/design.md`) — the artifact the
pipeline's handoff counts them from. This file is the brainstorming record: the problem, the scope
and the rule.
