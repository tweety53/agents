## Context

`lpFirstReviewType` (`stats/internal/guard/loop.go`) returns the subagent type `review-1` must run
on, from the `first-review-effort` decision in the loop's `design.md`. It reads
`refs/heads/<ref>:spectre/changes/<lp.change>/design.md`, `<ref>` being the loop branch
(`lp.token`) or, when that does not exist, `spectre/<lp.change>`.

A loop's plan is always the spectre change named after the loop's branch name, `lp.token`:
`pipeline-fix-<name>-<k>`, or `self-review-<name>` — on its own branch or on the change's own
branch alike (`skills/flow-self-review/SKILL.md` step 3 item 1: the planner "plans them all in full
as one spectre change, `self-review-<name>`"). For a pipeline-fix loop and a `self-review-<name>`
loop branch, `lp.change == lp.token`, so the read was right. For a self-review on the change's own
branch, `lp.change` is `<name>` — the change its flowd rows are filed under — and the read misses
the plan: the checker always demands `flow-medium`. Pipeline-fix 8 (a4a32afc), which first made an
own-branch loop in another project runnable, wrote its test fixture at the same wrong path, so its
test encoded the defect.

On a `/flow-fast` change the change's own branch is `<name>`, not `spectre/<name>` (step 3's
`<fix-branch>` paragraph), so the fallback ref misses there too.

The `-C` is right: `checkLoopDispatch` passes `lp.main`, which for an own-branch loop is the main
checkout of the caller's working directory (fix 8); branches are shared across a repository's
worktrees, so `refs/heads/spectre/<name>` resolves from it.

Other readers of a loop's plan directory are right: `looplanding.go` reads no plan; `pfLand`
(`pipelinefix.go`) checks `spectre/changes/<loop branch>/` and lands only loop branches; the
skill's ticks and archive (step 3 item 4) name `self-review-<name>`; `lpRows`, `dispatch begin`,
`dispatch end` and the escalation markers key on `lp.change` and `lp.token` as fix 8 intends.

Reproduced at plan time at 711c69c9: with the task's test hunks alone,
`TestCheckLoopDispatchOwnBranchInAnotherProject` and
`TestCheckLoopDispatchOwnFastBranchReadsLoopPlan` both fail with the gymie run's exact refusal; with
the whole diff, `cd stats && go test ./internal/guard/ -count=1` is green.

Blast radius — every file naming `lpFirstReviewType` or `first-review-effort` outside `.git` and
the spectre archive — not `big`: 2 files.
<!-- measured: grep -rl -e lpFirstReviewType -e first-review-effort --exclude-dir=.git --exclude-dir=archive . @ 711c69c9 -->

## Decisions

### Read the loop's plan by the loop's branch name

**ID:** plan-dir-is-the-token
**Status:** active
**Chosen:** the path is `spectre/changes/<lp.token>/design.md` in every case — the plan is always
named after the loop's branch, which `lp.token` already is; one expression, no case split.
**Considered:** `"self-review-" + lp.change` when `lp.change != lp.token` — a branch restating
what the token already carries; a new `lpLoop.plan` field — stored state derivable from the token.

### Fall back to the /flow-fast own branch

**ID:** fast-own-branch-fallback
**Status:** active
**Chosen:** when neither the loop branch nor `spectre/<name>` exists, read from `<name>`, the
`/flow-fast` change's own branch — the same defect class, closed in the same few lines.
**Considered:** leaving it — a `/flow-fast` own-branch loop would keep being forced onto
`flow-medium` the same way.

### First review on low

**ID:** first-review-effort
**Status:** active
**Chosen:** the first review, `pipeline-fix-9-review-1`, runs on `flow-low` (effort `low`) on
`opus`: a one-expression path fix and a short ref fallback in one read-only function, with both
tests reproducing the live refusal.
**Considered:** `medium` — more than a two-file change to one read-only function needs.

### One fixer group

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group — Task 1 in one `pipeline-fix-9-fix` dispatch.
**Considered:** none other — the plan has one task.

## Open questions

## Live check

No runtime state changes: the guard is a one-shot process per call. After landing, the gymie run's
`self-review-kan-954-user-row-loaded-twice-per-request-review-1` dispatch on `flow-low` / `opus` is
admitted (`LOOP-DISPATCH-OK`); failure looks like the same `PAIR — … runs on flow-medium, as the
loop's design.md first-review-effort decision records` refusal.
