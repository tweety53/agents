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
