# Final review panel — kan-181-dashboard-period-presets-and-reset

## Roster

**Preset:** `light`. Three required slots, all on `sonnet` (`models.reviewPanel`): Primary,
Principles (`[LENS]` = Merged), Code review (low).

**Operator instruction for this run:** light roster with **capped fix rounds** — stop after the
findings converge rather than chasing every new one, and bring anything remaining to the operator as
a decision. Recorded here because it governed how the panel was run, not just what it found.

**Optional slots — two triggers fired, all four declined by the operator.** Declined is recorded
distinctly from a trigger that never fired.

| Slot | Trigger | Disposition |
|------|---------|-------------|
| 4 — Security | borderline: the change parses period bounds out of an untrusted URL | declined by operator |
| 5 — Adversarial | 722 changed lines (>300); modified test files | declined by operator |
| 6 — Lens B (simplicity & state) | 722 changed lines (>200) | declined by operator |
| 6 — Lens C (robustness & ops) | did not fire | not triggered |

## Panel diff size

`scripts/check-panel-diff-size.sh <worktree> 0256da3` measured **722** changed lines, **under cap**,
exit 0.

## Pass 1

**Mode:** full roster (pass 1 always runs the full roster selected for the change).
**Diff read:** `.superpowers/sdd/final-review.diff`, from `git diff 0256da3`.
**Bounced findings:** none.

**All three slots returned READY, and all three independently raised the same single Minor.** That
convergence is why it was fixed rather than withdrawn: three reviewers reaching the same line by
different routes is evidence the comment genuinely misleads.

## Findings

Per-task review found four defects before the panel ran; the panel found one. All five are fixed.

| ID | Stage | Slot | Severity | Location | Note |
|---|---|---|---|---|---|
| F1 | per-task (1,2,5) | combined reviewer | Important | `stats/web/src/App.tsx` | the URL write-back clobbered an explicit period arriving via `hashchange` with the stale in-memory one |
| F2 | per-task (3) | combined reviewer | Critical | `stats/web/src/components/PeriodPicker.tsx` | preset marking and reset visibility compared a stored period against a freshly recomputed `now` |
| F3 | per-task (4) | combined reviewer | Critical | `stats/web/src/components/DataTable.tsx` | the same moving-clock comparison made the "period may be why" hint appear under the default period |
| F4 | per-task (3,4) | combined reviewer | Minor | `stats/web/src/components/DataTable.tsx` | the instant-equality comparison was reimplemented rather than shared |
| F5 | panel pass 1 | Primary · Principles · Code review | Minor | `stats/web/src/App.tsx` | a comment claimed `onUrlPeriod` had a stable identity across renders when it is a fresh closure each render |

### F2 and F3 — one root cause, and the tests were complicit

`defaultPeriod()` and each preset's `range()` recompute `now` per call, while the stored `period.to`
is a snapshot. Comparing them compares a fixed instant against a moving target, so any re-render a
few seconds later — an unrelated filter change, a fetch resolving — unmarked a just-selected preset
and made the reset appear on an untouched default. In `DataTable` the same comparison broke the
scenario the change exists for: the "period may be why" hint appeared on a genuinely empty view under
the default period, which is precisely the outcome `design.md` says must not happen.

**Both were invisible to the suite because every test used fake timers and asserted immediately after
render, never across a re-render with the clock advanced.** The fix therefore included three
regression tests that do advance the clock and re-render; without them the same class of defect
returns unseen. The fix itself is one exported `sameRange` with a ±60s tolerance, shared by
`PeriodPicker` and `DataTable` — which also closed F4.

### F5 — why a comment-only finding was fixed rather than withdrawn

The comment stated the wrong reason the code is safe. Safety here is a property of the closure's
**body** — it forwards to `setPeriod`'s functional update form and reads nothing from `App`'s scope —
not of the dependency array. A future edit that made `onUrlPeriod` read `route` or `period` directly
would read that value's mount-time state forever, because the effect never resubscribes. A comment
that misdirects an audit of exactly that omission is worth correcting.

## Verification

Measured against the final tree:

- `cd stats/web && npm test` — 141 passed, 8 files (baseline was 112).
- `cd stats/web && npx tsc -b` — clean.
- `TZ="America/New_York" npx vitest run src/components/DashboardBar.test.tsx` — passed; the `Today`
  preset is local-midnight based and does not regress outside UTC.
- Five commits, one per task, each non-empty, each carrying its `Task-Id` trailer, no stray `fixup!`.

<!-- measured: all four run in the worktree on 2026-08-15 against the final commits -->

## Recorded, not fixed

**A bound with no timezone parses in the browser's local zone rather than UTC.** This is inherent
`Date` string-parsing behaviour, not introduced by this change, and the application's own write-back
always emits `toISOString()`, so it is reachable only through a hand-crafted external link. Noted so
the next reader does not rediscover it as a defect.

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed

reproducers-total: 5
finding-reproducer: F1 npm test
finding-reproducer: F2 npm test
finding-reproducer: F3 npm test
finding-reproducer: F4 none — a duplication finding with no runnable check
finding-reproducer: F5 none — a comment-accuracy finding with no runnable check
