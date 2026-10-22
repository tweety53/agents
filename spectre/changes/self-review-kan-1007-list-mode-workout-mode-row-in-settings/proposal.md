# self-review-kan-1007-list-mode-workout-mode-row-in-settings

## Why

kan-1007's self-review pass found eight defects, none `big`:

- **F1 (flow-fix).** 54603278 made integrate's `prepare` (and kickoff) create a satellite's missing
  `--base` on its origin, pushing `refs/remotes/origin/<default>:refs/heads/<base>` unasked, while
  **What still stops** (`skills/flow-contracts/operator-prompts-auto-resolution.md`) still lists
  every push and every integrate prompt as confirmed — the contract and the script disagree.
- **F2 (flow-fix).** `hooks/protect-main-checkout.py` collects a `push`'s and a `commit`'s
  arguments up to a separator only, so `2>&1` and `>/dev/null` are read as a remote, a refspec or a
  pathspec; `git -C $r push -q 2>&1 | tail -1` therefore names no `spectre/` destination and the
  untasked-commit check never runs. 889037cb reached origin with no trailer at 12:50:48Z that way.
  The checkout/restore collector already stops at a redirect.
- **F3 (flow-fix).** A stale `--base` was never compared with the branch it tracks: the creating
  run merged develop into the change branches instead of syncing the base, and integrate's reshape
  folded KAN-1010 into the squash. The check exists only as contract prose.
- **F4 (flow-fix).** Panel round 4 began at 12:18:45Z, sixteen seconds after task 8's inline
  commit (12:18:29Z) and before its `close-task`: `flow record dispatch begin` admits a
  `panel-<round>-<slot>` key while tasks.md holds an unticked task. The same run left dispatch 36
  (`task-8-implementer-retry`, store id 7043) stamped 12:21:37Z–12:23:29Z instead of its true
  ~11:59:47Z–12:23:29Z span.
- **F5 (flow-fix).** A fix-round implementer called a first-try 1dp offset clean: the "measures
  before it hands back" paragraph (`skills/flow/verify-fix-loop.md`) does not say a sub-pixel
  leftover is still a leftover.
- **F6 (flow-cost).** A plan-time capture task baselining against a mockup was judged by eye;
  KAN-1000's measure-until-clean rule binds fix rounds only.
- **F7 (flow-automation).** A `.mockups` sidecar line can name a frame the mockups root does not
  hold; `check-sidecar-reach` checks only the screenshot half of each line.
- **F8 (flow-speed).** A round whose diff touched only the mockups root re-ran the whole visual
  suite (18m28s) where composing the affected sidecars' frames answers the question.

## What changes

- **What still stops** carves out the missing-base creation 54603278 automates.
- The hook's push, commit and checkout/restore argument collectors share one walk that stops at a
  redirect.
- A new guard, `check-base-behind.sh`, run at the first `flow.verify`, names every worktree whose
  recorded base lacks commits of the branch it tracks and prints the sync commands; the contract
  forbids merging the tracked branch into a change branch.
- `flow record dispatch begin` refuses a panel reviewer key while tasks.md holds an unticked task,
  and dispatch 7043's window and figures are repaired.
- The fix-loop measurement rule names a 1dp leftover; section D extends the measure-until-clean
  rule to plan-time capture tasks; `check-sidecar-reach` also checks each line's frame exists; a
  mockups-only round composes instead of dispatching a verifier.
