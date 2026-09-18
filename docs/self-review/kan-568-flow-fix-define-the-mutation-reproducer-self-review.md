# kan-568-flow-fix-define-the-mutation-reproducer — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-568-flow-fix-define-the-mutation-reproducer-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** panel F2 (deferred, message-wording only): the exit-contract guard's exit-1 message reads backwards for a declared mutation reproducer whose healthy not-demonstrated reading is non-zero; the guard verdict is already correct, so the fix is the message plus its harness pins — filed: KAN-622
- **[flow-fix]** panel F4 (deferred): the documented 10-line edge of the mutation-reproducer window is unpinned (the fixture's prepended lines put case 28.a's marker at line 15, so a window of 1-14 passes); add marker-at-line-10 and marker-at-line-11 boundary cases — filed: KAN-623

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

_none — this angle produced no findings._

## What went well, and how to reproduce it — `flow-improvement`

_none — this angle produced no findings._

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** panel F5 (principles, deferred): the `# mutation-reproducer` literal and the 10-line window each live in review-panel.md and run-reproducer.sh with no mechanical pin keeping them together; extend a guard to pin the pair so the two statements cannot drift — filed: KAN-624

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** all three dispatch rows read `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
