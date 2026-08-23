# Self-review — kan-212-persist-per-dispatch-cost-tokens-model-and-role

**Operator rating:** 3 — mixed.

**Measured, from the `dispatches` table this change itself repaired:** 34 dispatches, ~204.7M
subagent tokens, 33 stage runs over ~235 minutes, across three `/myflow-fast` invocations.

| Role | Dispatches | Tokens | Share |
|------|-----------:|-------:|------:|
| reviewer (per-task + panel) | 16 | 91.3M | 45% |
| implementer | 8 | 67.3M | 33% |
| red-partner | 5 | 30.5M | 15% |
| panel-fix | 5 | 15.6M | 8% |

<!-- measured: SELECT role, count(*), sum over metrics->'tokens'->'sidechain' FROM dispatches WHERE change = kan-212 @ 2026-08-23 -->

## 1. Problems encountered — `myflow-fix`

- **[myflow-fix]** `check-task-commit-fields.sh` exits 2 in a worktree holding more than one unarchived change, so it was bypassed by hand for all 14 tasks — declined
- **[myflow-fix]** a panel fix spanning several tasks has no defined `--fixup` target, so F16/F17/F18 was committed standalone against the contract — declined
- **[myflow-fix]** the pipeline renders records with an unchecked CLI version, and a stale binary silently under-reported ten stamped rows — declined

## 2. Token and time cost — `myflow-cost`

- **[myflow-cost]** a red/green task pair pays for context acquisition twice, five times over, at 30.5M tokens for the red halves alone — filed: KAN-318
- **[myflow-cost]** a red task's per-task review reads a strict subset of what its green partner's review reads minutes later — filed: KAN-319

## 3. What went well — `myflow-improvement`

- **[myflow-improvement]** mutation proof caught four defects that passing tests had hidden — F1, F9, F10 and F18 — filed: KAN-320
- **[myflow-improvement]** the live-verification task found the change's headline claim false where no unit test could — filed: KAN-321
- **[myflow-improvement]** agents reported gaps rather than inventing, which produced task 6.1 and the four-state spec correction — declined

## 4. What could be automated — `myflow-automation`

- **[myflow-automation]** the `-agent-id` capture is the hand-written call KAN-212 predicted would be forgotten, and all 34 dispatches forgot it — filed: KAN-322
- **[myflow-automation]** a run hand-types ~100 bookkeeping calls, and two stages this run were marked outside the work they time — filed: KAN-323

## 5. What could move to the Go app or its storage — `myflow-stats-app`

- **[myflow-stats-app]** the daemon should stamp a dispatch's start and end, since agent-typed windows rounded to the minute are the root cause of the interval being unreliable — filed: KAN-324
- **[myflow-stats-app]** rendering should live in the daemon rather than in whichever CLI is installed, since a stale one produced a well-formed under-reporting ledger at exit 0 — filed: KAN-325

## What this change turned out to be

KAN-212 asked for per-dispatch cost to be persisted. That had already shipped with KAN-258, four days
after the ticket was filed, so the change became a repair of the figures being recorded: 20 of 57
rows costed, and the review panel's concurrent slots collapsing onto one row so that one slot carried
the whole group's spend while its siblings recorded zero.

The scope was put to the operator once the audit established it, rather than assumed. The aggregation
views KAN-198, KAN-208 and KAN-201 each want were deliberately left out, on the grounds that building
a per-slot cost view on the old attribution would have published a confident number attributing a
whole panel's spend to whichever slot sorted first.

The two `myflow-fix` findings about the guard and the fixup target were declined for filing; the
third is filed from angle 5 as KAN-325, which proposes the structural repair rather than a check.
