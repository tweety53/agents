# Plan Correction blocks and superseded decisions

When a re-plan overturns a premise, a task, or a decision, never rewrite the record: append a
dated Correction block stating what changed and why, and mark a superseded decision superseded,
naming what supersedes it. The plan then shows the path the run actually took — wrong premises
included — instead of a smooth history that never happened.

## Why

kan-749's plan stayed truthful through three re-plans and eight corrections: task 3's premise
did not reproduce (the 360dp test reached the last card on the unmodified tree) and the task was
re-planned around the measured cause (the page 56px taller than the screen), with the superseded
decision marked `rule-without-intrinsics` → superseded by `standalone-height-ios-only`; every
later change carried a dated Correction block naming what no longer holds — in task 6's case,
explicitly retiring two of its own earlier statements. A rewritten history hides exactly the
detour a later reader needs in order to trust the rest of the plan.

## How to run one

- Never rewrite a task's or a decision's history. The original text stays; what no longer holds
  is corrected by an appended block, not an edit.
- A Correction block opens `Correction (<date>` — naming the round, finding or measurement that
  forced it when one exists — and states in a sentence or two what changed and why.
- A superseded decision is marked, not deleted: `X` superseded by `Y`, so a reader who lands on
  `X` finds where the run went instead.
- Wrong premises stay in the record. A later reader is entitled to see that the premise failed
  and what replaced it.

Recorded from KAN-812, filed from the deferred self-review of
kan-749-mobile-scroll-runner-state-safe-area.
