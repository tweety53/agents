# self-review-kan-960-group-session-floor-hidden-tab-polling

## Why

The self-review pass over kan-960 found three non-`big` defects, each held against the tree
(`design.md`, **Context**):

- `flow record dispatch end` refused a correctly recorded inline verify: a cached docs-only verify
  finished in 7.6 s, under the store's ten-second inline-window floor, and the forced late close
  skewed the row's window. The floor exists to refuse a begin and end recorded in one Bash call,
  milliseconds apart.
- An implementer left a scratch worktree and branch (`gymie-playwright-worktrees/kan-960-measure`,
  branch `kan-960-measure`, an untracked harness spec) in a repository outside the change's set;
  cleanup removes only the change's own worktrees and branches, so it was left behind.
- kan-960's planned hidden-tab harness was void twice: Playwright's focus emulation keeps attached
  pages `visible`, and CDP `/json/activate` fires `visibilitychange` but not the window
  `blur`/`focus` a real tab switch fires. Nothing a gymie planner or implementer reads says so.

## What changes

- The inline-window floor drops from ten seconds to three: a one-second close is still refused, a
  seven-second one closes; `skills/flow/implement.md` names the new floor.
- The implementer dispatch carries a SCRATCH HARNESS paragraph: a throwaway harness lives in the
  session scratchpad or the change's state dir, never in a new worktree or branch.
- gymie's `CLAUDE.md` and `AGENTS.md` Frontend section gains one bullet naming the Playwright/CDP
  limits and how to measure hidden-tab behaviour instead — committed on gymie's kan-960 change
  branch.
