# pipeline-fix-kan-1062-flow-implement-one-shot-subagents-50 — design

## Context

Evidence: in run kan-1062, row `panel-fix-1-2` recorded commit
`7d3de8536059b219be89a65729b0b33b2a9cd3ba` (`fix(stats): rewrite -from findings atomically and pin
literal HTML characters`, trailer `Flow-Key: panel-fix-1`, on branch
`spectre/kan-1062-flow-implement-one-shot-subagents`), and the guard printed
`panel-fix-1-2 records commit 7d3de853… which is no task's Commit: subject and carries no
Flow-Key: panel-fix-1-2 trailer`.

Cause: `pfdCheckCommits` accepts a task's `Commit:` subject, the row's own key (`-retry`
stripped), or that key one round low (`pfdOneRoundLow`) — never the round's key a chunk row's
commit carries by contract.

The diff in Task 1 was applied to this worktree at `5302fcb2` and checked there: the guard's tests
pass with it, and the new test fails with `panelfixsingledispatch.go` reverted, printing
`panel-fix-1-2 records commit … carries no Flow-Key: panel-fix-1-2 trailer`.

Blast radius: 16 files outside archives cite `check-panel-fix-single-dispatch`.
<!-- measured: git grep -l check-panel-fix-single-dispatch -- . ':!spectre/changes/archive' ':!spectre/changes/pipeline-fix-kan-1062-flow-implement-one-shot-subagents-50' | wc -l @ 5302fcb2 → 16 -->
The fix touches the guard, its test and the shim's header; the other citers state the gate, not
its accepted-trailer list, and stay unchanged.

## Decisions

### Which extra key a chunk row's commit may carry

**ID:** accept-round-key
**Status:** active
**Chosen:** the row's round key, `panel-fix-<r>`, from a new `pfdRoundKey` that strips the `-<n>` chunk suffix — the key the fix-round contract writes.
**Considered:** changing `review-panel-fix-round.md` to key each chunk's commit `panel-fix-<r>-<n>` — contradicts `git-boundaries.md`'s round keying too and breaks commits already written; accepting any `panel-fix-*` trailer — drops the check that a row's commit belongs to its round.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — one added predicate in one guard, pinned by one test.
**Considered:** `flow-medium` — the guard gates every panel close, but the change only widens the accepted trailers by the round key and keeps every other verdict.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — the plan holds one task.
**Considered:** none — a single task has no split.

## Open questions
