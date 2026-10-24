# Self-review report for kan-1030-auto-land-by-default

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** The panel's base sync ran before the before-review planning commit, so an inline run's uncommitted change folder made git refuse the rebase; the planning commit now comes first — fixed: aa323919
- **[flow-fix]** The auto-land path dropped the `IN_PROGRESS` block's `Decisions:` lines, landing over open `?` questions; an open question now stops at the handoff and 🤖 lines print before landing — fixed: cd44ba7b
- **[flow-fix]** The archived narrative claimed the live check was substituted and `close-task.sh` was missing, contradicting `live-verification.md` and the base tree; both lines corrected — fixed: 631fe5b8
- **[flow-fix]** `stage-task-files.sh` absent from the worktree is skew from the agents repo running its own newer installed skill; the run used the installed scripts dir — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Inline parent context made task 3 (2.25M cache-read) and the verifier (2.0M) heavy; inherent to the `execution: inline` decision — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Reusing the `withdrawn` precedent (column, DTO, contract) kept tasks 1 and 2 under 40s each — noted
- **[flow-improvement]** The live check built `flowd` at base and branch on a scratch port and DB, showing the journal fallback before and the pass after, without touching the dev daemon; mutation killed 3 of 3 — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** The auto-land path still rebuilt and restarted the project stack at `flow.run-instructions` with no handoff to serve; a run that lands straight away now starts nothing — fixed: 68086aac
- **[flow-speed]** 7m20s between task 3 and the panel launch was partly the base-sync refusal, removed by aa323919 — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case — F1, auto-land path skipped the post-guard summary re-record; not offered for mining — noted
  - severity: Important; location: `skills/flow/verify-and-handoff.md:533`
  - reviewer: "auto-land path prints no block, so summary bullets changed after the guard are never re-recorded"
  - fixed in: 96cbb4af, squashed into a6704fba
  - why: a gap that appears only once a new branch skips a block another rule relied on
- **[flow-eval]** impl case — Task 2, API decode and serve `reviewBeforeLanding`; not offered for mining — noted
  - code: `/Users/tweety53/Projects/agents-worktrees/kan-1030-auto-land-by-default/stats/internal/api/changes.go:37`
  - why: the `DisallowUnknownFields` trap — without the DTO field `state set` silently falls back to the journal
