# Known bugs — fixed in the run that finds them

**This file is canonical for what a `/flow*` run does with a defect it finds in the project's own
product code — its change's own, or one the change did not introduce: a pre-existing test failure,
a visual departure present at the merge base, a finding a verification change's sweep makes on a
surface another change owns.** Skills reference it as **the sweep**; none of them restate it.

## The sweep

A product defect a run finds is fixed inside that run, before integrate lands the change — never deferred to self-review.

**A defect the change did not introduce is no exception.** It blocks exactly as the change's own
does and takes the same course — **The loop** (`skills/flow/verify-fix-loop.md`) in verify and
visual verification, an appended task for a verification change's sweep (**D. Basic Workflow #3 —
Writing plans**, `skills/flow/brainstorm-planner.md`). It is never logged to a `<project>/KNOWN-BUGS.md`
file, never filed for later, and never skipped, worked around or
tolerance-widened (**Fix determinism at the source, never by widening tolerance**,
`rules/fix-determinism-at-the-source.mdc`).

**A product defect the self-review pass meets is never dropped.** It stays out of the pass's
report and filing prompt, and stops the landing. Met inside a run, the pass stops at the end of its
step 2 — no fix, no prompt, no store row, no report: nothing is committed for the pass, so the
next run meets no committed report, runs the pass again, and meets the defect again if it was not
fixed. The `flow.self-review` mark closes `-outcome stopped`, nothing lands, the change keeps its
state, and the run ends on the block below.

- **`/flow`'s integrate run 1** stops before step 5's route. The fix run writes its plan above the
  archive commit, so the next bare `/flow <name>` is an archived re-run with new work, which runs
  the pass (**Run 1 — the branch is not merged**, `skills/flow-contracts/finish-contract-run1.md`).
- **`/flow-fast`'s verify** stops before `flow.stage-diff`; the fix re-run lands normally.

Met by a standalone `/flow-self-review` after landing, the pass's closing report names it for the
operator to start that fix run.

```
## Landing stopped — a product defect to fix first

**Change:** <name>
**Defect:** <the defect, one line>

Next:
/clear
/flow <name> <the defect, one line>
```

`/flow-fast` prints `/flow-fast <name> <the defect, one line>` as the last line instead.
