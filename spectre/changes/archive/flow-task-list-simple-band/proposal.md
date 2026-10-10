# flow-task-list-simple-band

## Why

Operator, 2026-10-10: "flow-task-list is too chaotic. I don't understand what is happening during the run." The band drew the main session's own status lines as rows beside the subagent rows, so the same unit appeared twice (`Faster integrate fix-1 (review findings) — fix` and `Faster integrate fix-1 (4 Important, 5 Minor review findings) — in progress`), a unit finished long ago lingered (`Self-review review-2 … — in review`), and between turns no main-agent row was drawn at all.

The operator's spec, verbatim: "Main agent - separate line at the top, always present and displays what task he is doing (NOT subtask!!!) subagents should be indentated and: display currently working agents as in-progress/review/fix/re-review, if there are pending tasks from the task list the should be displayed bellow as pending, if there is no pending -> show the last 2 done above.. If the task is from the task list it must have Task X/N prefix, otherwise no prefix. If the tasks were grouped for example 1+2+3/10 then when it is finished I MUST see 3/10 at the bottom. Not 1/10"

## What changes

- `mods/flow-task-list/hooks/register.tsx`:
  - The `main` row is always the band's first row while the band is drawn, `in progress` while the main turn runs and `waiting` between turns while a subagent runs. Its label is the top-level unit the main agent is on: the latest open status-line unit of its own naming a task of the running plan; else the stage key of the run's latest `flow stage begin`; else the latest open status-line unit of its own; else none. A Bash command's description no longer labels it.
  - A status line is never a row; a unit is the main agent's own when no subagent row shares its key or names one of its task numbers.
  - Under `main`: the running subagent rows, each with a state word of `in progress`, `review`, `fix` or `re-review` (a review key past round 1, one re-run after a fix — `visual-verify-fix-<k>`, `…-reviewer-fix-<k>` — or one naming a re-review; a `Task <x>/<n>` row is a review or re-review only when its words are exactly `review` or `re-review`, otherwise `in progress`), and among them, in spawn order, a failed subagent's `blocked` row until the next main turn starts; then the pending plan tasks no such row and not the `main` row names; and only when none is pending, the last two done subagent rows of the marked `/flow` change — with none marked, those done since the current main turn began — above the running rows, in the room the running rows leave, so no running row is cut for a done one. Five rows at most under `main`.
  - A plan-task row reads `Task <x>/<n>`; a group reads `Task 1+2+3/<n>` while it runs, and any done row with a `Task(s) …/<n>` prefix reads `Task 3/<n>`, its highest task over its own `<n>`, whether or not a plan is still loaded.
- `skills/flow/gated-review-fix.md` names the gated reviewer's re-dispatch description, `Tasks <ids>/<N> (re-review)`.
  - The hint line's tally counts a verifier under the review emoji; the `visual verify` and `verify` words are gone.
- `skills/flow-contracts/pipeline.md` **Progress visibility** and `README.md` state the band as above.
