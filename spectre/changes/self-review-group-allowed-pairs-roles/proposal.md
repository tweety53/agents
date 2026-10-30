# self-review-group-allowed-pairs-roles

## Why

The self-review pass of `group-allowed-pairs-roles` (`.superpowers/sdd/self-review-pass.md` in the
change's worktree) found four non-big defects in the pipeline it ran on:

- **F1** — the mutation slot's own mutant runs carry no test-scope rule: the whole-package /
  name-prefix rule binds only a surviving mutant's reproducer, and its threshold reads one run
  against a third of the ceiling, never runs times mutants. `panel-0-mutation+failure-modes` timed
  out; the prefix-filtered retry finished quickly; kan-1025 was the same failure.
- **F2** — an approved prototype rendered as a region capture's frame is under-specified: step 10
  of the verifier names no region-width sizing (the prototype's standalone panel width against a
  wider capture), never says `frame none` composes with an n/a ratio and no `.frame.png` by
  design, and step 8 plus the planner's capture-task paragraph forbid the clipped region such a
  prototype needs.
- **F3** — the round-1 verifiers named only the prototype-framed roles captures; the full-page
  captures this change moved were never stated or swept, so the clipped last x-axis label of the
  weekly-trend chart surfaced one round later and cost `verify-fix-1`. The rule exists in prose;
  `check-verify-report.sh` does not enforce it.
- **F4** — the full-app-suite zip rebuild is a separate manual step after `capture`; task 6's first
  commit missed it, and the script refuses a relative root.

## What changes

- The mutation-testing brief scopes the slot's own mutant runs by the planned mutant count.
- The verifier's step 8 admits a region capture beside the view's full-page capture where an
  approved prototype draws that region alone; step 10 sizes the prototype's region to the capture
  and states the `frame none` reads; the planner's capture-task paragraph names the exception.
- `check-verify-report.sh` refuses a report that leaves a resolved capture without its
  `- <view>: <path>` line, or a view with no composed frame without its capture-scope `sweeps:`
  line; step 9 names the refusal.
- This repository's `capture` row rebuilds `full-app-suite.zip` after Playwright with absolute
  arguments; step 8 states that a chained `capture` has already run the rebuild.
