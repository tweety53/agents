# kan-536-extract-a-shared-hook-registration-helper — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-536-extract-a-shared-hook-registration-helper-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** one `flow record pass` call dropped `-change` and briefly masked a guard run; the CLI should refuse a pass record with no change association instead of recording one — filed: KAN-590
- **[flow-fix]** panel F1 (deferred as pre-existing, reproducer available): the protect-main-checkout pins assert a warning about a hook `make_fixture_repo` never installs, so the pin passes without its precondition ever being real; the fixture should install the hook or the pin should be conditional — filed: KAN-591
- **[flow-fix]** the bundled panel reply again omitted the MODEL HANDSHAKE line and the run again recorded the breach rather than re-dispatching; filed from the kan-533 pass — declined (duplicate of KAN-582)
- **[flow-fix]** `check-plan-shape.sh` rejected the plan's indented field lines until they moved to column 0; the guard held and the author corrected, so no change filed — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** the bundled primary+principles review dispatch ran about 24 minutes; the scoped-entry-context reducer is already filed from the kan-512 pass — declined (duplicate of KAN-521)

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** the run pinned the five previously unasserted warning/snippet outputs before refactoring and proved byte-identity twice (the 664-assertion pinned suite plus an old-vs-new sandbox installer diff); reproduce by pinning current outputs first and landing with both proofs — filed: KAN-592

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** the byte-identity proof was a hand-rolled sandbox diff that took three attempts (fixture without setup.sh at its root, a lost exec bit, a reused HOME turning run two into an idempotent refresh); one tested old-vs-new installer-diff helper would retire the per-change hand work — filed: KAN-589

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** the dispatch row reads `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
