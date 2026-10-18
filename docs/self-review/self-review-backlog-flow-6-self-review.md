# Self-review report for self-review-backlog-flow-6

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** 4 — good

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `check-task-records.sh` missed a task commit landed in a peer repository's worktree (in-run pipeline fix) — fixed: 646ac917
- **[flow-fix]** Two inline dispatch rows begun 30 ms apart: the store capped the first row's window at the sibling's begin, so `verify-agents` got ~145k cache-read tokens against `verify-gymie`'s 7.18M; a window is now capped only at a later inline begin at least `minInlineWindow` after it — fixed: d8cdc673
- **[flow-fix]** A plan task created the gymie worktree with a bare `git worktree add`, duplicating **2. Isolate the workspace**; `brainstorm-planner.md` section D now says a task never creates a worktree — fixed: a8d90f57
- **[flow-fix]** `integrate-change.sh` not found on PATH; **Guard resolution** resolves guards to `<skill-dir>/scripts/` by design — a one-off agent slip, recovered on the next call — noted
- **[flow-fix]** The main-checkout hook denied a Python heredoc containing `>`; reproduced, and it is the hook's documented `strip_data_heredocs` approximation — noted
- **[flow-fix]** The store refused task 1's inline row closed under 3 s; `implement.md` already says to record begin before the work — noted
- **[flow-fix]** The plan held the adapter to be the only `FriendshipReadRepository` implementation; two test fakes also implement it, caught by the first compile — a planner judgement, not a rule gap — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** The verifier rows' token counts are mostly the parent's in-run fix loop inside the verify window — the attribution defect fixed under `flow-fix` — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Task 5's RED was taken on an unlocked first cut, and a second mutation proved each lock separately — noted
- **[flow-improvement]** Tasks 1–4's diffs were pre-checked at plan time and applied clean with every guard green on the first run — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** The parent hand-wrote a verify runner every run from `verify-and-handoff.md`'s prose; `scripts/run-verify-list.sh <worktree> lint|test` now runs a project's verify list, stdin detached — fixed: b8a5c1d0

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** The 24-minute verify window includes the 14-minute in-run pipeline fix; a red lint blocks the handoff, so the wait is unavoidable — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case — Task 5, lock the friendship row through an access grant (KAN-733); operator chose not to mine — noted
  - Commit: gymie `3d51d3aa`; code `/Users/tweety53/Agents/flow/state/gymie-7c1f238a/flow-eval/self-review-backlog-flow-6/AccessServiceImpl.kt`
  - Tests: `SocialGraphAdapterTest.kt:435`, `:474` (30 → 32)
  - Why: a real check-then-act race needing a lock on both sides; the lock-wait tests reject a `@Transactional`-only fix; the new port method breaks two test fakes
