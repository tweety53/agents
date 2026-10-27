# Design — pipeline-fix-kan-1039-end-to-end-autonomy-rule-36

## Context

A creating `/flow` run records its brainstorming dispatches before `spectre new`; the record
writer's checkout check finds no directory naming the change yet and refuses them (`proposal.md`).
`changeCheckouts` has three callers: `requireChangeCheckout` (the refusal), the inline-row tree
resolution in `record.go`'s dispatch begin, and `gatePanelFindingsClosed` in `stage.go`. The last
two pick the change's own worktree from the carrying list; a worktree on branch `spectre/<change>`
is that worktree, so the widened match keeps their answer correct.

## Decisions

### Accept the change's branch as carrying it

**ID:** branch-carries-change
**Status:** active
**Chosen:** `changeCheckouts` also counts a `git worktree list --porcelain` entry whose `branch` line
is `refs/heads/spectre/<change>` — one comparison in the loop that already parses that output, and
it covers every creating run, including runs whose worktree already exists.
**Considered:** `kickoff-worktree.sh` (Go in `stats/internal/guard/kickoffworktree.go`) creating
`<worktree>/.superpowers/sdd/<name>/`, mirroring `pipeline-fix.sh start` — rejected: it fixes only
worktrees kicked off after the fix lands (the kan-1039 worktree in flight stays refused), and it
leaves an empty directory as a side-channel marker for a fact git already states.

### Nothing to repair retroactively

**ID:** no-retroactive-repair
**Status:** active
**Chosen:** no data migration — a refused `dispatch begin` exits 2 before writing, so no row was
filed under any project and nothing stored is wrong.
**Considered:** back-filling the missing rows — rejected: the dispatches' timings and keys are not
recoverable from any store, and `-started-at` back-fill is the operator's call per run.

### Effort gap stays out

**ID:** xhigh-effort-out-of-scope
**Status:** active
**Chosen:** leave `recordEfforts` unchanged; the kan-1039 change decides `xhigh`.
**Considered:** adding `xhigh` here — rejected: the dispatch said not to, and it is a separate decision.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-low — a one-condition Go change with one test; the review surface is a dozen lines.
**Considered:** flow-medium — rejected: no concurrency, interop or data-model seam to reason over.

## Open questions

## Recommended approach

Widen `changeCheckouts`' match in `stats/cmd/flow/record.go`: while scanning
`git worktree list --porcelain`, remember each entry's `branch` line; an entry whose branch is
`refs/heads/spectre/<change>` is carrying even when none of the three directories exists. Update
`requireChangeCheckout`'s comment and refusal message to name the branch. The kickoff-side marker
directory lost on `branch-carries-change`; back-fill lost on `no-retroactive-repair`.

## Critical files

- `stats/cmd/flow/record.go` — `changeCheckouts` and `requireChangeCheckout`, edited.
- `stats/cmd/flow/record_test.go` — the regression test, added.
- `stats/cmd/flow/stage.go` — `gatePanelFindingsClosed`, a caller of `changeCheckouts`, unchanged.

## Existing code to reuse

- `changeCheckouts`' existing porcelain loop (`stats/cmd/flow/record.go`).
- `gitRepo(t)` test helper and the `git worktree add` pattern in `stats/cmd/flow/record_test.go`.

## Verification

- `cd stats && go test ./cmd/flow -run 'TestRequireChangeCheckout'` passes; with the branch match
  reverted, the new test fails.
- End to end: in a fresh kickoff worktree before `spectre new`, `flow record dispatch begin -change
  <name> -C <worktree> -role explorer …` exits 0 instead of 2.
